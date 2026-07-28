package routes

import (
	"github.com/gin-gonic/gin"

	handler "backend/handlers"
	dashboardHandler "backend/handlers/dashboard" // 🔥 [เพิ่มใหม่] นำเข้า dashboard handler
	repairHandler "backend/handlers/repair"
	userHandler "backend/handlers/user"
)

func SetupRoutes(r *gin.Engine) {

	// สร้างกลุ่ม api ครอบทุกเส้นทาง (แนะนำให้คงไว้เพื่อความเป็นระเบียบและมาตรฐาน)
	api := r.Group("/api")
	{
		// ==========================================
		// 🟢 AUTH (เข้าสู่ระบบ Admin/Technician)
		// ==========================================
		api.POST("/login", handler.Login)

		// ==========================================
		// 🟢 USER MANAGEMENT (Admin จัดการช่าง)
		// ==========================================
		api.GET("/users/search", handler.SearchUsers)
		api.GET("/users", handler.GetUsers)
		api.POST("/users", handler.CreateUser)
		api.PUT("/users/:id", handler.UpdateUser)
		api.DELETE("/users/:id", handler.DeleteUser)

		// ==========================================
		// 🟢 MASTER DATA (ข้อมูล Dropdown)
		// ==========================================
		api.GET("/locations", userHandler.GetLocations)
		api.GET("/locations/:id/floors", userHandler.GetFloorsByLocation) // 🔥 ดึงข้อมูลชั้นตามตึก
		api.GET("/floors/:id/rooms", userHandler.GetRoomsByFloor)         // 🔥 [เพิ่มใหม่] ดึงข้อมูลห้องตามชั้น
		api.GET("/rooms/:id/equipments", userHandler.GetEquipmentsByRoom) // 🔥 [เพิ่มใหม่] ดึงข้อมูลอุปกรณ์ตามห้อง
		api.GET("/problem-types", userHandler.GetProblemTypes)

		// ==========================================
		// 🟢 REPAIRS (Transaction แจ้งซ่อม)
		// ==========================================

		// ฝั่ง Public (แจ้งซ่อม/ติดตาม)
		api.GET("/repairs", repairHandler.GetAllRepairs)
		api.GET("/repairs/:id", repairHandler.GetRepairByID)
		api.POST("/repairs", repairHandler.CreateRepair)

		// ฝั่ง Admin (มอบหมายงาน / ดึงงานกลับ)
		api.PUT("/repairs/:id/assign", repairHandler.AssignRepair)
		api.PUT("/repairs/:id/revoke", repairHandler.RevokeRepair)

		// ฝั่ง Technician (ปฏิเสธงาน / ปิดงาน)
		api.PUT("/repairs/:id/reject", repairHandler.RejectRepair)
		api.PUT("/repairs/:id/status", repairHandler.UpdateRepairStatus)

		// ==========================================
		// 🟢 ADMIN DASHBOARD (สถิติและรายงาน)
		// ==========================================
		dashboard := api.Group("/admin/dashboard")
		{
			// สถิติภาพรวม
			dashboard.GET("/overview", dashboardHandler.GetDashboardOverview)
			// รายงานการซ่อม
			dashboard.GET("/performance", dashboardHandler.GetTechnicianPerformance)
			// จุดคุ้มทุน
			dashboard.GET("/breakeven", dashboardHandler.GetBreakEvenAnalysis)
		}
	}
}
