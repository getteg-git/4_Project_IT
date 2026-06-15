import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import "./UserHome.css";

function UserHome() {
  const navigate = useNavigate();
  
  // State สำหรับจัดการ Popup Login
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [loginRole, setLoginRole] = useState(""); 
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  // ฟังก์ชันเปิด Popup ล็อคอิน (ปรับให้เซต role เป็น "technician" หรือ "admin")
  const openLoginModal = (role) => {
    setLoginRole(role === "tech" ? "technician" : "admin");
    setIsModalOpen(true);
  };

  const closeLoginModal = () => {
    setIsModalOpen(false);
    setUsername("");
    setPassword("");
  };

  // ฟังก์ชันยิง API Login ไปเช็คที่ Backend
  const handleLoginSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    
    try {
      // 🔥 แก้ไข URL ให้ชี้ไปที่ /api/login ให้ตรงกับ routes.go
      const response = await fetch("http://localhost:8080/api/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ 
          username: username, 
          password: password,
          expected_role: loginRole // 🔥 แก้คีย์ให้ตรงกับ ExpectedRole ใน Go
        }),
      });

      const data = await response.json();

      if (response.ok) {
        // เก็บข้อมูล User ลง LocalStorage
        localStorage.setItem("user", JSON.stringify(data.user));

        // นำทางไปยังหน้าของแต่ละ Role
        if (data.user.role === "admin") {
          navigate("/admin/home");
        } else if (data.user.role === "technician") {
          navigate("/tech/home");
        }
        
        closeLoginModal();
      } else {
        alert(`❌ ไม่สามารถเข้าสู่ระบบได้: ${data.error}`);
      }
    } catch (error) {
      console.error("Login Error:", error);
      alert("❌ ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้ โปรดตรวจสอบระบบหลังบ้าน");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="home-container">
      {/* ส่วนหัวของหน้า */}
      <div className="home-header">
        <h1>ระบบแจ้งซ่อมและติดตามงาน</h1>
        <p>คณะวิทยาศาสตร์ มหาวิทยาลัยศิลปากร</p>
      </div>

      {/* ส่วนการ์ดเมนู 4 ปุ่ม (ปรับ UX ให้สื่อความหมายชัดเจนที่สุด) */}
      <div className="cards-grid">
        <div className="menu-card card-user" onClick={() => navigate("/repair/create")}>
          <div className="card-icon">📝</div>
          <h2>แจ้งซ่อมอุปกรณ์ / แจ้งปัญหาพื้นที่</h2>
          <span className="card-desc">คลิกเพื่อกรอกแบบฟอร์มแจ้งปัญหาใหม่</span>
        </div>

        <div className="menu-card card-user" onClick={() => navigate("/repair/history")}>
          <div className="card-icon">🔍</div>
          <h2>ติดตามสถานะงานแจ้งซ่อม</h2>
          <span className="card-desc">คลิกเพื่อดูความคืบหน้าของงานที่คุณแจ้งไว้</span>
        </div>

        <div className="menu-card card-staff" onClick={() => openLoginModal("admin")}>
          <div className="card-icon">⚙️</div>
          <h2>เข้าสู่ระบบแอดมิน (Admin)</h2>
          <span className="card-desc">คลิกเพื่อล็อกอิน (สำหรับเจ้าหน้าที่ดูแลระบบ)</span>
        </div>

        <div className="menu-card card-staff" onClick={() => openLoginModal("tech")}>
          <div className="card-icon">🛠️</div>
          <h2>เข้าสู่ระบบช่างเทคนิค (Technician)</h2>
          <span className="card-desc">คลิกเพื่อล็อกอิน (สำหรับช่างรับงาน/ปิดงาน)</span>
        </div>
      </div>

      {/* ส่วนของ POPUP LOGIN */}
      {isModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <span className="close-btn" onClick={closeLoginModal}>&times;</span>
            
            <h3>{loginRole === "admin" ? "⚙️ ล็อกอินสำหรับ แอดมิน" : "🛠️ ล็อกอินสำหรับ ช่างเทคนิค"}</h3>
            <p className="modal-subtitle">กรุณากรอกชื่อผู้ใช้งานและรหัสผ่านเพื่อเข้าสู่ระบบ</p>

            <form onSubmit={handleLoginSubmit}>
              <div className="input-group">
                <label>ชื่อผู้ใช้งาน <span className="required">*</span></label>
                <input 
                  type="text" 
                  placeholder="ระบุ Username ของคุณ"
                  value={username} 
                  onChange={(e) => setUsername(e.target.value)} 
                  required 
                />
              </div>
              <div className="input-group">
                <label>รหัสผ่าน <span className="required">*</span></label>
                <input 
                  type="password" 
                  placeholder="ระบุ Password"
                  value={password} 
                  onChange={(e) => setPassword(e.target.value)} 
                  required 
                />
              </div>
              
              <div className="modal-actions">
                <button type="button" className="btn-cancel" onClick={closeLoginModal} disabled={isLoading}>
                  ❌ ปิดหน้าต่าง
                </button>
                <button type="submit" className="btn-submit" disabled={isLoading}>
                  {isLoading ? "⏳ กำลังตรวจสอบ..." : "✅ ยืนยันเพื่อเข้าสู่ระบบ"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default UserHome;