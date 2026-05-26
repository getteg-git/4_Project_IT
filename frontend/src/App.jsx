import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom"; // อย่าลืม import Navigate เพิ่มนะ

import Login from "./pages/auth/Login/Login";
import Register from "./pages/auth/Register/Register";
import Users from "./pages/Users/Users";
import UserHome from "./pages/User/UserHome/UserHome";
import CreateRepair from "./pages/User/CreateRepair/CreateRepair";
import MyRepairs from "./pages/User/MyRepairs/MyRepairs";
import AdminHome from "./pages/Admin/AdminHome/AdminHome";
import TechHome from "./pages/Technician/TechHome/TechHome";

function App() {
  return (
    <BrowserRouter>
      <Routes>

        {/* Auth */}
        <Route path="/" element={<Navigate to="/login" replace />} />
        
        {/* เพิ่ม path /login ให้เรียกหน้า Login */}
        <Route path="/login" element={<Login />} />
        
        <Route path="/register" element={<Register />} />

        {/* Old Users page */}
        <Route path="/users" element={<Users />} />

        {/* User Repair System */}
        <Route path="/user/home" element={<UserHome />} />
        <Route path="/repair/create" element={<CreateRepair />} />
        <Route path="/repair/history" element={<MyRepairs />} />

        {/* Admin */}
        <Route path="/admin/home" element={<AdminHome />} />
        
        {/* Technician */}
        <Route path="/tech/home" element={<TechHome />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;