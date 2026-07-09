package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCClient wraps the OIDC provider configurations and verifiers.
type OIDCClient struct {
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
}

// NewOIDCClient fetches the provider configurations and builds the OAuth2 flow.
func NewOIDCClient(ctx context.Context, issuer, clientID, clientSecret, redirectURL string) (*OIDCClient, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve OIDC provider: %w", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &OIDCClient{
		oauth2Config: oauth2Config,
		verifier:     verifier,
	}, nil
}

// AuthCodeURL returns the login redirection URL.
func (c *OIDCClient) AuthCodeURL(state string) string {
	return c.oauth2Config.AuthCodeURL(state)
}

// ExchangeAndVerify exchanges the authorization code for an ID Token and verifies it.
func (c *OIDCClient) ExchangeAndVerify(ctx context.Context, code string) (*oidc.IDToken, error) {
	oauth2Token, err := c.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange auth code: %w", err)
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("id_token parameter missing in token response")
	}

	idToken, err := c.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID Token: %w", err)
	}

	return idToken, nil
}
