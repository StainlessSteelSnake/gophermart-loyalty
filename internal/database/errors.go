package database

import "fmt"

type DBError struct {
	Entity    string
	Duplicate bool
	Err       error
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

	if err.Entity != e.Entity || err.Duplicate != e.Duplicate {
		return false
	}

	return true
}

func NewDBError(entity string, duplicate bool, err error) error {
	return &DBError{
		Entity:    entity,
		Duplicate: duplicate,
		Err:       err,
	}
}
