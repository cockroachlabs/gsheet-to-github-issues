package main

import (
	"fmt"
	"os"

	"github.com/google/go-github/v30/github"
	"golang.org/x/oauth2"
)

type tokenSource struct {
	token string
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.token,
	}
	return token, nil
}

func getGithubClient() (*github.Client, error) {
	apiKey, ok := os.LookupEnv("GITHUB_API_KEY")
	if !ok {
		return nil, fmt.Errorf("cannot find GITHUB_API_KEY")
	}
	tokenSource := &tokenSource{
		token: apiKey,
	}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	c := github.NewClient(oauthClient)
	return c, nil
}
