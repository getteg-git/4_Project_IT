import { API_BASE_URL, apiFetch as fetch } from "../../../api/baseUrl.js";
import React, { useState } from "react";
import { CalendarDays, CircleX, Eye, FileText, Mail, UserRound, Wrench, X, ClipboardCheck, Clock } from "lucide-react";
import AppBreadcrumb from "../../../components/AppBreadcrumb";
import RepairImageGallery from "../../../components/RepairImageGallery";
import BackButton from "../../../components/ui/BackButton";
import RepairCardNotes from "../../../components/RepairCardNotes";
import RepairTimelineModal from "../Timeline/RepairTimelineModal";
import { getRepairStatusLabel } from "../../../utils/repairStatus.js";
import "./MyRepairs.css";

function MyRepairs() {
  const [reporterEmail, setReporterEmail] = useState("");
  const [trackingCredentials, setTrackingCredentials] = useState(null);
  const [searchError, setSearchError] = useState("");
  const [repairs, setRepairs] = useState([]);
  const [isLoading, setIsLoading] = useState(false);

  const [selectedRepair, setSelectedRepair] = useState(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const [isTimelineOpen, setIsTimelineOpen] = useState(false);
  const trackingHeaders = trackingCredentials
    ? {
        "X-Reporter-Email": trackingCredentials.email,
      }
    : undefined;

  const handleTrackRepair = async (event) => {
    event.preventDefault();
    const email = reporterEmail.trim();
    if (!email) {
      setSearchError("กรุณากรอกอีเมลที่ใช้แจ้งซ่อม");
      return;
    }

    setIsLoading(true);
    setSearchError("");
    setRepairs([]);
    setTrackingCredentials(null);
    setSelectedRepair(null);
    setIsDetailsOpen(false);
    setIsTimelineOpen(false);
    try {
      const query = new URLSearchParams({ email });
      const response = await fetch(`${API_BASE_URL}/api/repairs/track?${query}`);
      const data = await response.json();
      if (!response.ok) {
        setSearchError(data.error || "ไม่พบรายการแจ้งซ่อมที่ใช้อีเมลนี้");
        return;
      }
      setRepairs(Array.isArray(data) ? data : []);
      setTrackingCredentials({ email });
    } catch (error) {
      setSearchError(error.message || "ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้");
    } finally {
      setIsLoading(false);
    }
  };

  const filteredRepairs = repairs;
  const paginatedRepairs = repairs;
  const getStatusClass = (status) => {
    switch (status) {
      case "รอซ่อม": return "status-waiting";
      case "รับงานแล้ว": return "status-pending";
      case "กำลังซ่อม": return "status-progress";
      case "เสร็จเรียบร้อย": return "status-completed";
      case "เสร็จสิ้น": return "status-completed";
      case "ซ่อมไม่ได้": return "status-failed";
      case "ส่งซ่อมภายนอก": return "status-outsourced";
      case "ส่งซ่อมภายนอก / คณะ": return "status-outsourced";
      case "ส่งซ่อมให้คณะ": return "status-faculty";
      default: return "status-default";
    }
  };

  const formatDate = (dateString) => {
    const options = { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' };
    return new Date(dateString).toLocaleDateString('th-TH', options);
  };

  const getOutcomeNotes = (repair) => {
    const notes = [];
    if (repair.status === "ซ่อมไม่ได้") {
      const adminRejection = repair.admin_note?.replace(/^ไม่อนุมัติการซ่อม:\s*/u, "");
      const reason = adminRejection || repair.technician_note || repair.admin_note;
      if (reason?.trim()) notes.push({ label: "เหตุผลที่ซ่อมไม่สำเร็จ", value: reason.trim() });
    } else if (repair.status?.startsWith("ส่งซ่อม")) {
      if (repair.technician_note?.trim()) {
        notes.push({ label: "เหตุผลที่ซ่อมไม่สำเร็จ", value: repair.technician_note.trim() });
      }
      const handoffDetails = repair.admin_note?.replace(/^ส่งซ่อม(?:ภายนอก|ให้คณะ)(?:\s*\/\s*คณะ)?:\s*/u, "");
      if (handoffDetails?.trim()) {
        notes.push({ label: "รายละเอียดการส่งซ่อมต่อ", value: handoffDetails.trim() });
      }
    }
    return notes;
  };

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
        <AppBreadcrumb items={[
          { label: "หน้าหลัก", to: "/" },
          { label: "ติดตามสถานะการซ่อม" },
        ]} />

        <section className="tracking-hero">
          <nav className="page-navigation" aria-label="Breadcrumb">
            <BackButton to="/" label="กลับหน้าหลัก" />
          </nav>

          <div className="tracking-header">
            <h2>ติดตามสถานะการแจ้งซ่อม</h2>
            <p>ค้นหาและตรวจสอบความคืบหน้าของรายการแจ้งซ่อมภายในคณะวิทยาศาสตร์</p>
          </div>

          <form className="track-lookup-form" onSubmit={handleTrackRepair}>
            <label>
              อีเมลที่ใช้แจ้งซ่อม
              <input
                type="email"
                value={reporterEmail}
                onChange={(event) => setReporterEmail(event.target.value)}
                placeholder="name@example.com"
                maxLength={254}
                required
              />
            </label>
            <button type="submit" disabled={isLoading}>
              {isLoading ? "กำลังค้นหา..." : "ค้นหางานของฉัน"}
            </button>
            {searchError && <p className="track-lookup-error" role="alert">{searchError}</p>}
            <p className="track-lookup-hint">
              ระบบจะแสดงรายการแจ้งซ่อมทั้งหมดที่ผูกกับอีเมลนี้ โดย Ticket จะแสดงบนแต่ละรายการ
            </p>
          </form>
        </section>

        <div className="repair-list">
          {isLoading ? (
            <div className="loading-state">กำลังค้นหารายการแจ้งซ่อม...</div>
          ) : filteredRepairs.length > 0 ? (
            paginatedRepairs.map((repair) => {
              const displayLocation = repair.other_location || repair.location_name;
              const displayProblemType = repair.other_problem_type || repair.problem_type;
              const isCompleted = repair.status === "เสร็จเรียบร้อย" || repair.status === "เสร็จสิ้น";

              return (
                <div className="repair-item-card" key={repair.id}>
                  <div className="card-top">
                    <span className="repair-id">{repair.ticket_number || `Ticket #${repair.id}`}</span>
                    <span className={`status-badge ${getStatusClass(repair.status)}`}>
                      {getRepairStatusLabel(repair.status)}
                    </span>
                  </div>

                  <h3 className="repair-title">[{displayProblemType}] {displayLocation}</h3>

                  <p className="repair-room">
                    <strong>ชั้น / พิกัด:</strong> {repair.floor_name || "-"} {(repair.other_room || repair.room_number) ? `(${repair.other_room || repair.room_number})` : ""}
                  </p>

                  {repair.reporter_email && (
                    <p className="repair-reporter-email">
                      <Mail size={15} aria-hidden="true" /> <strong>อีเมลผู้แจ้ง:</strong> {repair.reporter_email}
                    </p>
                  )}

                  {repair.equipment_name && (
                    <p className="repair-equipment" style={{ margin: "4px 0", fontSize: "0.9rem", color: "#e67e22" }}>
                      <Wrench size={15} aria-hidden="true" /> <strong>อุปกรณ์:</strong> {repair.equipment_name} {repair.asset_code ? `(${repair.asset_code})` : ""}
                    </p>
                  )}

                  <p className="repair-desc"><strong>รายละเอียด:</strong> {repair.description}</p>

                  <RepairCardNotes
                    repair={repair}
                    showTechnicianNote={false}
                    showAdminNote={getOutcomeNotes(repair).length === 0}
                    extraNotes={getOutcomeNotes(repair)}
                  />

                  {repair.technician_name && (
                    <p className="repair-tech-badge"><strong><UserRound size={15} aria-hidden="true" /> ช่าง:</strong> {repair.technician_name}</p>
                  )}

                  {isCompleted && (
                    <div style={{ marginTop: "12px", marginBottom: "4px" }}>
                      <a
                        href="https://docs.google.com/forms/d/1sztsmrllXVDAAAhN9DSuE_zrYxRylyuxKWiOURnJWTw/viewform?edit_requested=true"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="btn-evaluation"
                        style={{ display: "inline-flex", alignItems: "center", gap: "6px", backgroundColor: "#8e44ad", color: "#fff", padding: "8px 14px", borderRadius: "6px", textDecoration: "none", fontSize: "0.9rem", fontWeight: "500" }}
                      >
                        <ClipboardCheck size={16} /> ทำแบบประเมินความพึงพอใจ
                      </a>
                    </div>
                  )}

                  {/* 🔥 แก้ไขการจัดวางปุ่มในการ์ดให้เรียงกันสวยงาม */}
                  <div className="card-bottom">
                    <span className="repair-date"><CalendarDays size={15} aria-hidden="true" /> {formatDate(repair.created_at)}</span>
                    <div style={{ display: 'flex', gap: '8px' }}>
                      <button className="btn-details" onClick={() => openDetailsModal(repair)}>
                        <Eye size={17} aria-hidden="true" /> ดูรายละเอียด
                      </button>
                      <button
                        className="btn-timeline"
                        onClick={() => { setSelectedRepair(repair); setIsTimelineOpen(true); }}
                      >
                        <Clock size={17} aria-hidden="true" /> ไทม์ไลน์
                      </button>
                    </div>
                  </div>
                </div>
              );
            })
          ) : (
            <div className="no-results">
              <p><CircleX size={18} aria-hidden="true" /> กรอกอีเมลที่ใช้แจ้งซ่อมเพื่อค้นหารายการของคุณ</p>
            </div>
          )}
        </div>

      </div>

      {/* ==========================================
          POPUP: ดูรายละเอียดงาน (Modal)
          ========================================== */}
      {isDetailsOpen && selectedRepair && (
        <div className="modal-overlay">
          <div className="modal-box details-box">
            <button className="close-btn" type="button" onClick={closeDetailsModal} aria-label="ปิดรายละเอียด"><X aria-hidden="true" /></button>
            <h3>รายละเอียดการแจ้งซ่อม</h3>
            <p className="ticket-subtitle">Ticket ID: {selectedRepair.ticket_number || `#${selectedRepair.id}`}</p>
            <hr />

            <div className="modal-details-content">
              <RepairImageGallery
                images={selectedRepair.images}
                ticketLabel={selectedRepair.ticket_number || `#${selectedRepair.id}`}
                requestHeaders={trackingHeaders}
              />

              <div className="info-grid">
                <p><strong>สถานะปัจจุบัน:</strong> <span className={`status-badge ${getStatusClass(selectedRepair.status)}`}>{getRepairStatusLabel(selectedRepair.status)}</span></p>
                <p><strong>หมวดหมู่งาน:</strong> {selectedRepair.other_problem_type || selectedRepair.problem_type}</p>
                <p><strong>สถานที่:</strong> {selectedRepair.other_location || selectedRepair.location_name}</p>
                <p>
                  <strong>ชั้น / พิกัด:</strong> {selectedRepair.floor_name || "-"} {(selectedRepair.other_room || selectedRepair.room_number) ? `(${selectedRepair.other_room || selectedRepair.room_number})` : ""}
                </p>
                {selectedRepair.equipment_name && (
                  <p>
                    <strong>อุปกรณ์ชำรุด:</strong> {selectedRepair.equipment_name} {selectedRepair.asset_code ? `(${selectedRepair.asset_code})` : ""}
                  </p>
                )}
                <p><strong>อีเมลผู้แจ้ง:</strong> {selectedRepair.reporter_email}</p>
                {selectedRepair.reporter_type && (
                  <p><strong>ประเภทผู้แจ้ง:</strong> {selectedRepair.reporter_type}</p>
                )}
                {selectedRepair.reporter_degree_level && (
                  <p><strong>ระดับการศึกษา:</strong> {selectedRepair.reporter_degree_level}</p>
                )}
                {selectedRepair.reporter_department && (
                  <p><strong>สาขาวิชา / หน่วยงาน:</strong> {selectedRepair.reporter_department}</p>
                )}
                <p><strong>วันที่แจ้งเรื่อง:</strong> {formatDate(selectedRepair.created_at)}</p>
              </div>

              <div className="issue-desc-box">
                <strong><FileText size={17} aria-hidden="true" /> รายละเอียดปัญหา:</strong>
                <p>{selectedRepair.description}</p>
              </div>

              <RepairCardNotes
                repair={selectedRepair}
                showTechnicianNote={false}
                showAdminNote={false}
                extraNotes={getOutcomeNotes(selectedRepair)}
              />

              {selectedRepair.status !== "รอซ่อม" && (
                <div className="tech-info-box">
                  <p><strong><UserRound size={17} aria-hidden="true" /> ช่างผู้รับผิดชอบ:</strong> {selectedRepair.technician_name || "อยู่ระหว่างดำเนินการ"}</p>
                  {(selectedRepair.status === "เสร็จเรียบร้อย" || selectedRepair.status === "เสร็จสิ้น") && (
                    <div style={{ marginTop: "15px", textAlign: "center" }}>
                      <a
                        href="https://docs.google.com/forms/d/1sztsmrllXVDAAAhN9DSuE_zrYxRylyuxKWiOURnJWTw/viewform?edit_requested=true"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="btn-evaluation"
                        style={{ display: "inline-flex", alignItems: "center", justifyContent: "center", gap: "6px", backgroundColor: "#8e44ad", color: "#fff", padding: "10px 16px", borderRadius: "6px", textDecoration: "none", fontSize: "0.95rem", fontWeight: "bold", width: "100%" }}
                      >
                        <ClipboardCheck size={18} /> กรุณาทำแบบประเมินความพึงพอใจ
                      </a>
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* ==========================================
          POPUP: ไทม์ไลน์ (Timeline Modal)
          ========================================== */}
      <RepairTimelineModal 
        repairId={selectedRepair?.id} 
        images={selectedRepair?.images}
        ticketLabel={selectedRepair?.ticket_number || `#${selectedRepair?.id}`}
        trackingCredentials={trackingCredentials}
        assignedTechnician={selectedRepair?.technician_name}
        view="summary"
        isOpen={isTimelineOpen} 
        onClose={() => { setIsTimelineOpen(false); setSelectedRepair(null); }} 
      />
    </div>
  );
}

export default MyRepairs;