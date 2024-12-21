package application

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	calculations "github.com/pAran0k/calc_go/pkg"
)

type Config struct {
	Addr string
}

func ConfigFromEnv() *Config {
	cfg := new(Config)
	cfg.Addr = os.Getenv("PORT")
	if cfg.Addr == "" {
		cfg.Addr = "8080"
	}
	return cfg
}

type Application struct {
	cfg    *Config
	logger *log.Logger
}

func New() *Application {
	return &Application{
		cfg:    ConfigFromEnv(),
		logger: log.New(os.Stdout, "[APP]", log.Flags()),
	}
}
func (a *Application) Run() error {
	for {
		a.logger.Println("Your input:")
		var input string
		_, err := fmt.Scan(&input)
		if err != nil {
			a.logger.Println("failed to read input")
			return err
		}
		res, err := calculations.Calc(input)
		if err != nil {
			a.logger.Printf("failed to calculation %s with error %v",
				input, err.Error())
		} else {
			a.logger.Printf("%s = %f", input, res)
		}

	}
}

type Request struct {
	Request string `json:"request"`
}
type Response struct {
	Result float64 `json:"result,omitempty"`
	Error  error   `json:"error,omitempty"`
}

func CalcHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "not POST method", http.StatusMethodNotAllowed)
		return
	}
	logger := log.New(os.Stdout, "[HTTP]", log.Flags())
	w.Header().Set("Content-Type", "application/json")
	request := new(Request)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Print("invalid JSON format")
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Error: err,
		})
		return
	} else {
		res, err := calculations.Calc(request.Request)
		if err != nil {
			logger.Printf("error: %s", err.Error())
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(Response{
				Error: err,
			})
		} else {
			json.NewEncoder(w).Encode(Response{
				Result: res,
			})
		}

	}
}
func (a *Application) RunServer() error {
	a.logger.Printf("server run in port %s", a.cfg.Addr)
	http.HandleFunc("/api/v1/calculate", CalcHandler)
	return http.ListenAndServe(":"+a.cfg.Addr, nil)

}
