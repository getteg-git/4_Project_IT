import { useState } from "react";
import { useNavigate } from "react-router-dom";
import api from "../../../api/axios";
import "./Register.css";

function Register() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  const navigate = useNavigate();

  const handleRegister = async () => {
    if (!username || !password) {
      setError("กรุณากรอกข้อมูลให้ครบ");
      return;
    }

    try {
      await api.post("/register", {
        username,
        password,
      });

      setError("");
      setUsername("");
      setPassword("");

      // เด้งไปหน้า login ทันทีโดยไม่ต้องรอ
      navigate("/login");

    } catch (err) {
      console.error(err);
      setError("สมัครไม่สำเร็จ (username ซ้ำ?)");
    }
  };

  return (
    <div className="register-container">
      <h2>Register</h2>

      {error && <p className="error">{error}</p>}

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

      <button onClick={handleRegister}>Register</button>

      <p className="link" onClick={() => navigate("/login")}>
        Already have account? Login
      </p>
    </div>
  );
}

export default Register;