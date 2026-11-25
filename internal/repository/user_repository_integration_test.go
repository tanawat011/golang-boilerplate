//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tanawat011/golang-boilerplate/internal/model"
	"github.com/tanawat011/golang-boilerplate/internal/repository"
	"github.com/tanawat011/golang-boilerplate/internal/testutil"
)

func newUser(name string) *model.User {
	return &model.User{
		Username: name,
		Email:    name + "@example.com",
		Password: "hash",
		Role:     model.RoleUser,
		IsActive: true,
	}
}

func TestUserRepositoryAgainstPostgres(t *testing.T) {
	db := testutil.NewDatabase(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	alice := newUser("alice")
	if err := repo.Create(ctx, alice); err != nil {
		t.Fatalf("create: %v", err)
	}
	if alice.ID == "" || alice.CreatedAt.IsZero() {
		t.Fatalf("ID / timestamps not set: %+v", alice)
	}

	t.Run("find", func(t *testing.T) {
		for name, find := range map[string]func() (*model.User, error){
			"by id":       func() (*model.User, error) { return repo.FindByID(ctx, alice.ID) },
			"by email":    func() (*model.User, error) { return repo.FindByEmail(ctx, "alice@example.com") },
			"by username": func() (*model.User, error) { return repo.FindByUsername(ctx, "alice") },
		} {
			u, err := find()
			if err != nil || u.ID != alice.ID {
				t.Errorf("%s: %v %+v", name, err, u)
			}
		}
		if _, err := repo.FindByEmail(ctx, "nobody@example.com"); !errors.Is(err, repository.ErrNotFound) {
			t.Errorf("missing user: %v", err)
		}
	})

	t.Run("unique constraints map to ErrDuplicate", func(t *testing.T) {
		dup := newUser("alice2")
		dup.Email = "alice@example.com"
		if err := repo.Create(ctx, dup); !errors.Is(err, repository.ErrDuplicate) {
			t.Errorf("duplicate email: %v", err)
		}
	})

	t.Run("update", func(t *testing.T) {
		alice.FullName = "Alice Liddell"
		if err := repo.Update(ctx, alice); err != nil {
			t.Fatal(err)
		}
		u, _ := repo.FindByID(ctx, alice.ID)
		if u.FullName != "Alice Liddell" {
			t.Errorf("not saved: %+v", u)
		}
	})

	t.Run("list newest first with total", func(t *testing.T) {
		for _, n := range []string{"bob", "carol"} {
			if err := repo.Create(ctx, newUser(n)); err != nil {
				t.Fatal(err)
			}
		}
		users, total, err := repo.FindAll(ctx, 0, 2)
		if err != nil {
			t.Fatal(err)
		}
		if total != 3 || len(users) != 2 || users[0].Username != "carol" {
			t.Errorf("page: total %d, %v", total, usernames(users))
		}
	})

	t.Run("soft delete", func(t *testing.T) {
		if err := repo.Delete(ctx, alice.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := repo.FindByID(ctx, alice.ID); !errors.Is(err, repository.ErrNotFound) {
			t.Errorf("deleted user found: %v", err)
		}
		if err := repo.Delete(ctx, alice.ID); !errors.Is(err, repository.ErrNotFound) {
			t.Errorf("second delete: %v", err)
		}
		var count int64
		db.Unscoped().Model(&model.User{}).Where("id = ?", alice.ID).Count(&count)
		if count != 1 {
			t.Errorf("row was hard-deleted")
		}
		// The address of a deleted account stays reserved.
		again := newUser("alice3")
		again.Email = "alice@example.com"
		if err := repo.Create(ctx, again); !errors.Is(err, repository.ErrDuplicate) {
			t.Errorf("re-using a deleted account's email: %v", err)
		}
	})
}

func usernames(users []model.User) []string {
	out := make([]string, len(users))
	for i, u := range users {
		out[i] = u.Username
	}
	return out
}
