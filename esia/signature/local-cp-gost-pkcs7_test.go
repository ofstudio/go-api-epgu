package signature

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type suiteLocalCryptoProPKCS7 struct {
	suite.Suite
	signer *LocalCryptoProPKCS7
}

func TestLocalCryptoProPKCS7(t *testing.T) {
	suite.Run(t, new(suiteLocalCryptoProPKCS7))
}

func (suite *suiteLocalCryptoProPKCS7) SetupTest() {
	suite.signer = NewLocalCryptoProPKCS7("test", "test_thumbprint")
	suite.signer.cmd = newTestPKCS7Cmd(suite.T())
}

func (suite *suiteLocalCryptoProPKCS7) TestSign() {
	suite.Run("success", func() {
		signature, err := suite.signer.Sign([]byte(testDataToSign))
		suite.NoError(err)
		suite.Equal(testSignature, string(signature))
	})

	suite.Run("error", func() {
		signature, err := suite.signer.Sign([]byte{})
		suite.ErrorIs(err, ErrCPTestExec)
		suite.ErrorContains(err, "csptest output")
		suite.Nil(signature)
	})
}

func (suite *suiteLocalCryptoProPKCS7) TestCertHash() {
	suite.Equal("", suite.signer.CertHash())
}

type testPKCS7Cmd struct {
	t *testing.T
}

func newTestPKCS7Cmd(t *testing.T) *testPKCS7Cmd {
	return &testPKCS7Cmd{t: t}
}

func (c *testPKCS7Cmd) Run(name string, args ...string) ([]byte, error) {
	require.Equal(c.t, 12, len(args))
	require.Equal(c.t, "-sfsign", args[0])
	require.Equal(c.t, "-sign", args[1])
	require.Equal(c.t, "-detached", args[2])
	require.Equal(c.t, "-add", args[3])
	require.Equal(c.t, "-alg", args[4])
	require.Equal(c.t, "GOST12_256", args[5])
	require.Equal(c.t, "-in", args[6])
	require.NotEmpty(c.t, args[7])
	require.Equal(c.t, "-out", args[8])
	require.NotEmpty(c.t, args[9])
	require.Equal(c.t, "-my", args[10])
	require.Equal(c.t, "test_thumbprint", args[11])

	inFname := args[7]
	outFname := args[9]

	inBytes, err := os.ReadFile(inFname)
	require.NoError(c.t, err)

	if len(inBytes) == 0 {
		return []byte("csptest output"), errors.New("some error")
	}

	require.Equal(c.t, testDataToSign, string(inBytes))

	err = os.WriteFile(outFname, []byte(testSignature), 0644)
	require.NoError(c.t, err)

	return nil, nil
}
