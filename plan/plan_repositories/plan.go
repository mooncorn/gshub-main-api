package plan_repositories

import (
	"github.com/mooncorn/gshub-main-api/plan/plan_models"
	"gorm.io/gorm"
)

type PlanRepository struct {
	DB *gorm.DB
}

func NewPlanRepository(db *gorm.DB) *PlanRepository {
	return &PlanRepository{DB: db}
}

func (r *PlanRepository) GetPlan(planID uint) (*plan_models.Plan, error) {
	var plan plan_models.Plan
	err := r.DB.First(&plan, planID).Error
	return &plan, err
}

func (r *PlanRepository) GetPlans() (*[]plan_models.Plan, error) {
	var plans []plan_models.Plan
	err := r.DB.Find(&plans).Error
	return &plans, err
}
