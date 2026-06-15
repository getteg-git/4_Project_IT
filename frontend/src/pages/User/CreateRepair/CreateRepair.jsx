import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import "./CreateRepair.css";

function CreateRepair() {
  const navigate = useNavigate();
  
  // State สำหรับเก็บข้อมูลที่ผู้ใช้กรอก
  const [email, setEmail] = useState("");
  const [location, setLocation] = useState("");
  const [floor, setFloor] = useState("");
  const [problemType, setProblemType] = useState("");
  const [details, setDetails] = useState("");
  const [image, setImage] = useState(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // State สำหรับเก็บข้อมูลจาก Backend (ห้าม Mockup)
  const [locations, setLocations] = useState([]);
  const [floors, setFloors] = useState([]);
  const [problemTypes, setProblemTypes] = useState([]);

  // 1. ดึงข้อมูลตึกและประเภทปัญหา เมื่อเปิดหน้าเว็บครั้งแรก
  useEffect(() => {
    const fetchMasterData = async () => {
      try {
        const [locRes, typeRes] = await Promise.all([
          fetch("http://localhost:8080/api/locations"),
          fetch("http://localhost:8080/api/problem-types")
        ]);
        
        if (locRes.ok) setLocations(await locRes.json());
        if (typeRes.ok) setProblemTypes(await typeRes.json());
      } catch (err) {
        console.error("ดึงข้อมูลหลักไม่สำเร็จ:", err);
      }
    };
    fetchMasterData();
  }, []);

  // 2. ดึงข้อมูล "ชั้น" เมื่อผู้ใช้ทำการเลือก "สถานที่ (ตึก)"
  useEffect(() => {
    if (!location) {
      setFloors([]);
      setFloor("");
      return;
    }

    const fetchFloors = async () => {
      try {
        const res = await fetch(`http://localhost:8080/api/locations/${location}/floors`);
        if (res.ok) {
          const data = await res.json();
          setFloors(data || []);
        }
      } catch (err) {
        console.error("ดึงข้อมูลชั้นไม่สำเร็จ:", err);
      }
    };
    
    fetchFloors();
  }, [location]);

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!email.endsWith("@su.ac.th")) {
      alert("⚠️ กรุณาใช้อีเมลของมหาวิทยาลัย (@su.ac.th) เท่านั้นครับ");
      return;
    }

    setIsSubmitting(true);

    const formData = new FormData();
    formData.append("reporter_email", email); // ชื่อ key ต้องตรงกับ backend (reporter_email)
    formData.append("location_id", location);
    formData.append("floor_id", floor);
    formData.append("problem_type_id", problemType);
    formData.append("description", details); // ชื่อ key ต้องตรงกับ backend (description)
    
    if (image) {
      formData.append("image", image);
    }

    try {
      const response = await fetch("http://localhost:8080/api/repairs", {
        method: "POST",
        body: formData, 
      });

      if (response.ok) {
        alert("✅ ระบบได้รับเรื่องแจ้งซ่อมของคุณเรียบร้อยแล้ว");
        navigate("/repair/history"); // ส่งไปหน้าติดตามสถานะ
      } else {
        const errorData = await response.json();
        alert(`❌ เกิดข้อผิดพลาด: ${errorData.error || "ไม่สามารถส่งข้อมูลได้"}`);
      }
    } catch (error) {
      console.error("Error:", error);
      alert("❌ ไม่สามารถติดต่อเซิร์ฟเวอร์ได้ กรุณาลองใหม่อีกครั้ง");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleImageChange = (e) => {
    if (e.target.files && e.target.files[0]) {
      setImage(e.target.files[0]);
    }
  };

  return (
    <div className="create-repair-container">
      <div className="repair-card">
        <div className="repair-header">
          <h2>ฟอร์มแจ้งปัญหา / งานซ่อมบำรุง</h2>
          <p>กรุณากรอกข้อมูลให้ครบถ้วนเพื่อให้ช่างเทคนิคเข้าดำเนินการได้อย่างรวดเร็ว</p>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>E-mail ผู้แจ้ง (@su.ac.th) <span className="required">*</span></label>
            <input
              type="email"
              placeholder="ระบุอีเมลมหาวิทยาลัยของคุณ"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>

          <div className="form-group">
            <label>อาคาร / สถานที่ <span className="required">*</span></label>
            <select value={location} onChange={(e) => setLocation(e.target.value)} required>
              <option value="">-- กรุณาเลือกอาคาร --</option>
              {locations.map((loc) => (
                <option key={loc.id} value={loc.id}>{loc.name}</option>
              ))}
            </select>
          </div>

          {/* ซ่อนช่องเลือกชั้นจนกว่าจะมีการเลือกอาคาร */}
          {location && (
            <div className="form-group slide-down">
              <label>ชั้น <span className="required">*</span></label>
              <select value={floor} onChange={(e) => setFloor(e.target.value)} required>
                <option value="">-- กรุณาเลือกชั้น --</option>
                {floors.length > 0 ? (
                  floors.map((f) => (
                    <option key={f.id} value={f.id}>{f.floor_name}</option>
                  ))
                ) : (
                  <option value="" disabled>ไม่มีข้อมูลชั้นสำหรับอาคารนี้</option>
                )}
              </select>
            </div>
          )}

          <div className="form-group">
            <label>หมวดหมู่งานซ่อม <span className="required">*</span></label>
            <select value={problemType} onChange={(e) => setProblemType(e.target.value)} required>
              <option value="">-- กรุณาเลือกหมวดหมู่ปัญหา --</option>
              {problemTypes.map((type) => (
                <option key={type.id} value={type.id}>{type.name}</option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label>อธิบายรายละเอียดของปัญหา <span className="required">*</span></label>
            <textarea
              placeholder="เช่น แอร์น้ำหยดตรงมุมห้อง, หลอดไฟกะพริบ, หรือพิกัดจุดเกิดเหตุที่ชัดเจน"
              rows="4"
              value={details}
              onChange={(e) => setDetails(e.target.value)}
              required
            ></textarea>
          </div>

          <div className="form-group">
            <label>แนบรูปภาพประกอบ (ถ้ามี)</label>
            <input type="file" accept="image/*" onChange={handleImageChange} className="file-input" id="file-upload" />
            <label htmlFor="file-upload" className="file-upload-label">
              📸 คลิกเพื่อเลือกรูปภาพ หรือ ลากไฟล์มาวางที่นี่
            </label>
            {image && <p className="file-name-preview">✔️ ไฟล์ที่เลือก: {image.name}</p>}
          </div>

          <div className="form-actions">
            <button 
              type="button" 
              className="btn-cancel" 
              onClick={() => navigate("/")}
              disabled={isSubmitting}
            >
              ❌ ยกเลิก และกลับหน้าแรก
            </button>
            <button 
              type="submit" 
              className="btn-submit"
              disabled={isSubmitting}
            >
              {isSubmitting ? "⏳ กำลังส่งข้อมูล..." : "✅ ยืนยัน และส่งข้อมูลแจ้งซ่อม"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default CreateRepair;