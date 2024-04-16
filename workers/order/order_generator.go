package order

import (
	"gitlab.com/ptflp/geotask/module/order/service"
	"golang.org/x/net/context"
	"log"
	"time"
)

const (
	// order generation interval
	orderGenerationInterval = 10 * time.Millisecond
	maxOrdersCount          = 200
)

// worker generates orders and put them into redis
type OrderGenerator struct {
	orderService service.Orderer
}

func NewOrderGenerator(orderService service.Orderer) *OrderGenerator {
	return &OrderGenerator{orderService: orderService}
}

func (o *OrderGenerator) Run(ctx context.Context) {
	// запускаем горутину, которая будет генерировать заказы не более чем раз в 10 миллисекунд
	// не более 200 заказов используя константы orderGenerationInterval и maxOrdersCount
	// нужно использовать метод orderService.GetCount() для получения количества заказов
	// и метод orderService.GenerateOrder() для генерации заказа
	// если количество заказов меньше maxOrdersCount, то нужно сгенерировать новый заказ
	// если количество заказов больше или равно maxOrdersCount, то не нужно ничего делать
	// если при генерации заказа произошла ошибка, то нужно вывести ее в лог
	// если при получении количества заказов произошла ошибка, то нужно вывести ее в лог
	// внутри горутины нужно использовать select и time.NewTicker()
	go func() {
		// Создаем таймер, который будет срабатывать с интервалом orderGenerationInterval
		ticker := time.NewTicker(orderGenerationInterval)
		defer ticker.Stop()

		// Главный цикл генерации заказов
		for {
			select {
			case <-ticker.C:
				// Получаем текущее количество заказов
				count, err := o.orderService.GetCount(ctx)
				if err != nil {
					log.Printf("Error getting order count: %v", err)
					continue
				}

				// Если количество заказов меньше maxOrdersCount, генерируем новый заказ
				if count < maxOrdersCount {
					err = o.orderService.GenerateOrder(ctx)
					if err != nil {
						log.Printf("Error generating order: %v", err)
					}
				} else {
					time.Sleep(time.Second * 15)
				}
			}
		}
	}()

}
