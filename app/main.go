package main

import (
	"context"
	"encoding/json"
	"net/http"
	"patrickyau/interview-test-server/api"
	"time"

	// Import the ginzerolog package

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-openapi/runtime/middleware"
	chimiddleware "github.com/oapi-codegen/nethttp-middleware"

	// "github.com/gin-gonic/gin"
	"github.com/go-chi/chi"
	// "github.com/go-openapi/runtime/middleware"

	// ginzerolog "github.com/dn365/gin-zerolog"
	// "github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	// Import the api package
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339 //zerolog.TimeFormatUnix

	// gin.SetMode(gin.DebugMode)
	// router := gin.New()

	// router.Use(ginzerolog.Logger("gin")) // Use the ginzerolog middleware
	// router.Use(gin.Recovery())           // to recover gin automatically

	// router.GET("/", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{"message": "Hello, this is the Interview Test Server!"})
	// })
	// router.GET("/health", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{"Status": "OK"})
	// })
	// router.GET("/tax-calculator", getTaxCalculatorInstructionsByYear)
	// router.GET("/tax-calculator/tax-year", getAllTaxCalculatorInstructions)
	// router.GET("/tax-calculator/tax-year/:year", getTaxCalculatorInstructionsByYear)
	// router.POST("/tax-calculator/tax-year/:year", postTaxCalculationsByYear)

	// router.Run(":8080")

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

	log.Info().Msgf("Server listening on", addr)
	err = httpServer.ListenAndServe()
	if err != nil {
		log.Fatal().Msgf("error: %v", err)
		return
	}
}

func NewServer(taxServiice *TaxService) api.StrictServerInterface {
	return &server{
		// ThingService: thingService,
	}
}

type server struct {
	// *ThingService
}

// Calculate implements api.StrictServerInterface.
func (s *server) Calculate(ctx context.Context, request api.CalculateRequestObject) (api.CalculateResponseObject, error) {
	panic("unimplemented")
}

// Check implements api.StrictServerInterface.
func (s *server) Check(ctx context.Context, request api.CheckRequestObject) (api.CheckResponseObject, error) {
	panic("unimplemented")
}

// GetAllTaxCalculator implements api.StrictServerInterface.
func (s *server) GetAllTaxCalculator(ctx context.Context, request api.GetAllTaxCalculatorRequestObject) (api.GetAllTaxCalculatorResponseObject, error) {
	panic("unimplemented")
}

// GetTaxCalculator implements api.StrictServerInterface.
func (s *server) GetTaxCalculator(ctx context.Context, request api.GetTaxCalculatorRequestObject) (api.GetTaxCalculatorResponseObject, error) {
	panic("unimplemented")
}

// GetTaxCalculatorByYear implements api.StrictServerInterface.
func (s *server) GetTaxCalculatorByYear(ctx context.Context, request api.GetTaxCalculatorByYearRequestObject) (api.GetTaxCalculatorByYearResponseObject, error) {
	panic("unimplemented")
}

type TaxService struct {
}

func NewTaxService() *TaxService {
	return &TaxService{}
}

func NewSecurityMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			scopes, ok := ctx.Value(api.ApiKeyScopes).([]string)
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

func (s *TaxService) Check(ctx context.Context, request api.CheckRequestObject) (api.CheckResponseObject, error) {
	return api.Check200JSONResponse{
		Status: "ok",
	}, nil
}

func (s *TaxService) GetTaxCalculater(ctx context.Context, request api.GetTaxCalculatorRequestObject) (api.GetTaxCalculatorResponseObject, error) {
	return nil, nil
}

func (s *TaxService) GetAllTaxCalculator(ctx context.Context, request api.GetAllTaxCalculatorRequestObject) (api.GetAllTaxCalculatorResponseObject, error) {
	return nil, nil
}

func (s *TaxService) GetTaxCalculatorByYear(ctx context.Context, request api.GetTaxCalculatorByYearRequestObject) (api.GetTaxCalculatorByYearResponseObject, error) {
	return nil, nil
}

func (s *TaxService) Calculate(ctx context.Context, request api.CalculateRequestObject) (api.CalculateResponseObject, error) {
	return nil, nil
}

// getTaxCalculatorInstructionsByYear responds with the tax calculator instructions by the year.
// If the year is not provided, it defaults to year 2022.
// func getTaxCalculatorInstructionsByYear(c *gin.Context) {
// 	year := c.Param("year")
// 	// gin.DebugPrintFunc("getTaxCalculatorInstructions called with year: %s\n", year)
// 	taxBrackets, err := GetTaxCalculatorInstructionsByYear(year)
// 	if err != nil {
// 		c.IndentedJSON(http.StatusNotFound, err)
// 		return
// 	}
// 	c.IndentedJSON(http.StatusOK, taxBrackets)
// }

// // getAllTaxCalculatorInstructions responds with the list of tax calculator instructions.
// func getAllTaxCalculatorInstructions(c *gin.Context) {
// 	c.IndentedJSON(http.StatusOK, TaxBrackets)
// }

// type Salary struct {
// 	Salary float64 `json:"salary"`
// }

// type Err struct {
// 	Code    uint   `json:"code"`
// 	Field   string `json:"field"`
// 	Message string `json:"message"`
// }

// type TaxOwed struct {
// 	EffectiveTaxRate string       `json:"effective_tax_rate"`
// 	Salary           float64      `json:"salary"`
// 	TaxYear          string       `json:"tax_year"`
// 	TaxOwnedPerBand  []TaxBracket `json:"tax_owned_per_band"`
// 	TotalTaxOwed     float64      `json:"total_tax_owed"`
// }

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

// func GetTaxCalculatorInstructionsByYear(year string) ([]TaxBracket, *Err) {
// 	if year == "" {
// 		year = "2022"
// 	}
// 	taxBrackets := TaxBrackets[year]
// 	if len(taxBrackets) == 0 {
// 		return nil, &Err{
// 			Code:    http.StatusNotFound,
// 			Field:   "tax-year",
// 			Message: fmt.Sprintf("tax brackets for the tax year %s is not found", year),
// 		}
// 	}
// 	return taxBrackets, nil
// }

// func ValidateSalary(salary float64) *Err {
// 	if salary < 0 {
// 		return &Err{
// 			Code:    http.StatusBadRequest,
// 			Field:   "salary",
// 			Message: fmt.Sprintf("the salary for the tax year must be greater than 0. Invalid value: %.2f", salary),
// 		}
// 	}
// 	return nil
// }
//
// func CalculateTaxAmount(year string, taxBrackets []TaxBracket, salary float64) TaxOwed {
// 	var taxAmount float64
// 	var taxPerBracket []TaxBracket
// 	totalTaxAmount := 0.0
// 	for _, bracket := range taxBrackets {
// 		if salary > bracket.Min {
// 			leftover := salary
// 			if bracket.Max > 0 && salary > bracket.Max {
// 				leftover = math.Min(salary, bracket.Max)
// 			}
// 			taxableIncome := leftover - bracket.Min
// 			taxAmount = taxableIncome * bracket.Rate
// 			totalTaxAmount += taxAmount
// 			bracket.TaxOwed = math.Round(taxAmount*100) / 100

// 			taxPerBracket = append(taxPerBracket, bracket)
// 		}
// 	}
// 	effectiveRate := totalTaxAmount / salary
// 	return TaxOwed{
// 		EffectiveTaxRate: fmt.Sprintf("%.2f", effectiveRate*100) + "%",
// 		Salary:           salary,
// 		TaxYear:          year,
// 		TaxOwnedPerBand:  taxPerBracket,
// 		TotalTaxOwed:     math.Round(totalTaxAmount*100) / 100,
// 	}
// }
