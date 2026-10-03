// Package lorawan contains the small LoRaWAN frame codec needed by the
// simulation boundary. It intentionally models the wire-visible protocol
// fields and security primitives without trying to be a complete MAC-command
// implementation.
package lorawan

import (
	"crypto/aes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	MTypeJoinRequest         byte = 0
	MTypeJoinAccept          byte = 1
	MTypeUnconfirmedDataUp   byte = 2
	MTypeUnconfirmedDataDown byte = 3
	MTypeConfirmedDataUp     byte = 4
	MTypeConfirmedDataDown   byte = 5
)

var ErrInvalidFrame = errors.New("invalid LoRaWAN frame")

type Packet struct {
	MType      byte
	Confirmed  bool
	DevAddr    uint32
	FCnt       uint32
	FPort      *byte
	FRMPayload []byte
	DevEUI     uint64
	JoinEUI    uint64
	DevNonce   uint16
	MIC        [4]byte
	Raw        []byte
}

type JoinAcceptOptions struct {
	AppNonce          uint32
	NetID             uint32
	DevAddr           uint32
	RX1DataRate       int
	RX1DataRateOffset int
	RXDelay           int
	AppKey            []byte
}

func Parse(raw []byte) (Packet, error) {
	if len(raw) < 5 {
		return Packet{}, fmt.Errorf("%w: frame is too short", ErrInvalidFrame)
	}
	packet := Packet{MType: raw[0] >> 5, Raw: append([]byte(nil), raw...)}
	copy(packet.MIC[:], raw[len(raw)-4:])

	switch packet.MType {
	case MTypeJoinRequest:
		if len(raw) != 23 {
			return Packet{}, fmt.Errorf("%w: join request must contain 23 bytes", ErrInvalidFrame)
		}
		packet.JoinEUI = binary.LittleEndian.Uint64(raw[1:9])
		packet.DevEUI = binary.LittleEndian.Uint64(raw[9:17])
		packet.DevNonce = binary.LittleEndian.Uint16(raw[17:19])
		return packet, nil
	case MTypeUnconfirmedDataUp, MTypeUnconfirmedDataDown,
		MTypeConfirmedDataUp, MTypeConfirmedDataDown:
		return parseData(packet, raw)
	default:
		return packet, nil
	}
}

func parseData(packet Packet, raw []byte) (Packet, error) {
	if len(raw) < 12 {
		return Packet{}, fmt.Errorf("%w: data frame is too short", ErrInvalidFrame)
	}
	packet.Confirmed = packet.MType == MTypeConfirmedDataUp || packet.MType == MTypeConfirmedDataDown
	packet.DevAddr = binary.LittleEndian.Uint32(raw[1:5])
	fCtrl := raw[5]
	fOptsLength := int(fCtrl & 0x0f)
	// The fixed FHDR occupies seven bytes after MHDR; raw indexes therefore
	// start at eight when the FOpts field is empty.
	fhdrLength := 8 + fOptsLength
	if len(raw) < fhdrLength+4 {
		return Packet{}, fmt.Errorf("%w: invalid frame options length", ErrInvalidFrame)
	}
	packet.FCnt = uint32(binary.LittleEndian.Uint16(raw[6:8]))
	if fhdrLength+4 == len(raw) {
		return packet, nil
	}
	fPort := raw[fhdrLength]
	packet.FPort = &fPort
	packet.FRMPayload = append([]byte(nil), raw[fhdrLength+1:len(raw)-4]...)
	return packet, nil
}

type DataFrameOptions struct {
	DevAddr   uint32
	FCnt      uint32
	FPort     *byte
	Payload   []byte
	Confirmed bool
	ACK       bool
	Direction byte
	NwkSKey   []byte
	AppSKey   []byte
}

func BuildDataFrame(options DataFrameOptions) ([]byte, error) {
	if len(options.NwkSKey) != 16 || len(options.AppSKey) != 16 {
		return nil, fmt.Errorf("%w: data frame keys must be 16 bytes", ErrInvalidFrame)
	}
	if options.Direction > 1 {
		return nil, fmt.Errorf("%w: direction must be 0 or 1", ErrInvalidFrame)
	}
	if options.FPort != nil && (*options.FPort == 0 || *options.FPort > 223) {
		return nil, fmt.Errorf("%w: FPort must be between 1 and 223", ErrInvalidFrame)
	}
	mType := MTypeUnconfirmedDataDown
	if options.Direction == 0 {
		mType = MTypeUnconfirmedDataUp
	}
	if options.Confirmed {
		mType += 2
	}
	frame := []byte{mType << 5}
	devAddr := make([]byte, 4)
	binary.LittleEndian.PutUint32(devAddr, options.DevAddr)
	frame = append(frame, devAddr...)
	fCtrl := byte(0)
	if options.ACK {
		fCtrl = 0x20
	}
	frame = append(frame, fCtrl, byte(options.FCnt), byte(options.FCnt>>8))
	if options.FPort != nil || len(options.Payload) > 0 {
		if options.FPort == nil {
			return nil, fmt.Errorf("%w: payload requires FPort", ErrInvalidFrame)
		}
		frame = append(frame, *options.FPort)
		if len(options.Payload) > 0 {
			encrypted, err := cryptPayload(options.AppSKey, options.Direction, options.DevAddr, options.FCnt, options.Payload)
			if err != nil {
				return nil, err
			}
			frame = append(frame, encrypted...)
		}
	}
	mic, err := computeMIC(options.NwkSKey, options.Direction, options.DevAddr, options.FCnt, frame)
	if err != nil {
		return nil, err
	}
	return append(frame, mic...), nil
}

// BuildJoinAccept builds the encrypted LoRaWAN 1.0 Join-Accept PHYPayload.
// The simulator uses deterministic AppNonce/NetID values for both logical
// and real-gateway sessions so the derived session keys stay aligned.
func BuildJoinAccept(options JoinAcceptOptions) ([]byte, error) {
	if len(options.AppKey) != 16 || options.AppNonce > 0xffffff || options.NetID > 0xffffff {
		return nil, fmt.Errorf("%w: invalid join accept input", ErrInvalidFrame)
	}
	if options.RX1DataRate < 0 || options.RX1DataRate > 15 || options.RX1DataRateOffset < 0 || options.RX1DataRateOffset > 7 || options.RXDelay < 0 || options.RXDelay > 255 {
		return nil, fmt.Errorf("%w: invalid join accept radio settings", ErrInvalidFrame)
	}
	payload := make([]byte, 0, 12)
	payload = append(payload, byte(options.AppNonce), byte(options.AppNonce>>8), byte(options.AppNonce>>16))
	payload = append(payload, byte(options.NetID), byte(options.NetID>>8), byte(options.NetID>>16))
	devAddr := make([]byte, 4)
	binary.LittleEndian.PutUint32(devAddr, options.DevAddr)
	payload = append(payload, devAddr...)
	payload = append(payload, byte(options.RX1DataRateOffset<<4|options.RX1DataRate), byte(options.RXDelay))

	mhdr := []byte{MTypeJoinAccept << 5}
	mac, err := cmac(options.AppKey, append(append([]byte(nil), mhdr...), payload...))
	if err != nil {
		return nil, err
	}
	plaintext := append(payload, mac[:4]...)
	encrypted, err := aesDecrypt(options.AppKey, plaintext)
	if err != nil {
		return nil, err
	}
	return append(mhdr, encrypted...), nil
}

func VerifyDataMIC(packet Packet, nwkSKey []byte, direction byte) bool {
	if len(packet.Raw) < 5 || len(nwkSKey) != 16 || direction > 1 {
		return false
	}
	actual, err := computeMIC(nwkSKey, direction, packet.DevAddr, packet.FCnt, packet.Raw[:len(packet.Raw)-4])
	if err != nil {
		return false
	}
	for index := range actual {
		if actual[index] != packet.MIC[index] {
			return false
		}
	}
	return true
}

func VerifyJoinRequestMIC(packet Packet, appKey []byte) bool {
	if packet.MType != MTypeJoinRequest || len(packet.Raw) != 23 || len(appKey) != 16 {
		return false
	}
	mac, err := cmac(appKey, packet.Raw[:19])
	if err != nil {
		return false
	}
	for index := 0; index < 4; index++ {
		if mac[index] != packet.MIC[index] {
			return false
		}
	}
	return true
}

func DecodeHex(value string) ([]byte, error) {
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decode LoRaWAN key: %w", err)
	}
	return decoded, nil
}

func DeriveSessionKeys(appKey []byte, appNonce uint32, netID uint32, devNonce uint16) (nwkSKey, appSKey []byte, err error) {
	if len(appKey) != 16 || appNonce > 0xffffff || netID > 0xffffff {
		return nil, nil, fmt.Errorf("%w: invalid OTAA key derivation input", ErrInvalidFrame)
	}
	var block [16]byte
	block[1] = byte(appNonce)
	block[2] = byte(appNonce >> 8)
	block[3] = byte(appNonce >> 16)
	block[4] = byte(netID)
	block[5] = byte(netID >> 8)
	block[6] = byte(netID >> 16)
	binary.LittleEndian.PutUint16(block[7:9], devNonce)
	nwkBlock := block
	nwkBlock[0] = 0x01
	appBlock := block
	appBlock[0] = 0x02
	nwkSKey, err = aesEncrypt(appKey, nwkBlock[:])
	if err != nil {
		return nil, nil, err
	}
	appSKey, err = aesEncrypt(appKey, appBlock[:])
	return nwkSKey, appSKey, err
}

func cryptPayload(key []byte, direction byte, devAddr uint32, fCnt uint32, payload []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, fmt.Errorf("%w: payload key must be 16 bytes", ErrInvalidFrame)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := make([]byte, len(payload))
	for offset, counter := 0, byte(1); offset < len(payload); counter++ {
		actual := make([]byte, aes.BlockSize)
		actual[0] = 0x01
		actual[5] = direction
		binary.LittleEndian.PutUint32(actual[6:10], devAddr)
		binary.LittleEndian.PutUint32(actual[10:14], fCnt)
		actual[15] = counter
		keystream := make([]byte, aes.BlockSize)
		block.Encrypt(keystream, actual)
		count := len(payload) - offset
		if count > aes.BlockSize {
			count = aes.BlockSize
		}
		for index := 0; index < count; index++ {
			stream[offset+index] = payload[offset+index] ^ keystream[index]
		}
		offset += count
	}
	return stream, nil
}

func computeMIC(key []byte, direction byte, devAddr uint32, fCnt uint32, message []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, fmt.Errorf("%w: MIC key must be 16 bytes", ErrInvalidFrame)
	}
	b0 := make([]byte, 16)
	b0[0] = 0x49
	b0[5] = direction
	binary.LittleEndian.PutUint32(b0[6:10], devAddr)
	binary.LittleEndian.PutUint32(b0[10:14], fCnt)
	binary.BigEndian.PutUint16(b0[14:16], uint16(len(message)))
	input := append(b0, message...)
	mac, err := cmac(key, input)
	if err != nil {
		return nil, err
	}
	return mac[:4], nil
}

func cmac(key, message []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	l := make([]byte, aes.BlockSize)
	block.Encrypt(l, make([]byte, aes.BlockSize))
	k1 := leftShift(l)
	if l[0]&0x80 != 0 {
		k1[aes.BlockSize-1] ^= 0x87
	}
	k2 := leftShift(k1)
	if k1[0]&0x80 != 0 {
		k2[aes.BlockSize-1] ^= 0x87
	}
	blocks := (len(message) + aes.BlockSize - 1) / aes.BlockSize
	if blocks == 0 {
		blocks = 1
	}
	last := make([]byte, aes.BlockSize)
	complete := len(message) > 0 && len(message)%aes.BlockSize == 0
	lastOffset := (blocks - 1) * aes.BlockSize
	if complete {
		copy(last, message[lastOffset:])
		xor(last, k1)
	} else {
		remaining := message[lastOffset:]
		copy(last, remaining)
		last[len(remaining)] = 0x80
		xor(last, k2)
	}
	state := make([]byte, aes.BlockSize)
	for index := 0; index < blocks-1; index++ {
		input := make([]byte, aes.BlockSize)
		copy(input, message[index*aes.BlockSize:(index+1)*aes.BlockSize])
		xor(input, state)
		block.Encrypt(state, input)
	}
	xor(last, state)
	block.Encrypt(state, last)
	return state, nil
}

func leftShift(value []byte) []byte {
	shifted := make([]byte, len(value))
	var carry byte
	for index := len(value) - 1; index >= 0; index-- {
		shifted[index] = value[index]<<1 | carry
		carry = value[index] >> 7
	}
	return shifted
}

func xor(left, right []byte) {
	for index := range left {
		left[index] ^= right[index]
	}
}

func aesEncrypt(key, value []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	result := make([]byte, aes.BlockSize)
	block.Encrypt(result, value)
	return result, nil
}

func aesDecrypt(key, value []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(value)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("%w: AES input must be a multiple of 16 bytes", ErrInvalidFrame)
	}
	result := make([]byte, len(value))
	for offset := 0; offset < len(value); offset += aes.BlockSize {
		block.Decrypt(result[offset:offset+aes.BlockSize], value[offset:offset+aes.BlockSize])
	}
	return result, nil
}
