-- Seed initial spare parts for each tenant
INSERT INTO campus_spare_parts (tenant_id, part_name, part_code, category, quantity, unit, min_stock, cost_per_unit)
SELECT id, 'แบตเตอรี่ลิเธียม 48V 15Ah', 'BAT-48V-15AH', 'battery', 12, 'ลูก', 5, 4500.00
FROM tenants
ON CONFLICT (tenant_id, part_code) DO NOTHING;

INSERT INTO campus_spare_parts (tenant_id, part_name, part_code, category, quantity, unit, min_stock, cost_per_unit)
SELECT id, 'ผ้าเบรกดิสก์หน้า-หลัง', 'BRK-PAD-D01', 'brake', 24, 'ชุด', 10, 150.00
FROM tenants
ON CONFLICT (tenant_id, part_code) DO NOTHING;

INSERT INTO campus_spare_parts (tenant_id, part_name, part_code, category, quantity, unit, min_stock, cost_per_unit)
SELECT id, 'ยางนอกเรเดียล 10 นิ้ว', 'TIRE-RAD-10', 'tire', 8, 'เส้น', 6, 380.00
FROM tenants
ON CONFLICT (tenant_id, part_code) DO NOTHING;

INSERT INTO campus_spare_parts (tenant_id, part_name, part_code, category, quantity, unit, min_stock, cost_per_unit)
SELECT id, 'กล่องคอนโทรลเลอร์ IoT 4G', 'IOT-CTL-4G', 'electrical', 4, 'กล่อง', 5, 1200.00
FROM tenants
ON CONFLICT (tenant_id, part_code) DO NOTHING;

INSERT INTO campus_spare_parts (tenant_id, part_name, part_code, category, quantity, unit, min_stock, cost_per_unit)
SELECT id, 'มือเบรกพร้อมสวิตช์ตัดไฟ', 'BRK-LEV-SW', 'body', 10, 'คู่', 4, 250.00
FROM tenants
ON CONFLICT (tenant_id, part_code) DO NOTHING;
