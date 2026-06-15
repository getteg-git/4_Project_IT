import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import "./AdminHome.css";

function AdminHome() {
  const navigate = useNavigate();

  // ข้อมูลจำลอง (Mock Data)
  const [repairs, setRepairs] = useState([
    { id: "REP-001", date: "10 มิ.ย. 2569", location: "อาคารวิทยาศาสตร์ 1", room: "ห้อง 1102", type: "งานไฟฟ้า", status: "รอซ่อม", desc: "หลอดไฟหน้าห้องขาด", tech: null },
    { id: "REP-002", date: "09 มิ.ย. 2569", location: "ห้องปฏิบัติการเคมี", room: "Lab 3", type: "งานประปา", status: "กำลังซ่อม", tech: "ช่างสมชาย" },
    { id: "REP-003", date: "08 มิ.ย. 2569", location: "อาคารวิทยาศาสตร์ 2", room: "ห้อง 2205", type: "งานอิเล็กทรอนิกส์", status: "ซ่อมไม่ได้", desc: "แอร์ไม่เย็น มีน้ำหยด", tech: "ช่างสมเกียรติ" },
    { id: "REP-004", date: "05 มิ.ย. 2569", location: "อาคารวิทยาศาสตร์ 1", room: "ห้องน้ำชาย ชั้น 1", type: "งานประปา", status: "เสร็จเรียบร้อย", tech: "ช่างสมชาย" },
  ]);

  // รายชื่อช่างสำหรับดึงใส่ Dropdown
  const technicians = ["ช่างสมชาย", "ช่างสมศักดิ์", "ช่างสมเกียรติ"];

  const [searchTerm, setSearchTerm] = useState("");
  
  // 🔥 States สำหรับจัดการ Popup Modals
  const [selectedRepair, setSelectedRepair] = useState(null); // เก็บข้อมูลงานที่กำลังกดดู/มอบหมาย
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);   // เปิด-ปิด Popup รายละเอียด
  const [isAssignOpen, setIsAssignOpen] = useState(false);     // เปิด-ปิด Popup มอบหมายงาน
  const [chosenTech, setChosenTech] = useState("");            // เก็บช่างที่เลือกจาก Dropdown

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

  // เปิด Popup รายละเอียด
  const openDetailsModal = (repair) => {
    setSelectedRepair(repair);
    setIsDetailsOpen(true);
  };

  // เปิด Popup มอบหมายงาน
  const openAssignModal = (repair) => {
    setSelectedRepair(repair);
    setChosenTech(""); // รีเซ็ตค่าช่างที่เลือกเป็นค่าว่างก่อน
    setIsAssignOpen(true);
  };

  // ยืนยันมอบหมายงานจาก Dropdown
  const handleAssignSubmit = (e) => {
    e.preventDefault();
    if (!chosenTech) {
      alert("กรุณาเลือกช่างก่อนครับ");
      return;
    }

    setRepairs(repairs.map(r => 
      r.id === selectedRepair.id ? { ...r, status: "กำลังซ่อม", tech: chosenTech } : r
    ));

    setIsAssignOpen(false);
    alert(`✅ มอบหมายงาน #${selectedRepair.id} ให้ ${chosenTech} สำเร็จ`);
  };

  const handleRevoke = (id) => {
    const confirmRevoke = window.confirm(`คุณแน่ใจหรือไม่ที่จะ "ดึงงานกลับ/ยกเลิกการจ่ายงาน" สำหรับ Ticket: #${id} ?`);
    if (confirmRevoke) {
      setRepairs(repairs.map(r => r.id === id ? { ...r, status: "รอซ่อม", tech: null } : r));
    }
  };

  const handleLogout = () => {
    if (window.confirm("ต้องการออกจากระบบใช่หรือไม่?")) {
      navigate("/");
    }
  };

  return (
    <div className="admin-container">
      <div className="admin-wrapper">
        <div className="admin-header">
          <div>
            <h2>แผงควบคุมผู้ดูแลระบบ (Admin Panel)</h2>
            <p>จัดการและมอบหมายงานแจ้งซ่อม คณะวิทยาศาสตร์</p>
          </div>
          <button className="btn-logout" onClick={handleLogout}>ออกจากระบบ</button>
        </div>

        <div className="search-section">
          <input 
            type="text" 
            placeholder="🔍 ค้นหารหัสงาน, สถานที่, หมายเลขห้อง หรือสถานะ..." 
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
        </div>

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
                <p className="repair-info"><strong>รายละเอียด:</strong> {repair.desc}</p>
                
                {repair.tech && (
                  <p className="repair-tech"><strong>ผู้รับผิดชอบ:</strong> 👷‍♂️ {repair.tech}</p>
                )}
                
                <div className="card-actions">
                  <span className="repair-date">📅 {repair.date}</span>
                  
                  <div className="action-buttons">
                    <button className="btn-details" onClick={() => openDetailsModal(repair)}>
                      ดูรายละเอียด
                    </button>

                    {repair.status === "รอซ่อม" && (
                      <button className="btn-assign" onClick={() => openAssignModal(repair)}>
                        มอบหมายงาน
                      </button>
                    )}

                    {repair.status === "กำลังซ่อม" && (
                      <button className="btn-revoke" onClick={() => handleRevoke(repair.id)}>
                        ยกเลิกงาน
                      </button>
                    )}
                  </div>
                </div>
              </div>
            ))
          ) : (
            <div className="no-results">
              <p>ไม่พบรายการที่ค้นหา</p>
            </div>
          )}
        </div>
      </div>

      {/* ==========================================
          1. POPUP: ดูรายละเอียด (Details Modal)
          ========================================== */}
      {isDetailsOpen && selectedRepair && (
        <div className="modal-overlay">
          <div className="modal-box details-box">
            <span className="close-btn" onClick={() => setIsDetailsOpen(false)}>&times;</span>
            <h3>📄 รายละเอียดงานซ่อม #{selectedRepair.id}</h3>
            <hr />
            <div className="modal-details-content">
              <p><strong>ประเภทปัญหา:</strong> {selectedRepair.type}</p>
              <p><strong>สถานที่:</strong> {selectedRepair.location}</p>
              <p><strong>ห้อง/จุดเกิดเหตุ:</strong> {selectedRepair.room}</p>
              <p><strong>รายละเอียดเพิ่มเติม:</strong> {selectedRepair.desc}</p>
              <p><strong>วันที่แจ้งเรื่อง:</strong> {selectedRepair.date}</p>
              <p><strong>สถานะปัจจุบัน:</strong> <span className={`status-badge ${getStatusClass(selectedRepair.status)}`}>{selectedRepair.status}</span></p>
              <p><strong>ช่างที่รับผิดชอบ:</strong> {selectedRepair.tech ? `👷‍♂️ ${selectedRepair.tech}` : "ยังไม่มอบหมายงาน"}</p>
            </div>
            <div className="modal-actions">
              <button className="btn-cancel" onClick={() => setIsDetailsOpen(false)}>ปิดหน้าต่าง</button>
            </div>
          </div>
        </div>
      )}

      {/* ==========================================
          2. POPUP: มอบหมายงานพร้อม Dropdown ช่าง (Assign Modal)
          ========================================== */}
      {isAssignOpen && selectedRepair && (
        <div className="modal-overlay">
          <div className="modal-box">
            <span className="close-btn" onClick={() => setIsAssignOpen(false)}>&times;</span>
            <h3>👷‍♂️ มอบหมายงานซ่อม #{selectedRepair.id}</h3>
            <p style={{ fontSize: "14px", color: "#666", marginBottom: "15px" }}>
              งาน: [{selectedRepair.type}] {selectedRepair.location}
            </p>
            
            <form onSubmit={handleAssignSubmit}>
              <div className="input-group">
                <label style={{ fontWeight: "bold", display: "block", marginBottom: "8px" }}>
                  เลือกช่างเทคนิคที่รับผิดชอบ <span style={{ color: "red" }}>*</span>
                </label>
                <select 
                  value={chosenTech} 
                  onChange={(e) => setChosenTech(e.target.value)}
                  required
                  style={{ width: "100%", padding: "10px", borderRadius: "5px", border: "1px solid #ccc" }}
                >
                  <option value="">-- โปรดเลือกช่างซ่อม --</option>
                  {technicians.map((tech, idx) => (
                    <option key={idx} value={tech}>{tech}</option>
                  ))}
                </select>
              </div>

              <div className="modal-actions" style={{ marginTop: "20px" }}>
                <button type="button" className="btn-cancel" onClick={() => setIsAssignOpen(false)}>
                  ยกเลิก
                </button>
                <button type="submit" className="btn-submit">
                  ยืนยันมอบหมายงาน
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default AdminHome;