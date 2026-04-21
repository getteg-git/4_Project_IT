package routes

import (
	"github.com/gin-gonic/gin"

	handler "backend/handlers"
	repairHandler "backend/handlers/repair"
	userHandler "backend/handlers/user"
)

func SetupRoutes(r *gin.Engine) {

	// Auth
	r.POST("/login", handler.Login)
	r.POST("/register", handler.Register)

	// User Management
	r.GET("/users/search", handler.SearchUsers)
	r.GET("/users", handler.GetUsers)
	r.POST("/users", handler.CreateUser)
	r.PUT("/users/:id", handler.UpdateUser)
	r.DELETE("/users/:id", handler.DeleteUser)

	// Repair System
	r.GET("/locations", userHandler.GetLocations)
	r.GET("/problem-types", userHandler.GetProblemTypes)

	// Repairs
	r.POST("/repairs", repairHandler.CreateRepair)
}
