package user_repositories

import (
	"github.com/mooncorn/gshub-main-api/user/user_models"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) GetUser(ID uint) (*user_models.User, error) {
	var user user_models.User
	err := r.DB.Where("id = ?", ID).First(&user).Error
	return &user, err
}

func (r *UserRepository) GetUserByEmail(email string) (*user_models.User, error) {
	var user user_models.User
	err := r.DB.Where(&user_models.User{Email: email}).First(&user).Error
	return &user, err
}

func (r *UserRepository) CreateUser(user *user_models.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRepository) UpdateUser(existingUser *user_models.User, updatedUser *user_models.User) error {
	return r.DB.Model(&existingUser).Updates(*updatedUser).Error
}
