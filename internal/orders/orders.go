package orders

import (
	"errors"
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/database"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	processChannelCount     = 10
	ordersToSaveChannelSize = 10
	errorQueueSize          = 10
	orderStatusNew          = "NEW"
)

type Order struct {
	ID         string
	UserLogin  string
	Status     string
	UploadedAt time.Time
	Amount     int
}

type OrderAdderGetter interface {
	AddOrder(user, order string) error
	GetOrders(user string) ([]database.OrderWithAccrual, error)
	Close()
}

type orderController struct {
	ordersToProcess    chan *Order
	ordersToSave       chan *Order
	processingChannels []chan *Order
	errors             chan error
	waitForRetry       bool
	retryMutex         sync.Mutex
	retryAfter         *sync.Cond

	mu    sync.Mutex
	pause *sync.Cond

	model  OrderAdderGetter
	client http.Client
}

func NewOrders(m OrderAdderGetter) OrderAdderGetter {
	result := orderController{
		ordersToProcess: make(chan *Order),
		ordersToSave:    make(chan *Order, ordersToSaveChannelSize),
		errors:          make(chan error, errorQueueSize),

		model:  m,
		client: http.Client{},
	}

	result.retryAfter = sync.NewCond(&result.retryMutex)

	result.pause = sync.NewCond(&result.mu)

	result.initOrderProcessing(processChannelCount)

	return &result
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
		return NewOrderError(orderID, true, false, userLogin, errors.New("номер заказа содержит символы, отличные от цифр"))
	}

	if (orderNumber%10+lunhChecksum(orderNumber/10))%10 != 0 {
		return NewOrderError(orderID, true, false, userLogin, errors.New("контрольное число указано неправильно в номере заказа"))
	}

	var dbError *database.DBError
	err = o.model.AddOrder(userLogin, orderID)
	if err != nil && errors.As(err, &dbError) {
		return NewOrderError(orderID, false, dbError.Duplicate, dbError.User, err)
	}

	if err != nil {
		return err
	}

	//go o.addOrderToProcess(&Order{ID: orderID, UserLogin: userLogin, Status: orderStatusNew}, nil)

	return nil
}

func (o *orderController) GetOrders(user string) ([]database.OrderWithAccrual, error) {
	orders, err := o.model.GetOrders(user)
	if err != nil {
		return nil, err
	}

	return orders, nil
}
