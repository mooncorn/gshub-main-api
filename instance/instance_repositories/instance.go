package instance_repositories

import (
	"errors"

	"github.com/mooncorn/gshub-main-api/instance/instance_models"
	"gorm.io/gorm"
)

type InstanceRepository struct {
	DB *gorm.DB
}

func NewInstanceRepository(db *gorm.DB) *InstanceRepository {
	return &InstanceRepository{DB: db}
}

func (r *InstanceRepository) GetUserInstances(userID uint) (*[]instance_models.Instance, error) {
	var instances []instance_models.Instance
	if err := r.DB.Where("user_id = ?", userID).Find(&instances).Error; err != nil {
		return nil, err
	}
	return &instances, nil
}

func (r *InstanceRepository) GetUserInstance(userEmail string, instanceID uint) (*instance_models.Instance, error) {
	var instance instance_models.Instance
	err := r.DB.Where("id = ? AND user_id = (SELECT id FROM users WHERE email = ?)", instanceID, userEmail).First(&instance).Error
	return &instance, err
}

func (r *InstanceRepository) GetInstance(instanceID uint) (*instance_models.Instance, error) {
	var instance instance_models.Instance
	err := r.DB.Where("id = ?", instanceID).First(&instance).Error
	return &instance, err
}

func (r *InstanceRepository) CreateInstance(instance *instance_models.Instance) error {
	return r.DB.Create(instance).Error
}

func (r *InstanceRepository) DeleteUserInstance(userEmail string, instanceID uint) error {
	result := r.DB.Where("id = ? AND user_id = (SELECT id FROM users WHERE email = ?)", instanceID, userEmail).Delete(&instance_models.Instance{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no matching record found")
	}

	return nil
}

func (r *InstanceRepository) SaveInstance(instance *instance_models.Instance) error {
	return r.DB.Save(instance).Error
}
