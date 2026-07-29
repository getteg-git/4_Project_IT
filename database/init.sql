SET search_path TO public;

-- ==========================================
-- โครงสร้างฐานข้อมูล (Schema)
-- ==========================================

-- 1. ตารางผู้ใช้งาน
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL, -- 'admin' หรือ 'technician'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. ตารางสถานที่ (ตึก/อาคาร)
CREATE TABLE locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- 2.1 ตารางชั้น (ผูกกับตึกด้วย location_id)
CREATE TABLE floors (
    id SERIAL PRIMARY KEY,
    location_id INT REFERENCES locations(id) ON DELETE CASCADE,
    floor_name VARCHAR(50) NOT NULL
);

-- 2.2 ตารางห้อง (ผูกกับชั้นด้วย floor_id)
CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    floor_id INT REFERENCES floors(id) ON DELETE CASCADE,
    room_number VARCHAR(100) NOT NULL
);

-- 2.3 ตารางอุปกรณ์/ครุภัณฑ์ (ผูกกับห้องด้วย room_id)
CREATE TABLE equipments (
    id SERIAL PRIMARY KEY,
    room_id INT REFERENCES rooms(id) ON DELETE CASCADE,
    asset_code VARCHAR(100) UNIQUE NOT NULL, -- รหัสครุภัณฑ์ (เช่น AC-101-A)
    name VARCHAR(255) NOT NULL,              -- ชื่ออุปกรณ์
    category VARCHAR(100),                   -- ประเภทอุปกรณ์ (แอร์, PC ฯลฯ)
    base_price DECIMAL(10, 2) DEFAULT 0.00,  -- ราคาต้นทุน
    is_active BOOLEAN DEFAULT TRUE           -- สถานะใช้งาน
);

-- 3. ตารางประเภทปัญหา
CREATE TABLE problem_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- 3.1 ตารางเชื่อมความถนัดของช่าง (Junction Table)
CREATE TABLE technician_specialties (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    problem_type_id INT REFERENCES problem_types(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, problem_type_id)
);

-- 4. ตารางการแจ้งซ่อม (แก้ไขและเพิ่มฟิลด์ใหม่)
CREATE TABLE repairs (
    id SERIAL PRIMARY KEY,
    reporter_email VARCHAR(255) NOT NULL, 
    technician_id INT REFERENCES users(id) ON DELETE SET NULL,
    location_id INT REFERENCES locations(id),
    other_location VARCHAR(255),                 -- ระบุสถานที่/อาคารเพิ่มเติม
    floor_id INT REFERENCES floors(id),
    room_id INT REFERENCES rooms(id),            
    equipment_id INT REFERENCES equipments(id),  
    problem_type_id INT REFERENCES problem_types(id),
    other_problem_type VARCHAR(255),             -- ระบุหมวดหมู่งานซ่อมเพิ่มเติม
    description TEXT NOT NULL,
    technician_note TEXT, 
    admin_note TEXT,                             -- หมายเหตุจากแอดมิน (เช่น เหตุผลที่ยกเลิกงาน)
    status VARCHAR(50) DEFAULT 'รอซ่อม', 
    estimated_cost DECIMAL(10, 2) DEFAULT 0.00,  -- 🔥 [เพิ่มใหม่] ราคาประเมินเบื้องต้นจากช่าง
    actual_cost DECIMAL(10, 2) DEFAULT 0.00,     -- 🔥 [ปรับปรุง] ค่าใช้จ่ายจริงเมื่อซ่อมเสร็จ
    accepted_at TIMESTAMP,                       
    completed_at TIMESTAMP,                      
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. ตารางเก็บรูปภาพ
CREATE TABLE repair_images (
    id SERIAL PRIMARY KEY,
    repair_id INT REFERENCES repairs(id) ON DELETE CASCADE,
    image_url TEXT,
    image_type VARCHAR(20) DEFAULT 'before' 
);

-- ==========================================
-- ข้อมูลเริ่มต้น (Master Data)
-- ==========================================

INSERT INTO users (username, password, full_name, role) VALUES
('admin', '1234', 'ผู้ดูแลระบบสูงสุด', 'admin'),
('tech1', '1234', 'นายสมชาย ยอดช่าง', 'technician'); 

INSERT INTO locations (name) VALUES
('อาคารวิทยาศาสตร์ 1'),
('อาคารวิทยาศาสตร์ 2'),
('อาคารวิทยาศาสตร์ 3'),
('อาคารวิทยาศาสตร์ 4'),
('อื่นๆ');

INSERT INTO floors (location_id, floor_name) VALUES
(1, 'ชั้น 1'),
(1, 'ชั้น 2'),
(1, 'ชั้น 3'),
(1, 'ชั้น 4'),
(1, 'ชั้น 5'),
(1, 'ชั้น 6'),
(2, 'ชั้น 1'),
(2, 'ชั้น 2'),
(3, 'ชั้น 1'),
(3, 'ชั้น 2'),
(4, 'ชั้น 1'),
(4, 'ชั้น 2'),
(4, 'ชั้น 3'),
(4, 'ชั้น 4');

INSERT INTO rooms (floor_id, room_number) VALUES
(2, 'ห้อง 1227/1'), -- อาคาร 1 ชั้น 2
(2, 'ห้อง 1227/2'), -- อาคาร 1 ชั้น 2
(2, 'ห้อง 1239'), -- อาคาร 1 ชั้น 2
(6, 'ห้อง 1601'), -- อาคาร 1 ชั้น 6
(11, 'ห้อง 4101'); -- อาคาร 4 ชั้น 1

INSERT INTO equipments (room_id, asset_code, name, category, base_price) VALUES
(1, 'AC-1227/1-A', 'แอร์ Daikin 24000 BTU ตัวซ้าย', 'Air Conditioner', 25000.00),
(2, 'AC-1227/2-B', 'แอร์ Daikin 24000 BTU ตัวขวา', 'Air Conditioner', 25000.00),
(4, 'PC-1601-01', 'คอมพิวเตอร์อาจารย์', 'IT Equipment', 30000.00);

INSERT INTO problem_types (name) VALUES
('งานประปา'),
('งานไฟฟ้า'),
('งานอิเล็กทรอนิกส์/โทรศัพท์'),
('งานห้องเรียน'),       
('งานเสียงและภาพ'),    
('อื่นๆ');