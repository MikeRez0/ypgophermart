package auth

import (
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/MikeRez0/ypgophermart/internal/core/domain"
	"github.com/MikeRez0/ypgophermart/internal/core/port"
)

type PasetoToken struct {
	parser *paseto.Parser
	key    *paseto.V4SymmetricKey
	token  *paseto.Token
}

func New() (port.TokenService, error) {
	parser := paseto.NewParser()
	key := paseto.NewV4SymmetricKey()
	token := paseto.NewToken()

	s := PasetoToken{
		parser: &parser,
		key:    &key,
		token:  &token,
	}

	return &s, nil
}

func (p *PasetoToken) CreateToken(user *domain.User) (string, error) {
	p.token.SetExpiration(time.Now().Add(1000 * time.Hour))

	payload := port.TokenPayload{UserID: user.ID}
	err := p.token.Set("payload", payload)
	if err != nil {
		return "", domain.ErrTokenCreation
	}

	return p.token.V4Encrypt(*p.key, nil), nil
}
func (p *PasetoToken) VerifyToken(token string) (*port.TokenPayload, error) {
	parsedToken, err := p.parser.ParseV4Local(*p.key, token, nil)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	payload := port.TokenPayload{}
	err = parsedToken.Get("payload", &payload)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	return &payload, nil
}
