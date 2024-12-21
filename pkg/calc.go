package calculations

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
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
func toRPN(tokens []string) ([]string, error) {
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
				return nil, ErrInvalidExpression
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
				return nil, ErrInvalidExpression
			}
			out = append(out, stack[len(stack)-1])
			stack = stack[:len(stack)-1]
		}
	}
	return out, nil
}
func Calc(expression string) (float64, error) {
	if len(expression) == 0 {
		return 0, ErrEmptyExpression
	}
	expr, err := tokenize(expression)
	if err != nil {
		return 0, err
	}
	rpn, err := toRPN(expr)
	if err != nil {
		return 0, err
	}
	stack := []float64{}
	for i := 0; i < len(rpn); i++ {
		if a, err := strconv.ParseFloat(rpn[i], 64); err == nil {
			stack = append(stack, a)

		} else if _, ok := operators[rpn[i]]; ok {
			if len(stack) < 2 {
				return 0, ErrInvalidExpression
			}
			f := stack[len(stack)-2]
			s := stack[len(stack)-1]
			stack = stack[:len(stack)-2]
			temp := 0.0
			switch rpn[i] {
			case "+":
				temp = f + s
			case "-":
				temp = f - s
			case "*":
				temp = f * s
			case "/":
				if s == 0 {
					return 0, errors.New("division by 0")
				}
				temp = f / s
			}

			stack = append(stack, temp)
		}
	}
	if len(stack) != 1 {
		return 0, ErrInvalidExpression
	}
	return stack[0], nil
}
