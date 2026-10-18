package validation

import (
	"regexp"
	"strings"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/Giuseppe-Compagnone/lwn-engine/types"
	"github.com/dop251/goja/parser"
)

var payloadDecoderDeclaration = regexp.MustCompile(`(?m)\b(function\s+decodeUplink\s*\(|(const|let|var)\s+decodeUplink\s*=\s*(function\s*\(|(\([^)]*\)|[A-Za-z_$][\w$]*)\s*=>))`)

func validatePayloadCodec(issues *types.ValidationErrors, prefix string, codec *contracts.PayloadCodec) {
	if codec == nil || !codec.Enabled {
		return
	}

	script := strings.TrimSpace(codec.Script)
	if script == "" {
		issues.Add(prefix+".script", "required", "payload codec script is required when the codec is enabled")
		return
	}

	if _, err := parser.ParseFile(nil, "payload-codec.js", script, 0); err != nil {
		issues.Add(prefix+".script", "invalid_javascript", "payload codec script contains invalid JavaScript syntax")
		return
	}

	if !payloadDecoderDeclaration.MatchString(script) {
		issues.Add(prefix+".script", "missing_decoder", "payload codec script must define decodeUplink(input)")
	}
}
