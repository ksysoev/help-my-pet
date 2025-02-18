package pet

import (
	"fmt"
	"time"
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
	age := calculateAge(p.DateOfBirth)

	return fmt.Sprintf(`
Pet Profile:
Name: %s
Species: %s
Breed: %s
Date of Birth: %s
Age in years: %s
Gender: %s
Weight: %s
`, p.Name, p.Species, p.Breed, p.DateOfBirth, age, p.Gender, p.Weight)
}

// calculateAge calculates the age in years based on the provided date of birth string in "YYYY-MM-DD" format.
// It handles invalid input by returning "Not provided" and assumes the input string is correctly formatted.
// Returns the age as a string or "Not provided" in case of an error during parsing.
func calculateAge(dateOfBirth string) string {
	dob, err := time.Parse("2006-01-02", dateOfBirth)
	if err != nil {
		return "Not provided"
	}

	now := time.Now()
	age := now.Year() - dob.Year()
	if now.YearDay() < dob.YearDay() {
		age--
	}

	return fmt.Sprintf("%d", age)
}
