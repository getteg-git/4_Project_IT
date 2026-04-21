import { useNavigate } from "react-router-dom";
import "./UserHome.css";

function UserHome() {

  const navigate = useNavigate();

  return (
    <div className="user-home-container">

      <h1 className="title">SC-REP ระบบแจ้งซ่อม</h1>

      <div className="button-group">

        <button
          className="btn-primary"
          onClick={() => navigate("/repair/create")}
        >
          แจ้งซ่อม
        </button>

        <button
          className="btn-secondary"
          onClick={() => navigate("/repair/my")}
        >
          ดูรายการแจ้งซ่อมของฉัน
        </button>

      </div>

    </div>
  );
}

export default UserHome;