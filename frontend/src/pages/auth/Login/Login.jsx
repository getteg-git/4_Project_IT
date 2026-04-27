import { useState } from "react";
import { useNavigate } from "react-router-dom";
import api from "../../../api/axios";
import "./Login.css";

function Login() {
  const navigate = useNavigate();

  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  const handleLogin = async (e) => {
    e.preventDefault();

    try {
      const res = await api.post("/login", {
        username,
        password,
      });

      console.log("LOGIN RESPONSE:", res.data);

      setError("");

      // 1. บันทึกข้อมูล User ลง localStorage เพื่อให้หน้าอื่น (เช่น หน้าช่าง) ดึงไปใช้งานได้
      localStorage.setItem("user", JSON.stringify(res.data.user));

      // 2. เช็ค Role แล้วแยกทางไปตามหน้าของแต่ละคน
      if (res.data.user.role === "admin") {
        navigate("/admin/home");
      } else if (res.data.user.role === "technician") {
        navigate("/tech/home"); // พาช่างเทคนิคไปที่หน้า TechHome
      } else {
        navigate("/user/home");
      }

    } catch (err) {
      console.log("LOGIN ERROR:", err.response);
      setError("Username หรือ Password ไม่ถูกต้อง");
    }
  };

  return (
    <div className="login-container">
      <h2>Login</h2>

      {error && <p className="error">{error}</p>}

      <form onSubmit={handleLogin}>
        <input
          type="text"
          placeholder="Username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />

        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />

        <button type="submit">Login</button>
      </form>

      <p className="link" onClick={() => navigate("/register")}>
        No account? Register
      </p>
    </div>
  );
}

export default Login;