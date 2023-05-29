package main

import (
	"context"
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/auth"
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/orders"
	"log"

	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/config"
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/database"
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/handlers"
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/server"
)

func main() {
	cfg := config.NewConfiguration()
	if cfg == nil {
		log.Fatal("Не удалось получить конфигурацию сервиса системы лояльности")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbStorage := database.NewDatabaseStorage(ctx, cfg.DatabaseURI)
	if dbStorage == nil {
		log.Fatal("Не удалось инициализировать БД сервиса системы лояльности")
	}
	defer dbStorage.Close()

	authenticator, err := auth.NewAuth(dbStorage)
	if err != nil {
		log.Fatal(err)
	}

	orderController, err := orders.NewOrders(dbStorage, cfg.AccrualSystemAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer orderController.Close()

	//orderController.ProcessOrder("12345678903")

	handler := handlers.NewHandler(cfg.BaseURL, authenticator, orderController)

	srv := server.NewServer(cfg.RunAddress, handler)
	log.Fatal(srv.ListenAndServe())
}
