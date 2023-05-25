package database

import "fmt"

type DBError struct {
	User        string
	AnotherUser bool
	Duplicate   bool
	Err         error
}

type DBUserError DBError

type DBOrderError struct {
	Order string
	DBError
}

func (e DBError) Error() string {
	if e.Duplicate {
		return fmt.Sprintf("При попытке добавления записи в БД обнаружен дубликат. Ошибка: %v", e.Err)
	}

	return e.Err.Error()
}

func (e DBUserError) Error() string {
	if e.Duplicate {
		return fmt.Sprintf("При попытке добавления пользователя в БД обнаружен дубликат. Ошибка: %v", e.Err)
	}

	return e.Err.Error()
}

func (e DBOrderError) Error() string {
	if e.Duplicate {
		return fmt.Sprintf("При попытке добавления заказа в БД обнаружен дубликат. Ошибка: %v", e.Err)
	}

	return e.Err.Error()
}

func (e DBError) Is(target error) bool {
	err, ok := target.(DBError)
	if !ok {
		return false
	}

	if err.Duplicate != e.Duplicate || err.AnotherUser != e.AnotherUser {
		return false
	}

	return true
}

func (e DBUserError) Is(target error) bool {
	err, ok := target.(DBUserError)
	if !ok {
		return false
	}

	if err.Duplicate != e.Duplicate || err.AnotherUser != e.AnotherUser {
		return false
	}

	return true
}

func (e DBOrderError) Is(target error) bool {
	err, ok := target.(DBOrderError)
	if !ok {
		return false
	}

	if err.Duplicate != e.Duplicate || err.AnotherUser != e.AnotherUser {
		return false
	}

	return true
}

func NewDBError(user string, anotherUser bool, duplicate bool, err error) error {
	return &DBError{
		User:        user,
		Duplicate:   duplicate,
		AnotherUser: anotherUser,
		Err:         err,
	}
}

func NewDBUserError(user string, anotherUser bool, duplicate bool, err error) error {
	return &DBUserError{
		User:        user,
		Duplicate:   duplicate,
		AnotherUser: anotherUser,
		Err:         err,
	}
}

func NewDBOrderError(order string, user string, anotherUser bool, duplicate bool, err error) error {
	return &DBOrderError{
		Order: order,
		DBError: DBError{
			User:        user,
			Duplicate:   duplicate,
			AnotherUser: anotherUser,
			Err:         err,
		},
	}
}
