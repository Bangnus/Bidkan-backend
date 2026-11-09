# 🚲 Project Context & Agent Instructions: "BidKan" (บิดกัน) - Backend

คู่มือและคำแนะนำสำหรับ Agent ในโปรเจกต์ **Bidkan Backend** (RESTful API & MQTT Broker สำหรับระบบสมาร์ทอีไบค์แชร์ริ่ง BidKan)

---

## 🛠️ Tech Stack
- **Language:** Go (1.23+)
- **Framework:** `github.com/gofiber/fiber/v2`
- **Database Toolkit:** `sqlc` + `database/sql` + `github.com/lib/pq` (Pure SQL ไม่ใช้ ORM)
- **Database:** PostgreSQL (PostGIS สำหรับระบบพิกัดจักรยาน/โซน)
- **Message Broker:** Eclipse Mosquitto (MQTT สำหรับตำแหน่งจักรยานและ IoT ESP32)
- **Hot Reload:** Air (`.air.toml`)

---

## 🧠 Core Business Logic
- **Geofencing System:** ตรวจสอบว่าพิกัดรถอยู่ใน `Parking Zone (P)` ก่อนอนุญาตให้จบการเดินทาง และตรวจจับการออกนอก `Riding Limit Zone` เพื่อสั่งตัดไฟ Relay
- **Real-time Wallet Deduction & Auto-Cutoff:** หักเงินจากกระเป๋าตามเวลาขับขี่จริง และสั่งตัดไฟผ่าน MQTT ทันทีเมื่อยอดเงินหมด (`wallet_balance <= 0`)
- **Ranking & Loyalty System:** คำนวณระดับแรงค์ย้อนหลัง 3 เดือนเพื่อให้ส่วนลดอัตราค่าบริการรายนาที

---

## 🏛️ Architecture Rules (Pure SQL Go Clean Architecture)
โปรเจกต์นี้ใช้ Go Clean Architecture แยกชั้นการทำงานอย่างเคร่งครัด:

```text
internal/app/
├── domain/                    # Enterprise Business Rules
│   ├── entity/                # Data Structs บริสุทธิ์
│   └── repository/            # Repository Interfaces สำหรับ Database
├── infrastructure/            # Implementations
│   ├── database/              # PostgreSQL Connection & Pool
│   └── repository/            # SQL Execution ผ่าน sqlc Queries
├── usecase/                   # Application Business Logic
│   └── [feature]/             # แยกตาม Feature ย่อย (Service & Handler)
├── router/                    # Fiber Route Declarations
└── middleware/                # Auth, CORS, Logger, Validator
```

### ⚠️ กฎความยาวไฟล์และการจัดโครงสร้างโค้ด (Code Organization Rules)
- **📏 แต่ละไฟล์ห้ามเกิน 200 บรรทัด (Max 200 Lines per File):** หากไฟล์ใดมีความยาวเกิน 200 บรรทัด ต้องแยกออกเป็นไฟล์ย่อยตามหน้าที่หรือฟีเจอร์ทันที เพื่อให้โค้ดอ่านง่ายและดูแลรักษาง่าย
- **🚫 กฎการจัดการข้อมูลจำลองและค่าเริ่มต้น (No Mock / Hardcoded Fallback Rule):** ห้ามใส่ข้อมูลจำลองหรือเบอร์โทรศัพท์ปลอม (เช่น `"065-458-5254"`) เด็ดขาด หากไม่มีข้อมูลหรือข้อมูลเป็น null / undefined / ค่าว่าง ให้แสดงผลเป็น `"-"` เสมอ เพื่อความถูกต้องและโปร่งใสของข้อมูลจริง

---

## 🚀 Common Commands
```bash
# Start Development with Live Reload
air

# Run SQLC Generation (หลังจากแก้ไฟล์ใน sql/queries หรือ sql/schema)
sqlc generate

# Docker Compose Services
docker compose up -d
```

