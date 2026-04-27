import { useNavigate } from "react-router-dom";
import "./UserHome.css";

function UserHome() {
  const navigate = useNavigate();

  // ฟังก์ชันสำหรับออกจากระบบ
  const handleLogout = () => {
    localStorage.removeItem("user"); // ล้างข้อมูล user ที่ล็อกอินค้างไว้
    navigate("/login"); // เด้งกลับไปหน้า Login
  };

  return (
    <div className="user-home-container">
      
      {/* จัดกลุ่ม Header เพื่อให้อยู่บรรทัดเดียวกันและมีปุ่ม Logout มุมขวา */}
      <div className="user-header">
        <h1 className="title">SC-REP ระบบแจ้งซ่อม</h1>
        <button className="logout-btn" onClick={handleLogout}>
          ออกจากระบบ
        </button>
      </div>

      <div className="button-group">
        <button
          className="btn-primary"
          onClick={() => navigate("/repair/create")}
        >
          แจ้งซ่อม
        </button>

        <button
          className="btn-secondary"
          onClick={() => navigate("/repair/history")}
        >
          ดูรายการแจ้งซ่อมของฉัน
        </button>
      </div>

    </div>
  );
}

export default UserHome;