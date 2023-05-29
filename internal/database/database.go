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
	TransactionTypeAccrual    = "ACCRUAL"
	TransactionTypeWithdrawal = "WITHDRAWAL"
	dbExistError              = "42601"
)

type locker struct {
	user, account sync.RWMutex
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
	Accrual    float32        `json:"accrual,omitempty"`
	UploadedAt CustomDateTime `json:"uploaded_at"`
}

type Account struct {
	UserLogin string  `json:"-"`
	Balance   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type Transaction struct {
	OrderNumber string
	UserLogin   string
	Type        string
	Amount      float32
	CreatedAt   time.Time
}

type Storager interface {
	AddUser(userID string, password string) error
	GetUserPassword(login string) (string, error)

	AddOrder(user string, order string) error
	GetOrders(user string) ([]OrderWithAccrual, error)
	GetOrdersToProcess() ([]Order, error)
	UpdateOrder(*Order, float32) error

	AddTransaction(transaction *Transaction) error
	UpdateUserAccount(account *Account) error

	GetUserAccount(user string) (*Account, error)

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

	_, err = s.conn.Exec(ctx, queryInsertUserAccount, user, 0, 0)
	if err != nil {
		log.Println("Ошибка при добавлении балльного счёта пользователя в БД:", err)
		return err
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

func (s *databaseStorage) UpdateOrder(order *Order, amount float32) error {
	log.Printf("Обновление заказа '%v' пользователя '%v', статус '%v'\n", order.ID, order.UserLogin, order.Status)

	s.locker.account.Lock()
	defer s.locker.account.Unlock()

	account, err := s.GetUserAccount(order.UserLogin)
	if err != nil {
		log.Println("Ошибка при обновлении заказа "+order.ID+":", err)
		return err
	}

	transaction, err := s.getTransaction(order.ID)
	if err != nil {
		log.Println("Ошибка при обновлении заказа "+order.ID+":", err)
		return err
	}

	if amount > 0 && transaction != nil && transaction.Type == TransactionTypeAccrual {
		err = errors.New("Для заказа " + order.ID + " уже существует транзакция начисления от " + transaction.CreatedAt.String())
		return err
	}

	ctx := context.Background()

	_, err = s.conn.Exec(ctx, queryUpdateOrder, order.ID, order.Status)
	if err != nil {
		log.Println("Ошибка при обновлении заказа "+order.ID+":", err)
		return err
	}

	if amount > 0 {
		account.Balance += amount

		transaction = &Transaction{
			OrderNumber: order.ID,
			UserLogin:   order.UserLogin,
			Type:        TransactionTypeAccrual,
			Amount:      amount,
			CreatedAt:   time.Now(),
		}

		_, err = s.conn.Exec(ctx, queryUpdateUserAccount, account.UserLogin, account.Balance, account.Withdrawn)
		if err != nil {
			log.Println("Ошибка при обновлении заказа "+order.ID+":", err)
			return err
		}

		_, err = s.conn.Exec(ctx, queryInsertTransaction, transaction.OrderNumber, transaction.UserLogin, transaction.Type, transaction.Amount, transaction.CreatedAt)
		if err != nil {
			log.Println("Ошибка при обновлении заказа "+order.ID+":", err)
			return err
		}
	}

	log.Println("Заказ " + order.ID + " успешно обновлён")
	return nil
}

func (s *databaseStorage) GetUserAccount(user string) (*Account, error) {
	log.Printf("Получение балльного счёта пользователя '%v'\n", user)

	ctx := context.Background()
	var account Account

	row := s.conn.QueryRow(ctx, queryGetUserAccount, user)
	err := row.Scan(&account.UserLogin, &account.Balance, &account.Withdrawn)
	if err != nil {
		log.Println("Ошибка при считывании балльного счёта пользователя "+user+" из БД:", err)
		return nil, err
	}

	return &account, nil
}

func (s *databaseStorage) UpdateUserAccount(account *Account) error {
	log.Printf("Обновление балльного счёта пользователя '%v', баланс '%v', всего списано '%v'\n", account.UserLogin, account.Balance, account.Withdrawn)

	ctx := context.Background()

	_, err := s.conn.Exec(ctx, queryUpdateUserAccount, account.UserLogin, account.Balance, account.Withdrawn)
	if err != nil {
		log.Println("Ошибка при обновлении балльного счёта пользователя:", err)
		return err
	}

	return nil
}

func (s *databaseStorage) getTransaction(orderID string) (*Transaction, error) {
	log.Printf("Получение транзакции по заказу '%v'\n", orderID)

	ctx := context.Background()
	var transaction Transaction

	row := s.conn.QueryRow(ctx, queryGetTransaction, orderID)
	err := row.Scan(&transaction.OrderNumber, &transaction.UserLogin, &transaction.Type, &transaction.Amount, &transaction.CreatedAt)

	if err != nil && err == pgx.ErrNoRows {
		log.Println("Транзакции по заказу " + orderID + " не найдены")
		return nil, nil
	}

	if err != nil {
		log.Println("Ошибка при считывании транзакции по заказу "+orderID+" из БД:", err)
		return nil, err
	}

	return &transaction, nil
}

func (s *databaseStorage) AddTransaction(transaction *Transaction) error {
	log.Printf("Добавление транзакции для заказа '%v', пользователя '%v', тип '%v', сумма '%v'\n", transaction.OrderNumber, transaction.UserLogin, transaction.Type, transaction.Amount)

	ctx := context.Background()

	_, err := s.conn.Exec(ctx, queryInsertTransaction, transaction.OrderNumber, transaction.UserLogin, transaction.Type, transaction.Amount, transaction.CreatedAt)
	if err != nil {
		log.Println("Ошибка при добавлении транзакции:", err)
		return err
	}

	return nil
}
