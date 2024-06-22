package repositories

import (
	"github.com/mooncorn/gshub-main-api/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.DB.Where(&models.User{Email: email}).First(&user).Error
	return &user, err
}

func (r *UserRepository) CreateUser(user *models.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRepository) UpdateUser(existingUser *models.User, updatedUser *models.User) error {
	return r.DB.Model(&existingUser).Updates(*updatedUser).Error
}
