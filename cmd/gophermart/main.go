package main

import (
	"context"
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/config"
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/database"
	"log"
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
}
