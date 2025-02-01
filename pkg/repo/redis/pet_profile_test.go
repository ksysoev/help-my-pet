package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestPetProfileRepository_SaveProfiles(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := NewPetProfileRepository(client)

	profile := &core.PetProfile{
		Name:        "Max",
		Species:     "Dog",
		Breed:       "Golden Retriever",
		DateOfBirth: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		Gender:      "Male",
		Weight:      30.5,
	}

	data, err := json.Marshal(core.PetProfiles{Profiles: []core.PetProfile{*profile}})
	assert.NoError(t, err)

	mock.ExpectHSet(petProfilesKey, "user123", data).SetVal(1)

	err = repo.SaveProfile(context.Background(), "user123", profile)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPetProfileRepository_GetProfiles_Success(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := NewPetProfileRepository(client)

	profile := &core.PetProfile{
		Name:        "Max",
		Species:     "Dog",
		Breed:       "Golden Retriever",
		DateOfBirth: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		Gender:      "Male",
		Weight:      30.5,
	}

	data, err := json.Marshal(core.PetProfiles{Profiles: []core.PetProfile{*profile}})
	assert.NoError(t, err)

	mock.ExpectHGet(petProfilesKey, "user123").SetVal(string(data))

	result, err := repo.GetCurrentProfile(context.Background(), "user123")
	assert.NoError(t, err)
	assert.Equal(t, profile, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestPetProfileRepository_GetProfiles_Empty(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := NewPetProfileRepository(client)

	mock.ExpectHGet(petProfilesKey, "user456").RedisNil()

	result, err := repo.GetCurrentProfile(context.Background(), "user456")
	assert.ErrorIs(t, err, core.ErrProfileNotFound)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())

}
