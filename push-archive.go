package apipgu

import (
	"archive/zip"
	"bytes"
	"fmt"
)

type PushAttachment struct {
	Filename string
	Data     []byte
}

type PushArchive struct {
	Name  string
	Files []PushAttachment
}

func (a *PushArchive) Zip() ([]byte, error) {
	if len(a.Files) == 0 {
		return nil, ErrNoFiles
	}

	var b bytes.Buffer
	zipWriter := zip.NewWriter(&b)

	for _, file := range a.Files {
		fileWriter, err := zipWriter.Create(file.Filename)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrZipCreate, err)
		}
		if _, err = fileWriter.Write(file.Data); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrZipWrite, err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrZipClose, err)
	}

	return b.Bytes(), nil
}
