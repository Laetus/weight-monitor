package util

import (
	"net/http"
	"strings"

	"google.golang.org/api/oauth2/v2"
)

var httpClient = &http.Client{}

func trimOauthClientId() {
	splits := strings.Split(ConfigInstance.OauthClientId, "/")
	if len(splits) == 1 {
		return
	}
	ConfigInstance.OauthClientId = splits[len(splits)-1]
}

func Validate(r *http.Request) (*oauth2.Tokeninfo, error) {
	idToken := r.Header.Get("Authorization")

	idToken = strings.Replace(idToken, "Bearer ", "", 1)

	oauth2Service, err := oauth2.New(httpClient)
	tokenInfoCall := oauth2Service.Tokeninfo()
	tokenInfoCall.IdToken(idToken)
	tokenInfo, err := tokenInfoCall.Do()
	return tokenInfo, err
}
