package api

import (
	"github.com/cloudfoundry-community/go-cfclient"
)

type apiClient struct {
	clientConfig *cfclient.Config
	client       *cfclient.Client
}

func NewAPIClient(apiUrl string, username string, password string, sslSkipVerify bool) (*apiClient, error) {
	config := &cfclient.Config{
		ApiAddress:        apiUrl,
		Username:          username,
		Password:          password,
		SkipSslValidation: sslSkipVerify,
	}

	client, err := cfclient.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &apiClient{
		clientConfig: config,
		client:       client,
	}, nil
}

func (api *apiClient) FetchTrafficControllerURL() string {
	return api.client.Endpoint.DopplerEndpoint
}

func (api *apiClient) FetchAuthToken() (string, error) {
	token, err := api.client.GetToken()
	if err != nil {
		return "", err
	}
	return token, nil
}
