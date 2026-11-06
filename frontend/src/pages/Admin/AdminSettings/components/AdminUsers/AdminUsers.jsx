import { API_BASE_URL } from "../../../../../api/baseUrl.js";
import React, { useState, useEffect } from "react";
import { CircleX, Eye, Pencil, Plus, Save, ShieldCheck, Trash2, UserPlus, UsersRound, Wrench, X, Zap, Mail, Building2, MapPin, UserRound } from "lucide-react";
import SearchField from "../../../../../components/ui/SearchField";
import FilterBar from "../../../../../components/ui/FilterBar";
import Pagination from "../../../../../components/ui/Pagination";
import { getPageItems } from "../../../../../components/ui/paginationUtils";
import useToast from "../../../../../hooks/useToast";
import "./AdminUsers.css";

const USER_PAGE_SIZE = 4;

function AdminUsers() {
  const { toast, confirm } = useToast();
  const [users, setUsers] = useState([]);
  const [problemTypes, setProblemTypes] = useState([]); 
  const [departments, setDepartments] = useState([]); // 🔥 เพิ่ม State เก็บสาขาวิชา
  const [locations, setLocations] = useState([]);
  const [areaFloors, setAreaFloors] = useState([]);
  const [areaRooms, setAreaRooms] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [roleFilter, setRoleFilter] = useState("");
  const [specialtyFilter, setSpecialtyFilter] = useState("");
  const [currentPage, setCurrentPage] = useState(1);

  // States สำหรับ Modal
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);

  // States สำหรับฟอร์ม
  const [editId, setEditId] = useState(null);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [fullName, setFullName] = useState(""); 
  const [email, setEmail] = useState(""); // 🔥 เพิ่มฟิลด์อีเมล
  const [departmentId, setDepartmentId] = useState(""); // 🔥 เพิ่มฟิลด์สาขา
  const [role, setRole] = useState("technician"); 
  const [isHeadTechnician, setIsHeadTechnician] = useState(false);
  const [specialties, setSpecialties] = useState([]); 
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [updatingHeadTechnicianId, setUpdatingHeadTechnicianId] = useState(null);
  const [areaTechnician, setAreaTechnician] = useState(null);
  const [detailTechnician, setDetailTechnician] = useState(null);
  const [detailAreas, setDetailAreas] = useState([]);
  const [isDetailLoading, setIsDetailLoading] = useState(false);
  const [detailLoadError, setDetailLoadError] = useState("");
  const [responsibleAreas, setResponsibleAreas] = useState([]);
  const [areaScope, setAreaScope] = useState("location");
  const [areaLocationId, setAreaLocationId] = useState("");
  const [areaFloorId, setAreaFloorId] = useState("");
  const [areaRoomId, setAreaRoomId] = useState("");
  const [isAreaLoading, setIsAreaLoading] = useState(false);
  const [isAreaSubmitting, setIsAreaSubmitting] = useState(false);
  const [areaLoadError, setAreaLoadError] = useState("");

  const validateUserForm = (allowEmptyPassword = false) => {
    const trimmedUsername = username.trim();
    const trimmedFullName = fullName.trim();
    const trimmedEmail = email.trim();
    const trimmedPassword = password.trim();

    if (!trimmedUsername || !/^[A-Za-z0-9._-]{3,30}$/.test(trimmedUsername)) {
      return "username ต้องมี 3-30 ตัวอักษร และใช้ได้เฉพาะ a-z, A-Z, 0-9, '.', '_', '-'";
    }
    if (!trimmedFullName || trimmedFullName.length < 2) {
      return "กรุณาระบุชื่อ-นามสกุลจริงอย่างน้อย 2 ตัวอักษร";
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmedEmail)) {
      return "อีเมลไม่ถูกต้อง กรุณากรอกเป็นรูปแบบ email ที่ถูกต้อง";
    }
    if (trimmedPassword !== "") {
      if (trimmedPassword.length < 3 || trimmedPassword.length > 30) {
        return "password ต้องมีความยาว 3-30 ตัวอักษร";
      }
    } else if (!allowEmptyPassword) {
      return "password ต้องมีความยาว 3-30 ตัวอักษร";
    }
    if (role === "technician" && !departmentId) {
      return "ช่างเทคนิคจำเป็นต้องเลือกสาขาวิชา";
    }
    if (role === "technician" && specialties.length === 0) {
      return "ช่างเทคนิคต้องเลือกความถนัดอย่างน้อย 1 หมวดหมู่";
    }
    if (role === "technician" && specialties.length > 3) {
      return "ช่างเทคนิคสามารถเลือกความถนัดได้สูงสุด 3 หมวดหมู่";
    }
    return null;
  };

  // ดึงข้อมูลหลักจาก Backend พร้อมกัน
  const fetchMasterData = async () => {
    try {
      setIsLoading(true);
      const [usersRes, typesRes, deptsRes, locationsRes] = await Promise.all([
        fetch(API_BASE_URL + "/api/users"),
        fetch(API_BASE_URL + "/api/problem-types"),
        fetch(API_BASE_URL + "/api/departments"), // 🔥 ดึงข้อมูลสาขาวิชา
        fetch(API_BASE_URL + "/api/locations"),
      ]);
      
      if (usersRes.ok) setUsers(await usersRes.json() || []);
      if (typesRes.ok) setProblemTypes(await typesRes.json() || []);
      if (deptsRes.ok) setDepartments(await deptsRes.json() || []);
      if (locationsRes.ok) setLocations(await locationsRes.json() || []);
    } catch (error) {
      console.error("Fetch Users Error:", error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchMasterData();
  }, []);

  useEffect(() => {
    if (!areaLocationId) {
      setAreaFloors([]);
      setAreaFloorId("");
      setAreaRooms([]);
      setAreaRoomId("");
      return;
    }

    const fetchAreaFloors = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/api/locations/${areaLocationId}/floors`);
        if (!response.ok) throw new Error("ไม่สามารถดึงข้อมูลชั้นได้");
        setAreaFloors(await response.json() || []);
      } catch (error) {
        console.error("Fetch responsible-area floors error:", error);
        toast.error("โหลดข้อมูลชั้นไม่สำเร็จ", { description: "กรุณาลองใหม่อีกครั้ง" });
      }
    };
    fetchAreaFloors();
  }, [areaLocationId, toast]);

  useEffect(() => {
    if (!areaFloorId) {
      setAreaRooms([]);
      setAreaRoomId("");
      return;
    }

    const fetchAreaRooms = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/api/floors/${areaFloorId}/rooms`);
        if (!response.ok) throw new Error("ไม่สามารถดึงข้อมูลห้องได้");
        setAreaRooms(await response.json() || []);
      } catch (error) {
        console.error("Fetch responsible-area rooms error:", error);
        toast.error("โหลดข้อมูลห้องไม่สำเร็จ", { description: "กรุณาลองใหม่อีกครั้ง" });
      }
    };
    fetchAreaRooms();
  }, [areaFloorId, toast]);

  const technicianUsers = users.filter((user) => user.role === "technician");
  const specialtyCounts = new Map(
    problemTypes.map((type) => [
      String(type.id),
      technicianUsers.filter((user) =>
        (user.specialties || []).some((id) => String(id) === String(type.id))
      ).length,
    ])
  );
  const normalizedSearchTerm = searchTerm.trim().toLowerCase();
  const filteredUsers = users.filter((user) => {
    const matchesSearch = [user.username, user.full_name, user.role, user.email]
      .some((value) => String(value || "").toLowerCase().includes(normalizedSearchTerm));
    const matchesRole = !roleFilter || user.role === roleFilter;
    const matchesSpecialty = !specialtyFilter || (
      user.role === "technician" &&
      (user.specialties || []).some((id) => String(id) === specialtyFilter)
    );
    return matchesSearch && matchesRole && matchesSpecialty;
  });
  const filteredTechnicianCount = filteredUsers.filter((user) => user.role === "technician").length;
  const paginatedUsers = getPageItems(filteredUsers, currentPage, USER_PAGE_SIZE);
  const userFilters = [
    {
      id: "role",
      label: "ประเภทบัญชี",
      value: roleFilter,
      options: [
        { value: "admin", label: "แอดมิน" },
        { value: "technician", label: "ช่างเทคนิค" },
      ],
      onChange: (value) => {
        setRoleFilter(value);
        setCurrentPage(1);
      },
    },
    {
      id: "specialty",
      label: "งานที่ถนัด",
      value: specialtyFilter,
      options: problemTypes.map((type) => ({
        value: String(type.id),
        label: `${type.name} (${specialtyCounts.get(String(type.id)) || 0} คน)`,
      })),
      onChange: (value) => {
        setSpecialtyFilter(value);
        setCurrentPage(1);
      },
    },
  ];

  // -----------------------------------------
  // ฟังก์ชัน Action และการทำงานต่างๆ
  // -----------------------------------------
  const resetForm = () => {
    setEditId(null);
    setUsername("");
    setPassword("");
    setFullName("");
    setEmail("");
    setDepartmentId("");
    setRole("technician");
    setIsHeadTechnician(false);
    setSpecialties([]);
    setIsCreateModalOpen(false);
    setIsEditModalOpen(false);
  };

  const openCreateModal = () => {
    resetForm();
    setIsCreateModalOpen(true);
  };

  const openEditModal = (user) => {
    resetForm();
    setEditId(user.id);
    setUsername(user.username);
    setFullName(user.full_name || "");
    setEmail(user.email || "");
    setDepartmentId(user.department_id || "");
    setRole(user.role);
    setIsHeadTechnician(user.is_head_technician || false);
    setSpecialties(user.specialties || []);
    setIsEditModalOpen(true);
  };

  const handleSpecialtyChange = (typeId) => {
    if (specialties.includes(typeId)) {
      setSpecialties(specialties.filter((id) => id !== typeId));
    } else {
      if (specialties.length >= 3) {
        toast.warning("เลือกความถนัดได้สูงสุด 3 รายการ", { description: "ยกเลิกตัวเลือกเดิมก่อนเลือกหมวดหมู่เพิ่ม" });
        return;
      }
      setSpecialties([...specialties, typeId]);
    }
  };

  const handleCreateSubmit = async (e) => {
    e.preventDefault();
    const validationMessage = validateUserForm();
    if (validationMessage) {
      toast.warning("ข้อมูลไม่ถูกต้อง", { description: validationMessage });
      return;
    }
    setIsSubmitting(true);
    try {
      const response = await fetch(API_BASE_URL + "/api/users", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ 
          username, 
          password, 
          full_name: fullName, 
          email,
          department_id: departmentId ? parseInt(departmentId) : null,
          is_head_technician: role === "technician" ? isHeadTechnician : false,
          role,
          specialties: role === "technician" ? specialties : []
        })
      });

      if (response.ok) {
        toast.success("สร้างผู้ใช้งานสำเร็จ", { description: "เพิ่มบัญชีเข้าสู่ระบบแล้ว" });
        fetchMasterData();
        resetForm();
      } else {
        const data = await response.json();
        toast.error("สร้างผู้ใช้งานไม่สำเร็จ", { description: data.error || "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch {
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleEditSubmit = async (e) => {
    e.preventDefault();
    const validationMessage = validateUserForm(true);
    if (validationMessage) {
      toast.warning("ข้อมูลไม่ถูกต้อง", { description: validationMessage });
      return;
    }

    setIsSubmitting(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/users/${editId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ 
          username, 
          password, 
          full_name: fullName,
          email,
          role,
          department_id: departmentId ? parseInt(departmentId) : null,
          is_head_technician: role === "technician" ? isHeadTechnician : false,
          specialties: role === "technician" ? specialties : []
        })
      });

      if (response.ok) {
        toast.success("อัปเดตข้อมูลสำเร็จ", { description: "บันทึกข้อมูลผู้ใช้งานแล้ว" });
        fetchMasterData();
        resetForm();
      } else {
        const data = await response.json();
        toast.error("อัปเดตข้อมูลไม่สำเร็จ", { description: data.error || "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch {
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async (id, name) => {
    const confirmDelete = await confirm({
      title: "ลบบัญชีผู้ใช้",
      description: `บัญชี “${name}” จะถูกลบออกจากระบบถาวร และไม่สามารถกู้คืนได้`,
      confirmLabel: "ลบบัญชี",
      variant: "danger",
    });
    if (!confirmDelete) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/users/${id}`, {
        method: "DELETE"
      });

      if (response.ok) {
        toast.success("ลบผู้ใช้งานสำเร็จ", { description: "นำบัญชีออกจากระบบแล้ว" });
        fetchMasterData();
      } else {
        const data = await response.json();
        toast.error("ลบผู้ใช้งานไม่สำเร็จ", { description: data.error || "กรุณาลองใหม่อีกครั้ง" });
      }
    } catch {
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", { description: "กรุณาลองใหม่อีกครั้ง" });
    }
  };

  const handleToggleHeadTechnician = async (user) => {
    const nextValue = !user.is_head_technician;
    setUpdatingHeadTechnicianId(user.id);
    try {
      const response = await fetch(
        `${API_BASE_URL}/api/users/${user.id}/head-technician`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ is_head_technician: nextValue }),
        }
      );

      const responseBody = await response.text();
      let responseData = {};
      try {
        responseData = responseBody ? JSON.parse(responseBody) : {};
      } catch {
        responseData = { error: responseBody };
      }

      if (!response.ok) {
        if (response.status === 404) {
          toast.error("Backend ยังไม่มี API ตั้งหัวหน้าช่าง", {
            description: "กรุณา deploy backend เวอร์ชันล่าสุด แล้วลองใหม่",
          });
          return;
        }
        toast.error("เปลี่ยนสถานะหัวหน้าช่างไม่สำเร็จ", {
          description: responseData.error || "กรุณาลองใหม่อีกครั้ง",
        });
        return;
      }

      setUsers((currentUsers) =>
        currentUsers.map((currentUser) =>
          currentUser.id === user.id
            ? { ...currentUser, is_head_technician: responseData.is_head_technician }
            : currentUser
        )
      );
      toast.success(
        responseData.is_head_technician ? "ตั้งเป็นหัวหน้าช่างแล้ว" : "ยกเลิกสถานะหัวหน้าช่างแล้ว",
        { description: "ป้ายสถานะในรายชื่อผู้ใช้ได้รับการอัปเดตแล้ว" }
      );
    } catch (error) {
      console.error("Update Head Technician Error:", error);
      toast.error("ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์", {
        description: error instanceof TypeError
          ? "เชื่อมต่อ backend ไม่สำเร็จ ตรวจสอบสถานะเซิร์ฟเวอร์หรือ CORS แล้วลองใหม่"
          : "เปลี่ยนสถานะหัวหน้าช่างไม่สำเร็จ กรุณาลองใหม่อีกครั้ง",
      });
    } finally {
      setUpdatingHeadTechnicianId(null);
    }
  };

  const openResponsibleAreas = async (technician) => {
    setAreaTechnician(technician);
    setResponsibleAreas([]);
    setAreaScope("location");
    setAreaLocationId("");
    setAreaFloorId("");
    setAreaRoomId("");
    setAreaLoadError("");
    setIsAreaLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/users/${technician.id}/responsible-areas`);
      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || "ไม่สามารถโหลดพื้นที่รับผิดชอบได้");
      }
      setResponsibleAreas(await response.json() || []);
    } catch (error) {
      console.error("Fetch technician responsible areas error:", error);
      setAreaLoadError(error.message || "กรุณาลองใหม่อีกครั้ง");
      toast.error("โหลดพื้นที่รับผิดชอบไม่สำเร็จ", { description: error.message || "กรุณาลองใหม่อีกครั้ง" });
    } finally {
      setIsAreaLoading(false);
    }
  };

  const openTechnicianDetails = async (technician) => {
    setDetailTechnician(technician);
    setDetailAreas([]);
    setDetailLoadError("");
    setIsDetailLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/users/${technician.id}/responsible-areas`);
      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.error || "ไม่สามารถโหลดพื้นที่รับผิดชอบได้");
      }
      setDetailAreas(data || []);
    } catch (error) {
      console.error("Fetch technician detail areas error:", error);
      setDetailLoadError(error.message || "กรุณาลองใหม่อีกครั้ง");
      toast.error("โหลดรายละเอียดพื้นที่ช่างไม่สำเร็จ", {
        description: error.message || "กรุณาลองใหม่อีกครั้ง",
      });
    } finally {
      setIsDetailLoading(false);
    }
  };

  const addResponsibleArea = () => {
    const newArea = areaScope === "location"
      ? { scope: "location", location_id: Number(areaLocationId) }
      : areaScope === "floor"
        ? { scope: "floor", floor_id: Number(areaFloorId) }
        : { scope: "room", room_id: Number(areaRoomId) };
    const selectedLocation = locations.find((location) => String(location.id) === String(areaLocationId));
    const selectedFloor = areaFloors.find((floor) => String(floor.id) === String(areaFloorId));
    const selectedRoom = areaRooms.find((room) => String(room.id) === String(areaRoomId));
    const hasSelection = areaScope === "location"
      ? Boolean(areaLocationId)
      : areaScope === "floor"
        ? Boolean(areaFloorId)
        : Boolean(areaRoomId);

    if (!hasSelection) {
      toast.warning("กรุณาเลือกพื้นที่", { description: "เลือกอาคาร ชั้น หรือห้องก่อนเพิ่มรายการ" });
      return;
    }

    const duplicate = responsibleAreas.some((area) =>
      area.scope === newArea.scope &&
      String(area.location_id || area.floor_id || area.room_id) ===
        String(newArea.location_id || newArea.floor_id || newArea.room_id)
    );
    if (duplicate) {
      toast.warning("พื้นที่นี้ถูกเพิ่มแล้ว");
      return;
    }

    setResponsibleAreas((currentAreas) => [
      ...currentAreas,
      {
        ...newArea,
        location_name: selectedLocation?.name,
        floor_name: selectedFloor?.floor_name,
        room_number: selectedRoom?.room_number,
      },
    ]);
  };

  const removeResponsibleArea = (indexToRemove) => {
    setResponsibleAreas((currentAreas) => currentAreas.filter((_, index) => index !== indexToRemove));
  };

  const saveResponsibleAreas = async () => {
    if (!areaTechnician) return;
    setIsAreaSubmitting(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/users/${areaTechnician.id}/responsible-areas`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          areas: responsibleAreas.map(({ scope, location_id, floor_id, room_id }) => ({
            scope,
            location_id,
            floor_id,
            room_id,
          })),
        }),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "ไม่สามารถบันทึกพื้นที่รับผิดชอบได้");
      toast.success("บันทึกพื้นที่รับผิดชอบแล้ว");
      setAreaTechnician(null);
    } catch (error) {
      console.error("Save technician responsible areas error:", error);
      toast.error("บันทึกพื้นที่ไม่สำเร็จ", { description: error.message || "กรุณาลองใหม่อีกครั้ง" });
    } finally {
      setIsAreaSubmitting(false);
    }
  };

  return (
    <div className="admin-users-container" style={{ padding: 0 }}>
      <div className="admin-wrapper">
        
        {/* Toolbar: ค้นหา & เพิ่มผู้ใช้ */}
        <div className="users-toolbar">
          <div className="search-section">
            <SearchField
              id="admin-user-search"
              label="ค้นหาบัญชีผู้ใช้"
              placeholder="ชื่อผู้ใช้ ชื่อจริง อีเมล หรือสิทธิ์"
              value={searchTerm}
              onChange={(value) => { setSearchTerm(value); setCurrentPage(1); }}
            />
          </div>
          <div className="toolbar-actions">
            <FilterBar filters={userFilters} />
            <button className="btn-create-user" onClick={openCreateModal}>
              <UserPlus size={18} aria-hidden="true" /> เพิ่มผู้ใช้งานใหม่
            </button>
          </div>
        </div>

        {!isLoading && (
          <p className="users-filter-summary" aria-live="polite">
            พบช่าง {filteredTechnicianCount} คน จากทั้งหมด {technicianUsers.length} คน
          </p>
        )}

        {/* ตารางรายชื่อผู้ใช้งาน */}
        <div className="users-list">
          {isLoading ? (
            <div className="loading-state">กำลังโหลดข้อมูล...</div>
          ) : filteredUsers.length > 0 ? (
            paginatedUsers.map((user) => (
              <div className="user-item-card" key={user.id}>
                <div className="user-info">
                  <div className="user-avatar">
                    {user.role === "admin" ? <ShieldCheck size={22} aria-hidden="true" /> : <Wrench size={22} aria-hidden="true" />}
                  </div>
                  <div className="user-details-block">
                    <h3 className="user-name">
                      {user.full_name} <span className="user-username-label">(@{user.username})</span>
                    </h3>
                    
                    <div className="user-meta-info" style={{ display: 'flex', gap: '10px', flexWrap: 'wrap', margin: '6px 0', fontSize: '0.85rem', color: '#64748b' }}>
                      {user.email && <span style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}><Mail size={14} /> {user.email}</span>}
                      {user.department_name && <span style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}><Building2 size={14} /> สาขา: {user.department_name}</span>}
                    </div>
                    
                    <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap', alignItems: 'center' }}>
                      <span className={`role-badge ${user.role === "admin" ? "role-admin" : "role-tech"}`}>
                        {user.role === "admin" ? "แอดมิน (Admin)" : "ช่างเทคนิค (Technician)"}
                      </span>

                      {user.role === "technician" && user.is_head_technician && (
                        <span className="role-badge" style={{ backgroundColor: '#fef3c7', color: '#b45309', border: '1px solid #fde68a' }}>
                          <Wrench size={13} style={{ marginRight: '4px' }} /> หัวหน้าช่าง
                        </span>
                      )}
                    </div>

                    {user.role === "technician" && user.specialty_names && user.specialty_names.length > 0 && (
                      <div className="specialty-container" style={{ marginTop: '8px' }}>
                        {user.specialty_names.map((name, index) => (
                          <span className="specialty-tag" key={index}><Zap size={13} aria-hidden="true" /> {name}</span>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
                
                <div className="user-actions">
                  {user.role === "technician" && (
                    <>
                      <button
                        className="btn-user-details"
                        type="button"
                        onClick={() => openTechnicianDetails(user)}
                      >
                        <Eye size={16} aria-hidden="true" /> ดูรายละเอียด
                      </button>
                      <button
                        className="btn-head-tech"
                        type="button"
                        onClick={() => openResponsibleAreas(user)}
                      >
                        <MapPin size={16} aria-hidden="true" /> กำหนดพื้นที่
                      </button>
                      <button
                        className={`btn-head-tech${user.is_head_technician ? " is-active" : ""}`}
                        type="button"
                        onClick={() => handleToggleHeadTechnician(user)}
                        disabled={updatingHeadTechnicianId === user.id}
                        aria-pressed={Boolean(user.is_head_technician)}
                      >
                        <Wrench size={16} aria-hidden="true" />
                        {updatingHeadTechnicianId === user.id
                          ? "กำลังบันทึก..."
                          : user.is_head_technician
                            ? "ยกเลิกหัวหน้าช่าง"
                            : "ตั้งเป็นหัวหน้าช่าง"}
                      </button>
                    </>
                  )}
                  <button className="btn-edit" onClick={() => openEditModal(user)}>
                    <Pencil size={16} aria-hidden="true" /> แก้ไข
                  </button>
                  <button className="btn-delete" onClick={() => handleDelete(user.id, user.full_name)}>
                    <Trash2 size={16} aria-hidden="true" /> ลบ
                  </button>
                </div>
              </div>
            ))
          ) : (
            <div className="no-results">
              <p><CircleX size={18} aria-hidden="true" /> ไม่พบข้อมูลบัญชีผู้ใช้งาน</p>
            </div>
          )}
        </div>
        <Pagination
          currentPage={currentPage}
          totalItems={filteredUsers.length}
          pageSize={USER_PAGE_SIZE}
          onPageChange={setCurrentPage}
        />
      </div>

      {/* ==========================================
          POPUP MODAL: สร้างผู้ใช้งานใหม่
          ========================================== */}
      {isCreateModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <button className="close-btn" type="button" onClick={resetForm} aria-label="ปิดหน้าต่าง"><X aria-hidden="true" /></button>
            <h3 className="modal-title"><Plus size={20} aria-hidden="true" /> สร้างบัญชีผู้ใช้ใหม่</h3>
            
            <form onSubmit={handleCreateSubmit}>
              <div className="input-group">
                <label>สิทธิ์การใช้งาน <span className="required">*</span></label>
                <select value={role} onChange={(e) => setRole(e.target.value)} required>
                  <option value="technician">ช่างเทคนิค (Technician)</option>
                  <option value="admin">ผู้ดูแลระบบ (Admin)</option>
                </select>
              </div>

              <div className="input-group">
                <label>ชื่อ-นามสกุลจริง <span className="required">*</span></label>
                <input 
                  type="text" 
                  placeholder="ระบุชื่อและนามสกุล (เช่น นายสมชาย ยอดช่าง)"
                  value={fullName} 
                  onChange={(e) => setFullName(e.target.value)} 
                  required 
                />
              </div>

              {/* 🔥 เพิ่มฟิลด์อีเมล */}
              <div className="input-group">
                <label>อีเมลติดต่อ <span className="required">*</span></label>
                <input 
                  type="email" 
                  placeholder="เช่น @silpakorn.edu หรือ @gmail.com"
                  value={email} 
                  onChange={(e) => setEmail(e.target.value)} 
                  required 
                />
              </div>

              {/* 🔥 เพิ่มฟิลด์สาขา */}
              <div className="input-group">
                <label>สาขาวิชา / สังกัด {role === "technician" && <span className="required">*</span>}</label>
                <select value={departmentId} onChange={(e) => setDepartmentId(e.target.value)} required={role === "technician"}>
                  <option value="">{role === "technician" ? "-- โปรดเลือกสาขาวิชา --" : "-- ไม่บังคับสำหรับแอดมิน --"}</option>
                  {departments.map(dept => (
                    <option key={dept.id} value={dept.id}>{dept.name}</option>
                  ))}
                </select>
              </div>

              {role === "technician" && (
                <>
                  <div className="input-group" style={{ backgroundColor: '#f8fafc', padding: '10px 15px', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
                    <label style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer', margin: 0 }}>
                      <input 
                        type="checkbox" 
                        checked={isHeadTechnician}
                        onChange={(e) => setIsHeadTechnician(e.target.checked)}
                        style={{ width: '18px', height: '18px', cursor: 'pointer' }}
                      />
                      <strong>ตั้งเป็นหัวหน้าช่าง (เพิ่มปุ่มมอบหมายงานให้ช่างในสาขา)</strong>
                    </label>
                  </div>

                  <div className="input-group">
                    <label>หมวดหมู่งานซ่อมที่ถนัด (เลือกได้สูงสุด 3 หมวดหมู่) <span className="required">*</span></label>
                    <div className="specialties-checkbox-grid">
                      {problemTypes.map((type) => (
                        <label className="checkbox-item-label" key={type.id}>
                          <input 
                            type="checkbox"
                            checked={specialties.includes(type.id)}
                            onChange={() => handleSpecialtyChange(type.id)}
                          />
                          {type.name}
                        </label>
                      ))}
                    </div>
                  </div>
                </>
              )}

              <div className="input-group">
                <label>ชื่อผู้ใช้งาน (Username สำหรับใช้ Login) <span className="required">*</span></label>
                <input 
                  type="text" 
                  placeholder="ตั้งชื่อล็อกอินภาษาอังกฤษ"
                  value={username} 
                  onChange={(e) => setUsername(e.target.value)} 
                  required 
                />
              </div>

              <div className="input-group">
                <label>รหัสผ่าน (Password) <span className="required">*</span></label>
                <input 
                  type="password" 
                  placeholder="ตั้งรหัสผ่านเบื้องต้น (3-30 ตัว)"
                  value={password} 
                  onChange={(e) => setPassword(e.target.value)} 
                  required 
                />
              </div>

              <div className="modal-actions">
                <button type="button" className="btn-cancel" onClick={resetForm} disabled={isSubmitting}>ยกเลิก</button>
                <button type="submit" className="btn-submit" disabled={isSubmitting}>
                  {isSubmitting ? "กำลังบันทึก..." : <><Save size={18} aria-hidden="true" /> ยืนยันสร้างบัญชี</>}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ==========================================
          POPUP MODAL: แก้ไขผู้ใช้งาน
          ========================================== */}
      {isEditModalOpen && (
        <div className="modal-overlay">
          <div className="modal-box">
            <button className="close-btn" type="button" onClick={resetForm} aria-label="ปิดหน้าต่าง"><X aria-hidden="true" /></button>
            <h3 className="modal-title"><Pencil size={20} aria-hidden="true" /> แก้ไขข้อมูลบัญชี</h3>
            
            <form onSubmit={handleEditSubmit}>
              
              <div className="input-group">
                <label>สิทธิ์การใช้งาน <span className="required">*</span></label>
                <select value={role} onChange={(e) => setRole(e.target.value)} required>
                  <option value="technician">ช่างเทคนิค (Technician)</option>
                  <option value="admin">ผู้ดูแลระบบ (Admin)</option>
                </select>
              </div>

              <div className="input-group">
                <label>ชื่อ-นามสกุลจริง <span className="required">*</span></label>
                <input 
                  type="text" 
                  value={fullName} 
                  onChange={(e) => setFullName(e.target.value)} 
                  required 
                />
              </div>

              <div className="input-group">
                <label>อีเมลติดต่อ <span className="required">*</span></label>
                <input 
                  type="email" 
                  value={email} 
                  onChange={(e) => setEmail(e.target.value)} 
                  required 
                />
              </div>

              <div className="input-group">
                <label>สาขาวิชา / สังกัด {role === "technician" && <span className="required">*</span>}</label>
                <select value={departmentId} onChange={(e) => setDepartmentId(e.target.value)} required={role === "technician"}>
                  <option value="">{role === "technician" ? "-- โปรดเลือกสาขาวิชา --" : "-- ไม่บังคับสำหรับแอดมิน --"}</option>
                  {departments.map(dept => (
                    <option key={dept.id} value={dept.id}>{dept.name}</option>
                  ))}
                </select>
              </div>

              {role === "technician" && (
                <>
                  <div className="input-group" style={{ backgroundColor: '#f8fafc', padding: '10px 15px', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
                    <label style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer', margin: 0 }}>
                      <input 
                        type="checkbox" 
                        checked={isHeadTechnician}
                        onChange={(e) => setIsHeadTechnician(e.target.checked)}
                        style={{ width: '18px', height: '18px', cursor: 'pointer' }}
                      />
                      <strong>ตั้งเป็นหัวหน้าช่าง (เพิ่มปุ่มมอบหมายงานให้ช่างในสาขา)</strong>
                    </label>
                  </div>

                  <div className="input-group">
                    <label>ปรับปรุงงานซ่อมที่ถนัด (สูงสุด 3 หมวดหมู่) <span className="required">*</span></label>
                    <div className="specialties-checkbox-grid">
                      {problemTypes.map((type) => (
                        <label className="checkbox-item-label" key={type.id}>
                          <input 
                            type="checkbox"
                            checked={specialties.includes(type.id)}
                            onChange={() => handleSpecialtyChange(type.id)}
                          />
                          {type.name}
                        </label>
                      ))}
                    </div>
                  </div>
                </>
              )}

              <div className="input-group">
                <label>ชื่อผู้ใช้งาน (Username) <span className="required">*</span></label>
                <input 
                  type="text" 
                  value={username} 
                  onChange={(e) => setUsername(e.target.value)} 
                  required 
                />
              </div>

              <div className="input-group">
                <label>รีเซ็ตรหัสผ่านใหม่ (เว้นว่างไว้หากไม่ต้องการเปลี่ยน)</label>
                <input 
                  type="password" 
                  placeholder="ระบุรหัสผ่านใหม่หากต้องการเปลี่ยน (3-30 ตัว)"
                  value={password} 
                  onChange={(e) => setPassword(e.target.value)}
                />
                <small className="password-warning-text">
                  * หากปล่อยว่างไว้ ระบบจะคงรหัสผ่านเดิมไว้ และถ้ากรอกใหม่ต้องอยู่ในช่วง 3-30 ตัว
                </small>
              </div>

              <div className="modal-actions">
                <button type="button" className="btn-cancel" onClick={resetForm} disabled={isSubmitting}>ยกเลิก</button>
                <button type="submit" className="btn-submit" disabled={isSubmitting}>
                  {isSubmitting ? "กำลังบันทึก..." : <><Save size={18} aria-hidden="true" /> บันทึกการเปลี่ยนแปลง</>}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {areaTechnician && (
        <div className="modal-overlay">
          <div className="modal-box responsible-area-modal">
            <button
              className="close-btn"
              type="button"
              onClick={() => setAreaTechnician(null)}
              disabled={isAreaSubmitting}
              aria-label="ปิดหน้าต่าง"
            >
              <X aria-hidden="true" />
            </button>
            <h3 className="modal-title"><MapPin size={20} aria-hidden="true" /> พื้นที่รับผิดชอบของช่าง</h3>
            <p className="responsible-area-technician">{areaTechnician.full_name}</p>
            <p className="responsible-area-help">เพิ่มพื้นที่เป็นระดับอาคาร ชั้น หรือห้องได้ งานที่ไม่มีพื้นที่ตรงกับรายการนี้จะไม่แสดงช่างคนนี้ในตัวเลือกมอบหมายงาน</p>

            {isAreaLoading ? (
              <div className="loading-state">กำลังโหลดพื้นที่...</div>
            ) : areaLoadError ? (
              <div className="error-state" role="alert">
                โหลดพื้นที่รับผิดชอบไม่สำเร็จ: {areaLoadError}
              </div>
            ) : (
              <>
                <div className="input-group">
                  <label htmlFor="responsible-area-scope">ระดับพื้นที่</label>
                  <select
                    id="responsible-area-scope"
                    value={areaScope}
                    onChange={(event) => {
                      setAreaScope(event.target.value);
                      setAreaFloorId("");
                      setAreaRoomId("");
                    }}
                  >
                    <option value="location">ทั้งอาคาร</option>
                    <option value="floor">เฉพาะชั้น</option>
                    <option value="room">เฉพาะห้อง</option>
                  </select>
                </div>

                <div className="input-group">
                  <label htmlFor="responsible-area-location">อาคาร</label>
                  <select
                    id="responsible-area-location"
                    value={areaLocationId}
                    onChange={(event) => {
                      setAreaLocationId(event.target.value);
                      setAreaFloorId("");
                      setAreaRoomId("");
                    }}
                  >
                    <option value="">-- เลือกอาคาร --</option>
                    {locations.map((location) => (
                      <option key={location.id} value={location.id}>{location.name}</option>
                    ))}
                  </select>
                </div>

                {areaScope !== "location" && (
                  <div className="input-group">
                    <label htmlFor="responsible-area-floor">ชั้น</label>
                    <select
                      id="responsible-area-floor"
                      value={areaFloorId}
                      disabled={!areaLocationId}
                      onChange={(event) => {
                        setAreaFloorId(event.target.value);
                        setAreaRoomId("");
                      }}
                    >
                      <option value="">-- เลือกชั้น --</option>
                      {areaFloors.map((floor) => (
                        <option key={floor.id} value={floor.id}>{floor.floor_name}</option>
                      ))}
                    </select>
                  </div>
                )}

                {areaScope === "room" && (
                  <div className="input-group">
                    <label htmlFor="responsible-area-room">ห้อง</label>
                    <select
                      id="responsible-area-room"
                      value={areaRoomId}
                      disabled={!areaFloorId}
                      onChange={(event) => setAreaRoomId(event.target.value)}
                    >
                      <option value="">-- เลือกห้อง --</option>
                      {areaRooms.map((room) => (
                        <option key={room.id} value={room.id}>{room.room_number}</option>
                      ))}
                    </select>
                  </div>
                )}

                <button className="btn-add-area" type="button" onClick={addResponsibleArea}>
                  <Plus size={16} aria-hidden="true" /> เพิ่มพื้นที่
                </button>

                <div className="responsible-area-list">
                  <h4>พื้นที่ที่เลือก ({responsibleAreas.length})</h4>
                  {responsibleAreas.length === 0 ? (
                    <p className="responsible-area-empty">ยังไม่ได้กำหนดพื้นที่ ช่างจะยังไม่ปรากฏในรายชื่อมอบหมายงาน</p>
                  ) : responsibleAreas.map((area, index) => (
                    <div className="responsible-area-item" key={`${area.scope}-${area.location_id || area.floor_id || area.room_id}`}>
                      <span>
                        {area.scope === "location"
                          ? `ทั้งอาคาร: ${area.location_name || locations.find((location) => String(location.id) === String(area.location_id))?.name || ""}`
                          : area.scope === "floor"
                            ? `ชั้น: ${area.location_name ? `${area.location_name} / ` : ""}${area.floor_name || areaFloors.find((floor) => String(floor.id) === String(area.floor_id))?.floor_name || `ID ${area.floor_id}`}`
                            : `ห้อง: ${area.location_name ? `${area.location_name} / ` : ""}${area.floor_name ? `${area.floor_name} / ` : ""}${area.room_number || `ID ${area.room_id}`}`}
                      </span>
                      <button type="button" onClick={() => removeResponsibleArea(index)} aria-label="นำพื้นที่ออก">
                        <X size={15} aria-hidden="true" />
                      </button>
                    </div>
                  ))}
                </div>

                <div className="modal-actions">
                  <button type="button" className="btn-cancel" onClick={() => setAreaTechnician(null)} disabled={isAreaSubmitting}>ยกเลิก</button>
                  <button type="button" className="btn-submit" onClick={saveResponsibleAreas} disabled={isAreaSubmitting || isAreaLoading}>
                    {isAreaSubmitting ? "กำลังบันทึก..." : <><Save size={18} aria-hidden="true" /> บันทึกพื้นที่</>}
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {detailTechnician && (
        <div className="modal-overlay">
          <div
            className="modal-box technician-account-detail-modal"
            role="dialog"
            aria-modal="true"
            aria-labelledby="technician-account-detail-title"
          >
            <button
              className="close-btn"
              type="button"
              onClick={() => setDetailTechnician(null)}
              aria-label="ปิดรายละเอียดช่าง"
            >
              <X aria-hidden="true" />
            </button>
            <h3 className="modal-title" id="technician-account-detail-title">
              <UserRound size={21} aria-hidden="true" /> รายละเอียดช่าง
            </h3>
            <div className="technician-account-detail-summary">
              <strong>{detailTechnician.full_name}</strong>
              <span>@{detailTechnician.username}</span>
            </div>
            <div className="technician-account-detail-grid">
              <div>
                <strong>อีเมล</strong>
                <span>{detailTechnician.email || "ไม่ได้ระบุ"}</span>
              </div>
              <div>
                <strong>สาขา / สังกัด</strong>
                <span>{detailTechnician.department_name || "ไม่ได้ระบุ"}</span>
              </div>
              <div>
                <strong>สถานะบัญชี</strong>
                <span>{detailTechnician.is_active === false ? "ปิดใช้งาน" : "ใช้งานอยู่"}</span>
              </div>
              <div>
                <strong>ตำแหน่ง</strong>
                <span>{detailTechnician.is_head_technician ? "หัวหน้าช่าง" : "ช่างเทคนิค"}</span>
              </div>
            </div>
            <section className="technician-account-detail-section">
              <h4><Zap size={16} aria-hidden="true" /> ความเชี่ยวชาญ</h4>
              {detailTechnician.specialty_names?.length ? (
                <div className="technician-detail-tags">
                  {detailTechnician.specialty_names.map((name) => (
                    <span key={name}>{name}</span>
                  ))}
                </div>
              ) : (
                <p>ยังไม่ได้ระบุความเชี่ยวชาญ</p>
              )}
            </section>
            <section className="technician-account-detail-section">
              <h4><MapPin size={16} aria-hidden="true" /> พื้นที่รับผิดชอบ</h4>
              {isDetailLoading ? (
                <p>กำลังโหลดพื้นที่รับผิดชอบ...</p>
              ) : detailLoadError ? (
                <p className="technician-detail-error" role="alert">{detailLoadError}</p>
              ) : detailAreas.length ? (
                <ul className="technician-account-area-list">
                  {detailAreas.map((area) => {
                    const locationName = area.location_name || `อาคาร ID ${area.location_id || ""}`;
                    const label = area.scope === "location"
                      ? `ทั้งอาคาร: ${locationName}`
                      : area.scope === "floor"
                        ? `ชั้น: ${locationName} / ${area.floor_name || `ID ${area.floor_id}`}`
                        : `ห้อง: ${locationName} / ${area.floor_name || `ชั้น ID ${area.floor_id || ""}`} / ${area.room_number || `ID ${area.room_id}`}`;
                    return (
                      <li key={`${area.scope}-${area.location_id || area.floor_id || area.room_id}`}>
                        {label}
                      </li>
                    );
                  })}
                </ul>
              ) : (
                <p>ยังไม่ได้กำหนดพื้นที่รับผิดชอบ</p>
              )}
            </section>
          </div>
        </div>
      )}

    </div>
  );
}

export default AdminUsers;