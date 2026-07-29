import React, { useState, useEffect, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import {
  PieChart, Pie, Cell, Tooltip as RechartsTooltip, ResponsiveContainer,
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Legend
} from "recharts";
import { AlertTriangle, Ban, Calendar } from 'lucide-react';
import dayjs from 'dayjs';
import "./AdminDashboard.css";

function AdminDashboard() {
  const navigate = useNavigate();
  const [repairs, setRepairs] = useState([]);
  const [isLoading, setIsLoading] = useState(true);

  // State สำหรับตัวกรองเวลาและ Tab
  const [selectedMonth, setSelectedMonth] = useState("all");
  const [selectedYear, setSelectedYear] = useState(dayjs().year().toString());
  const [activeTab, setActiveTab] = useState("status");

  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        const response = await fetch("http://localhost:8080/api/repairs");
        if (response.ok) {
          const data = await response.json();
          setRepairs(data || []);
        }
      } catch (error) {
        console.error("ดึงข้อมูล Dashboard ไม่สำเร็จ:", error);
      } finally {
        setIsLoading(false);
      }
    };
    fetchDashboardData();
  }, []);

  // ==========================================
  // Logic การกรองข้อมูลตาม เดือน/ปี ที่เลือก
  // ==========================================
  const filteredRepairs = useMemo(() => {
    return repairs.filter(r => {
      if (!r.created_at) return true;
      const date = dayjs(r.created_at);
      const matchYear = selectedYear === "all" || date.year().toString() === selectedYear;
      const matchMonth = selectedMonth === "all" || (date.month() + 1).toString() === selectedMonth;
      return matchYear && matchMonth;
    });
  }, [repairs, selectedMonth, selectedYear]);

  const availableYears = useMemo(() => {
    const years = repairs.map(r => r.created_at ? dayjs(r.created_at).year() : dayjs().year());
    return [...new Set(years)].sort((a, b) => b - a);
  }, [repairs]);

  // ==========================================
  // Logic การคำนวณสถิติ 4 สถานะหลัก
  // ==========================================
  const totalRepairs = filteredRepairs.length;
  const pendingCount = filteredRepairs.filter(r => r.status === "รอซ่อม").length;
  const progressCount = filteredRepairs.filter(r => r.status === "กำลังซ่อม").length;
  const cannotRepairCount = filteredRepairs.filter(r => r.status === "ซ่อมไม่ได้").length;
  // เผื่อกรณี Database เก็บคำว่า 'เสร็จสิ้น' หรือ 'เสร็จเรียบร้อย'
  const completedCount = filteredRepairs.filter(r => r.status === "เสร็จเรียบร้อย" || r.status === "เสร็จสิ้น").length;

  // ฟังก์ชันคำนวณหาเวลาซ่อมเฉลี่ย (MTTR)
  const calculateAverageRepairTime = (repairsList) => {
    // 1. กรองเอาเฉพาะงานที่ "เสร็จเรียบร้อย" และมีเวลาบันทึกครบถ้วน
    const completedJobs = repairsList.filter(
      (repair) => (repair.status === "เสร็จเรียบร้อย" || repair.status === "เสร็จสิ้น") && repair.accepted_at && repair.completed_at
    );

    if (completedJobs.length === 0) return 0;

    let totalDiffMs = 0;

    // 2. หาผลรวมของเวลาที่ใช้ซ่อมแต่ละงาน (มิลลิวินาที)
    completedJobs.forEach((repair) => {
      const startTime = new Date(repair.accepted_at);
      const endTime = new Date(repair.completed_at);

      // ป้องกันกรณีเวลาติดลบหรือ Error
      if (!isNaN(startTime) && !isNaN(endTime) && endTime >= startTime) {
        totalDiffMs += (endTime - startTime);
      }
    });

    // 3. แปลงจากมิลลิวินาที เป็น "ชั่วโมง"
    const totalHours = totalDiffMs / (1000 * 60 * 60);

    // 4. หาค่าเฉลี่ย
    const averageHours = totalHours / completedJobs.length;

    // คืนค่า ทศนิยม 2 ตำแหน่ง เพื่อให้เห็นแม้เศษนาที (เช่น 0.05 ชั่วโมง)
    return averageHours.toFixed(2);
  };

  // ✅ [เพิ่มใหม่] ฟังก์ชันแปลงเวลาทศนิยม (ชั่วโมง) ให้อ่านง่าย
  const formatRepairTime = (hoursString) => {
    const hours = parseFloat(hoursString);
    if (!hours || isNaN(hours)) return "0 นาที";
  
    const totalMinutes = Math.round(hours * 60);
  
    if (totalMinutes < 60) {
      return `${totalMinutes} นาที`;
    } 
    
    const h = Math.floor(totalMinutes / 60);
    const m = totalMinutes % 60;
    
    if (m === 0) {
      return `${h} ชม.`; 
    }
    
    return `${h} ชม. ${m} นาที`;
  };

  // ✅ [อัปเดต] เติมตัวแปร filteredRepairs เข้าไปในฟังก์ชันเพื่อแก้ปัญหาหน้าขาว
  const mttrDays = calculateAverageRepairTime(filteredRepairs);

  // 🔥 [อัปเดต] ดึงงานที่เกินคุ้มทุนโดยเช็กจาก admin_note ที่ Backend ส่งมาให้
  const warningRepairs = filteredRepairs.filter(r => {
    return r.status === "รอซ่อม" && r.admin_note && r.admin_note !== "";
  });

  // ==========================================
  // ข้อมูลสำหรับกราฟแต่ละแบบ
  // ==========================================
  const statusChartData = [
    { name: "รอซ่อม", value: pendingCount, color: "#f39c12" },
    { name: "กำลังซ่อม", value: progressCount, color: "#2980b9" },
    { name: "ซ่อมไม่ได้", value: cannotRepairCount, color: "#e74c3c" },
    { name: "เสร็จสิ้น", value: completedCount, color: "#007A53" }
  ].filter(item => item.value > 0);

  const problemTypeData = useMemo(() => {
    const counts = {};
    filteredRepairs.forEach(r => { counts[r.problem_type || "ไม่ระบุ"] = (counts[r.problem_type || "ไม่ระบุ"] || 0) + 1; });
    return Object.keys(counts).map(key => ({ name: key, จำนวน: counts[key] })).sort((a, b) => b.จำนวน - a.จำนวน);
  }, [filteredRepairs]);

  const equipmentData = useMemo(() => {
    const counts = {};
    filteredRepairs.forEach(r => { counts[r.equipment_name || "ไม่ระบุ"] = (counts[r.equipment_name || "ไม่ระบุ"] || 0) + 1; });
    return Object.keys(counts).map(key => ({ name: key, จำนวน: counts[key] })).sort((a, b) => b.จำนวน - a.จำนวน).slice(0, 10);
  }, [filteredRepairs]);

  // ==========================================
  // โครงสร้างเมนู Tabs
  // ==========================================
  const TABS_CONFIG = [
    { id: "status", label: "⭕ สัดส่วนสถานะ" },
    { id: "problem", label: "📈 ปัญหาที่พบบ่อย" },
    { id: "equipment", label: "💻 อุปกรณ์ที่ซ่อมบ่อย" },
    { id: "warning", label: `⚠️ งานรอพิจารณา (${warningRepairs.length})` }
  ];

  return (
    <div className="admin-dashboard-container">
      <header className="dashboard-header">
        <div>
          <h1>📊 แผงควบคุมผู้ดูแลระบบ</h1>
          <p>สรุปภาพรวมระบบแจ้งซ่อมบำรุง</p>
        </div>
        <div className="header-actions">
          <button className="btn-users" onClick={() => navigate("/admin/users")}>👥 จัดการบัญชี</button>
          <button className="btn-manage" onClick={() => navigate("/admin/manage")}>🛠️ มอบหมายงาน</button>
          <button className="btn-logout" onClick={() => { if (window.confirm("ออก?")) { localStorage.removeItem("user"); navigate("/"); } }}>🚪 ออกจากระบบ</button>
        </div>
      </header>

      {/* ส่วนตัวกรอง เดือน/ปี */}
      <div className="filter-section">
        <div className="filter-group">
          <Calendar size={20} color="#007A53" />
          <span className="filter-label">ดูข้อมูลประจำ:</span>
          <select value={selectedMonth} onChange={(e) => setSelectedMonth(e.target.value)} className="filter-select">
            <option value="all">รวมทุกเดือน</option>
            {[...Array(12)].map((_, i) => (
              <option key={i + 1} value={i + 1}>เดือน {i + 1}</option>
            ))}
          </select>
          <select value={selectedYear} onChange={(e) => setSelectedYear(e.target.value)} className="filter-select">
            <option value="all">รวมทุกปี</option>
            {availableYears.map(year => (
              <option key={year} value={year}>ปี {year}</option>
            ))}
          </select>
        </div>
      </div>

      {isLoading ? (
        <div className="loading-state">⏳ กำลังโหลดข้อมูล...</div>
      ) : (
        <>
          {/* การ์ดสรุปตัวเลข */}
          <div className="summary-cards-grid">
            <div className="summary-card total"><h3>รวมทั้งหมด</h3><div className="number">{totalRepairs}</div><span>รายการ</span></div>
            <div className="summary-card pending"><h3>รอซ่อม/พิจารณา</h3><div className="number">{pendingCount}</div><span>รายการ</span></div>
            <div className="summary-card progress"><h3>กำลังซ่อม</h3><div className="number">{progressCount}</div><span>ดำเนินการ</span></div>
            
            {/* ✅ [อัปเดต] เรียกใช้ formatRepairTime ตรงนี้ และเอา <span>ชั่วโมง</span> ออก */}
            <div className="summary-card mttr">
              <h3>ระยะเวลาซ่อมเฉลี่ย</h3>
              <div className="number">{formatRepairTime(mttrDays)}</div>
            </div>
            
            <div className="summary-card warning"><h3>งานรอพิจารณาคุ้มทุน</h3><div className="number">{warningRepairs.length}</div><span>รายการ</span></div>
          </div>

          {/* ปุ่มเลือกสถิติ (Tabs) */}
          <div className="tabs-container">
            {TABS_CONFIG.map(tab => (
              <button
                key={tab.id}
                className={`tab-button ${activeTab === tab.id ? 'active' : ''}`}
                onClick={() => setActiveTab(tab.id)}
              >
                {tab.label}
              </button>
            ))}
          </div>

          {/* พื้นที่แสดงผลกราฟตาม Tab ที่เลือก */}
          <div className="chart-display-area">

            {activeTab === "status" && (
              <div className="chart-card">
                <h3>สัดส่วนสถานะงานซ่อม</h3>
                <div className="chart-wrapper">
                  {statusChartData.length > 0 ? (
                    <ResponsiveContainer width="100%" height={350}>
                      <PieChart>
                        <Pie data={statusChartData} cx="50%" cy="50%" innerRadius={80} outerRadius={120} paddingAngle={5} dataKey="value">
                          {statusChartData.map((entry, index) => <Cell key={`cell-${index}`} fill={entry.color} />)}
                        </Pie>
                        <RechartsTooltip formatter={(value) => [`${value} รายการ`, "จำนวน"]} />
                        <Legend verticalAlign="bottom" height={36} />
                      </PieChart>
                    </ResponsiveContainer>
                  ) : <p className="no-data-text">ไม่มีข้อมูลในช่วงเวลานี้</p>}
                </div>
              </div>
            )}

            {activeTab === "problem" && (
              <div className="chart-card">
                <h3>สถิติประเภทปัญหาที่เกิดบ่อยที่สุด</h3>
                <div className="chart-wrapper">
                  {problemTypeData.length > 0 ? (
                    <ResponsiveContainer width="100%" height={350}>
                      <BarChart data={problemTypeData} margin={{ top: 20, right: 30, left: 0, bottom: 5 }}>
                        <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#eee" />
                        <XAxis dataKey="name" tick={{ fill: '#555' }} />
                        <YAxis allowDecimals={false} tick={{ fill: '#555' }} />
                        <RechartsTooltip cursor={{ fill: '#f4f6f5' }} />
                        <Bar dataKey="จำนวน" fill="#007A53" radius={[4, 4, 0, 0]} />
                      </BarChart>
                    </ResponsiveContainer>
                  ) : <p className="no-data-text">ไม่มีข้อมูลในช่วงเวลานี้</p>}
                </div>
              </div>
            )}

            {activeTab === "equipment" && (
              <div className="chart-card">
                <h3>Top 10 อุปกรณ์ที่ถูกแจ้งซ่อมบ่อยที่สุด</h3>
                <div className="chart-wrapper">
                  {equipmentData.length > 0 ? (
                    <ResponsiveContainer width="100%" height={350}>
                      <BarChart data={equipmentData} layout="vertical" margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
                        <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="#eee" />
                        <XAxis type="number" allowDecimals={false} />
                        <YAxis dataKey="name" type="category" width={120} tick={{ fill: '#555', fontSize: 12 }} />
                        <RechartsTooltip cursor={{ fill: '#f4f6f5' }} />
                        <Bar dataKey="จำนวน" fill="#2980b9" radius={[0, 4, 4, 0]} />
                      </BarChart>
                    </ResponsiveContainer>
                  ) : <p className="no-data-text">ไม่มีข้อมูลในช่วงเวลานี้</p>}
                </div>
              </div>
            )}

            {activeTab === "warning" && (
              <div className="chart-card warning-section">
                <h3 className="warning-title"><AlertTriangle size={20} /> งานรอพิจารณา (เกินจุดคุ้มทุน)</h3>
                <div className="warning-table-wrapper">
                  {warningRepairs.length > 0 ? (
                    <table className="warning-table">
                      <thead>
                        <tr>
                          <th>รหัสงานซ่อม</th>
                          <th>อุปกรณ์ / ปัญหา</th>
                          <th>ราคาประเมิน</th>
                          <th>หมายเหตุจากระบบ (เหตุผลที่ระงับ)</th>
                          <th className="text-center">จัดการ</th>
                        </tr>
                      </thead>
                      <tbody>
                        {warningRepairs.map(r => (
                          <tr key={r.id}>
                            <td className="font-bold">{r.id}</td>
                            <td>{r.equipment_name || "ไม่ระบุ"} <br /><span className="sub-text">{r.problem_type}</span></td>
                            <td>฿{Number(r.estimated_cost || 0).toLocaleString()}</td>
                            {/* ✅ [อัปเดต] เอา Inline CSS ออก แล้วใช้ class แทน */}
                            <td className="warning-note-text">{r.admin_note}</td>
                            <td className="text-center">
                              <button className="btn-cancel-repair" onClick={() => navigate(`/admin/manage/${r.id}`)}>
                                <Ban size={14} /> ตรวจสอบ
                              </button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  ) : <p className="no-data-text">ไม่มีรายการที่รอพิจารณาในขณะนี้</p>}
                </div>
              </div>
            )}

          </div>
        </>
      )}
    </div>
  );
}

export default AdminDashboard;