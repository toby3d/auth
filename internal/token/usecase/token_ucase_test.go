package usecase_test

import (
	"context"
	"path"
	"sync"
	"testing"

	"source.toby3d.me/toby3d/auth/internal/domain"
	"source.toby3d.me/toby3d/auth/internal/profile"
	profilerepo "source.toby3d.me/toby3d/auth/internal/profile/repository/memory"
	"source.toby3d.me/toby3d/auth/internal/session"
	sessionrepo "source.toby3d.me/toby3d/auth/internal/session/repository/memory"
	"source.toby3d.me/toby3d/auth/internal/token"
	tokenrepo "source.toby3d.me/toby3d/auth/internal/token/repository/memory"
	usecase "source.toby3d.me/toby3d/auth/internal/token/usecase"
)

type Dependencies struct {
	config   *domain.Config
	profile  *domain.Profile
	profiles profile.Repository
	session  *domain.Session
	sessions session.Repository
	store    *sync.Map
	token    *domain.Token
	tokens   token.Repository
}

func TestExchange(t *testing.T) {
	t.Parallel()

	deps := NewDependencies(t)
	deps.store.Store(path.Join(profilerepo.DefaultPathPrefix, deps.session.Me.String()), deps.profile)

	if err := deps.sessions.Create(context.Background(), deps.session); err != nil {
		t.Fatal(err)
	}

	opts := token.ExchangeOptions{
		ClientID:     deps.session.ClientID,
		Code:         deps.session.Code,
		CodeVerifier: deps.session.CodeChallenge,
		RedirectURI:  deps.session.RedirectURI,
	}

	tkn, userInfo, err := usecase.NewTokenUseCase(usecase.Config{
		Config:   deps.config,
		Profiles: deps.profiles,
		Sessions: deps.sessions,
		Tokens:   deps.tokens,
	}).Exchange(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}

	if tkn == nil {
		t.Errorf("Exchange(ctx, %v) = nil, want not nil", opts)
	}

	if userInfo == nil {
		t.Errorf("Exchange(ctx, %v) = nil, want not nil", opts)
	}
}

func TestVerify(t *testing.T) {
	t.Parallel()

	deps := NewDependencies(t)
	ucase := usecase.NewTokenUseCase(usecase.Config{
		Config:   deps.config,
		Profiles: deps.profiles,
		Sessions: deps.sessions,
		Tokens:   deps.tokens,
	})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		testToken := domain.TestToken(t)
		result, _, err := ucase.Verify(context.Background(), testToken.AccessToken)
		if err != nil {
			t.Fatal(err)
		}

		if testToken.AccessToken != result.AccessToken ||
			testToken.Scope.String() != result.Scope.String() ||
			testToken.ClientID.String() != result.ClientID.String() ||
			testToken.Me.String() != result.Me.String() {
			t.Errorf("Verify(%s) = %v, want %v", testToken.AccessToken, result, testToken)
		}
	})

	t.Run("revoked", func(t *testing.T) {
		t.Parallel()

		testToken := domain.TestToken(t)
		if err := deps.tokens.Create(context.Background(), testToken); err != nil {
			t.Fatal(err)
		}

		result, _, err := ucase.Verify(context.Background(), testToken.AccessToken)
		if err == nil {
			t.Errorf("Verify(%s) = %v, want error", testToken.AccessToken, err)
		}

		if result != nil {
			t.Errorf("Verify(%s) = %v, want nil", testToken.AccessToken, result)
		}
	})
}

func TestRevoke(t *testing.T) {
	t.Parallel()

	deps := NewDependencies(t)
	if err := usecase.NewTokenUseCase(usecase.Config{
		Config:   deps.config,
		Profiles: deps.profiles,
		Sessions: deps.sessions,
		Tokens:   deps.tokens,
	}).Revoke(context.Background(), deps.token.AccessToken); err != nil {
		t.Fatal(err)
	}

	result, err := deps.tokens.Get(context.Background(), deps.token.AccessToken)
	if err != nil {
		t.Error(err)
	}

	if result.AccessToken != deps.token.AccessToken {
		t.Errorf("Get(%s) = %s, want %s", deps.token.AccessToken, result.AccessToken, deps.token.AccessToken)
	}
}

func NewDependencies(tb testing.TB) Dependencies {
	tb.Helper()

	store := new(sync.Map)
	config := domain.TestConfig(tb)

	return Dependencies{
		config:   config,
		profile:  domain.TestProfile(tb),
		profiles: profilerepo.NewMemoryProfileRepository(store),
		session:  domain.TestSession(tb),
		sessions: sessionrepo.NewMemorySessionRepository(store, config),
		store:    store,
		token:    domain.TestToken(tb),
		tokens:   tokenrepo.NewMemoryTokenRepository(store),
	}
}
