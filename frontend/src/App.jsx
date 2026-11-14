import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";

import UserHome from "./pages/User/UserHome/UserHome";
import CreateRepair from "./pages/User/CreateRepair/CreateRepair";
import MyRepairs from "./pages/User/MyRepairs/MyRepairs";
import TechHome from "./pages/Technician/TechHome/TechHome";
import AdminDashboard from "./pages/Admin/AdminDashboard/AdminDashboard"; 
import AdminManage from "./pages/Admin/AdminManage/AdminManage";
import AdminSettings from "./pages/Admin/AdminSettings/AdminSettings";
import AdminLayout from "./pages/Admin/AdminLayout/AdminLayout";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<UserHome />} />

        {/* ฝั่งผู้ใช้งานทั่วไป */}
        <Route path="/repair/create" element={<CreateRepair />} />
        <Route path="/repair/history" element={<MyRepairs />} />

        {/* ฝั่งแอดมิน */}
        <Route path="/admin" element={<AdminLayout />}>
          <Route index element={<Navigate to="/admin/manage" replace />} />
          <Route path="home" element={<AdminDashboard />} />
          <Route path="manage" element={<AdminManage />} />
          <Route path="settings" element={<AdminSettings />} />
        </Route>
        
        {/* ฝั่งช่าง */}
        <Route path="/tech/home" element={<TechHome />} />


        {/* หากผู้ใช้เข้า URL ที่ไม่มีอยู่จริง (เช่น /user/home, /login) ให้เด้งกลับมาหน้า "/" ทันที */}
        <Route path="*" element={<Navigate to="/" replace />} />

      </Routes>
    </BrowserRouter>
  );
}

export default App;