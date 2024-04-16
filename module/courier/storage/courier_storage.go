package storage

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis"
	"gitlab.com/ptflp/geotask/module/courier/models"
)

type CourierStorager interface {
	Save(ctx context.Context, courier models.Courier) error // сохранить курьера по ключу courier
	GetOne(ctx context.Context) (*models.Courier, error)    // получить курьера по ключу courier
}

type CourierStorage struct {
	storage *redis.Client
}

func NewCourierStorage(storage *redis.Client) CourierStorager {
	return &CourierStorage{storage: storage}
}

func (cs *CourierStorage) Save(ctx context.Context, courier models.Courier) error {
	courierJSON, err := json.Marshal(courier)
	if err != nil {
		return err
	}

	if err = cs.storage.Set("courier", courierJSON, 0).Err(); err != nil {
		return err
	}
	return nil
}

func (cs *CourierStorage) GetOne(ctx context.Context) (*models.Courier, error) {
	courierJSON, err := cs.storage.Get("courier").Result()
	if err == redis.Nil {
		cs.Save(ctx, models.Courier{
			Location: models.Point{
				Lat: 59.9311,
				Lng: 30.3609,
			},
		})
	}
	if err != nil {
		return nil, err
	}
	var courier models.Courier
	err = json.Unmarshal([]byte(courierJSON), &courier)
	if err != nil {
		return nil, err
	}

	return &courier, nil
}
