import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { 
  PieChart, Pie, Cell, Tooltip, ResponsiveContainer, 
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Legend 
} from "recharts";
import "./AdminDashboard.css";

function AdminDashboard() {
  const navigate = useNavigate();
  const [repairs, setRepairs] = useState([]);
  const [isLoading, setIsLoading] = useState(true);

  // ดึงข้อมูลจริงจาก Backend
  useEffect(() => {
    const fetchRepairs = async () => {
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
    fetchRepairs();
  }, []);

  // ==========================================
  // Logic การคำนวณข้อมูลสถิติ (Real Data)
  // ==========================================
  
  // 1. สรุปตัวเลขการ์ดด้านบน
  const totalRepairs = repairs.length;
  const pendingCount = repairs.filter(r => r.status === "รอซ่อม").length;
  const progressCount = repairs.filter(r => r.status === "กำลังซ่อม").length;
  const completedCount = repairs.filter(r => r.status === "เสร็จเรียบร้อย").length;
  const failedCount = repairs.filter(r => r.status === "ซ่อมไม่ได้").length;

  // 2. ข้อมูลสำหรับกราฟโดนัท (สัดส่วนสถานะงาน)
  const statusChartData = [
    { name: "รอซ่อม", value: pendingCount, color: "#f39c12" },      // สีส้ม
    { name: "กำลังซ่อม", value: progressCount, color: "#2980b9" },    // สีฟ้า
    { name: "เสร็จเรียบร้อย", value: completedCount, color: "#007A53" },// สีเขียวศิลปากร
    { name: "ซ่อมไม่ได้", value: failedCount, color: "#e74c3c" }      // สีแดง
  ].filter(item => item.value > 0); // เอาเฉพาะอันที่มีค่าไปแสดง

  // 3. ข้อมูลสำหรับกราฟแท่ง (สรุปประเภทปัญหาที่พบบ่อย)
  const getProblemTypeStats = () => {
    const typeCounts = {};
    repairs.forEach(r => {
      const type = r.problem_type || "ไม่ระบุ";
      typeCounts[type] = (typeCounts[type] || 0) + 1;
    });
    return Object.keys(typeCounts).map(key => ({
      name: key,
      จำนวน: typeCounts[key]
    })).sort((a, b) => b.จำนวน - a.จำนวน); // เรียงจากมากไปน้อย
  };
  const problemTypeData = getProblemTypeStats();

  // ฟังก์ชันออกจากระบบ
  const handleLogout = () => {
    localStorage.removeItem("user");
    navigate("/");
  };

  return (
    <div className="admin-dashboard-container">
      {/* 1. ส่วนหัว (Header & Navigation) */}
      <header className="dashboard-header">
        <div>
          <h1>📊 แผงควบคุมผู้ดูแลระบบ (Admin Dashboard)</h1>
          <p>สรุปภาพรวมระบบแจ้งซ่อมบำรุง คณะวิทยาศาสตร์ มหาวิทยาลัยศิลปากร</p>
        </div>
        
        {/* 🔥 เพิ่มปุ่ม จัดการบัญชีผู้ใช้ ตรงนี้ครับ */}
        <div className="header-actions">
          <button className="btn-users" onClick={() => navigate("/admin/users")}>
            👥 จัดการบัญชีผู้ใช้
          </button>
          <button className="btn-manage" onClick={() => navigate("/admin/manage")}>
            🛠️ ไปหน้าจัดการมอบหมายงาน
          </button>
          <button className="btn-logout" onClick={handleLogout}>
            🚪 ออกจากระบบ
          </button>
        </div>
      </header>

      {isLoading ? (
        <div className="loading-state">⏳ กำลังโหลดและคำนวณข้อมูล...</div>
      ) : (
        <>
          {/* 2. ส่วนการ์ดสรุปตัวเลข (Summary Cards) */}
          <div className="summary-cards-grid">
            <div className="summary-card total">
              <h3>รวมทั้งหมด</h3>
              <div className="number">{totalRepairs}</div>
              <span>รายการแจ้งซ่อม</span>
            </div>
            <div className="summary-card pending">
              <h3>รอจ่ายงานช่าง</h3>
              <div className="number">{pendingCount}</div>
              <span>รายการใหม่</span>
            </div>
            <div className="summary-card progress">
              <h3>ช่างกำลังซ่อม</h3>
              <div className="number">{progressCount}</div>
              <span>อยู่ระหว่างดำเนินการ</span>
            </div>
            <div className="summary-card completed">
              <h3>งานซ่อมสำเร็จ</h3>
              <div className="number">{completedCount}</div>
              <span>ปิดงานเรียบร้อย</span>
            </div>
          </div>

          {/* 3. ส่วนแสดงกราฟ (Charts) */}
          <div className="charts-grid">
            
            {/* กราฟโดนัท: สัดส่วนสถานะงาน */}
            <div className="chart-card">
              <h3>⭕ สัดส่วนสถานะงานซ่อมทั้งหมด</h3>
              <div className="chart-wrapper">
                {statusChartData.length > 0 ? (
                  <ResponsiveContainer width="100%" height={300}>
                    <PieChart>
                      <Pie 
                        data={statusChartData} 
                        cx="50%" cy="50%" 
                        innerRadius={70} 
                        outerRadius={100} 
                        paddingAngle={5} 
                        dataKey="value"
                      >
                        {statusChartData.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={entry.color} />
                        ))}
                      </Pie>
                      <Tooltip formatter={(value) => [`${value} รายการ`, "จำนวน"]} />
                      <Legend verticalAlign="bottom" height={36}/>
                    </PieChart>
                  </ResponsiveContainer>
                ) : (
                  <p className="no-data-text">ยังไม่มีข้อมูลรายการแจ้งซ่อม</p>
                )}
              </div>
            </div>

            {/* กราฟแท่ง: ปัญหาที่พบบ่อย */}
            <div className="chart-card">
              <h3>📈 สถิติประเภทปัญหาที่เกิดบ่อยที่สุด</h3>
              <div className="chart-wrapper">
                {problemTypeData.length > 0 ? (
                  <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={problemTypeData} margin={{ top: 20, right: 30, left: 0, bottom: 5 }}>
                      <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#eee" />
                      <XAxis dataKey="name" tick={{fill: '#555'}} />
                      <YAxis allowDecimals={false} tick={{fill: '#555'}} />
                      <Tooltip cursor={{fill: '#f4f6f5'}} />
                      <Bar dataKey="จำนวน" fill="#007A53" radius={[4, 4, 0, 0]} />
                    </BarChart>
                  </ResponsiveContainer>
                ) : (
                  <p className="no-data-text">ยังไม่มีข้อมูลหมวดหมู่ปัญหา</p>
                )}
              </div>
            </div>

          </div>
        </>
      )}
    </div>
  );
}

export default AdminDashboard;