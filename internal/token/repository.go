package token

import (
	"context"

	"golang.org/x/xerrors"
	"source.toby3d.me/website/oauth/internal/model"
)

type Repository interface {
	Get(ctx context.Context, accessToken string) (*model.Token, error)
	Create(ctx context.Context, accessToken *model.Token) error
	Update(ctx context.Context, accessToken *model.Token) error
	Remove(ctx context.Context, accessToken string) error
}

var ErrExist error = model.Error{
	Code:        "invalid_request",
	Description: "this token is already exists",
	URI:         "",
	Frame:       xerrors.Caller(1),
}
