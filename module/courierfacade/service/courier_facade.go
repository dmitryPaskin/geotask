package service

import (
	"context"
	cservice "gitlab.com/ptflp/geotask/module/courier/service"
	cfm "gitlab.com/ptflp/geotask/module/courierfacade/models"
	oservice "gitlab.com/ptflp/geotask/module/order/service"
	"log"
)

const (
	CourierVisibilityRadius = 2800 // 2.8km
)

type CourierFacer interface {
	MoveCourier(ctx context.Context, direction, zoom int) // отвечает за движение курьера по карте direction - направление движения, zoom - уровень зума
	GetStatus(ctx context.Context) cfm.CourierStatus      // отвечает за получение статуса курьера и заказов вокруг него
}

// CourierFacade фасад для курьера и заказов вокруг него (для фронта)
type CourierFacade struct {
	courierService cservice.Courierer
	orderService   oservice.Orderer
}

func NewCourierFacade(courierService cservice.Courierer, orderService oservice.Orderer) CourierFacer {
	return &CourierFacade{courierService: courierService, orderService: orderService}
}

func (cf *CourierFacade) MoveCourier(ctx context.Context, direction, zoom int) {
	courier, err := cf.courierService.GetCourier(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}
	if err = cf.courierService.MoveCourier(*courier, direction, zoom); err != nil {
		log.Fatal(err)
		return
	}
}

func (cf *CourierFacade) GetStatus(ctx context.Context) cfm.CourierStatus {
	courier, err := cf.courierService.GetCourier(ctx)
	if err != nil {
		log.Fatal(err)
		return cfm.CourierStatus{} // Возвращаем пустой статус
	}

	orders, err := cf.orderService.GetByRadius(ctx, courier.Location.Lng, courier.Location.Lat, CourierVisibilityRadius, "km")
	if err != nil {
		log.Println(err)
		return cfm.CourierStatus{Courier: *courier}
	}

	courierStatus := cfm.CourierStatus{
		Courier: *courier,
		Orders:  orders,
	}

	return courierStatus
}
