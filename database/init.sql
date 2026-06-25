SET search_path TO public;

-- 1. ตารางผู้ใช้งาน
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    full_name VARCHAR(255) NOT NULL, -- [เพิ่มใหม่] ชื่อแสดงผลจริงบนหน้าเว็บ
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

-- 3. ตารางประเภทปัญหา
CREATE TABLE problem_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- [เพิ่มใหม่] 3.1 ตารางเชื่อมความถนัดของช่าง (Junction Table)
CREATE TABLE technician_specialties (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    problem_type_id INT REFERENCES problem_types(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, problem_type_id)
);

-- 4. ตารางการแจ้งซ่อม 
CREATE TABLE repairs (
    id SERIAL PRIMARY KEY,
    reporter_email VARCHAR(255) NOT NULL, 
    technician_id INT REFERENCES users(id) ON DELETE SET NULL, -- เปลี่ยนเป็น SET NULL เพื่อไม่ให้ใบงานหายหากเผลอลบช่างออก
    location_id INT REFERENCES locations(id),
    floor_id INT REFERENCES floors(id),
    room VARCHAR(100), -- [เพิ่มใหม่] เก็บข้อมูลเลขห้อง หรือพิกัดจุดเกิดเหตุ
    problem_type_id INT REFERENCES problem_types(id),
    description TEXT NOT NULL,
    technician_note TEXT, 
    status VARCHAR(50) DEFAULT 'รอซ่อม', 
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

-- [แก้ไข] เพิ่ม full_name ให้กับบัญชีเริ่มต้น
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
(2, 'ชั้น 1'),
(2, 'ชั้น 2');

INSERT INTO problem_types (name) VALUES
('งานประปา'),
('งานไฟฟ้า'),
('งานอิเล็กทรอนิกส์/โทรศัพท์'),
('อื่นๆ');