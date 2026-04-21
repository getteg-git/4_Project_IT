package user

import (
	"backend/database"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetLocations(c *gin.Context) {

	rows, err := database.DB.Query(
		"SELECT id, name FROM locations ORDER BY id",
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer rows.Close()

	var locations []models.Location

	for rows.Next() {

		var loc models.Location

		err := rows.Scan(
			&loc.ID,
			&loc.Name,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		locations = append(locations, loc)
	}

	c.JSON(http.StatusOK, locations)
}
