package utils

import (
	"encoding/json"
	"net/url"
)

// PrettyJSON возвращает переданную структуру в виде форматированной JSON-строки с отступами.
func PrettyJSON(data any) string {
	indented, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return ""
	}

	return string(indented)
}

// PrettyQuery возвращает параметры запроса в виде форматированной строки.
func PrettyQuery(query url.Values) string {
	s := ""
	for key := range query {
		s += key + "=" + query.Get(key) + "\n"
	}
	return s
}
