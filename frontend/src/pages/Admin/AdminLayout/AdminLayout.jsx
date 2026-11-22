import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { ClipboardList, LayoutDashboard, LogOut, Settings } from "lucide-react";
import useToast from "../../../hooks/useToast";
import { API_BASE_URL, apiFetch as fetch } from "../../../api/baseUrl.js";
import "./AdminLayout.css";

function AdminLayout() {
  const navigate = useNavigate();
  const { confirm, toast } = useToast();

  const handleLogout = async () => {
    const approved = await confirm({
      title: "ยืนยันออกจากระบบ",
      description: "คุณแน่ใจหรือไม่ว่าต้องการออกจากระบบผู้ดูแล?",
      confirmLabel: "ออกจากระบบ",
      cancelLabel: "ยกเลิก",
      variant: "danger",
    });

    if (approved) {
      try {
        const response = await fetch(`${API_BASE_URL}/api/logout`, { method: "POST" });
        if (!response.ok) throw new Error("ไม่สามารถออกจากระบบได้ กรุณาลองอีกครั้ง");
        localStorage.removeItem("user");
        navigate("/", { replace: true });
      } catch (error) {
        toast.error("ออกจากระบบไม่สำเร็จ", { description: error.message });
      }
    }
  };

  return (
    <div className="admin-layout">
      <header className="admin-shell-header">
        <div className="admin-shell-brand">ระบบจัดการงานซ่อม</div>
        <nav className="admin-shell-navigation" aria-label="เมนูผู้ดูแลระบบ">
          <NavLink to="/admin/manage" className={({ isActive }) => isActive ? "is-active" : ""}>
            <ClipboardList size={18} aria-hidden="true" />
            <span>มอบหมายงาน</span>
          </NavLink>
          <NavLink to="/admin/home" className={({ isActive }) => isActive ? "is-active" : ""}>
            <LayoutDashboard size={18} aria-hidden="true" />
            <span>แดชบอร์ด</span>
          </NavLink>
          <NavLink to="/admin/settings" className={({ isActive }) => isActive ? "is-active" : ""}>
            <Settings size={18} aria-hidden="true" />
            <span>ตั้งค่าระบบ</span>
          </NavLink>
        </nav>
        <button className="admin-shell-logout" type="button" onClick={handleLogout}>
          <LogOut size={17} aria-hidden="true" />
          <span>ออกจากระบบ</span>
        </button>
      </header>
      <Outlet />
    </div>
  );
}

export default AdminLayout;
