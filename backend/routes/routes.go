package routes

import (
	"github.com/gin-gonic/gin"

	handler "backend/handlers"
	dashboardHandler "backend/handlers/dashboard"
	repairHandler "backend/handlers/repair"
	userHandler "backend/handlers/user"
)

func SetupRoutes(r *gin.Engine) {

	// สร้างกลุ่ม api ครอบทุกเส้นทาง (แนะนำให้คงไว้เพื่อความเป็นระเบียบและมาตรฐาน)
	api := r.Group("/api")
	{
		// Login ของช่างและแอดมิน
		api.POST("/login", handler.Login)
		// ใช้สำหรับค้นหา
		api.GET("/users/search", handler.SearchUsers)
		// ดึงข้อมูล users ทั้งหมดในระบบ
		api.GET("/users", handler.GetUsers)
		// ใช้สำหรับสร้างผู้ใช้งาน
		api.POST("/users", handler.CreateUser)
		// ใช้สำหรับแก้ไขข้อมูลผู้ใช้งาน
		api.PUT("/users/:id", handler.UpdateUser)
		// ใช้สำหรับลบข้อมูลผู้ใช้งาน
		api.DELETE("/users/:id", handler.DeleteUser)
		// ใช้สำหรับดึงข้อมูลอาคารทั้งหมด
		api.GET("/locations", userHandler.GetLocations)
		// ใช้สำหรับดึงข้อมูลชั้น ของแต่ละอาคารตาม id ของอาคาร
		api.GET("/locations/:id/floors", userHandler.GetFloorsByLocation)
		// ใช้สำหรับดึงข้อมูลห้อง ของแต่ละชั้นตาม id ของชั้น
		api.GET("/floors/:id/rooms", userHandler.GetRoomsByFloor)
		// ใช้สำหรับดึงข้อมูลอุปกรณ์ ของแต่ละห้องตาม id ของห้อง
		api.GET("/rooms/:id/equipments", userHandler.GetEquipmentsByRoom)
		// ใช้สำหรับดึงข้อมูลหมวดหมู่การซ่อม
		api.GET("/problem-types", userHandler.GetProblemTypes)
		// ใช้สำหรับดึงข้อมูลรายการการแจ้งซ่อมทั้งหมด
		api.GET("/repairs", repairHandler.GetAllRepairs)
		// ใช้สำหรับดึงข้อมูลรายการการแจ้งซ่อมตาม id
		api.GET("/repairs/:id", repairHandler.GetRepairByID)
		// ใช้สำหรับสร้างรายการการแจ้งซ่อม
		api.POST("/repairs", repairHandler.CreateRepair)
		// ใช้สำหรับการมอบหมายงานให้ช่างแต่ละคน
		api.PUT("/repairs/:id/assign", repairHandler.AssignRepair)
		// ใช้สำหรับการดึงงานกลับ ในกรณีที่แอดมินมอบหมายงานให้ช่างผิดความถนัด
		api.PUT("/repairs/:id/revoke", repairHandler.RevokeRepair)
		// ใช้สำหรับยกเลิกงาน ในกรณีไม่คุ้มทุนในการซ่อม
		api.PUT("/repairs/:id/cancel", repairHandler.CancelRepairByAdmin)
		// ใช้สำหรับอนุมัติงานที่ ติดจุดคุ้มทุน แต่แอดมินคิดเห็นว่าควรซ่อมอยู่ดี
		api.PUT("/repairs/:id/approve", repairHandler.ApproveRepairThreshold)
		// ใช้สำหรับการประเมินราคาเบื้องต้นของช่าง
		api.PUT("/repairs/:id/estimate", repairHandler.EstimateRepair)
		// ใช้สำหรับปฏิเสธงาน ในกรณีที่ช่างไม่สามารถซ่อมงานนี้ได้
		api.PUT("/repairs/:id/reject", repairHandler.RejectRepair)
		// ใช้สำหรับกดปิดงาน สำหรับช่างที่ซ่อมงานนั้นๆ เสร็จแล้ว
		api.PUT("/repairs/:id/status", repairHandler.UpdateRepairStatus)
		dashboard := api.Group("/admin/dashboard")
		{
			// ใช้สำหรับดึงข้อมูลสถิติเอามาไว้ในหน้า Dashboard
			dashboard.GET("/overview", dashboardHandler.GetDashboardOverview)
			// ใช้สำหรับคำนวณเวลาเฉลี่ยนในการทำงานของช่าง (คำนวณจากการแจ้งซ่อมที่เสร็จสิ้นทั้งหมด)
			dashboard.GET("/performance", dashboardHandler.GetTechnicianPerformance)
			// ใช้สำหรับวิเคราะห์จุดคุ้มทุน แล้วนำมาใส่ในหน้า Dashboard
			dashboard.GET("/breakeven", dashboardHandler.GetBreakEvenAnalysis)
			// ใช้สำหรับดึงสัดส่วนหมวดหมู่ปัญหา
			dashboard.GET("/problem-stats", dashboardHandler.GetProblemTypeStats)
		}
	}
}
