import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { CheckCircle2, ClipboardPenLine, Cog, Hammer, Info, LoaderCircle, LogIn, Search, ShieldCheck, Wrench, X } from "lucide-react";
import useToast from "../../../hooks/useToast";
import "./UserHome.css";

function UserHome() {
  const navigate = useNavigate();
  const { toast } = useToast();

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isHelpOpen, setIsHelpOpen] = useState(false);
  const [loginRole, setLoginRole] = useState("admin");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isLoginSuccess, setIsLoginSuccess] = useState(false);

  const openLoginModal = () => setIsModalOpen(true);
  const closeHelpModal = () => setIsHelpOpen(false);

  useEffect(() => {
    if (!isHelpOpen) return undefined;

    const handleKeyDown = (event) => {
      if (event.key === "Escape") closeHelpModal();
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isHelpOpen]);

  const closeLoginModal = () => {
    setIsModalOpen(false);
    setUsername("");
    setPassword("");
    setLoginRole("admin");
    setIsLoginSuccess(false);
  };

  const handleLoginSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setIsLoginSuccess(false);

    try {
      const response = await fetch("https://4projectit-production.up.railway.app/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          username: username,
          password: password,
          expected_role: loginRole,
        }),
      });

      const data = await response.json();

      if (response.ok) {
        localStorage.setItem("user", JSON.stringify(data.user));
        setIsLoginSuccess(true);
        toast.success("เข้าสู่ระบบสำเร็จ", {
          description: `ยินดีต้อนรับ ${data.user?.role === "admin" ? "แอดมิน" : "ช่างเทคนิค"}`,
        });
        await new Promise((resolve) => window.setTimeout(resolve, 650));

        if (data.user.role === "admin") {
          navigate("/admin/home", { replace: true });
        } else if (data.user.role === "technician") {
          navigate("/tech/home", { replace: true });
        }

        closeLoginModal();
      } else {
        toast.error("เข้าสู่ระบบไม่สำเร็จ", {
          description: data.error || "กรุณาตรวจสอบชื่อผู้ใช้และรหัสผ่าน",
        });
      }
    } catch (error) {
      console.error("Login Error:", error);
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", {
        description: "กรุณาลองใหม่อีกครั้ง หรือติดต่อผู้ดูแลระบบ",
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="home-container">
      <div className="top-band" aria-hidden="true">
        <span className="tool-float tool-float-1"><Wrench /></span>
        <span className="tool-float tool-float-2"><Cog /></span>
        <span className="tool-float tool-float-3"><Hammer /></span>
      </div>

      <nav className="top-nav">
        <button
          className="btn-help"
          type="button"
          onClick={() => setIsHelpOpen(true)}
          aria-label="วิธีใช้งานระบบ"
          aria-haspopup="dialog"
          aria-expanded={isHelpOpen}
        >
          <Info size={20} aria-hidden="true" />
          <span>วิธีใช้งาน</span>
        </button>
        <button className="btn-staff-login" onClick={openLoginModal}>
          <span className="lock-icon" aria-hidden="true">🔐</span>
          เข้าสู่ระบบเจ้าหน้าที่
        </button>
      </nav>

      <header className="home-header">
        {/* <span className="eyebrow">SCIENCE FACULTY SERVICE</span> */}
        <h1>แจ้งซ่อมและติดตามงาน</h1>
        <p>คณะวิทยาศาสตร์ มหาวิทยาลัยศิลปากร</p>
      </header>

      <div className="cards-grid user-only-grid">
        <button type="button" className="menu-card card-user" onClick={() => navigate("/repair/create")}>
          <div className="card-icon"><ClipboardPenLine aria-hidden="true" /></div>
          <h2>แจ้งซ่อมอุปกรณ์ได้ที่นี่เลย ทันที!!!</h2>
          <span className="card-desc">คลิกเพื่อกรอกแบบฟอร์มแจ้งปัญหาใหม่</span>
        </button>

        <button type="button" className="menu-card card-user" onClick={() => navigate("/repair/history")}>
          <div className="card-icon"><Search aria-hidden="true" /></div>
          <h2>ติดตามสถานะงานแจ้งซ่อม</h2>
          <span className="card-desc">คลิกเพื่อดูความคืบหน้าของงานที่คุณแจ้งไว้</span>
        </button>
      </div>

      {/* Footer */}
      <footer className="home-footer">
        <div className="footer-content">
          <h3>แจ้งซ่อมและติดตามงาน</h3>
          <p>คณะวิทยาศาสตร์ มหาวิทยาลัยศิลปากร</p>
          <span>© 2026 Faculty of Science</span>
        </div>
      </footer>

      {isHelpOpen && (
        <div className="modal-overlay" role="presentation" onMouseDown={(event) => {
          if (event.target === event.currentTarget) closeHelpModal();
        }}>
          <section
            className="modal-box help-modal-box"
            role="dialog"
            aria-modal="true"
            aria-labelledby="help-modal-title"
            aria-describedby="help-modal-description"
          >
            <button className="close-btn" type="button" onClick={closeHelpModal} aria-label="ปิดวิธีใช้งาน">
              <X aria-hidden="true" />
            </button>
            <h3 id="help-modal-title"><Info size={22} aria-hidden="true" /> วิธีใช้งานระบบ</h3>
            <p className="modal-subtitle" id="help-modal-description">เลือกเมนูให้ตรงกับสิ่งที่ต้องการทำ</p>
            <ol className="help-steps">
              <li>
                <strong>แจ้งซ่อม</strong>
                <span>เลือก “แจ้งซ่อมอุปกรณ์” แล้วกรอกรายละเอียดปัญหาและสถานที่</span>
              </li>
              <li>
                <strong>ติดตามงาน</strong>
                <span>เลือก “ติดตามสถานะงานแจ้งซ่อม” เพื่อดูความคืบหน้าของรายการที่แจ้งไว้</span>
              </li>
              <li>
                <strong>รอรับการซ่อม</strong>
                <span>เจ้าหน้าที่รับเรื่องและช่างจะดำเนินการ ก่อนแจ้งผลเพื่อให้ตรวจสอบและประเมินงาน</span>
              </li>
            </ol>
          </section>
        </div>
      )}

      {isModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <button className="close-btn" type="button" onClick={closeLoginModal} disabled={isLoading} aria-label="ปิดหน้าต่างเข้าสู่ระบบ"><X aria-hidden="true" /></button>

            <h3><ShieldCheck size={22} aria-hidden="true" /> เข้าสู่ระบบเจ้าหน้าที่</h3>
            <p className="modal-subtitle">กรุณาระบุสิทธิ์และข้อมูลเพื่อเข้าสู่ระบบ</p>

            <form onSubmit={handleLoginSubmit}>
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
