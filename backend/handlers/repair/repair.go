package repair

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"backend/database"
)

func CreateRepair(c *gin.Context) {

	userID := c.PostForm("user_id")
	locationID := c.PostForm("location_id")
	problemTypeID := c.PostForm("problem_type_id")
	description := c.PostForm("description")

	// สร้าง repair ก่อน
	var repairID int

	err := database.DB.QueryRow(
		`INSERT INTO repairs
		(user_id, location_id, problem_type_id, description)
		VALUES ($1,$2,$3,$4)
		RETURNING id`,
		userID,
		locationID,
		problemTypeID,
		description,
	).Scan(&repairID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// รับรูปหลายรูป
	form, _ := c.MultipartForm()
	files := form.File["images"]

	for _, file := range files {

		filename := fmt.Sprintf(
			"%d_%d%s",
			repairID,
			time.Now().UnixNano(),
			filepath.Ext(file.Filename),
		)

		path := "uploads/" + filename

		err := c.SaveUploadedFile(file, path)
		if err != nil {
			continue
		}

		imageURL := "/uploads/" + filename

		_, err = database.DB.Exec(
			`INSERT INTO repair_images
			(repair_id, image_url)
			VALUES ($1,$2)`,
			repairID,
			imageURL,
		)

		if err != nil {
			fmt.Println(err)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "repair created",
	})
}
