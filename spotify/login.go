package spotify

import "github.com/eolso/librespot-golang/librespot/core"

var tokenChan chan core.OAuth

func StartLocalOAuth(id string, secret string, callback string) string {
	var url string
	url, tokenChan = core.StartLocalOAuthServer(id, secret, callback)
	return url
}

func GetOAuthToken() string {
	if tokenChan != nil {
		oauth := <-tokenChan
		tokenChan = nil
		return oauth.AccessToken
	}

	return ""
}
