package adapters

import (
	"better-console-backend/config"
	"better-console-backend/dtos"
	"encoding/json"
	"fmt"
	"github.com/bettercode-oss/rest"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type GoogleOAuthAdapter struct {
}

func (adapter GoogleOAuthAdapter) Authenticate(code string, setting dtos.GoogleWorkspaceLoginSetting) (dtos.GoogleMember, error) {
	accessToken, err := adapter.getAccessToken(code, setting)
	if err != nil {
		return dtos.GoogleMember{}, err
	}

	client := rest.Client{}
	googleMember := dtos.GoogleMember{}
	client.
		Request().
		SetResult(&googleMember).
		Get(fmt.Sprintf("%v?access_token=%v", config.Config.GoogleOAuth.AuthUri, accessToken))

	return googleMember, nil
}

func (GoogleOAuthAdapter) getAccessToken(code string, setting dtos.GoogleWorkspaceLoginSetting) (string, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", setting.ClientId)
	data.Set("client_secret", setting.ClientSecret)
	data.Set("redirect_uri", setting.RedirectUri)
	data.Set("grant_type", "authorization_code")

	client := &http.Client{}
	r, err := http.NewRequest("POST", config.Config.GoogleOAuth.TokenUri, strings.NewReader(data.Encode())) // URL-encoded payload
	if err != nil {
		return "", err
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	res, err := client.Do(r)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	responseBody := map[string]interface{}{}
	json.Unmarshal(body, &responseBody)

	return responseBody["access_token"].(string), nil
}
