import { useEffect, useRef, useState } from "react";
import { CheckCircle2, Info, TriangleAlert, X, XCircle } from "lucide-react";
import "./feedback.css";

const toastIcons = {
  success: CheckCircle2,
  info: Info,
  warning: TriangleAlert,
  error: XCircle,
};

function Toast({ toast, onDismiss }) {
  const [paused, setPaused] = useState(false);
  const remaining = useRef(toast.duration);
  const startedAt = useRef(0);
  const timeoutRef = useRef();
  const Icon = toastIcons[toast.type] ?? Info;

  useEffect(() => {
    if (paused) {
      remaining.current -= Date.now() - startedAt.current;
      window.clearTimeout(timeoutRef.current);
      return undefined;
    }

    startedAt.current = Date.now();
    timeoutRef.current = window.setTimeout(() => onDismiss(toast.id), remaining.current);
    return () => window.clearTimeout(timeoutRef.current);
  }, [onDismiss, paused, toast.id]);

  return (
    <article
      className={`ui-toast ui-toast--${toast.type}`}
      role={toast.type === "error" ? "alert" : "status"}
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => setPaused(false)}
    >
      <Icon className="ui-toast__icon" aria-hidden="true" size={22} />
      <div className="ui-toast__content">
        <strong>{toast.title}</strong>
        {toast.description && <p>{toast.description}</p>}
        {toast.action && (
          <button className="ui-toast__action" type="button" onClick={toast.action.onClick}>
            {toast.action.label}
          </button>
        )}
      </div>
      <button className="ui-icon-button ui-toast__close" type="button" onClick={() => onDismiss(toast.id)} aria-label="ปิดการแจ้งเตือน">
        <X size={18} aria-hidden="true" />
      </button>
      <span className={`ui-toast__progress ${paused ? "is-paused" : ""}`} style={{ animationDuration: `${toast.duration}ms` }} />
    </article>
  );
}

export default function ToastContainer({ toasts, onDismiss }) {
  return (
    <section className="ui-toast-container" aria-label="การแจ้งเตือนของระบบ" aria-live="polite" aria-relevant="additions">
      {toasts.map((toast) => <Toast key={toast.id} toast={toast} onDismiss={onDismiss} />)}
    </section>
  );
}
