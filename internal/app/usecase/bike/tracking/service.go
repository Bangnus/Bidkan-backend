package tracking

import (
	"bidkan/internal/app/domain/entity"
	"context"
	"fmt"
)

type Service interface {
	ProcessTracking(ctx context.Context, data entity.BikeData) error
}

type service struct {
	// ในอนาคตจะใส่ BikeRepository ตรงนี้เพื่อเซฟลง DB
}

func NewService() Service {
	return &service{}
}

func (s *service) ProcessTracking(ctx context.Context, data entity.BikeData) error {
	// TODO: ใส่คำสั่งบันทึกลง PostgreSQL ผ่าน Repository ตรงนี้
	fmt.Printf("📍 [ได้รับข้อมูล] รถ: %s | พิกัด: %f, %f | แบต: %d%%\n", 
		data.BikeID, data.Lat, data.Lon, data.Battery)
	
	return nil
}
