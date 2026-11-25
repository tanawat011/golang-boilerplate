package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tanawat011/golang-boilerplate/internal/model"
	"github.com/tanawat011/golang-boilerplate/internal/service"
)

type fixture struct {
	users service.UserService
	alice service.Actor
	bob   service.Actor
	admin service.Actor
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	authSvc, repo, _ := newAuth(t)
	a := register(t, authSvc, "alice", "alice@example.com")
	b := register(t, authSvc, "bob", "bob@example.com")
	root := register(t, authSvc, "root", "root@example.com")
	repo.SetRole(root.ID, model.RoleAdmin)
	return fixture{
		users: service.NewUserService(repo),
		alice: service.Actor{UserID: a.ID, Role: model.RoleUser},
		bob:   service.Actor{UserID: b.ID, Role: model.RoleUser},
		admin: service.Actor{UserID: root.ID, Role: model.RoleAdmin},
	}
}

func TestGetByIDOwnershipRules(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	if u, err := f.users.GetByID(ctx, f.alice, f.alice.UserID); err != nil || u.Username != "alice" {
		t.Errorf("self read: %v %+v", err, u)
	}
	if _, err := f.users.GetByID(ctx, f.alice, f.bob.UserID); !errors.Is(err, service.ErrForbidden) {
		t.Errorf("reading another user: err = %v, want ErrForbidden", err)
	}
	if _, err := f.users.GetByID(ctx, f.admin, f.bob.UserID); err != nil {
		t.Errorf("admin read: %v", err)
	}
	if _, err := f.users.GetByID(ctx, f.admin, "00000000-0000-0000-0000-000000000000"); !errors.Is(err, service.ErrUserNotFound) {
		t.Errorf("missing user: err = %v", err)
	}
}

func TestListIsAdminOnly(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	if _, _, err := f.users.List(ctx, f.alice, 1, 10); !errors.Is(err, service.ErrForbidden) {
		t.Errorf("user list: err = %v", err)
	}

	page, total, err := f.users.List(ctx, f.admin, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(page) != 2 || page[0].Username != "root" {
		t.Errorf("page 1 = %+v, total %d", page, total)
	}
	page, _, _ = f.users.List(ctx, f.admin, 3, 2)
	if page == nil || len(page) != 0 {
		t.Errorf("past the end: %#v, want an empty (non-nil) slice", page)
	}
}

func TestUpdateRules(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	yes := true
	name := "Alice Liddell"

	u, err := f.users.Update(ctx, f.alice, f.alice.UserID, service.UpdateUserRequest{FullName: &name, Email: "Alice2@Example.com"})
	if err != nil {
		t.Fatalf("self update: %v", err)
	}
	if u.FullName != name || u.Email != "alice2@example.com" {
		t.Errorf("unexpected: %+v", u)
	}

	cases := map[string]struct {
		actor service.Actor
		id    string
		req   service.UpdateUserRequest
		want  error
	}{
		"user edits someone else":    {f.alice, f.bob.UserID, service.UpdateUserRequest{FullName: &name}, service.ErrForbidden},
		"user promotes themself":     {f.alice, f.alice.UserID, service.UpdateUserRequest{Role: model.RoleAdmin}, service.ErrForbidden},
		"user reactivates themself":  {f.alice, f.alice.UserID, service.UpdateUserRequest{IsActive: &yes}, service.ErrForbidden},
		"email taken":                {f.alice, f.alice.UserID, service.UpdateUserRequest{Email: "bob@example.com"}, service.ErrEmailAlreadyExists},
		"username taken":             {f.alice, f.alice.UserID, service.UpdateUserRequest{Username: "bob"}, service.ErrUsernameExists},
		"admin edits a missing user": {f.admin, "00000000-0000-0000-0000-000000000000", service.UpdateUserRequest{FullName: &name}, service.ErrUserNotFound},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := f.users.Update(ctx, tc.actor, tc.id, tc.req); !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}

	promoted, err := f.users.Update(ctx, f.admin, f.bob.UserID, service.UpdateUserRequest{Role: model.RoleAdmin})
	if err != nil || promoted.Role != model.RoleAdmin {
		t.Errorf("admin promotes bob: %v %+v", err, promoted)
	}
}

func TestDeleteRules(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	if err := f.users.Delete(ctx, f.alice, f.bob.UserID); !errors.Is(err, service.ErrForbidden) {
		t.Errorf("deleting someone else: err = %v", err)
	}
	if err := f.users.Delete(ctx, f.alice, f.alice.UserID); err != nil {
		t.Errorf("self delete: %v", err)
	}
	if _, err := f.users.GetByID(ctx, f.admin, f.alice.UserID); !errors.Is(err, service.ErrUserNotFound) {
		t.Errorf("deleted user still readable: %v", err)
	}
	if err := f.users.Delete(ctx, f.admin, f.alice.UserID); !errors.Is(err, service.ErrUserNotFound) {
		t.Errorf("deleting twice: err = %v", err)
	}
}
