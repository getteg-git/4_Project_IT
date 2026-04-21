import { useState } from "react";
import api from "../../../api/axios";
import "./Register.css";

function Register({ setPage }) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

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

      setSuccess("สมัครสมาชิกสำเร็จ 🎉");
      setError("");

      setUsername("");
      setPassword("");

      // กลับไปหน้า Login หลัง 1.5 วิ
      setTimeout(() => {
        setPage("login");
      }, 1500);

    } catch (err) {
      console.error(err);
      setError("สมัครไม่สำเร็จ (username ซ้ำ?)");
      setSuccess("");
    }
  };

  return (
    <div className="register-container">
      <h2>Register</h2>

      {error && <p className="error">{error}</p>}
      {success && <p className="success">{success}</p>}

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

      <p className="link" onClick={() => setPage("login")}>
        Already have account? Login
      </p>
    </div>
  );
}

export default Register;