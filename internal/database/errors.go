package database

import "fmt"

type DBError struct {
	Entity      string
	User        string
	AnotherUser bool
	Duplicate   bool
	Err         error
}

func (e DBError) Error() string {
	if e.Duplicate {
		return fmt.Sprintf("При попытке добавления записи '%v' в БД обнаружен дубликат. Ошибка: %v", e.Entity, e.Err)
	}

	return e.Err.Error()
}

func (e DBError) Is(target error) bool {
	err, ok := target.(DBError)
	if !ok {
		return false
	}

	if err.Entity != e.Entity || err.Duplicate != e.Duplicate || err.AnotherUser != e.AnotherUser {
		return false
	}

	return true
}

func NewDBError(entity, user string, anotherUser bool, duplicate bool, err error) error {
	return &DBError{
		Entity:      entity,
		User:        user,
		Duplicate:   duplicate,
		AnotherUser: anotherUser,
		Err:         err,
	}
}
