package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	dbExistError = "42601"
)

type locker struct {
	user, orders, transactions, account sync.RWMutex
}

type databaseStorage struct {
	conn   *pgx.Conn
	dbUser string
	locker
}

type CustomDateTime struct {
	time.Time
}

type Order struct {
	ID         string
	UserLogin  string
	Status     string
	UploadedAt time.Time
}

type OrderWithAccrual struct {
	ID         string         `json:"number"`
	Status     string         `json:"status"`
	Accrual    int            `json:"accrual,omitempty"`
	UploadedAt CustomDateTime `json:"uploaded_at"`
}

type Storager interface {
	AddUser(userID string, password string) error
	GetUserPassword(login string) (string, error)
	AddOrder(user string, order string) error
	GetOrders(user string) ([]OrderWithAccrual, error)
	GetOrdersToProcess() ([]Order, error)
	UpdateOrder(Order, int) error
	Close()
}

func (c *CustomDateTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`) // remove quotes
	if s == "null" {
		return
	}
	c.Time, err = time.Parse(time.RFC3339, s)
	return
}

func (c *CustomDateTime) MarshalJSON() ([]byte, error) {
	if c.Time.IsZero() {
		return nil, nil
	}
	return []byte(fmt.Sprintf(`"%s"`, c.Time.Format(time.RFC3339))), nil
}

func NewDatabaseStorage(ctx context.Context, databaseURI string) Storager {
	dbStorage := &databaseStorage{}

	var err error
	dbStorage.conn, err = pgx.Connect(ctx, databaseURI)
	if err != nil {
		log.Fatal(err)
		return dbStorage
	}

	dbCfg := strings.Split(databaseURI, ":")
	if len(dbCfg) < 2 {
		log.Fatal(errors.New("в URI базы данных отсутствует информация о пользователе"))
		return dbStorage
	}

	dbStorage.dbUser = strings.TrimPrefix(dbCfg[1], "//")
	if dbStorage.dbUser == "" {
		log.Fatal(errors.New("в URI базы данных отсутствует информация о пользователе"))
		return dbStorage
	}
	log.Println("Пользователь БД:", dbStorage.dbUser)

	err = dbStorage.init(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return dbStorage
}

func (s *databaseStorage) init(ctx context.Context) error {

	var pgErr *pgconn.PgError

	_, err := s.conn.Exec(ctx, sqlCreateDatabase, s.dbUser)
	if err != nil && !errors.As(err, &pgErr) {
		return err
	}

	if err != nil && pgErr.Code != dbExistError {
		return err
	}

	_, err = s.conn.Exec(ctx, sqlCreateTableUsers)
	if err != nil {
		return err
	}

	_, err = s.conn.Exec(ctx, sqlCreateTableOrders)
	if err != nil {
		return err
	}

	_, err = s.conn.Exec(ctx, sqlCreateTableAccounts)
	if err != nil {
		return err
	}

	_, err = s.conn.Exec(ctx, sqlCreateTableAccounts)
	if err != nil {
		return err
	}

	_, err = s.conn.Exec(ctx, sqlCreateTableTransactions)
	if err != nil {
		return err
	}

	log.Println("Таблицы успешно инициализированы в БД")
	return nil
}

func (s *databaseStorage) Close() {
	if s.conn == nil {
		return
	}

	ctx := context.Background()
	err := s.conn.Close(ctx)
	if err != nil {
		log.Println(err)
		return
	}
}

func (s *databaseStorage) AddUser(user string, password string) error {
	log.Printf("Добавление в БД пользователя '%v' с хэшем пароля '%v'\n", user, password)

	ctx := context.Background()
	var pgErr *pgconn.PgError
	ct, err := s.conn.Exec(ctx, queryInsertUser, user, password)
	if err != nil && !errors.As(err, &pgErr) {
		log.Println("Ошибка при добавлении пользователя в БД:", err)
		return err
	}

	if err != nil && pgErr.Code != pgerrcode.UniqueViolation {
		log.Println("Ошибка при добавлении пользователя в БД, код:", pgErr.Code, ", сообщение:", pgErr.Error())
		return err
	}

	if err != nil {
		log.Println("Ошибка при добавлении пользователя в БД, код:", pgErr.Code, ", сообщение:", pgErr.Error())
		return NewDBUserError(user, false, true, err)
	}

	log.Println("Добавлено записей пользователей в таблицу БД:", ct.RowsAffected())
	return nil
}

func (s *databaseStorage) GetUserPassword(user string) (string, error) {
	var passwordHash string
	ctx := context.Background()

	row := s.conn.QueryRow(ctx, querySelectPassword, user)
	err := row.Scan(&passwordHash)
	if err != nil {
		log.Println("Ошибка при считывании пароля пользователя из БД:", err)
		return "", NewDBUserError(user, false, false, err)
	}

	return passwordHash, nil
}

func (s *databaseStorage) AddOrder(user, order string) error {
	log.Printf("Добавление в БД заказа '%v' для пользователя '%v'\n", order, user)

	ctx := context.Background()
	var pgErr *pgconn.PgError
	ct, err := s.conn.Exec(ctx, queryInsertOrder, order, user, time.Now())
	if err != nil && !errors.As(err, &pgErr) {
		log.Println("Ошибка при добавлении заказа '"+order+"' под пользователем '"+user+"' в БД:", err)
		return err
	}

	if err != nil && pgErr.Code != pgerrcode.UniqueViolation {
		log.Println("Ошибка при добавлении заказа '"+order+"' под пользователем '"+user+"' в БД, код:", pgErr.Code, ", сообщение:", pgErr.Error())
		return err
	}

	var orderUser string
	if err != nil {
		row := s.conn.QueryRow(ctx, queryGetOrderUserByID, order)

		err = row.Scan(&orderUser)
		if err != nil {
			log.Println("Ошибка при добавлении заказа '"+order+"' под пользователем '"+user+"' в БД, код:", pgErr.Code, ", сообщение:", pgErr.Error())
			return err
		}

		if orderUser != user {
			log.Println("Заказ '" + order + "' уже был загружен ранее другим пользователем '" + orderUser + "'")
			return NewDBOrderError(order, orderUser, true, true, errors.New("заказ '"+order+"' уже был загружен ранее другим пользователем"))
		}

		log.Println("Заказ '" + order + "' уже был загружен ранее текущим пользователем '" + user + "'")
		return NewDBOrderError(order, user, false, true, errors.New("заказ '"+order+"' уже был загружен ранее текущим пользователем '"+user+"'"))

	}

	log.Println("Добавлено записей заказов в таблицу БД:", ct.RowsAffected())
	return nil
}

func (s *databaseStorage) GetOrders(user string) ([]OrderWithAccrual, error) {
	ctx := context.Background()

	rows, err := s.conn.Query(ctx, queryGetOrdersByUser, user)
	if err != nil {
		log.Println("Ошибка при запросе списка заказов пользователя:", err)
		return nil, err
	}

	defer rows.Close()

	result := make([]OrderWithAccrual, 0)

	for rows.Next() {
		var order OrderWithAccrual
		err = rows.Scan(&order.ID, &order.Status, &order.Accrual, &order.UploadedAt.Time)
		if err != nil {
			log.Println("Ошибка при считывании записи заказа пользователя из списка:", err)
			return nil, err
		}

		result = append(result, order)
	}

	err = rows.Err()
	if err != nil {
		log.Println("Ошибка при считывании записей заказов пользователя из списка:", err)
		return nil, err
	}

	return result, nil
}

func (s *databaseStorage) GetOrdersToProcess() ([]Order, error) {
	ctx := context.Background()

	rows, err := s.conn.Query(ctx, queryGetOrdersToProcess)
	if err != nil {
		log.Println("Ошибка при запросе заказов для обработки начисления баллов:", err)
		return nil, err
	}

	defer rows.Close()

	result := make([]Order, 0)

	for rows.Next() {
		var order Order
		err = rows.Scan(&order.ID, &order.UserLogin, &order.Status, &order.UploadedAt)
		if err != nil {
			log.Println("Ошибка при считывании записи заказа пользователя из списка:", err)
			return nil, err
		}

		result = append(result, order)
	}

	err = rows.Err()
	if err != nil {
		log.Println("Ошибка при считывании записей заказов пользователя из списка:", err)
		return nil, err
	}

	return result, nil
}

func (s *databaseStorage) UpdateOrder(order Order, accrual int) error {
	return nil
}
