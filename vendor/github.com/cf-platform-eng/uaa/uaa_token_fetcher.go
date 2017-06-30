package uaa

import (
	"github.com/cloudfoundry-incubator/uaago"
)

type uaaTokenFetcher struct {
	uaaUrl   string
	username string
	password string
	skipSSL  bool
}

type UAATokenFetcher interface {
	FetchAuthToken() (string, error)
}

func NewUAATokenFetcher(url string, username string, password string, sslSkipVerify bool) UAATokenFetcher {
	return &uaaTokenFetcher{
		uaaUrl:   url,
		username: username,
		password: password,
		skipSSL:  sslSkipVerify,
	}
}

func (uaa *uaaTokenFetcher) FetchAuthToken() (string, error) {
	uaaClient, err := uaago.NewClient(uaa.uaaUrl)
	if err != nil {
		return "", err
	}

	authToken, err := uaaClient.GetAuthToken(uaa.username, uaa.password, uaa.skipSSL)
	if err != nil {
		return "", err
	}
	return authToken, nil
}
