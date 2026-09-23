SET search_path TO public;

-- ==========================================
-- โครงสร้างฐานข้อมูลใหม่ (Enterprise Schema)
-- ==========================================

-- 1. [เพิ่มใหม่] ตารางสาขาวิชา/หน่วยงาน (Departments)
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

-- 2. ตารางผู้ใช้งาน (อัปเกรด)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE, -- [เพิ่มใหม่] อีเมลของช่าง/แอดมิน
    role VARCHAR(20) NOT NULL, -- 'admin' หรือ 'technician'
    department_id INT REFERENCES departments(id) ON DELETE SET NULL, -- [เพิ่มใหม่] สังกัดสาขาวิชา
    is_central BOOLEAN DEFAULT FALSE, -- [เพิ่มใหม่] ควบตำแหน่งช่างส่วนกลางหรือไม่?
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. ตารางสถานที่ (ตึก/อาคาร)
CREATE TABLE locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- 3.1 ตารางชั้น (ผูกกับตึกด้วย location_id)
CREATE TABLE floors (
    id SERIAL PRIMARY KEY,
    location_id INT REFERENCES locations(id) ON DELETE CASCADE,
    floor_name VARCHAR(50) NOT NULL
);

-- 3.2 ตารางห้อง (ผูกกับชั้นด้วย floor_id)
CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    floor_id INT REFERENCES floors(id) ON DELETE CASCADE,
    room_number VARCHAR(100) NOT NULL
);

-- 3.3 ตารางอุปกรณ์/ครุภัณฑ์ (ผูกกับห้องด้วย room_id)
CREATE TABLE equipments (
    id SERIAL PRIMARY KEY,
    room_id INT REFERENCES rooms(id) ON DELETE CASCADE, -- ยอมให้เป็น NULL ได้ กรณีหาห้องไม่เจอตอน Import
    asset_code VARCHAR(100) UNIQUE NOT NULL, 
    serial_number VARCHAR(100),                 -- 🔥 [เพิ่มใหม่] เลขที่ผลิตภัณฑ์
    name VARCHAR(255) NOT NULL,              
    category VARCHAR(100),                   
    status VARCHAR(100),                        -- 🔥 [เพิ่มใหม่] สถานะของสินทรัพย์
    acquired_date DATE,                         -- 🔥 [เพิ่มใหม่] วันที่ได้มาครั้งแรก
    expected_life_years INT,                    -- 🔥 [เพิ่มใหม่] อายุการใช้งาน
    base_price DECIMAL(10, 2) DEFAULT 0.00,  
    accumulated_depreciation DECIMAL(10, 2) DEFAULT 0.00, -- 🔥 [เพิ่มใหม่] ค่าเสื่อมสะสม
    book_value DECIMAL(10, 2) DEFAULT 0.00,             -- 🔥 [เพิ่มใหม่] มูลค่าตามบัญชี
    is_active BOOLEAN DEFAULT TRUE           
);

-- 4. ตารางประเภทปัญหา / หมวดหมู่งาน
CREATE TABLE problem_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- 4.1 ตารางเชื่อมความถนัดของช่าง (Junction Table) 
CREATE TABLE technician_specialties (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    problem_type_id INT REFERENCES problem_types(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, problem_type_id)
);

-- 5. ตารางการแจ้งซ่อม (อัปเกรด)
CREATE TABLE repairs (
    id SERIAL PRIMARY KEY,
    ticket_number VARCHAR(50) UNIQUE, -- [เพิ่มใหม่] รหัสอ้างอิงตั๋ว (เช่น REQ-2609001)
    reporter_email VARCHAR(255) NOT NULL, 
    department_id INT REFERENCES departments(id) ON DELETE SET NULL, -- [เพิ่มใหม่] งานนี้เป็นของสาขาไหน?
    technician_id INT REFERENCES users(id) ON DELETE SET NULL,
    location_id INT REFERENCES locations(id),
    other_location VARCHAR(255),                 
    floor_id INT REFERENCES floors(id),
    room_id INT REFERENCES rooms(id),            
    equipment_id INT REFERENCES equipments(id),  
    problem_type_id INT REFERENCES problem_types(id),
    other_problem_type VARCHAR(255),             
    description TEXT NOT NULL,
    technician_note TEXT, -- [จะถูกใช้เพื่อ] บังคับกรอกสรุปงานก่อนปิดจ๊อบ
    admin_note TEXT,                             
    status VARCHAR(50) DEFAULT 'รอซ่อม', 
    estimated_cost DECIMAL(10, 2) DEFAULT 0.00,  
    actual_cost DECIMAL(10, 2) DEFAULT 0.00,     
    accepted_at TIMESTAMP,                       
    completed_at TIMESTAMP,                      
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 6. ตารางเก็บรูปภาพ
CREATE TABLE repair_images (
    id SERIAL PRIMARY KEY,
    repair_id INT REFERENCES repairs(id) ON DELETE CASCADE,
    image_url TEXT,
    image_type VARCHAR(20) DEFAULT 'before' 
);

-- 7. [เพิ่มใหม่] ตารางประวัติการทำรายการ (Activity Log)
CREATE TABLE repair_logs (
    id SERIAL PRIMARY KEY,
    repair_id INT REFERENCES repairs(id) ON DELETE CASCADE, -- ผูกกับใบแจ้งซ่อม
    user_id INT REFERENCES users(id) ON DELETE SET NULL, -- ใครเป็นคนทำแอคชันนี้
    action VARCHAR(100) NOT NULL, -- เช่น 'CREATED', 'STATUS_CHANGED', 'ASSIGNED'
    old_status VARCHAR(50), -- สถานะก่อนเปลี่ยน
    new_status VARCHAR(50), -- สถานะที่เปลี่ยนใหม่
    note TEXT, -- ข้อความเช่น 'ช่างสมชายรับงานแล้ว', 'เปลี่ยนสถานะเป็นกำลังดำเนินการ'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==========================================
-- ข้อมูลเริ่มต้น (Master Data)
-- ==========================================

-- 1. สร้างข้อมูลสาขาวิชา
INSERT INTO departments (name) VALUES
('สาขาชีววิทยา'), -- id 1
('สาขาเคมี'), -- id 2
('สาขาฟิสิกส์'), -- id 3
('สาขาวิทยาศาสตร์สิ่งแวดล้อม'), -- id 4
('สาขาเทคโนโลยีสารสนเทศ'); -- id 5

INSERT INTO users (username, password, full_name, email, role, department_id, is_central) VALUES
('admin', '$2a$12$ixARKu9I12oi7m51r.TSbO0L3QeLKB/9yT8711IWvvBF7radYi3YC', 'ผู้ดูแลระบบสูงสุด', 'admin@science.uni.ac.th', 'admin', null, false),
('tech_bio', '$2a$12$ixARKu9I12oi7m51r.TSbO0L3QeLKB/9yT8711IWvvBF7radYi3YC', 'นายชีวะ ซ่อมได้', 'tech.bio@science.uni.ac.th', 'technician', 1, false), -- ช่างชีวะ (รับแค่งานชีวะ)
('tech_it', '$2a$12$ixARKu9I12oi7m51r.TSbO0L3QeLKB/9yT8711IWvvBF7radYi3YC', 'นายไอที รับจบ', 'tech.it@science.uni.ac.th', 'technician', 5, true); -- ช่างไอที (is_central = true รับงานสาขาอื่นได้)

-- 2. สร้างข้อมูลอาคาร
INSERT INTO locations (name) VALUES
('อาคารวิทยาศาสตร์ 1'), -- id 1
('อาคารวิทยาศาสตร์ 2'), -- id 2
('อาคารวิทยาศาสตร์ 3'), -- id 3
('อาคารวิทยาศาสตร์ 4'), -- id 4
('อื่นๆ');               -- id 5

-- 3. สร้างข้อมูลชั้น (เพิ่มให้ครบตามจริงทั้งหมด)
INSERT INTO floors (location_id, floor_name) VALUES
-- อาคาร 1 (มี 9 ชั้น) -> ได้ id 1-9
(1, 'ชั้น 1'), (1, 'ชั้น 2'), (1, 'ชั้น 3'), (1, 'ชั้น 4'), (1, 'ชั้น 5'), 
(1, 'ชั้น 6'), (1, 'ชั้น 7'), (1, 'ชั้น 8'), (1, 'ชั้น 9'),
-- อาคาร 2 (มี 2 ชั้น) -> ได้ id 10-11
(2, 'ชั้น 1'), (2, 'ชั้น 2'),
-- อาคาร 3 (มี 5 ชั้น) -> ได้ id 12-16
(3, 'ชั้น 1'), (3, 'ชั้น 2'), (3, 'ชั้น 3'), (3, 'ชั้น 4'), (3, 'ชั้น 5'),
-- อาคาร 4 (มี 5 ชั้น) -> ได้ id 17-21
(4, 'ชั้น 1'), (4, 'ชั้น 2'), (4, 'ชั้น 3'), (4, 'ชั้น 4'), (4, 'ชั้น 5');

-- 4. นำเข้าข้อมูลห้อง (ไม่ซ้ำ, เรียงตัวหนังสือ -> ตัวเลข)
INSERT INTO rooms (floor_id, room_number) VALUES
-- =========================================
-- อาคารวิทยาศาสตร์ 1
-- =========================================
-- วิทย์ 1 ชั้น 2 (floor_id = 2)
(2, 'ห้องน้ำชาย'),
(2, 'ห้องน้ำหญิง'),
(2, 'ห้อง 1227/1'),
(2, 'ห้อง 1227/2'),
(2, 'ห้อง 1231'),
(2, 'ห้อง 1232'),
(2, 'ห้อง 1233'),
(2, 'ห้อง 1234'),
(2, 'ห้อง 1235'),
(2, 'ห้อง 1236'),
(2, 'ห้อง 1239'),
(2, 'ห้อง 1240'),
(2, 'ห้อง 1241'),
(2, 'ห้อง 1242'),

-- วิทย์ 1 ชั้น 3 (floor_id = 3)
(3, 'ห้องไววิทย์'),
(3, 'ห้อง 1301'),
(3, 'ห้อง 1320'),
(3, 'ห้อง 1321'),
(3, 'ห้อง 1323'),

-- วิทย์ 1 ชั้น 4 (floor_id = 4)
(4, 'ห้องพักครู'),
(4, 'ห้องลิฟ'),
(4, 'ห้อง 1437'),

-- วิทย์ 1 ชั้น 6 (floor_id = 6)
(6, 'ห้องพักครู'),
(6, 'ห้อง 1601'),
(6, 'ห้อง 1637'),
(6, 'ห้อง 1642'),
(6, 'ห้อง 1646'),
(6, 'ห้อง 1655'),
(6, 'ห้อง 1656'),
(6, 'ห้อง 1657'),
(6, 'ห้อง 1658'),

-- =========================================
-- อาคารวิทยาศาสตร์ 2
-- =========================================
-- วิทย์ 2 ชั้น 1 (floor_id = 10)
(10, 'ห้องกระจก'),

-- วิทย์ 2 ชั้น 2 (floor_id = 11)
(11, 'ห้อง ร.วท.1'),
(11, 'ห้อง ร.วท.2'),

-- =========================================
-- อาคารวิทยาศาสตร์ 3
-- =========================================
-- วิทย์ 3 ชั้น 2 (floor_id = 13)
(13, 'ห้องสัมมนา'),
(13, 'ห้อง 3203'),
(13, 'ห้อง 3204'),

-- =========================================
-- อาคารวิทยาศาสตร์ 4 (แบบมีประเภทห้องต่อท้าย)
-- =========================================
-- วิทย์ 4 ชั้น 1 (floor_id = 17)
(17, 'ห้อง 4101'),
(17, 'ห้อง 4103 (ห้องบรรยาย 150 คน)'),
(17, 'ห้อง 4104 (ห้องบรรยาย 50 คน)'),
(17, 'ห้อง 4105 (ห้องบรรยาย 50 คน)'),
(17, 'ห้อง 4106 (ห้องบรรยาย 50 คน)'),
(17, 'ห้อง 4111 (ห้องปฏิบัติการพื้นฐานฟิสิกส์)'),
(17, 'ห้อง 4112 (ห้องปฏิบัติการพื้นฐานฟิสิกส์)'),
(17, 'ห้อง 4113 (ห้องเตรียมปฏิบัติการพื้นฐานฟิสิกส์)'),
(17, 'ห้อง 4121 (ห้องเตรียมปฏิบัติการพื้นฐานฟิสิกส์)'),
(17, 'ห้อง 4122 (ห้องปฏิบัติการพื้นฐานฟิสิกส์)'),
(17, 'ห้อง 4123 (ห้องปฏิบัติการพื้นฐานฟิสิกส์)'),
(17, 'ห้อง 5104 (ศูนย์เครื่องมือวิทยาศาสตร์)'),
(17, 'ห้อง 5105 (สำนักงานศูนย์เครื่องมือวิทยาศาสตร์)'),

-- วิทย์ 4 ชั้น 2 (floor_id = 18)
(18, 'ห้อง 4203 (ห้องบรรยาย 100 คน)'),
(18, 'ห้อง 4204 (ห้องบรรยาย 100 คน)'),
(18, 'ห้อง 4205 (ห้องบรรยาย 100 คน)'),
(18, 'ห้อง 4206 (ห้องบรรยาย 100 คน)'),
(18, 'ห้อง 4210 (ห้องพักอาจารย์ภาควิชาเคมี)'),
(18, 'ห้อง 4219 (ห้องปฏิบัติการอิเลคทรอนิค)'),
(18, 'ห้อง 4222 (ห้องควบคุมอิเลคทรอนิค)'),
(18, 'ห้อง 4223 (ห้องปฏิบัติการโมเดิร์นฟิสิกส์)'),
(18, 'ห้อง 4224 (ห้องควบคุมโมเดิร์นฟิสิกส์)'),
(18, 'ห้อง 4232 (ห้องเตรียมปฏิบัติการ)'),
(18, 'ห้อง 4233 (ห้องปฏิบัติการเคมีวิเคราะห์)'),
(18, 'ห้อง 4237 (ห้องปฏิบัติการอินทรีย์เคมี)'),
(18, 'ห้อง 5204 (ศูนย์เครื่องมือวิทยาศาสตร์)'),
(18, 'ห้อง 5205 (สำนักงานศูนย์เครื่องมือวิทยาศาสตร์)'),

-- วิทย์ 4 ชั้น 3 (floor_id = 19)
(19, 'ห้อง 4303 (ห้องเตรียมปฏิบัติการพื้นฐานเคมี)'),
(19, 'ห้อง 4304 (ห้องเครื่องชั่ง)'),
(19, 'ห้อง 4305 (ห้องเตรียมปฏิบัติการชีวเคมี)'),
(19, 'ห้อง 4306 (ห้องปฏิบัติการชีวเคมี)'),
(19, 'ห้อง 4310 (ห้องปฏิบัติการพื้นฐานเคมี)'),
(19, 'ห้อง 4317 (ห้องกลั่นน้ำ)'),
(19, 'ห้อง 4318 (ห้องปฏิบัติการพื้นฐานเคมี)'),
(19, 'ห้อง 4320 (ห้องประชุมสัมมนา)'),
(19, 'ห้อง 4321 (ห้องประชุมสัมมนา)'),
(19, 'ห้อง 4324 (ห้องประชุมสัมมนา)'),
(19, 'ห้อง 5304 (ห้องแสดงนิทรรศการ)'),
(19, 'ห้อง 5305 (ห้องเก็บพัสดุ)'),

-- วิทย์ 4 ชั้น 4 (floor_id = 20)
(20, 'ห้อง 4403 (ห้องเตรียมปฏิบัติการพื้นฐานชีววิทยา)'),
(20, 'ห้อง 4405 (ห้องเตรียมปฏิบัติการพฤกษศาสตร์)'),
(20, 'ห้อง 4407 (ห้องปฏิบัติการพฤกษศาสตร์)'),
(20, 'ห้อง 4408 (ห้องปฏิบัติการพฤกษศาสตร์)'),
(20, 'ห้อง 4409 (ห้องเตรียมปฏิบัติการสัตววิทยา)'),
(20, 'ห้อง 4411 (ห้องปฏิบัติการสัตววิทยา)'),
(20, 'ห้อง 4415 (ห้องพักอาจารย์ภาควิชาชีววิทยา)'),
(20, 'ห้อง 4422 (ห้องปฏิบัติการพื้นฐานชีววิทยา)'),
(20, 'ห้อง 4423 (ห้องปฏิบัติการพื้นฐานชีววิทยา)'),
(20, 'ห้อง 4430 (ห้องเตรียมปฏิบัติการพื้นฐานชีววิทยา)'),
(20, 'ห้อง 4431 (ห้องปฏิบัติการพื้นฐานชีววิทยา)'),
(20, 'ห้อง 4432 (ห้องปฏิบัติการพื้นฐานชีววิทยา)'),
(20, 'ห้อง 5406 (ห้องบรรยายรวม)'),
(20, 'ห้อง 5408 (ห้องโสตทัศนอุปกรณ์)'),

-- วิทย์ 4 ชั้น 5 (floor_id = 21)
(21, 'ห้อง 4503 (ห้องปฏิบัติการจุลชีววิทยา)'),
(21, 'ห้อง 4510 (ห้องเตรียมปฏิบัติการจุลชีววิทยา)'),
(21, 'ห้อง 4511 (ห้องปฏิบัติการสิ่งแวดล้อม)'),
(21, 'ห้อง 4515 (ห้องปฏิบัติการสิ่งแวดล้อม)'),
(21, 'ห้อง 4516 (ห้องปฏิบัติการสิ่งแวดล้อม)'),
(21, 'ห้อง 4521 (ห้องพักอาจารย์ภาควิชาสิ่งแวดล้อม)'),
(21, 'ห้อง 4522 (ห้องประชุมภาควิชาสิ่งแวดล้อม)'),
(21, 'ห้อง 4528 (ห้องปฏิบัติการจุลชีววิทยา)'),
(21, 'ห้อง 4529 (ห้องปฏิบัติการจุลชีววิทยา)'),
(21, 'ห้อง 4540 (ห้องพักอาจารย์สิ่งแวดล้อม)'),
(21, 'ห้อง 4541 (ห้องปฏิบัติการวิจัยสิ่งแวดล้อม)'),
(21, 'ห้อง 4544 (ห้องปฏิบัติการวิจัยสิ่งแวดล้อม)'),
(21, 'ห้อง 4546 (ห้องปฏิบัติการวิจัยสิ่งแวดล้อม)'),
(21, 'ห้อง 4549 (ห้องปฏิบัติการวิจัยสิ่งแวดล้อม)');

INSERT INTO equipments (room_id, asset_code, serial_number, name, category, status, base_price) VALUES
(3, 'AC-1227/1-A', 'SN-001', 'แอร์ Daikin 24000 BTU ตัวซ้าย', 'Air Conditioner', 'พร้อมใช้งาน', 25000.00),
(4, 'AC-1227/2-B', 'SN-002', 'แอร์ Daikin 24000 BTU ตัวขวา', 'Air Conditioner', 'พร้อมใช้งาน', 25000.00);

INSERT INTO problem_types (name) VALUES
('งานประปา'),
('งานไฟฟ้า'),
('งานอิเล็กทรอนิกส์/โทรศัพท์'),
('งานห้องเรียน'),       
('งานเสียงและภาพ'),    
('อื่นๆ');