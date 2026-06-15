import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import "./TechHome.css";

function TechHome() {
  const navigate = useNavigate();

  // ข้อมูลจำลอง (สมมติว่าล็อคอินเข้ามาเป็น ช่างสมชาย)
  const [repairs, setRepairs] = useState([
    { id: "REP-002", date: "09 มิ.ย. 2569", location: "ห้องปฏิบัติการเคมี", room: "Lab 3", type: "งานประปา", status: "กำลังซ่อม", desc: "ก๊อกน้ำอ่างล้างมือรั่วซึม", techDetail: "" },
    { id: "REP-005", date: "10 มิ.ย. 2569", location: "อาคารวิทยาศาสตร์ 1", room: "ห้องน้ำหญิง ชั้น 2", type: "งานประปา", status: "กำลังซ่อม", desc: "ท่อน้ำทิ้งตัน น้ำเอ่อล้น", techDetail: "" },
    { id: "REP-004", date: "05 มิ.ย. 2569", location: "อาคารวิทยาศาสตร์ 1", room: "ห้องน้ำชาย ชั้น 1", type: "งานประปา", status: "เสร็จเรียบร้อย", desc: "ชักโครกกดไม่ลง", techDetail: "ทำการเปลี่ยนชุดลูกลอยและยางรองใหม่เรียบร้อยแล้ว ใช้งานได้ปกติ" },
  ]);

  const [searchTerm, setSearchTerm] = useState("");

  // State สำหรับ Modal พิมพ์รายละเอียดการปิดงาน
  const [modalOpen, setModalOpen] = useState(false);
  const [currentTicket, setCurrentTicket] = useState(null);
  const [closeType, setCloseType] = useState(""); // "success" (เสร็จเรียบร้อย) หรือ "fail" (ซ่อมไม่ได้)
  const [techDetail, setTechDetail] = useState("");

  // กรองข้อมูล
  const filteredRepairs = repairs.filter((repair) =>
    repair.location.includes(searchTerm) ||
    repair.room.includes(searchTerm) ||
    repair.status.includes(searchTerm) ||
    repair.id.includes(searchTerm)
  );

  const getStatusClass = (status) => {
    switch (status) {
      case "รอซ่อม": return "status-pending";
      case "กำลังซ่อม": return "status-progress";
      case "เสร็จเรียบร้อย": return "status-completed";
      case "ซ่อมไม่ได้": return "status-failed";
      default: return "status-default";
    }
  };

  // -----------------------------------------
  // ฟังก์ชัน Action ของช่างเทคนิค
  // -----------------------------------------

  // 1. ปฏิเสธงาน (ส่งคืน Admin)
  const handleReject = (id) => {
    const confirmReject = window.confirm(`คุณต้องการ "ปฏิเสธงาน" #${id} และส่งคืนแอดมิน ใช่หรือไม่? (อาจเพราะงานไม่ตรงสาย)`);
    if (confirmReject) {
      // ในระบบจริง จะส่ง API ไปเปลี่ยนสถานะเป็น "รอซ่อม" และลบชื่อช่างออก
      // ส่วนในโค้ดจำลองนี้ เราจะลบออกจากหน้าของช่างสมชายไปเลย
      setRepairs(repairs.filter(r => r.id !== id));
      alert("ส่งงานคืนระบบให้แอดมินพิจารณาใหม่เรียบร้อยแล้ว");
    }
  };

  // 2. เปิดหน้าต่าง Modal เพื่อพิมพ์รายละเอียดปิดงาน
  const openCloseModal = (id, type) => {
    setCurrentTicket(id);
    setCloseType(type);
    setTechDetail(""); // เคลียร์ข้อความเก่า
    setModalOpen(true);
  };

  // 3. ยืนยันการปิดงาน (บันทึกข้อมูล)
  const handleSubmitClose = (e) => {
    e.preventDefault();
    const newStatus = closeType === "success" ? "เสร็จเรียบร้อย" : "ซ่อมไม่ได้";
    
    setRepairs(repairs.map(r => 
      r.id === currentTicket 
        ? { ...r, status: newStatus, techDetail: techDetail } 
        : r
    ));
    
    setModalOpen(false);
    alert(`บันทึกสถานะ "${newStatus}" เรียบร้อยแล้ว`);
  };

  const handleLogout = () => {
    if (window.confirm("ต้องการออกจากระบบใช่หรือไม่?")) {
      navigate("/");
    }
  };

  return (
    <div className="tech-container">
      <div className="tech-wrapper">
        
        {/* ส่วนหัวของ Tech */}
        <div className="tech-header">
          <div>
            <h2>ระบบจัดการงานซ่อม (สำหรับช่าง)</h2>
            <p>ยินดีต้อนรับ, ช่างสมชาย (แผนกประปา)</p>
          </div>
          <button className="btn-logout" onClick={handleLogout}>ออกจากระบบ</button>
        </div>

        <div className="search-section">
          <input 
            type="text" 
            placeholder="🔍 ค้นหางานซ่อมของคุณ..." 
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
        </div>

        {/* รายการใบงาน */}
        <div className="repair-list">
          {filteredRepairs.length > 0 ? (
            filteredRepairs.map((repair) => (
              <div className="repair-item-card" key={repair.id}>
                
                <div className="card-top">
                  <span className="repair-id">#{repair.id}</span>
                  <span className={`status-badge ${getStatusClass(repair.status)}`}>
                    {repair.status}
                  </span>
                </div>
                
                <h3 className="repair-title">[{repair.type}] {repair.location}</h3>
                <p className="repair-info"><strong>ห้อง/จุดเกิดเหตุ:</strong> {repair.room}</p>
                <p className="repair-info"><strong>อาการที่แจ้ง:</strong> {repair.desc}</p>
                
                {/* ถ้าช่างพิมพ์รายละเอียดการซ่อมไว้ ให้แสดงตรงนี้ */}
                {repair.techDetail && (
                  <div className="tech-note">
                    <strong>บันทึกจากช่าง:</strong> {repair.techDetail}
                  </div>
                )}
                
                <div className="card-actions">
                  <span className="repair-date">📅 {repair.date}</span>
                  
                  {/* แสดงปุ่มเฉพาะงานที่กำลังซ่อม */}
                  {repair.status === "กำลังซ่อม" && (
                    <div className="action-buttons tech-actions">
                      {/* ปุ่มปฏิเสธงาน (ขอบแดง) */}
                      <button className="btn-reject" onClick={() => handleReject(repair.id)}>
                        ปฏิเสธงาน
                      </button>

                      {/* ปุ่มซ่อมไม่ได้ (แดงทึบ) */}
                      <button className="btn-cannot-fix" onClick={() => openCloseModal(repair.id, "fail")}>
                        ซ่อมไม่ได้
                      </button>

                      {/* ปุ่มเสร็จเรียบร้อย (เขียวทึบ) */}
                      <button className="btn-complete" onClick={() => openCloseModal(repair.id, "success")}>
                        ปิดงาน (สำเร็จ)
                      </button>
                    </div>
                  )}
                </div>
              </div>
            ))
          ) : (
            <div className="no-results">
              <p>ไม่มีรายการงานซ่อมในขณะนี้</p>
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
            <p className="modal-subtitle">รหัสใบงาน: #{currentTicket}</p>

            <form onSubmit={handleSubmitClose}>
              <div className="input-group">
                <label>
                  {closeType === "success" ? "รายละเอียดการซ่อมแซม" : "เหตุผลขัดข้อง"} 
                  <span className="required">*</span>
                </label>
                <textarea 
                  rows="4"
                  placeholder={closeType === "success" ? "เช่น เปลี่ยนอะไหล่..., ล้างแอร์..." : "เช่น อะไหล่หมด, เกินขอบเขต, ต้องจ้างช่างภายนอก..."}
                  value={techDetail}
                  onChange={(e) => setTechDetail(e.target.value)}
                  required
                ></textarea>
              </div>

              {/* ในอนาคตสามารถใส่ปุ่มแนบรูปภาพ After ตรงนี้ได้ */}
              
              <div className="modal-actions">
                <button type="button" className="btn-cancel" onClick={() => setModalOpen(false)}>
                  ยกเลิก
                </button>
                <button 
                  type="submit" 
                  className={closeType === "success" ? "btn-submit-success" : "btn-submit-fail"}
                >
                  บันทึกข้อมูล
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