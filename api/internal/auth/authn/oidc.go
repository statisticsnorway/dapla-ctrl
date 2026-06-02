package authn

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

type OIDC struct {
	oauth2.Config
	provider    *oidc.Provider
	tokenSource oauth2.TokenSource
}

func NewOIDC(ctx context.Context, issuer, clientID, redirectURL string, additionalScopes []string) (*OIDC, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}

	ts, err := idtoken.NewTokenSource(ctx, "api://AzureADTokenExchange")
	if err != nil {
		return nil, err
	}

	return &OIDC{
		provider: provider,
		Config: oauth2.Config{
			ClientID:    clientID,
			Endpoint:    provider.Endpoint(),
			RedirectURL: redirectURL,
			Scopes:      append([]string{oidc.ScopeOpenID, "profile", "email"}, additionalScopes...),
		},
		tokenSource: ts,
	}, nil
}

func (o *OIDC) Exchange(ctx context.Context, code string, opt ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	token, err := o.tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("get google token: %w", err)
	}

	return o.Config.Exchange(ctx, code,
		oauth2.SetAuthURLParam("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"),
		oauth2.SetAuthURLParam("client_assertion", token.AccessToken), // Actually the ID token
	)
}

func (o *OIDC) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return o.provider.Verifier(&oidc.Config{ClientID: o.Config.ClientID}).Verify(ctx, rawIDToken)
}
