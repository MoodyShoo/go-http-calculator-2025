package calculation

import (
	"fmt"
	"strings"
	"unicode"
)

func isDigitOrDot(symbol rune) bool {
	return unicode.IsDigit(symbol) || symbol == '.'
}

func buildNumber(expression string, i *int) string {
	var num strings.Builder
	if expression[*i] == '-' {
		num.WriteByte('-')
		*i++
	}

	for *i < len(expression) && (unicode.IsDigit(rune(expression[*i])) || expression[*i] == '.') {
		num.WriteByte(expression[*i])
		*i++
	}

	*i--

	return num.String()
}

func ShuntingYard(expression string) ([]string, error) {
	var out []string
	var operators []string

	for i := 0; i < len(expression); i++ {
		current := rune(expression[i])

		if unicode.IsSpace(current) {
			continue
		}

		if isDigitOrDot(current) || (current == '-' && (i == 0 || expression[i-1] == '(')) {
			out = append(out, buildNumber(expression, &i))
		} else if IsOperator(current) {
			for len(operators) > 0 && IsOperator(rune(operators[len(operators)-1][0])) &&
				precedence(rune(operators[len(operators)-1][0])) >= precedence(current) {
				out = append(out, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}

			operators = append(operators, string(current))
		} else if current == '(' {
			operators = append(operators, string(current))
		} else if current == ')' {
			for len(operators) > 0 && operators[len(operators)-1] != "(" {
				out = append(out, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}

			if len(operators) == 0 {
				return nil, fmt.Errorf("mismatched parentheses")
			}

			operators = operators[:len(operators)-1]
		} else {
			return nil, fmt.Errorf("invalid character: %c", current)
		}
	}

	for len(operators) > 0 {
		if operators[len(operators)-1] == "(" || operators[len(operators)-1] == ")" {
			return nil, fmt.Errorf("mismatched parentheses")
		}

		out = append(out, operators[len(operators)-1])
		operators = operators[:len(operators)-1]
	}

	return out, nil
}
