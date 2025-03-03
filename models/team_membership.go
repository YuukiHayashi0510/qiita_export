package models

import "time"

type TeamMembership struct {
	Description    string    `json:"description"`
	Email          string    `json:"email"`
	ID             string    `json:"id"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
	Name           string    `json:"name"`
}
