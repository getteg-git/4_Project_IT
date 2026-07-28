import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import "./MyRepairs.css";

function MyRepairs() {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const [repairs, setRepairs] = useState([]);
  const [isLoading, setIsLoading] = useState(true);

  // States สำหรับ Popup
  const [selectedRepair, setSelectedRepair] = useState(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);

  // ดึงข้อมูลทั้งหมดจาก Backend
  useEffect(() => {
    const fetchRepairs = async () => {
      try {
        const response = await fetch("http://localhost:8080/api/repairs");
        if (response.ok) {
          const data = await response.json();
          setRepairs(data || []);
        } else {
          console.error("ดึงข้อมูลไม่สำเร็จ");
        }
      } catch (error) {
        console.error("Error fetching repairs:", error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchRepairs();
  }, []);

  // ฟังก์ชัน Smart Search 
  const normalizedSearch = searchTerm.replace(/\s+/g, '').replace(/วิทย์/g, 'วิทยาศาสตร์').toLowerCase();

  const filteredRepairs = repairs.filter((repair) => {
    if (!normalizedSearch) return true;

    // 🔥 ปรับแก้ฟิลด์ให้ตรงกับ JSON จาก Backend และเพิ่มการค้นหา Other Location/Problem, อุปกรณ์
    const loc = (repair.location_name || "").replace(/\s+/g, '').toLowerCase();
    const otherLoc = (repair.other_location || "").replace(/\s+/g, '').toLowerCase();
    const floor = (repair.floor_name || "").replace(/\s+/g, '').toLowerCase();
    const room = (repair.room_number || "").replace(/\s+/g, '').toLowerCase(); 
    const equip = (repair.equipment_name || "").replace(/\s+/g, '').toLowerCase(); 
    const type = (repair.problem_type || "").replace(/\s+/g, '').toLowerCase();
    const otherType = (repair.other_problem_type || "").replace(/\s+/g, '').toLowerCase();
    const desc = (repair.description || "").replace(/\s+/g, '').toLowerCase();
    const status = (repair.status || "").replace(/\s+/g, '').toLowerCase();
    const ticketId = String(repair.id);
    const techName = (repair.technician_name || "").replace(/\s+/g, '').toLowerCase(); 

    return (
      loc.includes(normalizedSearch) ||
      otherLoc.includes(normalizedSearch) ||
      floor.includes(normalizedSearch) ||
      room.includes(normalizedSearch) ||
      equip.includes(normalizedSearch) ||
      type.includes(normalizedSearch) ||
      otherType.includes(normalizedSearch) ||
      desc.includes(normalizedSearch) ||
      status.includes(normalizedSearch) ||
      ticketId.includes(normalizedSearch) ||
      techName.includes(normalizedSearch)
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

  // ฟังก์ชันแปลงวันที่ให้อ่านง่าย
  const formatDate = (dateString) => {
    const options = { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' };
    return new Date(dateString).toLocaleDateString('th-TH', options);
  };

  // ฟังก์ชันเปิด-ปิด Popup
  const openDetailsModal = (repair) => {
    setSelectedRepair(repair);
    setIsDetailsOpen(true);
  };

  const closeDetailsModal = () => {
    setIsDetailsOpen(false);
    setSelectedRepair(null);
  };

  return (
    <div className="tracking-container">
      <div className="tracking-wrapper">
        <div className="tracking-header">
          <h2>ติดตามสถานะการแจ้งซ่อมทั้งหมด</h2>
          <p>ระบบแสดงรายการแจ้งปัญหาทั้งหมดภายในคณะวิทยาศาสตร์</p>
        </div>

        <div className="search-section">
          <input 
            type="text" 
            placeholder="🔍 พิมพ์เพื่อค้นหา (เช่น ห้อง 1102, งานไฟฟ้า, รอซ่อม, รหัส #123)..." 
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
        </div>

        <div className="repair-list">
          {isLoading ? (
            <div className="loading-state">⏳ กำลังโหลดข้อมูล...</div>
          ) : filteredRepairs.length > 0 ? (
            filteredRepairs.map((repair) => {
              // 🔥 ตรรกะแสดงผล Location และ Problem Type (ถ้าระบุ 'อื่นๆ' มา ให้แสดงค่า Other แทน)
              const displayLocation = repair.other_location || repair.location_name;
              const displayProblemType = repair.other_problem_type || repair.problem_type;

              return (
                <div className="repair-item-card" key={repair.id}>
                  <div className="card-top">
                    <span className="repair-id">Ticket #{repair.id}</span>
                    <span className={`status-badge ${getStatusClass(repair.status)}`}>
                      {repair.status}
                    </span>
                  </div>
                  
                  <h3 className="repair-title">[{displayProblemType}] {displayLocation}</h3>
                  
                  <p className="repair-room">
                    <strong>ชั้น / พิกัด:</strong> {repair.floor_name || "-"} {repair.room_number ? `(${repair.room_number})` : ""}
                  </p>

                  {/* 🔥 แสดงอุปกรณ์ชำรุด (ถ้ามี) */}
                  {repair.equipment_name && (
                    <p className="repair-equipment" style={{ margin: "4px 0", fontSize: "0.9rem", color: "#e67e22" }}>
                      <strong>⚙️ อุปกรณ์:</strong> {repair.equipment_name} {repair.asset_code ? `(${repair.asset_code})` : ""}
                    </p>
                  )}
                  
                  <p className="repair-desc"><strong>รายละเอียด:</strong> {repair.description}</p>
                  
                  {repair.technician_name && (
                    <p className="repair-tech-badge"><strong>👷‍♂️ ช่าง:</strong> {repair.technician_name}</p>
                  )}

                  <div className="card-bottom">
                    <span className="repair-date">📅 {formatDate(repair.created_at)}</span>
                    <button 
                      className="btn-details"
                      onClick={() => openDetailsModal(repair)}
                    >
                      👁️ กดดูรายละเอียดงานนี้
                    </button>
                  </div>
                </div>
              );
            })
          ) : (
            <div className="no-results">
              <p>❌ ไม่พบรายการแจ้งซ่อมที่ตรงกับคำค้นหาของคุณ</p>
            </div>
          )}
        </div>

        <div className="back-action">
          <button className="btn-back" onClick={() => navigate("/")}>⬅️ กลับสู่หน้าหลัก</button>
        </div>
      </div>

      {/* ==========================================
          POPUP: ดูรายละเอียดงาน (Modal)
          ========================================== */}
      {isDetailsOpen && selectedRepair && (
        <div className="modal-overlay">
          <div className="modal-box details-box">
            <span className="close-btn" onClick={closeDetailsModal}>&times;</span>
            <h3>📄 ข้อมูลการแจ้งซ่อมอย่างละเอียด</h3>
            <p className="ticket-subtitle">Ticket ID: #{selectedRepair.id}</p>
            <hr />
            
            <div className="modal-details-content">
              <div className="image-gallery">
                {selectedRepair.images && selectedRepair.images.length > 0 ? (
                  selectedRepair.images.map((img, index) => (
                    <div key={index} className="image-container" style={{ position: 'relative', display: 'inline-block' }}>
                      <img 
                        src={`http://localhost:8080${img.url}`} 
                        alt="รูปปัญหาหน้างาน" 
                        className="repair-image"
                      />
                      {/* 🔥 เพิ่ม Label เพื่อแยกรูปก่อน-หลังซ่อม */}
                      <span style={{ position: 'absolute', top: 5, left: 5, background: 'rgba(0,0,0,0.6)', color: '#fff', padding: '2px 6px', borderRadius: '4px', fontSize: '0.8rem' }}>
                        {img.type === 'after' ? '📸 หลังซ่อม' : '📸 ก่อนซ่อม'}
                      </span>
                    </div>
                  ))
                ) : (
                  <div className="no-image-box">ไม่มีรูปภาพประกอบ</div>
                )}
              </div>

              <div className="info-grid">
                <p><strong>สถานะปัจจุบัน:</strong> <span className={`status-badge ${getStatusClass(selectedRepair.status)}`}>{selectedRepair.status}</span></p>
                <p><strong>หมวดหมู่งาน:</strong> {selectedRepair.other_problem_type || selectedRepair.problem_type}</p>
                <p><strong>สถานที่:</strong> {selectedRepair.other_location || selectedRepair.location_name}</p>
                <p>
                  <strong>ชั้น / พิกัด:</strong> {selectedRepair.floor_name || "-"} {selectedRepair.room_number ? `(${selectedRepair.room_number})` : ""}
                </p>
                {selectedRepair.equipment_name && (
                  <p>
                    <strong>อุปกรณ์ชำรุด:</strong> {selectedRepair.equipment_name} {selectedRepair.asset_code ? `(${selectedRepair.asset_code})` : ""}
                  </p>
                )}
                <p><strong>อีเมลผู้แจ้ง:</strong> {selectedRepair.reporter_email}</p>
                <p><strong>วันที่แจ้งเรื่อง:</strong> {formatDate(selectedRepair.created_at)}</p>
              </div>
              
              <div className="issue-desc-box">
                <strong>📝 รายละเอียดปัญหา:</strong>
                <p>{selectedRepair.description}</p>
              </div>
              
              {selectedRepair.status !== "รอซ่อม" && (
                <div className="tech-info-box">
                  <p><strong>👷‍♂️ ช่างผู้รับผิดชอบ:</strong> {selectedRepair.technician_name || "อยู่ระหว่างดำเนินการ"}</p>
                  {selectedRepair.technician_note && (
                    <p className="tech-note"><strong>💬 หมายเหตุจากช่าง:</strong> {selectedRepair.technician_note}</p>
                  )}
                </div>
              )}
            </div>

            <div className="modal-actions">
              <button className="btn-close-modal" onClick={closeDetailsModal}>❌ ปิดหน้าต่างนี้</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default MyRepairs;