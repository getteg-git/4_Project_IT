import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import "./TechHome.css";

function TechHome() {
  const navigate = useNavigate();

  const [currentUser, setCurrentUser] = useState(null);
  const [repairs, setRepairs] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");

  const [selectedRepair, setSelectedRepair] = useState(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);

  // States สำหรับ Modal ปิดงาน (ซ่อมสำเร็จ/ซ่อมไม่ได้)
  const [modalOpen, setModalOpen] = useState(false);
  const [currentTicket, setCurrentTicket] = useState(null);
  const [closeType, setCloseType] = useState(""); 
  const [techDetail, setTechDetail] = useState("");
  const [actualCost, setActualCost] = useState(""); 
  const [isSubmitting, setIsSubmitting] = useState(false);

  // 🔥 States สำหรับ Modal "รับงานและประเมินราคา" (Estimate Modal)
  const [estimateModalOpen, setEstimateModalOpen] = useState(false);
  const [estimateTicketId, setEstimateTicketId] = useState(null);
  const [estimatedCost, setEstimatedCost] = useState("");

  const fetchMyJobs = async (user) => {
    try {
      setIsLoading(true);
      const response = await fetch("http://localhost:8080/api/repairs");
      if (response.ok) {
        const allRepairs = await response.json();
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
      navigate("/");
    }
  }, [navigate]);

  const normalizedSearch = searchTerm.replace(/\s+/g, '').replace(/วิทย์/g, 'วิทยาศาสตร์').toLowerCase();
  
  const filteredRepairs = repairs.filter((repair) => {
    if (!normalizedSearch) return true;
    const loc = (repair.location_name || "").replace(/\s+/g, '').toLowerCase();
    const floor = (repair.floor_name || "").replace(/\s+/g, '').toLowerCase();
    const room = (repair.room_number || "").replace(/\s+/g, '').toLowerCase(); 
    const equip = (repair.equipment_name || "").replace(/\s+/g, '').toLowerCase();
    const status = (repair.status || "").replace(/\s+/g, '').toLowerCase();
    const ticketId = String(repair.id);
    
    return (
      loc.includes(normalizedSearch) || floor.includes(normalizedSearch) || room.includes(normalizedSearch) ||
      equip.includes(normalizedSearch) || status.includes(normalizedSearch) || ticketId.includes(normalizedSearch)
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

  const handleReject = async (id) => {
    const confirmReject = window.confirm(`⚠️ คุณต้องการ "ปฏิเสธงาน" Ticket #${id} และส่งคืนให้แอดมิน ใช่หรือไม่?`);
    if (!confirmReject) return;

    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${id}/reject`, {
        method: "PUT"
      });

      if (response.ok) {
        alert("✅ ส่งงานคืนระบบให้แอดมินพิจารณาใหม่เรียบร้อยแล้ว");
        fetchMyJobs(currentUser); 
      } else {
        alert("❌ เกิดข้อผิดพลาด ไม่สามารถปฏิเสธงานได้");
      }
    } catch (error) {
      console.error("Reject Error:", error);
    }
  };

  // -------------------------------------------------------------
  // 🔥 1. เปิดหน้าต่างประเมินราคาก่อนรับงาน (เช็ก Break-even)
  // -------------------------------------------------------------
  const openEstimateModal = (id) => {
    setEstimateTicketId(id);
    setEstimatedCost("");
    setEstimateModalOpen(true);
  };

  const handleSubmitEstimate = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);

    // ✅ [แก้บั๊กเลข 0] แปลงค่าเป็น Number ให้ชัวร์ และรับเลข 0 โดดๆ ได้
    const costValue = estimatedCost === "" ? 0 : Number(estimatedCost);

    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${estimateTicketId}/estimate`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ estimated_cost: costValue })
      });

      if (response.ok) {
        const data = await response.json();
        
        // ถ้าระบบตีกลับสถานะเป็นรอซ่อม แสดงว่าเกินจุดคุ้มทุน
        if (data.status === "รอซ่อม") {
          alert(`⚠️ ${data.message}\n(งานนี้ถูกส่งเรื่องกลับไปที่ Admin แล้ว)`);
        } else {
          alert("✅ ประเมินราคาผ่านเกณฑ์! ระบบเริ่มนับเวลาดำเนินการซ่อมแล้ว");
        }

        setEstimateModalOpen(false);
        fetchMyJobs(currentUser); 
      } else {
        const errorData = await response.json();
        alert(`❌ เกิดข้อผิดพลาด: ${errorData.error}`);
      }
    } catch (error) {
      console.error("Estimate Error:", error);
      alert("❌ ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้");
    } finally {
      setIsSubmitting(false);
    }
  };

  // -------------------------------------------------------------
  // 2. ปิดงานซ่อม (ใช้ FormData ส่งเข้า API /status)
  // -------------------------------------------------------------
  const openCloseModal = (id, type) => {
    setCurrentTicket(id);
    setCloseType(type);
    setTechDetail("");
    setActualCost(""); 
    setModalOpen(true);
  };

  const handleSubmitClose = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    
    const newStatus = closeType === "success" ? "เสร็จเรียบร้อย" : "ซ่อมไม่ได้";
    const costValue = closeType === "success" && actualCost ? actualCost : "0";
    
    // สร้าง FormData ให้ตรงกับที่ Backend (repair.go) รอรับ
    const formData = new FormData();
    formData.append("status", newStatus);
    formData.append("technician_note", techDetail);
    formData.append("actual_cost", costValue);

    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${currentTicket}/status`, {
        method: "PUT",
        body: formData 
      });

      if (response.ok) {
        alert(`✅ บันทึกสถานะ "${newStatus}" เรียบร้อยแล้ว`);
        setModalOpen(false);
        fetchMyJobs(currentUser); 
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
    <div className="admin-manage-container"> 
      <div className="admin-wrapper">
        
        <div className="admin-header">
          <div>
            <h2>🛠️ ระบบจัดการงานซ่อม (Technician Panel)</h2>
            <p>ยินดีต้อนรับ, {currentUser ? (currentUser.full_name || currentUser.username) : "กำลังโหลด..."}</p>
          </div>
          <div className="header-actions">
            <button className="btn-logout" onClick={handleLogout}>
              🚪 ออกจากระบบ
            </button>
          </div>
        </div>

        <div className="search-section">
          <input 
            type="text" 
            placeholder="🔍 ค้นหารหัสงาน, สถานที่, เลขห้อง, อุปกรณ์ หรือสถานะ..." 
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
        </div>

        <div className="repair-list">
          {isLoading ? (
            <div className="loading-state">⏳ กำลังโหลดรายการงานของคุณ...</div>
          ) : filteredRepairs.length > 0 ? (
            filteredRepairs.map((repair) => (
              <div className="repair-item-card" key={repair.id}>
                
                <div className="card-top">
                  <span className="repair-ticket-id">Ticket #{repair.id}</span>
                  <span className={`status-badge ${getStatusClass(repair.status)}`}>
                    {repair.status}
                  </span>
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
                    <strong>อาการที่แจ้ง:</strong> {repair.description}
                  </p>
                </div>
                
                <div className="card-divider"></div>
                
                <div className="card-actions-row">
                  <span className="repair-date">📅 แจ้งเมื่อ: {formatDate(repair.created_at)}</span>
                  
                  <div className="action-buttons">
                    <button className="btn-details" onClick={() => openDetailsModal(repair)}>
                      👁️ ดูรายละเอียด
                    </button>

                    {repair.status === "รอซ่อม" && (
                      <>
                        {/* ✅ [อัปเดต] เช็กว่ามี admin_note หรือไม่ ถ้ามีให้แสดงปุ่ม รออนุมัติ แบบกดไม่ได้ */}
                        {repair.admin_note && repair.admin_note.trim() !== "" ? (
                          <button className="btn-waiting-approval" disabled>
                            ⏳ รออนุมัติ
                          </button>
                        ) : (
                          <button className="btn-accept" onClick={() => openEstimateModal(repair.id)}>
                            👍 รับงาน
                          </button>
                        )}

                        <button className="btn-reject" onClick={() => handleReject(repair.id)}>
                          ↩️ ปฏิเสธงาน
                        </button>
                      </>
                    )}

                    {repair.status === "กำลังซ่อม" && (
                      <>
                        <button className="btn-cannot-fix" onClick={() => openCloseModal(repair.id, "fail")}>
                          ❌ ซ่อมไม่ได้
                        </button>
                        <button className="btn-complete" onClick={() => openCloseModal(repair.id, "success")}>
                          ✅ ซ่อมสำเร็จ
                        </button>
                      </>
                    )}
                  </div>
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
          POPUP 1: ดูรายละเอียดงาน
          ========================================== */}
      {isDetailsOpen && selectedRepair && (
        <div className="modal-overlay">
          <div className="modal-box details-box">
            <span className="close-btn" onClick={() => setIsDetailsOpen(false)}>&times;</span>
            
            <div className="modal-header-center">
              <h3 className="modal-title-primary">📄 ข้อมูลการแจ้งซ่อมอย่างละเอียด</h3>
              <p className="modal-subtitle">Ticket ID: #{selectedRepair.id}</p>
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
                <p className="repair-info-item">
                  <strong>สถานะปัจจุบัน:</strong> <span className={`status-badge ${getStatusClass(selectedRepair.status)}`}>{selectedRepair.status}</span>
                </p>
                <p className="repair-info-item"><strong>หมวดหมู่งาน:</strong> {selectedRepair.problem_type}</p>
                <p className="repair-info-item"><strong>สถานที่:</strong> {selectedRepair.location_name}</p>
                <p className="repair-info-item">
                  <strong>ชั้น / พิกัด:</strong> {selectedRepair.floor_name || "-"} {selectedRepair.room_number ? `(ห้อง ${selectedRepair.room_number})` : ""}
                </p>
                
                {selectedRepair.equipment_name && (
                  <p className="repair-info-item">
                    <strong>อุปกรณ์ชำรุด:</strong> {selectedRepair.equipment_name}
                  </p>
                )}
                
                <p className="repair-info-item"><strong>อีเมลผู้แจ้ง:</strong> {selectedRepair.reporter_email}</p>
                <p className="repair-info-item"><strong>วันที่แจ้งเรื่อง:</strong> {formatDate(selectedRepair.created_at)}</p>
              </div>

              <div className="issue-desc-box">
                <strong>📝 รายละเอียดปัญหา:</strong>
                <p className="issue-desc-text">{selectedRepair.description}</p>
              </div>
            </div>
            
            <div className="modal-actions">
              <button className="btn-close-modal" onClick={() => setIsDetailsOpen(false)}>❌ ปิดหน้าต่างนี้</button>
            </div>
          </div>
        </div>
      )}

      {/* ==========================================
          🔥 POPUP 2: รับงานและประเมินราคา (ระบบจะเช็ก Break-even ตรงนี้)
          ========================================== */}
      {estimateModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <span className="close-btn" onClick={() => setEstimateModalOpen(false)}>&times;</span>
            
            <h3 className="modal-title-primary">💸 ประเมินราคาก่อนดำเนินการซ่อม</h3>
            <p className="modal-subtitle">รหัสใบงาน: Ticket #{estimateTicketId}</p>

            <form onSubmit={handleSubmitEstimate}>
              <div className="input-group">
                <label>💰 ราคาประเมินเบื้องต้น (บาท) <span className="required">*</span></label>
                <input 
                  type="number" 
                  min="0"
                  step="0.01"
                  className="cost-input"
                  placeholder="ใส่ราคาประเมิน (กรอก 0 หากไม่เสียค่าใช้จ่าย)"
                  value={estimatedCost}
                  onChange={(e) => setEstimatedCost(e.target.value)}
                  required
                />
                <small className="cost-hint">
                  * หากค่าซ่อมเกินจุดคุ้มทุน (50% ของราคาซื้อ หรือยอดสะสมเกิน 70%) งานนี้จะถูกระงับและส่งให้แอดมินพิจารณาใหม่
                </small>
              </div>

              <div className="modal-actions">
                <button type="button" className="btn-cancel" onClick={() => setEstimateModalOpen(false)} disabled={isSubmitting}>
                  ❌ ยกเลิก
                </button>
                <button type="submit" className="btn-accept" disabled={isSubmitting}>
                  {isSubmitting ? "⏳ กำลังตรวจสอบ..." : "👍 ยืนยันรับงาน"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ==========================================
          POPUP 3: ปิดงาน (ซ่อมสำเร็จ / ซ่อมไม่ได้)
          ========================================== */}
      {modalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <span className="close-btn" onClick={() => setModalOpen(false)}>&times;</span>
            
            <h3 className={closeType === "success" ? "modal-title-success" : "modal-title-fail"}>
              {closeType === "success" ? "✅ ปิดงาน: ซ่อมเสร็จเรียบร้อย" : "❌ ปิดงาน: ซ่อมไม่ได้"}
            </h3>
            <p className="modal-subtitle">รหัสใบงาน: Ticket #{currentTicket}</p>

            <form onSubmit={handleSubmitClose}>
              {closeType === "success" && (
                <div className="input-group">
                  <label>💰 ค่าใช้จ่ายเบิกจริง (บาท) <span className="required">*</span></label>
                  <input 
                    type="number" 
                    min="0"
                    step="0.01"
                    className="cost-input"
                    placeholder="ใส่ยอดใช้จ่ายจริงเพื่อเก็บเป็นสถิติสะสม"
                    value={actualCost}
                    onChange={(e) => setActualCost(e.target.value)}
                    required
                  />
                </div>
              )}

              <div className="input-group">
                <label>
                  {closeType === "success" ? "📝 รายละเอียดการซ่อมแซม" : "📝 ระบุเหตุผลที่ขัดข้อง"} 
                  <span className="required">*</span>
                </label>
                <textarea 
                  rows="4"
                  className="textarea-input"
                  placeholder={closeType === "success" ? "เช่น ล้างแอร์ เติมน้ำยาแอร์..." : "เช่น อะไหล่ต้องสั่งจากต่างประเทศ อาการหนัก..."}
                  value={techDetail}
                  onChange={(e) => setTechDetail(e.target.value)}
                  required
                ></textarea>
              </div>
              
              <div className="modal-actions">
                <button type="button" className="btn-cancel" onClick={() => setModalOpen(false)} disabled={isSubmitting}>
                  ❌ ยกเลิก
                </button>
                <button 
                  type="submit" 
                  className={closeType === "success" ? "btn-submit-success" : "btn-submit-fail"}
                  disabled={isSubmitting}
                >
                  {isSubmitting ? "⏳ กำลังบันทึก..." : "💾 ยืนยันปิดงาน"}
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