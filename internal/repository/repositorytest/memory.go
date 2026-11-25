// Package repositorytest provides an in-memory UserRepository for unit tests.
package repositorytest

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/tanawat011/golang-boilerplate/internal/model"
	"github.com/tanawat011/golang-boilerplate/internal/repository"
)

// UserRepository is a goroutine-safe in-memory repository.UserRepository.
// Like the GORM one, it enforces unique emails and usernames and soft-deletes
// (a deleted user's email and username stay taken).
type UserRepository struct {
	mu      sync.Mutex
	users   map[string]model.User
	deleted map[string]bool
	clock   time.Time

	// Err, when set, is returned by every call (to test error paths).
	Err error
}

var _ repository.UserRepository = (*UserRepository)(nil)

// NewUserRepository returns an empty repository.
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:   map[string]model.User{},
		deleted: map[string]bool{},
		clock:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// Create stores a copy of user, assigning an ID and timestamps.
func (r *UserRepository) Create(_ context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Err != nil {
		return r.Err
	}
	for _, u := range r.users {
		if u.Email == user.Email || u.Username == user.Username {
			return repository.ErrDuplicate
		}
	}
	if user.ID == "" {
		user.ID = uuid.NewString()
	}
	r.clock = r.clock.Add(time.Second)
	user.CreatedAt, user.UpdatedAt = r.clock, r.clock
	r.users[user.ID] = *user
	return nil
}

// FindByID returns a live (not deleted) user.
func (r *UserRepository) FindByID(_ context.Context, id string) (*model.User, error) {
	return r.find(func(u model.User) bool { return u.ID == id })
}

// FindByEmail returns a live user by email.
func (r *UserRepository) FindByEmail(_ context.Context, email string) (*model.User, error) {
	return r.find(func(u model.User) bool { return u.Email == email })
}

// FindByUsername returns a live user by username.
func (r *UserRepository) FindByUsername(_ context.Context, username string) (*model.User, error) {
	return r.find(func(u model.User) bool { return u.Username == username })
}

// FindAll returns live users, newest first.
func (r *UserRepository) FindAll(_ context.Context, offset, limit int) ([]model.User, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Err != nil {
		return nil, 0, r.Err
	}
	var live []model.User
	for id, u := range r.users {
		if !r.deleted[id] {
			live = append(live, u)
		}
	}
	sort.Slice(live, func(i, j int) bool { return live[i].CreatedAt.After(live[j].CreatedAt) })
	total := int64(len(live))
	if offset >= len(live) {
		return []model.User{}, total, nil
	}
	end := min(offset+limit, len(live))
	return live[offset:end], total, nil
}

// Update replaces a stored user.
func (r *UserRepository) Update(_ context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Err != nil {
		return r.Err
	}
	for id, u := range r.users {
		if id != user.ID && (u.Email == user.Email || u.Username == user.Username) {
			return repository.ErrDuplicate
		}
	}
	r.clock = r.clock.Add(time.Second)
	user.UpdatedAt = r.clock
	r.users[user.ID] = *user
	return nil
}

// Delete soft-deletes a user.
func (r *UserRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Err != nil {
		return r.Err
	}
	if _, ok := r.users[id]; !ok || r.deleted[id] {
		return repository.ErrNotFound
	}
	r.deleted[id] = true
	return nil
}

// Get returns the stored user with that ID, deleted or not (for assertions).
func (r *UserRepository) Get(id string) (model.User, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	return u, ok
}

// SetRole changes a stored user's role (to set up admins in tests).
func (r *UserRepository) SetRole(id, role string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u := r.users[id]
	u.Role = role
	r.users[id] = u
}

func (r *UserRepository) find(match func(model.User) bool) (*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Err != nil {
		return nil, r.Err
	}
	for id, u := range r.users {
		if !r.deleted[id] && match(u) {
			found := u
			return &found, nil
		}
	}
	return nil, repository.ErrNotFound
}
