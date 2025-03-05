package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/ksysoev/help-my-pet/pkg/core/pet"
	"github.com/stretchr/testify/assert"
)

func TestPetProfileRepository_SaveProfiles(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		profile     *pet.Profile
		mockSetup   func(mock redismock.ClientMock, userID string, profile *pet.Profile) []byte
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "success case",
			userID: "user123",
			profile: &pet.Profile{
				Name:        "Max",
				Species:     "Dog",
				Breed:       "Golden Retriever",
				DateOfBirth: "2020-01-01",
				Gender:      "Male",
				Weight:      "30.5",
			},
			mockSetup: func(mock redismock.ClientMock, userID string, profile *pet.Profile) []byte {
				data, _ := json.Marshal(pet.Profiles{Profiles: []pet.Profile{*profile}})
				mock.ExpectHSet(petProfilesKey, userID, data).SetVal(1)
				return data
			},
		},
		{
			name:   "marshaling error",
			userID: "user123",
			profile: &pet.Profile{
				Name: string(make([]byte, 0)), // Simulating invalid data input.
			},
			mockSetup: func(mock redismock.ClientMock, userID string, profile *pet.Profile) []byte {
				return nil
			},
			wantErr: true,
		},
		{
			name:    "empty profile",
			userID:  "user123",
			profile: &pet.Profile{},
			mockSetup: func(mock redismock.ClientMock, userID string, profile *pet.Profile) []byte {
				data, _ := json.Marshal(pet.Profiles{Profiles: []pet.Profile{*profile}})
				mock.ExpectHSet(petProfilesKey, userID, data).SetVal(1)
				return data
			},
		},
		{
			name:   "redis failure",
			userID: "user456",
			profile: &pet.Profile{
				Name:        "Bella",
				Species:     "Cat",
				Breed:       "Siamese",
				DateOfBirth: "2022-06-15",
				Gender:      "Female",
				Weight:      "4.1",
			},
			mockSetup: func(mock redismock.ClientMock, userID string, profile *pet.Profile) []byte {
				data, _ := json.Marshal(pet.Profiles{Profiles: []pet.Profile{*profile}})
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
		expected    *pet.Profile
		expectedErr error
	}{
		{
			name:   "profile exists",
			userID: "user123",
			mockSetup: func(mock redismock.ClientMock, userID string) string {
				profile := pet.Profile{
					Name:        "Max",
					Species:     "Dog",
					Breed:       "Golden Retriever",
					DateOfBirth: "2020-01-01",
					Gender:      "Male",
					Weight:      "30.5",
				}
				data, _ := json.Marshal(pet.Profiles{Profiles: []pet.Profile{profile}})
				mock.ExpectHGet(petProfilesKey, userID).SetVal(string(data))
				return string(data)
			},
			expected: &pet.Profile{
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
				data, _ := json.Marshal(pet.Profiles{Profiles: []pet.Profile{}})
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
				profile1 := pet.Profile{
					Name:        "Max",
					Species:     "Dog",
					Breed:       "Golden Retriever",
					DateOfBirth: "2020-01-01",
					Gender:      "Male",
					Weight:      "30.5",
				}
				profile2 := pet.Profile{
					Name:        "Bella",
					Species:     "Cat",
					Breed:       "Siamese",
					DateOfBirth: "2022-06-15",
					Gender:      "Female",
					Weight:      "4.0",
				}
				data, _ := json.Marshal(pet.Profiles{Profiles: []pet.Profile{profile1, profile2}})
				mock.ExpectHGet(petProfilesKey, userID).SetVal(string(data))
				return string(data)
			},
			expected: &pet.Profile{
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

func TestPetProfileRepository_RemoveUserProfiles(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		mockSetup   func(mock redismock.ClientMock, userID string)
		expectedErr error
	}{
		{
			name:   "successful removal",
			userID: "user123",
			mockSetup: func(mock redismock.ClientMock, userID string) {
				mock.ExpectHDel(petProfilesKey, userID).SetVal(1) // Simulating Redis HDel success
			},
			expectedErr: nil,
		},
		{
			name:   "profile does not exist",
			userID: "user404",
			mockSetup: func(mock redismock.ClientMock, userID string) {
				mock.ExpectHDel(petProfilesKey, userID).SetVal(0) // Simulating Redis HDel success with no key deleted
			},
			expectedErr: nil,
		},
		{
			name:   "redis error",
			userID: "user500",
			mockSetup: func(mock redismock.ClientMock, userID string) {
				mock.ExpectHDel(petProfilesKey, userID).SetErr(fmt.Errorf("redis unavailable")) // Simulating Redis error
			},
			expectedErr: fmt.Errorf("failed to remove pet profiles: redis unavailable"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			repo := NewPetProfileRepository(client)

			tt.mockSetup(mock, tt.userID)

			err := repo.RemoveUserProfiles(context.Background(), tt.userID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
