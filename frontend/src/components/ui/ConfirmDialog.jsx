import { useEffect, useRef } from "react";
import { AlertTriangle, X } from "lucide-react";
import "./feedback.css";

export default function ConfirmDialog({ dialog, onClose }) {
  const cancelRef = useRef(null);

  useEffect(() => {
    cancelRef.current?.focus();
    const handleKeyDown = (event) => {
      if (event.key === "Escape") onClose(false);
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  if (!dialog) return null;

  const isDanger = dialog.variant === "danger";
  return (
    <div className="ui-dialog-backdrop" role="presentation" onMouseDown={() => onClose(false)}>
      <section
        className="ui-confirm-dialog"
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="confirm-dialog-title"
        aria-describedby="confirm-dialog-description"
        onMouseDown={(event) => event.stopPropagation()}
      >
        <button className="ui-icon-button ui-confirm-dialog__close" type="button" onClick={() => onClose(false)} aria-label="ปิดหน้าต่างยืนยัน">
          <X size={20} aria-hidden="true" />
        </button>
        <div className={`ui-confirm-dialog__icon ${isDanger ? "is-danger" : ""}`}>
          <AlertTriangle size={24} aria-hidden="true" />
        </div>
        <h2 id="confirm-dialog-title">{dialog.title}</h2>
        {dialog.description && <p id="confirm-dialog-description">{dialog.description}</p>}
        <div className="ui-confirm-dialog__actions">
          <button className={`ui-button ${isDanger ? "ui-button--danger" : "ui-button--primary"}`} type="button" onClick={() => onClose(true)}>
            {dialog.confirmLabel ?? "ยืนยัน"}
          </button>
          <button ref={cancelRef} className="ui-button ui-button--secondary" type="button" onClick={() => onClose(false)}>
            {dialog.cancelLabel ?? "ยกเลิก"}
          </button>
        </div>
      </section>
    </div>
  );
}
