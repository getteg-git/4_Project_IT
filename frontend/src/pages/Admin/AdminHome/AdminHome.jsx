import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import api from "../../../api/axios";
import "./AdminHome.css";

function AdminHome() {
  const [repairs, setRepairs] = useState([]);
  const [technicians, setTechnicians] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // State สำหรับจัดการ Modal (Popup จ่ายงาน)
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedRepair, setSelectedRepair] = useState(null);
  const [selectedTechId, setSelectedTechId] = useState("");

  const navigate = useNavigate();

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      // ดึงข้อมูล 2 อย่างพร้อมกัน เพื่อความรวดเร็ว
      const [repairsRes, usersRes] = await Promise.all([
        api.get("/repairs"),
        api.get("/users")
      ]);

      // เก็บรายการแจ้งซ่อมทั้งหมด (สำหรับแอดมินต้องเห็นหมด)
      setRepairs(repairsRes.data);

      // กรองเอาเฉพาะ User ที่เป็นช่างเทคนิค
      const techList = usersRes.data.filter(user => user.role === "technician");
      setTechnicians(techList);

      setLoading(false);
    } catch (err) {
      console.error(err);
      setError("ไม่สามารถดึงข้อมูลได้ กรุณาลองใหม่อีกครั้ง");
      setLoading(false);
    }
  };

  // เปิด Popup จ่ายงาน
  const openAssignModal = (repair) => {
    setSelectedRepair(repair);
    setSelectedTechId(""); // รีเซ็ตค่า dropdown
    setIsModalOpen(true);
  };

  // ปิด Popup
  const closeModal = () => {
    setIsModalOpen(false);
    setSelectedRepair(null);
  };

  // ฟังก์ชันกดยืนยันจ่ายงาน
  const handleAssignTask = async () => {
    if (!selectedTechId) {
      alert("กรุณาเลือกช่างเทคนิคก่อนกดยืนยัน");
      return;
    }

    try {
      // ยิง API ไปเปลี่ยนสถานะและใส่ช่างเทคนิค
      await api.put(`/repairs/${selectedRepair.id}/approve`, {
        technician_id: parseInt(selectedTechId)
      });

      // ถ้าสำเร็จ ให้ดึงข้อมูลใหม่ เพื่ออัปเดตตาราง
      await fetchData();
      closeModal();
      
    } catch (err) {
      console.error("Assign error:", err);
      alert("เกิดข้อผิดพลาดในการมอบหมายงาน (ช่างอาจถูกลบไปแล้ว?)");
    }
  };

  // จัดการป้ายสถานะ
  const getStatusBadge = (status) => {
    switch (status) {
      case "pending": return <span className="badge badge-pending">รอดำเนินการ</span>;
      case "approved": return <span className="badge badge-approved">รอช่างรับงาน</span>;
      case "in_progress": return <span className="badge badge-progress">กำลังซ่อม</span>;
      case "done": return <span className="badge badge-done">เสร็จสิ้น</span>;
      default: return <span className="badge">{status}</span>;
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return "-";
    const options = { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' };
    return new Date(dateString).toLocaleDateString('th-TH', options);
  };

  return (
    <div className="admin-home-container">
      
      <div className="admin-header">
        <h2>🛠️ Admin Dashboard (ระบบจัดการงานซ่อม)</h2>
        {/* เผื่อทำระบบ Logout ในอนาคต พี่ใส่ปุ่มหลอกๆ ไว้ให้ก่อน */}
        <button className="logout-btn" onClick={() => navigate("/login")}>ออกจากระบบ</button>
      </div>

      {loading ? (
        <p className="loading-text">กำลังโหลดข้อมูลระบบ... ⏳</p>
      ) : error ? (
        <p className="error-message">{error}</p>
      ) : (
        <div className="table-wrapper">
          <table className="admin-table">
            <thead>
              <tr>
                <th>Ticket ID</th>
                <th>ผู้แจ้ง</th>
                <th>สถานที่</th>
                <th>ปัญหา</th>
                <th>สถานะ</th>
                <th>วันที่แจ้ง</th>
                <th>การจัดการ</th>
              </tr>
            </thead>
            <tbody>
              {repairs.length === 0 ? (
                <tr><td colSpan="7" className="empty-text">ไม่มีรายการแจ้งซ่อมในระบบ</td></tr>
              ) : (
                repairs.map((repair) => (
                  <tr key={repair.id}>
                    <td>#{repair.id}</td>
                    <td>{repair.requester || "Unknown"}</td>
                    <td>{repair.location}</td>
                    <td>{repair.problem_type}</td>
                    <td>{getStatusBadge(repair.status)}</td>
                    <td>{formatDate(repair.created_at)}</td>
                    <td>
                      {/* ถ้าสถานะเป็น pending ถึงจะให้กดจ่ายงานได้ */}
                      {repair.status === "pending" ? (
                        <button 
                          className="assign-btn"
                          onClick={() => openAssignModal(repair)}
                        >
                          จ่ายงาน
                        </button>
                      ) : (
                        <span className="assigned-text">
                          {repair.status === "done" ? "จบงานแล้ว" : "มอบหมายแล้ว"}
                        </span>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* --- Modal (Popup) สำหรับเลือกช่าง --- */}
      {isModalOpen && selectedRepair && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>มอบหมายงานซ่อม Ticket #{selectedRepair.id}</h3>
              <button className="close-btn" onClick={closeModal}>&times;</button>
            </div>
            
            <div className="modal-body">
              <div className="repair-info-box">
                <p><strong>สถานที่:</strong> {selectedRepair.location}</p>
                <p><strong>ปัญหา:</strong> {selectedRepair.problem_type}</p>
                <p><strong>รายละเอียด:</strong> {selectedRepair.description}</p>
              </div>

              <div className="assign-form">
                <label>เลือกช่างเทคนิคที่ต้องการมอบหมาย:</label>
                <select 
                  value={selectedTechId} 
                  onChange={(e) => setSelectedTechId(e.target.value)}
                >
                  <option value="">-- กรุณาเลือกช่าง --</option>
                  {technicians.map((tech) => (
                    <option key={tech.id} value={tech.id}>
                      {tech.username} (ID: {tech.id})
                    </option>
                  ))}
                </select>

                {technicians.length === 0 && (
                  <p className="error-message" style={{marginTop: "10px", fontSize: "12px"}}>
                    * ไม่พบข้อมูลช่างในระบบ กรุณาไปเพิ่ม User ที่มี Role: technician ก่อน
                  </p>
                )}
              </div>
            </div>

            <div className="modal-footer">
              <button className="cancel-assign-btn" onClick={closeModal}>ยกเลิก</button>
              <button className="confirm-assign-btn" onClick={handleAssignTask}>ยืนยันการจ่ายงาน</button>
            </div>
          </div>
        </div>
      )}

    </div>
  );
}

export default AdminHome;