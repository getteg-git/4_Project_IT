import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import api from "../../../api/axios";
import "./TechHome.css";

function TechHome() {
    const [repairs, setRepairs] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");

    // State สำหรับ Modal ปิดงานและอัปโหลดรูป
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [selectedRepair, setSelectedRepair] = useState(null);
    const [images, setImages] = useState([]);

    const navigate = useNavigate();

    useEffect(() => {
        fetchMyTasks();
    }, []);

    const fetchMyTasks = async () => {
        try {
            const res = await api.get("/repairs");
            const currentUser = JSON.parse(localStorage.getItem("user"));
            const currentTechId = currentUser ? currentUser.id : null;

            // กรองเอาเฉพาะงานที่ technician_id ตรงกับ ID ของช่างที่ล็อกอินอยู่
            // และไม่เอางานที่ status เป็น 'pending' (เพราะแอดมินยังไม่จ่ายงาน)
            const myTasks = res.data.filter(
                (repair) => repair.technician_id === currentTechId && repair.status !== "pending"
            );

            setRepairs(myTasks);
            setLoading(false);
        } catch (err) {
            console.error(err);
            setError("ไม่สามารถดึงข้อมูลงานได้");
            setLoading(false);
        }
    };

    // ฟังก์ชันกด "รับงาน" (เปลี่ยนสถานะเป็น In Progress)
    // ฟังก์ชันกด "รับงาน" (เปลี่ยนสถานะเป็น In Progress)
    const handleStartWork = async (id) => {
        try {
            // เปลี่ยนเส้นทางให้ตรงกับ r.PUT("/repairs/:id/progress", ...) ใน Go
            await api.put(`/repairs/${id}/progress`);
            fetchMyTasks(); // โหลดข้อมูลใหม่
        } catch (err) {
            console.error(err);
            alert("เกิดข้อผิดพลาดในการรับงาน");
        }
    };

    // ฟังก์ชันเปิด Popup เตรียม "ปิดงาน"
    const openCompleteModal = (repair) => {
        setSelectedRepair(repair);
        setImages([]); // เคลียร์รูปเก่า
        setIsModalOpen(true);
    };

    const closeCompleteModal = () => {
        setIsModalOpen(false);
        setSelectedRepair(null);
    };

    // จัดการเลือกรูปภาพ
    const handleImageChange = (e) => {
        const files = Array.from(e.target.files);
        setImages((prev) => [...prev, ...files]);
    };

    const removeImage = (index) => {
        setImages(images.filter((_, i) => i !== index));
    };

    // ฟังก์ชันยืนยันการ "ปิดงาน" พร้อมส่งรูป
    const handleCompleteWork = async () => {
        try {
            const formData = new FormData();
            formData.append("status", "done"); // ส่งสถานะว่าเสร็จแล้ว

            // แนบรูประหว่างซ่อม/ซ่อมเสร็จ
            images.forEach((img) => {
                formData.append("images", img);
            });

            // ยิง API อัปเดตงาน (ใช้ endpoint ที่รับ FormData ได้)
            await api.put(`/repairs/${selectedRepair.id}/complete`, formData, {
                headers: { "Content-Type": "multipart/form-data" }
            });

            closeCompleteModal();
            fetchMyTasks(); // โหลดข้อมูลใหม่
        } catch (err) {
            console.error(err);
            alert("เกิดข้อผิดพลาดในการปิดงาน");
        }
    };

    const getStatusBadge = (status) => {
        switch (status) {
            case "approved": return <span className="badge badge-approved">งานใหม่ (รอรับ)</span>;
            case "in_progress": return <span className="badge badge-progress">กำลังดำเนินการ</span>;
            case "done": return <span className="badge badge-done">ปิดงานแล้ว</span>;
            default: return <span className="badge">{status}</span>;
        }
    };

    return (
        <div className="tech-home-container">
            <div className="tech-header">
                <h2>🔧 Technician Dashboard</h2>
                <button className="logout-btn" onClick={() => navigate("/login")}>ออกจากระบบ</button>
            </div>

            {loading ? (
                <p className="loading-text">กำลังโหลดงานของคุณ... ⏳</p>
            ) : error ? (
                <p className="error-message">{error}</p>
            ) : repairs.length === 0 ? (
                <div className="empty-state">
                    <p>ตอนนี้คุณยังไม่มีงานที่ได้รับมอบหมายครับ ไปพักผ่อนได้เลย! ☕</p>
                </div>
            ) : (
                <div className="task-grid">
                    {repairs.map((repair) => (
                        <div className={`task-card ${repair.status === 'done' ? 'task-done' : ''}`} key={repair.id}>
                            <div className="task-header">
                                <span className="task-id">Ticket #{repair.id}</span>
                                {getStatusBadge(repair.status)}
                            </div>

                            <div className="task-body">
                                <h3>{repair.location}</h3>
                                <p><strong>ปัญหา:</strong> {repair.problem_type}</p>
                                <p><strong>รายละเอียด:</strong> {repair.description}</p>
                            </div>

                            <div className="task-footer">
                                {/* ปุ่มรับงาน (โชว์ตอนแอดมินเพิ่งจ่ายงานมา) */}
                                {repair.status === "approved" && (
                                    <button className="start-btn" onClick={() => handleStartWork(repair.id)}>
                                        ▶️ กดรับงาน
                                    </button>
                                )}

                                {/* ปุ่มปิดงาน (โชว์ตอนกำลังซ่อม) */}
                                {repair.status === "in_progress" && (
                                    <button className="complete-btn" onClick={() => openCompleteModal(repair)}>
                                        ✅ แจ้งซ่อมเสร็จสิ้น
                                    </button>
                                )}

                                {/* ถ้าง่ายเสร็จแล้ว */}
                                {repair.status === "done" && (
                                    <span className="done-text">งานนี้เสร็จสมบูรณ์แล้ว 🎉</span>
                                )}
                            </div>
                        </div>
                    ))}
                </div>
            )}

            {/* --- Modal สำหรับปิดงานและอัปโหลดรูป --- */}
            {isModalOpen && selectedRepair && (
                <div className="modal-overlay" onClick={closeCompleteModal}>
                    <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                        <div className="modal-header">
                            <h3>ปิดงานซ่อม: Ticket #{selectedRepair.id}</h3>
                            <button className="close-btn" onClick={closeCompleteModal}>&times;</button>
                        </div>

                        <div className="modal-body">
                            <p className="instruction-text">กรุณาแนบรูปภาพหลังซ่อมเสร็จ เพื่อเป็นหลักฐานให้ผู้แจ้งและแอดมินตรวจสอบ</p>

                            <input
                                type="file"
                                multiple
                                accept="image/*"
                                onChange={handleImageChange}
                                className="file-input"
                            />

                            <div className="preview-container">
                                {images.map((img, index) => (
                                    <div key={index} className="preview-item">
                                        <img src={URL.createObjectURL(img)} alt="preview" />
                                        <button type="button" className="remove-btn" onClick={() => removeImage(index)}>ลบ</button>
                                    </div>
                                ))}
                            </div>
                        </div>

                        <div className="modal-footer">
                            <button className="cancel-btn" onClick={closeCompleteModal}>ยกเลิก</button>
                            <button className="confirm-btn" onClick={handleCompleteWork} disabled={images.length === 0}>
                                บันทึกการปิดงาน
                            </button>
                        </div>
                    </div>
                </div>
            )}

        </div>
    );
}

export default TechHome;