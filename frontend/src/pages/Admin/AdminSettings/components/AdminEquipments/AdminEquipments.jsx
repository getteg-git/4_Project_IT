import React, { useState, useEffect, useRef } from "react";
import { CircleX, Pencil, Plus, Save, Trash2, X, FileUp } from "lucide-react";
import SearchField from "../../../../../components/ui/SearchField";
import Pagination from "../../../../../components/ui/Pagination";
import { getPageItems } from "../../../../../components/ui/paginationUtils";
import useToast from "../../../../../hooks/useToast";
import "./AdminEquipments.css";

function AdminEquipments() {
  const { toast, confirm } = useToast();
  const [equipments, setEquipments] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [currentPage, setCurrentPage] = useState(1);

  // States สำหรับ Modal
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);

  // States สำหรับฟอร์ม
  const [editId, setEditId] = useState(null);
  const [name, setName] = useState("");
  const [category, setCategory] = useState("");
  const [basePrice, setBasePrice] = useState(""); 
  const [isSubmitting, setIsSubmitting] = useState(false);

  // State & Ref สำหรับ Import Excel
  const fileInputRef = useRef(null);
  const [isImporting, setIsImporting] = useState(false); 

  // ดึงข้อมูลอุปกรณ์จาก Backend
  const fetchEquipments = async () => {
    try {
      setIsLoading(true);
      const response = await fetch("http://localhost:8080/api/equipments");
      if (response.ok) {
        const data = await response.json();
        setEquipments(data || []);
      }
    } catch (error) {
      console.error("Fetch Equipments Error:", error);
      toast.error("ดึงข้อมูลไม่สำเร็จ", { description: "ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้" });
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchEquipments();
  }, []);

  // กรองข้อมูลจากการค้นหา
  const filteredEquipments = equipments.filter((eq) =>
    eq.name.toLowerCase().includes(searchTerm.toLowerCase()) || 
    (eq.category && eq.category.toLowerCase().includes(searchTerm.toLowerCase())) ||
    (eq.asset_code && eq.asset_code.toLowerCase().includes(searchTerm.toLowerCase()))
  );
  const paginatedEquipments = getPageItems(filteredEquipments, currentPage);

  const resetForm = () => {
    setEditId(null);
    setName("");
    setCategory("");
    setBasePrice("");
    setIsCreateModalOpen(false);
    setIsEditModalOpen(false);
  };

  const openCreateModal = () => {
    resetForm();
    setIsCreateModalOpen(true);
  };

  const openEditModal = (equipment) => {
    resetForm();
    setEditId(equipment.id);
    setName(equipment.name);
    setCategory(equipment.category || "");
    setBasePrice(equipment.base_price || "");
    setIsEditModalOpen(true);
  };

  const handleCreateSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      const response = await fetch("http://localhost:8080/api/equipments", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ 
          name,
          category,
          base_price: parseFloat(basePrice) || 0 
        })
      });

      if (response.ok) {
        toast.success("เพิ่มอุปกรณ์สำเร็จ", { description: `เพิ่ม "${name}" เข้าสู่ระบบแล้ว` });
        fetchEquipments();
        resetForm();
      } else {
        const data = await response.json();
        toast.error("เพิ่มอุปกรณ์ไม่สำเร็จ", { description: data.error || "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch {
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleEditSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      const response = await fetch(`http://localhost:8080/api/equipments/${editId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ 
          name,
          category,
          base_price: parseFloat(basePrice) || 0
        })
      });

      if (response.ok) {
        toast.success("แก้ไขอุปกรณ์สำเร็จ", { description: "อัปเดตข้อมูลอุปกรณ์แล้ว" });
        fetchEquipments();
        resetForm();
      } else {
        const data = await response.json();
        toast.error("แก้ไขอุปกรณ์ไม่สำเร็จ", { description: data.error || "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch {
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async (id, eqName) => {
    const confirmDelete = await confirm({
      title: "ลบอุปกรณ์",
      description: `คุณแน่ใจหรือไม่ว่าต้องการลบ “${eqName}”? (หากอุปกรณ์นี้ถูกใช้งานอยู่จะไม่สามารถลบได้)`,
      confirmLabel: "ลบข้อมูล",
      variant: "danger",
    });
    if (!confirmDelete) return;

    try {
      const response = await fetch(`http://localhost:8080/api/equipments/${id}`, {
        method: "DELETE"
      });

      if (response.ok) {
        toast.success("ลบอุปกรณ์สำเร็จ", { description: "นำข้อมูลออกจากระบบแล้ว" });
        fetchEquipments();
      } else {
        const data = await response.json();
        toast.error("ลบข้อมูลไม่สำเร็จ", { description: data.error || "อาจมีการใช้งานอุปกรณ์นี้ค้างอยู่ในระบบ" });
      }
    } catch {
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    }
  };

  const triggerFileInput = () => {
    if (fileInputRef.current) {
      fileInputRef.current.click();
    }
  };

  const handleFileUpload = async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    e.target.value = null;
    setIsImporting(true);

    const formData = new FormData();
    formData.append("file", file);

    try {
      const response = await fetch("http://localhost:8080/api/equipments/import", {
        method: "POST",
        body: formData,
      });

      const data = await response.json();

      if (response.ok) {
        toast.success("นำเข้าไฟล์เสร็จสิ้น", { description: `เพิ่ม/อัปเดตสำเร็จ ${data.success_count} รายการ` });
        
        if (data.errors && data.errors.length > 0) {
           toast.error("พบข้อผิดพลาดบางรายการ", { 
             description: `มีข้อมูลตกหล่น ${data.errors.length} แถว (เช่น ไม่มีรหัสครุภัณฑ์)` 
           });
           console.log("รายละเอียดข้อผิดพลาดจากการ Import:", data.errors);
        }

        fetchEquipments(); 
      } else {
        toast.error("นำเข้าไฟล์ไม่สำเร็จ", { description: data.error || "รูปแบบไฟล์อาจไม่ถูกต้อง" });
      }
    } catch (error) {
      console.error("Import Error:", error);
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    } finally {
      setIsImporting(false);
    }
  };

  return (
    <div className="admin-equipments-container">
      
      {/* Toolbar */}
      <div className="equipments-toolbar">
        <div className="search-section">
          <SearchField
            id="admin-equipment-search"
            label="ค้นหาอุปกรณ์"
            placeholder="ค้นหารหัส หรือ ชื่ออุปกรณ์..."
            value={searchTerm}
            onChange={(value) => { setSearchTerm(value); setCurrentPage(1); }}
          />
        </div>
        
        <div className="toolbar-actions">
          <input 
            type="file" 
            accept=".xlsx" 
            ref={fileInputRef} 
            onChange={handleFileUpload} 
            style={{ display: "none" }} 
          />
          <button 
            className="btn-import-equipment" 
            onClick={triggerFileInput} 
            disabled={isImporting}
          >
            <FileUp size={18} aria-hidden="true" />
            {isImporting ? "กำลังนำเข้า..." : "นำเข้า Excel"}
          </button>

          <button className="btn-create-equipment" onClick={openCreateModal}>
            <Plus size={18} aria-hidden="true" /> เพิ่มอุปกรณ์ใหม่
          </button>
        </div>
      </div>

      {isImporting && (
         <div className="import-loading-state">
           กำลังประมวลผลไฟล์ Excel (อาจใช้เวลาสักครู่)...
         </div>
      )}

      {/* แบบตาราง (Data Table) */}
      <div className="equipments-list-wrapper">
        {isLoading && !isImporting ? (
          <div className="loading-state">กำลังโหลดข้อมูล...</div>
        ) : filteredEquipments.length > 0 ? (
          <div className="table-responsive">
            <table className="equipments-table">
              <thead>
                <tr>
                  <th width="15%">รหัสครุภัณฑ์</th>
                  <th width="30%">ชื่ออุปกรณ์</th>
                  <th width="15%">หมวดหมู่</th>
                  <th width="15%">สถานะ</th>
                  <th width="15%" className="text-right">มูลค่า (฿)</th>
                  <th width="10%" className="text-center">จัดการ</th>
                </tr>
              </thead>
              <tbody>
                {paginatedEquipments.map((eq) => (
                  <tr key={eq.id}>
                    <td className="font-mono text-gray">
                      {eq.asset_code || "-"}
                    </td>
                    <td className="font-medium">
                      {eq.name}
                    </td>
                    <td>
                      {eq.category ? (
                        <span className="badge badge-category">{eq.category}</span>
                      ) : "-"}
                    </td>
                    <td>
                      {eq.status ? (
                         <span className={`badge badge-status ${eq.status === 'พร้อมใช้งาน' ? 'status-active' : 'status-inactive'}`}>
                           {eq.status}
                         </span>
                      ) : "-"}
                    </td>
                    <td className="text-right font-numeric">
                      {eq.base_price > 0 ? Number(eq.base_price).toLocaleString() : "-"}
                    </td>
                    <td className="text-center">
                      <div className="table-actions">
                        <button className="btn-icon btn-edit-icon" onClick={() => openEditModal(eq)} title="แก้ไข">
                          <Pencil size={16} />
                        </button>
                        <button className="btn-icon btn-delete-icon" onClick={() => handleDelete(eq.id, eq.name)} title="ลบ">
                          <Trash2 size={16} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="no-results">
            <p><CircleX size={18} aria-hidden="true" /> ไม่พบข้อมูลอุปกรณ์</p>
          </div>
        )}
      </div>

      {filteredEquipments.length > 0 && (
        <Pagination currentPage={currentPage} totalItems={filteredEquipments.length} onPageChange={setCurrentPage} />
      )}

      {/* POPUP: สร้าง/แก้ไข ... (โค้ดเดิม) */}
      {isCreateModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <button className="close-btn" type="button" onClick={resetForm} aria-label="ปิดหน้าต่าง"><X aria-hidden="true" /></button>
            <h3 className="modal-title">เพิ่มรายการอุปกรณ์ใหม่</h3>
            <form onSubmit={handleCreateSubmit}>
              <div className="input-group">
                <label>ชื่ออุปกรณ์ <span className="required">*</span></label>
                <input type="text" value={name} onChange={(e) => setName(e.target.value)} required />
              </div>
              <div className="input-group">
                <label>หมวดหมู่ (Category)</label>
                <input type="text" value={category} onChange={(e) => setCategory(e.target.value)} />
              </div>
              <div className="input-group">
                <label>มูลค่าอ้างอิง / ราคาซื้อใหม่ (บาท)</label>
                <input type="number" min="0" step="0.01" value={basePrice} onChange={(e) => setBasePrice(e.target.value)} />
              </div>
              <div className="modal-actions">
                <button type="button" className="btn-cancel" onClick={resetForm} disabled={isSubmitting}>ยกเลิก</button>
                <button type="submit" className="btn-submit" disabled={isSubmitting}>
                  {isSubmitting ? "กำลังบันทึก..." : <><Save size={18} aria-hidden="true" /> ยืนยันเพิ่มอุปกรณ์</>}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {isEditModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <button className="close-btn" type="button" onClick={resetForm} aria-label="ปิดหน้าต่าง"><X aria-hidden="true" /></button>
            <h3 className="modal-title">แก้ไขรายการอุปกรณ์</h3>
            <form onSubmit={handleEditSubmit}>
              <div className="input-group">
                <label>ชื่ออุปกรณ์ <span className="required">*</span></label>
                <input type="text" value={name} onChange={(e) => setName(e.target.value)} required />
              </div>
              <div className="input-group">
                <label>หมวดหมู่ (Category)</label>
                <input type="text" value={category} onChange={(e) => setCategory(e.target.value)} />
              </div>
              <div className="input-group">
                <label>มูลค่าอ้างอิง / ราคาซื้อใหม่ (บาท)</label>
                <input type="number" min="0" step="0.01" value={basePrice} onChange={(e) => setBasePrice(e.target.value)} />
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

export default AdminEquipments;