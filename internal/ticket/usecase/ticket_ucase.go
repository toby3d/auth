package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/oauth/internal/common"
	"source.toby3d.me/website/oauth/internal/domain"
	"source.toby3d.me/website/oauth/internal/ticket"
)

type (
	//nolint: tagliatelle
	Response struct {
		Me          *domain.Me    `json:"me"`
		Scope       domain.Scopes `json:"scope"`
		AccessToken string        `json:"access_token"`
		TokenType   string        `json:"token_type"`
	}

	ticketUseCase struct {
		client *http.Client
	}
)

func NewTicketUseCase(client *http.Client) ticket.UseCase {
	return &ticketUseCase{
		client: client,
	}
}

func (useCase *ticketUseCase) Redeem(ctx context.Context, endpoint *domain.URL, ticket string) (*domain.Token, error) {
	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.Header.SetMethod(http.MethodPost)
	req.SetRequestURIBytes(endpoint.FullURI())
	req.Header.SetContentType(common.MIMEApplicationForm)
	req.Header.Set(http.HeaderAccept, common.MIMEApplicationJSON)
	req.PostArgs().Set("grant_type", domain.GrantTypeTicket.String())
	req.PostArgs().Set("ticket", ticket)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := useCase.client.Do(req, resp); err != nil {
		return nil, err
	}

	data := new(Response)
	if err := json.Unmarshal(resp.Body(), data); err != nil {
		return nil, fmt.Errorf("cannot unmarshal access token response: %w", err)
	}

	// TODO(toby3d): should this also include client_id?
	// https://github.com/indieweb/indieauth/issues/85
	return &domain.Token{
		ClientID:    nil,
		AccessToken: data.AccessToken,
		Me:          data.Me,
		Scope:       data.Scope,
	}, nil
}
