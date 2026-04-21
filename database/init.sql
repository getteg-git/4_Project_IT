SET search_path TO public;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE problem_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE repairs (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    location_id INT REFERENCES locations(id),
    problem_type_id INT REFERENCES problem_types(id),
    description TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE repair_images (
    id SERIAL PRIMARY KEY,
    repair_id INT REFERENCES repairs(id),
    image_url TEXT
);

INSERT INTO users (username, password, role) VALUES
('GETTEG', '1234', 'user'),
('admin', '1234', 'admin');

INSERT INTO locations (name) VALUES
('อาคารวิทยาศาสตร์ 1'),
('อาคารวิทยาศาสตร์ 2'),
('ห้องปฏิบัติการ'),
('ห้องเรียน'),
('ห้องน้ำ');

INSERT INTO problem_types (name) VALUES
('คอมพิวเตอร์'),
('อินเทอร์เน็ต'),
('เครื่องพิมพ์'),
('ไฟฟ้า'),
('อื่นๆ');