package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/ethereum/go-ethereum/common"
)

type JwtValidator struct {
	settings *settings.JwtSettings
	now      func() time.Time
}

func NewJwtValidator(settings *settings.JwtSettings) *JwtValidator {
	return &JwtValidator{settings: settings, now: time.Now}
}

func (v *JwtValidator) Verify(token string) (common.Address, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return common.Address{}, errors.New("malformed token")
	}

	encoder := base64.RawURLEncoding
	headerBytes, err := encoder.DecodeString(parts[0])
	if err != nil {
		return common.Address{}, errors.New("malformed header")
	}
	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return common.Address{}, errors.New("malformed header")
	}
	if header.Alg != "HS256" {
		return common.Address{}, errors.New("unsupported alg")
	}

	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(v.settings.SigningKey))
	mac.Write([]byte(signingInput))
	expectedSig := mac.Sum(nil)
	gotSig, err := encoder.DecodeString(parts[2])
	if err != nil {
		return common.Address{}, errors.New("malformed signature")
	}
	if !hmac.Equal(expectedSig, gotSig) {
		return common.Address{}, errors.New("signature mismatch")
	}

	claimsBytes, err := encoder.DecodeString(parts[1])
	if err != nil {
		return common.Address{}, errors.New("malformed claims")
	}
	var claims jwtClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return common.Address{}, errors.New("malformed claims")
	}

	now := v.now().Unix()
	if claims.Exp > 0 && now > claims.Exp {
		return common.Address{}, errors.New("token expired")
	}
	if claims.Nbf > 0 && now < claims.Nbf {
		return common.Address{}, errors.New("token not yet valid")
	}
	if !common.IsHexAddress(claims.Sub) {
		return common.Address{}, errors.New("subject is not an address")
	}
	return common.HexToAddress(claims.Sub), nil
}
