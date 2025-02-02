package core

import (
	"context"
	"fmt"
)

var ErrProfileNotFound = fmt.Errorf("pet profile not found")

// PetProfile represents a pet's profile information
type PetProfile struct {
	Name        string `json:"name"`
	Species     string `json:"species"`
	Breed       string `json:"breed"`
	DateOfBirth string `json:"date_of_birth"`
	Gender      string `json:"gender"`
	Weight      string `json:"weight"`
}

// PetProfiles represents a collection of pet profiles for a user
type PetProfiles struct {
	Profiles []PetProfile `json:"profiles"`
}

// PetProfileRepository defines the interface for pet profile storage operations
type PetProfileRepository interface {
	SaveProfile(ctx context.Context, userID string, profile *PetProfile) error
	GetCurrentProfile(ctx context.Context, userID string) (*PetProfile, error)
}

func (p PetProfile) String() string {
	return fmt.Sprintf(`
Pet Profile
Name: %s
Species: %s
Breed: %s
Date of Birth: %s
Gender: %s
Weight: %f
`, p.Name, p.Species, p.Breed, p.DateOfBirth, p.Gender, p.Weight)
}
