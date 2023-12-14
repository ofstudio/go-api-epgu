package aas

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type errorSuite struct {
	suite.Suite
}

func TestError(t *testing.T) {
	suite.Run(t, new(errorSuite))
}

func (suite *errorSuite) Test_esiaError() {
	suite.Run("known ESIA errors", func() {
		suite.Equal(ErrESIA_036700, esiaError("ESIA-036700: Не указана мнемоника типа согласия"))
		suite.Equal(ErrESIA_036701, esiaError("ESIA-036701: Не найден тип согласия"))
		suite.Equal(ErrESIA_036702, esiaError("ESIA-036702: Не указан обязательный скоуп для типа согласия"))
	})

	suite.Run("short and unknown", func() {
		suite.Equal(ErrESIA_unknown, esiaError("ESIA-999999"))
		suite.Equal(ErrESIA_unknown, esiaError("ESIA"))
		suite.Equal(ErrESIA_unknown, esiaError(""))
	})
}
