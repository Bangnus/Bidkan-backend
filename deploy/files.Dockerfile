# ไฟล์ตั้งค่าที่ /opt/bidkan/deploy.sh บน Droplet ดึงไปวาง (GHCR: bidkan-deploy-files:<commit>)
# ส่งผ่าน GHCR แทน scp เพราะ key ของ GitHub Actions รันได้แค่ deploy.sh ส่งไฟล์เข้าเครื่องไม่ได้
# build จาก root ของ repo (ต้องใช้ sql/schema ด้วย)
FROM scratch
COPY deploy/docker-compose.prod.yml deploy/bidkan.caddy deploy/mosquitto.conf deploy/deploy.sh /
COPY sql/schema/ /schema/
