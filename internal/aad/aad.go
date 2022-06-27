package aad

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	msalErrors "github.com/AzureAD/microsoft-authentication-library-for-go/apps/errors"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"

	"github.com/ubuntu/aad-auth/internal/pam"
)

const (
	endpoint = "https://login.microsoftonline.com"

	invalidCredCode = 50126
	requiresMFACode = 50076
	noSuchUserCode  = 50034
)

var (
	// NoNetworkErr is returned in case of no network available
	NoNetworkErr = errors.New("NO NETWORK")
	// DenyErr is returned in case of denial returned by AAD
	DenyErr = errors.New("DENY")
)

type aadErr struct {
	ErrorCodes []int `json:"error_codes"`
}

// Authenticate tries to authenticate username against AAD.
func Authenticate(ctx context.Context, tenantID, appID, username, password string) error {
	authority := fmt.Sprintf("%s/%s", endpoint, tenantID)
	pam.LogDebug(ctx, "Connecting to %q, with clientID %q for user %q", authority, appID, username)

	// Get client from network
	app, errAcquireToken := public.New(appID, public.WithAuthority(authority))
	if errAcquireToken != nil {
		pam.LogErr(ctx, "Connection to authority failed: %v", errAcquireToken)
		return NoNetworkErr
	}

	// Authentify the user
	_, errAcquireToken = app.AcquireTokenByUsernamePassword(context.Background(), nil, username, password)

	var callErr msalErrors.CallErr
	if errors.As(errAcquireToken, &callErr) {
		data, err := io.ReadAll(callErr.Resp.Body)
		if err != nil {
			pam.LogErr(ctx, "Can't read server response: %v", err)
			return DenyErr
		}
		var addErrWithCodes aadErr
		if err := json.Unmarshal(data, &addErrWithCodes); err != nil {
			pam.LogErr(ctx, "Invalid server response, not a json object: %v", err)
			return DenyErr
		}
		for _, errcode := range addErrWithCodes.ErrorCodes {
			if errcode == invalidCredCode {
				pam.LogDebug(ctx, "Got response: Invalid credentials")
				return DenyErr
			}
			if errcode == noSuchUserCode {
				pam.LogDebug(ctx, "Got response: User doesn't exist")
				return DenyErr
			}
			if errcode == requiresMFACode {
				pam.LogDebug(ctx, "Authentication successful even if requiring MFA")
				return nil
			}
		}
		pam.LogErr(ctx, "Unknown error code(s) from server: %v", addErrWithCodes.ErrorCodes)
		return DenyErr
	}

	if errAcquireToken != nil {
		pam.LogDebug(ctx, "acquiring token failed: %v", errAcquireToken)
		return NoNetworkErr
	}

	pam.LogDebug(ctx, "Authentication successful with user/password")
	return nil
}