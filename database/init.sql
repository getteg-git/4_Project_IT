SET search_path TO public;

-- ตารางผู้ใช้งาน
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ตารางสถานที่ (ตึก/ห้อง)
CREATE TABLE locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- ตารางประเภทปัญหา
CREATE TABLE problem_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- ตารางการแจ้งซ่อม (แก้ไขเพิ่ม technician_id)
CREATE TABLE repairs (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    technician_id INT REFERENCES users(id), -- [เพิ่มใหม่] เก็บ ID ของช่างที่แอดมินมอบหมายงานให้
    location_id INT REFERENCES locations(id),
    problem_type_id INT REFERENCES problem_types(id),
    description TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ตารางเก็บรูปภาพ (แก้ไขเพิ่ม image_type)
CREATE TABLE repair_images (
    id SERIAL PRIMARY KEY,
    repair_id INT REFERENCES repairs(id),
    image_url TEXT,
    image_type VARCHAR(20) DEFAULT 'before' -- [เพิ่มใหม่] 'before' = รูปตอนผู้ใช้แจ้ง, 'during' = รูปตอนช่างซ่อม
);

-- Insert ข้อมูลผู้ใช้งานเริ่มต้น (เพิ่มช่าง)
INSERT INTO users (username, password, role) VALUES
('GETTEG', '1234', 'user'),
('admin', '1234', 'admin'),
('tech1', '1234', 'technician'); 

-- Insert ข้อมูลสถานที่
INSERT INTO locations (name) VALUES
('อาคารวิทยาศาสตร์ 1'),
('อาคารวิทยาศาสตร์ 2'),
('ห้องปฏิบัติการ'),
('ห้องเรียน'),
('ห้องน้ำ');

-- Insert ข้อมูลประเภทปัญหา
INSERT INTO problem_types (name) VALUES
('คอมพิวเตอร์'),
('อินเทอร์เน็ต'),
('เครื่องพิมพ์'),
('ไฟฟ้า'),
('อื่นๆ');