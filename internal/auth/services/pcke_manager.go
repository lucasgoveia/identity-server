package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/oklog/ulid/v2"
	"identity-server/pkg/providers/cache"
	"identity-server/pkg/security"
	"time"
)

type PCKEManager struct {
	secureKeyGenerator *security.SecureKeyGenerator
	cache              cache.Cache
}

func NewPCKEManager(secureKeyGenerator *security.SecureKeyGenerator, cache cache.Cache) *PCKEManager {
	return &PCKEManager{
		secureKeyGenerator: secureKeyGenerator,
		cache:              cache,
	}
}

type PCKEData struct {
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	RedirectUri         string `json:"redirect_uri"`
	UserId              string `json:"user_id"`
	CredentialId        string `json:"credential_id"`
	RememberMe          bool   `json:"remember_me"`
}

func buildPCKEKey(code string) string {
	return fmt.Sprintf("pcke:%s", code)
}

func (p *PCKEManager) New(ctx context.Context, userId ulid.ULID, credentialId ulid.ULID, codeChallenge string, codeChallengeMethod string, redirectUri string, rememberMe bool) (string, error) {

	code, err := p.secureKeyGenerator.Generate([]rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"), 128)

	if err != nil {
		return "", err
	}

	hashedCode := fmt.Sprintf("%x", sha256.Sum256([]byte(code)))

	pckeData := PCKEData{
		UserId:              userId.String(),
		CredentialId:        credentialId.String(),
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		RedirectUri:         redirectUri,
		RememberMe:          rememberMe,
	}

	pckeDataJson, err := json.Marshal(pckeData)

	if err != nil {
		return "", err
	}

	err = p.cache.Set(ctx, buildPCKEKey(hashedCode), pckeDataJson, time.Minute*5)

	if err != nil {
		return "", nil
	}

	return code, nil
}

type ExchangeResponse struct {
	UserId       ulid.ULID
	CredentialId ulid.ULID
	RememberMe   bool
}

func (p *PCKEManager) Exchange(ctx context.Context, code string, codeVerifier string, redirectUri string) (*ExchangeResponse, error) {
	hashedCode := fmt.Sprintf("%x", sha256.Sum256([]byte(code)))

	res, exists := p.cache.GetAndRemove(ctx, buildPCKEKey(hashedCode))

	if !exists {
		return nil, fmt.Errorf("no data found for code %s", code)
	}

	var data PCKEData
	err := json.Unmarshal([]byte(res.(string)), &data)

	if err != nil {
		return nil, err
	}

	if data.CodeChallengeMethod != "S256" {
		plainChallengeVerified := data.CodeChallenge == codeVerifier && data.RedirectUri == redirectUri

		if !plainChallengeVerified {
			return nil, fmt.Errorf("invalid code")
		}

		return &ExchangeResponse{
			UserId:       ulid.MustParse(data.UserId),
			CredentialId: ulid.MustParse(data.CredentialId),
			RememberMe:   data.RememberMe,
		}, nil
	}

	hash := sha256.Sum256([]byte(codeVerifier))
	codeVerifierHash := base64.URLEncoding.EncodeToString(hash[:])

	challengeVerified := data.CodeChallenge == codeVerifierHash && data.RedirectUri == redirectUri

	if !challengeVerified {
		return nil, fmt.Errorf("invalid code")
	}

	return &ExchangeResponse{
		UserId:       ulid.MustParse(data.UserId),
		CredentialId: ulid.MustParse(data.CredentialId),
		RememberMe:   data.RememberMe,
	}, nil
}
