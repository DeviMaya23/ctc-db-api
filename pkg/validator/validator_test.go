package validator

import (
	"lizobly/ctc-db-api/pkg/constants"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ValidatorTestSuite struct {
	suite.Suite
	validator *CustomValidator
}

func TestValidatorSuite(t *testing.T) {
	suite.Run(t, new(ValidatorTestSuite))
}

func (s *ValidatorTestSuite) SetupTest() {
	var err error
	s.validator, err = NewValidator()
	s.NoError(err)
	s.NotNil(s.validator)
}

// TestNewValidator_Success tests successful creation of validator
func (s *ValidatorTestSuite) TestNewValidator_Success() {
	validator, err := NewValidator()
	s.NoError(err)
	s.NotNil(validator)
	s.NotNil(validator.Validator)
	s.NotNil(validator.Translator)
}

// TestNewValidator_ValidTranslator tests validator creation (assumes en translator always exists)
func (s *ValidatorTestSuite) TestNewValidator_ValidTranslator() {
	validator, err := NewValidator()
	s.NoError(err)
	s.NotNil(validator)

	// Verify translator has English translator
	trans, ok := validator.Translator.GetTranslator("en")
	s.True(ok)
	s.NotNil(trans)
}

type TestStructWithInfluence struct {
	Influence string `validate:"influence"`
}

type TestStructWithJob struct {
	Job string `validate:"job"`
}

type TestStructWithInvalidInfluence struct {
	Influence string `validate:"influence"`
}

// TestValidate_ValidStruct tests validation of valid struct
func (s *ValidatorTestSuite) TestValidate_ValidStruct() {
	validStruct := TestStructWithInfluence{
		Influence: constants.InfluenceWealth,
	}

	err := s.validator.Validate(validStruct)
	s.NoError(err)
}

// TestValidate_InvalidStruct tests validation with validation errors
func (s *ValidatorTestSuite) TestValidate_InvalidStruct() {
	invalidStruct := TestStructWithInfluence{
		Influence: "InvalidInfluence",
	}

	err := s.validator.Validate(invalidStruct)
	s.Error(err)
}

// TestValidateInfluence tests influence validation with valid and invalid values
func (s *ValidatorTestSuite) TestValidateInfluence() {
	tests := []struct {
		name      string
		influence string
		shouldErr bool
	}{
		{name: "valid influence Wealth", influence: constants.InfluenceWealth, shouldErr: false},
		{name: "valid influence Power", influence: constants.InfluencePower, shouldErr: false},
		{name: "valid influence Fame", influence: constants.InfluenceFame, shouldErr: false},
		{name: "valid influence Opulence", influence: constants.InfluenceOpulence, shouldErr: false},
		{name: "valid influence Dominance", influence: constants.InfluenceDominance, shouldErr: false},
		{name: "valid influence Prestige", influence: constants.InfluencePrestige, shouldErr: false},
		{name: "invalid influence", influence: "NotAnInfluence", shouldErr: true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			testStruct := TestStructWithInfluence{
				Influence: tt.influence,
			}
			err := s.validator.Validate(testStruct)
			if tt.shouldErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

// TestValidateJob tests job validation with valid and invalid values
func (s *ValidatorTestSuite) TestValidateJob() {
	tests := []struct {
		name      string
		job       string
		shouldErr bool
	}{
		{name: "valid job Warrior", job: constants.JobWarrior, shouldErr: false},
		{name: "valid job Merchant", job: constants.JobMerchant, shouldErr: false},
		{name: "valid job Thief", job: constants.JobThief, shouldErr: false},
		{name: "valid job Apothecary", job: constants.JobApothecary, shouldErr: false},
		{name: "valid job Hunter", job: constants.JobHunter, shouldErr: false},
		{name: "valid job Cleric", job: constants.JobCleric, shouldErr: false},
		{name: "valid job Scholar", job: constants.JobScholar, shouldErr: false},
		{name: "valid job Dancer", job: constants.JobDancer, shouldErr: false},
		{name: "invalid job", job: "NotAJob", shouldErr: true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			testStruct := TestStructWithJob{
				Job: tt.job,
			}
			err := s.validator.Validate(testStruct)
			if tt.shouldErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}
