package apipgu

import (
	"archive/zip"
	"bytes"
	"fmt"
)

// ArchiveFile - файл вложения для формирования архива [Archive] к создаваемому заявлению
type ArchiveFile struct {
	Filename string // Имя файла с расширением. Пример: "req_346ee59c-a428-42f6-342e-c780dd2e278e.xml"
	Data     []byte // Содержимое файла
}

// Archive - архив вложений к создаваемому заявлению.
// Используется для методов [Client.OrderPush] и [Client.OrderPushChunked].
type Archive struct {
	Name string // Имя архива (без расширения). Пример: "35002123456-archive"
	Data []byte // Содержимое архива в zip-формате
}

// NewArchive - создает архив из файлов вложений.
// В случае ошибки возвращает [ErrZip].
func NewArchive(name string, files ...ArchiveFile) (*Archive, error) {
	if len(files) == 0 {
		return nil, ErrNoFiles
	}

	var b bytes.Buffer
	zipWriter := zip.NewWriter(&b)
	for _, file := range files {
		fileWriter, err := zipWriter.Create(file.Filename)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrZip, err)
		}
		if _, err = fileWriter.Write(file.Data); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrZip, err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrZip, err)
	}

	return &Archive{
		Name: name,
		Data: b.Bytes(),
	}, nil
}
