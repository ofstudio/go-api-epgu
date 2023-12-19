package apipgu

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestDateTime(t *testing.T) {
	suite.Run(t, new(suiteTestDateTime))
}

type suiteTestDateTime struct {
	suite.Suite
}

func (suite *suiteTestDateTime) TestMarshalJSON() {
	suite.Run("2023-11-02T07:27:22", func() {
		dt := DateTime{time.Date(2023, 11, 2, 7, 27, 22, 0, time.UTC)}
		b, err := json.Marshal(dt)
		suite.NoError(err)
		suite.Equal(`"2023-11-02T07:27:22.000+0000"`, string(b))
	})

	suite.Run("null time", func() {
		dt := DateTime{time.Time{}}
		b, err := json.Marshal(dt)
		suite.NoError(err)
		suite.Equal(`null`, string(b))
	})

}

func (suite *suiteTestDateTime) TestUnmarshalJSON() {
	suite.Run("2023-11-02T07:27:22", func() {
		var dt DateTime

		err := json.Unmarshal([]byte(`"2023-11-02T07:27:22.000+0000"`), &dt)
		suite.NoError(err)
		suite.NoError(err)
		suite.Equal(time.Date(2023, 11, 2, 7, 27, 22, 0, dt.Time.Location()), dt.Time)
	})

	suite.Run("null time", func() {
		var dt DateTime
		err := json.Unmarshal([]byte(`null`), &dt)
		suite.NoError(err)
		suite.Equal(time.Time{}, dt.Time)
	})

	suite.Run("empty string", func() {
		var dt DateTime
		err := json.Unmarshal([]byte(`""`), &dt)
		suite.Error(err)
		suite.Equal(time.Time{}, dt.Time)
	})

	suite.Run("invalid string", func() {
		var dt DateTime
		err := json.Unmarshal([]byte(`"invalid string"`), &dt)
		suite.Error(err)
		suite.Equal(time.Time{}, dt.Time)
	})
}
