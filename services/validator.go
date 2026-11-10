package services

import "github.com/go-playground/validator/v10"

var validate = validator.New()

type ValidatorType struct {
	Validate *validator.Validate
}

var Validator = &ValidatorType{
	Validate: validate,
}
