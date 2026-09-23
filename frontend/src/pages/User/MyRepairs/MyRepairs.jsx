import React, { useState, useEffect } from "react";
import { Camera, CalendarDays, CircleX, Eye, FileText, MessageSquareText, UserRound, Wrench, X, ClipboardCheck, Clock } from "lucide-react";
import BackButton from "../../../components/ui/BackButton";
import SearchField from "../../../components/ui/SearchField";
import Pagination from "../../../components/ui/Pagination";
import { getPageItems } from "../../../components/ui/paginationUtils";
import FilterBar from "../../../components/ui/FilterBar";
import RepairTimelineModal from "../Timeline/RepairTimelineModal"; // 🔥 แก้ไข Path ให้ถูกต้องตามโครงสร้างโฟลเดอร์
import "./MyRepairs.css";

function MyRepairs() {
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [locationFilter, setLocationFilter] = useState("");
  const [floorFilter, setFloorFilter] = useState("");
  const [roomFilter, setRoomFilter] = useState("");
  const [problemFilter, setProblemFilter] = useState("");
  const [repairs, setRepairs] = useState([]);
  const [locations, setLocations] = useState([]);
  const [floors, setFloors] = useState([]);
  const [rooms, setRooms] = useState([]);
  const [problemTypes, setProblemTypes] = useState([]);
  const [isLoading, setIsLoading] = useState(true);

  // States สำหรับ Popup
  const [selectedRepair, setSelectedRepair] = useState(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const [isTimelineOpen, setIsTimelineOpen] = useState(false);

  // ดึงข้อมูลทั้งหมดจาก Backend
  useEffect(() => {
    const fetchRepairs = async () => {
      try {
        const response = await fetch("http://localhost:8080/api/repairs");
        if (response.ok) {
          const data = await response.json();
          setRepairs(data || []);
        } else {
          console.error("ดึงข้อมูลไม่สำเร็จ");
        }
      } catch (error) {
        console.error("Error fetching repairs:", error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchRepairs();
  }, []);

  useEffect(() => {
    const fetchMasterData = async () => {
      try {
        const [locationResponse, problemResponse] = await Promise.all([
          fetch("http://localhost:8080/api/locations"),
          fetch("http://localhost:8080/api/problem-types"),
        ]);
        if (locationResponse.ok) setLocations(await locationResponse.json() || []);
        if (problemResponse.ok) setProblemTypes(await problemResponse.json() || []);
      } catch (error) {
        console.error("Error fetching filter master data:", error);
      }
    };
    fetchMasterData();
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
        console.error("Error fetching filter floors:", error);
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
        console.error("Error fetching filter rooms:", error);
      }
    };
    fetchRooms();
  }, [floorFilter, floors]);

  // ฟังก์ชัน Smart Search 
  const normalizedSearch = searchTerm.replace(/\s+/g, '').replace(/วิทย์/g, 'วิทยาศาสตร์').toLowerCase();

  const uniqueOptions = (values) => [...new Set(values.filter(Boolean))].sort((a, b) => String(a).localeCompare(String(b), "th"));
  const locationOptions = uniqueOptions([
    ...locations.map((location) => location.name),
    ...repairs.map((repair) => repair.other_location || repair.location_name),
  ]);
  const floorOptions = uniqueOptions(floors.map((floor) => floor.floor_name));
  const roomOptions = uniqueOptions(rooms.map((room) => room.room_number));
  const problemOptions = uniqueOptions([
    ...problemTypes.map((problemType) => problemType.name),
    ...repairs.map((repair) => repair.other_problem_type || repair.problem_type),
  ]);

  const matchesStatusFilter = (repair, filter) => {
    if (filter === "all") return true;
    if (filter === "completed") return repair.status === "เสร็จเรียบร้อย" || repair.status === "เสร็จสิ้น";
    if (filter === "pending") return repair.status === "รอซ่อม";
    if (filter === "progress") return repair.status === "กำลังซ่อม";
    if (filter === "failed") return repair.status === "ซ่อมไม่ได้";
    return true;
  };

  const filteredRepairs = repairs.filter((repair) => {
    if (!matchesStatusFilter(repair, statusFilter)) return false;
    if (locationFilter && (repair.other_location || repair.location_name) !== locationFilter) return false;
    if (floorFilter && repair.floor_name !== floorFilter) return false;
    if (roomFilter && repair.room_number !== roomFilter) return false;
    if (problemFilter && (repair.other_problem_type || repair.problem_type) !== problemFilter) return false;
    if (!normalizedSearch) return true;

    const loc = (repair.location_name || "").replace(/\s+/g, '').toLowerCase();
    const otherLoc = (repair.other_location || "").replace(/\s+/g, '').toLowerCase();
    const floor = (repair.floor_name || "").replace(/\s+/g, '').toLowerCase();
    const room = (repair.room_number || "").replace(/\s+/g, '').toLowerCase();
    const equip = (repair.equipment_name || "").replace(/\s+/g, '').toLowerCase();
    const type = (repair.problem_type || "").replace(/\s+/g, '').toLowerCase();
    const otherType = (repair.other_problem_type || "").replace(/\s+/g, '').toLowerCase();
    const desc = (repair.description || "").replace(/\s+/g, '').toLowerCase();
    const status = (repair.status || "").replace(/\s+/g, '').toLowerCase();
    const ticketId = (repair.ticket_number || String(repair.id)).toLowerCase();
    const techName = (repair.technician_name || "").replace(/\s+/g, '').toLowerCase();

    return (
      loc.includes(normalizedSearch) ||
      otherLoc.includes(normalizedSearch) ||
      floor.includes(normalizedSearch) ||
      room.includes(normalizedSearch) ||
      equip.includes(normalizedSearch) ||
      type.includes(normalizedSearch) ||
      otherType.includes(normalizedSearch) ||
      desc.includes(normalizedSearch) ||
      status.includes(normalizedSearch) ||
      ticketId.includes(normalizedSearch) ||
      techName.includes(normalizedSearch)
    );
  });

  const statusFilters = [
    { id: "all", label: "ทั้งหมด", count: repairs.length },
    { id: "pending", label: "รอซ่อม", count: repairs.filter((repair) => matchesStatusFilter(repair, "pending")).length },
    { id: "progress", label: "กำลังซ่อม", count: repairs.filter((repair) => matchesStatusFilter(repair, "progress")).length },
    { id: "completed", label: "เสร็จแล้ว", count: repairs.filter((repair) => matchesStatusFilter(repair, "completed")).length },
    { id: "failed", label: "ซ่อมไม่ได้", count: repairs.filter((repair) => matchesStatusFilter(repair, "failed")).length },
  ];
  const paginatedRepairs = getPageItems(filteredRepairs, currentPage);

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

  const openDetailsModal = (repair) => {
    setSelectedRepair(repair);
    setIsDetailsOpen(true);
  };

  const closeDetailsModal = () => {
    setIsDetailsOpen(false);
    setSelectedRepair(null);
  };

  return (
    <div className="tracking-container">
      <div className="tracking-wrapper">
        <section className="tracking-hero">
          <nav className="page-navigation" aria-label="การนำทางย้อนกลับ">
            <BackButton to="/" label="กลับหน้าหลัก" />
          </nav>

          <div className="tracking-header">
            <h2>ติดตามสถานะการแจ้งซ่อม</h2>
            <p>ค้นหาและตรวจสอบความคืบหน้าของรายการแจ้งซ่อมภายในคณะวิทยาศาสตร์</p>
          </div>

          <div className="search-section">
            <SearchField
              id="repair-search"
              label="ค้นหารายการ"
              placeholder="ห้อง หมวดหมู่งาน สถานะ หรือหมายเลข Ticket"
              value={searchTerm}
              onChange={(value) => { setSearchTerm(value); setCurrentPage(1); }}
            />
            <FilterBar filters={[
              { id: "location", label: "อาคาร / สถานที่", value: locationFilter, options: locationOptions, onChange: (value) => { setLocationFilter(value); setFloorFilter(""); setRoomFilter(""); setCurrentPage(1); } },
              { id: "floor", label: "ชั้น", value: floorFilter, options: floorOptions, onChange: (value) => { setFloorFilter(value); setRoomFilter(""); setCurrentPage(1); } },
              { id: "room", label: "ห้อง", value: roomFilter, options: roomOptions, onChange: (value) => { setRoomFilter(value); setCurrentPage(1); } },
              { id: "problem", label: "ประเภทงาน", value: problemFilter, options: problemOptions, onChange: (value) => { setProblemFilter(value); setCurrentPage(1); } },
            ]} />
            <div className="status-filters" role="group" aria-label="กรองตามสถานะ">
              {statusFilters.map((filter) => (
                <button
                  type="button"
                  key={filter.id}
                  className={statusFilter === filter.id ? "is-active" : ""}
                  aria-pressed={statusFilter === filter.id}
                  onClick={() => { setStatusFilter(filter.id); setCurrentPage(1); }}
                >
                  {filter.label} <span>{filter.count}</span>
                </button>
              ))}
            </div>
          </div>
        </section>

        <div className="repair-list">
          {isLoading ? (
            <div className="loading-state">กำลังโหลดข้อมูล...</div>
          ) : filteredRepairs.length > 0 ? (
            paginatedRepairs.map((repair) => {
              const displayLocation = repair.other_location || repair.location_name;
              const displayProblemType = repair.other_problem_type || repair.problem_type;
              const isCompleted = repair.status === "เสร็จเรียบร้อย" || repair.status === "เสร็จสิ้น";

              return (
                <div className="repair-item-card" key={repair.id}>
                  <div className="card-top">
                    <span className="repair-id">{repair.ticket_number || `Ticket #${repair.id}`}</span>
                    <span className={`status-badge ${getStatusClass(repair.status)}`}>
                      {repair.status}
                    </span>
                  </div>

                  <h3 className="repair-title">[{displayProblemType}] {displayLocation}</h3>

                  <p className="repair-room">
                    <strong>ชั้น / พิกัด:</strong> {repair.floor_name || "-"} {repair.room_number ? `(${repair.room_number})` : ""}
                  </p>

                  {repair.equipment_name && (
                    <p className="repair-equipment" style={{ margin: "4px 0", fontSize: "0.9rem", color: "#e67e22" }}>
                      <Wrench size={15} aria-hidden="true" /> <strong>อุปกรณ์:</strong> {repair.equipment_name} {repair.asset_code ? `(${repair.asset_code})` : ""}
                    </p>
                  )}

                  <p className="repair-desc"><strong>รายละเอียด:</strong> {repair.description}</p>

                  {repair.technician_name && (
                    <p className="repair-tech-badge"><strong><UserRound size={15} aria-hidden="true" /> ช่าง:</strong> {repair.technician_name}</p>
                  )}

                  {isCompleted && (
                    <div style={{ marginTop: "12px", marginBottom: "4px" }}>
                      <a
                        href="https://docs.google.com/forms/d/1sztsmrllXVDAAAhN9DSuE_zrYxRylyuxKWiOURnJWTw/viewform?edit_requested=true"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="btn-evaluation"
                        style={{ display: "inline-flex", alignItems: "center", gap: "6px", backgroundColor: "#8e44ad", color: "#fff", padding: "8px 14px", borderRadius: "6px", textDecoration: "none", fontSize: "0.9rem", fontWeight: "500" }}
                      >
                        <ClipboardCheck size={16} /> ทำแบบประเมินความพึงพอใจ
                      </a>
                    </div>
                  )}

                  {/* 🔥 แก้ไขการจัดวางปุ่มในการ์ดให้เรียงกันสวยงาม */}
                  <div className="card-bottom">
                    <span className="repair-date"><CalendarDays size={15} aria-hidden="true" /> {formatDate(repair.created_at)}</span>
                    <div style={{ display: 'flex', gap: '8px' }}>
                      <button
                        className="btn-details"
                        onClick={() => openDetailsModal(repair)}
                      >
                        <Eye size={17} aria-hidden="true" /> ดูรายละเอียด
                      </button>
                      <button 
                        className="btn-timeline" 
                        onClick={() => { setSelectedRepair(repair); setIsTimelineOpen(true); }} 
                        style={{ backgroundColor: '#f1f5f9', color: '#334155', border: 'none', padding: '6px 12px', borderRadius: '6px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '4px' }}
                      >
                        <Clock size={17} aria-hidden="true" /> ไทม์ไลน์
                      </button>
                    </div>
                  </div>
                </div>
              );
            })
          ) : (
            <div className="no-results">
              <p><CircleX size={18} aria-hidden="true" /> ไม่พบรายการแจ้งซ่อมที่ตรงกับคำค้นหาของคุณ</p>
            </div>
          )}
        </div>

        <Pagination currentPage={currentPage} totalItems={filteredRepairs.length} onPageChange={setCurrentPage} />
      </div>

      {/* ==========================================
          POPUP: ดูรายละเอียดงาน (Modal)
          ========================================== */}
      {isDetailsOpen && selectedRepair && (
        <div className="modal-overlay">
          <div className="modal-box details-box">
            <button className="close-btn" type="button" onClick={closeDetailsModal} aria-label="ปิดรายละเอียด"><X aria-hidden="true" /></button>
            <h3>รายละเอียดการแจ้งซ่อม</h3>
            <p className="ticket-subtitle">Ticket ID: {selectedRepair.ticket_number || `#${selectedRepair.id}`}</p>
            <hr />

            <div className="modal-details-content">
              <div className="image-gallery">
                {selectedRepair.images && selectedRepair.images.length > 0 ? (
                  selectedRepair.images.map((img, index) => (
                    <div key={index} className="image-container" style={{ position: 'relative', display: 'inline-block' }}>
                      <img
                        src={`http://localhost:8080${img.url}`}
                        alt="รูปปัญหาหน้างาน"
                        className="repair-image"
                      />
                      <span style={{ position: 'absolute', top: 5, left: 5, display: 'inline-flex', alignItems: 'center', gap: 4, background: 'rgba(0,0,0,0.6)', color: '#fff', padding: '2px 6px', borderRadius: '4px', fontSize: '0.8rem' }}>
                        <Camera size={13} aria-hidden="true" /> {img.type === 'after' ? 'หลังซ่อม' : 'ก่อนซ่อม'}
                      </span>
                    </div>
                  ))
                ) : (
                  <div className="no-image-box">ไม่มีรูปภาพประกอบ</div>
                )}
              </div>

              <div className="info-grid">
                <p><strong>สถานะปัจจุบัน:</strong> <span className={`status-badge ${getStatusClass(selectedRepair.status)}`}>{selectedRepair.status}</span></p>
                <p><strong>หมวดหมู่งาน:</strong> {selectedRepair.other_problem_type || selectedRepair.problem_type}</p>
                <p><strong>สถานที่:</strong> {selectedRepair.other_location || selectedRepair.location_name}</p>
                <p>
                  <strong>ชั้น / พิกัด:</strong> {selectedRepair.floor_name || "-"} {selectedRepair.room_number ? `(${selectedRepair.room_number})` : ""}
                </p>
                {selectedRepair.equipment_name && (
                  <p>
                    <strong>อุปกรณ์ชำรุด:</strong> {selectedRepair.equipment_name} {selectedRepair.asset_code ? `(${selectedRepair.asset_code})` : ""}
                  </p>
                )}
                <p><strong>อีเมลผู้แจ้ง:</strong> {selectedRepair.reporter_email}</p>
                <p><strong>วันที่แจ้งเรื่อง:</strong> {formatDate(selectedRepair.created_at)}</p>
              </div>

              <div className="issue-desc-box">
                <strong><FileText size={17} aria-hidden="true" /> รายละเอียดปัญหา:</strong>
                <p>{selectedRepair.description}</p>
              </div>

              {selectedRepair.status !== "รอซ่อม" && (
                <div className="tech-info-box">
                  <p><strong><UserRound size={17} aria-hidden="true" /> ช่างผู้รับผิดชอบ:</strong> {selectedRepair.technician_name || "อยู่ระหว่างดำเนินการ"}</p>
                  {selectedRepair.technician_note && (
                    <p className="tech-note"><strong><MessageSquareText size={17} aria-hidden="true" /> หมายเหตุจากช่าง:</strong> {selectedRepair.technician_note}</p>
                  )}
                  {(selectedRepair.status === "เสร็จเรียบร้อย" || selectedRepair.status === "เสร็จสิ้น") && (
                    <div style={{ marginTop: "15px", textAlign: "center" }}>
                      <a
                        href="https://docs.google.com/forms/d/1sztsmrllXVDAAAhN9DSuE_zrYxRylyuxKWiOURnJWTw/viewform?edit_requested=true"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="btn-evaluation"
                        style={{ display: "inline-flex", alignItems: "center", justifyContent: "center", gap: "6px", backgroundColor: "#8e44ad", color: "#fff", padding: "10px 16px", borderRadius: "6px", textDecoration: "none", fontSize: "0.95rem", fontWeight: "bold", width: "100%" }}
                      >
                        <ClipboardCheck size={18} /> กรุณาทำแบบประเมินความพึงพอใจ
                      </a>
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* ==========================================
          POPUP: ไทม์ไลน์ (Timeline Modal)
          ========================================== */}
      <RepairTimelineModal 
        repairId={selectedRepair?.id} 
        isOpen={isTimelineOpen} 
        onClose={() => { setIsTimelineOpen(false); setSelectedRepair(null); }} 
      />
    </div>
  );
}

export default MyRepairs;