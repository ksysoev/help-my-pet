package redis

import (
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

	profiles := &core.PetProfiles{
		Profiles: []core.PetProfile{
			{
				Name:        "Max",
				Species:     "Dog",
				Breed:       "Golden Retriever",
				DateOfBirth: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				Gender:      "Male",
				Weight:      30.5,
			},
		},
	}

	data, err := json.Marshal(profiles)
	assert.NoError(t, err)

	mock.ExpectHSet(petProfilesKey, "user123", data).SetVal(1)

	err = repo.SaveProfiles("user123", profiles)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPetProfileRepository_GetProfiles(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := NewPetProfileRepository(client)

	t.Run("existing profiles", func(t *testing.T) {
		profiles := &core.PetProfiles{
			Profiles: []core.PetProfile{
				{
					Name:        "Max",
					Species:     "Dog",
					Breed:       "Golden Retriever",
					DateOfBirth: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
					Gender:      "Male",
					Weight:      30.5,
				},
			},
		}

		data, err := json.Marshal(profiles)
		assert.NoError(t, err)

		mock.ExpectHGet(petProfilesKey, "user123").SetVal(string(data))

		result, err := repo.GetProfiles("user123")
		assert.NoError(t, err)
		assert.Equal(t, profiles, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("non-existing profiles", func(t *testing.T) {
		mock.ExpectHGet(petProfilesKey, "user456").RedisNil()

		result, err := repo.GetProfiles("user456")
		assert.NoError(t, err)
		assert.Equal(t, &core.PetProfiles{Profiles: []core.PetProfile{}}, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
