package calculation_test

import (
	"slices"
	"testing"

	"github.com/MoodyShoo/go-http-calculator/pkg/calculation"
)

func TestShuntingYard(t *testing.T) {
	cases := []struct {
		name       string
		expression string
		want       []string
		wantErr    bool
	}{
		{
			name:       "Valid TwoSum",
			expression: "2+2",
			want:       []string{"2", "2", "+"},
			wantErr:    false,
		},
		{
			name:       "Valid Expression with Priority",
			expression: "2+3*4",
			want:       []string{"2", "3", "4", "*", "+"},
			wantErr:    false,
		},
		{
			name:       "Valid Expression with Minus",
			expression: "-2+1",
			want:       []string{"-2", "1", "+"},
			wantErr:    false,
		},
		{
			name:       "Valid Expression with Parentheses",
			expression: "(2+3)*4",
			want:       []string{"2", "3", "+", "4", "*"},
			wantErr:    false,
		},
		{
			name:       "Invalid Expression (Mismatched Parentheses)",
			expression: "2+(3*4",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "Invalid Expression (Unknown Character)",
			expression: "2+3$4",
			want:       nil,
			wantErr:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := calculation.ShuntingYard(tc.expression)

			if (err != nil) != tc.wantErr {
				t.Errorf("ShuntingYard() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr && !slices.Equal(got, tc.want) {
				t.Errorf("ShuntingYard() = %v, want %v", got, tc.want)
			}
		})
	}
}
