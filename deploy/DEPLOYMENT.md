# Deploy BidKan backend ขึ้น Droplet

บน Droplet `kripnet` (159.223.37.4) รันแค่ API + Postgres + Redis + MQTT แบบเดียวกับ JodMenu
หน้าเว็บ (bidkan-backoffice, bidkan-client) อยู่บน Vercel (ท้ายไฟล์) ส่วน HTTPS ใช้ Caddy กลางของเครื่อง (`/opt/edge`)

```text
อินเทอร์เน็ต ──► edge-caddy :443 ──(network edge)──► bidkan-api:8080 ──► bidkan-db / bidkan-redis
ESP32 บนจักรยาน ──► 159.223.37.4:1883 ──► bidkan-mqtt ◄── bidkan-api
```

| บนเครื่อง | มาจาก |
|---|---|
| `/opt/bidkan/deploy.sh` | `deploy/deploy.sh` (อัปเดตตัวเองทุกครั้งที่ deploy) |
| `/opt/bidkan/docker-compose.prod.yml`, `bidkan.caddy`, `mosquitto.conf`, `schema/` | image `bidkan-deploy-files:<commit>` บน GHCR |
| `/opt/bidkan/.env.prod` | สร้างเองด้วยมือ จาก `deploy/.env.prod.example` |
| `/opt/bidkan/firebase-service-account.json` | ไฟล์ `configs/bidkan-service-account.json` ในเครื่อง dev |

Branch: พัฒนาบน `dev` ตลอด แล้ว merge `dev` เข้า `master` เมื่อจะขึ้น production (push เข้า `master` = deploy อัตโนมัติ)

## 1) DNS

เพิ่ม A record ของ `kriptnet.com`: `bidkan.api` → `159.223.37.4`

## 2) GitHub (repo Bangnus/Bidkan-backend)

Settings → Environments → New environment ชื่อ `production`
→ Deployment branches: Selected branches → เพิ่ม `master`
→ เพิ่ม secret 3 ตัวในข้อ 3.4 (`DROPLET_HOST`, `DROPLET_SSH_KEY`, `DROPLET_KNOWN_HOSTS`)

## 3) Droplet (ทำครั้งเดียว)

**3.1 โฟลเดอร์ + สคริปต์ deploy ตัวแรก** (ครั้งต่อไปสคริปต์อัปเดตตัวเอง)

```bash
mkdir -p /opt/bidkan
curl -fsSL https://raw.githubusercontent.com/Bangnus/Bidkan-backend/master/deploy/deploy.sh -o /opt/bidkan/deploy.sh
chmod 755 /opt/bidkan/deploy.sh
```

ถ้า repo เป็น private ให้ scp จากเครื่อง dev แทน: `scp deploy/deploy.sh root@159.223.37.4:/opt/bidkan/`

**3.2 ให้ Droplet ดึง image จาก GHCR ได้** — ถ้าเคย `docker login ghcr.io` ตอนทำ JodMenu ด้วยบัญชีที่เห็น package ของ Bangnus อยู่แล้ว ข้ามได้
ไม่งั้นสร้าง GitHub token แบบ classic ที่มีสิทธิ์ `read:packages` อย่างเดียว แล้ว:

```bash
docker login ghcr.io -u <ชื่อ GitHub>
```

**3.3 ไฟล์ลับ**

```bash
nano /opt/bidkan/.env.prod          # วางจาก deploy/.env.prod.example แล้วเติมค่า
chmod 600 /opt/bidkan/.env.prod
```

จากเครื่อง dev (PowerShell):

```powershell
scp D:\Bidkan-backend\configs\bidkan-service-account.json root@159.223.37.4:/opt/bidkan/firebase-service-account.json
```

แล้วบน Droplet (API รันเป็น uid 1000):

```bash
chown 1000:1000 /opt/bidkan/firebase-service-account.json
chmod 600 /opt/bidkan/firebase-service-account.json
```

**3.4 SSH key ของ GitHub Actions** (รันได้แค่ deploy.sh)

```bash
ssh-keygen -t ed25519 -N "" -C bidkan-ci -f /root/.ssh/bidkan_ci
echo "command=\"bash /opt/bidkan/deploy.sh\",no-pty,no-port-forwarding,no-agent-forwarding,no-X11-forwarding $(cat /root/.ssh/bidkan_ci.pub)" >> /root/.ssh/authorized_keys
cat /root/.ssh/bidkan_ci          # → secret DROPLET_SSH_KEY (ทั้งก้อน รวมบรรทัด BEGIN/END)
ssh-keyscan -t ed25519 159.223.37.4   # → secret DROPLET_KNOWN_HOSTS
rm /root/.ssh/bidkan_ci
```

`DROPLET_HOST` = `159.223.37.4` · private key วางได้ที่เดียวคือ secret ของ GitHub

**3.5 พอร์ต MQTT** — ถ้ามี DigitalOcean Cloud Firewall หรือ ufw ให้เปิด TCP `1883` ขาเข้า

## 4) Deploy ครั้งแรก

merge `dev` เข้า `master` แล้ว push → GitHub → Actions → "Deploy backend" ต้องเขียว
(หรือกด Run workflow บน branch `master`)

ฐานข้อมูลถูกสร้างจาก `sql/schema/*.sql` ตอน deploy แรก (volume ว่าง) เท่านั้น
ถ้าสร้างไม่สำเร็จ (`docker logs bidkan-db` มี ERROR) ฐานข้อมูลจะค้างครึ่ง ๆ ต้องลบ volume แล้ว deploy ใหม่ — ทำได้เฉพาะตอนยังไม่มีข้อมูลจริง:
`docker compose -p bidkan -f /opt/bidkan/docker-compose.prod.yml --env-file /opt/bidkan/.env.prod down` แล้ว `docker volume rm bidkan_pgdata`

แก้ schema ทีหลังต้องรันเองด้วยมือ:

```bash
docker exec -i bidkan-db psql -U bidkan -d bidkan < ไฟล์.sql
```

## 5) ตรวจหลัง deploy

```bash
docker compose -p bidkan -f /opt/bidkan/docker-compose.prod.yml --env-file /opt/bidkan/.env.prod ps
docker logs --tail 50 bidkan-api      # ต้องเห็น Connected to Redis, Firebase initialized
docker logs --tail 20 bidkan-mqtt
curl -I https://bidkan.api.kriptnet.com/api/v1/
```

## 6) กล่อง IoT (BidKan-iot)

`include/config.h`: `MQTT_SERVER "159.223.37.4"`, `MQTT_USER "device"`, `MQTT_PASS` = `MQTT_DEVICE_PASSWORD` ใน `.env.prod`
ตั้งแล้วแฟลชใหม่ กล่องที่ไม่มีรหัสจะต่อไม่ได้อีก (Mosquitto ไม่รับ anonymous)

## หน้าเว็บบน Vercel (bidkan-client, bidkan-backoffice)

ทั้งสองตัวเรียก API จากฝั่ง server ของ Vercel (route `/api/auth/*` และ `/api/proxy/*`) เบราว์เซอร์ไม่ได้เรียก API ตรง
cookie ล็อกอินเป็นของโดเมน Vercel เอง จึงไม่ติดปัญหา cookie ข้ามโดเมน

Vercel → Project → Settings → Environment Variables (Production) ใส่ทั้งสอง project:

| ตัวแปร | ค่า |
|---|---|
| `NEXT_PUBLIC_API_URL` | `https://bidkan.api.kriptnet.com/api/v1` |
| `API_INTERNAL_URL` | `https://bidkan.api.kriptnet.com/api/v1` |
| `JWT_SECRET` | **ค่าเดียวกับ** `JWT_SECRET` ใน `/opt/bidkan/.env.prod` (ไม่ตรง = ล็อกอินแล้วเด้งออก) |
| `PROXY_SHARED_SECRET` | **ค่าเดียวกับ** `PROXY_SHARED_SECRET` ใน `.env.prod` — ส่ง IP จริงของผู้ใช้ให้ API นับ rate limit (ไม่ตั้ง = ผู้ใช้ทุกคนใช้โควตาล็อกอิน/refresh ร่วมกัน แล้วเจอ 429 สุ่ม ๆ) |
| `NEXT_PUBLIC_APP_ENV` | `production` |
| `NEXT_PUBLIC_GOOGLE_CLIENT_ID` | (เฉพาะ bidkan-client) ค่าเดียวกับ `GOOGLE_CLIENT_ID` ของ backend |

`NEXT_PUBLIC_*` ถูกฝังตอน build: แก้แล้วต้อง Redeploy บน Vercel
`NEXT_PUBLIC_MQTT_WS_URL` ไม่ต้องตั้ง (โค้ดหน้าเว็บยังไม่ได้ใช้ และเซิร์ฟเวอร์ไม่เปิด MQTT แบบ WebSocket)

ฝั่ง backend: ใส่โดเมน production ของทั้งสอง project ใน `CORS_ALLOWED_ORIGINS` ของ `.env.prod`
(เช่น `https://bidkan-client.vercel.app,https://bidkan-backoffice.vercel.app` หรือโดเมนจริงที่ผูกไว้)
ให้ `STRIPE_CLIENT_URL` / `STRIPE_BACKOFFICE_URL` เป็นโดเมนเดียวกันนี้ (Stripe พากลับมาหน้านี้หลังจ่ายเงิน)
แก้ `.env.prod` แล้วต้องสั่งขึ้นใหม่: `docker compose -p bidkan -f /opt/bidkan/docker-compose.prod.yml --env-file /opt/bidkan/.env.prod up -d`
