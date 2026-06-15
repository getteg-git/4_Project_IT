SET search_path TO public;

-- 1. ตารางผู้ใช้งาน
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role VARCHAR(20) NOT NULL, -- 'admin' หรือ 'technician'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. ตารางสถานที่ (ตึก/อาคาร)
CREATE TABLE locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- [เพิ่มใหม่] 2.1 ตารางชั้น (ผูกกับตึกด้วย location_id)
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

-- 4. ตารางการแจ้งซ่อม 
CREATE TABLE repairs (
    id SERIAL PRIMARY KEY,
    reporter_email VARCHAR(255) NOT NULL, 
    technician_id INT REFERENCES users(id), 
    location_id INT REFERENCES locations(id),
    floor_id INT REFERENCES floors(id), -- [เพิ่มใหม่] เก็บข้อมูลชั้น
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

INSERT INTO users (username, password, role) VALUES
('admin', '1234', 'admin'),
('tech1', '1234', 'technician'); 

INSERT INTO locations (name) VALUES
('อาคารวิทยาศาสตร์ 1'),
('อาคารวิทยาศาสตร์ 2'),
('อาคารวิทยาศาสตร์ 3'),
('อาคารวิทยาศาสตร์ 4'),
('อื่นๆ');

-- [เพิ่มใหม่] Insert ข้อมูล 4 ชั้น (วิทย์ 1 มีชั้น 1-2, วิทย์ 2 มีชั้น 1-2)
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