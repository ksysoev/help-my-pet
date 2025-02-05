package pet

import (
	"testing"
)

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
