package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Validator struct {
	validator *validator.Validate
}

func New() *Validator {
	return &Validator{validator: validator.New()}
}

func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

var _ echo.Validator = (*Validator)(nil)
