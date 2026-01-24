package repository

import (
	"goalkeeper-plan/internal/user/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *model.User) error
	GetByID(id string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	Update(user *model.User) error
	Delete(id string) error
	List(limit, offset int) ([]*model.User, error)
	WithTrx(tx *gorm.DB) UserRepository
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) WithTrx(tx *gorm.DB) UserRepository {
	return &userRepository{db: tx}
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id string) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id string) error {
	return r.db.Delete(&model.User{}, "id = ?", id).Error
}

func (r *userRepository) List(limit, offset int) ([]*model.User, error) {
	var users []*model.User
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

