package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/redis/go-redis/v9"
)

const petProfilesKey = "pet_profiles"

// PetProfileRepository implements core.PetProfileRepository using Redis
type PetProfileRepository struct {
	client *redis.Client
}

// NewPetProfileRepository creates a new Redis-based pet profile repository
func NewPetProfileRepository(client *redis.Client) *PetProfileRepository {
	return &PetProfileRepository{
		client: client,
	}
}

// SaveProfiles saves pet profiles for a user
func (r *PetProfileRepository) SaveProfiles(userID string, profiles *core.PetProfiles) error {
	data, err := json.Marshal(profiles)
	if err != nil {
		return fmt.Errorf("failed to marshal pet profiles: %w", err)
	}

	if err := r.client.HSet(context.Background(), petProfilesKey, userID, data).Err(); err != nil {
		return fmt.Errorf("failed to save pet profiles: %w", err)
	}

	return nil
}

// GetProfiles retrieves pet profiles for a user
func (r *PetProfileRepository) GetProfiles(userID string) (*core.PetProfiles, error) {
	data, err := r.client.HGet(context.Background(), petProfilesKey, userID).Bytes()
	if err == redis.Nil {
		// Return empty profiles if not found
		return &core.PetProfiles{Profiles: make([]core.PetProfile, 0)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pet profiles: %w", err)
	}

	var profiles core.PetProfiles
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pet profiles: %w", err)
	}

	return &profiles, nil
}
