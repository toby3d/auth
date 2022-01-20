package usecase

import (
	"context"
	"fmt"

	json "github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/website/indieauth/internal/common"
	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/ticket"
	"source.toby3d.me/website/indieauth/internal/util"
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
		client  *http.Client
		tickets ticket.Repository
	}
)

func NewTicketUseCase(tickets ticket.Repository, client *http.Client) ticket.UseCase {
	return &ticketUseCase{
		client:  client,
		tickets: tickets,
	}
}

func (useCase *ticketUseCase) Generate(ctx context.Context, ticket *domain.Ticket) error {
	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURIBytes(ticket.Subject.RequestURI())

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := useCase.client.Do(req, resp); err != nil {
		return fmt.Errorf("cannot discovery ticket subject: %w", err)
	}

	var ticketEndpoint *domain.URL

	// NOTE(toby3d): find metadata first
	if metadata, err := util.ExtractMetadata(resp, useCase.client); err == nil && metadata != nil {
		ticketEndpoint = metadata.TicketEndpoint
	} else { // NOTE(toby3d): fallback to old links searching
		if endpoints := util.ExtractEndpoints(resp, "ticket_endpoint"); endpoints != nil && len(endpoints) > 0 {
			ticketEndpoint = endpoints[len(endpoints)-1]
		}
	}

	if ticketEndpoint == nil {
		return fmt.Errorf("cannot discovery ticket_endpoint on ticket resource")
	}

	if err := useCase.tickets.Create(ctx, ticket); err != nil {
		return fmt.Errorf("cannot save ticket in store: %w", err)
	}

	req.Reset()
	req.Header.SetMethod(http.MethodPost)
	req.SetRequestURIBytes(ticketEndpoint.RequestURI())
	req.Header.SetContentType(common.MIMEApplicationForm)
	req.PostArgs().Set("ticket", ticket.Ticket)
	req.PostArgs().Set("subject", ticket.Subject.String())
	req.PostArgs().Set("resource", ticket.Resource.String())
	resp.Reset()

	if err := useCase.client.Do(req, resp); err != nil {
		return fmt.Errorf("cannot send ticket to subject ticket_endpoint: %w", err)
	}

	return nil
}

func (useCase *ticketUseCase) Exchange(ctx context.Context, ticket *domain.Ticket) (*domain.Token, error) {
	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.SetRequestURI(ticket.Resource.String())
	req.Header.SetMethod(http.MethodGet)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := useCase.client.Do(req, resp); err != nil {
		return nil, fmt.Errorf("cannot discovery ticket resource: %w", err)
	}

	var tokenEndpoint *domain.URL

	// NOTE(toby3d): find metadata first
	if metadata, err := util.ExtractMetadata(resp, useCase.client); err == nil && metadata != nil {
		tokenEndpoint = metadata.TokenEndpoint
	} else { // NOTE(toby3d): fallback to old links searching
		if endpoints := util.ExtractEndpoints(resp, "token_endpoint"); endpoints != nil && len(endpoints) > 0 {
			tokenEndpoint = endpoints[len(endpoints)-1]
		}
	}

	if tokenEndpoint == nil {
		return nil, fmt.Errorf("cannot discovery token_endpoint on ticket resource")
	}

	req.Reset()
	req.Header.SetMethod(http.MethodPost)
	req.SetRequestURI(tokenEndpoint.String())
	req.Header.SetContentType(common.MIMEApplicationForm)
	req.Header.Set(http.HeaderAccept, common.MIMEApplicationJSON)
	req.PostArgs().Set("grant_type", domain.GrantTypeTicket.String())
	req.PostArgs().Set("ticket", ticket.Ticket)
	resp.Reset()

	if err := useCase.client.Do(req, resp); err != nil {
		return nil, fmt.Errorf("cannot exchange ticket on token_endpoint: %w", err)
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
