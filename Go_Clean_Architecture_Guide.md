# 🚀 คู่มือ Go Clean Architecture (Production Ready - Pure SQL)

คู่มือนี้สรุปการสร้าง REST API ด้วย Go ที่เน้นประสิทธิภาพและความคลีนของโค้ด โดยใช้ SQL ดิบ (Pure SQL) เพื่อหลีกเลี่ยง Overhead ของ ORM และรองรับการทำระบบสเกลใหญ่

## 🛠️ Tech Stack สเปคใช้งานจริง

- **Framework:** `github.com/gofiber/fiber/v2` (เบาและเร็วมาก)
- **Database Toolkit:** `sqlc` + `database/sql` + `github.com/lib/pq` (เขียนไฟล์ SQL ดิบ แล้วให้ sqlc Gen โค้ด Go ที่ Type-safe ให้แบบอัตโนมัติ ไม่ต้องเขียน Struct รับค่าเอง)
- **Security:** `golang.org/x/crypto/bcrypt` (มาตรฐานการ Hash Password)
- **Validation:** `github.com/go-playground/validator/v10`
- **UUID:** `github.com/google/uuid`

---

## 📖 1. โครงสร้างโฟลเดอร์ (Folder Structure)

```text
my-clean-api/
├── cmd/
│   └── main.go                    # จุด Start ของระบบ (Initialize ทุกอย่างที่นี่)
├── internal/
│   └── app/
│       ├── domain/
│       │   ├── entity/            # Business Object (โครงสร้างข้อมูลดิบ)
│       │   └── repository/        # Interface ของ Database
│       ├── infrastructure/
│       │   ├── database/          # การเชื่อมต่อ PostgreSQL และ Connection Pool
│       │   └── repository/        # การเขียน SQL Query จริงๆ (Implementation)
│       ├── usecase/
│       │   └── user/
│       │       └── create/        # แยก Usecase ย่อยชัดเจน (Service, Handler)
│       ├── router/                # ผูก URL Routes เข้ากับ Handler
│       └── initial/               # Dependency Injection (ประกอบร่าง)
├── .env                           # เก็บความลับเช่น DB_PASSWORD
└── go.mod
```
