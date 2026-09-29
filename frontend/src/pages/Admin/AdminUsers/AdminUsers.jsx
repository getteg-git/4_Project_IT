import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import "./AdminUsers.css";

function AdminUsers() {
  const navigate = useNavigate();
  const [users, setUsers] = useState([]);
  const [problemTypes, setProblemTypes] = useState([]); // ดึงข้อมูลประเภทปัญหาจาก DB
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");

  // States สำหรับ Modal
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);

  // States สำหรับฟอร์ม
  const [editId, setEditId] = useState(null);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [fullName, setFullName] = useState(""); // ฟิลด์ชื่อจริงใหม่
  const [role, setRole] = useState("technician");
  const [specialties, setSpecialties] = useState([]); // อาเรย์เก็บ ID ความถนัด เช่น [1, 2]
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [selectedUser, setSelectedUser] = useState(null);
  const [isViewOpen, setIsViewOpen] = useState(false);

  // ดึงข้อมูลหลักจาก Backend พร้อมกันแบบไร้ Mockup
  const fetchMasterData = async () => {
    try {
      setIsLoading(true);
      const [usersRes, typesRes] = await Promise.all([
        fetch("http://localhost:8080/api/users"),
        fetch("http://localhost:8080/api/problem-types")
      ]);

      if (usersRes.ok) setUsers(await usersRes.json() || []);
      if (typesRes.ok) setProblemTypes(await typesRes.json() || []);
    } catch (error) {
      console.error("Fetch Users Error:", error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchMasterData();
  }, []);

  // กรองข้อมูลจากการค้นหา (รองรับทั้งการค้นหาด้วยชื่อจริง, username หรือสิทธิ์)
  const filteredUsers = users.filter((u) =>
    u.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
    (u.full_name && u.full_name.toLowerCase().includes(searchTerm.toLowerCase())) ||
    u.role.toLowerCase().includes(searchTerm.toLowerCase())
  );

  // -----------------------------------------
  // ฟังก์ชัน Action และการทำงานต่างๆ
  // -----------------------------------------
  const resetForm = () => {
    setEditId(null);
    setUsername("");
    setPassword("");
    setFullName("");
    setRole("technician");
    setSpecialties([]);
    setIsCreateModalOpen(false);
    setIsEditModalOpen(false);
  };

  const openCreateModal = () => {
    resetForm();
    setIsCreateModalOpen(true);
  };

  const openEditModal = (user) => {
    resetForm();
    setEditId(user.id);
    setUsername(user.username);
    setFullName(user.full_name || "");
    setRole(user.role);
    setSpecialties(user.specialties || []);
    setIsEditModalOpen(true);
  };

  const openViewModal = (user) => {
    setSelectedUser(user);
    setIsViewOpen(true);
  };

  // จัดการกับการเพิ่ม/ลดความถนัดของช่างเทคนิค (ห้ามเกิน 3 รายการ)
  const handleSpecialtyChange = (typeId) => {
    if (specialties.includes(typeId)) {
      setSpecialties(specialties.filter((id) => id !== typeId));
    } else {
      if (specialties.length >= 3) {
        alert("⚠️ ช่าง 1 คน สามารถเลือกหมวดหมู่งานที่ถนัดได้สูงสุด 3 หมวดหมู่เท่านั้นครับ");
        return;
      }
      setSpecialties([...specialties, typeId]);
    }
  };

  // สร้างผู้ใช้งานใหม่ (POST)
  const handleCreateSubmit = async (e) => {
    e.preventDefault();
    if (!password || password.trim().length < 3 || password.trim().length > 30) {
      alert("⚠️ รหัสผ่านต้องมีความยาว 3-30 ตัวอักษร");
      return;
    }
    setIsSubmitting(true);
    try {
      const response = await fetch("http://localhost:8080/api/users", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          username,
          password,
          full_name: fullName,
          role,
          specialties: role === "technician" ? specialties : [] // ส่งความถนัดไปเฉพาะสิทธิ์ช่าง
        })
      });

      if (response.ok) {
        alert("✅ สร้างผู้ใช้งานสำเร็จ");
        fetchMasterData();
        resetForm();
      } else {
        const data = await response.json();
        alert(`❌ ผิดพลาด: ${data.error}`);
      }
    } catch (error) {
      alert("❌ ไม่สามารถติดต่อเซิร์ฟเวอร์ได้");
    } finally {
      setIsSubmitting(false);
    }
  };

  // แก้ไขผู้ใช้งาน (PUT) - บันทึกชื่อจริงและความถนัดชุดใหม่ได้จริง
  const handleEditSubmit = async (e) => {
    e.preventDefault();
    if (password && (password.trim().length < 3 || password.trim().length > 30)) {
      alert("⚠️ รหัสผ่านใหม่ต้องมีความยาว 3-30 ตัวอักษร");
      return;
    }

    setIsSubmitting(true);
    try {
      const response = await fetch(`http://localhost:8080/api/users/${editId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          username,
          password,
          full_name: fullName,
          specialties: role === "technician" ? specialties : []
        })
      });

      if (response.ok) {
        alert("✅ อัปเดตข้อมูลสำเร็จ");
        fetchMasterData();
        resetForm();
      } else {
        const data = await response.json();
        alert(`❌ ผิดพลาด: ${data.error}`);
      }
    } catch (error) {
      alert("❌ ไม่สามารถติดต่อเซิร์ฟเวอร์ได้");
    } finally {
      setIsSubmitting(false);
    }
  };

  // ลบผู้ใช้งาน (DELETE)
  const handleDelete = async (id, name) => {
    const confirmDelete = window.confirm(`⚠️ คุณแน่ใจหรือไม่ที่จะลบบัญชี "${name}" ออกจากระบบถาวร?`);
    if (!confirmDelete) return;

    try {
      const response = await fetch(`http://localhost:8080/api/users/${id}`, {
        method: "DELETE"
      });

      if (response.ok) {
        alert("✅ ลบผู้ใช้งานสำเร็จ");
        fetchMasterData();
      } else {
        const data = await response.json();
        alert(`❌ ผิดพลาด: ${data.error}`);
      }
    } catch (error) {
      alert("❌ ไม่สามารถติดต่อเซิร์ฟเวอร์ได้");
    }
  };

  return (
    <div className="admin-users-container">
      <div className="admin-wrapper">

        {/* Header */}
        <div className="admin-header">
          <div>
            <h2>👥 จัดการบัญชีผู้ใช้งาน (User Management)</h2>
            <p>เพิ่ม แก้ไข หรือลบบัญชีของแอดมินและช่างเทคนิค</p>
          </div>
          <div className="header-actions">
            <button className="btn-back-dashboard" onClick={() => navigate("/admin/home")}>
              📊 กลับหน้า Dashboard
            </button>
          </div>
        </div>

        {/* Toolbar: ค้นหา & เพิ่มผู้ใช้ */}
        <div className="users-toolbar">
          <div className="search-section">
            <input
              type="text"
              placeholder="🔍 ค้นหาชื่อผู้ใช้ ชื่อจริง หรือสิทธิ์..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="search-input"
            />
          </div>
          <button className="btn-create-user" onClick={openCreateModal}>
            ➕ เพิ่มผู้ใช้งานใหม่
          </button>
        </div>

        {/* ตารางรายชื่อผู้ใช้งาน */}
        <div className="users-list">
          {isLoading ? (
            <div className="loading-state">⏳ กำลังโหลดข้อมูล...</div>
          ) : filteredUsers.length > 0 ? (
            filteredUsers.map((user) => (
              <div className="user-item-card" key={user.id}>
                <div className="user-info">
                  <div className="user-avatar">
                    {user.role === "admin" ? "⚙️" : "🛠️"}
                  </div>
                  <div className="user-details-block">
                    {/* แสดงชื่อจริงภาษาไทยเป็นหลัก และต่อท้ายด้วยบัญชีล็อกอิน */}
                    <h3 className="user-name">
                      {user.full_name} <span className="user-username-label">(@{user.username})</span>
                    </h3>

                    <span className={`role-badge ${user.role === "admin" ? "role-admin" : "role-tech"}`}>
                      {user.role === "admin" ? "แอดมิน (Admin)" : "ช่างเทคนิค (Technician)"}
                    </span>

                    {/* แสดงป้ายความถนัดของช่างเทคนิคดึงสดจากระบบหลังบ้าน */}
                    {user.role === "technician" && user.specialty_names && user.specialty_names.length > 0 && (
                      <div className="specialty-container">
                        {user.specialty_names.map((name, index) => (
                          <span className="specialty-tag" key={index}>⚡ {name}</span>
                        ))}
                      </div>
                    )}
                  </div>
                </div>

                <div className="user-actions">
                  <button
                    className="btn-view"
                    onClick={() => openViewModal(user)}
                  >
                    👁️ ดูข้อมูล
                  </button>

                  <button
                    className="btn-edit"
                    onClick={() => openEditModal(user)}
                  >
                    ✏️ แก้ไข
                  </button>

                  <button
                    className="btn-delete"
                    onClick={() => handleDelete(user.id, user.full_name)}
                  >
                    🗑️ ลบ
                  </button>
                </div>
              </div>
            ))
          ) : (
            <div className="no-results">
              <p>❌ ไม่พบข้อมูลบัญชีผู้ใช้งาน</p>
            </div>
          )}
        </div>
      </div>

      {/* ==========================================
          POPUP MODAL: สร้างผู้ใช้งานใหม่
          ========================================== */}
      {isCreateModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <span className="close-btn" onClick={resetForm}>&times;</span>
            <h3 className="modal-title">➕ สร้างบัญชีผู้ใช้ใหม่</h3>

            <form onSubmit={handleCreateSubmit}>
              <div className="input-group">
                <label>สิทธิ์การใช้งาน <span className="required">*</span></label>
                <select value={role} onChange={(e) => setRole(e.target.value)} required>
                  <option value="technician">🛠️ ช่างเทคนิค</option>
                  <option value="admin">⚙️ แอดมิน</option>
                </select>
              </div>

              <div className="input-group">
                <label>ชื่อ-นามสกุลจริง <span className="required">*</span></label>
                <input
                  type="text"
                  placeholder="ระบุชื่อและนามสกุลช่าง/แอดมิน (เช่น นายสมชาย ยอดช่าง)"
                  value={fullName}
                  onChange={(e) => setFullName(e.target.value)}
                  required
                />
              </div>

              {/* แสดงโซนระบุความถนัดเมื่อสิทธิ์ที่เลือกเป็นช่างเทคนิค */}
              {role === "technician" && (
                <div className="input-group">
                  <label>หมวดหมู่งานซ่อมที่ถนัด (เลือกได้สูงสุด 3 หมวดหมู่) <span className="required">*</span></label>
                  <div className="specialties-checkbox-grid">
                    {problemTypes.map((type) => (
                      <label className="checkbox-item-label" key={type.id}>
                        <input
                          type="checkbox"
                          checked={specialties.includes(type.id)}
                          onChange={() => handleSpecialtyChange(type.id)}
                        />
                        {type.name}
                      </label>
                    ))}
                  </div>
                </div>
              )}

              <div className="input-group">
                <label>ชื่อผู้ใช้งาน (Username สำหรับใช้ Login) <span className="required">*</span></label>
                <input
                  type="text"
                  placeholder="ตั้งชื่อล็อกอินภาษาอังกฤษ (เช่น tech_somch)"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  required
                />
              </div>

              <div className="input-group">
                <label>รหัสผ่าน (Password) <span className="required">*</span></label>
                <input
                  type="password"
                  placeholder="ตั้งรหัสผ่านเบื้องต้น"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
              </div>

              <div className="modal-actions">
                <button type="button" className="btn-cancel" onClick={resetForm} disabled={isSubmitting}>ยกเลิก</button>
                <button type="submit" className="btn-submit" disabled={isSubmitting}>
                  {isSubmitting ? "⏳ กำลังบันทึก..." : "💾 ยืนยันสร้างบัญชี"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ==========================================
          POPUP MODAL: แก้ไขผู้ใช้งาน
          ========================================== */}
      {isEditModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <span className="close-btn" onClick={resetForm}>&times;</span>
            <h3 className="modal-title">✏️ แก้ไขข้อมูลบัญชี</h3>

            <form onSubmit={handleEditSubmit}>
              <div className="input-group">
                <label>ชื่อ-นามสกุลจริง <span className="required">*</span></label>
                <input
                  type="text"
                  value={fullName}
                  onChange={(e) => setFullName(e.target.value)}
                  required
                />
              </div>

              <div className="input-group">
                <label>ชื่อผู้ใช้งาน (Username) <span className="required">*</span></label>
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  required
                />
              </div>

              {/* แก้ไขความถนัดของช่างเทคนิค */}
              {role === "technician" && (
                <div className="input-group">
                  <label>ปรับปรุงงานซ่อมที่ถนัด (สูงสุด 3 หมวดหมู่) <span className="required">*</span></label>
                  <div className="specialties-checkbox-grid">
                    {problemTypes.map((type) => (
                      <label className="checkbox-item-label" key={type.id}>
                        <input
                          type="checkbox"
                          checked={specialties.includes(type.id)}
                          onChange={() => handleSpecialtyChange(type.id)}
                        />
                        {type.name}
                      </label>
                    ))}
                  </div>
                </div>
              )}

              <div className="input-group">
                <label>รีเซ็ตรหัสผ่านใหม่ (เว้นว่างไว้หากไม่ต้องการเปลี่ยน)</label>
                <input
                  type="password"
                  placeholder="ระบุรหัสผ่านใหม่หากต้องการเปลี่ยน (3-30 ตัว)"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
                <small className="password-warning-text">
                  * หากปล่อยว่างไว้ ระบบจะคงรหัสผ่านเดิมไว้ และถ้ากรอกใหม่ต้องอยู่ในช่วง 3-30 ตัว
                </small>
              </div>

              <div className="modal-actions">
                <button type="button" className="btn-cancel" onClick={resetForm} disabled={isSubmitting}>ยกเลิก</button>
                <button type="submit" className="btn-submit" disabled={isSubmitting}>
                  {isSubmitting ? "⏳ กำลังบันทึก..." : "💾 บันทึกการเปลี่ยนแปลง"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ==========================================
    POPUP MODAL: ดูข้อมูลผู้ใช้งาน
    ========================================== */}
      {isViewOpen && selectedUser && (
        <div className="modal-overlay">
          <div className="modal-box user-view-modal">

            <span
              className="close-btn"
              onClick={() => {
                setIsViewOpen(false);
                setSelectedUser(null);
              }}
            >
              &times;
            </span>

            <h3 className="modal-title">👤 ข้อมูลผู้ใช้งาน</h3>

            <div className="user-view-content">

              <div className="user-view-avatar">
                {selectedUser.role === "admin" ? "⚙️" : "🛠️"}
              </div>

              <div className="user-view-row">
                <strong>ชื่อ-นามสกุล</strong>
                <span>{selectedUser.full_name || "-"}</span>
              </div>

              <div className="user-view-row">
                <strong>Username</strong>
                <span>@{selectedUser.username || "-"}</span>
              </div>

              <div className="user-view-row">
                <strong>สิทธิ์</strong>
                <span>
                  {selectedUser.role === "admin"
                    ? "แอดมิน (Admin)"
                    : "ช่างเทคนิค (Technician)"}
                </span>
              </div>

              {selectedUser.email && (
                <div className="user-view-row">
                  <strong>อีเมล</strong>
                  <span>{selectedUser.email}</span>
                </div>
              )}

              {selectedUser.department_name && (
                <div className="user-view-row">
                  <strong>สาขา</strong>
                  <span>{selectedUser.department_name}</span>
                </div>
              )}

              {selectedUser.role === "technician" && (
                <div className="user-view-row">
                  <strong>ความเชี่ยวชาญ</strong>

                  <div className="user-view-specialties">
                    {selectedUser.specialty_names?.length > 0
                      ? selectedUser.specialty_names.map((name, index) => (
                        <span className="specialty-tag" key={index}>
                          ⚡ {name}
                        </span>
                      ))
                      : <span>- ยังไม่ได้ระบุ -</span>}
                  </div>
                </div>
              )}

            </div>

            <div className="modal-actions">
              <button
                type="button"
                className="btn-cancel"
                onClick={() => {
                  setIsViewOpen(false);
                  setSelectedUser(null);
                }}
              >
                ปิด
              </button>

              <button
                type="button"
                className="btn-submit"
                onClick={() => {
                  setIsViewOpen(false);
                  openEditModal(selectedUser);
                }}
              >
                ✏️ แก้ไขข้อมูล
              </button>
            </div>

          </div>
        </div>
      )}

    </div>
  );

}

export default AdminUsers;