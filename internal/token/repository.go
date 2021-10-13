package token

import (
	"context"

	"golang.org/x/xerrors"

	"source.toby3d.me/website/oauth/internal/domain"
)

type Repository interface {
	Get(ctx context.Context, accessToken string) (*domain.Token, error)
	Create(ctx context.Context, accessToken *domain.Token) error
}

var ErrExist error = domain.Error{
	Code:        "invalid_request",
	Description: "this token is already exists",
	URI:         "",
	Frame:       xerrors.Caller(1),
}
