package orders

/*
func (o *orderController) processOrder(order *Order) {
	o.retryMutex.Lock()
	if o.waitForRetry {
		o.retryAfter.Wait()
	}
	o.retryMutex.Unlock()

	response, err := o.client.Get(o.accrualSystemAddress + "/api/orders/" + order.ID)
	if err != nil {
		go o.addOrderToProcess(order, err)
		return
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case 200:

	case 204:
		go o.addOrderToProcess(order, errors.New("заказ "+order.ID+" не зарегистрирован в системе"))
		return

	case 429:
		retry := response.Header.Get("Retry-After")

		retryAfter, err := strconv.Atoi(retry)
		if err != nil {
			go o.addOrderToProcess(order, errors.New("превышено количество запросов к сервису: "+response.Status+", некорректный заголовок Retry-After: "+retry))
			return
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			o.postponeProcessing(time.Second * time.Duration(retryAfter))
			go o.addOrderToProcess(order, errors.New("превышено количество запросов к сервису: "+response.Status+". "+err.Error()))
			return
		}

		o.postponeProcessing(time.Second * time.Duration(retryAfter))
		go o.addOrderToProcess(order, errors.New("превышено количество запросов к сервису: "+response.Status+". "+string(body)))
		return

	case 500:
		go o.addOrderToProcess(order, errors.New("ошибка сервера рассчёта баллов лояльности"))
		return

	default:
		go o.addOrderToProcess(order, errors.New("неизвестная ошибка, код "+strconv.Itoa(response.StatusCode)+", описание: "+response.Status))
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		go o.addOrderToProcess(order, err)
		return
	}

	var orderBonuses OrderBonuses
	err = json.Unmarshal(body, &orderBonuses)
	if err != nil {
		go o.addOrderToProcess(order, err)
		return
	}

	order.Amount = orderBonuses.BonusAmount
	order.Status = orderBonuses.Status
	log.Println("Получены статус " + order.Status + " по заказу " + order.ID + " для пользователя " + order.UserLogin + ". Кол-во начисленных бонусов: " + strconv.Itoa(order.Amount) + ".")

	go o.addOrderToSave(order)

}

*/

/*
func (o *orderController) addOrderToProcess(order *Order, err error) {
	o.ordersToProcess <- order

	if err == nil {
		return
	}
	o.errors <- err
}
*/

/*
func (o *orderController) postponeProcessing(delay time.Duration) {
	o.retryMutex.Lock()
	defer o.retryMutex.Unlock()

	o.waitForRetry = true
	time.Sleep(delay)
	o.waitForRetry = false
	o.retryAfter.Broadcast()
}
*/

/*
func (o *orderController) ProcessOrder(orderID string) {

	response, err := o.client.Get("http://localhost:8080/api/orders/" + orderID)
	if err != nil {
		log.Println(err)
		return
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case 200:

	case 204:
		log.Println("заказ " + orderID + " не зарегистрирован в системе")
		return

	case 429:
		retry := response.Header.Get("Retry-After")

		retryAfter, err := strconv.Atoi(retry)
		if err != nil {
			log.Println("превышено количество запросов к сервису: " + response.Status + ", некорректный заголовок Retry-After: " + retry)
			return
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			o.postponeProcessing(time.Second * time.Duration(retryAfter))
			log.Println("превышено количество запросов к сервису: " + response.Status + ". " + err.Error())
			return
		}

		log.Println("превышено количество запросов к сервису: " + response.Status + ". " + string(body))
		return

	case 500:
		log.Println("ошибка сервера рассчёта баллов лояльности")
		return

	default:
		log.Println("неизвестная ошибка, код " + strconv.Itoa(response.StatusCode) + ", описание: " + response.Status)
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var orderBonuses OrderBonuses
	err = json.Unmarshal(body, &orderBonuses)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Получены статус " + orderBonuses.Status + " по заказу " + orderBonuses.ID + ". Кол-во начисленных бонусов: " + strconv.Itoa(orderBonuses.BonusAmount) + ".")
}
*/
