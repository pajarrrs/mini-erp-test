INSERT INTO products (sku, name, price, stock, is_active, created_at) VALUES
    ('001', 'Beras', 78500, 120, TRUE, now() - interval '30 days'),
    ('002', 'Minyak', 38900, 90, TRUE, now() - interval '30 days'),
    ('003', 'Gula', 17500, 200, TRUE, now() - interval '30 days'),
    ('004', 'Tepung Terigu', 13200, 150, TRUE, now() - interval '30 days'),
    ('005', 'Mi Instan', 118000, 40, TRUE, now() - interval '25 days'),
    ('006', 'Kopi Bubuk', 27500, 8, TRUE, now() - interval '25 days'),
    ('007', 'Teh', 8900, 180, TRUE, now() - interval '25 days'),
    ('008', 'Susu UHT', 19800, 96, TRUE, now() - interval '20 days'),
    ('009', 'Air Mineral', 52000, 60, TRUE, now() - interval '20 days'),
    ('010', 'Sabun Mandi', 4500, 6, TRUE, now() - interval '20 days'),
    ('011', 'Deterjen', 24900, 75, TRUE, now() - interval '15 days'),
    ('012', 'Kecap Manis', 21500, 64, TRUE, now() - interval '15 days'),
    ('013', 'Sarden', 28700, 5, FALSE, now() - interval '10 days'),
    ('014', 'Telur Ayam', 56000, 0, TRUE, now() - interval '1 day');

INSERT INTO stock_movements (product_id, type, qty, note, created_at)
SELECT id, 'in', stock, 'Opening stock', now() - interval '3 days'
FROM products
WHERE stock > 0;

INSERT INTO sales_orders (customer, total, created_at) VALUES
    ('Toko Testing', 5 * 78500 + 10 * 38900, now() - interval '1 day'),
    ('Warung Testing', 20 * 17500, now() - interval '5 hours');

INSERT INTO sales_order_items (order_id, sku, qty, unit_price) VALUES
    (1, '001', 5, 78500),
    (1, '002', 10, 38900),
    (2, '003', 20, 17500);

INSERT INTO stock_movements (product_id, type, qty, note, created_at)
SELECT p.id, 'out', i.qty, 'Order #' || o.id, o.created_at
FROM sales_order_items i
JOIN sales_orders o ON o.id = i.order_id
JOIN products p ON p.sku = i.sku;

UPDATE products p
SET stock = p.stock - (SELECT COALESCE(SUM(i.qty), 0) FROM sales_order_items i WHERE i.sku = p.sku);
