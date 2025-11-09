package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"udp-current-time/config"
	"udp-current-time/server"
)

func main() {
	// Загрузить конфигурацию
	cfg := config.NewConfig()

	// Создать логгер
	logger := log.New(os.Stdout, "logger: ", log.Lshortfile)

	// Создать экземпляр сервера
	srv := server.NewServer(cfg, logger)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := srv.Serve(ctx); err != nil {
		logger.Fatal(err)
	}

}
