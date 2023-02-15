package database

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"log"
	"strings"
	"sync"
)

const dbUser = "gophermart_app"

type locker struct {
	user, orders, transactions, account sync.RWMutex
}

type databaseStorage struct {
	conn   *pgx.Conn
	dbUser string
	locker
}

type Storager interface {
	AddUser(login string, password string) error
	GetUser(login string, password string) error
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

	err = dbStorage.init(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return dbStorage
}

func (s *databaseStorage) init(ctx context.Context) error {

	_, err := s.conn.Exec(ctx, sqlCreateDatabase, s.dbUser)
	if err != nil {
		return err
	}

	_, err = s.conn.Exec(ctx, sqlCreateTableUsers, s.dbUser)
	if err != nil {
		return err
	}

	_, err = s.conn.Exec(ctx, sqlCreateTableOrders, s.dbUser)
	if err != nil {
		return err
	}

	_, err = s.conn.Exec(ctx, sqlCreateTableAccounts, s.dbUser)
	if err != nil {
		return err
	}

	_, err = s.conn.Exec(ctx, sqlCreateTableAccounts, s.dbUser)
	if err != nil {
		return err
	}

	_, err = s.conn.Exec(ctx, sqlCreateTableTransactions, s.dbUser)
	if err != nil {
		return err
	}

	log.Println("Таблицы успешно инициализированы в БД")
	return nil
}

func (s *databaseStorage) AddUser(login string, password string) error {
	return nil
}
func (s *databaseStorage) GetUser(login string, password string) error {
	return nil
}
