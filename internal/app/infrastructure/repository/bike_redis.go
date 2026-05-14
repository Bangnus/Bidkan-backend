package repository

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/repository"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type bikeRedisRepository struct {
	rdb *redis.Client
}

func NewBikeRedisRepository(rdb *redis.Client) repository.BikeCacheRepository {
	return &bikeRedisRepository{rdb: rdb}
}

func (r *bikeRedisRepository) SetLatestLocation(ctx context.Context, data entity.BikeData) error {
	key := fmt.Sprintf("bike:location:%s", data.BikeID)
	
	// แปลงข้อมูลเป็น JSON เพื่อเก็บใน Redis
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// เก็บข้อมูลโดยตั้งค่า Expire ไว้ 24 ชั่วโมง (ถ้าไม่ส่งสัญญาณมาเลยให้ลบออก)
	return r.rdb.Set(ctx, key, payload, 24*time.Hour).Err()
}

func (r *bikeRedisRepository) GetLatestLocation(ctx context.Context, bikeID string) (*entity.BikeData, error) {
	key := fmt.Sprintf("bike:location:%s", bikeID)
	
	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var data entity.BikeData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}

	return &data, nil
}
