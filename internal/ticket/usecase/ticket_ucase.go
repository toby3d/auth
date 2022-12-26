package usecase

import (
	"context"
	"fmt"
	"time"

	json "github.com/goccy/go-json"
	http "github.com/valyala/fasthttp"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/httputil"
	"source.toby3d.me/toby3d/auth/internal/ticket"
)

type (
	//nolint:tagliatelle // https://indieauth.net/source/#access-token-response
	AccessToken struct {
		Me           *domain.Me `json:"me"`
		Profile      *Profile   `json:"profile,omitempty"`
		AccessToken  string     `json:"access_token"`
		RefreshToken string     `json:"refresh_token"`
		ExpiresIn    int64      `json:"expires_in,omitempty"`
	}

	Profile struct {
		Email *domain.Email `json:"email,omitempty"`
		Photo *domain.URL   `json:"photo,omitempty"`
		URL   *domain.URL   `json:"url,omitempty"`
		Name  string        `json:"name,omitempty"`
	}

	ticketUseCase struct {
		config  *domain.Config
		client  *http.Client
		tickets ticket.Repository
	}
)

func NewTicketUseCase(tickets ticket.Repository, client *http.Client, config *domain.Config) ticket.UseCase {
	return &ticketUseCase{
		client:  client,
		tickets: tickets,
		config:  config,
	}
}

func (useCase *ticketUseCase) Generate(ctx context.Context, tkt *domain.Ticket) error {
	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURI(tkt.Subject.String())

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := useCase.client.Do(req, resp); err != nil {
		return fmt.Errorf("cannot discovery ticket subject: %w", err)
	}

	var ticketEndpoint *domain.URL

	// NOTE(toby3d): find metadata first
	if metadata, err := httputil.ExtractMetadata(resp, useCase.client); err == nil && metadata != nil {
		ticketEndpoint = metadata.TicketEndpoint
	} else { // NOTE(toby3d): fallback to old links searching
		if endpoints := httputil.ExtractEndpoints(resp, "ticket_endpoint"); len(endpoints) > 0 {
			ticketEndpoint = endpoints[len(endpoints)-1]
		}
	}

	if ticketEndpoint == nil {
		return ticket.ErrTicketEndpointNotExist
	}

	if err := useCase.tickets.Create(ctx, tkt); err != nil {
		return fmt.Errorf("cannot save ticket in store: %w", err)
	}

	req.Reset()
	req.Header.SetMethod(http.MethodPost)
	req.SetRequestURIBytes(ticketEndpoint.RequestURI())
	req.Header.SetContentType(common.MIMEApplicationForm)
	req.PostArgs().Set("ticket", tkt.Ticket)
	req.PostArgs().Set("subject", tkt.Subject.String())
	req.PostArgs().Set("resource", tkt.Resource.String())
	resp.Reset()

	if err := useCase.client.Do(req, resp); err != nil {
		return fmt.Errorf("cannot send ticket to subject ticket_endpoint: %w", err)
	}

	return nil
}

func (useCase *ticketUseCase) Redeem(ctx context.Context, tkt *domain.Ticket) (*domain.Token, error) {
	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)
	req.SetRequestURI(tkt.Resource.String())
	req.Header.SetMethod(http.MethodGet)

	resp := http.AcquireResponse()
	defer http.ReleaseResponse(resp)

	if err := useCase.client.Do(req, resp); err != nil {
		return nil, fmt.Errorf("cannot discovery ticket resource: %w", err)
	}

	var tokenEndpoint *domain.URL

	// NOTE(toby3d): find metadata first
	if metadata, err := httputil.ExtractMetadata(resp, useCase.client); err == nil && metadata != nil {
		tokenEndpoint = metadata.TokenEndpoint
	} else { // NOTE(toby3d): fallback to old links searching
		if endpoints := httputil.ExtractEndpoints(resp, "token_endpoint"); len(endpoints) > 0 {
			tokenEndpoint = endpoints[len(endpoints)-1]
		}
	}

	if tokenEndpoint == nil {
		return nil, ticket.ErrTokenEndpointNotExist
	}

	req.Reset()
	req.Header.SetMethod(http.MethodPost)
	req.SetRequestURI(tokenEndpoint.String())
	req.Header.SetContentType(common.MIMEApplicationForm)
	req.Header.Set(http.HeaderAccept, common.MIMEApplicationJSON)
	req.PostArgs().Set("grant_type", domain.GrantTypeTicket.String())
	req.PostArgs().Set("ticket", tkt.Ticket)
	resp.Reset()

	if err := useCase.client.Do(req, resp); err != nil {
		return nil, fmt.Errorf("cannot exchange ticket on token_endpoint: %w", err)
	}

	data := new(AccessToken)
	if err := json.Unmarshal(resp.Body(), data); err != nil {
		return nil, fmt.Errorf("cannot unmarshal access token response: %w", err)
	}

	return &domain.Token{
		CreatedAt: time.Now().UTC(),
		Expiry:    time.Unix(data.ExpiresIn, 0),
		Scope:     nil, // TODO(toby3d)
		// TODO(toby3d): should this also include client_id?
		// https://github.com/indieweb/indieauth/issues/85
		ClientID:     nil,
		Me:           data.Me,
		AccessToken:  data.AccessToken,
		RefreshToken: "", // TODO(toby3d)
	}, nil
}

func (useCase *ticketUseCase) Exchange(ctx context.Context, ticket string) (*domain.Token, error) {
	tkt, err := useCase.tickets.GetAndDelete(ctx, ticket)
	if err != nil {
		return nil, fmt.Errorf("cannot find provided ticket: %w", err)
	}

	token, err := domain.NewToken(domain.NewTokenOptions{
		Expiration:  useCase.config.JWT.Expiry,
		Scope:       domain.Scopes{domain.ScopeRead},
		Issuer:      nil,
		Subject:     tkt.Subject,
		Secret:      []byte(useCase.config.JWT.Secret),
		Algorithm:   useCase.config.JWT.Algorithm,
		NonceLength: useCase.config.JWT.NonceLength,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot generate a new access token: %w", err)
	}

	return token, nil
}
