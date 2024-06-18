package db

import (
	"github.com/mooncorn/gshub-core/models"
	"gorm.io/gorm"
)

type PlanRepository struct {
	DB *gorm.DB
}

func NewPlanRepository(db *gorm.DB) *PlanRepository {
	return &PlanRepository{DB: db}
}

func (r *PlanRepository) GetPlan(planID int) (*models.Plan, error) {
	var plan models.Plan
	err := r.DB.First(&plan, planID).Error
	return &plan, err
}

func (r *PlanRepository) GetPlans() (*[]models.Plan, error) {
	var plans []models.Plan
	err := r.DB.Find(&plans).Error
	return &plans, err
}
