package utils

import (
	"encoding/json"
	"net/http"
)

func ParseJson(r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(&dst)
}