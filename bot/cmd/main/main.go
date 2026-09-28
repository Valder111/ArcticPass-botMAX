// package main - точка входа: go run cmd/main/main.go из каталога bot/.
package main

import (
	"ArcticPassBot/internal/api"
	"ArcticPassBot/internal/config"
	"ArcticPassBot/internal/max"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.LoadConfig()

	httpClient, err := max.NewServiceHTTPClient()
	if err != nil {
		log.Fatalf("HTTP-клиент API: %v", err)
	}
	max.SetObjectClient(api.NewObjectClient(cfg.BaseURL, httpClient))

	bot, err := max.NewBotClient(cfg.MaxToken)
	if err != nil {
		log.Fatalf("Ошибка инициализации бота: %v", err)
	}

	max.ConfigureMiniApp(cfg.MiniAppWebApp, cfg.MiniAppStartParam, max.FetchBotInfo(context.Background(), bot))

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
		<-exit
		fmt.Println("\nЗавершение работы бота...")
		cancel()
	}()

	fmt.Println("Бот запущен и ожидает сообщений...")

	for upd := range bot.GetUpdates(ctx) {
		max.Dispatch(ctx, bot, upd)
	}
}
