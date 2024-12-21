package calculations_test

import (
	"testing"

	calculations "github.com/pAran0k/calc_go/pkg"
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
			res, err := calculations.Calc(testCase.expression)
			if err != nil {
				t.Fatalf("expected %f error %v", testCase.expected, err)
			} else if res != testCase.expected {
				t.Errorf("expected %f returns %v", testCase.expected, res)
			}

		})

	}
	testCasesFail := []string{
		"1/0",
		"2+2*",
		"()",
		"10+2*2/(10-2))",
		"-",
		"qwerty",
		"",
	}
	for _, testCase := range testCasesFail {
		t.Run(testCase, func(t *testing.T) {
			_, err := calculations.Calc(testCase)
			if err == nil {
				t.Errorf("calc %s error not nil", testCase)
			}
		})
	}

}
