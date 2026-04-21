package database

import (
	"backend/models"
)

func SearchUsers(keyword string) ([]models.User, error) {

	query := `
		SELECT id, username, password
		FROM users
		WHERE username ILIKE $1
	`

	rows, err := DB.Query(query, "%"+keyword+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Password,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
