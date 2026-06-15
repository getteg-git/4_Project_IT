import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom"; // นำ Navigate กลับมาใช้ดัก URL ผิด

import UserHome from "./pages/User/UserHome/UserHome";
import CreateRepair from "./pages/User/CreateRepair/CreateRepair";
import MyRepairs from "./pages/User/MyRepairs/MyRepairs";
import AdminHome from "./pages/Admin/AdminHome/AdminHome";
import TechHome from "./pages/Technician/TechHome/TechHome";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* ==========================================
            ✅ ROUTES ที่เปิดใช้งาน (Active Routes) 
            ========================================== */}
            
        {/* หน้า Landing Page หลัก (พิมพ์แค่ localhost:5173 ก็จะเจอหน้านี้เลย) */}
        <Route path="/" element={<UserHome />} />

        {/* ฝั่งผู้ใช้งานทั่วไป (ไม่ต้องเข้าสู่ระบบ) */}
        <Route path="/repair/create" element={<CreateRepair />} />
        <Route path="/repair/history" element={<MyRepairs />} />

        {/* ฝั่งเจ้าหน้าที่ (เข้าสู่ระบบแล้วถึงจะเด้งมาหน้านี้) */}
        <Route path="/admin/home" element={<AdminHome />} />
        <Route path="/tech/home" element={<TechHome />} />


        {/* ==========================================
            🛡️ ROUTE ดักจับ Error (Catch-all)
            ========================================== */}
        {/* หากผู้ใช้เข้า URL ที่ไม่มีอยู่จริง (เช่น /user/home, /login) ให้เด้งกลับมาหน้า "/" ทันที */}
        <Route path="*" element={<Navigate to="/" replace />} />

      </Routes>
    </BrowserRouter>
  );
}

export default App;