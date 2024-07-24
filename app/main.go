package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"patrickyau/interview-test-server/api"
	"strconv"
	"time"

	// Import the ginzerolog package

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-openapi/runtime/middleware"
	chimiddleware "github.com/oapi-codegen/nethttp-middleware"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339 //zerolog.TimeFormatUnix

	service := NewTaxService()
	s := NewServer(service)
	swagger, err := api.GetSwagger()
	if err != nil {
		log.Fatal().Msgf("error: %v", err)
	}

	router := chi.NewRouter()

	// Add swagger UI endpoints
	router.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(swagger)
	})
	router.Handle("/swagger/", middleware.SwaggerUI(middleware.SwaggerUIOpts{
		Path:    "/swagger/",
		SpecURL: "/swagger/doc.json",
	}, nil))

	// Enable validation of incoming requests
	validator := chimiddleware.OapiRequestValidatorWithOptions(
		swagger,
		&chimiddleware.Options{
			Options: openapi3filter.Options{
				AuthenticationFunc: func(c context.Context, input *openapi3filter.AuthenticationInput) error {
					return nil
				},
			},
		},
	)

	securityMiddleware := NewSecurityMiddleware()

	apiServer := api.HandlerWithOptions(
		api.NewStrictHandler(s, nil),
		api.ChiServerOptions{
			BaseURL:    "/tax-calculator",
			BaseRouter: router,
			Middlewares: []api.MiddlewareFunc{
				securityMiddleware,
				validator,
			},
		},
	)

	addr := ":8080"
	httpServer := http.Server{
		Addr:    addr,
		Handler: apiServer,
	}

	log.Info().Msgf("Server listening on %v", addr)
	err = httpServer.ListenAndServe()
	if err != nil {
		log.Fatal().Msgf("error: %v", err)
		return
	}
}

func NewServer(taxService *TaxService) api.StrictServerInterface {
	return &server{
		TaxService: taxService,
	}
}

type TaxService struct {
}

func NewTaxService() *TaxService {
	return &TaxService{}
}

type server struct {
	*TaxService
}

func (s *TaxService) Check(ctx context.Context, request api.CheckRequestObject) (api.CheckResponseObject, error) {
	return api.Check200JSONResponse{
		Status: "ok",
	}, nil
}

func (s *TaxService) GetTaxCalculator(ctx context.Context, request api.GetTaxCalculatorRequestObject) (api.GetTaxCalculatorResponseObject, error) {
	taxBrackets, err := GetTaxCalculatorInstructionsByYear("")
	if err != nil {
		// c.IndentedJSON(http.StatusNotFound, err)
		return nil, fmt.Errorf("error: %v", err)
	}
	var response api.GetTaxCalculator200JSONResponse
	for _, bracket := range taxBrackets {
		response = append(response, mapTaxBracketToAPITaxBracket(bracket))
	}
	return response, nil
}

func mapTaxBracketToAPITaxBracket(taxBracket TaxBracket) api.TaxBracket {
	apiTaxBracket := api.TaxBracket{
		Min:  taxBracket.Min,
		Max:  taxBracket.Max,
		Rate: taxBracket.Rate,
	}
	if taxBracket.TaxOwed != 0.0 {
		apiTaxBracket.TaxOwed = taxBracket.TaxOwed
	}
	return apiTaxBracket
}

func (s *TaxService) GetTaxCalculatorByYear(ctx context.Context, request api.GetTaxCalculatorByYearRequestObject) (api.GetTaxCalculatorByYearResponseObject, error) {

	if _, err := strconv.Atoi(request.Year); err != nil {
		return api.GetTaxCalculatorByYear400Response{}, nil
	}
	var response api.GetTaxCalculatorByYear200JSONResponse
	taxBrackets, err := GetTaxCalculatorInstructionsByYear(request.Year)
	if err != nil {
		return api.GetTaxCalculatorByYear404Response{}, nil
	}

	for _, bracket := range taxBrackets {
		response = append(response, mapTaxBracketToAPITaxBracket(bracket))
	}
	return response, nil
}

func mapTaxBracketsToAPITaxBrackets(taxBrackets []TaxBracket) []api.TaxBracket {
	apiTaxBrackets := make([]api.TaxBracket, len(taxBrackets))
	for i, bracket := range taxBrackets {
		apiTaxBrackets[i] = mapTaxBracketToAPITaxBracket(bracket)
	}
	return apiTaxBrackets
}

func (s *TaxService) GetAllTaxCalculator(ctx context.Context, request api.GetAllTaxCalculatorRequestObject) (api.GetAllTaxCalculatorResponseObject, error) {
	response := api.GetAllTaxCalculator200JSONResponse{}
	for year, taxBrackets := range TaxBrackets {
		response[year] = mapTaxBracketsToAPITaxBrackets(taxBrackets)
	}
	return response, nil
}

func (s *TaxService) Calculate(ctx context.Context, request api.CalculateRequestObject) (api.CalculateResponseObject, error) {
	salary := request.Body.Salary

	err := ValidateSalary(salary)
	if err != nil {
		return api.Calculate400Response{}, nil
	}

	year := request.Year
	log.Debug().Msgf("postTaxCalculationsByYear() year: %v, salary: %.2f", year, salary)
	taxBrackets, err := GetTaxCalculatorInstructionsByYear(year)
	if err != nil {
		return api.Calculate400Response{}, nil
	}

	taxOwed := CalculateTaxAmount(year, taxBrackets, salary)
	return api.Calculate200JSONResponse{
		TaxYear:          taxOwed.TaxYear,
		Salary:           taxOwed.Salary,
		EffectiveTaxRate: taxOwed.EffectiveTaxRate,
		TotalTaxOwed:     taxOwed.TotalTaxOwed,
		TaxOwedPerBand:   mapTaxBracketsToAPITaxBrackets(taxOwed.TaxOwnedPerBand),
	}, nil
}

func NewSecurityMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// const (
			// 	ApiKeyScopes = "apiKey.Scopes"
			// )
			scopes, ok := ctx.Value("apiKey.Scopes").([]string)
			if !ok {
				// no scopes required for this endpoint, no X-Api-Key required
				next.ServeHTTP(w, r)
				return
			}

			apiKey := r.Header.Get("X-Api-Key")
			if apiKey == "" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("header X-Api-Key not provided"))
				return
			}

			if apiKey != "test" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("invalid api key provided"))
				return
			}

			// This is where you check if api key has the required scope
			_, _ = apiKey, scopes

			next.ServeHTTP(w, r)
		})
	}
}

type Salary struct {
	Salary float64 `json:"salary"`
}

type Err struct {
	Code    uint   `json:"code"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

type TaxOwed struct {
	EffectiveTaxRate string       `json:"effective_tax_rate"`
	Salary           float32      `json:"salary"`
	TaxYear          string       `json:"tax_year"`
	TaxOwnedPerBand  []TaxBracket `json:"tax_owned_per_band"`
	TotalTaxOwed     float32      `json:"total_tax_owed"`
}

// // postTaxCalculationsByYear calculates the tax for the year from JSON received in the request body.
// func postTaxCalculationsByYear(c *gin.Context) {
// 	var newSalary Salary

// 	log.Debug().Msg("postTaxCalculationsByYear called")
// 	// Call BindJSON to bind the received JSON to newSalary.
// 	if err := c.BindJSON(&newSalary); err != nil {
// 		c.IndentedJSON(http.StatusBadRequest, gin.H{
// 			"code":    http.StatusBadRequest,
// 			"field":   "salary",
// 			"message": fmt.Errorf("the salary for the tax year is not found"),
// 		})
// 		return
// 	}

// 	year := c.Param("year")
// 	log.Debug().Msgf("postTaxCalculationsByYear() year: %v, salary: %.2f", year, newSalary.Salary)
// 	taxBrackets, err := GetTaxCalculatorInstructionsByYear(year)
// 	if err != nil {
// 		c.IndentedJSON(http.StatusNotFound, gin.H{
// 			"code":    err.Code,
// 			"field":   err.Field,
// 			"message": err.Message,
// 		})
// 		return
// 	}
// 	err = ValidateSalary(newSalary.Salary)
// 	if err != nil {
// 		c.IndentedJSON(http.StatusBadRequest, gin.H{
// 			"code":    err.Code,
// 			"field":   err.Field,
// 			"message": err.Message,
// 		})
// 		return
// 	}

// 	taxOwed := CalculateTaxAmount(year, taxBrackets, newSalary.Salary)
// 	c.IndentedJSON(http.StatusOK, gin.H{
// 		"effective_tax_rate": taxOwed.EffectiveTaxRate,
// 		"salary":             taxOwed.Salary,
// 		"tax_year":           taxOwed.TaxYear,
// 		"tax_owned_per_band": taxOwed.TaxOwnedPerBand,
// 		"total_tax_owed":     taxOwed.TotalTaxOwed,
// 	})
// }

func GetTaxCalculatorInstructionsByYear(year string) ([]TaxBracket, *Err) {
	if year == "" {
		year = "2022"
	}
	taxBrackets := TaxBrackets[year]
	if len(taxBrackets) == 0 {
		return nil, &Err{
			Code:    http.StatusNotFound,
			Field:   "year",
			Message: fmt.Sprintf("tax brackets for the tax year %v is not found", year),
		}
	}
	return taxBrackets, nil
}

// func ValidateYear(year string) error {
// 	if salary <= 0.0 {
// 		return fmt.Errorf("the salary for the tax year must be greater than 0. Invalid value: %.2f", salary)
// 	}
// 	return nil
// }

func ValidateSalary(salary float32) *Err {
	if salary < 0 {
		return &Err{
			Code:    http.StatusBadRequest,
			Field:   "salary",
			Message: fmt.Sprintf("the salary for the tax year must be greater than 0. Invalid value: %.2f", salary),
		}
	}
	return nil
}

func CalculateTaxAmount(year string, taxBrackets []TaxBracket, salary float32) TaxOwed {
	var taxAmount float32
	var taxPerBracket []TaxBracket
	totalTaxAmount := float32(0.0)
	for _, bracket := range taxBrackets {
		if salary > bracket.Min {
			leftover := salary
			if bracket.Max > 0 && salary > bracket.Max {
				leftover = float32(math.Min(float64(salary), float64(bracket.Max)))
			}
			taxableIncome := leftover - bracket.Min
			taxAmount = taxableIncome * bracket.Rate
			totalTaxAmount += taxAmount
			bracket.TaxOwed = float32(math.Round(float64(taxAmount*100)) / 100)

			taxPerBracket = append(taxPerBracket, bracket)
		}
	}
	effectiveRate := totalTaxAmount / salary
	return TaxOwed{
		EffectiveTaxRate: fmt.Sprintf("%.2f", effectiveRate*100) + "%",
		Salary:           salary,
		TaxYear:          year,
		TaxOwnedPerBand:  taxPerBracket,
		TotalTaxOwed:     float32(math.Round(float64(totalTaxAmount)*100) / 100),
	}
}
