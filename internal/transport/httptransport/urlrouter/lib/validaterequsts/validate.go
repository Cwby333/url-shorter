package validaterequests

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func Validate(errors validator.ValidationErrors) []string {
	out := make([]string, 0, len(errors))

	for _, err := range errors {
		switch err.ActualTag() {
		case "required":
			str := fmt.Sprintf("field %s is required", err.Field())

			out = append(out, str)
		case "url":
			str := fmt.Sprint("invalid url")

			out = append(out, str)
		}
	}

	return out
}
