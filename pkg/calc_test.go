package calculations_test

import (
	"testing"
)

func TestCalc(t *testing.T) {
	testCasesSuccess := []struct {
		expression string
		expected   float64
	}{
		{
			expression: "2+2",
			expected:   4,
		},
		{
			expression: "(2+2)*2",
			expected:   8,
		},
		{
			expression: "1/2",
			expected:   0.5,
		},
		{
			expression: "20+20/0.5*(3+2)",
			expected:   220,
		},
	}
	for _, testCase := range testCasesSuccess {
		t.Run(testCase.expression, func(t *testing.T) {
			val, err := Calc(TestCase.expression)

		})

	}
}
