import React, { useState, useEffect } from "react";
import { CircleX, Pencil, Plus, Save, ShieldCheck, Trash2, UserPlus, UsersRound, Wrench, X, Zap, Mail, Building2, Globe } from "lucide-react";
import SearchField from "../../../../../components/ui/SearchField";
import Pagination from "../../../../../components/ui/Pagination";
import { getPageItems } from "../../../../../components/ui/paginationUtils";
import useToast from "../../../../../hooks/useToast";
import "./AdminUsers.css";

function AdminUsers() {
  const { toast, confirm } = useToast();
  const [users, setUsers] = useState([]);
  const [problemTypes, setProblemTypes] = useState([]); 
  const [departments, setDepartments] = useState([]); // 🔥 เพิ่ม State เก็บสาขาวิชา
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [currentPage, setCurrentPage] = useState(1);

  // States สำหรับ Modal
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);

  // States สำหรับฟอร์ม
  const [editId, setEditId] = useState(null);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [fullName, setFullName] = useState(""); 
  const [email, setEmail] = useState(""); // 🔥 เพิ่มฟิลด์อีเมล
  const [departmentId, setDepartmentId] = useState(""); // 🔥 เพิ่มฟิลด์สาขา
  const [role, setRole] = useState("technician"); 
  const [isCentral, setIsCentral] = useState(false); // 🔥 เพิ่มฟิลด์ช่างส่วนกลาง
  const [specialties, setSpecialties] = useState([]); 
  const [isSubmitting, setIsSubmitting] = useState(false);

  // ดึงข้อมูลหลักจาก Backend พร้อมกัน
  const fetchMasterData = async () => {
    try {
      setIsLoading(true);
      const [usersRes, typesRes, deptsRes] = await Promise.all([
        fetch("http://localhost:8080/api/users"),
        fetch("http://localhost:8080/api/problem-types"),
        fetch("http://localhost:8080/api/departments") // 🔥 ดึงข้อมูลสาขาวิชา
      ]);
      
      if (usersRes.ok) setUsers(await usersRes.json() || []);
      if (typesRes.ok) setProblemTypes(await typesRes.json() || []);
      if (deptsRes.ok) setDepartments(await deptsRes.json() || []);
    } catch (error) {
      console.error("Fetch Users Error:", error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchMasterData();
  }, []);

  const filteredUsers = users.filter((u) =>
    u.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
    (u.full_name && u.full_name.toLowerCase().includes(searchTerm.toLowerCase())) ||
    u.role.toLowerCase().includes(searchTerm.toLowerCase()) ||
    (u.email && u.email.toLowerCase().includes(searchTerm.toLowerCase()))
  );
  const paginatedUsers = getPageItems(filteredUsers, currentPage);

  // -----------------------------------------
  // ฟังก์ชัน Action และการทำงานต่างๆ
  // -----------------------------------------
  const resetForm = () => {
    setEditId(null);
    setUsername("");
    setPassword("");
    setFullName("");
    setEmail("");
    setDepartmentId("");
    setRole("technician");
    setIsCentral(false);
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
    setEmail(user.email || "");
    setDepartmentId(user.department_id || "");
    setRole(user.role);
    setIsCentral(user.is_central || false);
    setSpecialties(user.specialties || []);
    setIsEditModalOpen(true);
  };

  const handleSpecialtyChange = (typeId) => {
    if (specialties.includes(typeId)) {
      setSpecialties(specialties.filter((id) => id !== typeId));
    } else {
      if (specialties.length >= 3) {
        toast.warning("เลือกความถนัดได้สูงสุด 3 รายการ", { description: "ยกเลิกตัวเลือกเดิมก่อนเลือกหมวดหมู่เพิ่ม" });
        return;
      }
      setSpecialties([...specialties, typeId]);
    }
  };

  const handleCreateSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      const response = await fetch("http://localhost:8080/api/users", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ 
          username, 
          password, 
          full_name: fullName, 
          email,
          department_id: departmentId ? parseInt(departmentId) : null,
          is_central: role === "technician" ? isCentral : false,
          role,
          specialties: role === "technician" ? specialties : []
        })
      });

      if (response.ok) {
        toast.success("สร้างผู้ใช้งานสำเร็จ", { description: "เพิ่มบัญชีเข้าสู่ระบบแล้ว" });
        fetchMasterData();
        resetForm();
      } else {
        const data = await response.json();
        toast.error("สร้างผู้ใช้งานไม่สำเร็จ", { description: data.error || "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch {
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleEditSubmit = async (e) => {
    e.preventDefault();
    if (!password) {
      toast.warning("กรุณาระบุรหัสผ่าน", { description: "กำหนดรหัสผ่านใหม่ หรือกรอกรหัสเดิมเพื่อยืนยันการแก้ไข" });
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
          email,
          department_id: departmentId ? parseInt(departmentId) : null,
          is_central: role === "technician" ? isCentral : false,
          specialties: role === "technician" ? specialties : []
        })
      });

      if (response.ok) {
        toast.success("อัปเดตข้อมูลสำเร็จ", { description: "บันทึกข้อมูลผู้ใช้งานแล้ว" });
        fetchMasterData();
        resetForm();
      } else {
        const data = await response.json();
        toast.error("อัปเดตข้อมูลไม่สำเร็จ", { description: data.error || "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch {
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async (id, name) => {
    const confirmDelete = await confirm({
      title: "ลบบัญชีผู้ใช้",
      description: `บัญชี “${name}” จะถูกลบออกจากระบบถาวร และไม่สามารถกู้คืนได้`,
      confirmLabel: "ลบบัญชี",
      variant: "danger",
    });
    if (!confirmDelete) return;

    try {
      const response = await fetch(`http://localhost:8080/api/users/${id}`, {
        method: "DELETE"
      });

      if (response.ok) {
        toast.success("ลบผู้ใช้งานสำเร็จ", { description: "นำบัญชีออกจากระบบแล้ว" });
        fetchMasterData();
      } else {
        const data = await response.json();
        toast.error("ลบผู้ใช้งานไม่สำเร็จ", { description: data.error || "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch {
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    }
  };

  return (
    <div className="admin-users-container" style={{ padding: 0 }}>
      <div className="admin-wrapper">
        
        {/* Toolbar: ค้นหา & เพิ่มผู้ใช้ */}
        <div className="users-toolbar">
          <div className="search-section">
            <SearchField
              id="admin-user-search"
              label="ค้นหาบัญชีผู้ใช้"
              placeholder="ชื่อผู้ใช้ ชื่อจริง อีเมล หรือสิทธิ์"
              value={searchTerm}
              onChange={(value) => { setSearchTerm(value); setCurrentPage(1); }}
            />
          </div>
          <button className="btn-create-user" onClick={openCreateModal}>
            <UserPlus size={18} aria-hidden="true" /> เพิ่มผู้ใช้งานใหม่
          </button>
        </div>

        {/* ตารางรายชื่อผู้ใช้งาน */}
        <div className="users-list">
          {isLoading ? (
            <div className="loading-state">กำลังโหลดข้อมูล...</div>
          ) : filteredUsers.length > 0 ? (
            paginatedUsers.map((user) => (
              <div className="user-item-card" key={user.id}>
                <div className="user-info">
                  <div className="user-avatar">
                    {user.role === "admin" ? <ShieldCheck size={22} aria-hidden="true" /> : <Wrench size={22} aria-hidden="true" />}
                  </div>
                  <div className="user-details-block">
                    <h3 className="user-name">
                      {user.full_name} <span className="user-username-label">(@{user.username})</span>
                    </h3>
                    
                    <div className="user-meta-info" style={{ display: 'flex', gap: '10px', flexWrap: 'wrap', margin: '6px 0', fontSize: '0.85rem', color: '#64748b' }}>
                      {user.email && <span style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}><Mail size={14} /> {user.email}</span>}
                      {user.department_name && <span style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}><Building2 size={14} /> สาขา: {user.department_name}</span>}
                    </div>
                    
                    <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap', alignItems: 'center' }}>
                      <span className={`role-badge ${user.role === "admin" ? "role-admin" : "role-tech"}`}>
                        {user.role === "admin" ? "แอดมิน (Admin)" : "ช่างเทคนิค (Technician)"}
                      </span>

                      {/* 🔥 ป้ายกำกับช่างส่วนกลาง */}
                      {user.role === "technician" && user.is_central && (
                        <span className="role-badge" style={{ backgroundColor: '#fef3c7', color: '#b45309', border: '1px solid #fde68a' }}>
                          <Globe size={13} style={{ marginRight: '4px' }} /> ช่างส่วนกลาง
                        </span>
                      )}
                    </div>

                    {user.role === "technician" && user.specialty_names && user.specialty_names.length > 0 && (
                      <div className="specialty-container" style={{ marginTop: '8px' }}>
                        {user.specialty_names.map((name, index) => (
                          <span className="specialty-tag" key={index}><Zap size={13} aria-hidden="true" /> {name}</span>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
                
                <div className="user-actions">
                  <button className="btn-edit" onClick={() => openEditModal(user)}>
                    <Pencil size={16} aria-hidden="true" /> แก้ไข
                  </button>
                  <button className="btn-delete" onClick={() => handleDelete(user.id, user.full_name)}>
                    <Trash2 size={16} aria-hidden="true" /> ลบ
                  </button>
                </div>
              </div>
            ))
          ) : (
            <div className="no-results">
              <p><CircleX size={18} aria-hidden="true" /> ไม่พบข้อมูลบัญชีผู้ใช้งาน</p>
            </div>
          )}
        </div>
        <Pagination currentPage={currentPage} totalItems={filteredUsers.length} onPageChange={setCurrentPage} />
      </div>

      {/* ==========================================
          POPUP MODAL: สร้างผู้ใช้งานใหม่
          ========================================== */}
      {isCreateModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <button className="close-btn" type="button" onClick={resetForm} aria-label="ปิดหน้าต่าง"><X aria-hidden="true" /></button>
            <h3 className="modal-title"><Plus size={20} aria-hidden="true" /> สร้างบัญชีผู้ใช้ใหม่</h3>
            
            <form onSubmit={handleCreateSubmit}>
              <div className="input-group">
                <label>สิทธิ์การใช้งาน <span className="required">*</span></label>
                <select value={role} onChange={(e) => setRole(e.target.value)} required>
                  <option value="technician">ช่างเทคนิค (Technician)</option>
                  <option value="admin">ผู้ดูแลระบบ (Admin)</option>
                </select>
              </div>

              <div className="input-group">
                <label>ชื่อ-นามสกุลจริง <span className="required">*</span></label>
                <input 
                  type="text" 
                  placeholder="ระบุชื่อและนามสกุล (เช่น นายสมชาย ยอดช่าง)"
                  value={fullName} 
                  onChange={(e) => setFullName(e.target.value)} 
                  required 
                />
              </div>

              {/* 🔥 เพิ่มฟิลด์อีเมล */}
              <div className="input-group">
                <label>อีเมลติดต่อ <span className="required">*</span></label>
                <input 
                  type="email" 
                  placeholder="เช่น @silpakorn.edu หรือ @gmail.com"
                  value={email} 
                  onChange={(e) => setEmail(e.target.value)} 
                  required 
                />
              </div>

              {/* 🔥 เพิ่มฟิลด์สาขา */}
              <div className="input-group">
                <label>สาขาวิชา / สังกัด <span className="required">*</span></label>
                <select value={departmentId} onChange={(e) => setDepartmentId(e.target.value)} required>
                  <option value="">-- โปรดเลือกสาขาวิชา --</option>
                  {departments.map(dept => (
                    <option key={dept.id} value={dept.id}>{dept.name}</option>
                  ))}
                </select>
              </div>

              {/* 🔥 เพิ่ม Checkbox ช่างส่วนกลาง (เฉพาะช่าง) */}
              {role === "technician" && (
                <>
                  <div className="input-group" style={{ backgroundColor: '#f8fafc', padding: '10px 15px', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
                    <label style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer', margin: 0 }}>
                      <input 
                        type="checkbox" 
                        checked={isCentral} 
                        onChange={(e) => setIsCentral(e.target.checked)} 
                        style={{ width: '18px', height: '18px', cursor: 'pointer' }}
                      />
                      <strong>ตั้งเป็น "ช่างส่วนกลาง" (รับผิดชอบงานได้ทุกอาคาร/สาขา)</strong>
                    </label>
                  </div>

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
                </>
              )}

              <div className="input-group">
                <label>ชื่อผู้ใช้งาน (Username สำหรับใช้ Login) <span className="required">*</span></label>
                <input 
                  type="text" 
                  placeholder="ตั้งชื่อล็อกอินภาษาอังกฤษ"
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
                  {isSubmitting ? "กำลังบันทึก..." : <><Save size={18} aria-hidden="true" /> ยืนยันสร้างบัญชี</>}
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
            <button className="close-btn" type="button" onClick={resetForm} aria-label="ปิดหน้าต่าง"><X aria-hidden="true" /></button>
            <h3 className="modal-title"><Pencil size={20} aria-hidden="true" /> แก้ไขข้อมูลบัญชี</h3>
            
            <form onSubmit={handleEditSubmit}>
              
              <div className="input-group">
                <label>สิทธิ์การใช้งาน <span className="required">*</span></label>
                <select value={role} onChange={(e) => setRole(e.target.value)} required>
                  <option value="technician">ช่างเทคนิค (Technician)</option>
                  <option value="admin">ผู้ดูแลระบบ (Admin)</option>
                </select>
              </div>

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
                <label>อีเมลติดต่อ <span className="required">*</span></label>
                <input 
                  type="email" 
                  value={email} 
                  onChange={(e) => setEmail(e.target.value)} 
                  required 
                />
              </div>

              <div className="input-group">
                <label>สาขาวิชา / สังกัด <span className="required">*</span></label>
                <select value={departmentId} onChange={(e) => setDepartmentId(e.target.value)} required>
                  <option value="">-- โปรดเลือกสาขาวิชา --</option>
                  {departments.map(dept => (
                    <option key={dept.id} value={dept.id}>{dept.name}</option>
                  ))}
                </select>
              </div>

              {role === "technician" && (
                <>
                  <div className="input-group" style={{ backgroundColor: '#f8fafc', padding: '10px 15px', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
                    <label style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer', margin: 0 }}>
                      <input 
                        type="checkbox" 
                        checked={isCentral} 
                        onChange={(e) => setIsCentral(e.target.checked)} 
                        style={{ width: '18px', height: '18px', cursor: 'pointer' }}
                      />
                      <strong>ตั้งเป็น "ช่างส่วนกลาง" (รับผิดชอบงานได้ทุกอาคาร/สาขา)</strong>
                    </label>
                  </div>

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
                </>
              )}

              <div className="input-group">
                <label>ชื่อผู้ใช้งาน (Username) <span className="required">*</span></label>
                <input 
                  type="text" 
                  value={username} 
                  onChange={(e) => setUsername(e.target.value)} 
                  required 
                />
              </div>

              <div className="input-group">
                <label>รีเซ็ตรหัสผ่านใหม่เพื่อยืนยัน <span className="required">*</span></label>
                <input 
                  type="text" 
                  placeholder="ระบุรหัสผ่านใหม่ (หรือพิมพ์รหัสเดิม)"
                  value={password} 
                  onChange={(e) => setPassword(e.target.value)} 
                  required 
                />
                <small className="password-warning-text">
                  * เนื่องจากความปลอดภัยของระบบฐานข้อมูลหลัก กรุณาระบุรหัสผ่านใหม่หรือรหัสผ่านเดิมเพื่อยืนยันการทำรายการอัปเดตทุกครั้ง
                </small>
              </div>

              <div className="modal-actions">
                <button type="button" className="btn-cancel" onClick={resetForm} disabled={isSubmitting}>ยกเลิก</button>
                <button type="submit" className="btn-submit" disabled={isSubmitting}>
                  {isSubmitting ? "กำลังบันทึก..." : <><Save size={18} aria-hidden="true" /> บันทึกการเปลี่ยนแปลง</>}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

    </div>
  );
}

export default AdminUsers;