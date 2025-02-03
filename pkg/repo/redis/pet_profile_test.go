package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestPetProfileRepository_SaveProfiles(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		profile     *core.PetProfile
		mockSetup   func(mock redismock.ClientMock, userID string, profile *core.PetProfile) []byte
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "success case",
			userID: "user123",
			profile: &core.PetProfile{
				Name:        "Max",
				Species:     "Dog",
				Breed:       "Golden Retriever",
				DateOfBirth: "2020-01-01",
				Gender:      "Male",
				Weight:      "30.5",
			},
			mockSetup: func(mock redismock.ClientMock, userID string, profile *core.PetProfile) []byte {
				data, _ := json.Marshal(core.PetProfiles{Profiles: []core.PetProfile{*profile}})
				mock.ExpectHSet(petProfilesKey, userID, data).SetVal(1)
				return data
			},
		},
		{
			name:   "marshaling error",
			userID: "user123",
			profile: &core.PetProfile{
				Name: string(make([]byte, 0)), // Simulating invalid data input.
			},
			mockSetup: func(mock redismock.ClientMock, userID string, profile *core.PetProfile) []byte {
				return nil
			},
			wantErr: true,
		},
		{
			name:    "empty profile",
			userID:  "user123",
			profile: &core.PetProfile{},
			mockSetup: func(mock redismock.ClientMock, userID string, profile *core.PetProfile) []byte {
				data, _ := json.Marshal(core.PetProfiles{Profiles: []core.PetProfile{*profile}})
				mock.ExpectHSet(petProfilesKey, userID, data).SetVal(1)
				return data
			},
		},
		{
			name:   "redis failure",
			userID: "user456",
			profile: &core.PetProfile{
				Name:        "Bella",
				Species:     "Cat",
				Breed:       "Siamese",
				DateOfBirth: "2022-06-15",
				Gender:      "Female",
				Weight:      "4.1",
			},
			mockSetup: func(mock redismock.ClientMock, userID string, profile *core.PetProfile) []byte {
				data, _ := json.Marshal(core.PetProfiles{Profiles: []core.PetProfile{*profile}})
				mock.ExpectHSet(petProfilesKey, userID, data).SetErr(fmt.Errorf("redis unavailable"))
				return data
			},
			wantErr:     true,
			expectedErr: fmt.Errorf("failed to save pet profiles: redis unavailable"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			repo := NewPetProfileRepository(client)

			tt.mockSetup(mock, tt.userID, tt.profile)

			err := repo.SaveProfile(context.Background(), tt.userID, tt.profile)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

func TestPetProfileRepository_GetCurrentProfile(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		mockSetup   func(mock redismock.ClientMock, userID string) string
		expected    *core.PetProfile
		expectedErr error
	}{
		{
			name:   "profile exists",
			userID: "user123",
			mockSetup: func(mock redismock.ClientMock, userID string) string {
				profile := core.PetProfile{
					Name:        "Max",
					Species:     "Dog",
					Breed:       "Golden Retriever",
					DateOfBirth: "2020-01-01",
					Gender:      "Male",
					Weight:      "30.5",
				}
				data, _ := json.Marshal(core.PetProfiles{Profiles: []core.PetProfile{profile}})
				mock.ExpectHGet(petProfilesKey, userID).SetVal(string(data))
				return string(data)
			},
			expected: &core.PetProfile{
				Name:        "Max",
				Species:     "Dog",
				Breed:       "Golden Retriever",
				DateOfBirth: "2020-01-01",
				Gender:      "Male",
				Weight:      "30.5",
			},
			expectedErr: nil,
		},
		{
			name:   "redis nil error",
			userID: "user404",
			mockSetup: func(mock redismock.ClientMock, userID string) string {
				mock.ExpectHGet(petProfilesKey, userID).RedisNil()
				return ""
			},
			expected:    nil,
			expectedErr: core.ErrProfileNotFound,
		},
		{
			name:   "redis failure",
			userID: "user500",
			mockSetup: func(mock redismock.ClientMock, userID string) string {
				mock.ExpectHGet(petProfilesKey, userID).SetErr(assert.AnError)
				return ""
			},
			expected:    nil,
			expectedErr: assert.AnError,
		},
		{
			name:   "empty profile list",
			userID: "user123",
			mockSetup: func(mock redismock.ClientMock, userID string) string {
				data, _ := json.Marshal(core.PetProfiles{Profiles: []core.PetProfile{}})
				mock.ExpectHGet(petProfilesKey, userID).SetVal(string(data))
				return string(data)
			},
			expected:    nil,
			expectedErr: core.ErrProfileNotFound,
		},
		{
			name:   "multiple profiles, only first returned",
			userID: "user789",
			mockSetup: func(mock redismock.ClientMock, userID string) string {
				profile1 := core.PetProfile{
					Name:        "Max",
					Species:     "Dog",
					Breed:       "Golden Retriever",
					DateOfBirth: "2020-01-01",
					Gender:      "Male",
					Weight:      "30.5",
				}
				profile2 := core.PetProfile{
					Name:        "Bella",
					Species:     "Cat",
					Breed:       "Siamese",
					DateOfBirth: "2022-06-15",
					Gender:      "Female",
					Weight:      "4.0",
				}
				data, _ := json.Marshal(core.PetProfiles{Profiles: []core.PetProfile{profile1, profile2}})
				mock.ExpectHGet(petProfilesKey, userID).SetVal(string(data))
				return string(data)
			},
			expected: &core.PetProfile{
				Name:        "Max",
				Species:     "Dog",
				Breed:       "Golden Retriever",
				DateOfBirth: "2020-01-01",
				Gender:      "Male",
				Weight:      "30.5",
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			repo := NewPetProfileRepository(client)

			tt.mockSetup(mock, tt.userID)

			result, err := repo.GetCurrentProfile(context.Background(), tt.userID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
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
