import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import "./AdminManage.css"; 

function AdminManage() {
  const navigate = useNavigate();
  
  // States สำหรับเก็บข้อมูลจริงจาก Backend
  const [repairs, setRepairs] = useState([]);
  const [technicians, setTechnicians] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  
  // States สำหรับจัดการ Popup Modals
  const [selectedRepair, setSelectedRepair] = useState(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const [isAssignOpen, setIsAssignOpen] = useState(false);
  const [chosenTechId, setChosenTechId] = useState(""); // เก็บ ID ของช่างที่ถูกเลือก

  // ฟังก์ชันดึงข้อมูลทั้งหมดจาก Backend
  const fetchData = async () => {
    try {
      setIsLoading(true);
      const [repairsRes, usersRes] = await Promise.all([
        fetch("http://localhost:8080/api/repairs"),
        fetch("http://localhost:8080/api/users") 
      ]);

      if (repairsRes.ok) {
        const repairsData = await repairsRes.json();
        setRepairs(repairsData || []);
      }

      if (usersRes.ok) {
        const usersData = await usersRes.json();
        // กรองเอาเฉพาะ User ที่มี Role เป็น technician และเก็บข้อมูลความถนัดมาด้วย
        const techList = usersData.filter(user => user.role === "technician");
        setTechnicians(techList);
      }
    } catch (error) {
      console.error("ดึงข้อมูลไม่สำเร็จ:", error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  // ฟังก์ชัน Smart Search (ค้นหาจาก ชื่อช่างจริง เลขห้อง สถานที่ ฯลฯ)
  const normalizedSearch = searchTerm.replace(/\s+/g, '').replace(/วิทย์/g, 'วิทยาศาสตร์').toLowerCase();
  
  const filteredRepairs = repairs.filter((repair) => {
    if (!normalizedSearch) return true;
    const loc = (repair.location || "").replace(/\s+/g, '').toLowerCase();
    const floor = (repair.floor_name || "").replace(/\s+/g, '').toLowerCase();
    const room = (repair.room || "").replace(/\s+/g, '').toLowerCase();
    const type = (repair.problem_type || "").replace(/\s+/g, '').toLowerCase();
    const status = (repair.status || "").replace(/\s+/g, '').toLowerCase();
    const ticketId = String(repair.id);
    const techName = (repair.technician_name || "").replace(/\s+/g, '').toLowerCase();

    return (
      loc.includes(normalizedSearch) || floor.includes(normalizedSearch) || room.includes(normalizedSearch) ||
      type.includes(normalizedSearch) || status.includes(normalizedSearch) ||
      ticketId.includes(normalizedSearch) || techName.includes(normalizedSearch)
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

  // ==========================================
  // Action Handlers
  // ==========================================
  
  const openDetailsModal = (repair) => {
    setSelectedRepair(repair);
    setIsDetailsOpen(true);
  };

  const openAssignModal = (repair) => {
    setSelectedRepair(repair);
    setChosenTechId(""); 
    setIsAssignOpen(true);
  };

  // ยืนยันมอบหมายงาน (ยิง API PUT /assign)
  const handleAssignSubmit = async (e) => {
    e.preventDefault();
    if (!chosenTechId) {
      alert("⚠️ กรุณาเลือกช่างจากรายชื่อก่อนครับ");
      return;
    }

    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${selectedRepair.id}/assign`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ technician_id: parseInt(chosenTechId) })
      });

      if (response.ok) {
        alert(`✅ มอบหมายงานให้ช่างสำเร็จ!`);
        setIsAssignOpen(false);
        fetchData(); // ดึงข้อมูลใหม่เพื่ออัปเดตหน้าจอ
      } else {
        const errorData = await response.json();
        alert(`❌ ผิดพลาด: ${errorData.error}`);
      }
    } catch (error) {
      console.error("Assign Error:", error);
      alert("❌ ไม่สามารถเชื่อมต่อระบบได้");
    }
  };

  // ดึงงานกลับ (ยิง API PUT /revoke)
  const handleRevoke = async (id) => {
    const confirmRevoke = window.confirm(`⚠️ คุณแน่ใจหรือไม่ที่จะ "ดึงงานกลับ (ยกเลิกการจ่ายงาน)" สำหรับ Ticket: #${id} ?`);
    if (!confirmRevoke) return;

    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${id}/revoke`, {
        method: "PUT"
      });

      if (response.ok) {
        alert(`✅ ดึงงานกลับสำเร็จ สถานะกลับเป็น "รอซ่อม"`);
        fetchData(); // ดึงข้อมูลใหม่
      } else {
        alert(`❌ เกิดข้อผิดพลาดในการดึงงานกลับ`);
      }
    } catch (error) {
      console.error("Revoke Error:", error);
    }
  };

  return (
    <div className="admin-manage-container">
      <div className="admin-wrapper">
        
        {/* Header และ Navigation */}
        <div className="admin-header">
          <div>
            <h2>🛠️ จัดการและมอบหมายงานซ่อม</h2>
            <p>ค้นหาและจ่ายงานให้ช่างเทคนิค</p>
          </div>
          <div className="header-actions">
            <button className="btn-back-dashboard" onClick={() => navigate("/admin/home")}>
              📊 กลับหน้า Dashboard
            </button>
          </div>
        </div>

        {/* ช่องค้นหา */}
        <div className="search-section">
          <input 
            type="text" 
            placeholder="🔍 ค้นหารหัสงาน, สถานที่, เลขห้อง หรือชื่อช่าง..." 
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
        </div>

        {/* รายการแจ้งซ่อม */}
        <div className="repair-list">
          {isLoading ? (
            <div className="loading-state">⏳ กำลังโหลดข้อมูลงานซ่อม...</div>
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
                
                {/* 🔥 เพิ่มการแสดงผลเลขห้อง/พิกัด */}
                <p className="repair-info">
                  <strong>ชั้น / พิกัด:</strong> {repair.floor_name || "-"} {repair.room ? `(ห้อง ${repair.room})` : ""}
                </p>
                <p className="repair-info"><strong>รายละเอียด:</strong> {repair.description}</p>
                
                {/* แสดงชื่อจริงของช่างถ้ามีการมอบหมายงานแล้ว */}
                {repair.technician_name && (
                  <div className="repair-tech-box">
                    <strong>👷‍♂️ ผู้รับผิดชอบ:</strong> {repair.technician_name}
                  </div>
                )}
                
                <div className="card-actions-row">
                  <span className="repair-date">📅 แจ้งเมื่อ: {formatDate(repair.created_at)}</span>
                  
                  <div className="action-buttons">
                    <button className="btn-details" onClick={() => openDetailsModal(repair)}>
                      👁️ ดูรายละเอียด
                    </button>

                    {repair.status === "รอซ่อม" && (
                      <button className="btn-assign" onClick={() => openAssignModal(repair)}>
                        👉 มอบหมายช่าง
                      </button>
                    )}

                    {repair.status === "กำลังซ่อม" && (
                      <button className="btn-revoke" onClick={() => handleRevoke(repair.id)}>
                        🔄 ดึงงานกลับ
                      </button>
                    )}
                  </div>
                </div>
              </div>
            ))
          ) : (
            <div className="no-results">
              <p>❌ ไม่พบรายการที่ตรงกับการค้นหา</p>
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
              
              {/* โชว์รูปภาพ */}
              <div className="image-gallery">
                {selectedRepair.images && selectedRepair.images.length > 0 ? (
                  selectedRepair.images.map((img, index) => (
                    <img key={index} src={`http://localhost:8080${img.url}`} alt="รูปปัญหา" className="repair-image"/>
                  ))
                ) : (
                  <div className="no-image-box">ไม่มีรูปภาพประกอบ</div>
                )}
              </div>

              <div className="info-grid">
                <p><strong>ประเภทปัญหา:</strong> {selectedRepair.problem_type}</p>
                {/* 🔥 เพิ่มการโชว์เลขห้องในหน้าดูรายละเอียด */}
                <p>
                  <strong>สถานที่:</strong> {selectedRepair.location} 
                  {selectedRepair.floor_name ? ` (${selectedRepair.floor_name})` : ""}
                  {selectedRepair.room ? ` ห้อง ${selectedRepair.room}` : ""}
                </p>
                <p><strong>วันที่แจ้งเรื่อง:</strong> {formatDate(selectedRepair.created_at)}</p>
                <p><strong>อีเมลผู้แจ้ง:</strong> {selectedRepair.reporter_email}</p>
                <p><strong>สถานะปัจจุบัน:</strong> <span className={`status-badge ${getStatusClass(selectedRepair.status)}`}>{selectedRepair.status}</span></p>
              </div>

              <div className="issue-desc-box">
                <strong>📝 รายละเอียดเพิ่มเติม:</strong>
                <p>{selectedRepair.description}</p>
              </div>

              {selectedRepair.technician_name && (
                <div className="tech-info-box">
                  <p><strong>👷‍♂️ ช่างผู้รับผิดชอบ:</strong> {selectedRepair.technician_name}</p>
                  {selectedRepair.technician_note && (
                    <p className="tech-note"><strong>💬 หมายเหตุจากช่าง:</strong> {selectedRepair.technician_note}</p>
                  )}
                </div>
              )}

            </div>
            <div className="modal-actions">
              <button className="btn-close-modal" onClick={() => setIsDetailsOpen(false)}>❌ ปิดหน้าต่าง</button>
            </div>
          </div>
        </div>
      )}

      {/* ==========================================
          2. POPUP: มอบหมายงาน (Assign Modal) พร้อมแสดงความถนัดช่าง
          ========================================== */}
      {isAssignOpen && selectedRepair && (
        <div className="modal-overlay">
          <div className="modal-box">
            <span className="close-btn" onClick={() => setIsAssignOpen(false)}>&times;</span>
            <h3 style={{ color: "#007A53" }}>👷‍♂️ มอบหมายงานให้ช่าง</h3>
            <p style={{ fontSize: "14px", color: "#666", marginBottom: "20px" }}>
              Ticket: #{selectedRepair.id} | {selectedRepair.problem_type}
            </p>
            
            <form onSubmit={handleAssignSubmit}>
              <div className="input-group">
                <label>เลือกช่างเทคนิคที่รับผิดชอบ <span className="required">*</span></label>
                <select 
                  value={chosenTechId} 
                  onChange={(e) => setChosenTechId(e.target.value)}
                  required
                  style={{ height: "auto", minHeight: "50px" }} // ขยายกล่องนิดนึงเพราะชื่อยาว
                >
                  <option value="">-- โปรดเลือกช่างจากรายชื่อ --</option>
                  {technicians.map((tech) => {
                    // จัดรูปแบบการแสดงผล: ช่างสมชาย (ความถนัด: ประปา, ไฟฟ้า)
                    const specialtiesText = tech.specialty_names && tech.specialty_names.length > 0 
                      ? `[ถนัด: ${tech.specialty_names.join(", ")}]` 
                      : "[ยังไม่ระบุความถนัด]";
                      
                    return (
                      <option key={tech.id} value={tech.id}>
                        {tech.full_name} {specialtiesText}
                      </option>
                    );
                  })}
                </select>
                <small style={{ display: "block", marginTop: "8px", color: "#666" }}>
                  💡 ระบบแสดงรายชื่อช่างพร้อมหมวดหมู่งานที่ถนัด เพื่อให้ท่านจ่ายงานได้ตรงสาย
                </small>
              </div>

              <div className="modal-actions" style={{ marginTop: "30px" }}>
                <button type="button" className="btn-cancel" onClick={() => setIsAssignOpen(false)}>
                  ❌ ยกเลิก
                </button>
                <button type="submit" className="btn-submit">
                  ✅ ยืนยันมอบหมายงาน
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default AdminManage;