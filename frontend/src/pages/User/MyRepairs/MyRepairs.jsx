import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import api from "../../../api/axios";
import "./MyRepairs.css";

function MyRepairs() {
  const [repairs, setRepairs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  
  // State สำหรับจัดการ Modal (Popup)
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedRepair, setSelectedRepair] = useState(null);
  const [modalLoading, setModalLoading] = useState(false);

  const navigate = useNavigate();

  useEffect(() => {
    fetchMyRepairs();
  }, []);

  const fetchMyRepairs = async () => {
    try {
      const res = await api.get("/repairs");
      const currentUser = JSON.parse(localStorage.getItem("user"));
      const currentUsername = currentUser ? currentUser.username : "GETTEG";

      const myData = res.data.filter(
        (repair) => repair.requester === currentUsername
      );

      setRepairs(myData);
      setLoading(false);
    } catch (err) {
      console.error(err);
      setError("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์เพื่อดึงข้อมูลได้");
      setLoading(false);
    }
  };

  // ฟังก์ชันเมื่อกดปุ่ม "ดูรายละเอียด"
  const handleViewDetails = async (id) => {
    setIsModalOpen(true); // เปิด Modal ทันที
    setModalLoading(true); // โชว์สถานะกำลังโหลดใน Modal
    
    try {
      // ยิง API ดึงข้อมูลรายตัว (ซึ่งมีรูปภาพติดมาด้วย)
      const res = await api.get(`/repairs/${id}`);
      setSelectedRepair(res.data);
    } catch (err) {
      console.error(err);
      alert("ไม่สามารถดึงข้อมูลรายละเอียดได้");
    } finally {
      setModalLoading(false);
    }
  };

  const closeModal = () => {
    setIsModalOpen(false);
    setSelectedRepair(null);
  };

  const getStatusBadge = (status) => {
    switch (status) {
      case "pending":
        return <span className="badge badge-pending">รอดำเนินการ</span>;
      case "approved":
        return <span className="badge badge-approved">รับเรื่องแล้ว (รอช่าง)</span>;
      case "in_progress":
        return <span className="badge badge-progress">กำลังซ่อม 🛠️</span>;
      case "done":
        return <span className="badge badge-done">เสร็จสิ้น ✅</span>;
      default:
        return <span className="badge">{status}</span>;
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return "-";
    const options = { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' };
    return new Date(dateString).toLocaleDateString('th-TH', options);
  };

  // Base URL ของ Backend สำหรับแสดงรูปภาพ
  const BACKEND_URL = "http://localhost:8080";

  return (
    <div className="my-repairs-container">
      <div className="header-action">
        <button className="back-btn" onClick={() => navigate("/user/home")}>
          &larr; กลับหน้าหลัก
        </button>
        <h2>รายการแจ้งซ่อมของฉัน</h2>
      </div>

      {loading ? (
        <p className="loading-text">กำลังโหลดข้อมูล... ⏳</p>
      ) : error ? (
        <p className="error-message">{error}</p>
      ) : repairs.length === 0 ? (
        <div className="empty-state">
          <p>คุณยังไม่มีประวัติการแจ้งซ่อม</p>
          <button onClick={() => navigate("/repair/create")} className="new-repair-btn">
            + แจ้งซ่อมใหม่
          </button>
        </div>
      ) : (
        <div className="repair-list">
          {repairs.map((repair) => (
            <div className="repair-card" key={repair.id}>
              <div className="card-header">
                <span className="repair-id">Ticket #{repair.id}</span>
                {getStatusBadge(repair.status)}
              </div>

              <div className="card-body">
                <h3 className="location-title">{repair.location}</h3>
                <p className="problem-type">
                  <strong>ประเภทปัญหา:</strong> {repair.problem_type}
                </p>
                <p className="description">
                  <strong>รายละเอียด:</strong> {repair.description}
                </p>
              </div>

              {/* --- เปลี่ยนโครงสร้างส่วนล่างของการ์ด --- */}
              <div className="card-footer">
                <span className="date-text">
                  แจ้งเมื่อ: {formatDate(repair.created_at)}
                </span>
              </div>
              
              <div className="card-action-center">
                <button 
                  className="view-btn" 
                  onClick={() => handleViewDetails(repair.id)}
                >
                  ดูรายละเอียด / รูปภาพ
                </button>
              </div>
              {/* -------------------------------------- */}
            </div>
          ))}
        </div>
      )}

      {/* --- ส่วนของ Modal (Popup) --- */}
      {isModalOpen && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            
            <div className="modal-header">
              <h3>รายละเอียดการแจ้งซ่อม</h3>
              <button className="close-btn" onClick={closeModal}>&times;</button>
            </div>

            <div className="modal-body">
              {modalLoading ? (
                <p className="loading-text">กำลังโหลดรูปภาพ... ⏳</p>
              ) : selectedRepair ? (
                <>
                  <p><strong>สถานที่:</strong> {selectedRepair.location}</p>
                  <p><strong>ประเภทปัญหา:</strong> {selectedRepair.problem_type}</p>
                  <p><strong>รายละเอียด:</strong> {selectedRepair.description}</p>
                  
                  <hr className="modal-divider" />
                  
                  <h4>รูปภาพประกอบ (ตอนแจ้งซ่อม)</h4>
                  {selectedRepair.images && selectedRepair.images.filter(img => img.type === 'before').length > 0 ? (
                    <div className="modal-images">
                      {selectedRepair.images.filter(img => img.type === 'before').map((img, idx) => (
                        <img 
                          key={idx} 
                          src={`${BACKEND_URL}${img.url}`} 
                          alt="Before Repair" 
                          className="repair-img"
                        />
                      ))}
                    </div>
                  ) : (
                    <p className="no-image-text">- ไม่ได้แนบรูปภาพ -</p>
                  )}

                  {/* เผื่ออนาคตช่างแนบรูปตอนซ่อมเสร็จ จะได้มาโชว์ตรงนี้ด้วยเลย! */}
                  {selectedRepair.images && selectedRepair.images.filter(img => img.type === 'during').length > 0 && (
                    <>
                      <h4 style={{ marginTop: '20px', color: '#a21caf' }}>รูประหว่างซ่อม (จากช่าง)</h4>
                      <div className="modal-images">
                        {selectedRepair.images.filter(img => img.type === 'during').map((img, idx) => (
                          <img 
                            key={idx} 
                            src={`${BACKEND_URL}${img.url}`} 
                            alt="During Repair" 
                            className="repair-img tech-img"
                          />
                        ))}
                      </div>
                    </>
                  )}
                </>
              ) : (
                <p>ไม่พบข้อมูล</p>
              )}
            </div>

          </div>
        </div>
      )}

    </div>
  );
}

export default MyRepairs;