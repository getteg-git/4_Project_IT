package routes

import (
	"github.com/gin-gonic/gin"

	"backend/auth"
	handler "backend/handlers"
	dashboardHandler "backend/handlers/dashboard"
	departmentHandler "backend/handlers/departments"
	equipmentHandler "backend/handlers/equipment"
	repairHandler "backend/handlers/repair"
	userHandler "backend/handlers/user"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/login", handler.Login)
		api.POST("/logout", handler.Logout)

		// Public reference data required by the repair submission form.
		api.GET("/locations", equipmentHandler.GetLocations)
		api.GET("/locations/:id/floors", equipmentHandler.GetFloorsByLocation)
		api.GET("/floors/:id/rooms", equipmentHandler.GetRoomsByFloor)
		api.GET("/rooms/:id/equipments", equipmentHandler.GetEquipmentsByRoom)
		api.GET("/problem-types", userHandler.GetProblemTypes)
		api.GET("/departments", userHandler.GetDepartments)
		api.GET("/degree-levels", departmentHandler.GetDegreeLevels)
		api.GET("/degree-levels/:id/departments", departmentHandler.GetDepartmentsByDegree)

		// Public repair access returns only explicitly selected public fields.
		api.POST("/repairs", repairHandler.CreateRepair)
		api.GET("/repairs/public-board", repairHandler.GetPublicRepairBoard)
		api.GET("/repairs/track", repairHandler.GetTrackedRepair)
		api.GET("/repairs/track/:id/logs", repairHandler.GetTrackedRepairLogs)
		api.GET("/repairs/track/:id/images/:image_id", repairHandler.GetTrackedRepairImage)

		staff := api.Group("")
		staff.Use(auth.RequireRoles("admin", "technician"))
		{
			staff.GET("/auth/me", handler.CurrentUser)
			staff.GET("/users", handler.GetUsers)
			staff.GET("/repairs", repairHandler.GetAllRepairs)
			staff.GET("/repairs/:id", repairHandler.GetRepairByID)
			staff.GET("/repairs/:id/logs", repairHandler.GetRepairLogs)
			staff.PUT("/repairs/:id/assign", repairHandler.AssignRepair)
		}

		admin := api.Group("")
		admin.Use(auth.RequireRoles("admin"))
		{
			admin.GET("/users/search", handler.SearchUsers)
			admin.POST("/users", handler.CreateUser)
			admin.PUT("/users/:id", handler.UpdateUser)
			admin.PUT("/users/:id/head-technician", handler.UpdateHeadTechnician)
			admin.GET("/users/:id/responsible-areas", userHandler.GetUserResponsibleAreas)
			admin.PUT("/users/:id/responsible-areas", userHandler.UpdateUserResponsibleAreas)
			admin.GET("/technician-responsible-areas", userHandler.GetTechnicianResponsibleAreas)
			admin.DELETE("/users/:id", handler.DeleteUser)

			admin.POST("/locations", equipmentHandler.CreateLocation)
			admin.PUT("/locations/:id", equipmentHandler.UpdateLocation)
			admin.DELETE("/locations/:id", equipmentHandler.DeleteLocation)
			admin.POST("/problem-types", userHandler.CreateProblemType)
			admin.PUT("/problem-types/:id", userHandler.UpdateProblemType)
			admin.DELETE("/problem-types/:id", userHandler.DeleteProblemType)
			admin.GET("/equipments", equipmentHandler.GetEquipments)
			admin.POST("/equipments", equipmentHandler.CreateEquipment)
			admin.POST("/equipments/import", equipmentHandler.ImportExcelHandler)
			admin.PUT("/equipments/:id", equipmentHandler.UpdateEquipment)
			admin.DELETE("/equipments/:id", equipmentHandler.DeleteEquipment)

			admin.PUT("/repairs/:id/revoke", repairHandler.RevokeRepair)
			admin.PUT("/repairs/:id/cancel", repairHandler.CancelRepairByAdmin)
			admin.PUT("/repairs/:id/close-early", repairHandler.CloseRepairEarly)
			admin.PUT("/repairs/:id/outsource", repairHandler.OutsourceRepair)
			admin.PUT("/repairs/:id/approve", repairHandler.ApproveRepairThreshold)

			dashboard := admin.Group("/admin/dashboard")
			{
				dashboard.GET("/overview", dashboardHandler.GetDashboardOverview)
				dashboard.GET("/performance", dashboardHandler.GetTechnicianPerformance)
				dashboard.GET("/breakeven", dashboardHandler.GetBreakEvenAnalysis)
				dashboard.GET("/problem-stats", dashboardHandler.GetProblemTypeStats)
			}
		}

		technician := api.Group("")
		technician.Use(auth.RequireRoles("technician"))
		{
			technician.GET("/repairs/:id/break-even", repairHandler.GetBreakEvenSummary)
			technician.PUT("/repairs/:id/estimate", repairHandler.EstimateRepair)
			technician.PUT("/repairs/:id/return", repairHandler.RejectRepair)
			technician.PUT("/repairs/:id/reject", repairHandler.RejectRepair)
			technician.PUT("/repairs/:id/status", repairHandler.UpdateRepairStatus)
		}

		uploads := r.Group("/uploads")
		uploads.Use(auth.RequireRoles("admin", "technician"))
		uploads.Static("/", "./uploads")
	}
}
