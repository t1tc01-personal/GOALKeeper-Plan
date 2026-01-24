package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Progress represents tracking information for goal advancement
type Progress struct {
	ID                      uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	GoalID                  uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"goal_id"`
	CompletionPercentage    float64    `gorm:"type:float;not null;default:0.0" json:"completion_percentage"`
	TasksCompletedCount     int        `gorm:"not null;default:0" json:"tasks_completed_count"`
	TasksRemainingCount     int        `gorm:"not null;default:0" json:"tasks_remaining_count"`
	LastActivityAt          *time.Time `gorm:"type:timestamp" json:"last_activity_at,omitempty"`
	EstimatedCompletionDate *time.Time `gorm:"type:timestamp" json:"estimated_completion_date,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`

	// Relationships
	Goal interface{} `gorm:"-" json:"goal,omitempty"`
}

// BeforeCreate hook to generate UUID
func (p *Progress) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (Progress) TableName() string {
	return "progress"
}

