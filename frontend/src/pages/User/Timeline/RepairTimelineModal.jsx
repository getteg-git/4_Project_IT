import React, { useState, useEffect } from "react";
import { X, Clock, CheckCircle2, AlertCircle, PlayCircle, Ban } from "lucide-react";
import "./RepairTimelineModal.css"; // เดี๋ยว CSS เราค่อยทำทีหลัง

function RepairTimelineModal({ repairId, isOpen, onClose }) {
  const [logs, setLogs] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!isOpen || !repairId) return;

    const fetchLogs = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const response = await fetch(`http://localhost:8080/api/repairs/${repairId}/logs`);
        if (response.ok) {
          const data = await response.json();
          setLogs(data || []);
        } else {
          setError("ไม่สามารถดึงข้อมูลประวัติได้");
        }
      } catch (err) {
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
      case "REVOKED": 
      case "REJECTED":
      case "CANCELLED": return { icon: <Ban size={16} />, color: "text-red", bg: "bg-red" };
      default: return { icon: <Clock size={16} />, color: "text-gray", bg: "bg-gray" };
    }
  };

  return (
    <div className="modal-overlay">
      <div className="modal-box timeline-box">
        <button className="close-btn" onClick={onClose} aria-label="ปิด">
          <X aria-hidden="true" />
        </button>
        
        <h3>ประวัติการดำเนินการ (Timeline)</h3>
        <p className="ticket-subtitle">Ticket ID: #{repairId}</p>
        <hr />

        <div className="timeline-content">
          {isLoading ? (
            <div className="loading-state">กำลังโหลดประวัติ...</div>
          ) : error ? (
            <div className="error-state">{error}</div>
          ) : logs.length === 0 ? (
            <div className="no-data">ไม่พบประวัติการทำรายการ</div>
          ) : (
            <div className="timeline-container">
              {logs.map((log, index) => {
                const config = getActionConfig(log.action);
                return (
                  <div key={log.id} className="timeline-item">
                    {/* เส้นเชื่อม (ยกเว้นอันสุดท้าย) */}
                    {index !== logs.length - 1 && <div className="timeline-line"></div>}
                    
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
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default RepairTimelineModal;