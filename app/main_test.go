package main

import (
	"fmt"
	"net/http"
	"testing"
)

// // A simple example of unit testing a function.
// // Adapted from: https://gobyexample.com/testing-and-benchmarking

// func TestIntMinBasic(t *testing.T) {
// 	ans := IntMin(2, -2)
// 	if ans != -2 {
// 		t.Errorf("IntMin(2, -2) = %d; want -2", ans)
// 	}
// }

//	func TestIntMinTableDriven(t *testing.T) {
//		var tests = []struct {
//			a, b int
//			want int
//		}{
//			{0, 1, 0},
//			{1, 0, 0},
//			{2, -2, -2},
//			{0, -1, -1},
//			{-1, 0, -1},
//		}
//		for _, tt := range tests {
//			testname := fmt.Sprintf("%d,%d", tt.a, tt.b)
//			t.Run(testname, func(t *testing.T) {
//				ans := IntMin(tt.a, tt.b)
//				if ans != tt.want {
//					t.Errorf("got %d, want %d", ans, tt.want)
//				}
//			})
//		}
//	}
//

// TestValidateSalary tests the ValidateSalary function.
// func TestValidateSalary(t *testing.T) {
// 	var tests = []struct {
// 		salary float64
// 		valid  bool
// 	}{
// 		{50000, true},
// 		{100000, true},
// 		{200000, true},
// 		{300000, true},
// 		{400000, true},
// 		{0, false},
// 		{-1000, false},
// 	}

// 	for _, tt := range tests {
// 		testname := fmt.Sprintf("Salary: %.2f", tt.salary)
// 		t.Run(testname, func(t *testing.T) {
// 			err := ValidateSalary(tt.salary)
// 			if err != nil {
// 				t.Errorf("got %v, want %v", valid, tt.valid)
// 			}
// 		})
// 	}
// }

// TestGetTaxCalculatorInstructionsByYear tests the GetTaxCalculatorInstructionsByYear function.
func TestGetTaxCalculatorInstructionsByYear(t *testing.T) {
	var tests = []struct {
		year  string
		want  []TaxBracket
		error *Err
	}{
		{
			"2022",
			[]TaxBracket{
				{
					Min:  0,
					Max:  50197,
					Rate: 0.15,
				},
				{
					Min:  50197,
					Max:  100392,
					Rate: 0.205,
				},
				{
					Min:  100392,
					Max:  155625,
					Rate: 0.26,
				},
				{
					Min:  155625,
					Max:  221708,
					Rate: 0.29,
				},
				{
					Min:  221708,
					Rate: 0.33,
				},
			},
			nil,
		},
		{
			"2018",
			[]TaxBracket{},
			&Err{
				Code:    http.StatusNotFound,
				Field:   "tax-year",
				Message: "tax brackets for the tax year 2018 is not found",
			},
		},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%v", tt.year)
		t.Run(testname, func(t *testing.T) {
			ans, err := GetTaxCalculatorInstructionsByYear(tt.year)
			if err == nil && tt.error != nil {
				t.Errorf("got error %v, want error %v", "nil", *tt.error)
			}
			if err != nil && *tt.error != *err {
				t.Errorf("got error %v, want error %v", *err, *tt.error)
			}
			for i, v := range ans {
				if v != tt.want[i] {
					t.Errorf("i: %d got %v, want %v", i, ans[i], tt.want[i])
				}
			}
		})
	}
}
