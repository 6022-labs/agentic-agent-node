package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/settings"
	"github.com/ethereum/go-ethereum/common"
)

type JwtIssuer struct {
	settings *settings.JwtSettings
	now      func() time.Time
}

func NewJwtIssuer(settings *settings.JwtSettings) *JwtIssuer {
	return &JwtIssuer{settings: settings, now: time.Now}
}

func (i *JwtIssuer) Issue(address common.Address) (string, time.Time, error) {
	now := i.now()
	expiresAt := now.Add(i.settings.TokenTtl)

	claims := jwtClaims{
		Sub: address.Hex(),
		Iat: now.Unix(),
		Nbf: now.Unix(),
		Exp: expiresAt.Unix(),
	}
	token, err := signHs256(claims, []byte(i.settings.SigningKey))
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

type jwtClaims struct {
	Sub string `json:"sub"`
	Iat int64  `json:"iat"`
	Nbf int64  `json:"nbf"`
	Exp int64  `json:"exp"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func signHs256(claims jwtClaims, key []byte) (string, error) {
	headerBytes, err := json.Marshal(jwtHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", err
	}
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encoder := base64.RawURLEncoding
	signingInput := encoder.EncodeToString(headerBytes) + "." + encoder.EncodeToString(claimsBytes)

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(signingInput))
	sig := mac.Sum(nil)
	return signingInput + "." + encoder.EncodeToString(sig), nil
}
