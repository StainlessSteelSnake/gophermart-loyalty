package orders

import (
	"encoding/json"
	"errors"
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/database"
	"io"
	"log"
	"strconv"
	"time"
)

type OrderBonuses struct {
	ID          string `json:"order"`
	Status      string `json:"status"`
	BonusAmount int    `json:"accrual"`
}

func (o *orderController) initOrderProcessing(channelCount int) {
	o.processingChannels = make([]chan *Order, 0, channelCount)

	for i := 0; i < channelCount; i++ {
		ch := make(chan *Order)
		o.processingChannels = append(o.processingChannels, ch)
		go o.processOrdersInChannel(o.processingChannels[i])
	}

	go o.processErrors()
	go o.getOrdersToProcess()
	go o.processOrdersToSave()

	go func() {
		defer o.closeProcessingChannels()

		for i := 0; ; i++ {
			if i == len(o.processingChannels) {
				i = 0
			}

			// Получение заказа на обработку из общей очереди и отсылка его на обработку в один из каналов
			order, ok := <-o.ordersToProcess
			if !ok {
				return
			}

			ch := o.processingChannels[i]
			ch <- order
		}
	}()
}

func (o *orderController) getOrdersToProcess() {
	select {
	case <-o.done:
		return
	default:

	}

	orders, err := o.model.GetOrdersToProcess()
	log.Printf("Найдено %v заказов для обработки\n", len(orders))

	if err != nil {
		o.errors <- err
		time.AfterFunc(time.Second*1, func() { o.getOrdersToProcess() })
		return
	}

	for _, order := range orders {
		orderToProcess := Order{
			ID:         order.ID,
			UserLogin:  order.UserLogin,
			Status:     order.Status,
			UploadedAt: order.UploadedAt,
		}

		o.ordersToProcess <- &orderToProcess
	}

	time.AfterFunc(time.Second*1, func() { o.getOrdersToProcess() })
}

// Отслеживание отдельного канала на обработку заказов
func (o *orderController) processOrdersInChannel(processingChannel <-chan *Order) {
	for order := range processingChannel {
		o.processOrder(order)
	}
}

// Чтение ошибок из канала errors и вывод их в лог программы
func (o *orderController) processErrors() {
	for err := range o.errors {
		log.Println(err)
	}
}

// Закрытие всех каналов на обработку заказов в случае закрытия общего канала ordersToProcess
func (o *orderController) closeProcessingChannels() {
	for _, ch := range o.processingChannels {
		close(ch)
	}
}

// Закрытие всех каналов при завершении работы main()
func (o *orderController) Close() {
	close(o.done)
	close(o.ordersToProcess)
	close(o.ordersToSave)
	close(o.errors)
}

func (o *orderController) processOrder(order *Order) {

	response, err := o.client.Get(o.accrualSystemAddress + "/api/orders/" + order.ID)
	if err != nil {
		o.errors <- err
		return
	}

	defer response.Body.Close()

	switch response.StatusCode {
	case 200:

	case 204:
		o.errors <- errors.New("заказ " + order.ID + " не зарегистрирован в системе")
		return

	case 429:
		retry := response.Header.Get("Retry-After")

		retryAfter, err := strconv.Atoi(retry)
		if err != nil {
			o.errors <- errors.New("превышено количество запросов к сервису: " + response.Status + ", некорректный заголовок Retry-After: " + retry)
			return
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			o.postponeProcessing(time.Second * time.Duration(retryAfter))
			o.errors <- errors.New("превышено количество запросов к сервису: " + response.Status + ". " + err.Error())
			return
		}

		o.postponeProcessing(time.Second * time.Duration(retryAfter))
		o.errors <- errors.New("превышено количество запросов к сервису: " + response.Status + ". " + string(body))
		return

	case 500:
		o.errors <- errors.New("ошибка сервера рассчёта баллов лояльности")
		return

	default:
		o.errors <- errors.New("неизвестная ошибка, код " + strconv.Itoa(response.StatusCode) + ", описание: " + response.Status)
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		o.errors <- err
		return
	}

	var orderBonuses OrderBonuses
	err = json.Unmarshal(body, &orderBonuses)
	if err != nil {
		o.errors <- err
		return
	}

	o.errors <- errors.New("Получены статус " + orderBonuses.Status + " по заказу " + orderBonuses.ID + ". Кол-во начисленных бонусов: " + strconv.Itoa(orderBonuses.BonusAmount) + ".")
	if orderBonuses.Status != orderStatusNew {
		orderToSave := &Order{
			ID:         orderBonuses.ID,
			UserLogin:  order.UserLogin,
			Status:     orderBonuses.Status,
			UploadedAt: order.UploadedAt,
			Amount:     orderBonuses.BonusAmount,
		}
		o.ordersToSave <- orderToSave
	}
}

func (o *orderController) processOrdersToSave() {
	for orderToSave := range o.ordersToSave {
		order := database.Order{
			ID:         orderToSave.ID,
			UserLogin:  orderToSave.UserLogin,
			Status:     orderToSave.Status,
			UploadedAt: orderToSave.UploadedAt,
		}

		err := o.model.UpdateOrder(&order, orderToSave.Amount)
		if err != nil {
			o.errors <- err
		}
	}
}

func (o *orderController) postponeProcessing(seconds time.Duration) {

}
