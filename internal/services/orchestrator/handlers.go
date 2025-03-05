package orchestrator

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/pAran0k/calc_go/models"
	calculations "github.com/pAran0k/calc_go/pkg/calc"
)

type Store struct {
	Mu           sync.Mutex
	Expressions  map[int]models.Expression
	Tasks        map[string]models.Task
	PendingTasks chan models.Task
}

func NewStore() *Store {
	return &Store{
		Expressions:  make(map[int]models.Expression),
		Tasks:        make(map[string]models.Task),
		PendingTasks: make(chan models.Task, 100),
	}
}

func (s *Store) AddExpression(expr models.Expression) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Expressions[expr.Id] = expr
	log.Printf("Добавлено выражение %d: %+v", expr.Id, expr)
}

func (s *Store) GetExpression(id int) (models.Expression, bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	expr, exists := s.Expressions[id]
	log.Printf("Запрошено выражение %d: найдено=%v, %+v", id, exists, expr)
	return expr, exists
}

func (s *Store) GetAllExpressions() []models.Expression {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	var expressions []models.Expression
	for _, expr := range s.Expressions {
		expressions = append(expressions, expr)
	}
	log.Printf("Возвращено %d выражений", len(expressions))
	return expressions
}

func (s *Store) AddTask(task models.Task) {
	s.Mu.Lock()
	s.Tasks[task.ID] = task
	log.Printf("Задача %s добавлена в Tasks: %+v, всего задач: %d", task.ID, task, len(s.Tasks))
	s.PendingTasks <- task
	s.Mu.Unlock()
}

func (s *Store) UpdateTask(result models.Result) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	task, exists := s.Tasks[result.TaskID]
	if !exists {
		log.Printf("Ошибка: задача %s не найдена в Tasks: %+v", result.TaskID, s.Tasks)
		return false
	}

	log.Printf("Обновление задачи %s: старое значение %+v, новый результат %f", result.TaskID, task, result.Value)
	task.Result = result.Value
	task.Completed = true
	s.Tasks[result.TaskID] = task
	log.Printf("Задача %s обновлена: %+v", result.TaskID, task)

	parts := strings.Split(result.TaskID, "-")
	if len(parts) < 3 {
		log.Printf("Ошибка: неверный формат TaskID %s", result.TaskID)
		return false
	}
	exprIDStr := parts[2]
	id, err := strconv.Atoi(exprIDStr)
	if err != nil {
		log.Printf("Ошибка разбора exprID из %s: %v", result.TaskID, err)
		return false
	}

	expr, exists := s.Expressions[id]
	if !exists {
		log.Printf("Выражение %d не найдено для задачи %s", id, result.TaskID)
		return false
	}

	allCompleted := true
	for _, t := range s.Tasks {
		if strings.Contains(t.ID, fmt.Sprintf("expr-%d", id)) && !t.Completed {
			log.Printf("Задача %s для выражения %d ещё не завершена: %+v", t.ID, id, t)
			allCompleted = false
			break
		}
	}

	if allCompleted {
		log.Printf("Все задачи для выражения %d завершены, пересчитываем результат", id)
		finalResult, err := s.calculateExpression(expr)
		if err != nil {
			expr.Status = 3
			s.Expressions[id] = expr
			log.Printf("Ошибка при вычислении выражения %d: %v", id, err)
			return true
		}
		expr.Result = finalResult
		expr.Status = 0
		s.Expressions[id] = expr
		log.Printf("Выражение %d завершено: %+v", id, expr)
	}

	return true
}

func (s *Store) calculateExpression(expr models.Expression) (float64, error) {
	return s.evaluateNode(expr.Node)
}

func (s *Store) evaluateNode(node *models.Node) (float64, error) {
	if node == nil {
		return 0, fmt.Errorf("nil node")
	}

	if !calculations.IsOperator(node.Value) {
		return strconv.ParseFloat(node.Value, 64)
	}

	leftVal, err := s.evaluateNode(node.Left)
	if err != nil {
		return 0, err
	}

	rightVal, err := s.evaluateNode(node.Right)
	if err != nil {
		return 0, err
	}

	switch node.Value {
	case "+":
		return leftVal + rightVal, nil
	case "-":
		return leftVal - rightVal, nil
	case "*":
		return leftVal * rightVal, nil
	case "/":
		if rightVal == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return leftVal / rightVal, nil
	default:
		return 0, fmt.Errorf("unsupported operation: %s", node.Value)
	}
}

func (s *Store) GetPendingTask() (models.Task, bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for i := 0; i < cap(s.PendingTasks); i++ {
		select {
		case task := <-s.PendingTasks:
			if s.isTaskReady(task) {
				log.Printf("Задача %s готова и выдана: %+v", task.ID, task)
				return task, true
			}
			s.PendingTasks <- task
		default:
			return models.Task{}, false
		}
	}
	return models.Task{}, false
}

func (s *Store) isTaskReady(task models.Task) bool {
	if isNumeric(task.Arg1) && isNumeric(task.Arg2) {
		return true
	}
	if !isNumeric(task.Arg1) {
		depTask, exists := s.Tasks[task.Arg1]
		if !exists || !depTask.Completed {
			return false
		}
	}
	if !isNumeric(task.Arg2) {
		depTask, exists := s.Tasks[task.Arg2]
		if !exists || !depTask.Completed {
			return false
		}
	}
	return true
}

func isNumeric(arg string) bool {
	_, err := strconv.ParseFloat(arg, 64)
	return err == nil
}

func HandleTask(st *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path == "/internal/task" {
				handleGetTask(w, r, st)
			} else {
				http.Error(w, "Not found", http.StatusNotFound)
			}
		case http.MethodPost:
			handlePostTask(w, r, st)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func HandleTaskResult(st *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handleGetTaskResult(w, r, st)
	}
}

func handleGetTask(w http.ResponseWriter, r *http.Request, st *Store) {
	task, exists := st.GetPendingTask()
	if !exists {
		http.Error(w, "No task available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Task models.Task `json:"task"`
	}{Task: task})
}

func handlePostTask(w http.ResponseWriter, r *http.Request, st *Store) {
	var result models.Result
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		log.Printf("Ошибка декодирования результата: %v", err)
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	log.Printf("Получен результат для задачи %s: %f", result.TaskID, result.Value)
	if result.TaskID == "" {
		log.Println("Отсутствует TaskID в результате")
		http.Error(w, "Missing task ID", http.StatusUnprocessableEntity)
		return
	}

	if !st.UpdateTask(result) {
		log.Printf("Задача %s не найдена при обновлении", result.TaskID)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	log.Printf("Результат задачи %s успешно принят: %f", result.TaskID, result.Value)
	w.WriteHeader(http.StatusOK)
}

func handleGetTaskResult(w http.ResponseWriter, r *http.Request, st *Store) {
	taskID := strings.TrimPrefix(r.URL.Path, "/internal/task/result/")
	if taskID == "" {
		log.Println("Отсутствует taskID в запросе результата")
		http.Error(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	st.Mu.Lock()
	defer st.Mu.Unlock()

	task, exists := st.Tasks[taskID]
	if !exists {
		log.Printf("Задача %s не найдена в Tasks", taskID)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if !task.Completed {
		log.Printf("Результат задачи %s ещё не готов: %+v", taskID, task)
		http.Error(w, "Task result not available", http.StatusNotFound)
	}

	log.Printf("Возвращён результат задачи %s: %f", taskID, task.Result)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Result float64 `json:"result"`
	}{Result: task.Result})
}
