package port

import "github.com/MikeRez0/ypgophermart/internal/core/domain"

type TokenPayload struct {
	UserID uint64
}

type TokenService interface {
	CreateToken(user *domain.User) (string, error)
	VerifyToken(token string) (*TokenPayload, error)
}
