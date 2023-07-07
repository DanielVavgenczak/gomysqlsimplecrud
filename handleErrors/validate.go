package handleerrors

import (
	"errors"

	"github.com/go-playground/validator/v10"
)


func ValidationInputs(obj interface{}) error {
	validate := validator.New()
	err := validate.Struct(obj)
	if err == nil {
		return nil
	}

	validationErrors := err.(validator.ValidationErrors) 
	var field string
	for _, v := range validationErrors {
		switch v.Tag(){
			case "required" :
				if v.Field() == "Name" {
					field ="Nome"
				}else {
					field = "Idade"
				}
				return errors.New(field + " é obrigatorio ")
			case "min":
				if v.Field() == "Name" {
					field ="Nome"
				}else {
					field = "Idade"
				}
				return errors.New(field + " dever ter no minimo caracteres "+v.Param())
			case "max":
				if v.Field() == "Name" {
					field ="Nome"
				}else {
					field = "Idade"
				}
				return errors.New(field + " dever ter no maxiomo "+v.Param())
			default:
				return nil
		}
	}

	return nil
}