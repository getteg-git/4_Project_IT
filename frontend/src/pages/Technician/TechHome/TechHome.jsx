import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import "./TechHome.css";

function TechHome() {
  const navigate = useNavigate();

  // State เก็บข้อมูลผู้ใช้งานปัจจุบัน (ช่างเทคนิค)
  const [currentUser, setCurrentUser] = useState(null);
  
  // State เก็บข้อมูลงานซ่อม
  const [repairs, setRepairs] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");

  // State สำหรับ Modal พิมพ์รายละเอียดการปิดงาน
  const [modalOpen, setModalOpen] = useState(false);
  const [currentTicket, setCurrentTicket] = useState(null);
  const [closeType, setCloseType] = useState(""); // "success" หรือ "fail"
  const [techDetail, setTechDetail] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  // 1. โหลดข้อมูลช่างจาก LocalStorage และดึงข้อมูลงานจาก API
  const fetchMyJobs = async (user) => {
    try {
      setIsLoading(true);
      const response = await fetch("http://localhost:8080/api/repairs");
      if (response.ok) {
        const allRepairs = await response.json();
        // กรองเอาเฉพาะงานที่ assigned ให้ช่างคนนี้
        const myJobs = allRepairs.filter(r => r.technician_id === user.id);
        setRepairs(myJobs);
      }
    } catch (error) {
      console.error("ดึงข้อมูลงานซ่อมไม่สำเร็จ:", error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    const storedUser = localStorage.getItem("user");
    if (storedUser) {
      const parsedUser = JSON.parse(storedUser);
      setCurrentUser(parsedUser);
      fetchMyJobs(parsedUser);
    } else {
      // ถ้าไม่มีการล็อกอิน ให้เด้งกลับหน้าแรก
      navigate("/");
    }
  }, [navigate]);

  // Smart Search แบบอัจฉริยะ (เพิ่มการหาเลขห้อง)
  const normalizedSearch = searchTerm.replace(/\s+/g, '').replace(/วิทย์/g, 'วิทยาศาสตร์').toLowerCase();
  
  const filteredRepairs = repairs.filter((repair) => {
    if (!normalizedSearch) return true;
    const loc = (repair.location || "").replace(/\s+/g, '').toLowerCase();
    const floor = (repair.floor_name || "").replace(/\s+/g, '').toLowerCase();
    const room = (repair.room || "").replace(/\s+/g, '').toLowerCase(); // 🔥 ค้นหาด้วยเลขห้องได้
    const status = (repair.status || "").replace(/\s+/g, '').toLowerCase();
    const ticketId = String(repair.id);
    
    return (
      loc.includes(normalizedSearch) || floor.includes(normalizedSearch) || room.includes(normalizedSearch) ||
      status.includes(normalizedSearch) || ticketId.includes(normalizedSearch)
    );
  });

  const getStatusClass = (status) => {
    switch (status) {
      case "รอซ่อม": return "status-pending";
      case "กำลังซ่อม": return "status-progress";
      case "เสร็จเรียบร้อย": return "status-completed";
      case "ซ่อมไม่ได้": return "status-failed";
      default: return "status-default";
    }
  };

  const formatDate = (dateString) => {
    const options = { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' };
    return new Date(dateString).toLocaleDateString('th-TH', options);
  };

  // -----------------------------------------
  // ฟังก์ชัน Action ของช่างเทคนิค
  // -----------------------------------------

  // 1. ปฏิเสธงาน (ยิง API Revoke ส่งคืน Admin)
  const handleReject = async (id) => {
    const confirmReject = window.confirm(`⚠️ คุณต้องการ "ปฏิเสธงาน" Ticket #${id} และส่งคืนให้แอดมิน ใช่หรือไม่?`);
    if (!confirmReject) return;

    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${id}/reject`, { // 🔥 เปลี่ยนปลายทางให้ถูก
        method: "PUT"
      });

      if (response.ok) {
        alert("✅ ส่งงานคืนระบบให้แอดมินพิจารณาใหม่เรียบร้อยแล้ว");
        fetchMyJobs(currentUser); // รีเฟรชหน้าจอ
      } else {
        alert("❌ เกิดข้อผิดพลาด ไม่สามารถปฏิเสธงานได้");
      }
    } catch (error) {
      console.error("Reject Error:", error);
    }
  };

  // 2. เปิดหน้าต่าง Modal เพื่อพิมพ์รายละเอียดปิดงาน
  const openCloseModal = (id, type) => {
    setCurrentTicket(id);
    setCloseType(type);
    setTechDetail("");
    setModalOpen(true);
  };

  // 3. ยืนยันการปิดงาน (ยิง API อัปเดตสถานะและบันทึกของช่าง)
  const handleSubmitClose = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    const newStatus = closeType === "success" ? "เสร็จเรียบร้อย" : "ซ่อมไม่ได้";
    
    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${currentTicket}/status`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ 
          status: newStatus,
          technician_note: techDetail
        })
      });

      if (response.ok) {
        alert(`✅ บันทึกสถานะ "${newStatus}" เรียบร้อยแล้ว`);
        setModalOpen(false);
        fetchMyJobs(currentUser); // รีเฟรชข้อมูล
      } else {
        const errorData = await response.json();
        alert(`❌ เกิดข้อผิดพลาด: ${errorData.error}`);
      }
    } catch (error) {
      console.error("Close Job Error:", error);
      alert("❌ ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleLogout = () => {
    if (window.confirm("ต้องการออกจากระบบใช่หรือไม่?")) {
      localStorage.removeItem("user");
      navigate("/");
    }
  };

  return (
    <div className="tech-container">
      <div className="tech-wrapper">
        
        {/* ส่วนหัวของ Tech */}
        <div className="tech-header">
          <div>
            <h2>🛠️ ระบบจัดการงานซ่อม (Technician Panel)</h2>
            {/* 🔥 ดึง full_name มาโชว์ ถ้าไม่มีให้ดึง username โชว์แทน */}
            <p>ยินดีต้อนรับ, {currentUser ? (currentUser.full_name || currentUser.username) : "กำลังโหลด..."}</p>
          </div>
          <button className="btn-logout" onClick={handleLogout}>🚪 ออกจากระบบ</button>
        </div>

        <div className="search-section">
          <input 
            type="text" 
            placeholder="🔍 ค้นหารหัสงาน, สถานที่, เลขห้อง หรือสถานะ..." 
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
        </div>

        {/* รายการใบงาน */}
        <div className="repair-list">
          {isLoading ? (
            <div className="loading-state">⏳ กำลังโหลดรายการงานของคุณ...</div>
          ) : filteredRepairs.length > 0 ? (
            filteredRepairs.map((repair) => (
              <div className="repair-item-card" key={repair.id}>
                
                <div className="card-top">
                  <span className="repair-id">Ticket #{repair.id}</span>
                  <span className={`status-badge ${getStatusClass(repair.status)}`}>
                    {repair.status}
                  </span>
                </div>
                
                <h3 className="repair-title">[{repair.problem_type}] {repair.location}</h3>
                
                {/* 🔥 เพิ่มการแสดงผลเลขห้อง/พิกัด ให้ช่างเห็น */}
                <p className="repair-info">
                  <strong>ชั้น / พิกัด:</strong> {repair.floor_name || "-"} {repair.room ? `(ห้อง ${repair.room})` : ""}
                </p>
                <p className="repair-info"><strong>อาการที่แจ้ง:</strong> {repair.description}</p>
                
                {/* ถ้าช่างพิมพ์รายละเอียดการซ่อมไว้ ให้แสดงตรงนี้ */}
                {repair.technician_note && (
                  <div className="tech-note-display">
                    <strong>📝 บันทึกจากคุณ:</strong> {repair.technician_note}
                  </div>
                )}
                
                <div className="card-actions-row">
                  <span className="repair-date">📅 แจ้งเมื่อ: {formatDate(repair.created_at)}</span>
                  
                  {/* แสดงปุ่มเฉพาะงานที่ "กำลังซ่อม" เท่านั้น (งานเสร็จแล้วจะกดไม่ได้) */}
                  {repair.status === "กำลังซ่อม" && (
                    <div className="action-buttons tech-actions">
                      <button className="btn-reject" onClick={() => handleReject(repair.id)}>
                        ↩️ ปฏิเสธงาน
                      </button>

                      <button className="btn-cannot-fix" onClick={() => openCloseModal(repair.id, "fail")}>
                        ❌ ซ่อมไม่ได้
                      </button>

                      <button className="btn-complete" onClick={() => openCloseModal(repair.id, "success")}>
                        ✅ ซ่อมสำเร็จ
                      </button>
                    </div>
                  )}
                </div>
              </div>
            ))
          ) : (
            <div className="no-results">
              <p>🎉 ไม่มีรายการงานซ่อมที่ต้องดำเนินการในขณะนี้</p>
            </div>
          )}
        </div>

      </div>

      {/* ==========================================
          POPUP MODAL สำหรับพิมพ์รายละเอียดตอนปิดงาน
          ========================================== */}
      {modalOpen && (
        <div className="modal-overlay">
          <div className="modal-box tech-modal">
            <span className="close-btn" onClick={() => setModalOpen(false)}>&times;</span>
            
            <h3 className={closeType === "success" ? "text-success" : "text-fail"}>
              {closeType === "success" ? "✅ ยืนยันการซ่อมเสร็จเรียบร้อย" : "❌ ระบุเหตุผลที่ซ่อมไม่ได้"}
            </h3>
            <p className="modal-subtitle">รหัสใบงาน: Ticket #{currentTicket}</p>

            <form onSubmit={handleSubmitClose}>
              <div className="input-group">
                <label>
                  {closeType === "success" ? "📝 ระบุรายละเอียดการซ่อมแซม" : "📝 ระบุเหตุผลขัดข้อง"} 
                  <span className="required">*</span>
                </label>
                <textarea 
                  rows="4"
                  placeholder={closeType === "success" ? "เช่น เปลี่ยนอะไหล่ชุดลูกลอยใหม่, ล้างแอร์และเติมน้ำยา..." : "เช่น อะไหล่ในสต็อกหมด, อาการหนักเกินขอบเขตต้องจ้างช่างภายนอก..."}
                  value={techDetail}
                  onChange={(e) => setTechDetail(e.target.value)}
                  required
                ></textarea>
              </div>
              
              <div className="modal-actions" style={{ marginTop: "25px" }}>
                <button type="button" className="btn-cancel" onClick={() => setModalOpen(false)} disabled={isSubmitting}>
                  ❌ ยกเลิก
                </button>
                <button 
                  type="submit" 
                  className={closeType === "success" ? "btn-submit-success" : "btn-submit-fail"}
                  disabled={isSubmitting}
                >
                  {isSubmitting ? "⏳ กำลังบันทึก..." : "💾 บันทึกและปิดงาน"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

    </div>
  );
}

export default TechHome;