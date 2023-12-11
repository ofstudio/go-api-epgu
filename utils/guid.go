package utils

import (
	"github.com/google/uuid"
)

func GUID() (string, error) {
	guid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return guid.String(), nil
}
