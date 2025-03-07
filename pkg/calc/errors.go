package calculations

import "errors"

var (
	ErrDivisionByZero    = errors.New("division by zero")
	ErrEmptyExpression   = errors.New("expression is empty")
	ErrInvalidExpression = errors.New("expression is not valid")
	ErrInvalidRpn        = errors.New("invalid RPN expression")
	ErrInvalidSymbol     = errors.New("invalid symbol in expression")
)
