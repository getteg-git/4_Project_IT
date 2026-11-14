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

  const normalizeStatus = (status) => {
    switch (status?.trim()) {
      case "รับงาน":
      case "รับงานแล้ว":
        return "รับงาน";
      case "กำลังซ่อม":
        return "กำลังซ่อม";
      case "เสร็จเรียบร้อย":
      case "เสร็จสิ้น":
        return "เสร็จสิ้น";
      default:
        return "";
    }
  };
  const visibleLogs = logs
    .map((log) => ({ ...log, displayStatus: normalizeStatus(log.new_status) }))
    .filter((log) => log.displayStatus);

  const getPublicProgress = () => {
    const chronologicalLogs = [...logs].sort(
      (left, right) => new Date(left.created_at) - new Date(right.created_at)
    );
    const acceptedLog = chronologicalLogs.find((log) => normalizeStatus(log.new_status) === "รับงาน");
    const startedLog = chronologicalLogs.find((log) => normalizeStatus(log.new_status) === "กำลังซ่อม");
    const completedLog = chronologicalLogs.find((log) => normalizeStatus(log.new_status) === "เสร็จสิ้น");
    return [
      { label: "รับงาน", log: acceptedLog, complete: Boolean(acceptedLog) },
      { label: "กำลังซ่อม", log: startedLog, complete: Boolean(startedLog) },
      { label: "เสร็จสิ้น", log: completedLog, complete: Boolean(completedLog) },
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
    }));
  };

  const paginatedLogs = getPageItems(visibleLogs, currentPage);
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
                  </div>
                </div>
              ))}
            </div>
          ) : visibleLogs.length === 0 ? (
            <div className="no-data">ไม่พบประวัติการทำรายการ</div>
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
                          {log.displayStatus}
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
                totalItems={visibleLogs.length}
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