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
func (r *PetProfileRepository) SaveProfile(ctx context.Context, userID string, profile *core.PetProfile) error {
	allProfiles := core.PetProfiles{Profiles: []core.PetProfile{*profile}}

	data, err := json.Marshal(allProfiles)
	if err != nil {
		return fmt.Errorf("failed to marshal pet profiles: %w", err)
	}

	if err := r.client.HSet(ctx, petProfilesKey, userID, data).Err(); err != nil {
		return fmt.Errorf("failed to save pet profiles: %w", err)
	}

	return nil
}

func (r *PetProfileRepository) GetCurrentProfile(ctx context.Context, userID string) (*core.PetProfile, error) {
	data, err := r.client.HGet(ctx, petProfilesKey, userID).Bytes()
	if err == redis.Nil {
		return nil, core.ErrProfileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pet profiles: %w", err)
	}

	var profiles core.PetProfiles
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pet profiles: %w", err)
	}

	if len(profiles.Profiles) == 0 {
		return nil, ErrProfileNotFound
	}

	return &profiles.Profiles[0], nil
}
