package calculations

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"unicode"

	"github.com/pAran0k/calc_go/models"
)

var operators = map[string]int{
	"+": 2,
	"-": 2,
	"*": 3,
	"/": 3,
	"(": 1,
}

func tokenize(expression string) ([]string, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	var tokens []string
	var currentNumber string
	for _, token := range expression {
		if unicode.IsDigit(token) || token == '.' && currentNumber != "" {
			currentNumber += string(token)
			continue
		}
		if currentNumber != "" {
			tokens = append(tokens, currentNumber)
			currentNumber = ""
		}
		if strings.Contains("+-*/()", string(token)) {
			tokens = append(tokens, string(token))
		} else if !(unicode.IsDigit(token) || token == '.' && currentNumber != "") {
			return nil, ErrInvalidExpression
		}
	}
	if currentNumber != "" {
		tokens = append(tokens, currentNumber)
	}
	return tokens, nil
}
func ToRPN(expression string) (string, error) {
	tokens, err := tokenize(expression)
	if err != nil {
		return "", err
	}
	var out []string
	var stack []string
	for _, token := range tokens {
		if _, err := strconv.ParseFloat(token, 64); err == nil {
			out = append(out, token)
		} else if token == "(" {
			stack = append(stack, token)
		} else if token == ")" {
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				out = append(out, stack[len(stack)-1])
				stack = stack[:len(stack)-1]

			}
			if len(stack) == 0 || stack[len(stack)-1] != "(" {
				return "", ErrInvalidExpression
			}
			stack = stack[:len(stack)-1]
		} else {
			if len(stack) == 0 || operators[token] > operators[stack[len(stack)-1]] {
				stack = append(stack, token)
			} else {
				for {
					if len(stack) == 0 || operators[token] > operators[stack[len(stack)-1]] {
						break
					}
					out = append(out, stack[len(stack)-1])
					stack = stack[:len(stack)-1]
				}
				stack = append(stack, token)

			}
		}

	}
	if len(stack) != 0 {
		for len(stack) > 0 {
			if stack[len(stack)-1] == "(" {
				return "", ErrInvalidExpression
			}
			out = append(out, stack[len(stack)-1])
			stack = stack[:len(stack)-1]
		}
	}
	return strings.Join(out, " "), nil
}
func ParseRPN(rpn string) (*models.Node, error) {
	if rpn == "" {
		return nil, ErrEmptyExpression
	}

	tokens := strings.Fields(rpn)
	stack := make([]*models.Node, 0)

	for _, token := range tokens {
		if IsOperator(token) {
			if len(stack) < 2 {
				return nil, ErrInvalidRpn
			}
			right := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			left := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			node := &models.Node{Value: token, Left: left, Right: right}
			stack = append(stack, node)
		} else {
			num, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return nil, ErrInvalidSymbol
			}
			node := &models.Node{Value: fmt.Sprintf("%f", num)}
			stack = append(stack, node)
		}
	}

	if len(stack) != 1 {
		return nil, fmt.Errorf("invalid RPN expression")
	}

	return stack[0], nil
}

func IsOperator(token string) bool {
	return token == "+" || token == "-" || token == "*" || token == "/"
}

func BuildTasks(exprID string, root *models.Node) ([]models.Task, error) {
	if root == nil {
		return nil, ErrEmptyExpression
	}

	var tasks []models.Task
	var taskCounter int

	var buildTask func(node *models.Node) (string, error)
	buildTask = func(node *models.Node) (string, error) {
		if node == nil {
			return "", nil
		}

		if !IsOperator(node.Value) {
			return node.Value, nil
		}

		if node.Value == "/" {
			if rightNum, err := strconv.ParseFloat(node.Right.Value, 64); err == nil && rightNum == 0 {
				return "", ErrDivisionByZero
			}
		}

		leftArg, err := buildTask(node.Left)
		if err != nil {
			return "", err
		}
		rightArg, err := buildTask(node.Right)
		if err != nil {
			return "", err
		}

		taskID := fmt.Sprintf("task-%s-%d", exprID, taskCounter)
		taskCounter++

		task := models.Task{
			ID:        taskID,
			Arg1:      leftArg,
			Arg2:      rightArg,
			Operation: node.Value,
			Completed: false,
		}
		tasks = append(tasks, task)
		return taskID, nil
	}

	_, err := buildTask(root)
	if err != nil {
		return nil, err
	}

	for i, j := 0, len(tasks)-1; i < j; i, j = i+1, j-1 {
		tasks[i], tasks[j] = tasks[j], tasks[i]
	}
	log.Printf("Сформированы задачи для %s: %+v", exprID, tasks)
	return tasks, nil
}
