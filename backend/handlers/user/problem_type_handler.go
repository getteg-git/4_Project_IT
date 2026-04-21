package user

import (
	"backend/database"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetProblemTypes(c *gin.Context) {

	rows, err := database.DB.Query(
		"SELECT id, name FROM problem_types ORDER BY id",
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer rows.Close()

	var types []models.ProblemType

	for rows.Next() {

		var t models.ProblemType

		err := rows.Scan(
			&t.ID,
			&t.Name,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		types = append(types, t)
	}

	c.JSON(http.StatusOK, types)
}
