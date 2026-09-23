import { useCallback, useMemo, useState } from "react";
import ConfirmDialog from "../components/ui/ConfirmDialog";
import ToastContainer from "../components/ui/ToastContainer";
import { ToastContext } from "./ToastContext";

const defaultDurations = { success: 4000, info: 5000, warning: 6000, error: 7000 };

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([]);
  const [dialog, setDialog] = useState(null);

  const dismiss = useCallback((id) => {
    setToasts((current) => current.filter((toast) => toast.id !== id));
  }, []);

  const show = useCallback((type, title, options = {}) => {
    const id = crypto.randomUUID();
    setToasts((current) => [...current, {
      id,
      type,
      title,
      description: options.description,
      action: options.action,
      duration: options.duration ?? defaultDurations[type],
    }].slice(-3));
    return id;
  }, []);

  const confirm = useCallback((options) => new Promise((resolve) => {
    setDialog({ ...options, resolve });
  }), []);

  const closeDialog = useCallback((result) => {
    setDialog((current) => {
      current?.resolve(result);
      return null;
    });
  }, []);

  const value = useMemo(() => ({
    toast: {
      success: (title, options) => show("success", title, options),
      info: (title, options) => show("info", title, options),
      warning: (title, options) => show("warning", title, options),
      error: (title, options) => show("error", title, options),
    },
    confirm,
  }), [confirm, show]);

  return (
    <ToastContext.Provider value={value}>
      {children}
      <ToastContainer toasts={toasts} onDismiss={dismiss} />
      <ConfirmDialog dialog={dialog} onClose={closeDialog} />
    </ToastContext.Provider>
  );
}
