package main

import (
	"context"

	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pAran0k/calc_go/env"
	"github.com/pAran0k/calc_go/internal/services/orchestrator"
)

func main() {
	config := env.LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	orch := orchestrator.NewOrchestrator(config.OrchestratorAddr)
	go func() {
		log.Printf("Оркестратор запущен на %s", config.OrchestratorAddr)
		if err := orch.Run(ctx); err != nil {
			log.Printf("Ошибка оркестратора: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Получен сигнал остановки")
	cancel()
	log.Println("Приложение остановлено")
}
