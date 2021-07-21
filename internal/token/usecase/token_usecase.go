package usecase

import (
	"context"
	"time"

	"gitlab.com/toby3d/indieauth/internal/auth"
	"gitlab.com/toby3d/indieauth/internal/model"
	"gitlab.com/toby3d/indieauth/internal/pkce"
	"gitlab.com/toby3d/indieauth/internal/random"
	"gitlab.com/toby3d/indieauth/internal/token"
)

type tokenUseCase struct {
	authRepo  auth.Repository
	tokenRepo token.Repository
}

func NewTokenUseCase(authRepo auth.Repository, tokenRepo token.Repository) token.UseCase {
	return &tokenUseCase{
		authRepo:  authRepo,
		tokenRepo: tokenRepo,
	}
}

func (useCase *tokenUseCase) Exchange(ctx context.Context, req *model.ExchangeRequest) (*model.Token, error) {
	login, err := useCase.authRepo.Get(ctx, req.Code)
	if err != nil {
		return nil, err
	}

	if login == nil {
		return nil, nil
	}

	_ = useCase.authRepo.Delete(ctx, login.Code)

	if time.Now().UTC().After(time.Unix(login.CreatedAt, 0).Add(10 * time.Minute)) {
		return nil, nil
	}

	if login.ClientID != req.ClientID {
		return nil, model.Error{
			Code:        model.ErrInvalidRequest.Code,
			Description: "'client_id' value must be identical to the value at the start of the authorization flow",
		}
	}

	if login.RedirectURI != req.RedirectURI {
		return nil, model.Error{
			Code:        model.ErrInvalidRequest.Code,
			Description: "'client_id' value must be identical to the value at the start of the authorization flow",
		}
	}

	if login.CodeChallenge != "" {
		code, err := pkce.New(login.CodeChallengeMethod)
		if err != nil {
			return nil, err
		}

		code.Verifier = req.CodeVerifier

		code.Generate()

		if login.CodeChallenge != code.Challenge {
			return nil, model.Error{
				Code:        model.ErrInvalidRequest.Code,
				Description: "PKCE validation failed, please go through the authorization flow again",
			}
		}
	}

	token := new(model.Token)
	token.AccessToken = random.New().String(64)
	token.ClientID = login.ClientID
	token.Me = login.Me
	token.Scope = login.Scope
	token.TokenType = "Bearer"

	if err := useCase.tokenRepo.Create(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

func (useCase *tokenUseCase) Verify(ctx context.Context, token string) (*model.Token, error) {
	return useCase.tokenRepo.Get(ctx, token)
}

func (useCase *tokenUseCase) Revoke(ctx context.Context, token string) error {
	return useCase.tokenRepo.Delete(ctx, token)
}
