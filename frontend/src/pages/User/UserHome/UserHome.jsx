import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import "./UserHome.css";

function UserHome() {
  const navigate = useNavigate();
  
  // State สำหรับจัดการ Popup Login
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [loginRole, setLoginRole] = useState("admin"); // ค่าเริ่มต้นให้เป็น admin
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  // ฟังก์ชันเปิด Popup ล็อคอินสำหรับเจ้าหน้าที่
  const openLoginModal = () => {
    setIsModalOpen(true);
  };

  const closeLoginModal = () => {
    setIsModalOpen(false);
    setUsername("");
    setPassword("");
    setLoginRole("admin"); // รีเซ็ตค่าเมื่อปิด
  };

  // ฟังก์ชันยิง API Login ไปเช็คที่ Backend
  const handleLoginSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    
    try {
      const response = await fetch("http://localhost:8080/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ 
          username: username, 
          password: password,
          expected_role: loginRole 
        }),
      });

      const data = await response.json();

      if (response.ok) {
        localStorage.setItem("user", JSON.stringify(data.user));

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
      
      {/* แถบ Navigation ด้านบน */}
      <nav className="top-nav">
        <button className="btn-staff-login" onClick={openLoginModal}>
          🔐 เข้าสู่ระบบ (สำหรับเจ้าหน้าที่)
        </button>
      </nav>

      {/* ส่วนหัวของหน้า */}
      <div className="home-header">
        <h1>ระบบแจ้งซ่อมและติดตามงาน</h1>
        <p>คณะวิทยาศาสตร์ มหาวิทยาลัยศิลปากร</p>
      </div>

      {/* ส่วนการ์ดเมนูสำหรับคนแจ้งซ่อม (เหลือแค่ 2 กล่อง) */}
      <div className="cards-grid user-only-grid">
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
      </div>

      {/* ส่วนของ POPUP LOGIN สำหรับเจ้าหน้าที่ */}
      {isModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <span className="close-btn" onClick={closeLoginModal}>&times;</span>
            
            <h3>🔐 ล็อกอินสำหรับเจ้าหน้าที่</h3>
            <p className="modal-subtitle">กรุณาระบุสิทธิ์และข้อมูลเพื่อเข้าสู่ระบบ</p>

            <form onSubmit={handleLoginSubmit}>
              
              {/* Dropdown เลือก Role */}
              <div className="input-group">
                <label>เข้าสู่ระบบในฐานะ <span className="required">*</span></label>
                <select 
                  value={loginRole} 
                  onChange={(e) => setLoginRole(e.target.value)}
                  className="role-select"
                >
                  <option value="admin">⚙️ แอดมิน (ดูแลระบบและจ่ายงาน)</option>
                  <option value="technician">🛠️ ช่างเทคนิค (รับงานและปิดงาน)</option>
                </select>
              </div>

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