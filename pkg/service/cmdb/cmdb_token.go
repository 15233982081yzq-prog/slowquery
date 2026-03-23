package cmdb

import (
	"smart-slowquery/pkg/log"

	envUtil "smart-slowquery/internal/util/env"

	"fmt"
	"time"

	"git.garena.com/shopee/go-shopeelib/spacelib/models/auth"
	"github.com/pkg/errors"
)

// Token uid bot token
type Token struct {
	Value  string
	Expiry time.Time
}

func (t *Token) IsExpired() bool {
	return !t.Expiry.IsZero() && t.Expiry.Before(time.Now())
}

const (
	// defaultTokenExpiry specifies the duration that tokens should live by default.
	defaultTokenExpiry = 6 * time.Hour
)

var (
	authHandler auth.AuthHandlerInterface = auth.NewAuthHandler()

	botToken *Token
)

// SpaceUserPassTokenFetcher is a TokenFetcher implementation that
// returns a SPACE token given a username and password.
type SpaceUserPassTokenFetcher struct {
	username string
	password string
	spaceEnv string
}

func NewSpaceUserPassTokenFetcher(username, password, spaceEnv string) *SpaceUserPassTokenFetcher {
	return &SpaceUserPassTokenFetcher{username: username, password: password, spaceEnv: spaceEnv}
}

func (t *SpaceUserPassTokenFetcher) FetchToken() (*Token, error) {
	// Ensure that credentials are not empty.
	if t.username == "" {
		return nil, fmt.Errorf("username not specified")
	}
	if t.password == "" {
		return nil, fmt.Errorf("password not specified")
	}

	var (
		token = &Token{}
		resp  *auth.Auth
		err   error
	)

	// Login to UIC.
	if resp, err = authHandler.LoginByBasicAuthWithEnv(t.spaceEnv, t.username, t.password); err != nil {
		return nil, errors.Wrapf(err, "login to UIC failed")
	}

	// Ensure that token is not empty.
	if resp == nil || resp.Token == "" {
		return nil, errors.Errorf("login failed, received empty token, resp is %v", resp)
	}

	token.Value = resp.Token
	token.Expiry = time.Now().Add(defaultTokenExpiry)
	return token, nil
}

// GetToken outer get the token
func GetToken(env string) (*Token, error) {
	if botToken == nil || botToken.IsExpired() {
		var (
			token *Token
			err   error
		)

		switch env {
		case envUtil.ServerLiveEnv:
			token, err = NewSpaceUserPassTokenFetcher("db_tools_archive", "u2uvqybLHeFf", "live").FetchToken()
		default:
			token, err = NewSpaceUserPassTokenFetcher("db_tools_archive", "A5RvMQUU7KDP", "test").FetchToken()
		}
		if err != nil || token == nil {
			log.Errorf("fetch token error %v", err)
			return nil, err
		}
		botToken = token
	}
	return botToken, nil
}
