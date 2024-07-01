package instance_repositories

import (
	"github.com/mooncorn/gshub-main-api/instance/instance_models"
	"gorm.io/gorm"
)

type InstanceBurnedCyclesRepository struct {
	DB *gorm.DB
}

func NewInstanceBurnedCyclesRepository(db *gorm.DB) *InstanceBurnedCyclesRepository {
	return &InstanceBurnedCyclesRepository{DB: db}
}

func (r *InstanceBurnedCyclesRepository) CreateBurnedInstanceBurnedCycles(burnedCycle *instance_models.InstanceBurnedCycle) error {
	return r.DB.Create(burnedCycle).Error
}

func (r *InstanceBurnedCyclesRepository) GetInstanceBurnedCycles(instanceID uint) (*[]instance_models.InstanceBurnedCycle, error) {
	var instanceBurnedCycles []instance_models.InstanceBurnedCycle
	err := r.DB.Where("instance_id = ?", instanceID).Find(&instanceBurnedCycles).Error
	return &instanceBurnedCycles, err
}

func (r *InstanceBurnedCyclesRepository) GetInstanceBurnedCycle(InstanceBurnedCycleID uint) (*instance_models.InstanceBurnedCycle, error) {
	var instanceBurnedCycles instance_models.InstanceBurnedCycle
	err := r.DB.Where("id = ?", InstanceBurnedCycleID).First(&instanceBurnedCycles).Error
	return &instanceBurnedCycles, err
}

func (r *InstanceBurnedCyclesRepository) GetInstanceBurnedCyclesSum(instanceID uint) (uint, error) {
	var sum uint
	err := r.DB.Where("instance_id = ?", instanceID).Select("SUM(amount)").Row().Scan(&sum)
	return sum, err
}
