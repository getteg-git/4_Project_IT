import { useEffect, useState } from "react";
import {
  BarChart3,
  CalendarDays,
  ClipboardCheck,
  Clock,
  Eye,
  FileText,
  Image as ImageIcon,
  Mail,
  MapPin,
  UserRound,
  Wrench,
  X,
  ZoomIn,
} from "lucide-react";
import { API_BASE_URL, apiFetch as fetch } from "../../../api/baseUrl.js";
import AppBreadcrumb from "../../../components/AppBreadcrumb";
import RepairCardNotes from "../../../components/RepairCardNotes";
import RepairImageGallery from "../../../components/RepairImageGallery";
import BackButton from "../../../components/ui/BackButton";
import FilterBar from "../../../components/ui/FilterBar";
import Pagination from "../../../components/ui/Pagination";
import SearchField from "../../../components/ui/SearchField";
import RepairTimelineModal from "../Timeline/RepairTimelineModal";
import {
  getRepairStatusClass,
  getRepairStatusLabel,
} from "../../../utils/repairStatus.js";
import "./PublicRepairBoard.css";

function PublicRepairBoard() {
  const [repairs, setRepairs] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [reporterEmail, setReporterEmail] = useState("");
  const [trackingEmail, setTrackingEmail] = useState("");
  const [myRepairs, setMyRepairs] = useState([]);
  const [isMineView, setIsMineView] = useState(false);
  const [isTracking, setIsTracking] = useState(false);
  const [trackingError, setTrackingError] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [locationFilter, setLocationFilter] = useState("");
  const [floorFilter, setFloorFilter] = useState("");
  const [roomFilter, setRoomFilter] = useState("");
  const [problemFilter, setProblemFilter] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [isDashboardOpen, setIsDashboardOpen] = useState(false);
  const [selectedRepair, setSelectedRepair] = useState(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const [isTimelineOpen, setIsTimelineOpen] = useState(false);
  const trackingHeaders = trackingEmail
    ? { "X-Reporter-Email": trackingEmail }
    : undefined;

  const handleTrackMyRepairs = async (event) => {
    event.preventDefault();
    const email = reporterEmail.trim();
    if (!email) return;

    setIsTracking(true);
    setTrackingError("");
    setIsMineView(false);
    setMyRepairs([]);
    setTrackingEmail("");
    try {
      const query = new URLSearchParams({ email });
      const response = await fetch(`${API_BASE_URL}/api/repairs/track?${query}`);
      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.error || "ไม่พบรายการแจ้งซ่อมที่ใช้อีเมลนี้");
      }

      setMyRepairs(Array.isArray(data) ? data : []);
      setTrackingEmail(email);
      setIsMineView(true);
      setCurrentPage(1);
      resetFilters();
      setSelectedRepair(null);
      setIsDetailsOpen(false);
      setIsTimelineOpen(false);
    } catch (trackError) {
      setTrackingError(trackError.message || "ไม่สามารถค้นหารายการแจ้งซ่อมได้");
    } finally {
      setIsTracking(false);
    }
  };

  const showAllRepairs = () => {
    setIsMineView(false);
    setTrackingEmail("");
    setMyRepairs([]);
    setTrackingError("");
    setReporterEmail("");
    setCurrentPage(1);
    resetFilters();
    setSelectedRepair(null);
    setIsDetailsOpen(false);
    setIsTimelineOpen(false);
  };

  const openRepairDetails = (repair) => {
    setSelectedRepair(repair);
    setIsDetailsOpen(true);
  };

  const getOutcomeNotes = (repair) => {
    if (repair.status === "ปิดงานก่อนครบขั้นตอน" && repair.closure_note?.trim()) {
      return [{ label: "หมายเหตุปิดงานก่อนครบขั้นตอน", value: repair.closure_note.trim() }];
    }
    if (repair.status === "ซ่อมไม่ได้") {
      const reason = repair.admin_note?.replace(/^ไม่อนุมัติการซ่อม:\s*/u, "") ||
        repair.technician_note ||
        repair.admin_note;
      return reason?.trim()
        ? [{ label: "เหตุผลที่ซ่อมไม่สำเร็จ", value: reason.trim() }]
        : [];
    }
    if (repair.status?.startsWith("ส่งซ่อม")) {
      const notes = [];
      if (repair.technician_note?.trim()) {
        notes.push({ label: "เหตุผลที่ซ่อมไม่สำเร็จ", value: repair.technician_note.trim() });
      }
      const handoffDetails = repair.admin_note?.replace(/^ส่งซ่อม(?:ภายนอก|ให้คณะ)(?:\s*\/\s*คณะ)?:\s*/u, "");
      if (handoffDetails?.trim()) {
        notes.push({ label: "รายละเอียดการส่งซ่อมต่อ", value: handoffDetails.trim() });
      }
      return notes;
    }
    return [];
  };
  const isBreakEvenAdminNote = (note) =>
    typeof note === "string" && note.trim().startsWith("แจ้งเตือน: ราคาประเมิน");

  useEffect(() => {
    let active = true;
    const loadRepairs = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/api/repairs/public-board`);
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || "ไม่สามารถโหลดรายการแจ้งซ่อมได้");
        if (active) setRepairs(Array.isArray(data) ? data : []);
      } catch (loadError) {
        if (active) setError(loadError.message || "ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้");
      } finally {
        if (active) setIsLoading(false);
      }
    };
    loadRepairs();
    return () => { active = false; };
  }, []);

  useEffect(() => {
    if (!isDashboardOpen) return undefined;
    const handleKeyDown = (event) => {
      if (event.key === "Escape") setIsDashboardOpen(false);
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isDashboardOpen]);

  const formatDate = (value) => new Date(value).toLocaleDateString("th-TH", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
  const getUniqueValues = (values) => [...new Set(values.filter(Boolean))].sort((a, b) => a.localeCompare(b, "th"));
  const visibleRepairs = isMineView
    ? myRepairs.filter((repair) =>
        !["ส่งคืนแอดมิน", "ส่งซ่อมให้คณะ"].includes(repair.status)
      )
    : repairs.filter((repair) => repair.status !== "ส่งคืนแอดมิน");
  const locationOptions = getUniqueValues(visibleRepairs.map((repair) => repair.other_location || repair.location_name));
  const floorOptions = getUniqueValues(
    visibleRepairs
      .filter((repair) => !locationFilter || (repair.other_location || repair.location_name) === locationFilter)
      .map((repair) => repair.floor_name)
  );
  const roomOptions = getUniqueValues(
    visibleRepairs
      .filter((repair) =>
        (!locationFilter || (repair.other_location || repair.location_name) === locationFilter) &&
        (!floorFilter || repair.floor_name === floorFilter)
      )
      .map((repair) => repair.other_room || repair.room_number)
  );
  const problemOptions = getUniqueValues(visibleRepairs.map((repair) => repair.other_problem_type || repair.problem_type));
  const statusOptions = getUniqueValues(visibleRepairs.map((repair) => getRepairStatusLabel(repair.status)));
  const normalizedSearch = searchTerm.trim().toLocaleLowerCase();
  const filteredRepairs = visibleRepairs.filter((repair) => {
    const location = repair.other_location || repair.location_name || "";
    const room = repair.other_room || repair.room_number || "";
    const problem = repair.other_problem_type || repair.problem_type || "";
    if (locationFilter && location !== locationFilter) return false;
    if (floorFilter && repair.floor_name !== floorFilter) return false;
    if (roomFilter && room !== roomFilter) return false;
    if (problemFilter && problem !== problemFilter) return false;
    if (statusFilter && getRepairStatusLabel(repair.status) !== statusFilter) return false;
    if (!normalizedSearch) return true;
    return [
      repair.ticket_number,
      repair.description,
      location,
      repair.floor_name,
      room,
      problem,
      repair.status,
    ].some((value) => String(value || "").toLocaleLowerCase().includes(normalizedSearch));
  });
  const pageSize = 10;
  const paginatedRepairs = filteredRepairs.slice((currentPage - 1) * pageSize, currentPage * pageSize);
  const getStatusCount = (statuses) => repairs.filter((repair) =>
    statuses.includes(repair.status) || statuses.includes(getRepairStatusLabel(repair.status))
  ).length;
  const resetFilters = () => {
    setLocationFilter("");
    setFloorFilter("");
    setRoomFilter("");
    setProblemFilter("");
    setStatusFilter("");
    setSearchTerm("");
    setCurrentPage(1);
  };

  return (
    <main className="public-board-page">
      <div className="public-board-wrapper">
        <AppBreadcrumb items={[
          { label: "หน้าหลัก", to: "/" },
          { label: "รายการแจ้งซ่อมทั้งหมด" },
        ]} />
        <header className="public-board-header">
          <BackButton to="/" label="กลับหน้าหลัก" />
          <h1>{isMineView ? "รายการแจ้งซ่อมของฉัน" : "รายการแจ้งซ่อมทั้งหมด"}</h1>
          <p>ดูรายการแจ้งซ่อมล่าสุด หรือค้นหารายการของคุณด้วยอีเมลที่ใช้แจ้ง</p>
          <form className="public-tracking-form" onSubmit={handleTrackMyRepairs}>
            <label htmlFor="public-tracking-email">ค้นหารายการของฉันด้วยอีเมล</label>
            <div className="public-tracking-form-row">
              <input
                id="public-tracking-email"
                type="email"
                value={reporterEmail}
                onChange={(event) => setReporterEmail(event.target.value)}
                placeholder="name@example.com"
                maxLength={254}
                required
              />
              <button type="submit" disabled={isTracking}>
                {isTracking ? "กำลังค้นหา..." : "ค้นหางานของฉัน"}
              </button>
              {isMineView && (
                <button type="button" className="public-show-all-button" onClick={showAllRepairs}>
                  กลับไปดูรายการทั้งหมด
                </button>
              )}
            </div>
            {trackingError && <p className="public-tracking-error" role="alert">{trackingError}</p>}
          </form>
        </header>

        {isMineView && trackingEmail && (
          <p className="public-personal-view-note">
            กำลังแสดงรายการที่ผูกกับอีเมล {trackingEmail}
          </p>
        )}

        {(isMineView || (!isLoading && !error)) && visibleRepairs.length > 0 && (
          <section className="public-board-controls" aria-label="ค้นหาและกรองรายการแจ้งซ่อม">
            <SearchField
              id="public-repair-search"
              label="ค้นหารายการ"
              placeholder="หมายเลข Ticket สถานที่ ห้อง หรือรายละเอียด"
              value={searchTerm}
              onChange={(value) => {
                setSearchTerm(value);
                setCurrentPage(1);
              }}
            />
            <FilterBar
              filters={[
                {
                  id: "location",
                  label: "อาคาร / สถานที่",
                  value: locationFilter,
                  options: locationOptions,
                  onChange: (value) => {
                    setLocationFilter(value);
                    setFloorFilter("");
                    setRoomFilter("");
                    setCurrentPage(1);
                  },
                },
                {
                  id: "floor",
                  label: "ชั้น",
                  value: floorFilter,
                  options: floorOptions,
                  onChange: (value) => {
                    setFloorFilter(value);
                    setRoomFilter("");
                    setCurrentPage(1);
                  },
                },
                {
                  id: "room",
                  label: "ห้อง",
                  value: roomFilter,
                  options: roomOptions,
                  onChange: (value) => {
                    setRoomFilter(value);
                    setCurrentPage(1);
                  },
                },
                {
                  id: "problem",
                  label: "ประเภทงาน",
                  value: problemFilter,
                  options: problemOptions,
                  onChange: (value) => {
                    setProblemFilter(value);
                    setCurrentPage(1);
                  },
                },
                {
                  id: "status",
                  label: "สถานะ",
                  value: statusFilter,
                  options: statusOptions,
                  onChange: (value) => {
                    setStatusFilter(value);
                    setCurrentPage(1);
                  },
                },
              ]}
            />
            <button type="button" className="public-filter-reset" onClick={resetFilters}>
              ล้างการค้นหาและตัวกรอง
            </button>
          </section>
        )}

        {!isLoading && !error && !isMineView && repairs.length > 0 && (
          <button
            className="mini-dashboard"
            type="button"
            onClick={() => setIsDashboardOpen(true)}
            aria-haspopup="dialog"
            aria-label={`เปิด Dashboard รายการแจ้งซ่อมทั้งหมด ${repairs.length} รายการ`}
          >
            <span className="mini-dashboard-title"><BarChart3 size={17} /> ภาพรวมงานซ่อม <ZoomIn size={15} /></span>
            <span className="mini-dashboard-count">{repairs.length}<small>รายการทั้งหมด</small></span>
            <span className="mini-dashboard-summary">
              <span>กำลังซ่อม <strong>{getStatusCount(["กำลังซ่อม"])}</strong></span>
              <span>เสร็จสิ้น <strong>{getStatusCount(["เสร็จสิ้น", "เสร็จเรียบร้อย"])}</strong></span>
            </span>
          </button>
        )}

        {!isMineView && isLoading ? (
          <p className="public-board-state" aria-live="polite">กำลังโหลดรายการ...</p>
        ) : !isMineView && error ? (
          <p className="public-board-state is-error" role="alert">{error}</p>
        ) : visibleRepairs.length === 0 ? (
          <p className="public-board-state">
            {isMineView
              ? myRepairs.length > 0
                ? "ไม่มีรายการที่แสดงสำหรับสถานะงานปัจจุบัน"
                : "ไม่พบรายการแจ้งซ่อมที่ผูกกับอีเมลนี้"
              : repairs.length > 0
                ? "ไม่มีรายการที่แสดงสำหรับสถานะงานปัจจุบัน"
                : "ยังไม่มีรายการแจ้งซ่อมในช่วง 90 วันที่ผ่านมา"}
          </p>
        ) : filteredRepairs.length === 0 ? (
          <p className="public-board-state">ไม่พบรายการที่ตรงกับคำค้นหาหรือตัวกรอง</p>
        ) : isMineView ? (
          <section className="public-board-list" aria-label="รายการแจ้งซ่อมของฉัน">
            {paginatedRepairs.map((repair) => {
              const outcomeNotes = getOutcomeNotes(repair);
              const isCompleted = ["เสร็จเรียบร้อย", "เสร็จสิ้น"].includes(repair.status);
              return (
                <article className="public-repair-card public-personal-card" key={repair.id}>
                  <div className="public-repair-card-top">
                    <strong>{repair.ticket_number || `Ticket #${repair.id}`}</strong>
                    <div className="public-personal-card-badges">
                      <span className={`public-repair-status ${getRepairStatusClass(repair.status)}`}>
                        {getRepairStatusLabel(repair.status)}
                      </span>
                      <span className={`repair-image-indicator ${repair.images?.length ? "has-images" : "no-images"}`}>
                        <ImageIcon size={15} aria-hidden="true" />
                        {repair.images?.length ? `มีรูปภาพ (${repair.images.length})` : "ไม่มีรูปภาพ"}
                      </span>
                    </div>
                  </div>
                  <h2>{repair.other_problem_type || repair.problem_type || "แจ้งซ่อม"} · {repair.other_location || repair.location_name || "ไม่ระบุสถานที่"}</h2>
                  <p className="public-repair-location">
                    <MapPin size={16} aria-hidden="true" />
                    {repair.floor_name || "ไม่ระบุชั้น"}
                    {(repair.other_room || repair.room_number) ? ` · ห้อง ${repair.other_room || repair.room_number}` : ""}
                  </p>
                  {repair.equipment_name && (
                    <p className="public-repair-equipment">
                      <Wrench size={16} aria-hidden="true" /> {repair.equipment_name}
                    </p>
                  )}
                  {repair.reporter_email && (
                    <p className="public-repair-location">
                      <Mail size={16} aria-hidden="true" /> {repair.reporter_email}
                    </p>
                  )}
                  <p className="public-repair-description">{repair.description}</p>
                  <RepairCardNotes
                    repair={repair}
                    showTechnicianNote={false}
                    showAdminNote={outcomeNotes.length === 0 && !isBreakEvenAdminNote(repair.admin_note)}
                    extraNotes={outcomeNotes}
                  />
                  <footer>
                    <span><CalendarDays size={15} aria-hidden="true" /> {formatDate(repair.created_at)}</span>
                    <div className="public-personal-card-actions">
                      {isCompleted && (
                        <a
                          href="https://docs.google.com/forms/d/1sztsmrllXVDAAAhN9DSuE_zrYxRylyuxKWiOURnJWTw/viewform?edit_requested=true"
                          target="_blank"
                          rel="noopener noreferrer"
                          className="public-evaluation-link"
                        >
                          <ClipboardCheck size={16} aria-hidden="true" /> ประเมินความพึงพอใจ
                        </a>
                      )}
                      <button type="button" onClick={() => openRepairDetails(repair)}>
                        <Eye size={16} aria-hidden="true" /> รายละเอียด
                      </button>
                      <button
                        type="button"
                        onClick={() => {
                          setSelectedRepair(repair);
                          setIsTimelineOpen(true);
                        }}
                      >
                        <Clock size={16} aria-hidden="true" /> ไทม์ไลน์
                      </button>
                    </div>
                  </footer>
                </article>
              );
            })}
          </section>
        ) : (
          <section className="public-board-list" aria-label="รายการแจ้งซ่อม">
            {paginatedRepairs.map((repair) => (
              <article className="public-repair-card" key={repair.id}>
                <div className="public-repair-card-top">
                  <strong>{repair.ticket_number || `Ticket #${repair.id}`}</strong>
                  <span className={`public-repair-status ${getRepairStatusClass(repair.status)}`}>
                    {getRepairStatusLabel(repair.status)}
                  </span>
                </div>
                <h2>{repair.other_problem_type || repair.problem_type || "แจ้งซ่อม"}</h2>
                <p className="public-repair-location">
                  <MapPin size={16} aria-hidden="true" />
                  {repair.other_location || repair.location_name || "ไม่ระบุสถานที่"}
                  {repair.floor_name ? ` · ${repair.floor_name}` : ""}
                  {(repair.other_room || repair.room_number) ? ` · ห้อง ${repair.other_room || repair.room_number}` : ""}
                </p>
                {repair.equipment_name && (
                  <p className="public-repair-equipment">
                    <Wrench size={16} aria-hidden="true" /> {repair.equipment_name}
                  </p>
                )}
                <p className="public-repair-description">{repair.description}</p>
                <footer>
                  <span><CalendarDays size={15} aria-hidden="true" /> {formatDate(repair.created_at)}</span>
                  <span className={`repair-image-indicator ${repair.has_images ? "has-images" : "no-images"}`}>
                    <ImageIcon size={15} aria-hidden="true" />
                    {repair.has_images ? "มีรูปภาพ" : "ไม่มีรูปภาพ"}
                  </span>
                  <button
                    type="button"
                    className="public-details-button"
                    onClick={() => openRepairDetails(repair)}
                  >
                    <Eye size={16} aria-hidden="true" /> รายละเอียด
                  </button>
                </footer>
              </article>
            ))}
          </section>
        )}
        {(isMineView || (!isLoading && !error)) && visibleRepairs.length > 0 && filteredRepairs.length > 0 && (
          <Pagination
            currentPage={currentPage}
            totalItems={filteredRepairs.length}
            pageSize={pageSize}
            onPageChange={setCurrentPage}
          />
        )}
      </div>
      {isDashboardOpen && (
        <div className="dashboard-overlay" onClick={() => setIsDashboardOpen(false)}>
          <section
            className="dashboard-modal"
            role="dialog"
            aria-modal="true"
            aria-labelledby="public-dashboard-title"
            onClick={(event) => event.stopPropagation()}
          >
            <button
              className="dashboard-close"
              type="button"
              onClick={() => setIsDashboardOpen(false)}
              aria-label="ปิด Dashboard"
            >
              <X size={20} />
            </button>
            <p className="dashboard-eyebrow"><BarChart3 size={17} /> สรุปสถานะงาน</p>
            <h2 id="public-dashboard-title">ภาพรวมการแจ้งซ่อม</h2>
            <p className="dashboard-description">สรุปรายการแจ้งซ่อมทั้งหมดในช่วง 90 วันที่ผ่านมา</p>
            <div className="dashboard-metrics">
              <article className="dashboard-metric"><span>รายการทั้งหมด</span><strong>{repairs.length}</strong><small>รายการ</small></article>
              <article className="dashboard-metric"><span>กำลังซ่อม</span><strong>{getStatusCount(["กำลังซ่อม"])}</strong><small>รายการ</small></article>
              <article className="dashboard-metric"><span>เสร็จสิ้น</span><strong>{getStatusCount(["เสร็จสิ้น", "เสร็จเรียบร้อย"])}</strong><small>รายการ</small></article>
              <article className="dashboard-metric"><span>ปิดงานก่อนครบขั้นตอน</span><strong>{getStatusCount(["ปิดงานก่อนครบขั้นตอน"])}</strong><small>รายการ</small></article>
            </div>
            <div className="dashboard-breakdown">
              <h3>สถานะแต่ละรายการ</h3>
              {statusOptions.map((label) => {
                const count = getStatusCount([label]);
                const percent = repairs.length ? Math.round((count / repairs.length) * 100) : 0;
                return (
                  <div className="dashboard-bar-row" key={label}>
                    <span>{label}</span>
                    <div className="dashboard-bar-track" aria-label={`${label} ${percent}%`}>
                      <span className="dashboard-bar-fill" style={{ width: `${percent}%` }} />
                    </div>
                    <strong>{count}</strong>
                  </div>
                );
              })}
            </div>
          </section>
        </div>
      )}
      {isDetailsOpen && selectedRepair && (
        <div className="modal-overlay" onClick={() => setIsDetailsOpen(false)}>
          <section
            className="modal-box public-repair-details"
            role="dialog"
            aria-modal="true"
            aria-labelledby="repair-details-title"
            onClick={(event) => event.stopPropagation()}
          >
            <button
              className="close-btn"
              type="button"
              onClick={() => setIsDetailsOpen(false)}
              aria-label="ปิดรายละเอียด"
            >
              <X aria-hidden="true" />
            </button>
            <h2 id="repair-details-title">รายละเอียดการแจ้งซ่อม</h2>
            <p className="ticket-subtitle">
              Ticket: {selectedRepair.ticket_number || `#${selectedRepair.id}`}
            </p>
            {isMineView ? (
              <RepairImageGallery
                images={selectedRepair.images}
                ticketLabel={selectedRepair.ticket_number || `#${selectedRepair.id}`}
                requestHeaders={trackingHeaders}
              />
            ) : selectedRepair.has_images ? (
              <p className="public-details-image-note">
                <ImageIcon size={16} aria-hidden="true" /> รายการนี้มีรูปภาพแนบ
              </p>
            ) : null}
            <div className="public-personal-details-grid">
              <p>
                <strong>สถานะ:</strong>{" "}
                <span className={`public-repair-status ${getRepairStatusClass(selectedRepair.status)}`}>
                  {getRepairStatusLabel(selectedRepair.status)}
                </span>
              </p>
              <p><strong>หมวดหมู่งาน:</strong> {selectedRepair.other_problem_type || selectedRepair.problem_type || "-"}</p>
              <p><strong>สถานที่:</strong> {selectedRepair.other_location || selectedRepair.location_name || "-"}</p>
              <p><strong>ชั้น / ห้อง:</strong> {[selectedRepair.floor_name, selectedRepair.other_room || selectedRepair.room_number].filter(Boolean).join(" · ") || "-"}</p>
              {selectedRepair.equipment_name && <p><strong>อุปกรณ์:</strong> {selectedRepair.equipment_name}</p>}
              {isMineView && <p><strong>อีเมลผู้แจ้ง:</strong> {selectedRepair.reporter_email}</p>}
              {isMineView && selectedRepair.reporter_type && <p><strong>สถานะผู้แจ้ง:</strong> {selectedRepair.reporter_type}</p>}
              {isMineView && selectedRepair.reporter_degree_level && <p><strong>ระดับการศึกษา:</strong> {selectedRepair.reporter_degree_level}</p>}
              {isMineView && selectedRepair.reporter_department && <p><strong>ภาควิชา / หน่วยงาน:</strong> {selectedRepair.reporter_department}</p>}
              <p><strong>วันที่แจ้ง:</strong> {formatDate(selectedRepair.created_at)}</p>
            </div>
            <div className="public-personal-description">
              <strong><FileText size={17} aria-hidden="true" /> รายละเอียดปัญหา</strong>
              <p>{selectedRepair.description}</p>
            </div>
            {isMineView && (
              <RepairCardNotes
                repair={selectedRepair}
                showTechnicianNote={false}
                showAdminNote={false}
                extraNotes={getOutcomeNotes(selectedRepair)}
              />
            )}
            {isMineView && selectedRepair.technician_name && (
              <p className="public-personal-technician">
                <UserRound size={17} aria-hidden="true" />
                <strong>ช่างผู้รับผิดชอบ:</strong> {selectedRepair.technician_name}
              </p>
            )}
          </section>
        </div>
      )}
      <RepairTimelineModal
        repairId={selectedRepair?.id}
        images={selectedRepair?.images}
        ticketLabel={selectedRepair?.ticket_number || `#${selectedRepair?.id}`}
        trackingCredentials={trackingEmail ? { email: trackingEmail } : null}
        assignedTechnician={selectedRepair?.technician_name}
        view="summary"
        isOpen={isTimelineOpen}
        onClose={() => {
          setIsTimelineOpen(false);
          setSelectedRepair(null);
        }}
      />
    </main>
  );
}

export default PublicRepairBoard;
