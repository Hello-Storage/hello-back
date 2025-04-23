package oauth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type GithubUser struct {
	ID     uint
	Name   string
	Avatar string
}

func GetGithubOAuthToken(code string) (string, error) {
	const rootURl = "https://github.com/login/oauth/access_token"

	values := url.Values{}
	values.Add("code", code)
	//values.Add("client_id", config.Env().GithubClientID)
	//values.Add("client_secret", config.Env().GithubClientSecret)

	query := values.Encode()

	queryString := fmt.Sprintf("%s?%s", rootURl, bytes.NewBufferString(query))
	req, err := http.NewRequest("POST", queryString, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := http.Client{
		Timeout: time.Second * 30,
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", errors.New("could not retrieve token")
	}

	var resBody bytes.Buffer
	_, err = io.Copy(&resBody, res.Body)
	if err != nil {
		return "", err
	}

	parsedQuery, err := url.ParseQuery(resBody.String())
	if err != nil {
		return "", err
	}

	tokenBody := parsedQuery["access_token"][0]

	return tokenBody, nil
}

func GetGithubUser(token string) (*GithubUser, error) {
	req, err := http.NewRequest("GET", github_api_url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	client := http.Client{
		Timeout: time.Second * 30,
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("could not retrieve user")
	}

	var resBody bytes.Buffer
	_, err = io.Copy(&resBody, res.Body)
	if err != nil {
		return nil, err
	}

	var GithubUserRes map[string]interface{}

	if err := json.Unmarshal(resBody.Bytes(), &GithubUserRes); err != nil {
		return nil, err
	}

	id := uint(GithubUserRes["id"].(float64))

	userBody := &GithubUser{
		ID:     id,
		Name:   GithubUserRes["login"].(string),
		Avatar: GithubUserRes["avatar_url"].(string),
	}

	return userBody, nil
}
