package validation

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/marlonnienaber/Factum/server/internal/shared"
	"roci.dev/fracdex"
)

type ValidationErrorDetails struct {
	Violations []ValidationViolation `json:"violations"`
}

type ValidationViolation struct {
	Field       string `json:"field"`
	Description string `json:"message"`
}

func RegisterJsonPropertyNames() {
	if validatorEngine, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validatorEngine.RegisterTagNameFunc(func(field reflect.StructField) string {
			name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
			return name
		})
	}
}

func RegisterRankValidation() {
	if validatorEngine, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := validatorEngine.RegisterValidation("rank", isValidRank); err != nil {
			panic("registering rank validation failed: " + err.Error())
		}
	}
}

func isValidRank(fieldLevel validator.FieldLevel) bool {
	_, err := fracdex.Float64Approx(fieldLevel.Field().String())
	return err == nil
}

func NewValidationErrorResponse(validationErrors validator.ValidationErrors) shared.DetailedErrorResponse[ValidationErrorDetails] {
	violations := make([]ValidationViolation, 0, len(validationErrors))

	for _, fieldError := range validationErrors {
		violations = append(violations, ValidationViolation{
			Field:       fieldError.Field(),
			Description: validationErrorMessage(fieldError),
		})
	}

	return shared.DetailedErrorResponse[ValidationErrorDetails]{
		ErrorCode: "validation_error",
		Message:   "Request body did not match the expected form.",
		Details: ValidationErrorDetails{
			Violations: violations,
		},
	}
}

func validationErrorMessage(fieldError validator.FieldError) string {
	field := fieldError.Field()
	switch fieldError.Tag() {
	case "email":
		return field + " must be a valid email address."
	case "min":
		return field + " must be at least " + fieldError.Param() + " characters long."
	case "max":
		return field + " must be at most " + fieldError.Param() + " characters long."
	case "datetime":
		switch fieldError.Param() {
		case "2006-01-02":
			return field + " must be a valid calendar date in YYYY-MM-DD format."
		default:
			return field + " does not follow the expected datetime format."
		}
	case "gt":
		return field + " must be greater than " + fieldError.Param() + "."
	case "rank":
		return field + " must be a valid fractional rank."
	default:
		return field + " is invalid."
	}
}
