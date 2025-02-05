package pet

import (
	"fmt"
)

// Profile represents a pet's profile information
type Profile struct {
	Name        string `json:"name"`
	Species     string `json:"species"`
	Breed       string `json:"breed"`
	DateOfBirth string `json:"date_of_birth"`
	Gender      string `json:"gender"`
	Weight      string `json:"weight"`
}

// Profiles represents a collection of pet profiles for a user
type Profiles struct {
	Profiles []Profile `json:"profiles"`
}

// String generates a formatted string representation of the Profile.
// It includes all fields of the Profile struct, presenting them in a readable layout.
// This method is safe to use with partially filled Profile instances, as it handles empty or missing fields gracefully.
// Returns a string representation of the Profile.
func (p Profile) String() string {
	return fmt.Sprintf(`
Pet Profile:
Name: %s
Species: %s
Breed: %s
Date of Birth: %s
Gender: %s
Weight: %s
`, p.Name, p.Species, p.Breed, p.DateOfBirth, p.Gender, p.Weight)
}
