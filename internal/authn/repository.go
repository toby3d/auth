package authn

import "context"

type Repository interface {
	Fetch(ctx context.Context, me string) ([]string, error)
}
