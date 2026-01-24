package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Plan represents the AI-generated strategy to achieve a goal
type Plan struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	GoalID         uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"goal_id"`
	Name           string     `gorm:"type:varchar(200);not null" json:"name"`
	Description    *string    `gorm:"type:text" json:"description,omitempty"`
	ConfidenceScore *float64   `gorm:"type:float" json:"confidence_score,omitempty"`
	GeneratedAt    time.Time  `gorm:"not null" json:"generated_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relationships
	Goal  interface{} `gorm:"-" json:"goal,omitempty"`
	Tasks interface{} `gorm:"-" json:"tasks,omitempty"`
}

// BeforeCreate hook to generate UUID
func (p *Plan) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (Plan) TableName() string {
	return "plans"
}

