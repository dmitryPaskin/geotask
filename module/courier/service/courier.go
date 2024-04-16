package service

import (
	"context"
	"errors"
	"gitlab.com/ptflp/geotask/geo"
	"gitlab.com/ptflp/geotask/module/courier/models"
	"gitlab.com/ptflp/geotask/module/courier/storage"
	"math"
)

// Направления движения курьера
const (
	DirectionUp    = 0
	DirectionDown  = 1
	DirectionLeft  = 2
	DirectionRight = 3
)

const (
	DefaultCourierLat = 59.9311
	DefaultCourierLng = 30.3609
)

type Courierer interface {
	GetCourier(ctx context.Context) (*models.Courier, error)
	MoveCourier(courier models.Courier, direction, zoom int) error
}

type CourierService struct {
	courierStorage storage.CourierStorager
	allowedZone    geo.PolygonChecker
	disabledZones  []geo.PolygonChecker
}

func NewCourierService(courierStorage storage.CourierStorager, allowedZone geo.PolygonChecker, disbledZones []geo.PolygonChecker) Courierer {
	return &CourierService{courierStorage: courierStorage, allowedZone: allowedZone, disabledZones: disbledZones}
}

func (c *CourierService) GetCourier(ctx context.Context) (*models.Courier, error) {
	// получаем курьера из хранилища используя метод GetOne из storage/courier.go

	// проверяем, что курьер находится в разрешенной зоне
	// если нет, то перемещаем его в случайную точку в разрешенной зоне
	// сохраняем новые координаты курьера

	courier, err := c.courierStorage.GetOne(ctx)
	if err != nil {
		return nil, err
	}
	if !c.allowedZone.Contains(geo.Point(courier.Location)) {
		newLocation := c.allowedZone.RandomPoint()
		courier.Location.Lat = newLocation.Lat
		courier.Location.Lng = newLocation.Lng

		if err := c.courierStorage.Save(ctx, *courier); err != nil {
			return nil, err
		}
	}

	return courier, nil
}

// MoveCourier : direction - направление движения курьера, zoom - зум карты
func (c *CourierService) MoveCourier(courier models.Courier, direction, zoom int) error {
	// точность перемещения зависит от зума карты использовать формулу 0.001 / 2^(zoom - 14)
	// 14 - это максимальный зум карты

	// далее нужно проверить, что курьер не вышел за границы зоны
	// если вышел, то нужно переместить его в случайную точку внутри зоны

	// далее сохранить изменения в хранилище
	precision := 0.001 / math.Pow(2, float64(zoom-14))

	switch direction {
	case DirectionUp:
		courier.Location.Lat += precision
	case DirectionDown:
		courier.Location.Lat -= precision
	case DirectionLeft:
		courier.Location.Lng -= precision
	case DirectionRight:
		courier.Location.Lng += precision
	default:
		return errors.New("unknown direction")
	}

	courierLocation := geo.Point{Lat: courier.Location.Lat, Lng: courier.Location.Lng}
	if !c.allowedZone.Contains(courierLocation) {
		randomPoint := c.allowedZone.RandomPoint()
		courier.Location.Lat = randomPoint.Lat
		courier.Location.Lng = randomPoint.Lng
	}

	// Сохраняем изменения в хранилище
	err := c.courierStorage.Save(context.Background(), courier)
	if err != nil {
		return err
	}

	return nil
}
