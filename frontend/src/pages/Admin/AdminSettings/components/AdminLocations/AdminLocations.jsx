import React, { useState, useEffect } from "react";
import { CircleX, Pencil, Plus, Save, Trash2, MapPin, Building2, X } from "lucide-react";
import SearchField from "../../../../../components/ui/SearchField";
import Pagination from "../../../../../components/ui/Pagination";
import { getPageItems } from "../../../../../components/ui/paginationUtils";
import useToast from "../../../../../hooks/useToast";
import "./AdminLocations.css";

function AdminLocations() {
  const { toast, confirm } = useToast();
  const [locations, setLocations] = useState([]);
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

  // ดึงข้อมูลสถานที่จาก Backend
  const fetchLocations = async () => {
    try {
      setIsLoading(true);
      const response = await fetch("http://localhost:8080/api/locations");
      if (response.ok) {
        const data = await response.json();
        setLocations(data || []);
      }
    } catch (error) {
      console.error("Fetch Locations Error:", error);
      toast.error("ดึงข้อมูลไม่สำเร็จ", { description: "ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้" });
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchLocations();
  }, []);

  // กรองข้อมูลจากการค้นหา
  const filteredLocations = locations.filter((loc) =>
    loc.name.toLowerCase().includes(searchTerm.toLowerCase())
  );
  const paginatedLocations = getPageItems(filteredLocations, currentPage);

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

  const openEditModal = (location) => {
    resetForm();
    setEditId(location.id);
    setName(location.name);
    setIsEditModalOpen(true);
  };

  // POST: เพิ่มสถานที่ใหม่
  const handleCreateSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      const response = await fetch("http://localhost:8080/api/locations", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name })
      });

      if (response.ok) {
        toast.success("เพิ่มสถานที่สำเร็จ", { description: `เพิ่ม "${name}" เข้าสู่ระบบแล้ว` });
        fetchLocations();
        resetForm();
      } else {
        const data = await response.json();
        toast.error("เพิ่มสถานที่ไม่สำเร็จ", { description: data.error || "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch {
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    } finally {
      setIsSubmitting(false);
    }
  };

  // PUT: แก้ไขสถานที่
  const handleEditSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      const response = await fetch(`http://localhost:8080/api/locations/${editId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name })
      });

      if (response.ok) {
        toast.success("แก้ไขสถานที่สำเร็จ", { description: "อัปเดตข้อมูลอาคาร/สถานที่แล้ว" });
        fetchLocations();
        resetForm();
      } else {
        const data = await response.json();
        toast.error("แก้ไขสถานที่ไม่สำเร็จ", { description: data.error || "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch {
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    } finally {
      setIsSubmitting(false);
    }
  };

  // DELETE: ลบสถานที่
  const handleDelete = async (id, locName) => {
    const confirmDelete = await confirm({
      title: "ลบสถานที่",
      description: `คุณแน่ใจหรือไม่ว่าต้องการลบ “${locName}” (หากสถานที่นี้มีประวัติแจ้งซ่อมอยู่จะไม่สามารถลบได้)`,
      confirmLabel: "ลบข้อมูล",
      variant: "danger",
    });
    if (!confirmDelete) return;

    try {
      const response = await fetch(`http://localhost:8080/api/locations/${id}`, {
        method: "DELETE"
      });

      if (response.ok) {
        toast.success("ลบสถานที่สำเร็จ", { description: "นำข้อมูลออกจากระบบแล้ว" });
        fetchLocations();
      } else {
        const data = await response.json();
        toast.error("ลบข้อมูลไม่สำเร็จ", { description: data.error || "อาจมีรายการแจ้งซ่อมค้างอยู่ในสถานที่นี้" });
      }
    } catch {
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    }
  };

  return (
    <div className="admin-locations-container" style={{ padding: 0 }}>
      
      {/* Toolbar: ค้นหา & เพิ่มสถานที่ */}
      <div className="locations-toolbar">
        <div className="search-section">
          <SearchField
            id="admin-location-search"
            label="ค้นหาสถานที่"
            placeholder="ค้นหาชื่ออาคาร หรือสถานที่..."
            value={searchTerm}
            onChange={(value) => { setSearchTerm(value); setCurrentPage(1); }}
          />
        </div>
        <button className="btn-create-location" onClick={openCreateModal}>
          <Plus size={18} aria-hidden="true" /> เพิ่มสถานที่ใหม่
        </button>
      </div>

      {/* รายการสถานที่ */}
      <div className="locations-list">
        {isLoading ? (
          <div className="loading-state">กำลังโหลดข้อมูล...</div>
        ) : filteredLocations.length > 0 ? (
          <div className="locations-grid">
            {paginatedLocations.map((loc) => (
              <div className="location-item-card" key={loc.id}>
                <div className="location-info">
                  <div className="location-avatar">
                    <Building2 size={24} aria-hidden="true" />
                  </div>
                  <h3 className="location-name">{loc.name}</h3>
                </div>
                
                <div className="location-actions">
                  <button className="btn-edit" onClick={() => openEditModal(loc)} aria-label="แก้ไข">
                    <Pencil size={16} />
                  </button>
                  <button className="btn-delete" onClick={() => handleDelete(loc.id, loc.name)} aria-label="ลบ">
                    <Trash2 size={16} />
                  </button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="no-results">
            <p><CircleX size={18} aria-hidden="true" /> ไม่พบข้อมูลสถานที่</p>
          </div>
        )}
      </div>

      {filteredLocations.length > 0 && (
        <Pagination currentPage={currentPage} totalItems={filteredLocations.length} onPageChange={setCurrentPage} />
      )}

      {/* ==========================================
          POPUP MODAL: สร้างสถานที่ใหม่
          ========================================== */}
      {isCreateModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <button className="close-btn" type="button" onClick={resetForm} aria-label="ปิดหน้าต่าง"><X aria-hidden="true" /></button>
            <h3 className="modal-title"><MapPin size={20} aria-hidden="true" /> เพิ่มอาคาร/สถานที่ใหม่</h3>
            
            <form onSubmit={handleCreateSubmit}>
              <div className="input-group">
                <label>ชื่อสถานที่ <span className="required">*</span></label>
                <input 
                  type="text" 
                  placeholder="ระบุชื่ออาคาร เช่น อาคารเรียนรวม 1"
                  value={name} 
                  onChange={(e) => setName(e.target.value)} 
                  required 
                />
              </div>

              <div className="modal-actions">
                <button type="button" className="btn-cancel" onClick={resetForm} disabled={isSubmitting}>ยกเลิก</button>
                <button type="submit" className="btn-submit" disabled={isSubmitting}>
                  {isSubmitting ? "กำลังบันทึก..." : <><Save size={18} aria-hidden="true" /> ยืนยันเพิ่มสถานที่</>}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ==========================================
          POPUP MODAL: แก้ไขสถานที่
          ========================================== */}
      {isEditModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <button className="close-btn" type="button" onClick={resetForm} aria-label="ปิดหน้าต่าง"><X aria-hidden="true" /></button>
            <h3 className="modal-title"><Pencil size={20} aria-hidden="true" /> แก้ไขชื่อสถานที่</h3>
            
            <form onSubmit={handleEditSubmit}>
              <div className="input-group">
                <label>ชื่อสถานที่ <span className="required">*</span></label>
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

export default AdminLocations;