package service_repositories

import (
	"github.com/mooncorn/gshub-main-api/service/service_models"
	"gorm.io/gorm"
)

type ServiceRepository struct {
	DB *gorm.DB
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{DB: db}
}

func (r *ServiceRepository) GetService(serviceID uint) (*service_models.Service, error) {
	var service service_models.Service
	err := r.DB.First(&service, serviceID).Error
	return &service, err
}

func (r *ServiceRepository) GetServices() (*[]service_models.Service, error) {
	var services []service_models.Service
	err := r.DB.Find(&services).Error
	return &services, err
}
