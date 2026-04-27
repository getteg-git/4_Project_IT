package routes

import (
	"github.com/gin-gonic/gin"

	handler "backend/handlers"
	repairHandler "backend/handlers/repair"
	userHandler "backend/handlers/user"
)

func SetupRoutes(r *gin.Engine) {

	// Auth: ระบบเข้าสู่ระบบและสมัครสมาชิก
	r.POST("/login", handler.Login)
	r.POST("/register", handler.Register)

	// User Management: ระบบจัดการข้อมูลผู้ใช้งาน (สำหรับ Admin)
	r.GET("/users/search", handler.SearchUsers)
	r.GET("/users", handler.GetUsers)
	r.POST("/users", handler.CreateUser)
	r.PUT("/users/:id", handler.UpdateUser)
	r.DELETE("/users/:id", handler.DeleteUser)

	// Repair System (Master Data): ดึงข้อมูลพื้นฐานไปแสดงใน Dropdown
	r.GET("/locations", userHandler.GetLocations)
	r.GET("/problem-types", userHandler.GetProblemTypes)

	// Repairs (Transaction): ระบบจัดการงานแจ้งซ่อม
	// ---------------------------------------------------------

	// สำหรับทุกคน: ดูรายการแจ้งซ่อมทั้งหมด หรือดูตาม ID
	r.GET("/repairs", repairHandler.GetAllRepairs)     // ดึงรายการซ่อมทั้งหมด (Admin ดูได้หมด, User ดูของตัวเอง)
	r.GET("/repairs/:id", repairHandler.GetRepairByID) // ดูรายละเอียดงานซ่อมเฉพาะเคส (พร้อมรูปภาพ)

	// สำหรับ User: สร้างรายการแจ้งซ่อมใหม่
	r.POST("/repairs", repairHandler.CreateRepair)

	// สำหรับ Admin: อนุมัติงานและมอบหมายช่าง
	r.PUT("/repairs/:id/approve", repairHandler.ApproveRepair)

	// สำหรับ Technician: อัปเดตสถานะเป็น 'กำลังซ่อม' และอัปโหลดรูปภาพระหว่างทำงาน
	r.PUT("/repairs/:id/progress", repairHandler.UpdateProgress)

	// สำหรับ Technician/Admin: อัปเดตสถานะเป็น 'เสร็จสิ้น' เพื่อปิดงาน
	r.PUT("/repairs/:id/complete", repairHandler.CompleteRepair)

	// ---------------------------------------------------------
}
