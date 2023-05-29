package orders

import (
	"fmt"
)

type OrderError struct {
	OrderID           string
	IncorrectID       bool
	Duplicate         bool
	InsufficientFunds bool
	User              string
	Err               error
}

func (e OrderError) Error() string {
	if e.IncorrectID {
		return fmt.Sprintf("Неверный формат номера заказа: %v. Ошибка: %v", e.OrderID, e.Err)
	}

	return e.Err.Error()
}

func (e OrderError) Is(target error) bool {
	err, ok := target.(OrderError)
	if !ok {
		return false
	}

	if err.OrderID != e.OrderID || err.IncorrectID != e.IncorrectID ||
		err.Duplicate != e.Duplicate || err.User != e.User {
		return false
	}

	return true
}

func NewOrderError(orderID string, incorrectID, duplicate, insufficientFunds bool, user string, err error) error {
	return &OrderError{
		OrderID:           orderID,
		IncorrectID:       incorrectID,
		Duplicate:         duplicate,
		InsufficientFunds: insufficientFunds,
		User:              user,
		Err:               err,
	}
}
