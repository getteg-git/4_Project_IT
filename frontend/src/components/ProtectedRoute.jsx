import { useCallback, useEffect, useState } from "react";
import { Navigate } from "react-router-dom";
import { API_BASE_URL, apiFetch } from "../api/baseUrl.js";

function ProtectedRoute({ role, children }) {
  const [state, setState] = useState({ status: "loading", error: "" });
  const [attempt, setAttempt] = useState(0);

  const verifySession = useCallback(async () => {
    setState({ status: "loading", error: "" });
    try {
      const response = await apiFetch(`${API_BASE_URL}/api/auth/me`);
      if (response.status === 401 || response.status === 403) {
        localStorage.removeItem("user");
        setState({ status: "denied", error: "" });
        return;
      }
      if (!response.ok) {
        throw new Error("ไม่สามารถตรวจสอบสิทธิ์ผู้ใช้ได้");
      }
      const data = await response.json();
      if (data.user?.role !== role) {
        setState({ status: "denied", error: "" });
        return;
      }
      localStorage.setItem("user", JSON.stringify(data.user));
      setState({ status: "allowed", error: "" });
    } catch (error) {
      setState({
        status: "error",
        error: error.message || "ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้",
      });
    }
  }, [role]);

  useEffect(() => {
    verifySession();
  }, [verifySession, attempt]);

  if (state.status === "denied") return <Navigate to="/" replace />;
  if (state.status === "error") {
    return (
      <main role="alert" style={{ maxWidth: 520, margin: "15vh auto", padding: 24, textAlign: "center" }}>
        <p>{state.error}</p>
        <button type="button" onClick={() => setAttempt((value) => value + 1)}>
          ลองอีกครั้ง
        </button>
      </main>
    );
  }
  if (state.status === "loading") {
    return <main aria-live="polite" style={{ padding: 32, textAlign: "center" }}>กำลังตรวจสอบสิทธิ์...</main>;
  }

  return children;
}

export default ProtectedRoute;
