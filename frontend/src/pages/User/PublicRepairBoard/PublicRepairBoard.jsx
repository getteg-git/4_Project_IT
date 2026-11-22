import { useEffect, useState } from "react";
import { CalendarDays, MapPin, Wrench } from "lucide-react";
import { API_BASE_URL, apiFetch as fetch } from "../../../api/baseUrl.js";
import AppBreadcrumb from "../../../components/AppBreadcrumb";
import BackButton from "../../../components/ui/BackButton";
import { getRepairStatusLabel } from "../../../utils/repairStatus.js";
import "./PublicRepairBoard.css";

function PublicRepairBoard() {
  const [repairs, setRepairs] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

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

  const formatDate = (value) => new Date(value).toLocaleDateString("th-TH", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });

  return (
    <main className="public-board-page">
      <div className="public-board-wrapper">
        <AppBreadcrumb items={[
          { label: "หน้าหลัก", to: "/" },
          { label: "รายการแจ้งซ่อมล่าสุด" },
        ]} />
        <header className="public-board-header">
          <BackButton to="/" label="กลับหน้าหลัก" />
          <h1>รายการแจ้งซ่อมล่าสุด</h1>
          <p>แสดงรายการที่แจ้งภายใน 90 วันที่ผ่านมา โดยไม่แสดงข้อมูลส่วนบุคคลหรือบันทึกภายใน</p>
        </header>

        {isLoading ? (
          <p className="public-board-state" aria-live="polite">กำลังโหลดรายการ...</p>
        ) : error ? (
          <p className="public-board-state is-error" role="alert">{error}</p>
        ) : repairs.length === 0 ? (
          <p className="public-board-state">ยังไม่มีรายการแจ้งซ่อมในช่วง 90 วันที่ผ่านมา</p>
        ) : (
          <section className="public-board-list" aria-label="รายการแจ้งซ่อม">
            {repairs.map((repair) => (
              <article className="public-repair-card" key={repair.id}>
                <div className="public-repair-card-top">
                  <strong>{repair.ticket_number || `Ticket #${repair.id}`}</strong>
                  <span>{getRepairStatusLabel(repair.status)}</span>
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
                </footer>
              </article>
            ))}
          </section>
        )}
      </div>
    </main>
  );
}

export default PublicRepairBoard;
