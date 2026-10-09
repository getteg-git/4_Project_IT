import { MessageSquareText } from "lucide-react";
import "./RepairCardNotes.css";

function RepairCardNotes({ repair }) {
  const notes = [
    { label: "หมายเหตุจากช่าง", value: repair.technician_note },
    { label: "หมายเหตุจากแอดมิน", value: repair.admin_note },
  ].filter(({ value }) => typeof value === "string" && value.trim());

  if (notes.length === 0) return null;

  return (
    <section className="repair-card-notes" aria-label="หมายเหตุการดำเนินงาน">
      {notes.map(({ label, value }) => (
        <p className="repair-card-note" key={label}>
          <strong>
            <MessageSquareText size={15} aria-hidden="true" />
            {label}:
          </strong>
          <span>{value.trim()}</span>
        </p>
      ))}
    </section>
  );
}

export default RepairCardNotes;
