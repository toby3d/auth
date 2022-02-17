package usecase

import (
	"context"
	"fmt"

	"source.toby3d.me/website/indieauth/internal/domain"
	"source.toby3d.me/website/indieauth/internal/profile"
)

type profileUseCase struct {
	profiles profile.Repository
}

func NewProfileUseCase(profiles profile.Repository) profile.UseCase {
	return &profileUseCase{
		profiles: profiles,
	}
}

func (uc *profileUseCase) Fetch(ctx context.Context, me *domain.Me) (*domain.Profile, error) {
	result, err := uc.profiles.Get(ctx, me)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch profile info: %w", err)
	}

	return result, nil
}
