package main

import (
	"context"
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

	ctx, _ := context.WithCancel(context.Background())
	dbStorage := database.NewDatabaseStorage(ctx, cfg.DatabaseURI)
	if dbStorage == nil {
		log.Fatal("Не удалось инициализировать БД сервиса системы лояльности")
	}

	handler := handlers.NewHandler(dbStorage, cfg.BaseURL)

	srv := server.NewServer(cfg.RunAddress, handler)
	log.Fatal(srv.ListenAndServe())
}
