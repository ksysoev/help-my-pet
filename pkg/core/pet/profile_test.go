package pet

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateAge(t *testing.T) {
	tests := []struct {
		name        string
		now         time.Time
		dateOfBirth string
		expectedAge string
		expectedErr bool
	}{
		{
			name:        "Valid age calculation",
			now:         time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			dateOfBirth: "2000-01-01",
			expectedAge: "23",
			expectedErr: false,
		},
		{
			name:        "Date of birth in the future",
			now:         time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			dateOfBirth: "2025-01-01",
			expectedAge: "Not provided",
			expectedErr: false,
		},
		{
			name:        "Invalid date format",
			now:         time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			dateOfBirth: "01-01-2000",
			expectedAge: "Not provided",
			expectedErr: false,
		},
		{
			name:        "Age less than a year",
			now:         time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			dateOfBirth: "2023-06-01",
			expectedAge: "Less than a year",
			expectedErr: false,
		},
		{
			name:        "Empty date of birth",
			now:         time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			dateOfBirth: "",
			expectedAge: "Not provided",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := calculateAge(tt.now, tt.dateOfBirth)

			// Assert
			assert.Equal(t, tt.expectedAge, result)
		})
	}
}

func TestPetProfile_String(t *testing.T) {
	tests := []struct {
		name     string
		profile  Profile
		expected string
	}{
		{
			name: "All fields set",
			profile: Profile{
				Name:        "Buddy",
				Species:     "Dog",
				Breed:       "Golden Retriever",
				DateOfBirth: "2018-05-10",
				Gender:      "Male",
				Weight:      "30.5",
			},
			expected: `
Pet Profile:
Name: Buddy
Species: Dog
Breed: Golden Retriever
Date of Birth: 2018-05-10
Age: 6
Gender: Male
Weight: 30.5
`,
		},
		{
			name: "Empty fields",
			profile: Profile{
				Name:        "",
				Species:     "",
				Breed:       "",
				DateOfBirth: "",
				Gender:      "",
				Weight:      "",
			},
			expected: `
Pet Profile:
Name: 
Species: 
Breed: 
Date of Birth: 
Age: Not provided
Gender: 
Weight: 
`,
		},
		{
			name: "Partial fields set",
			profile: Profile{
				Name:    "Whiskers",
				Species: "Cat",
				Weight:  "4.2",
			},
			expected: `
Pet Profile:
Name: Whiskers
Species: Cat
Breed: 
Date of Birth: 
Age: Not provided
Gender: 
Weight: 4.2
`,
		},
		{
			name: "Date of birth set alone",
			profile: Profile{
				DateOfBirth: "2020-07-15",
			},
			expected: `
Pet Profile:
Name: 
Species: 
Breed: 
Date of Birth: 2020-07-15
Age: 4
Gender: 
Weight: 
`,
		},
		{
			name: "Negative weight",
			profile: Profile{
				Name:   "Tiny",
				Weight: "-1.5",
			},
			expected: `
Pet Profile:
Name: Tiny
Species: 
Breed: 
Date of Birth: 
Age: Not provided
Gender: 
Weight: -1.5
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.profile.String()
			if result != tt.expected {
				t.Errorf("expected: %q, got: %q", tt.expected, result)
			}
		})
	}
}
