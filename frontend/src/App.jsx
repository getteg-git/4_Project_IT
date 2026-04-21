import { BrowserRouter, Routes, Route } from "react-router-dom";

import Login from "./pages/auth/Login/Login";
import Register from "./pages/auth/Register/Register";
import Users from "./pages/Users/Users";
import UserHome from "./pages/User/UserHome/UserHome";
import CreateRepair from "./pages/User/CreateRepair/CreateRepair";

function App() {
  return (
    <BrowserRouter>

      <Routes>

        {/* Auth */}
        <Route path="/" element={<Login />} />
        <Route path="/register" element={<Register />} />

        {/* Old Users page */}
        <Route path="/users" element={<Users />} />

        {/* User Repair System */}
        <Route path="/user/home" element={<UserHome />} />
        <Route path="/repair/create" element={<CreateRepair />} />

      </Routes>

    </BrowserRouter>
  );
}

export default App;