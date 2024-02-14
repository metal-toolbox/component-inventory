package routes

import "errors"

func apiEcho(m map[string]any) (map[string]any, error) {
	rm := make(map[string]any)

	for k, v := range m {
		rm[k] = v
	}

	return rm, nil
}

func apiError(_ map[string]any) (map[string]any, error) {
	return nil, errors.New("bad times")
}
