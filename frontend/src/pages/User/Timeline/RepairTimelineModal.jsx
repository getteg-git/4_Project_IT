import { API_BASE_URL } from "../../../api/baseUrl.js";
import React, { useState, useEffect } from "react";
import { X, Clock, CheckCircle2, AlertCircle, PlayCircle, Ban, Camera, Circle } from "lucide-react";
import RepairImageGallery from "../../../components/RepairImageGallery";
import Pagination from "../../../components/ui/Pagination";
import { getPageItems } from "../../../components/ui/paginationUtils";
import "./RepairTimelineModal.css"; // เดี๋ยว CSS เราค่อยทำทีหลัง

function RepairTimelineModal({
  repairId,
  isOpen,
  onClose,
  images = [],
  ticketLabel,
  assignedTechnician,
  view = "details",
}) {
  const [logs, setLogs] = useState([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!isOpen || !repairId) return;

    const fetchLogs = async () => {
      setIsLoading(true);
      setError(null);
      setCurrentPage(1);
      try {
        const response = await fetch(`${API_BASE_URL}/api/repairs/${repairId}/logs`);
        if (response.ok) {
          const data = await response.json();
          setLogs(data || []);
        } else {
          setError("ไม่สามารถดึงข้อมูลประวัติได้");
        }
      } catch {
        setError("เกิดข้อผิดพลาดในการเชื่อมต่อเซิร์ฟเวอร์");
      } finally {
        setIsLoading(false);
      }
    };

    fetchLogs();
  }, [isOpen, repairId]);

  if (!isOpen) return null;

  const formatDate = (dateString) => {
    const options = { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' };
    return new Date(dateString).toLocaleDateString('th-TH', options);
  };

  // เลือกไอคอนและสีตาม Action
  const getActionConfig = (action) => {
    switch (action) {
      case "CREATED": return { icon: <Clock size={16} />, color: "text-blue", bg: "bg-blue" };
      case "ASSIGNED": return { icon: <PlayCircle size={16} />, color: "text-indigo", bg: "bg-indigo" };
      case "STATUS_CHANGED": return { icon: <CheckCircle2 size={16} />, color: "text-green", bg: "bg-green" };
      case "ESTIMATED": return { icon: <AlertCircle size={16} />, color: "text-orange", bg: "bg-orange" };
      case "APPROVED": return { icon: <CheckCircle2 size={16} />, color: "text-emerald", bg: "bg-emerald" };
      case "OUTSOURCED": return { icon: <AlertCircle size={16} />, color: "text-orange", bg: "bg-orange" };
      case "REVOKED": 
      case "REJECTED":
      case "CANCELLED": return { icon: <Ban size={16} />, color: "text-red", bg: "bg-red" };
      default: return { icon: <Clock size={16} />, color: "text-gray", bg: "bg-gray" };
    }
  };

  const getPublicProgress = () => {
    const chronologicalLogs = [...logs].sort(
      (left, right) => new Date(left.created_at) - new Date(right.created_at)
    );
    const creationLog = chronologicalLogs.find((log) => log.action === "CREATED") || chronologicalLogs[0];
    let assignmentLog = null;
    let assignmentIsActive = false;
    let startedLog = null;
    let completedLog = null;
    let latestLog = null;

    chronologicalLogs.forEach((log) => {
      latestLog = log;
      if (log.action === "ASSIGNED") {
        assignmentLog = log;
        assignmentIsActive = true;
      } else if (log.action === "REVOKED") {
        assignmentIsActive = false;
      }

      if (log.new_status === "รับงานแล้ว" && !assignmentLog) {
        assignmentLog = log;
        assignmentIsActive = true;
      }
      if (!startedLog && (log.new_status === "กำลังซ่อม" || log.action === "APPROVED")) {
        startedLog = log;
      }
      if (!completedLog && ["เสร็จเรียบร้อย", "เสร็จสิ้น"].includes(log.new_status)) {
        completedLog = log;
      }
    });

    const terminalLabels = {
      "ซ่อมไม่ได้": "งานนี้ไม่สามารถซ่อมได้",
      "ไม่สามารถซ่อมได้": "งานนี้ไม่สามารถซ่อมได้",
      "ส่งซ่อมภายนอก": "ส่งงานซ่อมต่อภายนอก",
      "ส่งซ่อมให้คณะ": "ส่งงานซ่อมต่อ",
      "ยกเลิก": "ยกเลิกงานซ่อม",
      "ยกเลิกงาน": "ยกเลิกงานซ่อม",
    };
    let terminalMessage = latestLog?.action === "CANCELLED"
      ? "ยกเลิกงานซ่อม"
      : latestLog?.action === "OUTSOURCED"
        ? "ส่งงานซ่อมต่อภายนอก"
        : terminalLabels[latestLog?.new_status];

    if (!assignmentIsActive) {
      assignmentLog = null;
    }

    return [
      { label: "รับแจ้งแล้ว", log: creationLog, complete: Boolean(creationLog) },
      {
        label: "มอบหมายช่างแล้ว",
        log: assignmentLog,
        complete: Boolean(assignmentLog || assignedTechnician),
      },
      { label: "กำลังดำเนินการ", log: startedLog, complete: Boolean(startedLog) },
      { label: "ดำเนินการเสร็จแล้ว", log: completedLog, complete: Boolean(completedLog) },
    ].map((stage, index, stages) => ({
      ...stage,
      current: stage.complete && !stages.slice(index + 1).some((next) => next.complete),
    })).map((stage, index) => ({
      ...stage,
      icon: stage.complete
        ? <CheckCircle2 size={17} />
        : index === 0
          ? <Clock size={17} />
          : <Circle size={17} />,
      terminalMessage: index === 3 && terminalMessage && !completedLog ? terminalMessage : null,
    }));
  };

  const paginatedLogs = getPageItems(logs, currentPage);
  const publicProgress = view === "summary" ? getPublicProgress() : [];

  return (
    <div className="modal-overlay">
      <div className="modal-box timeline-box">
        <button className="close-btn" onClick={onClose} aria-label="ปิด">
          <X aria-hidden="true" />
        </button>
        
        <h3>{view === "summary" ? "ความคืบหน้างานซ่อม" : "ประวัติการดำเนินการ"}</h3>
        <p className="ticket-subtitle">{ticketLabel || `Ticket ID: #${repairId}`}</p>
        <hr />

        <div className="timeline-content">
          {images.length > 0 && (
            <section className="timeline-repair-images" aria-label="ภาพประกอบงานซ่อม">
              <h4><Camera size={17} aria-hidden="true" /> ภาพประกอบงานซ่อม</h4>
              <RepairImageGallery
                images={images}
                ticketLabel={ticketLabel || `#${repairId}`}
              />
            </section>
          )}
          {isLoading ? (
            <div className="loading-state">กำลังโหลดประวัติ...</div>
          ) : error ? (
            <div className="error-state">{error}</div>
          ) : logs.length === 0 ? (
            <div className="no-data">ไม่พบประวัติการทำรายการ</div>
          ) : view === "summary" ? (
            <div className="timeline-container timeline-summary">
              {publicProgress.map((stage, index) => (
                <div
                  className={`timeline-item ${stage.complete ? "is-complete" : "is-pending"} ${stage.current ? "is-current" : ""}`}
                  key={stage.label}
                >
                  {index !== publicProgress.length - 1 && <div className="timeline-line" />}
                  <div className={`timeline-icon ${stage.complete ? "bg-green" : "bg-gray"}`}>
                    {stage.icon}
                  </div>
                  <div className="timeline-details">
                    <div className="timeline-header">
                      <span className={`timeline-status ${stage.complete ? "text-green" : "text-gray"}`}>
                        {stage.label}
                      </span>
                      {stage.log && (
                        <span className="timeline-date">{formatDate(stage.log.created_at)}</span>
                      )}
                    </div>
                    {index === 1 && stage.complete && assignedTechnician && (
                      <p className="timeline-public-detail">ช่างผู้รับผิดชอบ: {assignedTechnician}</p>
                    )}
                    {stage.terminalMessage && (
                      <p className="timeline-terminal">{stage.terminalMessage}</p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="timeline-container">
              {paginatedLogs.map((log, index) => {
                const config = getActionConfig(log.action);
                return (
                  <div key={log.id} className="timeline-item">
                    {/* เส้นเชื่อม (ยกเว้นอันสุดท้าย) */}
                    {index !== paginatedLogs.length - 1 && <div className="timeline-line"></div>}
                    
                    <div className={`timeline-icon ${config.bg}`}>
                      {config.icon}
                    </div>
                    
                    <div className="timeline-details">
                      <div className="timeline-header">
                        <span className={`timeline-status ${config.color}`}>
                          {log.new_status || log.action}
                        </span>
                        <span className="timeline-date">{formatDate(log.created_at)}</span>
                      </div>
                      
                      <div className="timeline-body">
                        <p className="timeline-user">
                          <strong>ผู้ทำรายการ:</strong> {log.user_name || "ระบบ / ผู้แจ้ง"}
                        </p>
                        {log.note && (
                          <div className="timeline-note">
                            {log.note}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
              <Pagination
                currentPage={currentPage}
                totalItems={logs.length}
                onPageChange={setCurrentPage}
              />
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default RepairTimelineModal;