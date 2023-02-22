package orders

import (
	"errors"
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/database"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const orderBufferSize = 10

type Order struct {
	ID         string
	UserLogin  string
	Status     string
	UploadedAt time.Time
}

type OrderAdderGetter interface {
	AddOrder(user, order string) error
	GetOrders(user string) ([]database.Order, error)
}

type orderController struct {
	ordersToProcess chan Order
	model           OrderAdderGetter
	client          http.Client
}

func NewOrders(m OrderAdderGetter) OrderAdderGetter {
	return &orderController{model: m, client: http.Client{}, ordersToProcess: make(chan Order, orderBufferSize)}
}

func lunhChecksum(number int) int {
	var luhh int

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 { // even
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhh += cur
		number = number / 10
	}
	return luhh % 10
}

func (o *orderController) AddOrder(userLogin, orderID string) error {
	orderNumber, err := strconv.Atoi(orderID)
	if err != nil {
		log.Println("номер заказа содержит символы, отличные от цифр")
		return NewOrderError(orderID, true, false, userLogin, errors.New("номер заказа содержит символы, отличные от цифр"))
	}

	if (orderNumber%10+lunhChecksum(orderNumber/10))%10 != 0 {
		log.Println("контрольное число указано неправильно в номере заказа")
		return NewOrderError(orderID, true, false, userLogin, errors.New("контрольное число указано неправильно в номере заказа"))
	}

	var dbError *database.DBError
	err = o.model.AddOrder(userLogin, orderID)
	if err != nil && errors.As(err, &dbError) {
		log.Println("Ошибка при добавлении заказа в БД:", err)
		return NewOrderError(orderID, false, dbError.Duplicate, dbError.User, err)
	}

	//go o.GetBonuses(orderID)

	return nil
}

func (o *orderController) GetOrders(user string) ([]database.Order, error) {
	orders, err := o.model.GetOrders(user)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (o *orderController) GetBonuses(orderID string) {
	response, err := o.client.Get("/api/orders/" + orderID)
	if err != nil {
		return
	}

	defer response.Body.Close()

	_, err = io.ReadAll(response.Body)
	if err != nil {
		return
	}
}
