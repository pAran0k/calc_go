package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/pAran0k/calc_go/env"
	"github.com/pAran0k/calc_go/models"
)

type Agent struct {
	ind    int
	Tasks  []chan models.Task
	IsFree []bool
	Work   []models.Task
	Config env.Config
	Client *http.Client
	wg     sync.WaitGroup
}

func NewAgent() *Agent {
	config := env.LoadConfig()
	numWorkers := config.ComputingPower

	agent := &Agent{
		ind:    1,
		Tasks:  make([]chan models.Task, numWorkers),
		IsFree: make([]bool, numWorkers),
		Work:   make([]models.Task, numWorkers),
		Config: config,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for i := 0; i < numWorkers; i++ {
		agent.Tasks[i] = make(chan models.Task, 1)
		agent.IsFree[i] = true
	}

	return agent
}

func (a *Agent) Run(stop <-chan struct{}) {
	log.Printf("Запуск агента %d с %d вычислителями", a.ind, len(a.Tasks))

	for i := 0; i < len(a.Tasks); i++ {
		a.wg.Add(1)
		go a.worker(i, a.Tasks[i], stop)
	}

	baseURL := "http://localhost" + a.Config.OrchestratorAddr
	for {
		select {
		case <-stop:
			for i := 0; i < len(a.Tasks); i++ {
				close(a.Tasks[i])
			}
			a.wg.Wait()
			return
		default:
			workerID := a.getFreeWorker()
			if workerID == -1 {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			task, err := a.getTask(baseURL)
			if err != nil {
				if err.Error() == "no task available" {
					time.Sleep(1 * time.Second)
					continue
				}
				log.Printf("[Агент %d] Ошибка при получении задачи: %v", a.ind, err)
				continue
			}

			log.Printf("[Агент %d] Получена задача %s для вычислителя %d", a.ind, task.ID, workerID)
			a.IsFree[workerID] = false
			a.Work[workerID] = *task
			a.Tasks[workerID] <- *task
		}
	}
}

func (a *Agent) worker(workerID int, taskChan <-chan models.Task, stop <-chan struct{}) {
	defer a.wg.Done()
	baseURL := "http://localhost" + a.Config.OrchestratorAddr

	for {
		select {
		case <-stop:
			return
		case task := <-taskChan:
			log.Printf("[Агент %d] Вычислитель %d: Принята задача %s: %+v", a.ind, workerID, task.ID, task)
			result, err := a.processTask(&task, baseURL)
			if err != nil {
				log.Printf("[Агент %d] Вычислитель %d: Ошибка при обработке задачи %s: %v", a.ind, workerID, task.ID, err)
				a.IsFree[workerID] = true
				continue
			}

			log.Printf("[Агент %d] Вычислитель %d: Результат задачи %s готов к отправке: %f", a.ind, workerID, task.ID, result.Value)
			for retries := 0; retries < 5; retries++ {
				err = a.sendResult(baseURL, result)
				if err != nil {
					log.Printf("[Агент %d] Вычислитель %d: Ошибка при отправке результата для задачи %s: %v, попытка %d", a.ind, workerID, task.ID, err, retries+1)
					time.Sleep(500 * time.Millisecond)
					continue
				}
				log.Printf("[Агент %d] Вычислитель %d: Результат %s отправлен: %f", a.ind, workerID, task.ID, result.Value)
				break
			}
			if err != nil {
				log.Printf("[Агент %d] Вычислитель %d: Не удалось отправить результат для задачи %s после всех попыток: %v", a.ind, workerID, task.ID, err)
				a.IsFree[workerID] = true
				continue
			}

			log.Printf("[Агент %d] Вычислитель %d: Задача %s выполнена: %f", a.ind, workerID, task.ID, result.Value)
			a.IsFree[workerID] = true
		}
	}
}

func (a *Agent) getFreeWorker() int {
	for i, free := range a.IsFree {
		if free {
			return i
		}
	}
	return -1
}

func (a *Agent) getTask(baseURL string) (*models.Task, error) {
	resp, err := a.Client.Get(baseURL + "/internal/task")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no task available")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Task models.Task `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response.Task, nil
}

func (a *Agent) processTask(task *models.Task, baseURL string) (*models.Result, error) {
	var arg1, arg2 float64
	var err error

	if isNumeric(task.Arg1) {
		arg1, err = strconv.ParseFloat(task.Arg1, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid arg1: %v", err)
		}
	} else {
		for retries := 0; retries < 5; retries++ {
			arg1, err = a.getTaskResult(baseURL, task.Arg1)
			if err == nil {
				break
			}
			log.Printf("[Агент %d] Ожидание результата для %s: %v, попытка %d", a.ind, task.Arg1, err, retries+1)
			time.Sleep(1 * time.Second)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get result for Arg1 %s after retries: %v", task.Arg1, err)
		}
	}

	if isNumeric(task.Arg2) {
		arg2, err = strconv.ParseFloat(task.Arg2, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid arg2: %v", err)
		}
	} else {
		for retries := 0; retries < 5; retries++ {
			arg2, err = a.getTaskResult(baseURL, task.Arg2)
			if err == nil {
				break
			}
			log.Printf("[Агент %d] Ожидание результата для %s: %v, попытка %d", a.ind, task.Arg2, err, retries+1)
			time.Sleep(1 * time.Second)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get result for Arg2 %s after retries: %v", task.Arg2, err)
		}
	}

	var value float64
	var operationTime int
	switch task.Operation {
	case "+":
		value = arg1 + arg2
		operationTime = a.Config.TimeAdditionMS
	case "-":
		value = arg1 - arg2
		operationTime = a.Config.TimeSubtractionMS
	case "*":
		value = arg1 * arg2
		operationTime = a.Config.TimeMultiplicationMS
	case "/":
		if arg2 == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		value = arg1 / arg2
		operationTime = a.Config.TimeDivisionMS
	default:
		return nil, fmt.Errorf("unsupported operation: %s", task.Operation)
	}

	time.Sleep(time.Duration(operationTime) * time.Millisecond)

	return &models.Result{
		TaskID: task.ID,
		Value:  value,
	}, nil
}

func (a *Agent) getTaskResult(baseURL, taskID string) (float64, error) {
	url := baseURL + "/internal/task/result/" + taskID
	resp, err := a.Client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Result float64 `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, err
	}

	return response.Result, nil
}

func (a *Agent) sendResult(baseURL string, result *models.Result) error {
	body, err := json.Marshal(result)
	if err != nil {
		return err
	}

	maxRetries := 5
	for retries := 0; retries < maxRetries; retries++ {
		resp, err := a.Client.Post(baseURL+"/internal/task", "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Printf("[Агент %d] Ошибка отправки результата %s: %v, попытка %d", a.ind, result.TaskID, err, retries+1)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			log.Printf("[Агент %d] Результат %s отправлен: %f", a.ind, result.TaskID, result.Value)
			return nil
		case http.StatusInternalServerError:
			log.Printf("[Агент %d] Ошибка сервера 500 для задачи %s, попытка %d", a.ind, result.TaskID, retries+1)
			time.Sleep(1 * time.Second)
			continue
		default:
			log.Printf("[Агент %d] Неожиданный код ответа %d для задачи %s", a.ind, resp.StatusCode, result.TaskID)
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}

	log.Printf("[Агент %d] Не удалось отправить результат %s после %d попыток", a.ind, result.TaskID, maxRetries)
	return fmt.Errorf("failed to send result after %d retries", maxRetries)

}

func isNumeric(arg string) bool {
	_, err := strconv.ParseFloat(arg, 64)
	return err == nil
}
