package config

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema.json
var schemaJSON string

// validateJSONSchema validates JSON configuration against the ARM schema
func validateJSONSchema(jsonData []byte) error {
	schemaLoader := gojsonschema.NewStringLoader(schemaJSON)
	documentLoader := gojsonschema.NewBytesLoader(jsonData)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, formatValidationError(desc))
		}
		return fmt.Errorf("configuration validation failed:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// formatValidationError formats a validation error with helpful context
func formatValidationError(err gojsonschema.ResultError) string {
	field := err.Field()
	if field == "(root)" {
		field = "configuration"
	}

	switch err.Type() {
	case "required":
		return fmt.Sprintf("  • Missing required field '%s' in %s", err.Details()["property"], field)
	case "invalid_type":
		expected := err.Details()["expected"]
		actual := err.Details()["given"]
		return fmt.Sprintf("  • Field '%s' has invalid type: expected %s, got %s", field, expected, actual)
	case "enum":
		allowed := err.Details()["allowed"]
		return fmt.Sprintf("  • Field '%s' has invalid value: must be one of %v", field, allowed)
	case "pattern":
		pattern := err.Details()["pattern"]
		return fmt.Sprintf("  • Field '%s' format is invalid: must match pattern %s", field, pattern)
	case "minimum":
		minimum := err.Details()["min"]
		return fmt.Sprintf("  • Field '%s' value is too small: must be at least %v", field, minimum)
	case "additional_property_not_allowed":
		property := err.Details()["property"]
		return fmt.Sprintf("  • Unknown field '%s' in %s (not allowed by schema)", property, field)
	case "array_min_items":
		minimum := err.Details()["min"]
		return fmt.Sprintf("  • Field '%s' must have at least %v items", field, minimum)
	case "string_gte":
		minimum := err.Details()["min"]
		return fmt.Sprintf("  • Field '%s' must be at least %v characters long", field, minimum)
	default:
		return fmt.Sprintf("  • Field '%s': %s", field, err.Description())
	}
}
