import React, { useState, useEffect } from "react";
import { CircleX, Pencil, Plus, Save, Trash2, Wrench, Settings2, X } from "lucide-react";
import SearchField from "../../../../../components/ui/SearchField";
import Pagination from "../../../../../components/ui/Pagination";
import { getPageItems } from "../../../../../components/ui/paginationUtils";
import useToast from "../../../../../hooks/useToast";
import "./AdminProblemTypes.css";

function AdminProblemTypes() {
  const { toast, confirm } = useToast();
  const [problemTypes, setProblemTypes] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [currentPage, setCurrentPage] = useState(1);

  // States สำหรับ Modal
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);

  // States สำหรับฟอร์ม
  const [editId, setEditId] = useState(null);
  const [name, setName] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  // ดึงข้อมูลหมวดหมู่งานซ่อมจาก Backend
  const fetchProblemTypes = async () => {
    try {
      setIsLoading(true);
      const response = await fetch("http://localhost:8080/api/problem-types");
      if (response.ok) {
        const data = await response.json();
        setProblemTypes(data || []);
      }
    } catch (error) {
      console.error("Fetch Problem Types Error:", error);
      toast.error("ดึงข้อมูลไม่สำเร็จ", { description: "ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้" });
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchProblemTypes();
  }, []);

  // กรองข้อมูลจากการค้นหา
  const filteredProblems = problemTypes.filter((prob) =>
    prob.name.toLowerCase().includes(searchTerm.toLowerCase())
  );
  const paginatedProblems = getPageItems(filteredProblems, currentPage);

  // -----------------------------------------
  // ฟังก์ชัน Action
  // -----------------------------------------
  const resetForm = () => {
    setEditId(null);
    setName("");
    setIsCreateModalOpen(false);
    setIsEditModalOpen(false);
  };

  const openCreateModal = () => {
    resetForm();
    setIsCreateModalOpen(true);
  };

  const openEditModal = (problem) => {
    resetForm();
    setEditId(problem.id);
    setName(problem.name);
    setIsEditModalOpen(true);
  };

  // POST: เพิ่มหมวดหมู่ใหม่
  const handleCreateSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      const response = await fetch("http://localhost:8080/api/problem-types", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name })
      });

      if (response.ok) {
        toast.success("เพิ่มหมวดหมู่สำเร็จ", { description: `เพิ่ม "${name}" เข้าสู่ระบบแล้ว` });
        fetchProblemTypes();
        resetForm();
      } else {
        const data = await response.json();
        toast.error("เพิ่มหมวดหมู่ไม่สำเร็จ", { description: data.error || "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch {
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    } finally {
      setIsSubmitting(false);
    }
  };

  // PUT: แก้ไขหมวดหมู่
  const handleEditSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      const response = await fetch(`http://localhost:8080/api/problem-types/${editId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name })
      });

      if (response.ok) {
        toast.success("แก้ไขหมวดหมู่สำเร็จ", { description: "อัปเดตข้อมูลหมวดหมู่งานซ่อมแล้ว" });
        fetchProblemTypes();
        resetForm();
      } else {
        const data = await response.json();
        toast.error("แก้ไขหมวดหมู่ไม่สำเร็จ", { description: data.error || "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch {
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    } finally {
      setIsSubmitting(false);
    }
  };

  // DELETE: ลบหมวดหมู่
  const handleDelete = async (id, probName) => {
    const confirmDelete = await confirm({
      title: "ลบหมวดหมู่งานซ่อม",
      description: `คุณแน่ใจหรือไม่ว่าต้องการลบ “${probName}”? (หากหมวดหมู่นี้มีประวัติแจ้งซ่อมอยู่จะไม่สามารถลบได้)`,
      confirmLabel: "ลบข้อมูล",
      variant: "danger",
    });
    if (!confirmDelete) return;

    try {
      const response = await fetch(`http://localhost:8080/api/problem-types/${id}`, {
        method: "DELETE"
      });

      if (response.ok) {
        toast.success("ลบหมวดหมู่สำเร็จ", { description: "นำข้อมูลออกจากระบบแล้ว" });
        fetchProblemTypes();
      } else {
        const data = await response.json();
        toast.error("ลบข้อมูลไม่สำเร็จ", { description: data.error || "อาจมีรายการแจ้งซ่อมค้างอยู่ในหมวดหมู่นี้" });
      }
    } catch {
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    }
  };

  return (
    <div className="admin-problems-container" style={{ padding: 0 }}>
      
      {/* Toolbar: ค้นหา & เพิ่มหมวดหมู่ */}
      <div className="problems-toolbar">
        <div className="search-section">
          <SearchField
            id="admin-problem-search"
            label="ค้นหาหมวดหมู่"
            placeholder="ค้นหาชื่อหมวดหมู่งานซ่อม..."
            value={searchTerm}
            onChange={(value) => { setSearchTerm(value); setCurrentPage(1); }}
          />
        </div>
        <button className="btn-create-problem" onClick={openCreateModal}>
          <Plus size={18} aria-hidden="true" /> เพิ่มหมวดหมู่ใหม่
        </button>
      </div>

      {/* รายการหมวดหมู่ */}
      <div className="problems-list">
        {isLoading ? (
          <div className="loading-state">กำลังโหลดข้อมูล...</div>
        ) : filteredProblems.length > 0 ? (
          <div className="problems-grid">
            {paginatedProblems.map((prob) => (
              <div className="problem-item-card" key={prob.id}>
                <div className="problem-info">
                  <div className="problem-avatar">
                    <Settings2 size={24} aria-hidden="true" />
                  </div>
                  <h3 className="problem-name">{prob.name}</h3>
                </div>
                
                <div className="problem-actions">
                  <button className="btn-edit" onClick={() => openEditModal(prob)} aria-label="แก้ไข">
                    <Pencil size={16} />
                  </button>
                  <button className="btn-delete" onClick={() => handleDelete(prob.id, prob.name)} aria-label="ลบ">
                    <Trash2 size={16} />
                  </button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="no-results">
            <p><CircleX size={18} aria-hidden="true" /> ไม่พบข้อมูลหมวดหมู่งานซ่อม</p>
          </div>
        )}
      </div>

      {filteredProblems.length > 0 && (
        <Pagination currentPage={currentPage} totalItems={filteredProblems.length} onPageChange={setCurrentPage} />
      )}

      {/* ==========================================
          POPUP MODAL: สร้างหมวดหมู่ใหม่
          ========================================== */}
      {isCreateModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <button className="close-btn" type="button" onClick={resetForm} aria-label="ปิดหน้าต่าง"><X aria-hidden="true" /></button>
            <h3 className="modal-title"><Wrench size={20} aria-hidden="true" /> เพิ่มหมวดหมู่งานซ่อมใหม่</h3>
            
            <form onSubmit={handleCreateSubmit}>
              <div className="input-group">
                <label>ชื่อหมวดหมู่ <span className="required">*</span></label>
                <input 
                  type="text" 
                  placeholder="ระบุหมวดหมู่งานซ่อม เช่น งานแอร์, งานประปา"
                  value={name} 
                  onChange={(e) => setName(e.target.value)} 
                  required 
                />
              </div>

              <div className="modal-actions">
                <button type="button" className="btn-cancel" onClick={resetForm} disabled={isSubmitting}>ยกเลิก</button>
                <button type="submit" className="btn-submit" disabled={isSubmitting}>
                  {isSubmitting ? "กำลังบันทึก..." : <><Save size={18} aria-hidden="true" /> ยืนยันเพิ่มหมวดหมู่</>}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ==========================================
          POPUP MODAL: แก้ไขหมวดหมู่
          ========================================== */}
      {isEditModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <button className="close-btn" type="button" onClick={resetForm} aria-label="ปิดหน้าต่าง"><X aria-hidden="true" /></button>
            <h3 className="modal-title"><Pencil size={20} aria-hidden="true" /> แก้ไขหมวดหมู่งานซ่อม</h3>
            
            <form onSubmit={handleEditSubmit}>
              <div className="input-group">
                <label>ชื่อหมวดหมู่ <span className="required">*</span></label>
                <input 
                  type="text" 
                  value={name} 
                  onChange={(e) => setName(e.target.value)} 
                  required 
                />
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

export default AdminProblemTypes;