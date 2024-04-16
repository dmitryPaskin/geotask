package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.com/ptflp/geotask/module/order/models"
	"strconv"
	"time"
)

type OrderStorager interface {
	Save(ctx context.Context, order models.Order, maxAge time.Duration) error                       // сохранить заказ с временем жизни
	GetByID(ctx context.Context, orderID int) (*models.Order, error)                                // получить заказ по id
	GenerateUniqueID(ctx context.Context) (int64, error)                                            // сгенерировать уникальный id
	GetByRadius(ctx context.Context, lng, lat, radius float64, unit string) ([]models.Order, error) // получить заказы в радиусе от точки
	GetCount(ctx context.Context) (int, error)                                                      // получить количество заказов
	RemoveOldOrders(ctx context.Context, maxAge time.Duration) error                                // удалить старые заказы по истечению времени maxAge
}

type OrderStorage struct {
	storage *redis.Client
}

func NewOrderStorage(storage *redis.Client) OrderStorager {
	return &OrderStorage{storage: storage}
}

func (o *OrderStorage) Save(ctx context.Context, order models.Order, maxAge time.Duration) error {
	// save with geo redis
	if order.ID == 0 {
		return errors.New("ID is zero")
	}
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return err
	}

	if err := o.storage.Set(fmt.Sprintf("order:%d", order.ID), string(orderJSON), maxAge).Err(); err != nil {
		return err
	}

	orderGEOAddCmd := o.storage.GeoAdd("orders", &redis.GeoLocation{
		Name:      fmt.Sprintf("order:%d", order.ID),
		Longitude: order.Lng,
		Latitude:  order.Lat,
	})

	if orderGEOAddCmd.Err() != nil {
		return orderGEOAddCmd.Err()
	}
	return nil
}

func (o *OrderStorage) RemoveOldOrders(ctx context.Context, maxAge time.Duration) error {
	// получить ID всех старых ордеров, которые нужно удалить
	// используя метод ZRangeByScore
	// старые ордеры это те, которые были созданы две минуты назад
	// и более
	/**
	&redis.ZRangeBy{
		Max: использовать вычисление времени с помощью maxAge,
		Min: "0",
	}
	*/

	// Проверить количество старых ордеров
	// удалить старые ордеры из redis используя метод ZRemRangeByScore где ключ "orders" min "-inf" max "(время создания старого ордера)"
	// удалять ордера по ключу не нужно, они будут удалены автоматически по истечению времени жизни
	maxTime := time.Now().Add(-maxAge).Unix()

	oldOrderIDs, err := o.storage.ZRangeByScore("orders", redis.ZRangeBy{
		Max: fmt.Sprintf("(%d", maxTime),
		Min: "0",
	}).Result()
	if err != nil {
		return err
	}

	if len(oldOrderIDs) == 0 {
		return nil
	}

	var oldOrderIDsInterface []interface{}
	for _, id := range oldOrderIDs {
		oldOrderIDsInterface = append(oldOrderIDsInterface, id)
	}

	_, err = o.storage.ZRem("orders", oldOrderIDsInterface...).Result()
	if err != nil {
		return err
	}

	for _, orderID := range oldOrderIDs {
		err = o.storage.Del(fmt.Sprintf("order:%s", orderID)).Err()
		if err != nil {
			fmt.Printf("failed to remove data of old order %s: %v\n", orderID, err)
		}
	}

	return nil
}

func (o *OrderStorage) GetByID(ctx context.Context, orderID int) (*models.Order, error) {

	// получаем ордер из redis по ключу order:ID

	// проверяем что ордер не найден исключение redis.Nil, в этом случае возвращаем nil, nil

	// десериализуем ордер из json
	data, err := o.storage.Get(fmt.Sprintf("order:%d", orderID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var order models.Order
	if err = json.Unmarshal(data, &order); err != nil {
		return nil, err
	}

	return &order, nil
}

func (o *OrderStorage) GetCount(ctx context.Context) (int, error) {
	// получить количество ордеров в упорядоченном множестве используя метод ZCard
	count, err := o.storage.ZCard("orders").Result()
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (o *OrderStorage) GetByRadius(ctx context.Context, lng, lat, radius float64, unit string) ([]models.Order, error) {
	ordersLocation, err := o.storage.GeoRadius("orders", lng, lat, &redis.GeoRadiusQuery{
		Radius:      radius,
		Unit:        unit,
		WithCoord:   true,
		WithDist:    true,
		WithGeoHash: false,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to get orders by radius: %w", err)
	}

	orders := make([]models.Order, len(ordersLocation))

	for _, orderLocation := range ordersLocation {
		orderID, err := strconv.Atoi(orderLocation.Name)
		if err != nil {
			fmt.Printf("failed to parse order ID: %v\n", err)
			continue
		}

		order, err := o.GetByID(ctx, orderID)
		if err != nil {
			fmt.Printf("failed to get order by ID %d: %v\n", orderID, err)
			continue
		}

		orders = append(orders, *order)
	}

	return orders, nil
}

func (o *OrderStorage) GenerateUniqueID(ctx context.Context) (int64, error) {
	id, err := o.storage.Incr("order:id").Result()
	if err != nil {
		return 0, err
	}
	return id, nil
}
