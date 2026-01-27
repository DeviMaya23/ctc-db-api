package validator

import (
	"fmt"
	"lizobly/ctc-db-api/pkg/constants"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	id_translations "github.com/go-playground/validator/v10/translations/id"
)

type CustomValidator struct {
	Validator  *validator.Validate
	Translator *ut.UniversalTranslator
}

func NewValidator() (*CustomValidator, error) {

	newValidator := validator.New()

	en := en.New()
	id := id.New()

	uni := ut.New(en, en, id)

	english, ok := uni.GetTranslator("en")
	if !ok {
		return nil, fmt.Errorf("failed to get en translator")
	}
	en_translations.RegisterDefaultTranslations(newValidator, english)

	indonesian, ok := uni.GetTranslator("id")
	if !ok {
		return nil, fmt.Errorf("failed to get id translator")
	}
	id_translations.RegisterDefaultTranslations(newValidator, indonesian)

	// Register Custom Validator
	newValidator.RegisterValidation("influence", ValidateInfluence)
	newValidator.RegisterValidation("job", ValidateJob)

	// Register Custom Validator Message
	newValidator.RegisterTranslation("influence", english, func(ut ut.Translator) error {
		return ut.Add("influence", "{0} must be valid influence type.", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("influence", fe.Field())

		return t
	})

	newValidator.RegisterTranslation("job", english, func(ut ut.Translator) error {
		return ut.Add("job", "{0} must be valid job type.", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("job", fe.Field())

		return t
	})

	return &CustomValidator{
		Validator:  newValidator,
		Translator: uni,
	}, nil
}

func (cv *CustomValidator) Validate(i interface{}) error {
	err := cv.Validator.Struct(i)
	if err != nil {
		return err
	}
	return nil
}

func ValidateInfluence(fl validator.FieldLevel) bool {
	return constants.GetInfluenceID(fl.Field().String()) != 0
}

func ValidateJob(fl validator.FieldLevel) bool {
	return constants.GetJobID(fl.Field().String()) != 0
}
