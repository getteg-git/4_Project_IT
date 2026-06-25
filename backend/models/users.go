package models

import "time"

type User struct {
	ID             int       `json:"id"`
	Username       string    `json:"username"`
	Password       string    `json:"password,omitempty"`
	FullName       string    `json:"full_name"`
	Role           string    `json:"role"`
	Specialties    []int     `json:"specialties,omitempty"`
	SpecialtyNames []string  `json:"specialty_names,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
