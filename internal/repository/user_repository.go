// Package repository is the data-access layer.
package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/tanawat011/golang-boilerplate/internal/model"
)

// Errors returned by repositories. Callers never see GORM errors directly.
var (
	ErrNotFound  = errors.New("record not found")
	ErrDuplicate = errors.New("duplicate record")
)

// UserRepository defines the data access operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindAll(ctx context.Context, offset, limit int) ([]model.User, int64, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a GORM-backed user repository. The *gorm.DB must
// be opened with TranslateError: true so unique violations map to ErrDuplicate.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return translate(r.db.WithContext(ctx).Create(user).Error)
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.findOne(ctx, "email = ?", email)
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	return r.findOne(ctx, "username = ?", username)
}

func (r *userRepository) FindAll(ctx context.Context, offset, limit int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	db := r.db.WithContext(ctx)
	if err := db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("created_at DESC").Order("id").Offset(offset).Limit(limit).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return translate(r.db.WithContext(ctx).Save(user).Error)
}

// Delete soft-deletes a user.
func (r *userRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.User{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *userRepository) findOne(ctx context.Context, query string, arg any) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where(query, arg).First(&user).Error; err != nil {
		return nil, translate(err)
	}
	return &user, nil
}

func translate(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return ErrDuplicate
	default:
		return err
	}
}
