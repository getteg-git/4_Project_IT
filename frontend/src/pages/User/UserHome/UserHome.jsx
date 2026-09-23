import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { CheckCircle2, ClipboardPenLine, LoaderCircle, LogIn, Search, ShieldCheck, X } from "lucide-react";
import useToast from "../../../hooks/useToast";
import "./UserHome.css";

function UserHome() {
  const navigate = useNavigate();
  const { toast } = useToast();
  
  // State สำหรับจัดการ Popup Login
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [loginRole, setLoginRole] = useState("admin"); // ค่าเริ่มต้นให้เป็น admin
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isLoginSuccess, setIsLoginSuccess] = useState(false);

  // ฟังก์ชันเปิด Popup ล็อคอินสำหรับเจ้าหน้าที่
  const openLoginModal = () => {
    setIsModalOpen(true);
  };

  const closeLoginModal = () => {
    setIsModalOpen(false);
    setUsername("");
    setPassword("");
    setLoginRole("admin"); // รีเซ็ตค่าเมื่อปิด
    setIsLoginSuccess(false);
  };

  // ฟังก์ชันยิง API Login ไปเช็คที่ Backend
  const handleLoginSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setIsLoginSuccess(false);
    
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
        setIsLoginSuccess(true);
        await new Promise((resolve) => window.setTimeout(resolve, 650));

        if (data.user.role === "admin") {
          navigate("/admin/home", { replace: true });
        } else if (data.user.role === "technician") {
          navigate("/tech/home", { replace: true });
        }
        
        closeLoginModal();
      } else {
        toast.error("เข้าสู่ระบบไม่สำเร็จ", { description: data.error || "กรุณาตรวจสอบชื่อผู้ใช้และรหัสผ่าน" });
      }
    } catch (error) {
      console.error("Login Error:", error);
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง หรือติดต่อผู้ดูแลระบบ" });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="home-container">
      
      {/* แถบ Navigation ด้านบน */}
      <nav className="top-nav">
        <button className="btn-staff-login" onClick={openLoginModal}>
          <LogIn size={18} aria-hidden="true" />
          เข้าสู่ระบบเจ้าหน้าที่
        </button>
      </nav>

      {/* ส่วนหัวของหน้า */}
      <div className="home-header">
        <h1>ระบบแจ้งซ่อมและติดตามงาน</h1>
        <p>คณะวิทยาศาสตร์ มหาวิทยาลัยศิลปากร</p>
      </div>

      {/* ส่วนการ์ดเมนูสำหรับคนแจ้งซ่อม (เหลือแค่ 2 กล่อง) */}
      <div className="cards-grid user-only-grid">
        <button type="button" className="menu-card card-user" onClick={() => navigate("/repair/create")}>
          <div className="card-icon"><ClipboardPenLine aria-hidden="true" /></div>
          <h2>แจ้งซ่อมอุปกรณ์ / แจ้งปัญหาพื้นที่</h2>
          <span className="card-desc">คลิกเพื่อกรอกแบบฟอร์มแจ้งปัญหาใหม่</span>
        </button>

        <button type="button" className="menu-card card-user" onClick={() => navigate("/repair/history")}>
          <div className="card-icon"><Search aria-hidden="true" /></div>
          <h2>ติดตามสถานะงานแจ้งซ่อม</h2>
          <span className="card-desc">คลิกเพื่อดูความคืบหน้าของงานที่คุณแจ้งไว้</span>
        </button>
      </div>

      {/* ส่วนของ POPUP LOGIN สำหรับเจ้าหน้าที่ */}
      {isModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <button className="close-btn" type="button" onClick={closeLoginModal} disabled={isLoading} aria-label="ปิดหน้าต่างเข้าสู่ระบบ"><X aria-hidden="true" /></button>
            
            <h3><ShieldCheck size={22} aria-hidden="true" /> เข้าสู่ระบบเจ้าหน้าที่</h3>
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
                  <option value="admin">แอดมิน — ดูแลระบบและมอบหมายงาน</option>
                  <option value="technician">ช่างเทคนิค — รับงานและปิดงาน</option>
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
                <button type="submit" className="btn-submit" disabled={isLoading}>
                  {isLoginSuccess
                    ? <><CheckCircle2 size={18} aria-hidden="true" /> เข้าสู่ระบบสำเร็จ</>
                    : isLoading
                      ? <><LoaderCircle className="spin-icon" size={18} aria-hidden="true" /> กำลังเข้าสู่ระบบ...</>
                      : <><LogIn size={18} aria-hidden="true" /> เข้าสู่ระบบ</>}
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
