package memory_test

import (
	"context"
	"path"
	"reflect"
	"sync"
	"testing"
	"time"

	"source.toby3d.me/website/indieauth/internal/domain"
	repository "source.toby3d.me/website/indieauth/internal/ticket/repository/memory"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	store := new(sync.Map)
	ticket := domain.TestTicket(t)

	if err := repository.NewMemoryTicketRepository(store, domain.TestConfig(t)).
		Create(context.TODO(), ticket); err != nil {
		t.Fatal(err)
	}

	storePath := path.Join(repository.DefaultPathPrefix, ticket.Ticket)

	src, ok := store.Load(storePath)
	if !ok {
		t.Fatalf("Load(%s) = %t, want %t", storePath, ok, true)
	}

	if result, _ := src.(*repository.Ticket); !reflect.DeepEqual(result.Ticket, ticket) {
		t.Errorf("Create(%+v) = %+v, want %+v", ticket, result.Ticket, ticket)
	}
}

func TestGetAndDelete(t *testing.T) {
	t.Parallel()

	ticket := domain.TestTicket(t)

	store := new(sync.Map)
	store.Store(path.Join(repository.DefaultPathPrefix, ticket.Ticket), &repository.Ticket{
		CreatedAt: time.Now().UTC(),
		Ticket:    ticket,
	})

	result, err := repository.NewMemoryTicketRepository(store, domain.TestConfig(t)).
		GetAndDelete(context.TODO(), ticket.Ticket)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, ticket) {
		t.Errorf("GetAndDelete(%s) = %+v, want %+v", ticket.Ticket, result, ticket)
	}

	storePath := path.Join(repository.DefaultPathPrefix, ticket.Ticket)
	if src, _ := store.Load(storePath); src != nil {
		t.Errorf("Load(%s) = %+v, want %+v", storePath, src, nil)
	}
}
