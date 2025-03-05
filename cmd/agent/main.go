package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pAran0k/calc_go/env"
	"github.com/pAran0k/calc_go/internal/services/agent"
)

func main() {
	config := env.LoadConfig()
	stop := make(chan struct{})
	agt := agent.NewAgent()
	go func() {
		log.Printf("Агент запущен с %d вычислителями", config.ComputingPower)
		agt.Run(stop)
	}()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	close(stop)
	log.Println("Приложение остановлено")
}
