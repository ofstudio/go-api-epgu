package apipgu

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestArchive(t *testing.T) {
	suite.Run(t, new(suiteTestArchive))
}

type suiteTestArchive struct {
	suite.Suite
}

func (suite *suiteTestArchive) TestNewArchive() {

	suite.Run("success", func() {
		file1 := File{Filename: "file1.txt", Data: []byte("This is file 1")}
		file2 := File{Filename: "file2.txt", Data: []byte("This is file 2")}

		archive, err := NewArchive("test", file1, file2)
		suite.NoError(err)
		suite.Require().NotNil(archive)

		r, err := zip.NewReader(bytes.NewReader(archive.Data), int64(len(archive.Data)))
		suite.Require().NoError(err)
		suite.Require().Len(r.File, 2)

		suite.Equal(file1, suite.unZip(r.File[0]))
		suite.Equal(file2, suite.unZip(r.File[1]))
	})

	suite.Run("no files", func() {
		archive, err := NewArchive("test")
		suite.ErrorIs(err, ErrNoFiles)
		suite.Nil(archive)
	})

}

func (suite *suiteTestArchive) unZip(zipFile *zip.File) File {
	f, err := zipFile.Open()
	suite.Require().NoError(err)
	b := bytes.Buffer{}
	_, err = io.Copy(&b, f)
	suite.Require().NoError(err)
	suite.NoError(f.Close())
	return File{
		Filename: zipFile.Name,
		Data:     b.Bytes(),
	}
}
