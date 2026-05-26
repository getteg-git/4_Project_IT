import { useEffect, useState, useRef } from "react";
import { useNavigate } from "react-router-dom";
import api from "../../../api/axios";
import "./CreateRepair.css";

function CreateRepair() {
  const [locations, setLocations] = useState([]);
  const [problemTypes, setProblemTypes] = useState([]);

  const [locationId, setLocationId] = useState("");
  const [problemTypeId, setProblemTypeId] = useState("");
  const [description, setDescription] = useState("");
  const [images, setImages] = useState([]);
  
  // เพิ่ม State สำหรับจัดการ Error
  const [error, setError] = useState(""); 
  
  // เพิ่ม Ref สำหรับอ้างอิงถึง input file ที่เราจะซ่อนไว้
  const fileInputRef = useRef(null);
  const cameraInputRef = useRef(null);

  const navigate = useNavigate();

  useEffect(() => {
    fetchDropdownData();
  }, []);

  const fetchDropdownData = async () => {
    try {
      const locRes = await api.get("/locations");
      const probRes = await api.get("/problem-types");
      setLocations(locRes.data);
      setProblemTypes(probRes.data);
    } catch (err) {
      console.error(err);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    // 1. Validation: เช็คว่ากรอกข้อมูลสำคัญครบไหม
    if (!locationId || !problemTypeId || !description) {
      setError("กรุณาเลือกสถานที่, ประเภทปัญหา และระบุรายละเอียดให้ครบถ้วน");
      return;
    }

    try {
      const formData = new FormData();

      // 2. ดึง user_id จริงๆ จาก localStorage
      const currentUser = JSON.parse(localStorage.getItem("user"));
      const userId = currentUser ? currentUser.id : 1; 
      
      formData.append("user_id", userId);
      formData.append("location_id", locationId);
      formData.append("problem_type_id", problemTypeId);
      formData.append("description", description);

      images.forEach((img) => {
        formData.append("images", img);
      });

      await api.post("/repairs", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });

      // ถ้าผ่านเคลียร์ Error แล้วกลับหน้า Home
      setError("");
      navigate("/user/home");

    } catch (err) {
      console.error(err);
      setError("เกิดข้อผิดพลาดในการแจ้งซ่อม กรุณาลองใหม่อีกครั้ง");
    }
  };

  const handleImageChange = (e) => {
    const files = Array.from(e.target.files);
    setImages((prev) => [...prev, ...files]);
  };

  const removeImage = (index) => {
    setImages(images.filter((_, i) => i !== index));
  };

  return (
    <div className="repair-container">
      <h2>แจ้งซ่อม</h2>

      {/* แสดง Error Message ถ้ามี */}
      {error && <p className="error-message">{error}</p>}

      <form onSubmit={handleSubmit}>
        
        <label>สถานที่</label>
        <select
          value={locationId}
          onChange={(e) => setLocationId(e.target.value)}
        >
          <option value="">-- เลือกสถานที่ --</option>
          {locations.map((loc) => (
            <option key={loc.id} value={loc.id}>
              {loc.name}
            </option>
          ))}
        </select>

        <label>ประเภทปัญหา</label>
        <select
          value={problemTypeId}
          onChange={(e) => setProblemTypeId(e.target.value)}
        >
          <option value="">-- เลือกประเภทปัญหา --</option>
          {problemTypes.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name}
            </option>
          ))}
        </select>

        <label>รายละเอียด</label>
        <textarea
          rows="5"
          placeholder="เช่น แอร์น้ำหยด, ปลั๊กไฟช็อต, ระบุเลขห้องให้ชัดเจน..."
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />

        <label>แนบรูปภาพ (ประกอบการซ่อม)</label>
        
        {/* ซ่อน input ตัวเลือกรูปปกติ (รองรับการเลือกหลายรูป) */}
        <input
          type="file"
          multiple
          accept="image/*"
          ref={fileInputRef}
          onChange={handleImageChange}
          style={{ display: "none" }}
        />

        {/* ซ่อน input ตัวเปิดกล้องถ่ายรูป (ใช้ capture="environment" เปิดกล้องหลัง) */}
        <input
          type="file"
          accept="image/*"
          capture="environment"
          ref={cameraInputRef}
          onChange={handleImageChange}
          style={{ display: "none" }}
        />

        {/* สร้างปุ่ม 2 ปุ่มเพื่อเรียกใช้งาน input ที่ซ่อนไว้ */}
        <div style={{ display: "flex", gap: "10px", marginBottom: "15px" }}>
          <button
            type="button"
            className="cancel-btn" // ใช้ class เดิมเพื่อยืมสไตล์ชั่วคราว หรือแก้ CSS เองทีหลังได้
            onClick={() => fileInputRef.current.click()}
            style={{ flex: 1, padding: "10px", backgroundColor: "#6c757d", color: "white" }}
          >
            🖼️ เลือกรูปภาพ
          </button>
          <button
            type="button"
            className="submit-btn" // ใช้ class เดิมเพื่อยืมสไตล์ชั่วคราว หรือแก้ CSS เองทีหลังได้
            onClick={() => cameraInputRef.current.click()}
            style={{ flex: 1, padding: "10px", backgroundColor: "#0d6efd", color: "white" }}
          >
            📸 ถ่ายภาพ
          </button>
        </div>

        {/* ส่วน Preview รูปภาพ */}
        <div className="preview-container">
          {images.map((img, index) => (
            <div key={index} className="preview-item">
              <img src={URL.createObjectURL(img)} alt="preview" />
              <button
                type="button"
                className="remove-btn"
                onClick={() => removeImage(index)}
              >
                ลบ
              </button>
            </div>
          ))}
        </div>

        {/* ปุ่ม Submit และ ยกเลิก */}
        <div className="button-group">
          <button type="submit" className="submit-btn">แจ้งซ่อม</button>
          <button 
            type="button" 
            className="cancel-btn" 
            onClick={() => navigate("/user/home")}
          >
            ยกเลิก
          </button>
        </div>

      </form>
    </div>
  );
}

export default CreateRepair;