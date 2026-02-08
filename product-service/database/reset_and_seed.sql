-- Reset and Seed Product Data
-- Validates that the table is clean before inserting sample data

-- 1. Clear existing data and reset ID sequence
TRUNCATE TABLE products RESTART IDENTITY;

-- 2. Insert Sample Data

-- Electronics
INSERT INTO products (name, description, price, stock, category, image_url) VALUES
('MacBook Pro 16"', 'Apple M3 Max chip with 16-core CPU and 40-core GPU, 64GB unified memory, 1TB SSD storage', 3499.00, 25, 'Electronics', 'https://picsum.photos/seed/laptop1/400/300'),
('Sony WH-1000XM5 Headphones', 'Industry-leading noise canceling wireless headphones with 30-hour battery life', 399.99, 50, 'Electronics', 'https://picsum.photos/seed/headphones1/400/300'),
('iPhone 15 Pro Max', 'A17 Pro chip, titanium design, 6.7-inch Super Retina XDR display, 256GB', 1199.00, 100, 'Electronics', 'https://picsum.photos/seed/phone1/400/300'),
('Samsung 65" QLED TV', '4K Ultra HD Smart TV with Quantum Processor and HDR10+', 1299.99, 15, 'Electronics', 'https://picsum.photos/seed/tv1/400/300'),
('Dell UltraSharp Monitor', '27-inch 4K USB-C monitor with 99% sRGB color coverage', 549.00, 40, 'Electronics', 'https://picsum.photos/seed/monitor1/400/300');

-- Clothing
INSERT INTO products (name, description, price, stock, category, image_url) VALUES
('Levi''s 501 Original Jeans', 'Classic straight fit denim jeans with button fly, available in multiple washes', 69.99, 200, 'Clothing', 'https://picsum.photos/seed/jeans1/400/300'),
('Nike Air Max Sneakers', 'Comfortable running shoes with visible Air cushioning and breathable mesh', 129.99, 150, 'Clothing', 'https://picsum.photos/seed/shoes1/400/300'),
('Patagonia Down Jacket', 'Lightweight insulated jacket with 800-fill-power down, water-resistant shell', 229.00, 75, 'Clothing', 'https://picsum.photos/seed/jacket1/400/300'),
('Ralph Lauren Oxford Shirt', 'Classic fit button-down shirt in 100% cotton, available in multiple colors', 89.50, 120, 'Clothing', 'https://picsum.photos/seed/shirt1/400/300');

-- Books
INSERT INTO products (name, description, price, stock, category, image_url) VALUES
('The Pragmatic Programmer', 'Your journey to mastery, 20th Anniversary Edition by David Thomas and Andrew Hunt', 49.99, 80, 'Books', 'https://picsum.photos/seed/book1/400/300'),
('Atomic Habits', 'An easy and proven way to build good habits and break bad ones by James Clear', 27.00, 150, 'Books', 'https://picsum.photos/seed/book2/400/300'),
('The Art of War', 'Ancient Chinese military treatise by Sun Tzu, deluxe hardcover edition', 19.99, 200, 'Books', 'https://picsum.photos/seed/book3/400/300');

-- Home & Garden
INSERT INTO products (name, description, price, stock, category, image_url) VALUES
('Ergonomic Office Chair', 'Adjustable lumbar support, breathable mesh back, 360-degree swivel, up to 300 lbs', 299.99, 45, 'Home & Garden', 'https://picsum.photos/seed/chair1/400/300'),
('Dyson V15 Vacuum', 'Cordless stick vacuum with laser detection and LCD screen showing particle count', 649.99, 30, 'Home & Garden', 'https://picsum.photos/seed/vacuum1/400/300'),
('KitchenAid Stand Mixer', '5-quart tilt-head stand mixer with 10 speeds and stainless steel bowl', 379.99, 60, 'Home & Garden', 'https://picsum.photos/seed/mixer1/400/300'),
('Weber Gas Grill', '3-burner propane gas grill with 529 sq. in. cooking area and side burner', 499.00, 20, 'Home & Garden', 'https://picsum.photos/seed/grill1/400/300');
