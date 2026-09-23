import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  Banknote,
  CalendarDays,
  CircleCheck,
  CircleX,
  ClipboardList,
  Clock3,
  Eye,
  FileText,
  LogOut,
  Save,
  Wrench,
  X,
  Clock
} from "lucide-react";
import useToast from "../../../hooks/useToast";
import SearchField from "../../../components/ui/SearchField";
import Pagination from "../../../components/ui/Pagination";
import { getPageItems } from "../../../components/ui/paginationUtils";
import FilterBar from "../../../components/ui/FilterBar";
import RepairTimelineModal from "../../User/Timeline/RepairTimelineModal"; // 🔥 นำเข้า Timeline Modal
import "./TechHome.css";

function TechHome() {
  const navigate = useNavigate();
  const { toast, confirm } = useToast();

  const [currentUser, setCurrentUser] = useState(null);
  const [repairs, setRepairs] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [jobView, setJobView] = useState("active");
  const [problemTypeFilter, setProblemTypeFilter] = useState("");
  const [currentPage, setCurrentPage] = useState(1);

  const [selectedRepair, setSelectedRepair] = useState(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const [isTimelineOpen, setIsTimelineOpen] = useState(false); // 🔥 State เปิด Timeline

  // States สำหรับ Modal ปิดงาน (ซ่อมสำเร็จ/ซ่อมไม่ได้)
  const [modalOpen, setModalOpen] = useState(false);
  const [currentTicket, setCurrentTicket] = useState(null);
  const [closeType, setCloseType] = useState("");
  const [techDetail, setTechDetail] = useState("");
  const [actualCost, setActualCost] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  // States สำหรับ Modal "ประเมินราคา" (Estimate Modal)
  const [estimateModalOpen, setEstimateModalOpen] = useState(false);
  const [estimateTicketId, setEstimateTicketId] = useState(null);
  const [estimatedCost, setEstimatedCost] = useState("");

  const fetchMyJobs = async (user) => {
    try {
      setIsLoading(true);
      const response = await fetch("http://localhost:8080/api/repairs");
      if (response.ok) {
        const allRepairs = await response.json();
        // ช่างจะเห็นแค่งานที่ถูก Assign มาให้ตัวเองเท่านั้น
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
  const problemTypeOptions = [...new Set(repairs.map((repair) => repair.other_problem_type || repair.problem_type).filter(Boolean))]
    .sort((a, b) => String(a).localeCompare(String(b), "th"));

  const filteredRepairs = repairs.filter((repair) => {
    if (problemTypeFilter && (repair.other_problem_type || repair.problem_type) !== problemTypeFilter) return false;
    if (!normalizedSearch) return true;
    const loc = (repair.location_name || "").replace(/\s+/g, '').toLowerCase();
    const floor = (repair.floor_name || "").replace(/\s+/g, '').toLowerCase();
    const room = (repair.room_number || "").replace(/\s+/g, '').toLowerCase();
    const equip = (repair.equipment_name || "").replace(/\s+/g, '').toLowerCase();
    const status = (repair.status || "").replace(/\s+/g, '').toLowerCase();
    const ticketId = (repair.ticket_number || String(repair.id)).toLowerCase(); // 🔥 ค้นหาด้วย Ticket Number

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

  const formatRoom = (roomNumber) => {
    if (!roomNumber) return "";
    const roomText = String(roomNumber).trim();
    return roomText.startsWith("ห้อง") ? `(${roomText})` : `(ห้อง ${roomText})`;
  };

  const openDetailsModal = (repair) => {
    setSelectedRepair(repair);
    setIsDetailsOpen(true);
  };

  // 🔥 1. ฟังก์ชัน "รับงาน" (เริ่มจับเวลา + เปลี่ยนเป็นกำลังซ่อม)
  const handleAcceptJob = async (id) => {
    try {
      const formData = new FormData();
      formData.append("status", "กำลังซ่อม");
      formData.append("technician_id", currentUser.id);

      const response = await fetch(`http://localhost:8080/api/repairs/${id}/status`, {
        method: "PUT",
        body: formData
      });

      if (response.ok) {
        toast.success("รับงานสำเร็จ", { description: "ระบบเริ่มบันทึกเวลาการปฏิบัติงานแล้ว" });
        fetchMyJobs(currentUser);
      } else {
        toast.error("รับงานไม่สำเร็จ", { description: "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch (error) {
      console.error("Accept Job Error:", error);
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้");
    }
  };

  const handleReject = async (id) => {
    const confirmReject = await confirm({
      title: "ปฏิเสธงานซ่อม",
      description: `งานซ่อมนี้จะถูกส่งคืนให้ผู้ดูแลระบบพิจารณามอบหมายใหม่`,
      confirmLabel: "ปฏิเสธงาน",
      variant: "danger",
    });
    if (!confirmReject) return;

    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${id}/reject`, {
        method: "PUT"
      });

      if (response.ok) {
        toast.success("ส่งงานกลับผู้ดูแลแล้ว", { description: "ผู้ดูแลระบบจะพิจารณามอบหมายงานอีกครั้ง" });
        fetchMyJobs(currentUser);
      } else {
        toast.error("ปฏิเสธงานไม่สำเร็จ", { description: "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch (error) {
      console.error("Reject Error:", error);
    }
  };

  // 🔥 2. เปิดหน้าต่างประเมินราคา (สำหรับเช็ก Break-even)
  const openEstimateModal = (id) => {
    setEstimateTicketId(id);
    setEstimatedCost("");
    setEstimateModalOpen(true);
  };

  const handleSubmitEstimate = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    const costValue = estimatedCost === "" ? 0 : Number(estimatedCost);

    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${estimateTicketId}/estimate`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ estimated_cost: costValue })
      });

      if (response.ok) {
        const data = await response.json();
        if (data.status === "รอซ่อม") {
          toast.warning("งานรอการอนุมัติ", { description: data.message || "ระบบส่งเรื่องกลับให้ผู้ดูแลพิจารณาแล้ว" });
        } else {
          toast.success("ประเมินราคาสำเร็จ", { description: "คุณสามารถดำเนินการซ่อมต่อได้เลย" });
        }
        setEstimateModalOpen(false);
        fetchMyJobs(currentUser);
      } else {
        const errorData = await response.json();
        toast.error("ส่งราคาประเมินไม่สำเร็จ", { description: errorData.error || "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch (error) {
      console.error("Estimate Error:", error);
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้");
    } finally {
      setIsSubmitting(false);
    }
  };

  // 🔥 3. ปิดงานซ่อม (ใช้ FormData ส่งเข้า API /status)
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

    const formData = new FormData();
    formData.append("status", newStatus);
    formData.append("technician_note", techDetail);
    formData.append("actual_cost", costValue);
    formData.append("technician_id", currentUser.id);

    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${currentTicket}/status`, {
        method: "PUT",
        body: formData
      });

      if (response.ok) {
        toast.success("บันทึกสถานะสำเร็จ", { description: `อัปเดตงานเป็นสถานะ “${newStatus}” แล้ว` });
        setModalOpen(false);
        fetchMyJobs(currentUser);
      } else {
        const errorData = await response.json();
        toast.error("บันทึกสถานะไม่สำเร็จ", { description: errorData.error || "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch (error) {
      console.error("Close Job Error:", error);
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleLogout = async () => {
    const approved = await confirm({ title: "ออกจากระบบ", description: "คุณต้องการออกจากระบบช่างเทคนิคหรือไม่?", confirmLabel: "ออกจากระบบ", variant: "danger" });
    if (approved) {
      localStorage.removeItem("user");
      navigate("/");
    }
  };

  const activeRepairs = filteredRepairs.filter((repair) => repair.status === "รอซ่อม");
  const progressRepairs = filteredRepairs.filter((repair) => repair.status === "กำลังซ่อม");
  const completedRepairs = filteredRepairs.filter((repair) =>
    repair.status !== "รอซ่อม" && repair.status !== "กำลังซ่อม"
  );
  const displayedRepairs = jobView === "active"
    ? activeRepairs
    : jobView === "progress"
      ? progressRepairs
      : completedRepairs;
  const paginatedRepairs = getPageItems(displayedRepairs, currentPage);

  const renderRepairCard = (repair) => (
    <div className="repair-item-card" key={repair.id}>
      <div className="card-top">
        {/* 🔥 แสดง Ticket Number */}
        <span className="repair-ticket-id">{repair.ticket_number || `Ticket #${repair.id}`}</span>
        <span className={`status-badge ${getStatusClass(repair.status)}`}>{repair.status}</span>
      </div>

      <h3 className="repair-title">[{repair.problem_type}] {repair.location_name}</h3>

      <div className="repair-details-text">
        <p className="repair-info">
          <strong>ชั้น / พิกัด:</strong> {repair.floor_name || "-"} {formatRoom(repair.room_number)}
        </p>
        {repair.equipment_name && (
          <p className="repair-info equipment-info">
            <strong><Wrench size={16} aria-hidden="true" /> อุปกรณ์:</strong> {repair.equipment_name}
          </p>
        )}
        <p className="repair-info"><strong>อาการที่แจ้ง:</strong> {repair.description}</p>
      </div>

      <div className="card-actions-row">
        <span className="repair-date"><CalendarDays size={16} aria-hidden="true" /> แจ้งเมื่อ: {formatDate(repair.created_at)}</span>
        <div className="action-buttons" style={{ display: 'flex', gap: '8px' }}>
          <button className="btn-details" onClick={() => openDetailsModal(repair)}>
            <Eye size={16} aria-hidden="true" /> รายละเอียด
          </button>

          {/* 🔥 ปุ่มดู Timeline สำหรับช่าง */}
          <button
            className="btn-timeline"
            onClick={() => { setSelectedRepair(repair); setIsTimelineOpen(true); }}
            style={{ backgroundColor: '#f1f5f9', color: '#334155', border: 'none', padding: '6px 12px', borderRadius: '6px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '4px', fontSize: '0.85rem' }}
          >
            <Clock size={16} aria-hidden="true" /> ไทม์ไลน์
          </button>
        </div>
      </div>

      {repair.status === "รอซ่อม" && (
        <div className="job-decision-row">
          {repair.admin_note && repair.admin_note.trim() !== "" ? (
            <button className="btn-waiting-approval" disabled><Clock3 size={17} aria-hidden="true" /> รอแอดมินอนุมัติงบ</button>
          ) : (
            <button className="btn-accept" onClick={() => handleAcceptJob(repair.id)}>
              <CircleCheck size={17} aria-hidden="true" /> รับงาน
            </button>
          )}
          <button className="btn-reject" onClick={() => handleReject(repair.id)}>
            <X size={17} aria-hidden="true" /> ปฏิเสธงาน
          </button>
        </div>
      )}

      {/* 🔥 เพิ่มปุ่มประเมินราคาเมื่อสถานะเป็นกำลังซ่อม */}
      {repair.status === "กำลังซ่อม" && (
        <div className="job-completion-row" style={{ flexWrap: 'wrap' }}>
          {/* ถ้ายังไม่เคยประเมินราคา (เป็น null หรือ undefined) ให้แสดงปุ่มประเมิน */}
          {(repair.estimated_cost === null || repair.estimated_cost === undefined) && (
            <button className="btn-estimate" onClick={() => openEstimateModal(repair.id)} style={{ backgroundColor: '#f59e0b', color: '#fff', border: 'none', padding: '8px 16px', borderRadius: '6px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px', fontWeight: 'bold' }}>
              <Banknote size={17} aria-hidden="true" /> ประเมินราคา
            </button>
          )}
          <button className="btn-complete" onClick={() => openCloseModal(repair.id, "success")}>
            <CircleCheck size={17} aria-hidden="true" /> ซ่อมสำเร็จ
          </button>
          <button className="btn-cannot-fix" onClick={() => openCloseModal(repair.id, "fail")}>
            <CircleX size={17} aria-hidden="true" /> ซ่อมไม่สำเร็จ
          </button>
        </div>
      )}
    </div>
  );

  return (
    <div className="tech-container">
      <div className="tech-wrapper">

        <div className="tech-header">
          <div>
            <h2><ClipboardList size={24} aria-hidden="true" /> งานซ่อมที่รับผิดชอบ</h2>
            <p>ยินดีต้อนรับ, {currentUser ? (currentUser.full_name || currentUser.username) : "กำลังโหลด..."}</p>
          </div>
          <div className="header-actions">
            <button className="btn-logout" onClick={handleLogout}>
              <LogOut size={17} aria-hidden="true" /> <span>ออกจากระบบ</span>
            </button>
          </div>
        </div>

        <div className="search-section">
          <SearchField
            id="technician-repair-search"
            label="ค้นหางานที่รับผิดชอบ"
            placeholder="หมายเลข Ticket สถานที่ ห้อง อุปกรณ์ หรือสถานะ"
            value={searchTerm}
            onChange={(value) => { setSearchTerm(value); setCurrentPage(1); }}
          />
          <FilterBar filters={[
            { id: "problem-type", label: "ประเภทงาน", value: problemTypeFilter, options: problemTypeOptions, onChange: (value) => { setProblemTypeFilter(value); setCurrentPage(1); } },
          ]} />
        </div>

        {isLoading ? (
          <div className="loading-state">กำลังโหลดรายการงานของคุณ...</div>
        ) : (
          <div className="job-view-area">
            <div className="job-view-tabs" role="tablist" aria-label="เลือกประเภทงาน">
              <button type="button" role="tab" aria-selected={jobView === "active"} className={jobView === "active" ? "is-active" : ""} onClick={() => { setJobView("active"); setCurrentPage(1); }}>
                งานที่รอรับ <span>{activeRepairs.length}</span>
              </button>
              <button type="button" role="tab" aria-selected={jobView === "progress"} className={jobView === "progress" ? "is-active" : ""} onClick={() => { setJobView("progress"); setCurrentPage(1); }}>
                กำลังซ่อม <span>{progressRepairs.length}</span>
              </button>
              <button type="button" role="tab" aria-selected={jobView === "completed"} className={jobView === "completed" ? "is-active" : ""} onClick={() => { setJobView("completed"); setCurrentPage(1); }}>
                เสร็จแล้ว <span>{completedRepairs.length}</span>
              </button>
            </div>
            {displayedRepairs.length > 0 ? (
              <>
                <div className="repair-list">{paginatedRepairs.map(renderRepairCard)}</div>
                <Pagination currentPage={currentPage} totalItems={displayedRepairs.length} onPageChange={setCurrentPage} />
              </>
            ) : (
              <div className="no-results"><p>{jobView === "active" ? "ไม่มีงานที่รอรับในขณะนี้" : jobView === "progress" ? "ไม่มีงานที่กำลังซ่อม" : "ยังไม่มีประวัติงานที่เสร็จแล้ว"}</p></div>
            )}
          </div>
        )}

      </div>

      {/* ==========================================
          POPUP 1: ดูรายละเอียดงาน
          ========================================== */}
      {isDetailsOpen && selectedRepair && (
        <div className="modal-overlay">
          <div className="modal-box details-box">
            <button className="close-btn" type="button" onClick={() => setIsDetailsOpen(false)} aria-label="ปิดรายละเอียด"><X aria-hidden="true" /></button>

            <div className="modal-header-center">
              <h3 className="modal-title-primary"><FileText size={22} aria-hidden="true" /> ข้อมูลการแจ้งซ่อมอย่างละเอียด</h3>
              <p className="modal-subtitle">Ticket: {selectedRepair.ticket_number || `#${selectedRepair.id}`}</p>
            </div>

            <div className="modal-details-content">
              <div className="image-gallery">
                {selectedRepair.images && selectedRepair.images.length > 0 ? (
                  selectedRepair.images.map((img, index) => (
                    <img key={index} src={`http://localhost:8080${img.url}`} alt="รูปปัญหา" className="repair-image" />
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
                  <strong>ชั้น / พิกัด:</strong> {selectedRepair.floor_name || "-"} {formatRoom(selectedRepair.room_number)}
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
                <strong><FileText size={17} aria-hidden="true" /> รายละเอียดปัญหา:</strong>
                <p className="issue-desc-text">{selectedRepair.description}</p>
              </div>
            </div>

          </div>
        </div>
      )}

      {/* ==========================================
          🔥 POPUP 2: ประเมินราคา (ระบบจะเช็ก Break-even ตรงนี้)
          ========================================== */}
      {estimateModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <button className="close-btn" type="button" onClick={() => setEstimateModalOpen(false)} aria-label="ปิดหน้าต่าง"><X aria-hidden="true" /></button>

            <h3 className="modal-title-primary"><Banknote size={23} aria-hidden="true" /> ประเมินราคาก่อนดำเนินการซ่อม</h3>
            <p className="modal-subtitle">รหัสใบงาน: {repairs.find(r => r.id === estimateTicketId)?.ticket_number || `#${estimateTicketId}`}</p>

            <form onSubmit={handleSubmitEstimate}>
              <div className="input-group">
                <label><Banknote size={17} aria-hidden="true" /> ราคาประเมินเบื้องต้น (บาท) <span className="required">*</span></label>
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
                  * หากค่าซ่อมเกินจุดคุ้มทุน งานนี้จะถูกระงับชั่วคราวและส่งให้แอดมินพิจารณาอนุมัติงบใหม่
                </small>
              </div>

              <div className="modal-actions">
                <button type="submit" className="btn-accept" disabled={isSubmitting}>
                  {isSubmitting ? "กำลังตรวจสอบ..." : <><CircleCheck size={18} aria-hidden="true" /> บันทึกราคาประเมิน</>}
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
            <button className="close-btn" type="button" onClick={() => setModalOpen(false)} aria-label="ปิดหน้าต่าง"><X aria-hidden="true" /></button>

            <h3 className={closeType === "success" ? "modal-title-success" : "modal-title-fail"}>
              {closeType === "success"
                ? <><CircleCheck size={23} aria-hidden="true" /> ปิดงาน: ซ่อมเสร็จเรียบร้อย</>
                : <><CircleX size={23} aria-hidden="true" /> ปิดงาน: ซ่อมไม่สำเร็จ</>}
            </h3>
            <p className="modal-subtitle">รหัสใบงาน: {repairs.find(r => r.id === currentTicket)?.ticket_number || `#${currentTicket}`}</p>

            <form onSubmit={handleSubmitClose}>
              {closeType === "success" && (
                <div className="input-group">
                  <label><Banknote size={17} aria-hidden="true" /> ค่าใช้จ่ายเบิกจริง (บาท) <span className="required">*</span></label>
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
                  <FileText size={17} aria-hidden="true" />
                  {closeType === "success" ? "รายละเอียดการซ่อมแซม" : "ระบุเหตุผลที่ขัดข้อง"}
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
                <button
                  type="submit"
                  className={closeType === "success" ? "btn-submit-success" : "btn-submit-fail"}
                  disabled={isSubmitting}
                >
                  {isSubmitting ? "กำลังบันทึก..." : <><Save size={18} aria-hidden="true" /> ยืนยันปิดงาน</>}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* 🔥 POPUP 4: Timeline Modal */}
      <RepairTimelineModal
        repairId={selectedRepair?.id}
        isOpen={isTimelineOpen}
        onClose={() => { setIsTimelineOpen(false); setSelectedRepair(null); }}
      />

    </div>
  );
}

export default TechHome;