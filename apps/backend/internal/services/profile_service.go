package services

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	contracts "github.com/Giuseppe-Compagnone/lwn-contracts/generated"
	"github.com/google/uuid"
	"lwn-simulator-backend/internal/apperrors"
	"lwn-simulator-backend/internal/database"
)

var (
	ErrCannotDeleteLastProfile = errors.New("cannot delete the last profile")
)

type ProfileRepository interface {
	GetAll() ([]contracts.Profile, error)
	GetByID(id string) (contracts.Profile, error)
	Save(profiles []contracts.Profile) error
	Update(profile contracts.Profile) error
	Delete(profile contracts.Profile) error
}

type ProfileService struct {
	repository     ProfileRepository
	dataDir        string
	realtime       RealtimePublisher
	deleteGuard    func(string) error
	deleteObserver func(string)
	mu             sync.Mutex
}

func NewProfileService(repository ProfileRepository, dataDir string) *ProfileService {
	return &ProfileService{repository: repository, dataDir: dataDir}
}

func (s *ProfileService) SetRealtimePublisher(publisher RealtimePublisher) {
	s.mu.Lock()
	s.realtime = publisher
	s.mu.Unlock()
}

func (s *ProfileService) SetDeleteGuard(guard func(string) error) {
	s.mu.Lock()
	s.deleteGuard = guard
	s.mu.Unlock()
}

func (s *ProfileService) SetDeleteObserver(observer func(string)) {
	s.mu.Lock()
	s.deleteObserver = observer
	s.mu.Unlock()
}

// EnsureDefaultProfile bootstraps the stable default profile and is safe to
// call on every server startup. Existing profile metadata is preserved.
func (s *ProfileService) EnsureDefaultProfile() (contracts.Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	profiles, err := s.repository.GetAll()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			profiles = []contracts.Profile{}
		} else {
			return contracts.Profile{}, fmt.Errorf("load profiles: %w", err)
		}
	}
	for _, profile := range profiles {
		if profile.ID == database.DefaultProfileID {
			return profile, s.ensureStorage(profile.ID)
		}
	}

	profile := contracts.Profile{ID: database.DefaultProfileID, Name: "Default profile"}
	profiles = append(profiles, profile)
	if err := s.repository.Save(profiles); err != nil {
		return contracts.Profile{}, fmt.Errorf("save default profile: %w", err)
	}
	if err := s.ensureStorage(profile.ID); err != nil {
		return contracts.Profile{}, err
	}
	return profile, nil
}

func (s *ProfileService) List() ([]contracts.Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	profiles, err := s.repository.GetAll()
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	return profiles, nil
}

func (s *ProfileService) Get(id string) (contracts.Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	profile, err := s.repository.GetByID(id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return contracts.Profile{}, apperrors.NotFound("profile %q not found", id)
		}
		return contracts.Profile{}, fmt.Errorf("get profile: %w", err)
	}
	return profile, nil
}

func (s *ProfileService) Create(name string) (contracts.Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name, err := normalizeProfileName(name)
	if err != nil {
		return contracts.Profile{}, err
	}
	profiles, err := s.repository.GetAll()
	if err != nil {
		return contracts.Profile{}, fmt.Errorf("load profiles: %w", err)
	}
	if profileNameExists(profiles, name, "") {
		return contracts.Profile{}, apperrors.Conflict("profile named %q already exists", name)
	}
	profile := contracts.Profile{ID: uuid.NewString(), Name: name}
	profiles = append(profiles, profile)
	if err := s.repository.Save(profiles); err != nil {
		return contracts.Profile{}, fmt.Errorf("save profile: %w", err)
	}
	if err := s.ensureStorage(profile.ID); err != nil {
		rollbackErr := s.repository.Delete(profile)
		return contracts.Profile{}, formatRollbackError("create profile storage", err, rollbackErr)
	}
	s.publishRealtime(contracts.RealtimeProfileCreatedMessage, &profile)
	return profile, nil
}

func (s *ProfileService) Rename(id string, name string) (contracts.Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name, err := normalizeProfileName(name)
	if err != nil {
		return contracts.Profile{}, err
	}
	profiles, err := s.repository.GetAll()
	if err != nil {
		return contracts.Profile{}, fmt.Errorf("load profiles: %w", err)
	}
	var current contracts.Profile
	found := false
	for _, profile := range profiles {
		if profile.ID == id {
			current = profile
			found = true
			break
		}
	}
	if !found {
		return contracts.Profile{}, apperrors.NotFound("profile %q not found", id)
	}
	if profileNameExists(profiles, name, id) {
		return contracts.Profile{}, apperrors.Conflict("profile named %q already exists", name)
	}
	current.Name = name
	if err := s.repository.Update(current); err != nil {
		return contracts.Profile{}, fmt.Errorf("rename profile: %w", err)
	}
	s.publishRealtime(contracts.RealtimeProfileUpdatedMessage, &current)
	return current, nil
}

func (s *ProfileService) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.deleteGuard != nil {
		if err := s.deleteGuard(id); err != nil {
			return err
		}
	}
	profiles, err := s.repository.GetAll()
	if err != nil {
		return fmt.Errorf("load profiles: %w", err)
	}
	if len(profiles) <= 1 {
		return fmt.Errorf("%w: %w", apperrors.ErrConflict, ErrCannotDeleteLastProfile)
	}
	var profile contracts.Profile
	found := false
	for _, candidate := range profiles {
		if candidate.ID == id {
			profile = candidate
			found = true
			break
		}
	}
	if !found {
		return apperrors.NotFound("profile %q not found", id)
	}
	if err := s.repository.Delete(profile); err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}
	storage, storageErr := database.NewProfileStorage(s.dataDir, profile.ID)
	if storageErr != nil {
		return storageErr
	}
	if err := removeProfileStorage(storage); err != nil {
		if rollbackErr := s.repository.Save(profiles); rollbackErr != nil {
			return formatRollbackError("delete profile storage", err, rollbackErr)
		}
		return fmt.Errorf("delete profile storage: %w", err)
	}
	s.publishRealtime(contracts.RealtimeProfileDeletedMessage, &profile)
	if s.deleteObserver != nil {
		s.deleteObserver(profile.ID)
	}
	return nil
}

func (s *ProfileService) publishRealtime(messageType contracts.RealtimeWebSocketMessageType, profile *contracts.Profile) {
	publisher := s.realtime
	if publisher == nil || profile == nil {
		return
	}
	id := profile.ID
	publisher.Publish(contracts.RealtimeWebSocketMessage{
		Type:      messageType,
		ProfileID: &id,
		Profile:   profile,
	})
}

func (s *ProfileService) ensureStorage(id string) error {
	storage, err := database.NewProfileStorage(s.dataDir, id)
	if err != nil {
		return err
	}
	if err := storage.Ensure(); err != nil {
		return fmt.Errorf("ensure profile storage: %w", err)
	}
	if err := storage.EnsureDataFiles(); err != nil {
		return fmt.Errorf("ensure profile data files: %w", err)
	}
	return nil
}

func normalizeProfileName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", apperrors.Invalid("profile name is required")
	}
	if len([]rune(name)) > 100 {
		return "", apperrors.Invalid("profile name cannot exceed 100 characters")
	}
	return name, nil
}

func profileNameExists(profiles []contracts.Profile, name string, excludedID string) bool {
	for _, profile := range profiles {
		if profile.ID != excludedID && strings.EqualFold(profile.Name, name) {
			return true
		}
	}
	return false
}

func removeProfileStorage(storage database.ProfileStorage) error {
	if err := os.RemoveAll(storage.Directory()); err != nil {
		return err
	}
	return nil
}
