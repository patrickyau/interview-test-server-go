package main

// TaxBracket represents a tax bracket for a given year.
type TaxBracket struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Rate    float64 `json:"rate"`
	TaxOwed float64 `json:"tax_owed,omitempty"`
}

// TaxBrackets represents the tax brackets for all supported years.
var TaxBrackets = map[string][]TaxBracket{
	"2019": {
		{Min: 0,
			Max:  47630,
			Rate: 0.15,
		},
		{Min: 47630,
			Max:  95259,
			Rate: 0.205,
		},
		{
			Min:  95259,
			Max:  147667,
			Rate: 0.26,
		},
		{
			Min:  147667,
			Max:  210371,
			Rate: 0.29,
		},
		{
			Min:  210371,
			Rate: 0.33,
		},
	},
	"2020": {
		{
			Min:  0,
			Max:  48535,
			Rate: 0.15,
		},
		{
			Min:  48535,
			Max:  97069,
			Rate: 0.205,
		},
		{
			Min:  97069,
			Max:  150473,
			Rate: 0.26,
		},
		{
			Min:  150473,
			Max:  214368,
			Rate: 0.29,
		},
		{
			Min:  214368,
			Rate: 0.33,
		},
	},
	"2021": {
		{
			Min:  0,
			Max:  49020,
			Rate: 0.15,
		},
		{
			Min:  49020,
			Max:  98040,
			Rate: 0.205,
		},
		{
			Min:  98040,
			Max:  151978,
			Rate: 0.26,
		},
		{
			Min:  151978,
			Max:  216511,
			Rate: 0.29,
		},
		{
			Min:  216511,
			Rate: 0.33,
		},
	},
	"2022": {
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
	"2023": {
		{
			Min:  0,
			Max:  53359,
			Rate: 0.15,
		},
		{
			Min:  53359,
			Max:  106717,
			Rate: 0.205,
		},
		{
			Min:  106717,
			Max:  165430,
			Rate: 0.26,
		},
		{
			Min:  165430,
			Max:  235675,
			Rate: 0.29,
		},
		{
			Min:  235675,
			Rate: 0.33,
		},
	},
}
