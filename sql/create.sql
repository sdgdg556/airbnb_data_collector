CREATE DATABASE IF NOT EXISTS airbnb;
USE airbnb;

CREATE TABLE IF NOT EXISTS airbnb_bookings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    hotel_name VARCHAR(255) NOT NULL,
    star DECIMAL(2, 1),
    price DECIMAL(10, 2) NOT NULL,
    price_before_taxes DECIMAL(10, 2),
    check_in_date DATE NOT NULL,
    check_out_date DATE NOT NULL,
    guests INT NOT NULL
);