CREATE DATABASE IF NOT EXISTS test;
USE test;
CREATE TABLE IF NOT EXISTS users(
                                  id INT AUTO_INCREMENT PRIMARY KEY,
                                  email VARCHAR(50) NOT NULL,
                                  password VARCHAR(100) NOT NULL
);