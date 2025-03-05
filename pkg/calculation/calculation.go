package calculation

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

func IsOperator(r rune) bool {
	return r == '+' || r == '-' || r == '/' || r == '*'
}

func IsValidFormula(expression string) bool {
	prevWasOperator := false
	stack := 0

	for _, r := range expression {
		switch {
		case unicode.IsDigit(r) || r == '.':
			prevWasOperator = false
		case r == '(':
			stack++
			prevWasOperator = true
		case r == ')':
			if stack == 0 {
				return false
			}
			stack--
			prevWasOperator = false
		case IsOperator(r):
			if prevWasOperator {
				return false
			}
			prevWasOperator = true
		case r == ' ':
			continue
		default:
			return false
		}
	}

	return stack == 0 && !prevWasOperator
}

func applyOperation(numbers_stack *[]float64, operator rune) error {
	if len(*numbers_stack) < 2 {
		return errors.New("недостаточно чисел для операции")
	}

	b := (*numbers_stack)[len(*numbers_stack)-1]
	a := (*numbers_stack)[len(*numbers_stack)-2]
	*numbers_stack = (*numbers_stack)[:len(*numbers_stack)-2]

	var result float64
	switch operator {
	case '+':
		result = a + b
	case '-':
		result = a - b
	case '*':
		result = a * b
	case '/':
		if b == 0 {
			return errors.New("деление на ноль")
		}
		result = a / b
	}

	*numbers_stack = append(*numbers_stack, result)
	return nil
}

func precedence(op rune) int {
	switch op {
	case '+', '-':
		return 1
	case '*', '/':
		return 2
	}
	return 0
}

func Calc(expression string) (float64, error) {
	expression = strings.ReplaceAll(expression, "\r", "")
	expression = strings.ReplaceAll(expression, "\n", "")
	trimmed := strings.ReplaceAll(expression, " ", "")

	if !IsValidFormula(trimmed) {
		return 0.0, errors.New("некорректная формула")
	}

	var numbers []float64
	var operators []rune
	var buffer []rune

	for _, r := range trimmed {
		switch {
		case unicode.IsDigit(r) || r == '.':
			buffer = append(buffer, r)
		case IsOperator(r):
			if len(buffer) > 0 {
				num, err := strconv.ParseFloat(string(buffer), 64)
				if err != nil {
					return 0.0, errors.New("ошибка преобразования числа")
				}
				numbers = append(numbers, num)
				buffer = buffer[:0]
			}
			for len(operators) > 0 && precedence(operators[len(operators)-1]) >= precedence(r) {
				if operators[len(operators)-1] == '(' {
					break
				}
				if err := applyOperation(&numbers, operators[len(operators)-1]); err != nil {
					return 0.0, err
				}
				operators = operators[:len(operators)-1]
			}
			operators = append(operators, r)
		case r == '(':
			operators = append(operators, r)
		case r == ')':
			if len(buffer) > 0 {
				num, err := strconv.ParseFloat(string(buffer), 64)
				if err != nil {
					return 0.0, errors.New("ошибка преобразования числа")
				}
				numbers = append(numbers, num)
				buffer = buffer[:0]
			}
			for len(operators) > 0 && operators[len(operators)-1] != '(' {
				if err := applyOperation(&numbers, operators[len(operators)-1]); err != nil {
					return 0.0, err
				}
				operators = operators[:len(operators)-1]
			}
			operators = operators[:len(operators)-1]
		}
	}

	if len(buffer) > 0 {
		num, err := strconv.ParseFloat(string(buffer), 64)
		if err != nil {
			return 0.0, errors.New("ошибка преобразования числа")
		}
		numbers = append(numbers, num)
	}

	for len(operators) > 0 {
		if err := applyOperation(&numbers, operators[len(operators)-1]); err != nil {
			return 0.0, err
		}
		operators = operators[:len(operators)-1]
	}

	if len(numbers) != 1 {
		return 0.0, errors.New("ошибка вычислений")
	}

	return numbers[0], nil
}
