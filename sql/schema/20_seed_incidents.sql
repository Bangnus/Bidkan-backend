-- Seed sample incident for tenant KU
INSERT INTO incident_reports (
    tenant_id, reporter_name, reporter_phone, incident_type, severity, description,
    location_name, location_lat, location_lon, police_report_no, insurance_claim_no,
    insurance_status, claim_amount, approved_amount, hospital_name, notes, created_at
)
SELECT id, 'สมชาย รัตนวงศ์ (ผู้ขับขี่)', '0812345678', 'fall', 'minor',
       'จักรยานลื่นทรายบริเวณโค้งหน้าคณะวิศวกรรมศาสตร์ มีแผลถลอกที่หัวเข่าและข้อศอก',
       'หน้าอาคารชูชาติ กำภู คณะวิศวกรรมศาสตร์', 13.847500, 100.570100, 'POL-69/042', 'CLM-2026-0081',
       'approved', 2500.00, 2500.00, 'โรงพยาบาลวิภาวดี', 'อนุมัติวงเงินประกันภัยกลุ่มอุบัติเหตุ นศ. เรียบร้อย', NOW() - INTERVAL '2 days'
FROM tenants WHERE slug = 'ku'
LIMIT 1;

INSERT INTO incident_reports (
    tenant_id, reporter_name, reporter_phone, incident_type, severity, description,
    location_name, location_lat, location_lon, police_report_no, insurance_claim_no,
    insurance_status, claim_amount, approved_amount, hospital_name, notes, created_at
)
SELECT id, 'วิภาวรรณ แสนสุข (คู่กรณี)', '0898765432', 'collision', 'moderate',
       'เฉี่ยวชนกับรถจักรยานยนต์บริเวณแยกสหกรณ์ มก. ข้อเท้าแพลงและแฮนด์รถจักรยานเบี้ยว',
       'สี่แยกสหกรณ์ มก.', 13.849200, 100.568300, 'POL-69/055', 'CLM-2026-0094',
       'under_review', 8500.00, 0.00, 'โรงพยาบาลเปาโล เกษตร', 'อยู่ระหว่างรอใบรับรองแพทย์และบันทึกประจำวันฉบับจริง', NOW() - INTERVAL '5 hours'
FROM tenants WHERE slug = 'ku'
LIMIT 1;
