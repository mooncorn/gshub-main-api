package repositories

import (
	"github.com/mooncorn/gshub-main-api/models"
	"gorm.io/gorm"
)

type ServiceRepository struct {
	DB *gorm.DB
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{DB: db}
}

func (r *ServiceRepository) GetService(serviceID uint) (*models.Service, error) {
	var service models.Service
	err := r.DB.First(&service, serviceID).Error
	return &service, err
}

func (r *ServiceRepository) GetServicePreloaded(serviceID uint) (*models.Service, error) {
	var service models.Service
	err := r.DB.Model(&models.Service{}).Preload("Env.Values").Preload("Ports").Preload("Volumes").Where(serviceID).First(&service).Error
	return &service, err
}

func (r *ServiceRepository) GetServices() (*[]models.Service, error) {
	var services []models.Service
	err := r.DB.Find(&services).Error
	return &services, err
}
