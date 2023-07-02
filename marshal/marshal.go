package marshal

import "encoding/json"

func Unmarshal[T any](reqStr string) (T, error) {
	var req T
	err := json.Unmarshal([]byte(reqStr), &req)
	if err != nil {
		return req, err
	}
	return req, nil
}

func Marshal[T any](resp T) (string, error) {
	respStr, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return string(respStr), nil
}
