# 🚀 BidKan Platform — DigitalOcean Droplet Deployment Guide

คู่มือฉบับสมบูรณ์สำหรับการ Deploy แพลตฟอร์ม **BidKan (บิดกัน)** ขึ้นบน **DigitalOcean Droplet (Ubuntu 22.04 / 24.04 LTS)** ครอบคลุมทั้ง Backend, Backoffice, Client, Database, MQTT, Redis และ Nginx SSL Gateway

---

## 💻 1. สเปก Droplet ที่แนะนำ
- **OS:** Ubuntu 22.04 LTS หรือ 24.04 LTS x64
- **Plan:** Basic Droplet (Regular หรือ Premium Intel/AMD)
  - **แนะนำเริ่มต้น:** 2 vCPU / 2GB RAM / 50GB SSD ($12 - $18/เดือน) + สคริปต์จะสร้าง 2GB Swap Memory ให้อัตโนมัติ ป้องกัน RAM เต็มตอน Build Next.js
  - **เพื่อประสิทธิภาพสูงสุด:** 2 vCPU / 4GB RAM ($24/เดือน)
- **Datacenter:** Singapore (SGP1) เพื่อให้ Latency ต่ำที่สุดสำหรับประเทศไทย (ประมาณ 25-35 ms)

---

## 🌐 2. การตั้งค่า DNS Records (ชี้มาที่ IP ของ Droplet)
ไปที่ผู้ให้บริการโดเมนของคุณ (เช่น Cloudflare, GoDaddy, Namecheap) แล้วเพิ่ม A Records ชี้มาที่ Droplet IP:

| Type | Name | Content / Value | TTL |
|---|---|---|---|
| A | `@` (หรือ `bidkan.com`) | `YOUR_DROPLET_IP` | Auto / DNS Only |
| A | `app` (Client PWA) | `YOUR_DROPLET_IP` | Auto / DNS Only |
| A | `admin` (Backoffice Console) | `YOUR_DROPLET_IP` | Auto / DNS Only |
| A | `api` (Go Backend REST API) | `YOUR_DROPLET_IP` | Auto / DNS Only |
| A | `mqtt` (WebSocket & IoT Broker) | `YOUR_DROPLET_IP` | Auto / DNS Only |

---

## 🛠️ 3. ขั้นตอนการ Deploy บน Droplet

### Step 3.1: SSH เข้าไปยัง Droplet
```bash
ssh root@YOUR_DROPLET_IP
```

### Step 3.2: Clone โปรเจกต์ลงบน Droplet
```bash
# สร้างโฟลเดอร์หลักของโปรเจกต์
mkdir -p /var/www && cd /var/www

# Clone โปรเจกต์ (หรือดึงจาก Git Repository ของคุณ)
git clone https://github.com/Bangnus/Bidkan-backend.git
git clone https://github.com/Bangnus/bidkan-backoffice.git
git clone https://github.com/Bangnus/bidkan-client.git
```

### Step 3.3: สร้างไฟล์ Configuration `.env.production`
```bash
cd /var/www/Bidkan-backend/deploy
cp .env.production.example .env.production
nano .env.production
```
*แก้ไขค่าตัวแปร เช่น `DOMAIN_NAME`, `JWT_SECRET`, รหัสผ่านฐานข้อมูล, และ API Keys ให้ตรงกับของจริง*

### Step 3.4: รันสคริปต์ติดตั้งและเริ่มระบบอัตโนมัติ
```bash
bash deploy-droplet.sh
```
สคริปต์จะดำเนินการให้อัตโนมัติ:
- ✅ สร้าง 2GB Swap Memory ป้องกันหน่วยความจำเต็ม
- ✅ ติดตั้ง Docker Engine และ Docker Compose Plugin
- ✅ ตั้งค่าไฟร์วอลล์ UFW เปิดพอร์ต `22` (SSH), `80` (HTTP), `443` (HTTPS), และ `1883` (IoT MQTT)
- ✅ สั่ง Build และรันคอนเทนเนอร์ทั้งหมดผ่าน `docker-compose.prod.yml`

---

## 🔒 4. การออกใบรับรองความปลอดภัยฟรี (Let's Encrypt SSL)

เมื่อคอนเทนเนอร์ทำงานและ DNS อัปเดตเรียบร้อย ให้รันคำสั่งออก SSL Certificate:

```bash
cd /var/www/Bidkan-backend/deploy

# ออกใบรับรองสำหรับทุก Subdomain
docker compose -f docker-compose.prod.yml run --rm certbot certonly --webroot \
  -w /var/www/certbot \
  --email admin@bidkan.com \
  --agree-tos \
  --no-eff-email \
  -d bidkan.com \
  -d app.bidkan.com \
  -d admin.bidkan.com \
  -d api.bidkan.com \
  -d mqtt.bidkan.com

# รีโหลด Nginx เพื่อเริ่มใช้ SSL
docker compose -f docker-compose.prod.yml exec nginx nginx -s reload
```

---

## 📊 5. คำสั่งที่ใช้ดูแลรักษาระบบ (Operations & Monitoring)

```bash
# ตรวจสอบสถานะคอนเทนเนอร์ทั้งหมด
docker compose -f docker-compose.prod.yml ps

# ดู Live Logs ของทุกระบบ
docker compose -f docker-compose.prod.yml logs -f

# ดู Logs เฉพาะ Go Backend
docker compose -f docker-compose.prod.yml logs -f backend

# อัปเดตโค้ดเวอร์ชันใหม่
git pull
docker compose -f docker-compose.prod.yml up -d --build
```
