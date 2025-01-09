CREATE DATABASE IF NOT EXISTS airbnb;
USE airbnb;

CREATE TABLE IF NOT EXISTS airbnb_bookings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    task_name VARCHAR(255),
    property VARCHAR(255),
    price VARCHAR(50),
    description TEXT
);
