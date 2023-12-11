package utils

import (
	"encoding/json"
	"net/url"
)

func PrettyJSON(data any) string {
	indented, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return ""
	}

	return string(indented)
}

func PrettyQuery(query url.Values) string {
	s := ""
	for key := range query {
		s += key + "=" + query.Get(key) + "\n"
	}
	return s
}
