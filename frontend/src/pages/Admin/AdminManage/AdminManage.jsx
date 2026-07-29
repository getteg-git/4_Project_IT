import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import "./AdminManage.css"; 

function AdminManage() {
  const navigate = useNavigate();
  
  const [repairs, setRepairs] = useState([]);
  const [technicians, setTechnicians] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  
  // States สำหรับ Popup
  const [selectedRepair, setSelectedRepair] = useState(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const [isAssignOpen, setIsAssignOpen] = useState(false);
  const [chosenTechId, setChosenTechId] = useState(""); 
  
  // 🌟 เพิ่ม States สำหรับหน้าต่างกรอกเหตุผลไม่อนุมัติ (Reject Modal)
  const [isRejectModalOpen, setIsRejectModalOpen] = useState(false);
  const [rejectReason, setRejectReason] = useState("");
  const [rejectTargetId, setRejectTargetId] = useState(null);

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

  const normalizedSearch = searchTerm.replace(/\s+/g, '').replace(/วิทย์/g, 'วิทยาศาสตร์').toLowerCase();
  
  const filteredRepairs = repairs.filter((repair) => {
    if (!normalizedSearch) return true;
    const loc = (repair.location_name || "").replace(/\s+/g, '').toLowerCase();
    const floor = (repair.floor_name || "").replace(/\s+/g, '').toLowerCase();
    const room = (repair.room_number  || "").replace(/\s+/g, '').toLowerCase();
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

  const openDetailsModal = (repair) => {
    setSelectedRepair(repair);
    setIsDetailsOpen(true);
  };

  const openAssignModal = (repair) => {
    setSelectedRepair(repair);
    setChosenTechId(""); 
    setIsAssignOpen(true);
  };

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
        fetchData();
      } else {
        const errorData = await response.json();
        alert(`❌ ผิดพลาด: ${errorData.error}`);
      }
    } catch (error) {
      console.error("Assign Error:", error);
      alert("❌ ไม่สามารถเชื่อมต่อระบบได้");
    }
  };

  const handleRevoke = async (id) => {
    const confirmRevoke = window.confirm(`⚠️ คุณแน่ใจหรือไม่ที่จะ "ดึงงานกลับ (ยกเลิกการจ่ายงาน)" สำหรับ Ticket: #${id} ?`);
    if (!confirmRevoke) return;

    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${id}/revoke`, {
        method: "PUT"
      });
      if (response.ok) {
        alert(`✅ ดึงงานกลับสำเร็จ สถานะกลับเป็น "รอซ่อม"`);
        fetchData();
      } else {
        alert(`❌ เกิดข้อผิดพลาดในการดึงงานกลับ`);
      }
    } catch (error) {
      console.error("Revoke Error:", error);
    }
  };

  const handleApproveRepair = async (id) => {
    if (!window.confirm("✅ ยืนยัน 'อนุมัติ' ให้ช่างดำเนินการซ่อมใช่หรือไม่? (ระบบจะแจ้งให้ช่างเริ่มงานทันที)")) return;
    
    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${id}/approve`, {
        method: "PUT",
        // บางที Backend อาจจะต้องการ Header เพื่อให้รู้ว่าเป็น Request ที่ถูกต้อง
        headers: { "Content-Type": "application/json" } 
      });
      
      if (response.ok) {
        alert("✅ อนุมัติงานสำเร็จ ช่างสามารถเริ่มงานได้เลย");
        fetchData();
      } else {
        // ดึงข้อความ Error จาก Backend มาแสดง
        const errorData = await response.json();
        alert(`❌ เกิดข้อผิดพลาด: ${errorData.error || errorData.message || 'ไม่สามารถอนุมัติได้'}`);
        console.error("Backend Error Response:", errorData);
      }
    } catch (error) {
      console.error("Approve Error:", error);
      alert("❌ ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้");
    }
  };
  
  // 🌟 ฟังก์ชันเปิดหน้าต่างกรอกเหตุผลแทนการใช้ prompt()
  const openRejectModal = (id) => {
    setRejectTargetId(id);
    setRejectReason("");
    setIsRejectModalOpen(true);
  };

  // 🌟 ฟังก์ชันส่งข้อมูลการไม่อนุมัติ (ซ่อมไม่ได้)
  const submitRejectRepair = async (e) => {
    e.preventDefault();
    if (!rejectReason.trim()) {
      alert("⚠️ กรุณาระบุเหตุผลที่ไม่อนุมัติ");
      return;
    }
  
    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${rejectTargetId}/cancel`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ admin_note: "ไม่อนุมัติการซ่อม: " + rejectReason })
      });
      if (response.ok) {
        alert("❌ ยกเลิกงานสำเร็จ (สถานะเปลี่ยนเป็นซ่อมไม่ได้)");
        setIsRejectModalOpen(false);
        fetchData();
      } else {
        alert("❌ เกิดข้อผิดพลาดในการยกเลิกงาน");
      }
    } catch (error) {
      console.error("Cancel Error:", error);
    }
  };

  return (
    <div className="admin-manage-container">
      <div className="admin-wrapper">
        
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

        <div className="search-section">
          <input 
            type="text" 
            placeholder="🔍 ค้นหารหัสงาน, สถานที่, เลขห้อง หรือชื่อช่าง..." 
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
        </div>

        <div className="repair-list">
          {isLoading ? (
            <div className="loading-state">⏳ กำลังโหลดข้อมูลงานซ่อม...</div>
          ) : filteredRepairs.length > 0 ? (
            filteredRepairs.map((repair) => (
              <div className="repair-item-card" key={repair.id}>
                
                <div className="card-top">
                  <span className="repair-ticket-id">Ticket #{repair.id}</span>
                  <span className={`status-badge ${getStatusClass(repair.status)}`}>{repair.status}</span>
                </div>
                
                <h3 className="repair-title">[{repair.problem_type}] {repair.location_name}</h3>
                
                <div className="repair-details-text">
                  <p className="repair-info">
                    <strong>ชั้น / พิกัด:</strong> {repair.floor_name || "-"} {repair.room_number ? `(ห้อง ${repair.room_number})` : ""}
                  </p>
                  {repair.equipment_name && (
                    <p className="repair-info equipment-info">
                      <strong>⚙️ อุปกรณ์:</strong> {repair.equipment_name}
                    </p>
                  )}
                  <p className="repair-info">
                    <strong>รายละเอียด:</strong> {repair.description}
                  </p>
                </div>
                
                {repair.technician_name && (
                  <div className="repair-tech-box">
                    <strong>👷‍♂️ ผู้รับผิดชอบ:</strong> {repair.technician_name}
                  </div>
                )}
                
                <div className="card-divider"></div>
                
                {/* 🌟 แถวสำหรับปุ่มจัดการทั่วไป */}
                <div className="card-actions-row">
                  <span className="repair-date">🗓️ {formatDate(repair.created_at)}</span>
                  
                  <div className="action-buttons">
                    <button className="btn-details" onClick={() => openDetailsModal(repair)}>👁️ ดูรายละเอียด</button>

                    {repair.status === "รอซ่อม" && !repair.technician_name && (
                      <button className="btn-assign" onClick={() => openAssignModal(repair)}>👉 มอบหมายช่าง</button>
                    )}

                    {repair.status === "กำลังซ่อม" && (
                      <button className="btn-revoke" onClick={() => handleRevoke(repair.id)}>🔄 ดึงงานกลับ</button>
                    )}
                  </div>
                </div>

                {/* 🌟 ดึงกล่องอนุมัติมาอยู่นอก card-actions-row เพื่อให้กาง 100% เต็มบรรทัดด้านล่าง */}
                {repair.status === "รอซ่อม" && repair.technician_name && (
                  repair.admin_note && repair.admin_note.includes("ระบบระงับอัตโนมัติ") ? (
                    <div className="threshold-approval-box">
                      <div>
                        <p className="admin-warning-note">⚠️ {repair.admin_note}</p>
                        <p>💰 ราคาประเมินจากช่าง: <strong>฿{Number(repair.estimated_cost).toLocaleString()}</strong></p>
                      </div>
                      <div className="threshold-actions">
                        <button className="btn-approve" onClick={() => handleApproveRepair(repair.id)}>✅ อนุมัติให้ซ่อม</button>
                        <button className="btn-reject" onClick={() => openRejectModal(repair.id)}>❌ ไม่อนุมัติ</button>
                      </div>
                    </div>
                  ) : (
                    <div style={{ marginTop: '15px', textAlign: 'right' }}>
                      <button className="btn-assign btn-disabled" disabled style={{ backgroundColor: '#9ca3af', cursor: 'not-allowed' }}>
                        ⏳ รอช่างรับงาน
                      </button>
                    </div>
                  )
                )}

              </div>
            ))
          ) : (
            <div className="no-results"><p>❌ ไม่พบรายการที่ตรงกับการค้นหา</p></div>
          )}
        </div>
      </div>

      {/* POPUP: ดูรายละเอียด (Details Modal) */}
      {isDetailsOpen && selectedRepair && (
        <div className="modal-overlay">
          <div className="modal-box details-box">
            {/* โค้ดส่วนนี้เหมือนเดิม... */}
            <span className="close-btn" onClick={() => setIsDetailsOpen(false)}>&times;</span>
            <div className="modal-header-center" style={{ textAlign: 'center', marginBottom: '20px' }}>
              <h3 className="assign-modal-title">📄 ข้อมูลการแจ้งซ่อมอย่างละเอียด</h3>
              <p className="assign-modal-subtitle">Ticket ID: #{selectedRepair.id}</p>
            </div>
            
            <div className="modal-details-content">
              <div className="image-gallery">
                {selectedRepair.images && selectedRepair.images.length > 0 ? (
                  selectedRepair.images.map((img, index) => (
                    <img key={index} src={`http://localhost:8080${img.url}`} alt="รูปปัญหา" className="repair-image"/>
                  ))
                ) : (
                  <div className="no-image-box">ไม่มีรูปภาพประกอบ</div>
                )}
              </div>
              <div className="info-list-container">
                <p><strong>สถานะปัจจุบัน:</strong> <span className={`status-badge ${getStatusClass(selectedRepair.status)}`}>{selectedRepair.status}</span></p>
                <p><strong>หมวดหมู่งาน:</strong> {selectedRepair.problem_type}</p>
                <p><strong>สถานที่:</strong> {selectedRepair.location_name}</p>
                <p><strong>ชั้น / พิกัด:</strong> {selectedRepair.floor_name || "-"} {selectedRepair.room_number ? `(ห้อง ${selectedRepair.room_number})` : ""}</p>
                {selectedRepair.equipment_name && (<p><strong>อุปกรณ์ชำรุด:</strong> {selectedRepair.equipment_name}</p>)}
                <p><strong>อีเมลผู้แจ้ง:</strong> {selectedRepair.reporter_email}</p>
                <p><strong>วันที่แจ้งเรื่อง:</strong> {formatDate(selectedRepair.created_at)}</p>
              </div>
              <div className="issue-desc-box">
                <strong>📝 รายละเอียดปัญหา:</strong>
                <p>{selectedRepair.description}</p>
              </div>
              {(selectedRepair.estimated_cost > 0 || selectedRepair.actual_cost > 0 || selectedRepair.admin_note) && (
                <div className="admin-cost-info-box">
                  <h4>💰 ข้อมูลค่าใช้จ่ายและการประเมิน (Admin Only)</h4>
                  {selectedRepair.estimated_cost > 0 && (<p><strong>ราคาประเมิน:</strong> ฿{Number(selectedRepair.estimated_cost).toLocaleString()}</p>)}
                  {selectedRepair.actual_cost > 0 && (<p><strong>ค่าใช้จ่ายเบิกจริง:</strong> ฿{Number(selectedRepair.actual_cost).toLocaleString()}</p>)}
                  {selectedRepair.admin_note && (<p className="admin-warning-note"><strong>⚠️ หมายเหตุระบบ:</strong> {selectedRepair.admin_note}</p>)}
                </div>
              )}
              {selectedRepair.technician_name && (
                <div className="tech-info-box">
                  <p><strong>👷‍♂️ ช่างผู้รับผิดชอบ:</strong> {selectedRepair.technician_name}</p>
                  {selectedRepair.technician_note && (<p className="tech-note"><strong>💬 หมายเหตุจากช่าง:</strong> {selectedRepair.technician_note}</p>)}
                </div>
              )}
            </div>
            <div className="modal-actions">
              <button className="btn-close-modal" onClick={() => setIsDetailsOpen(false)}>❌ ปิดหน้าต่างนี้</button>
            </div>
          </div>
        </div>
      )}

      {/* POPUP: มอบหมายงาน (Assign Modal) */}
      {isAssignOpen && selectedRepair && (
        <div className="modal-overlay">
          <div className="modal-box">
            <span className="close-btn" onClick={() => setIsAssignOpen(false)}>&times;</span>
            <h3 className="assign-modal-title">👷‍♂️ มอบหมายงานให้ช่าง</h3>
            <p className="assign-modal-subtitle">Ticket: #{selectedRepair.id} | {selectedRepair.problem_type}</p>
            
            <form onSubmit={handleAssignSubmit}>
              <div className="input-group">
                <label>เลือกช่างเทคนิคที่รับผิดชอบ <span className="required">*</span></label>
                <select className="assign-select" value={chosenTechId} onChange={(e) => setChosenTechId(e.target.value)} required>
                  <option value="">-- โปรดเลือกช่างจากรายชื่อ --</option>
                  {technicians.map((tech) => {
                    const specialtiesText = tech.specialty_names && tech.specialty_names.length > 0 
                      ? `[ถนัด: ${tech.specialty_names.join(", ")}]` 
                      : "[ยังไม่ระบุความถนัด]";
                    return (
                      <option key={tech.id} value={tech.id}>{tech.full_name} {specialtiesText}</option>
                    );
                  })}
                </select>
                <small className="assign-modal-hint">💡 ระบบแสดงรายชื่อช่างพร้อมหมวดหมู่งานที่ถนัด เพื่อให้ท่านจ่ายงานได้ตรงสาย</small>
              </div>
              <div className="modal-actions assign-actions">
                <button type="button" className="btn-cancel" onClick={() => setIsAssignOpen(false)}>❌ ยกเลิก</button>
                <button type="submit" className="btn-submit">✅ ยืนยันมอบหมายงาน</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* 🌟 NEW POPUP: ระบุเหตุผลไม่อนุมัติ (Reject Modal) */}
      {isRejectModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <span className="close-btn" onClick={() => setIsRejectModalOpen(false)}>&times;</span>
            <h3 className="assign-modal-title" style={{ color: "#ef4444" }}>❌ ไม่อนุมัติการซ่อม</h3>
            <p className="assign-modal-subtitle">Ticket: #{rejectTargetId}</p>
            
            <form onSubmit={submitRejectRepair}>
              <div className="input-group">
                <label>โปรดระบุเหตุผลที่ไม่อนุมัติ <span className="required">*</span></label>
                <textarea 
                  className="reject-textarea"
                  placeholder="เช่น อุปกรณ์เสื่อมสภาพ ซื้อใหม่คุ้มกว่า..."
                  value={rejectReason}
                  onChange={(e) => setRejectReason(e.target.value)}
                  required
                />
              </div>
              <div className="modal-actions assign-actions">
                <button type="button" className="btn-cancel" onClick={() => setIsRejectModalOpen(false)}>ย้อนกลับ</button>
                <button type="submit" className="btn-submit" style={{ backgroundColor: "#ef4444" }}>ยืนยันการไม่อนุมัติ</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default AdminManage;