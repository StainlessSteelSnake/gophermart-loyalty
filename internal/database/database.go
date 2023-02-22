package database

import (
	"context"
	"errors"
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

type Storager interface {
	AddUser(userID string, password string) error
	GetUserPassword(login string) (string, error)
	AddOrder(user string, order string) error
	GetOrders(user string) ([]Order, error)
	Close()
}

type Order struct {
	ID         string
	UserLogin  string
	Status     string
	UploadedAt time.Time
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

func (s *databaseStorage) AddUser(login string, password string) error {
	s.locker.user.Lock()
	defer s.locker.user.Unlock()

	log.Printf("Добавление в БД пользователя '%v' с хэшем пароля '%v'\n", login, password)

	ctx := context.Background()
	var pgErr *pgconn.PgError
	ct, err := s.conn.Exec(ctx, queryInsertUser, login, password)
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
		return NewDBError(login, login, false, true, err)
	}

	log.Println("Добавлено записей пользователей в таблицу БД:", ct.RowsAffected())
	return nil
}

func (s *databaseStorage) GetUserPassword(login string) (string, error) {
	s.locker.user.RLock()
	defer s.locker.user.RUnlock()

	var passwordHash string
	ctx := context.Background()

	r := s.conn.QueryRow(ctx, querySelectPassword, login)
	err := r.Scan(&passwordHash)
	if err != nil {
		log.Println("Ошибка при считывании пароля пользователя из БД:", err)
		return "", NewDBError(login, login, false, false, err)
	}

	return passwordHash, nil
}

func (s *databaseStorage) AddOrder(user, order string) error {
	s.locker.orders.Lock()
	defer s.locker.orders.Unlock()

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
		r := s.conn.QueryRow(ctx, queryGetOrderUserByID, order)

		err = r.Scan(&orderUser)
		if err != nil {
			log.Println("Ошибка при добавлении заказа '"+order+"' под пользователем '"+user+"' в БД, код:", pgErr.Code, ", сообщение:", pgErr.Error())
			return err
		}

		if orderUser != user {
			log.Println("Заказ '" + order + "' уже был загружен ранее другим пользователем '" + orderUser + "'")
			return NewDBError(order, orderUser, true, true, errors.New("заказ '"+order+"' уже был загружен ранее другим пользователем"))
		}

		log.Println("Заказ '" + order + "' уже был загружен ранее текущим пользователем '" + user + "'")
		return NewDBError(order, user, false, true, errors.New("заказ '"+order+"' уже был загружен ранее текущим пользователем '"+user+"'"))

	}

	log.Println("Добавлено записей заказов в таблицу БД:", ct.RowsAffected())
	return nil
}

func (s *databaseStorage) GetOrders(user string) ([]Order, error) {
	return nil, nil
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
