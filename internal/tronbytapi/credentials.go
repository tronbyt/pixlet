package tronbytapi

import "github.com/tronbyt/pixlet/cmd/config"

type APICredentials struct {
	baseURL string
	token   string
}

func ResolveAPICredentials(baseURL, token string) (APICredentials, error) {
	var err error
	if baseURL == "" {
		if baseURL, err = config.GetURL(); err != nil {
			return APICredentials{}, err
		}
	}
	if token == "" {
		if token, err = config.GetToken(); err != nil {
			return APICredentials{}, err
		}
	}

	return APICredentials{
		baseURL: baseURL,
		token:   token,
	}, nil
}
