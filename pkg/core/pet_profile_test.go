package core

import (
	"testing"
	"time"
)

func TestPetProfile_String(t *testing.T) {
	tests := []struct {
		name     string
		profile  PetProfile
		expected string
	}{
		{
			name: "All fields set",
			profile: PetProfile{
				Name:        "Buddy",
				Species:     "Dog",
				Breed:       "Golden Retriever",
				DateOfBirth: time.Date(2018, 5, 10, 0, 0, 0, 0, time.UTC),
				Gender:      "Male",
				Weight:      30.5,
			},
			expected: `
Pet Profile
Name: Buddy
Species: Dog
Breed: Golden Retriever
Date of Birth: 2018-05-10 00:00:00 +0000 UTC
Gender: Male
Weight: 30.500000
`,
		},
		{
			name: "Empty fields",
			profile: PetProfile{
				Name:        "",
				Species:     "",
				Breed:       "",
				DateOfBirth: time.Time{},
				Gender:      "",
				Weight:      0.0,
			},
			expected: `
Pet Profile
Name: 
Species: 
Breed: 
Date of Birth: 0001-01-01 00:00:00 +0000 UTC
Gender: 
Weight: 0.000000
`,
		},
		{
			name: "Partial fields set",
			profile: PetProfile{
				Name:    "Whiskers",
				Species: "Cat",
				Weight:  4.2,
			},
			expected: `
Pet Profile
Name: Whiskers
Species: Cat
Breed: 
Date of Birth: 0001-01-01 00:00:00 +0000 UTC
Gender: 
Weight: 4.200000
`,
		},
		{
			name: "Date of birth set alone",
			profile: PetProfile{
				DateOfBirth: time.Date(2020, 7, 15, 12, 30, 45, 0, time.UTC),
			},
			expected: `
Pet Profile
Name: 
Species: 
Breed: 
Date of Birth: 2020-07-15 12:30:45 +0000 UTC
Gender: 
Weight: 0.000000
`,
		},
		{
			name: "Negative weight",
			profile: PetProfile{
				Name:   "Tiny",
				Weight: -1.5,
			},
			expected: `
Pet Profile
Name: Tiny
Species: 
Breed: 
Date of Birth: 0001-01-01 00:00:00 +0000 UTC
Gender: 
Weight: -1.500000
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
