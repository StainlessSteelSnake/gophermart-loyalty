package orders

import (
	"errors"
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/database"
	"net/http"
	"strconv"
	"time"
)

const (
	delayForGettingOrdersToProcess = 1
	processChannelCount            = 10
	ordersToSaveChannelSize        = 10
	errorQueueSize                 = 10
	orderStatusNew                 = "REGISTERED"
)

type Order struct {
	ID         string
	UserLogin  string
	Status     string
	UploadedAt time.Time
	Amount     float32
}

type OrderAdderGetter interface {
	AddOrder(user, order string) error
	GetOrders(user string) ([]database.OrderWithAccrual, error)
	GetUserAccount(user string) (*database.Account, error)
	WithdrawForOrder(user, orderID string, amount float32) error
	GetUserWithdrawals(user string) ([]database.Transaction, error)
	Close()
}

type orderController struct {
	accrualSystemAddress string

	model  database.Storager
	client http.Client

	ordersToProcess    chan *Order
	processingChannels []chan *Order
	ordersToSave       chan *Order
	errors             chan error
	done               chan struct{}
}

func NewOrders(m database.Storager, accrualSystemAddress string) (OrderAdderGetter, error) {
	if len(accrualSystemAddress) == 0 {
		return nil, errors.New("не задан путь к серверу расчёта баллов лояльности")
	}

	result := orderController{
		accrualSystemAddress: accrualSystemAddress,

		ordersToProcess: make(chan *Order),
		ordersToSave:    make(chan *Order, ordersToSaveChannelSize),
		errors:          make(chan error, errorQueueSize),

		model:  m,
		client: http.Client{},
	}

	result.initOrderProcessing(processChannelCount)

	return &result, nil
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

func (o *orderController) AddOrder(user, orderID string) error {
	orderNumber, err := strconv.Atoi(orderID)
	if err != nil {
		return NewOrderError(orderID, true, false, false, user, errors.New("номер заказа содержит символы, отличные от цифр"))
	}

	if (orderNumber%10+lunhChecksum(orderNumber/10))%10 != 0 {
		return NewOrderError(orderID, true, false, false, user, errors.New("контрольное число указано неправильно в номере заказа"))
	}

	var dbError *database.DBOrderError
	err = o.model.AddOrder(user, orderID)
	if err != nil && errors.As(err, &dbError) {
		return NewOrderError(orderID, false, dbError.Duplicate, false, dbError.User, err)
	}

	if err != nil {
		return err
	}

	return nil
}

func (o *orderController) GetOrders(user string) ([]database.OrderWithAccrual, error) {
	orders, err := o.model.GetOrders(user)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (o *orderController) GetUserAccount(user string) (*database.Account, error) {
	account, err := o.model.GetUserAccount(user)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (o *orderController) WithdrawForOrder(user string, orderID string, amount float32) error {
	orderNumber, err := strconv.Atoi(orderID)
	if err != nil {
		return NewOrderError(orderID, true, false, false, user, errors.New("номер заказа содержит символы, отличные от цифр"))
	}

	if (orderNumber%10+lunhChecksum(orderNumber/10))%10 != 0 {
		return NewOrderError(orderID, true, false, false, user, errors.New("контрольное число указано неправильно в номере заказа"))
	}

	account, err := o.GetUserAccount(user)
	if err != nil {
		return err
	}

	if amount > account.Balance {
		return NewOrderError(orderID, false, false, true, user, errors.New("на счёте пользователя "+user+" недостаточно средств ("+strconv.FormatFloat(float64(account.Balance), 'E', -1, 32)+") для списания "+strconv.FormatFloat(float64(amount), 'E', -1, 32)+" баллов"))
	}

	transaction := database.Transaction{
		OrderNumber: orderID,
		UserLogin:   user,
		Type:        database.TransactionTypeWithdrawal,
		Amount:      amount,
		CreatedAt:   database.CustomDateTime{Time: time.Now()},
	}

	err = o.model.AddTransaction(&transaction)
	if err != nil {
		return err
	}

	account.Balance -= amount
	account.Withdrawn += amount

	err = o.model.UpdateUserAccount(account)
	if err != nil {
		return err
	}

	return nil
}

func (o *orderController) GetUserWithdrawals(user string) ([]database.Transaction, error) {
	transactions, err := o.model.GetTransactions(user, database.TransactionTypeWithdrawal)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}
