package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/ksysoev/help-my-pet/pkg/core/pet"
	"github.com/redis/go-redis/v9"
)

const petProfilesKey = "pet_profiles"

// PetProfileRepository implements core.PetProfileRepository using Redis
type PetProfileRepository struct {
	client *redis.Client
}

// NewPetProfileRepository creates a new instance of PetProfileRepository with the provided Redis client.
// It initializes the repository for managing pet profiles stored in Redis.
// client Redis client used for database operations.
// Returns a pointer to the PetProfileRepository instance.
func NewPetProfileRepository(client *redis.Client) *PetProfileRepository {
	return &PetProfileRepository{
		client: client,
	}
}

// SaveProfile stores a pet profile for a specified user in the database.
// It serializes the pet profile data and saves it under the user's ID in Redis.
// ctx is the context for the operation, allowing cancellation and timeouts.
// userID is the unique identifier for the user owning the pet profile.
// profile is the pet profile information to be saved.
// Returns an error if data serialization fails or if the save operation encounters an issue.
func (r *PetProfileRepository) SaveProfile(ctx context.Context, userID string, profile *pet.Profile) error {
	allProfiles := pet.Profiles{Profiles: []pet.Profile{*profile}}

	data, err := json.Marshal(allProfiles)
	if err != nil {
		return fmt.Errorf("failed to marshal pet profiles: %w", err)
	}

	if err := r.client.HSet(ctx, petProfilesKey, userID, data).Err(); err != nil {
		return fmt.Errorf("failed to save pet profiles: %w", err)
	}

	return nil
}

// GetCurrentProfile retrieves the most recent pet profile associated with a given user from the database.
// It fetches and deserializes the profile data stored under the specified userID in Redis.
// ctx is the context for the operation, supporting cancellation and timeouts.
// userID is the unique identifier for the user whose pet profile is being retrieved.
// Returns the first pet.Profile if profiles exist or core.ErrProfileNotFound if no profiles are available.
// Returns an error if data retrieval or unmarshaling fails.
func (r *PetProfileRepository) GetCurrentProfile(ctx context.Context, userID string) (*pet.Profile, error) {
	data, err := r.client.HGet(ctx, petProfilesKey, userID).Bytes()
	if err == redis.Nil {
		return nil, core.ErrProfileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pet profiles: %w", err)
	}

	var profiles pet.Profiles
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pet profiles: %w", err)
	}

	if len(profiles.Profiles) == 0 {
		return nil, core.ErrProfileNotFound
	}

	return &profiles.Profiles[0], nil
}

// RemoveUserProfiles deletes all pet profiles associated with a specific user ID from the database.
// It removes the entry identified by userID from the Redis hash key storing pet profile data.
// ctx is the context for the operation, supporting cancellation and timeouts.
// userID is the unique identifier for the user whose profiles should be removed.
// Returns an error if the delete operation encounters an issue.
func (r *PetProfileRepository) RemoveUserProfiles(ctx context.Context, userID string) error {
	if err := r.client.HDel(ctx, petProfilesKey, userID).Err(); err != nil {
		return fmt.Errorf("failed to remove pet profiles: %w", err)
	}

	return nil
}
