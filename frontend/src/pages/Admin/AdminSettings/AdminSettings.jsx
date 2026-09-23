import React, { useState } from "react";
import { Settings, UsersRound, MapPin, Wrench, Monitor } from "lucide-react"; // เพิ่มไอคอน Monitor
import BackButton from "../../../components/ui/BackButton"; // เช็ค path ให้ตรงกับโปรเจกต์คุณด้วยนะครับ

import AdminUsers from "./components/AdminUsers/AdminUsers";
import AdminLocations from "./components/AdminLocations/AdminLocations"; 
import AdminProblemTypes from "./components/AdminProblemTypes/AdminProblemTypes"; 
import AdminEquipments from "./components/AdminEquipments/AdminEquipments"; // เพิ่ม Import ไฟล์หน้าอุปกรณ์

import "./AdminSettings.css";

function AdminSettings() {
  // สร้าง State สำหรับเก็บว่าตอนนี้เลือก Tab ไหนอยู่ (ค่าเริ่มต้นคือหน้า users)
  const [activeTab, setActiveTab] = useState("users");

  return (
    <div className="admin-settings-container">
      <div className="admin-settings-wrapper">
        <nav className="page-navigation" aria-label="การนำทางย้อนกลับ">
          <BackButton to="/admin/home" label="กลับหน้าแดชบอร์ด" />
        </nav>

        {/* Header ของหน้าการตั้งค่า */}
        <div className="settings-header">
          <div>
            <h2><Settings size={26} aria-hidden="true" /> ตั้งค่าระบบ (System Settings)</h2>
            <p>จัดการข้อมูลพื้นฐาน ผู้ใช้งาน สถานที่ อุปกรณ์ และหมวดหมู่งานซ่อม</p>
          </div>
        </div>

        {/* เมนู Tab สลับหน้า */}
        <div className="settings-tabs">
          <button 
            className={`tab-button ${activeTab === "users" ? "active" : ""}`}
            onClick={() => setActiveTab("users")}
          >
            <UsersRound size={18} /> จัดการผู้ใช้งาน
          </button>
          
          <button 
            className={`tab-button ${activeTab === "locations" ? "active" : ""}`}
            onClick={() => setActiveTab("locations")}
          >
            <MapPin size={18} /> จัดการสถานที่
          </button>

          {/* แท็บใหม่: จัดการอุปกรณ์ */}
          <button 
            className={`tab-button ${activeTab === "equipments" ? "active" : ""}`}
            onClick={() => setActiveTab("equipments")}
          >
            <Monitor size={18} /> จัดการอุปกรณ์
          </button>
          
          <button 
            className={`tab-button ${activeTab === "problems" ? "active" : ""}`}
            onClick={() => setActiveTab("problems")}
          >
            <Wrench size={18} /> หมวดหมู่งานซ่อม
          </button>
        </div>

        {/* ส่วนแสดงผลเนื้อหา (Content) ตาม Tab ที่เลือก */}
        <div className="tab-content">
          {activeTab === "users" && <AdminUsers />}
          
          {activeTab === "locations" && <AdminLocations />}

          {activeTab === "equipments" && <AdminEquipments />}
          
          {activeTab === "problems" && <AdminProblemTypes />}
        </div>
      </div>
    </div>
  );
}

export default AdminSettings;