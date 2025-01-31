package core

import "time"

// PetProfile represents a pet's profile information
type PetProfile struct {
	Name        string    `json:"name"`
	Species     string    `json:"species"`
	Breed       string    `json:"breed"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	Weight      float64   `json:"weight"`
}

// PetProfiles represents a collection of pet profiles for a user
type PetProfiles struct {
	Profiles []PetProfile `json:"profiles"`
}

// PetProfileRepository defines the interface for pet profile storage operations
type PetProfileRepository interface {
	SaveProfiles(userID string, profiles *PetProfiles) error
	GetProfiles(userID string) (*PetProfiles, error)
}
