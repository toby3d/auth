package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/goccy/go-json"

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

func (useCase *ticketUseCase) Generate(ctx context.Context, tkt domain.Ticket) error {
	resp, err := useCase.client.Get(tkt.Subject.String())
	if err != nil {
		return fmt.Errorf("cannot discovery ticket subject: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read response body: %w", err)
	}

	buf := bytes.NewReader(body)
	ticketEndpoint := new(url.URL)

	// NOTE(toby3d): find metadata first
	metadata, err := httputil.ExtractFromMetadata(useCase.client, tkt.Subject.String())
	if err == nil && metadata != nil {
		ticketEndpoint = metadata.TicketEndpoint
	} else { // NOTE(toby3d): fallback to old links searching
		endpoints := httputil.ExtractEndpoints(buf, tkt.Subject.URL(), resp.Header.Get(common.HeaderLink),
			"ticket_endpoint")
		if len(endpoints) > 0 {
			ticketEndpoint = endpoints[len(endpoints)-1]
		}
	}

	if ticketEndpoint == nil {
		return ticket.ErrTicketEndpointNotExist
	}

	if err = useCase.tickets.Create(ctx, tkt); err != nil {
		return fmt.Errorf("cannot save ticket in store: %w", err)
	}

	payload := make(url.Values)
	payload.Set("ticket", tkt.Ticket)
	payload.Set("subject", tkt.Subject.String())
	payload.Set("resource", tkt.Resource.String())

	if _, err = useCase.client.PostForm(ticketEndpoint.String(), payload); err != nil {
		return fmt.Errorf("cannot send ticket to subject ticket_endpoint: %w", err)
	}

	return nil
}

func (useCase *ticketUseCase) Redeem(ctx context.Context, tkt domain.Ticket) (*domain.Token, error) {
	resp, err := useCase.client.Get(tkt.Resource.String())
	if err != nil {
		return nil, fmt.Errorf("cannot discovery ticket resource: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	buf := bytes.NewReader(body)
	tokenEndpoint := new(url.URL)

	// NOTE(toby3d): find metadata first
	metadata, err := httputil.ExtractFromMetadata(useCase.client, tkt.Resource.String())
	if err == nil && metadata != nil {
		tokenEndpoint = metadata.TokenEndpoint
	} else { // NOTE(toby3d): fallback to old links searching
		endpoints := httputil.ExtractEndpoints(buf, tkt.Resource, resp.Header.Get(common.HeaderLink),
			"token_endpoint")
		if len(endpoints) > 0 {
			tokenEndpoint = endpoints[len(endpoints)-1]
		}
	}

	if tokenEndpoint == nil || tokenEndpoint.String() == "" {
		return nil, ticket.ErrTokenEndpointNotExist
	}

	payload := make(url.Values)
	payload.Set("grant_type", domain.GrantTypeTicket.String())
	payload.Set("ticket", tkt.Ticket)

	resp, err = useCase.client.PostForm(tokenEndpoint.String(), payload)
	if err != nil {
		return nil, fmt.Errorf("cannot exchange ticket on token_endpoint: %w", err)
	}

	data := new(AccessToken)
	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, fmt.Errorf("cannot unmarshal access token response: %w", err)
	}

	return &domain.Token{
		CreatedAt: time.Now().UTC(),
		Expiry:    time.Unix(data.ExpiresIn, 0),
		Scope:     nil, // TODO(toby3d)
		// TODO(toby3d): should this also include client_id?
		// https://github.com/indieweb/indieauth/issues/85
		ClientID:     domain.ClientID{},
		Me:           *data.Me,
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
		Issuer:      domain.ClientID{},
		Subject:     *tkt.Subject,
		Secret:      []byte(useCase.config.JWT.Secret),
		Algorithm:   useCase.config.JWT.Algorithm,
		NonceLength: useCase.config.JWT.NonceLength,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot generate a new access token: %w", err)
	}

	return token, nil
}
