package models

type Repair struct {
	ID            int    `json:"id"`
	UserID        int    `json:"user_id"`
	LocationID    int    `json:"location_id"`
	ProblemTypeID int    `json:"problem_type_id"`
	Description   string `json:"description"`
}
