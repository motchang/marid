CREATE DATABASE IF NOT EXISTS ecommerce;
USE ecommerce;

CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    comment TEXT NULL
) COMMENT 'Registered customers';

CREATE TABLE products (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(150) NOT NULL,
    sku VARCHAR(64) NOT NULL UNIQUE,
    price DECIMAL(10, 2) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) COMMENT 'Items available for purchase';

CREATE TABLE inventory (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL DEFAULT 0,
    location VARCHAR(64) NOT NULL DEFAULT 'default',
    CONSTRAINT fk_inventory_product FOREIGN KEY (product_id) REFERENCES products(id),
    UNIQUE KEY uk_inventory_product (product_id),
    UNIQUE KEY uk_inventory_location (location)
) COMMENT 'Current on-hand counts for each product and location';

CREATE TABLE orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL,
    total_amount DECIMAL(12, 2) NOT NULL,
    placed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users(id)
) COMMENT 'Customer purchase orders';

CREATE TABLE order_items (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders(id),
    CONSTRAINT fk_order_items_product FOREIGN KEY (product_id) REFERENCES products(id)
) COMMENT 'Line items within an order';

INSERT INTO users (name, email, comment) VALUES
    ('Alice', 'alice@example.com', 'Prefers weekend delivery'),
    ('Bob', 'bob@example.com', 'VIP customer'),
    ('Charlie', 'charlie@example.com', NULL);

INSERT INTO products (name, sku, price) VALUES
    ('Wireless Mouse', 'WM-001', 29.99),
    ('Mechanical Keyboard', 'MK-002', 119.00),
    ('USB-C Hub', 'UH-003', 49.50);

INSERT INTO inventory (product_id, quantity, location) VALUES
    (1, 150, 'tokyo'),
    (2, 80, 'tokyo'),
    (3, 200, 'osaka');

INSERT INTO orders (user_id, status, total_amount, placed_at) VALUES
    (1, 'processing', 148.99, '2024-01-15 10:30:00'),
    (2, 'shipped', 119.00, '2024-02-10 08:15:00');

INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES
    (1, 1, 1, 29.99),
    (1, 3, 2, 59.50),
    (2, 2, 1, 119.00);
