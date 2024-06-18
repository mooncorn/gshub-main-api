package db

import (
	"errors"

	"github.com/mooncorn/gshub-core/models"
	"gorm.io/gorm"
)

type ServerRepository struct {
	DB *gorm.DB
}

func NewServerRepository(db *gorm.DB) *ServerRepository {
	return &ServerRepository{DB: db}
}

func (r *ServerRepository) GetUserServers(userID uint) (*[]models.Server, error) {
	var servers []models.Server
	if err := r.DB.Where("user_id = ?", userID).Find(&servers).Error; err != nil {
		return nil, err
	}
	return &servers, nil
}

func (r *ServerRepository) GetServerByInstanceIDAndUserEmail(instanceID, userEmail string) (*models.Server, error) {
	var server models.Server
	err := r.DB.Where("instance_id = ? AND user_id = (SELECT id FROM users WHERE email = ?)", instanceID, userEmail).First(&server).Error
	return &server, err
}

func (r *ServerRepository) GetServerByServerIDAndUserEmail(serverID, userEmail string) (*models.Server, error) {
	var server models.Server
	err := r.DB.Where("id = ? AND user_id = (SELECT id FROM users WHERE email = ?)", serverID, userEmail).First(&server).Error
	return &server, err
}

func (r *ServerRepository) CreateServer(server *models.Server) error {
	return r.DB.Create(server).Error
}

func (r *ServerRepository) DeleteServer(serverId string, userEmail string) error {
	result := r.DB.Where("id = ? AND user_id = (SELECT id FROM users WHERE email = ?)", serverId, userEmail).Delete(&models.Server{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no matching record found")
	}

	return nil
}
