package usecase

import (
	"bytes"
	"context"
	"net/url"
	"time"

	http "github.com/valyala/fasthttp"
	"gitlab.com/toby3d/indieauth/internal/auth"
	"gitlab.com/toby3d/indieauth/internal/domain"
	"gitlab.com/toby3d/indieauth/internal/pkce"
	"gitlab.com/toby3d/indieauth/internal/random"
	"willnorris.com/go/microformats"
)

type authUseCase struct {
	client *http.Client
	repo   auth.Repository
}

func NewAuthUseCase(repo auth.Repository) auth.UseCase {
	return &authUseCase{
		client: new(http.Client),
		repo:   repo,
	}
}

func (useCase *authUseCase) Discovery(ctx context.Context, clientId string) (*domain.Client, error) {
	_, src, err := useCase.client.Get(nil, clientId)
	if err != nil {
		return nil, err
	}

	cid, err := url.Parse(clientId)
	if err != nil {
		return nil, err
	}

	data := microformats.Parse(bytes.NewReader(src), cid)

	client := new(domain.Client)
	client.RedirectURI = make([]string, 0)

	for i := range data.Items {
		if len(data.Items[i].Type) == 0 || data.Items[i].Type[0] != "h-app" {
			continue
		}

		for key, values := range data.Items[i].Properties {
			switch key {
			case "logo":
				for j := range values {
					switch val := values[j].(type) {
					case string:
						client.Logo = val
					case map[string]string:
						client.Logo = val["value"]
					}
				}
			case "name":
				for j := range values {
					client.Name, _ = values[j].(string)
				}
			case "url":
				for j := range values {
					client.URL, _ = values[j].(string)
				}
			}
		}
	}

	for key, values := range data.Rels {
		if key != "redirect_uri" {
			continue
		}

		client.RedirectURI = append(client.RedirectURI, values...)
	}

	if client.URL != clientId {
		return nil, domain.Error{
			Code:        domain.ErrInvalidRequest.Code,
			Description: "'client_id' does not match the actual client URL",
		}
	}

	return client, nil
}

func (useCase *authUseCase) Approve(ctx context.Context, login *domain.Login) (string, error) {
	login.Code = random.New().String(32)

	if err := useCase.repo.Create(ctx, login); err != nil {
		return "", err
	}

	return login.Code, nil
}

func (useCase *authUseCase) Exchange(ctx context.Context, req *domain.ExchangeRequest) (string, error) {
	login, err := useCase.repo.Get(ctx, req.Code)
	if err != nil {
		return "", err
	}

	if login == nil {
		return "", nil
	}

	_ = useCase.repo.Delete(ctx, req.Code)

	if time.Now().UTC().After(time.Unix(login.CreatedAt, 0).Add(10 * time.Minute)) {
		return "", nil
	}

	if login.ClientID != req.ClientID || login.RedirectURI != req.RedirectURI {
		return "", domain.ErrInvalidRequest
	}

	if login.CodeChallenge != "" {
		codeChallenge, err := pkce.New(login.CodeChallengeMethod)
		if err != nil {
			return "", err
		}

		codeChallenge.Verifier = req.CodeVerifier

		codeChallenge.Generate()

		if login.CodeChallenge != codeChallenge.Challenge {
			return "", domain.ErrInvalidRequest
		}
	}

	return login.Me, nil
}
