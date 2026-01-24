package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GoalStatus represents the status of a goal
type GoalStatus string

const (
	GoalStatusActive    GoalStatus = "active"
	GoalStatusPaused    GoalStatus = "paused"
	GoalStatusCompleted GoalStatus = "completed"
	GoalStatusAbandoned GoalStatus = "abandoned"
)

// Goal represents a user's objective or desired outcome
type Goal struct {
	ID                  uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID              uuid.UUID   `gorm:"type:uuid;not null;index" json:"user_id"`
	Description         string      `gorm:"type:text;not null" json:"description"`
	TargetCompletionDate *time.Time `gorm:"type:timestamp" json:"target_completion_date,omitempty"`
	Status              GoalStatus  `gorm:"type:varchar(20);not null;default:'active';index" json:"status"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`

	// Relationships
	User     interface{} `gorm:"-" json:"user,omitempty"`
	Plan     interface{} `gorm:"-" json:"plan,omitempty"`
	Progress interface{} `gorm:"-" json:"progress,omitempty"`
}

// BeforeCreate hook to generate UUID
func (g *Goal) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (Goal) TableName() string {
	return "goals"
}

