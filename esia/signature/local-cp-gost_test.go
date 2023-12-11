package signature

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type suiteLocalCryptoPro struct {
	suite.Suite
	signer *LocalCryptoPro
}

func TestLocalCryptoPro(t *testing.T) {
	suite.Run(t, new(suiteLocalCryptoPro))
}

func (suite *suiteLocalCryptoPro) SetupTest() {
	suite.signer = NewLocalCryptoPro("test", "test", "test_hash")
	suite.signer.cmd = newTestCmd(suite.T())
}

func (suite *suiteLocalCryptoPro) TestSign() {
	suite.Run("success", func() {
		signature, err := suite.signer.Sign([]byte(testDataToSign))
		suite.NoError(err)
		suite.Equal(testSignatureReversed, string(signature))
	})

	suite.Run("error", func() {
		signature, err := suite.signer.Sign([]byte{})
		suite.ErrorIs(err, ErrCPTestExec)
		suite.Nil(signature)
	})
}

func (suite *suiteLocalCryptoPro) TestCertHash() {
	suite.Equal("test_hash", suite.signer.CertHash())
}

const (
	testDataToSign        = "this is a test data"
	testSignature         = "this is a test signature"
	testSignatureReversed = "erutangis tset a si siht"
)

type testCmd struct {
	t *testing.T
}

func newTestCmd(t *testing.T) *testCmd {
	return &testCmd{t: t}
}

func (c *testCmd) Run(name string, args ...string) error {

	require.Equal(c.t, 11, len(args))
	require.Equal(c.t, "-keys", args[0])
	require.Equal(c.t, "-sign", args[1])
	require.Equal(c.t, "GOST12_256", args[2])
	require.Equal(c.t, "-cont", args[3])
	require.NotEmpty(c.t, args[4])
	require.Equal(c.t, "-keytype", args[5])
	require.Equal(c.t, "exchange", args[6])
	require.Equal(c.t, "-in", args[7])
	require.NotEmpty(c.t, args[8])
	require.Equal(c.t, "-out", args[9])
	require.NotEmpty(c.t, args[10])

	inFname := args[8]
	outFname := args[10]

	// читаем подписываемые данные из временного файла
	inBytes, err := os.ReadFile(inFname)
	require.NoError(c.t, err)

	// если входные данные пустые, то возвращаем ошибку
	if len(inBytes) == 0 {
		return errors.New("some error")
	}

	require.Equal(c.t, testDataToSign, string(inBytes))

	// записываем подпись во временный файл
	err = os.WriteFile(outFname, []byte(testSignature), 0644)
	require.NoError(c.t, err)

	if len(inBytes) == 0 {
		return errors.New("test error")
	}

	return nil
}
