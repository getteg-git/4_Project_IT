import React, { useState, useEffect } from "react";
import {
  AlertTriangle,
  Banknote,
  CalendarDays,
  CircleCheck,
  CircleX,
  ClipboardList,
  Clock3,
  Eye,
  FileText,
  Lightbulb,
  MessageSquareText,
  RotateCcw,
  UserRound,
  UserRoundPlus,
  Wrench,
  X,
  Clock
} from "lucide-react";
import BackButton from "../../../components/ui/BackButton";
import SearchField from "../../../components/ui/SearchField";
import Pagination from "../../../components/ui/Pagination";
import { getPageItems } from "../../../components/ui/paginationUtils";
import FilterBar from "../../../components/ui/FilterBar";
import useToast from "../../../hooks/useToast";
import RepairTimelineModal from "../../User/Timeline/RepairTimelineModal";
import "./AdminManage.css";

function AdminManage() {
  const { toast, confirm } = useToast();

  const [repairs, setRepairs] = useState([]);
  const [technicians, setTechnicians] = useState([]);
  const [locations, setLocations] = useState([]);
  const [floors, setFloors] = useState([]);
  const [rooms, setRooms] = useState([]);
  const [problemTypes, setProblemTypes] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [assignmentView, setAssignmentView] = useState("unassigned");
  const [locationFilter, setLocationFilter] = useState("");
  const [floorFilter, setFloorFilter] = useState("");
  const [roomFilter, setRoomFilter] = useState("");
  const [problemFilter, setProblemFilter] = useState("");
  const [currentPage, setCurrentPage] = useState(1);

  // States สำหรับ Popup
  const [selectedRepair, setSelectedRepair] = useState(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const [isAssignOpen, setIsAssignOpen] = useState(false);
  const [isTimelineOpen, setIsTimelineOpen] = useState(false); // 🔥 State สำหรับเปิดไทม์ไลน์
  const [chosenTechId, setChosenTechId] = useState("");

  // States สำหรับหน้าต่างกรอกเหตุผลไม่อนุมัติ (Reject Modal)
  const [isRejectModalOpen, setIsRejectModalOpen] = useState(false);
  const [rejectReason, setRejectReason] = useState("");
  const [rejectTargetId, setRejectTargetId] = useState(null);

  const fetchData = async () => {
    try {
      setIsLoading(true);
      const [repairsRes, usersRes, locationsRes, problemTypesRes] = await Promise.all([
        fetch("http://localhost:8080/api/repairs"),
        fetch("http://localhost:8080/api/users"),
        fetch("http://localhost:8080/api/locations"),
        fetch("http://localhost:8080/api/problem-types"),
      ]);

      if (repairsRes.ok) {
        const repairsData = await repairsRes.json();
        setRepairs(repairsData || []);
      }

      if (usersRes.ok) {
        const usersData = await usersRes.json();
        const techList = usersData.filter(user => user.role === "technician");
        setTechnicians(techList);
      }
      if (locationsRes.ok) setLocations(await locationsRes.json() || []);
      if (problemTypesRes.ok) setProblemTypes(await problemTypesRes.json() || []);
    } catch (error) {
      console.error("ดึงข้อมูลไม่สำเร็จ:", error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  useEffect(() => {
    const selectedLocation = locations.find((location) => location.name === locationFilter);
    if (!selectedLocation) {
      setFloors([]);
      return;
    }
    const fetchFloors = async () => {
      try {
        const response = await fetch(`http://localhost:8080/api/locations/${selectedLocation.id}/floors`);
        if (response.ok) setFloors(await response.json() || []);
      } catch (error) {
        console.error("ดึงข้อมูลชั้นสำหรับฟิลเตอร์ไม่สำเร็จ:", error);
      }
    };
    fetchFloors();
  }, [locationFilter, locations]);

  useEffect(() => {
    const selectedFloor = floors.find((floor) => floor.floor_name === floorFilter);
    if (!selectedFloor) {
      setRooms([]);
      return;
    }
    const fetchRooms = async () => {
      try {
        const response = await fetch(`http://localhost:8080/api/floors/${selectedFloor.id}/rooms`);
        if (response.ok) setRooms(await response.json() || []);
      } catch (error) {
        console.error("ดึงข้อมูลห้องสำหรับฟิลเตอร์ไม่สำเร็จ:", error);
      }
    };
    fetchRooms();
  }, [floorFilter, floors]);

  const normalizedSearch = searchTerm.replace(/\s+/g, '').replace(/วิทย์/g, 'วิทยาศาสตร์').toLowerCase();
  const uniqueOptions = (values) => [...new Set(values.filter(Boolean))].sort((a, b) => String(a).localeCompare(String(b), "th"));
  const locationOptions = uniqueOptions([...locations.map((location) => location.name), ...repairs.map((repair) => repair.other_location || repair.location_name)]);
  const floorOptions = uniqueOptions(floors.map((floor) => floor.floor_name));
  const roomOptions = uniqueOptions(rooms.map((room) => room.room_number));
  const problemOptions = uniqueOptions([...problemTypes.map((problemType) => problemType.name), ...repairs.map((repair) => repair.other_problem_type || repair.problem_type)]);

  const filteredRepairs = repairs.filter((repair) => {
    if (locationFilter && (repair.other_location || repair.location_name) !== locationFilter) return false;
    if (floorFilter && repair.floor_name !== floorFilter) return false;
    if (roomFilter && repair.room_number !== roomFilter) return false;
    if (problemFilter && (repair.other_problem_type || repair.problem_type) !== problemFilter) return false;
    if (!normalizedSearch) return true;
    const loc = (repair.location_name || "").replace(/\s+/g, '').toLowerCase();
    const floor = (repair.floor_name || "").replace(/\s+/g, '').toLowerCase();
    const room = (repair.room_number || "").replace(/\s+/g, '').toLowerCase();
    const type = (repair.problem_type || "").replace(/\s+/g, '').toLowerCase();
    const status = (repair.status || "").replace(/\s+/g, '').toLowerCase();
    const ticketId = (repair.ticket_number || String(repair.id)).toLowerCase(); // 🔥 อัปเดตให้ค้นหาด้วย Ticket Number ได้
    const techName = (repair.technician_name || "").replace(/\s+/g, '').toLowerCase();

    return (
      loc.includes(normalizedSearch) || floor.includes(normalizedSearch) || room.includes(normalizedSearch) ||
      type.includes(normalizedSearch) || status.includes(normalizedSearch) ||
      ticketId.includes(normalizedSearch) || techName.includes(normalizedSearch)
    );
  });
  const isCompletedRepair = (repair) => repair.status === "เสร็จเรียบร้อย" || repair.status === "เสร็จสิ้น";
  const isReviewRepair = (repair) => repair.status === "รอซ่อม" && repair.admin_note && repair.admin_note.trim() !== "";
  const unassignedRepairs = filteredRepairs.filter((repair) => !isCompletedRepair(repair) && repair.status !== "กำลังซ่อม" && !isReviewRepair(repair) && !repair.technician_name);
  const assignedRepairs = filteredRepairs.filter((repair) => !isCompletedRepair(repair) && repair.status !== "กำลังซ่อม" && !isReviewRepair(repair) && Boolean(repair.technician_name));
  const reviewRepairs = filteredRepairs.filter((repair) => !isCompletedRepair(repair) && repair.status !== "กำลังซ่อม" && isReviewRepair(repair));
  const progressRepairs = filteredRepairs.filter((repair) => repair.status === "กำลังซ่อม");
  const completedRepairs = filteredRepairs.filter(isCompletedRepair);
  const displayedRepairs = assignmentView === "unassigned"
    ? unassignedRepairs
    : assignmentView === "assigned"
      ? assignedRepairs
      : assignmentView === "review"
        ? reviewRepairs
        : assignmentView === "progress"
          ? progressRepairs
          : completedRepairs;
  const paginatedRepairs = getPageItems(displayedRepairs, currentPage);

  const getStatusClass = (status) => {
    switch (status) {
      case "รอซ่อม": return "status-pending";
      case "กำลังซ่อม": return "status-progress";
      case "เสร็จเรียบร้อย": return "status-completed";
      case "เสร็จสิ้น": return "status-completed";
      case "ซ่อมไม่ได้": return "status-failed";
      default: return "status-default";
    }
  };

  const formatDate = (dateString) => {
    const options = { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' };
    return new Date(dateString).toLocaleDateString('th-TH', options);
  };

  const formatRoom = (roomNumber) => {
    if (!roomNumber) return "";
    const roomText = String(roomNumber).trim();
    return roomText.startsWith("ห้อง") ? `(${roomText})` : `(ห้อง ${roomText})`;
  };

  const openDetailsModal = (repair) => {
    setSelectedRepair(repair);
    setIsDetailsOpen(true);
  };

  const openAssignModal = (repair) => {
    setSelectedRepair(repair);
    setChosenTechId("");
    setIsAssignOpen(true);
  };

  const handleAssignSubmit = async (e) => {
    e.preventDefault();
    if (!chosenTechId) {
      toast.warning("กรุณาเลือกช่าง", { description: "เลือกช่างจากรายชื่อก่อนมอบหมายงาน" });
      return;
    }

    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${selectedRepair.id}/assign`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ technician_id: parseInt(chosenTechId), admin_id: 1 }) // กำหนด admin_id ให้ตรงกับ log
      });

      if (response.ok) {
        toast.success("มอบหมายงานสำเร็จ", { description: "ระบบแจ้งช่างผู้รับผิดชอบแล้ว" });
        setIsAssignOpen(false);
        fetchData();
      } else {
        const errorData = await response.json();
        toast.error("มอบหมายงานไม่สำเร็จ", { description: errorData.error || "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch (error) {
      console.error("Assign Error:", error);
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    }
  };

  const handleRevoke = async (id) => {
    const confirmRevoke = await confirm({
      title: "ดึงงานกลับ",
      description: `งานซ่อมนี้จะกลับสู่สถานะรอซ่อม และช่างจะไม่สามารถดำเนินการต่อได้`,
      confirmLabel: "ดึงงานกลับ",
      variant: "danger",
    });
    if (!confirmRevoke) return;

    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${id}/revoke`, {
        method: "PUT",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: new URLSearchParams({ rejection_reason: "Admin เป็นผู้ดึงงานกลับ", admin_id: "1" })
      });
      if (response.ok) {
        toast.success("ดึงงานกลับสำเร็จ", { description: "สถานะงานกลับเป็นรอซ่อมแล้ว" });
        fetchData();
      } else {
        toast.error("ดึงงานกลับไม่สำเร็จ", { description: "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch (error) {
      console.error("Revoke Error:", error);
    }
  };

  const handleApproveRepair = async (id) => {
    const isApproved = await confirm({
      title: "อนุมัติให้ดำเนินการซ่อม",
      description: "เมื่อยืนยัน ระบบจะแจ้งให้ช่างเริ่มดำเนินการกับงานนี้ทันที",
      confirmLabel: "อนุมัติ",
    });
    if (!isApproved) return;

    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${id}/approve`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ admin_id: 1 })
      });

      if (response.ok) {
        toast.success("อนุมัติงานสำเร็จ", { description: "ช่างสามารถเริ่มดำเนินการได้แล้ว" });
        fetchData();
      } else {
        const errorData = await response.json();
        toast.error("อนุมัติงานไม่สำเร็จ", { description: errorData.error || errorData.message || "ไม่สามารถอนุมัติได้" });
      }
    } catch (error) {
      console.error("Approve Error:", error);
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    }
  };

  const openRejectModal = (id) => {
    setRejectTargetId(id);
    setRejectReason("");
    setIsRejectModalOpen(true);
  };

  const submitRejectRepair = async (e) => {
    e.preventDefault();
    if (!rejectReason.trim()) {
      toast.warning("กรุณาระบุเหตุผล", { description: "โปรดอธิบายเหตุผลที่ไม่อนุมัติการซ่อม" });
      return;
    }

    try {
      const response = await fetch(`http://localhost:8080/api/repairs/${rejectTargetId}/cancel`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ admin_note: "ไม่อนุมัติการซ่อม: " + rejectReason, admin_id: 1 })
      });
      if (response.ok) {
        toast.success("ยกเลิกงานสำเร็จ", { description: "สถานะงานเปลี่ยนเป็นซ่อมไม่ได้แล้ว" });
        setIsRejectModalOpen(false);
        fetchData();
      } else {
        toast.error("ยกเลิกงานไม่สำเร็จ", { description: "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch (error) {
      console.error("Cancel Error:", error);
    }
  };

  return (
    <div className="admin-manage-container">
      <div className="admin-wrapper">
        <nav className="page-navigation" aria-label="การนำทางย้อนกลับ">
          <BackButton to="/admin/home" label="กลับหน้าแดชบอร์ด" />
        </nav>

        <div className="admin-header">
          <div>
            <h2><ClipboardList size={24} aria-hidden="true" /> จัดการและมอบหมายงานซ่อม</h2>
            <p>ค้นหาและจ่ายงานให้ช่างเทคนิค</p>
          </div>
        </div>

        <div className="search-section">
          <SearchField
            id="admin-repair-search"
            label="ค้นหางานซ่อม"
            placeholder="หมายเลข Ticket สถานที่ ห้อง หรือชื่อช่าง"
            value={searchTerm}
            onChange={(value) => { setSearchTerm(value); setCurrentPage(1); }}
          />
          <FilterBar filters={[
            { id: "location", label: "อาคาร / สถานที่", value: locationFilter, options: locationOptions, onChange: (value) => { setLocationFilter(value); setFloorFilter(""); setRoomFilter(""); setCurrentPage(1); } },
            { id: "floor", label: "ชั้น", value: floorFilter, options: floorOptions, onChange: (value) => { setFloorFilter(value); setRoomFilter(""); setCurrentPage(1); } },
            { id: "room", label: "ห้อง", value: roomFilter, options: roomOptions, onChange: (value) => { setRoomFilter(value); setCurrentPage(1); } },
            { id: "problem", label: "ประเภทงาน", value: problemFilter, options: problemOptions, onChange: (value) => { setProblemFilter(value); setCurrentPage(1); } },
          ]} />
        </div>

        <div className="assignment-view-tabs" role="tablist" aria-label="กรองตามการมอบหมายงาน">
          <button type="button" role="tab" aria-selected={assignmentView === "unassigned"} className={assignmentView === "unassigned" ? "is-active" : ""} onClick={() => { setAssignmentView("unassigned"); setCurrentPage(1); }}>
            ยังไม่มอบหมาย <span>{unassignedRepairs.length}</span>
          </button>
          <button type="button" role="tab" aria-selected={assignmentView === "assigned"} className={assignmentView === "assigned" ? "is-active" : ""} onClick={() => { setAssignmentView("assigned"); setCurrentPage(1); }}>
            มอบหมายแล้ว <span>{assignedRepairs.length}</span>
          </button>
          <button type="button" role="tab" aria-selected={assignmentView === "review"} className={assignmentView === "review" ? "is-active" : ""} onClick={() => { setAssignmentView("review"); setCurrentPage(1); }}>
            พิจารณา <span>{reviewRepairs.length}</span>
          </button>
          <button type="button" role="tab" aria-selected={assignmentView === "progress"} className={assignmentView === "progress" ? "is-active" : ""} onClick={() => { setAssignmentView("progress"); setCurrentPage(1); }}>
            กำลังซ่อม <span>{progressRepairs.length}</span>
          </button>
          <button type="button" role="tab" aria-selected={assignmentView === "completed"} className={assignmentView === "completed" ? "is-active" : ""} onClick={() => { setAssignmentView("completed"); setCurrentPage(1); }}>
            เสร็จแล้ว <span>{completedRepairs.length}</span>
          </button>
        </div>

        <div className="repair-list">
          {isLoading ? (
            <div className="loading-state">กำลังโหลดข้อมูลงานซ่อม...</div>
          ) : displayedRepairs.length > 0 ? (
            paginatedRepairs.map((repair) => (
              <div className="repair-item-card" key={repair.id}>

                <div className="card-top">
                  {/* 🔥 อัปเดตแสดงรหัส Ticket แบบใหม่ */}
                  <span className="repair-ticket-id">{repair.ticket_number || `Ticket #${repair.id}`}</span>
                  <span className={`status-badge ${getStatusClass(repair.status)}`}>{repair.status}</span>
                </div>

                <h3 className="repair-title">[{repair.problem_type}] {repair.location_name}</h3>

                <div className="repair-details-text">
                  <p className="repair-info">
                    <strong>ชั้น / พิกัด:</strong> {repair.floor_name || "-"} {formatRoom(repair.room_number)}
                  </p>
                  {repair.equipment_name && (
                    <p className="repair-info equipment-info">
                      <strong><Wrench size={16} aria-hidden="true" /> อุปกรณ์:</strong> {repair.equipment_name}
                    </p>
                  )}
                  <p className="repair-info">
                    <strong>รายละเอียด:</strong> {repair.description}
                  </p>
                </div>

                {repair.technician_name && (
                  <div className="repair-tech-box">
                    <strong><UserRound size={16} aria-hidden="true" /> ผู้รับผิดชอบ:</strong> {repair.technician_name}
                  </div>
                )}

                {repair.status === "รอซ่อม" && repair.technician_name && (
                  repair.admin_note && repair.admin_note.includes("ระบบระงับอัตโนมัติ") ? (
                    <div className="threshold-approval-box">
                      <div>
                        <p className="admin-warning-note"><AlertTriangle size={17} aria-hidden="true" /> {repair.admin_note}</p>
                        <p><Banknote size={17} aria-hidden="true" /> ราคาประเมินจากช่าง: <strong>฿{Number(repair.estimated_cost).toLocaleString()}</strong></p>
                      </div>
                      <div className="threshold-actions">
                        <button className="btn-approve" onClick={() => handleApproveRepair(repair.id)}><CircleCheck size={17} aria-hidden="true" /> อนุมัติให้ซ่อม</button>
                        <button className="btn-reject" onClick={() => openRejectModal(repair.id)}><CircleX size={17} aria-hidden="true" /> ไม่อนุมัติ</button>
                      </div>
                    </div>
                  ) : (
                    <div className="waiting-assignment" role="status">
                      <Clock3 size={18} aria-hidden="true" />
                      <span><strong>มอบหมายแล้ว</strong><small>กำลังรอช่างกดรับงาน</small></span>
                    </div>
                  )
                )}

                <div className="card-divider"></div>

                {/* 🌟 จัดระเบียบแถวปุ่มกดให้เรียงต่อกัน */}
                <div className="card-actions-row">
                  <span className="repair-date"><CalendarDays size={16} aria-hidden="true" /> {formatDate(repair.created_at)}</span>

                  <div className="action-buttons" style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
                    <button className="btn-details" onClick={() => openDetailsModal(repair)}><Eye size={16} aria-hidden="true" /> ดูรายละเอียด</button>

                    {/* 🔥 เพิ่มปุ่มไทม์ไลน์ */}
                    <button
                      className="btn-timeline"
                      onClick={() => { setSelectedRepair(repair); setIsTimelineOpen(true); }}
                      style={{ backgroundColor: '#f1f5f9', color: '#334155', border: 'none', padding: '6px 12px', borderRadius: '6px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '4px', fontSize: '0.85rem' }}
                    >
                      <Clock size={16} aria-hidden="true" /> ไทม์ไลน์
                    </button>

                    {repair.status === "รอซ่อม" && !repair.technician_name && (
                      <button className="btn-assign" onClick={() => openAssignModal(repair)}><UserRoundPlus size={16} aria-hidden="true" /> มอบหมายช่าง</button>
                    )}

                    {repair.status === "กำลังซ่อม" && (
                      <button className="btn-revoke" onClick={() => handleRevoke(repair.id)}><RotateCcw size={16} aria-hidden="true" /> ดึงงานกลับ</button>
                    )}
                  </div>
                </div>

              </div>
            ))
          ) : (
            <div className="no-results">
              <p>
                {assignmentView === "unassigned"
                  ? "ไม่มีงานที่รอมอบหมาย"
                  : assignmentView === "assigned"
                    ? "ยังไม่มีงานที่มอบหมายแล้ว"
                    : assignmentView === "review"
                      ? "ไม่มีงานที่รอพิจารณา"
                      : assignmentView === "progress"
                        ? "ยังไม่มีงานที่กำลังซ่อม"
                        : "ยังไม่มีงานที่เสร็จแล้ว"}
              </p>
            </div>
          )}
        </div>
        <Pagination currentPage={currentPage} totalItems={displayedRepairs.length} onPageChange={setCurrentPage} />
      </div>

      {/* POPUP: ดูรายละเอียด (Details Modal) */}
      {isDetailsOpen && selectedRepair && (
        <div className="modal-overlay">
          <div className="modal-box details-box">
            <button className="close-btn" type="button" onClick={() => setIsDetailsOpen(false)} aria-label="ปิดรายละเอียด"><X aria-hidden="true" /></button>
            <div className="modal-header-center" style={{ textAlign: 'center', marginBottom: '20px' }}>
              <h3 className="assign-modal-title"><FileText size={22} aria-hidden="true" /> ข้อมูลการแจ้งซ่อมอย่างละเอียด</h3>
              <p className="assign-modal-subtitle">Ticket: {selectedRepair.ticket_number || `#${selectedRepair.id}`}</p>
            </div>

            <div className="modal-details-content">
              <div className="image-gallery">
                {selectedRepair.images && selectedRepair.images.length > 0 ? (
                  selectedRepair.images.map((img, index) => (
                    <img key={index} src={`http://localhost:8080${img.url}`} alt="รูปปัญหา" className="repair-image" />
                  ))
                ) : (
                  <div className="no-image-box">ไม่มีรูปภาพประกอบ</div>
                )}
              </div>
              <div className="info-list-container">
                <p><strong>สถานะปัจจุบัน:</strong> <span className={`status-badge ${getStatusClass(selectedRepair.status)}`}>{selectedRepair.status}</span></p>
                <p><strong>หมวดหมู่งาน:</strong> {selectedRepair.problem_type}</p>
                <p><strong>สถานที่:</strong> {selectedRepair.location_name}</p>
                <p><strong>ชั้น / พิกัด:</strong> {selectedRepair.floor_name || "-"} {formatRoom(selectedRepair.room_number)}</p>
                {selectedRepair.equipment_name && (<p><strong>อุปกรณ์ชำรุด:</strong> {selectedRepair.equipment_name}</p>)}
                <p><strong>อีเมลผู้แจ้ง:</strong> {selectedRepair.reporter_email}</p>
                <p><strong>วันที่แจ้งเรื่อง:</strong> {formatDate(selectedRepair.created_at)}</p>
              </div>
              <div className="issue-desc-box">
                <strong><FileText size={17} aria-hidden="true" /> รายละเอียดปัญหา:</strong>
                <p>{selectedRepair.description}</p>
              </div>
              {(selectedRepair.estimated_cost > 0 || selectedRepair.actual_cost > 0 || selectedRepair.admin_note) && (
                <div className="admin-cost-info-box">
                  <h4><Banknote size={18} aria-hidden="true" /> ข้อมูลค่าใช้จ่ายและการประเมิน</h4>
                  {selectedRepair.estimated_cost > 0 && (<p><strong>ราคาประเมิน:</strong> ฿{Number(selectedRepair.estimated_cost).toLocaleString()}</p>)}
                  {selectedRepair.actual_cost > 0 && (<p><strong>ค่าใช้จ่ายเบิกจริง:</strong> ฿{Number(selectedRepair.actual_cost).toLocaleString()}</p>)}
                  {selectedRepair.admin_note && (<p className="admin-warning-note"><strong><AlertTriangle size={17} aria-hidden="true" /> หมายเหตุระบบ:</strong> {selectedRepair.admin_note}</p>)}
                </div>
              )}
              {selectedRepair.technician_name && (
                <div className="tech-info-box">
                  <p><strong><UserRound size={17} aria-hidden="true" /> ช่างผู้รับผิดชอบ:</strong> {selectedRepair.technician_name}</p>
                  {selectedRepair.technician_note && (<p className="tech-note"><strong><MessageSquareText size={17} aria-hidden="true" /> หมายเหตุจากช่าง:</strong> {selectedRepair.technician_note}</p>)}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* 🔥 POPUP: มอบหมายงาน (Assign Modal) พร้อมระบบกรองความถนัดแบบ Ultimate */}
      {isAssignOpen && selectedRepair && (
        <div className="modal-overlay">
          <div className="modal-box">
            <button className="close-btn" type="button" onClick={() => setIsAssignOpen(false)} aria-label="ปิดหน้าต่าง"><X aria-hidden="true" /></button>
            <h3 className="assign-modal-title"><UserRoundPlus size={22} aria-hidden="true" /> มอบหมายงานให้ช่าง</h3>
            <p className="assign-modal-subtitle">Ticket: {selectedRepair.ticket_number || `#${selectedRepair.id}`} | {selectedRepair.problem_type}</p>

            <form onSubmit={handleAssignSubmit}>
              <div className="input-group">
                <label>เลือกช่างเทคนิคที่รับผิดชอบ <span className="required">*</span></label>
                <select className="assign-select" value={chosenTechId} onChange={(e) => setChosenTechId(e.target.value)} required>
                  <option value="">-- โปรดเลือกช่างจากรายชื่อ --</option>

                  {/* 🌟 กรองช่าง: "ความถนัดต้องตรง" และ "ต้องเป็นช่างส่วนกลาง หรือ อยู่สาขาเดียวกับจุดเกิดเหตุ" */}
                  {technicians.filter(tech => {
                    const isSkillMatch = tech.specialty_names && tech.specialty_names.includes(selectedRepair.problem_type);
                    const isAreaMatch = !selectedRepair.department_id || tech.is_central || tech.department_id === selectedRepair.department_id;
                    return isSkillMatch && isAreaMatch;
                  }).length > 0 ? (
                    technicians
                      .filter(tech => {
                        const isSkillMatch = tech.specialty_names && tech.specialty_names.includes(selectedRepair.problem_type);
                        const isAreaMatch = !selectedRepair.department_id || tech.is_central || tech.department_id === selectedRepair.department_id;
                        return isSkillMatch && isAreaMatch;
                      })
                      .map((tech) => (
                        <option key={tech.id} value={tech.id}>
                          {tech.full_name} {tech.is_central ? "[ช่างส่วนกลาง]" : `[ช่างประจำสาขา]`} [ถนัด: {tech.specialty_names.join(", ")}]
                        </option>
                      ))
                  ) : (
                    <option value="" disabled>-- ไม่มีช่างที่เหมาะสม (ไม่มีความถนัด/ไม่อยู่ในพื้นที่) --</option>
                  )}

                </select>
                <small className="assign-modal-hint"><Lightbulb size={16} aria-hidden="true" /> ระบบจะแสดงเฉพาะรายชื่อช่างที่มีความถนัดตรงและอยู่ในพื้นที่รับผิดชอบเท่านั้น</small>
              </div>
              <div className="modal-actions assign-actions">
                <button
                  type="submit"
                  className="btn-submit"
                  disabled={!chosenTechId}
                >
                  <CircleCheck size={18} aria-hidden="true" /> ยืนยันมอบหมายงาน
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* NEW POPUP: ระบุเหตุผลไม่อนุมัติ (Reject Modal) */}
      {isRejectModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box reject-modal-box">
            <button className="close-btn" type="button" onClick={() => setIsRejectModalOpen(false)} aria-label="ปิดหน้าต่าง"><X aria-hidden="true" /></button>
            <h3 className="assign-modal-title reject-modal-title"><CircleX size={23} aria-hidden="true" /> ไม่อนุมัติการซ่อม</h3>
            <p className="assign-modal-subtitle">Ticket: {repairs.find(r => r.id === rejectTargetId)?.ticket_number || `#${rejectTargetId}`}</p>

            <form onSubmit={submitRejectRepair}>
              <div className="input-group">
                <label>โปรดระบุเหตุผลที่ไม่อนุมัติ <span className="required">*</span></label>
                <textarea
                  className="reject-textarea"
                  placeholder="เช่น อุปกรณ์เสื่อมสภาพ ซื้อใหม่คุ้มกว่า..."
                  value={rejectReason}
                  onChange={(e) => setRejectReason(e.target.value)}
                  required
                />
              </div>
              <div className="modal-actions assign-actions">
                <button type="submit" className="btn-submit btn-submit-reject"><CircleX size={18} aria-hidden="true" /> ยืนยันการไม่อนุมัติ</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* 🔥 POPUP: ไทม์ไลน์ (Timeline Modal) */}
      <RepairTimelineModal
        repairId={selectedRepair?.id}
        isOpen={isTimelineOpen}
        onClose={() => { setIsTimelineOpen(false); setSelectedRepair(null); }}
      />

    </div>
  );
}

export default AdminManage;