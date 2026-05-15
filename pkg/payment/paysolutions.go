package payment

import (
	"context"
	"fmt"
	"os"
)

type PaySolutionsService interface {
	GeneratePromptPayQR(ctx context.Context, orderID string, amount float64) (string, error)
}

type paySolutionsService struct {
	merchantID string
	apiKey     string
	baseURL    string
}

func NewPaySolutionsService() PaySolutionsService {
	// ดึงค่าจาก ENV (สามารถใส่ในไฟล์ .env ได้)
	merchantID := os.Getenv("PAYSOLUTIONS_MERCHANT_ID")
	apiKey := os.Getenv("PAYSOLUTIONS_API_KEY")
	baseURL := os.Getenv("PAYSOLUTIONS_BASE_URL")

	if baseURL == "" {
		baseURL = "https://apis.paysolutions.asia"
	}

	return &paySolutionsService{
		merchantID: merchantID,
		apiKey:     apiKey,
		baseURL:    baseURL,
	}
}

func (s *paySolutionsService) GeneratePromptPayQR(ctx context.Context, orderID string, amount float64) (string, error) {
	// TODO: นำโค้ดเชื่อมต่อ HTTP Request จริงมาใส่เมื่อได้ API Spec จาก PaySolutions
	// เนื่องจากตอนนี้ยังไม่มี API Spec แบบเป๊ะๆ จึงเขียนจำลองการเชื่อมต่อไว้ก่อน
	// โดยปกติ API จะคืนเป็น URL ของรูป QR Code หรือ Raw Base64 มาให้

	if s.merchantID == "" || s.apiKey == "" {
		fmt.Println("⚠️ Warning: PaySolutions API credentials not set. Returning mock QR.")
		return fmt.Sprintf("https://mock-qr-generator.com/?amount=%.2f&ref=%s", amount, orderID), nil
	}

	/*
		// ตัวอย่างโครงสร้างการยิง API ของ PaySolutions
		url := fmt.Sprintf("%s/openapi/v1/promptpay", s.baseURL)
		payload := map[string]interface{}{
			"merchant_id": s.merchantID,
			"order_no":    orderID,
			"amount":      amount,
			"detail":      "Wallet Top-up",
			"customer_email": "user@example.com",
		}
		
		// แปลง JSON และยิง HTTP POST ไปพร้อม Header (Authorization: Bearer <ApiKey>)
		// คืนค่า QR Code URL ที่ได้จาก Response
	*/

	return fmt.Sprintf("https://mock-qr-generator.com/?amount=%.2f&ref=%s", amount, orderID), nil
}
