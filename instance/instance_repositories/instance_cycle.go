package instance_repositories

import (
	"database/sql"

	"github.com/mooncorn/gshub-main-api/instance/instance_models"
	"gorm.io/gorm"
)

type InstanceCyclesRepository struct {
	DB *gorm.DB
}

func NewInstanceCyclesRepository(db *gorm.DB) *InstanceCyclesRepository {
	return &InstanceCyclesRepository{DB: db}
}

func (r *InstanceCyclesRepository) CreateInstanceCycles(cycle *instance_models.InstanceCycle) error {
	return r.DB.Create(cycle).Error
}

func (r *InstanceCyclesRepository) GetInstanceCycles(instanceID uint) (*[]instance_models.InstanceCycle, error) {
	var instanceCycles []instance_models.InstanceCycle
	err := r.DB.Where("instance_id = ?", instanceID).Find(&instanceCycles).Error
	return &instanceCycles, err
}

func (r *InstanceCyclesRepository) GetInstanceCycle(instanceCycleID uint) (*instance_models.InstanceCycle, error) {
	var instanceCycles instance_models.InstanceCycle
	err := r.DB.Where("id = ?", instanceCycleID).First(&instanceCycles).Error
	return &instanceCycles, err
}

func (r *InstanceCyclesRepository) GetInstanceCyclesSum(instanceID uint) (uint, error) {
	var sum sql.NullInt64
	row := r.DB.Model(&instance_models.InstanceCycle{}).Where("instance_id = ?", instanceID).Select("SUM(amount)").Row()
	err := row.Scan(&sum)
	if err != nil {
		return 0, err
	}

	if !sum.Valid {
		return 0, nil
	}
	return uint(sum.Int64), nil
}
