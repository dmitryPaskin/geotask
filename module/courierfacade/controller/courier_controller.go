package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gitlab.com/ptflp/geotask/module/courierfacade/service"
	"golang.org/x/net/context"
	"log"
	"time"
)

type CourierController struct {
	courierService service.CourierFacer
}

func NewCourierController(courierService service.CourierFacer) *CourierController {
	return &CourierController{courierService: courierService}
}

func (c *CourierController) GetStatus(ctx *gin.Context) {
	// установить задержку в 50 миллисекунд
	time.Sleep(50 * time.Millisecond)
	// получить статус курьера из сервиса courierService используя метод GetStatus
	status := c.courierService.GetStatus(ctx)
	// отправить статус курьера в ответ
	ctx.JSON(200, gin.H{"status": status})
}

func (c *CourierController) MoveCourier(m webSocketMessage) {
	var cm CourierMove
	var err error
	// получить данные из m.Data и десериализовать их в структуру CourierMove
	switch data := m.Data.(type) {
	case string:
		err = json.Unmarshal([]byte(data), &cm)
	case []byte:
		err = json.Unmarshal(data, &cm)
	default:
		log.Println("Error: Unsupported data type in webSocketMessage")
		return
	}
	if err != nil {
		log.Println("Error deserializing courier move data:", err)
		return
	}
	// вызвать метод MoveCourier у courierService
	c.courierService.MoveCourier(context.TODO(), cm.Direction, cm.Zoom)
}
