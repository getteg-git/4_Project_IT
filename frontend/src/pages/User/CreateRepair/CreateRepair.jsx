import { API_BASE_URL, apiFetch as fetch } from "../../../api/baseUrl.js";
import React, { useState, useEffect, useRef, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { Camera, CheckCircle2, Image as ImageIcon, LoaderCircle, Send } from 'lucide-react';
import AppBreadcrumb from "../../../components/AppBreadcrumb";
import BackButton from "../../../components/ui/BackButton";
import useToast from "../../../hooks/useToast";
import "./CreateRepair.css";

const reporterTypes = ["นักศึกษา", "อาจารย์", "เจ้าหน้าที่", "บุคลากรภายนอก"];
const internalReporterTypes = ["นักศึกษา", "อาจารย์", "เจ้าหน้าที่"];

function CreateRepair() {
  const navigate = useNavigate();
  const { toast } = useToast();

  // State สำหรับเก็บข้อมูลที่ผู้ใช้กรอก
  const [email, setEmail] = useState("");
  const [reporterType, setReporterType] = useState("");
  const [otherReporterDepartment, setOtherReporterDepartment] = useState("");
  const [degreeLevel, setDegreeLevel] = useState(""); 
  const [department, setDepartment] = useState(""); 
  const [location, setLocation] = useState("");
  const [floor, setFloor] = useState("");
  const [room, setRoom] = useState("");
  const [customRoomName, setCustomRoomName] = useState(""); // 🔥 [เพิ่มใหม่] State เก็บชื่อห้องที่ผู้ใช้พิมพ์เองกรณีเลือก "อื่นๆ"
  const [equipment, setEquipment] = useState("");
  const [customLocationName, setCustomLocationName] = useState("");
  const [problemType, setProblemType] = useState("");
  const [customProblemName, setCustomProblemName] = useState("");
  const [details, setDetails] = useState("");
  const [image, setImage] = useState(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSubmitSuccess, setIsSubmitSuccess] = useState(false);
  const [validationErrors, setValidationErrors] = useState({});

  // State สำหรับเก็บข้อมูลจาก Backend
  const [degreeLevels, setDegreeLevels] = useState([]); 
  const [departments, setDepartments] = useState([]); 
  const [locations, setLocations] = useState([]);
  const [floors, setFloors] = useState([]);
  const [isFloorDataLoaded, setIsFloorDataLoaded] = useState(false);
  const [floorLoadError, setFloorLoadError] = useState(false);
  const [rooms, setRooms] = useState([]);
  const [equipments, setEquipments] = useState([]);
  const [problemTypes, setProblemTypes] = useState([]);

  // Ref สำหรับอ้างอิง Input File ที่ซ่อนไว้
  const fileInputRef = useRef(null);
  const cameraInputRef = useRef(null);
  const submitInProgressRef = useRef(false);

  const isOtherLocation = useCallback(() => {
    const selectedLoc = locations.find(loc => String(loc.id) === String(location));
    return selectedLoc && (selectedLoc.name.includes("อื่นๆ") || selectedLoc.name.toLowerCase() === "other");
  }, [location, locations]);

  // 🔥 [เพิ่มใหม่] เช็กว่าเลือกห้องแบบ "อื่นๆ" หรือไม่
  const isOtherRoom = useCallback(() => {
    if (room === "other") return true;
    const selectedRoom = rooms.find(r => String(r.id) === String(room));
    return selectedRoom && (selectedRoom.room_number.includes("อื่นๆ") || selectedRoom.room_number.toLowerCase() === "other");
  }, [room, rooms]);

  const isOtherProblem = () => {
    const selectedType = problemTypes.find(type => String(type.id) === String(problemType));
    return selectedType && (selectedType.name.includes("อื่นๆ") || selectedType.name.toLowerCase() === "other");
  };

  const normalizeEquipmentName = (name) =>
    name.replace(/\s*\(\s*\d[^)]*\)/g, "").replace(/\s+/g, " ").trim();

  const uniqueEquipments = [...new Map(
    equipments
      .map((item) => ({ ...item, displayName: normalizeEquipmentName(item.name || "") }))
      .filter((item) => item.displayName)
      .map((item) => [item.displayName.toLocaleLowerCase(), item])
  ).values()];

  // 1. โหลดข้อมูลเริ่มต้น (สถานที่, หมวดหมู่ปัญหา, ระดับการศึกษา)
  useEffect(() => {
    const fetchMasterData = async () => {
      try {
        const [locRes, typeRes, degRes] = await Promise.all([
          fetch(API_BASE_URL + "/api/locations"),
          fetch(API_BASE_URL + "/api/problem-types"),
          fetch(API_BASE_URL + "/api/degree-levels")
        ]);

        if (locRes.ok) setLocations(await locRes.json());
        if (typeRes.ok) setProblemTypes(await typeRes.json());
        if (degRes.ok) setDegreeLevels(await degRes.json());
      } catch (err) {
        console.error("ดึงข้อมูลหลักไม่สำเร็จ:", err);
      }
    };
    fetchMasterData();
  }, []);

  // Load department options from the database for the selected reporter type.
  useEffect(() => {
    setDepartment("");
    setDepartments([]);

    if (!internalReporterTypes.includes(reporterType) || (reporterType === "นักศึกษา" && !degreeLevel)) {
      return;
    }

    const fetchDepartments = async () => {
      try {
        const endpoint = reporterType === "นักศึกษา"
          ? `${API_BASE_URL}/api/degree-levels/${degreeLevel}/departments`
          : `${API_BASE_URL}/api/departments`;
        const res = await fetch(endpoint);
        if (!res.ok) throw new Error(`ไม่สามารถโหลดภาควิชาได้ (${res.status})`);
        setDepartments(await res.json() || []);
      } catch (err) {
        console.error("ดึงข้อมูลภาควิชาไม่สำเร็จ:", err);
      }
    };
    fetchDepartments();
  }, [degreeLevel, reporterType]);

  useEffect(() => {
    setFloor("");
    setRooms([]);
    setRoom("");
    setCustomRoomName("");
    setEquipments([]);
    setEquipment("");

    if (!location || isOtherLocation()) {
      setFloors([]);
      setIsFloorDataLoaded(true);
      setFloorLoadError(false);
      return;
    }

    setIsFloorDataLoaded(false);
    setFloorLoadError(false);
    const fetchFloors = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/locations/${location}/floors`);
        if (!res.ok) throw new Error(`ไม่สามารถโหลดข้อมูลชั้นได้ (${res.status})`);
        setFloors(await res.json() || []);
      } catch (err) {
        console.error("ดึงข้อมูลชั้นไม่สำเร็จ:", err);
        setFloorLoadError(true);
      } finally {
        setIsFloorDataLoaded(true);
      }
    };
    fetchFloors();
  }, [location, isOtherLocation]);

  useEffect(() => {
    setRoom("");
    setCustomRoomName("");
    setEquipments([]);
    setEquipment("");

    if (!floor) {
      setRooms([]);
      return;
    }

    const fetchRooms = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/floors/${floor}/rooms`);
        if (res.ok) {
          const data = await res.json() || [];
          // 🔥 เพิ่มตัวเลือก "อื่นๆ" ต่อท้ายรายการห้องเสมอ เพื่อให้ผู้ใช้เลือกกรอกเองได้
          setRooms([...data, { id: "other", room_number: "อื่นๆ (โปรดระบุ)" }]);
        }
      } catch (err) {
        console.error("ดึงข้อมูลห้องไม่สำเร็จ:", err);
        // เผื่อ API ดึงห้องพลาด อย่างน้อยให้มีตัวเลือกอื่นๆ ให้กดพิมพ์เองได้
        setRooms([{ id: "other", room_number: "อื่นๆ (โปรดระบุ)" }]);
      }
    };
    fetchRooms();
  }, [floor]);

  useEffect(() => {
    setEquipment("");

    if (!room || isOtherRoom()) {
      setEquipments([]);
      return;
    }

    const fetchEquipments = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/api/rooms/${room}/equipments`);
        if (res.ok) setEquipments(await res.json() || []);
      } catch (err) {
        console.error("ดึงข้อมูลอุปกรณ์ไม่สำเร็จ:", err);
      }
    };
    fetchEquipments();
  }, [room, isOtherRoom]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (submitInProgressRef.current) return;

    const reporterEmail = email.trim();
    const nextErrors = {};
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(reporterEmail)) nextErrors.email = "กรุณากรอกอีเมลให้ถูกต้อง";
    if (!reporterType) nextErrors.reporterType = "กรุณาเลือกสถานะผู้แจ้ง";
    if (reporterType === "นักศึกษา" && !degreeLevel) nextErrors.degreeLevel = "กรุณาเลือกระดับการศึกษา";
    if (internalReporterTypes.includes(reporterType) && !department) {
      nextErrors.department = "กรุณาเลือกภาควิชาสังกัดจากฐานข้อมูล";
    }
    if (reporterType === "บุคลากรภายนอก" && !otherReporterDepartment.trim()) {
      nextErrors.otherReporterDepartment = "กรุณาระบุหน่วยงานหรือความเกี่ยวข้อง";
    }
    if (!location) nextErrors.location = "กรุณาเลือกอาคารหรือสถานที่";
    if (isOtherLocation() && !customLocationName.trim()) nextErrors.customLocation = "กรุณาระบุรายละเอียดอาคารหรือสถานที่";
    if (location && !isOtherLocation() && !isFloorDataLoaded) nextErrors.floor = "กรุณารอให้โหลดข้อมูลชั้นเสร็จก่อน";
    if (location && !isOtherLocation() && floorLoadError) nextErrors.floor = "ไม่สามารถโหลดข้อมูลชั้นได้ กรุณาลองใหม่";
    if (location && !isOtherLocation() && isFloorDataLoaded && floors.length > 0 && !floor) nextErrors.floor = "กรุณาเลือกชั้น";
    if (location && !isOtherLocation() && isFloorDataLoaded && floors.length === 0 && !customRoomName.trim()) nextErrors.customRoom = "กรุณาระบุห้องหรือจุดเกิดเหตุ";
    if (floor && !room) nextErrors.room = "กรุณาเลือกห้องหรือจุดเกิดเหตุ";
    if (isOtherRoom() && !customRoomName.trim()) nextErrors.customRoom = "กรุณาระบุชื่อห้องหรือจุดเกิดเหตุเพิ่มเติม";
    if (!problemType) nextErrors.problemType = "กรุณาเลือกหมวดหมู่งานซ่อม";
    if (isOtherProblem() && !customProblemName.trim()) nextErrors.customProblem = "กรุณาระบุหมวดหมู่งานซ่อมอื่น ๆ";
    if (!details.trim()) nextErrors.details = "กรุณาอธิบายรายละเอียดของปัญหา";
    setValidationErrors(nextErrors);

    if (Object.keys(nextErrors).length > 0) {
      toast.warning("กรุณาตรวจสอบข้อมูล", { description: "โปรดแก้ไขข้อมูลในช่องที่มีข้อความแจ้งเตือน" });
      return;
    }

    submitInProgressRef.current = true;
    setIsSubmitting(true);
    setIsSubmitSuccess(false);

    const formData = new FormData();
    formData.append("reporter_email", reporterEmail);
    formData.append("reporter_type", reporterType);
    if (internalReporterTypes.includes(reporterType)) {
      const selectedDepartment = departments.find((item) => String(item.id) === String(department));
      formData.append("department_id", department);
      formData.append("reporter_department", selectedDepartment?.name || "");
      if (reporterType === "นักศึกษา") {
        const selectedDegreeLevel = degreeLevels.find((item) => String(item.id) === String(degreeLevel));
        formData.append("reporter_degree_level", selectedDegreeLevel?.name || "");
      }
    } else {
      formData.append("reporter_department", otherReporterDepartment.trim());
    }
    formData.append("location_id", location);
    formData.append("problem_type_id", problemType);
    formData.append("description", details);

    if (isOtherLocation()) {
      formData.append("other_location", customLocationName);
    } else if (floors.length > 0) {
      formData.append("floor_id", floor);
      
      // 🔥 ถ้าเลือกห้องเป็น "อื่นๆ" ให้ส่ง other_room ไปแทน room_id
      if (isOtherRoom()) {
        formData.append("other_room", customRoomName);
      } else {
        formData.append("room_id", room);
        if (equipment) formData.append("equipment_id", equipment);
      }
    } else if (isFloorDataLoaded && !floorLoadError && customRoomName.trim()) {
      formData.append("other_room", customRoomName.trim());
    }

    if (isOtherProblem()) {
      formData.append("other_problem_type", customProblemName);
    }

    if (image) {
      formData.append("image", image);
    }

    try {
      const response = await fetch(API_BASE_URL + "/api/repairs", {
        method: "POST",
        body: formData,
      });

      if (response.ok) {
        setIsSubmitSuccess(true);
        toast.success("รับเรื่องแจ้งซ่อมแล้ว", {
          description: "คุณสามารถติดตามสถานะงานได้จากหน้ารายการแจ้งซ่อม",
        });

        navigate("/repair/history", { replace: true });
      } else {
        const errorData = await response.json();
        toast.error("ส่งรายการไม่สำเร็จ", {
          description: errorData.error || "ไม่สามารถส่งข้อมูลได้",
        });
      }
    } catch (error) {
      console.error("Error:", error);
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    } finally {
      submitInProgressRef.current = false;
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
      <div className="create-repair-shell">
        <nav className="page-navigation" aria-label="การนำทางย้อนกลับ">
          <AppBreadcrumb items={[
            { label: "หน้าหลัก", to: "/" },
            { label: "แจ้งซ่อม" },
          ]} />
          <BackButton to="/" label="กลับหน้าหลัก" />
        </nav>

        <div className="repair-card">
          <div className="repair-header">
            <span className="form-eyebrow">บริการแจ้งซ่อมออนไลน์</span>
            <h2>ฟอร์มแจ้งปัญหา / งานซ่อมบำรุง</h2>
            <p>กรอกข้อมูลให้ครบเพื่อช่วยให้เจ้าหน้าที่ตรวจสอบและดำเนินการได้รวดเร็วขึ้น</p>
            <p className="public-repair-notice">
              ข้อมูลรายการแจ้งซ่อม เช่น สถานที่ อุปกรณ์ และรายละเอียดปัญหา อาจแสดงบนหน้ารายการสาธารณะ
              กรุณาอย่าระบุข้อมูลส่วนบุคคลหรือข้อมูลลับในรายละเอียด
            </p>
          </div>

          <form onSubmit={handleSubmit} noValidate>
            <div className="form-group">
              <label htmlFor="reporter-email">อีเมลผู้แจ้ง (ที่ต้องการรับข้อมูลการแจ้งซ่อม)<span className="required">*</span></label>
              <input
                id="reporter-email"
                type="email"
                placeholder="กรอกอีเมลที่ติดต่อได้"
                value={email}
                onChange={(e) => { setEmail(e.target.value); setValidationErrors((current) => ({ ...current, email: undefined })); }}
                aria-invalid={Boolean(validationErrors.email)}
              />
              {validationErrors.email && <p className="field-error">{validationErrors.email}</p>}
            </div>

            <div className="form-group">
              <label htmlFor="reporter-type">สถานะผู้แจ้ง <span className="required">*</span></label>
              <select
                id="reporter-type"
                value={reporterType}
                onChange={(event) => {
                  setReporterType(event.target.value);
                  setDegreeLevel("");
                  setDepartment("");
                  setOtherReporterDepartment("");
                  setValidationErrors((current) => ({
                    ...current,
                    reporterType: undefined,
                    degreeLevel: undefined,
                    department: undefined,
                    otherReporterDepartment: undefined,
                  }));
                }}
                aria-invalid={Boolean(validationErrors.reporterType)}
              >
                <option value="">-- เลือกสถานะผู้แจ้ง --</option>
                {reporterTypes.map((type) => <option key={type} value={type}>{type}</option>)}
              </select>
              {validationErrors.reporterType && <p className="field-error">{validationErrors.reporterType}</p>}
            </div>

            {reporterType === "นักศึกษา" && (
              <>
                <div className="form-group">
                  <label htmlFor="reporter-degree">ระดับการศึกษา <span className="required">*</span></label>
                  <select
                    id="reporter-degree"
                    value={degreeLevel}
                    onChange={(event) => {
                      setDegreeLevel(event.target.value);
                      setValidationErrors((current) => ({ ...current, degreeLevel: undefined, department: undefined }));
                    }}
                    aria-invalid={Boolean(validationErrors.degreeLevel)}
                  >
                    <option value="">-- กรุณาเลือกระดับการศึกษา --</option>
                    {degreeLevels.map((degree) => (
                      <option key={degree.id} value={degree.id}>{degree.name}</option>
                    ))}
                  </select>
                  {validationErrors.degreeLevel && <p className="field-error">{validationErrors.degreeLevel}</p>}
                </div>

                {degreeLevel && (
                  <div className="form-group">
                    <label htmlFor="student-department">ภาควิชาสังกัด <span className="required">*</span></label>
                    <select
                      id="student-department"
                      value={department}
                      onChange={(event) => {
                        setDepartment(event.target.value);
                        setValidationErrors((current) => ({ ...current, department: undefined }));
                      }}
                      aria-invalid={Boolean(validationErrors.department)}
                    >
                      <option value="">-- กรุณาเลือกภาควิชาสังกัด --</option>
                      {departments.map((item) => (
                        <option key={item.id} value={item.id}>{item.name}</option>
                      ))}
                    </select>
                    {validationErrors.department && <p className="field-error">{validationErrors.department}</p>}
                  </div>
                )}
              </>
            )}

            {internalReporterTypes.includes(reporterType) && reporterType !== "นักศึกษา" && (
              <div className="form-group slide-down">
                <label htmlFor="staff-department">ภาควิชาสังกัด <span className="required">*</span></label>
                <select
                  id="staff-department"
                  value={department}
                  onChange={(event) => {
                    setDepartment(event.target.value);
                    setValidationErrors((current) => ({ ...current, department: undefined }));
                  }}
                  aria-invalid={Boolean(validationErrors.department)}
                >
                  <option value="">-- เลือกภาควิชาสังกัด --</option>
                  {departments.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name}{item.degree_level_name ? ` (${item.degree_level_name})` : ""}
                    </option>
                  ))}
                </select>
                {validationErrors.department && <p className="field-error">{validationErrors.department}</p>}
              </div>
            )}

            {reporterType === "บุคลากรภายนอก" && (
              <div className="form-group">
                <label htmlFor="other-reporter-department">หน่วยงานหรือความเกี่ยวข้อง <span className="required">*</span></label>
                <input
                  id="other-reporter-department"
                  type="text"
                  placeholder="เช่น ผู้มาติดต่อ หรือหน่วยงานภายนอก"
                  value={otherReporterDepartment}
                  onChange={(event) => {
                    setOtherReporterDepartment(event.target.value);
                    setValidationErrors((current) => ({ ...current, otherReporterDepartment: undefined }));
                  }}
                  aria-invalid={Boolean(validationErrors.otherReporterDepartment)}
                />
                {validationErrors.otherReporterDepartment && <p className="field-error">{validationErrors.otherReporterDepartment}</p>}
              </div>
            )}

            <div className="form-group">
              <label>อาคาร / สถานที่ <span className="required">*</span></label>
              <select value={location} onChange={(e) => { setLocation(e.target.value); setValidationErrors((current) => ({ ...current, location: undefined })); }} aria-invalid={Boolean(validationErrors.location)}>
                <option value="">-- กรุณาเลือกอาคาร --</option>
                {locations.map((loc) => (
                  <option key={loc.id} value={loc.id}>{loc.name}</option>
                ))}
              </select>
              {validationErrors.location && <p className="field-error">{validationErrors.location}</p>}
            </div>

            {location && (
              isOtherLocation() ? (
                <div className="form-group slide-down">
                  <label>ระบุสถานที่/ อาคารเพิ่มเติม <span className="required">*</span></label>
                  <input
                    type="text"
                    placeholder="เช่น โรงอาหารกลาง, หน้าคณะวิทยาศาสตร์, ทางเดินเชื่อมตึก"
                    value={customLocationName}
                    onChange={(e) => { setCustomLocationName(e.target.value); setValidationErrors((current) => ({ ...current, customLocation: undefined })); }}
                    aria-invalid={Boolean(validationErrors.customLocation)}
                  />
                  {validationErrors.customLocation && <p className="field-error">{validationErrors.customLocation}</p>}
                </div>
              ) : (
                <>
                  {!isFloorDataLoaded ? (
                    <div className="form-group slide-down">
                      <label>ชั้น</label>
                      <p role="status">กำลังโหลดข้อมูลชั้น...</p>
                    </div>
                  ) : floorLoadError ? (
                    <p className="field-error" role="alert">ไม่สามารถโหลดข้อมูลชั้นได้ กรุณาลองใหม่</p>
                  ) : floors.length > 0 ? (
                    <>
                      <div className="form-group slide-down">
                        <label>ชั้น <span className="required">*</span></label>
                        <select value={floor} onChange={(e) => { setFloor(e.target.value); setValidationErrors((current) => ({ ...current, floor: undefined })); }} aria-invalid={Boolean(validationErrors.floor)}>
                          <option value="">-- กรุณาเลือกชั้น --</option>
                          {floors.map((f) => (
                            <option key={f.id} value={f.id}>{f.floor_name}</option>
                          ))}
                        </select>
                        {validationErrors.floor && <p className="field-error">{validationErrors.floor}</p>}
                      </div>

                      {floor && (
                        <div className="form-group slide-down">
                          <label>ห้อง / จุดเกิดเหตุ <span className="required">*</span></label>
                          <select value={room} onChange={(e) => { setRoom(e.target.value); setValidationErrors((current) => ({ ...current, room: undefined })); }} aria-invalid={Boolean(validationErrors.room)}>
                            <option value="">-- กรุณาเลือกห้อง --</option>
                            {rooms.map((r) => (
                              <option key={r.id} value={r.id}>{r.room_number}</option>
                            ))}
                          </select>
                          {validationErrors.room && <p className="field-error">{validationErrors.room}</p>}
                        </div>
                      )}

                      {room && isOtherRoom() && (
                        <div className="form-group slide-down">
                          <label>ระบุชื่อห้อง / จุดเกิดเหตุอื่นๆ <span className="required">*</span></label>
                          <input
                            type="text"
                            placeholder="เช่น โถงทางเดินหน้าห้องน้ำชั้น 2, ห้องเก็บของข้างบันได"
                            value={customRoomName}
                            onChange={(e) => { setCustomRoomName(e.target.value); setValidationErrors((current) => ({ ...current, customRoom: undefined })); }}
                            aria-invalid={Boolean(validationErrors.customRoom)}
                          />
                          {validationErrors.customRoom && <p className="field-error">{validationErrors.customRoom}</p>}
                        </div>
                      )}

                      {room && !isOtherRoom() && (
                        <div className="form-group slide-down">
                          <label>อุปกรณ์ที่ชำรุด (ไม่บังคับ)</label>
                          <select value={equipment} onChange={(e) => setEquipment(e.target.value)}>
                            <option value="">-- ไม่ระบุ / ซ่อมโครงสร้างห้อง --</option>
                            {uniqueEquipments.map((eq) => (
                              <option key={eq.id} value={eq.id}>{eq.displayName}</option>
                            ))}
                          </select>
                        </div>
                      )}
                    </>
                  ) : (
                    <div className="form-group slide-down">
                      <label>ห้อง / จุดเกิดเหตุ <span className="required">*</span></label>
                      <input
                        type="text"
                        placeholder="ระบุห้องหรือจุดเกิดเหตุในอาคารนี้"
                        value={customRoomName}
                        onChange={(e) => { setCustomRoomName(e.target.value); setValidationErrors((current) => ({ ...current, customRoom: undefined })); }}
                        aria-invalid={Boolean(validationErrors.customRoom)}
                      />
                      {validationErrors.customRoom && <p className="field-error">{validationErrors.customRoom}</p>}
                    </div>
                  )}

                </>
              )
            )}

            <div className="form-group">
              <label>หมวดหมู่งานซ่อม <span className="required">*</span></label>
              <select value={problemType} onChange={(e) => { setProblemType(e.target.value); setValidationErrors((current) => ({ ...current, problemType: undefined })); }} aria-invalid={Boolean(validationErrors.problemType)}>
                <option value="">-- กรุณาเลือกหมวดหมู่ปัญหา --</option>
                {problemTypes.map((type) => (
                  <option key={type.id} value={type.id}>{type.name}</option>
                ))}
              </select>
              {validationErrors.problemType && <p className="field-error">{validationErrors.problemType}</p>}
            </div>

            {problemType && isOtherProblem() && (
              <div className="form-group slide-down">
                <label>ระบุหมวดหมู่งานซ่อมอื่นๆ <span className="required">*</span></label>
                <input
                  type="text"
                  placeholder="เช่น ซ่อมบานพับประตู, ปรับทิศทางแอร์ ฯลฯ"
                  value={customProblemName}
                  onChange={(e) => { setCustomProblemName(e.target.value); setValidationErrors((current) => ({ ...current, customProblem: undefined })); }}
                  aria-invalid={Boolean(validationErrors.customProblem)}
                />
                {validationErrors.customProblem && <p className="field-error">{validationErrors.customProblem}</p>}
              </div>
            )}

            <div className="form-group">
              <label>อธิบายรายละเอียดของปัญหา <span className="required">*</span></label>
              <textarea
                placeholder="ระบุตำแหน่งให้ชัดเจน เช่น เครื่องปรับอากาศตัวซ้าย, เก้าอี้ปฏิบัติการแถวหน้าสุด ตัวที่ 2"
                rows="4"
                value={details}
                onChange={(e) => { setDetails(e.target.value); setValidationErrors((current) => ({ ...current, details: undefined })); }}
                aria-invalid={Boolean(validationErrors.details)}
              ></textarea>
              {validationErrors.details && <p className="field-error">{validationErrors.details}</p>}
            </div>

            <div className="form-group">
              <label>แนบรูปภาพประกอบ (ถ้ามี)</label>
              <div className="upload-box-container">
                <div className="upload-buttons-row">
                  <button
                    type="button"
                    className="btn-upload-gallery"
                    onClick={() => fileInputRef.current.click()}
                  >
                    <ImageIcon size={20} /> เลือกจากคลังภาพ
                  </button>

                  <button
                    type="button"
                    className="btn-upload-camera"
                    onClick={() => cameraInputRef.current.click()}
                  >
                    <Camera size={20} /> ถ่ายภาพ
                  </button>
                </div>

                {image && (
                  <p className="file-name-preview"><CheckCircle2 size={17} aria-hidden="true" /> ไฟล์ที่เลือก: {image.name}</p>
                )}

                <input
                  type="file"
                  ref={fileInputRef}
                  accept="image/*"
                  onChange={handleImageChange}
                  className="hidden-file-input"
                  style={{ display: 'none' }}
                />
                <input
                  type="file"
                  ref={cameraInputRef}
                  accept="image/*"
                  capture="environment"
                  onChange={handleImageChange}
                  className="hidden-file-input"
                  style={{ display: 'none' }}
                />
              </div>
            </div>

            <div className="form-actions">
              <button
                type="submit"
                className="btn-submit"
                disabled={isSubmitting}
              >
                {isSubmitSuccess
                  ? <><CheckCircle2 size={19} aria-hidden="true" /> ส่งข้อมูลสำเร็จ</>
                  : isSubmitting
                    ? <><LoaderCircle className="spin-icon" size={19} aria-hidden="true" /> กำลังส่งข้อมูล...</>
                    : <><Send size={19} aria-hidden="true" /> ส่งข้อมูลแจ้งซ่อม</>}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

export default CreateRepair;