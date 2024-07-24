package main

import (
	"context"
	"encoding/json"
	"errors"
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
	chimiddleware2 "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
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
	// chimiddleware2.Logger.WithLogger(zerolog.New(os.Stdout).With().Timestamp().Logger())
	router.Use(chimiddleware2.Logger)
	router.Use(chimiddleware2.RequestID)
	router.Use(chimiddleware2.Timeout(60 * time.Second))
	router.Use(chimiddleware2.URLFormat)
	router.Use(render.SetContentType(render.ContentTypeJSON))
	// logger := httplog.NewLogger("httplog-example", httplog.Options{
	// 	// JSON:             true,
	// 	LogLevel:         slog.LevelDebug,
	// 	Concise:          true,
	// 	RequestHeaders:   true,
	// 	MessageFieldName: "message",
	// 	// TimeFieldFormat: time.RFC850,
	// 	Tags: map[string]string{
	// 		"version": "v1.0-81aa4244d9fc8076a",
	// 		"env":     "dev",
	// 	},
	// 	QuietDownRoutes: []string{
	// 		"/",
	// 		"/ping",
	// 	},
	// 	QuietDownPeriod: 10 * time.Second,
	// 	// SourceFieldName: "source",
	// })
	// router.Use(httplog.RequestLogger(logger))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to PY's Tax Calculator API"))
	})

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
			ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				oplog := httplog.LogEntry(r.Context())
				oplog.Error("msg here", "err", errors.New("err here"))
				w.WriteHeader(500)
				w.Write([]byte("oops, err"))
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

// Check returns a 200 status code if the service is up.
func (s *TaxService) Check(ctx context.Context, request api.CheckRequestObject) (api.CheckResponseObject, error) {
	return api.Check200JSONResponse{
		Status: "ok",
	}, nil
}

// GetTaxCalculator returns the tax brackets for the default year 2022.
func (s *TaxService) GetTaxCalculator(ctx context.Context, request api.GetTaxCalculatorRequestObject) (api.GetTaxCalculatorResponseObject, error) {
	taxBrackets, err := GetTaxCalculatorInstructionsByYear("")
	if err != nil {
		// c.IndentedJSON(http.StatusNotFound, err)
		return api.GetTaxCalculator400JSONResponse{
			Code:    err.Code,
			Field:   err.Field,
			Message: err.Message,
		}, nil
	}
	var response api.GetTaxCalculator200JSONResponse
	for _, bracket := range taxBrackets {
		response = append(response, mapTaxBracketToAPITaxBracket(bracket))
	}
	return response, nil
}

// mapTaxBracketToAPITaxBracket maps a TaxBracket to an api.TaxBracket.
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

// GetTaxCalculatorByYear returns the tax brackets for the given year.
func (s *TaxService) GetTaxCalculatorByYear(ctx context.Context, request api.GetTaxCalculatorByYearRequestObject) (api.GetTaxCalculatorByYearResponseObject, error) {

	if err := ValidateYear(request.Year); err != nil {
		return api.GetTaxCalculatorByYear400JSONResponse{
			Code:    http.StatusBadRequest,
			Field:   "year",
			Message: fmt.Sprintf("the tax year %v is not a valid year", request.Year),
		}, nil
	}

	taxBrackets, err := GetTaxCalculatorInstructionsByYear(request.Year)
	if err != nil {
		return api.GetTaxCalculatorByYear404JSONResponse{
			Code:    err.Code,
			Field:   err.Field,
			Message: err.Message,
		}, nil
	}

	var response api.GetTaxCalculatorByYear200JSONResponse
	for _, bracket := range taxBrackets {
		response = append(response, mapTaxBracketToAPITaxBracket(bracket))
	}
	return response, nil
}

// mapTaxBracketsToAPITaxBrackets maps a slice of TaxBracket to a slice of api.TaxBracket.
func mapTaxBracketsToAPITaxBrackets(taxBrackets []TaxBracket) []api.TaxBracket {
	apiTaxBrackets := make([]api.TaxBracket, len(taxBrackets))
	for i, bracket := range taxBrackets {
		apiTaxBrackets[i] = mapTaxBracketToAPITaxBracket(bracket)
	}
	return apiTaxBrackets
}

// GetAllTaxCalculator returns all tax brackets for all supported years.
func (s *TaxService) GetAllTaxCalculator(ctx context.Context, request api.GetAllTaxCalculatorRequestObject) (api.GetAllTaxCalculatorResponseObject, error) {
	response := api.GetAllTaxCalculator200JSONResponse{}
	for year, taxBrackets := range TaxBrackets {
		response[year] = mapTaxBracketsToAPITaxBrackets(taxBrackets)
	}
	return response, nil
}

// Calculate calculates the tax for the year from JSON received in the request body.
func (s *TaxService) Calculate(ctx context.Context, request api.CalculateRequestObject) (api.CalculateResponseObject, error) {
	salary := request.Body.Salary
	err := ValidateSalary(salary)
	if err != nil {
		return api.Calculate400JSONResponse{
			Code:    err.Code,
			Field:   err.Field,
			Message: err.Message,
		}, nil
	}

	year := request.Year
	if err := ValidateYear(year); err != nil {
		return api.Calculate400JSONResponse{
			Code:    http.StatusBadRequest,
			Field:   "year",
			Message: fmt.Sprintf("the tax year %v is not a valid year", request.Year),
		}, nil
	}

	taxBrackets, err := GetTaxCalculatorInstructionsByYear(year)
	if err != nil {
		return api.Calculate400JSONResponse{
			Code:    err.Code,
			Field:   err.Field,
			Message: err.Message,
		}, nil
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

// NewSecurityMiddleware returns a new security middleware.
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

type Err struct {
	Code    int    `json:"code"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

// TaxOwed represents the tax owed for a given year.
type TaxOwed struct {
	EffectiveTaxRate string       `json:"effective_tax_rate"`
	Salary           float32      `json:"salary"`
	TaxYear          string       `json:"tax_year"`
	TaxOwnedPerBand  []TaxBracket `json:"tax_owned_per_band"`
	TotalTaxOwed     float32      `json:"total_tax_owed"`
}

// TaxBracket returns the tax bracket for a given year.
func GetTaxCalculatorInstructionsByYear(year string) ([]TaxBracket, *Err) {
	if year == "" {
		year = "2022"
	}
	taxBrackets := TaxBrackets[year]
	if len(taxBrackets) == 0 {
		return nil, &Err{
			Code:    http.StatusNotFound,
			Field:   "year",
			Message: fmt.Sprintf("tax brackets for the tax year '%v' is not found", year),
		}
	}
	return taxBrackets, nil
}

// ValidateYear validates the year for the tax year.
func ValidateYear(year string) *Err {
	if _, err := strconv.Atoi(year); err != nil {
		return &Err{
			Code:    http.StatusBadRequest,
			Field:   "year",
			Message: fmt.Sprintf("the tax year %v is not a valid year", year),
		}
	}
	return nil
}

// ValidateSalary validates the salary for the tax year.
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

// CalculateTaxAmount calculates the tax owed for a given year based on the tax brackets and salary.
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
