package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusToDo       TaskStatus = "to_do"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusBlocked    TaskStatus = "blocked"
)

// Task represents an actionable step within a plan
type Task struct {
	ID                        uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PlanID                    uuid.UUID   `gorm:"type:uuid;not null;index" json:"plan_id"`
	Description               string      `gorm:"type:text;not null" json:"description"`
	Rationale                 *string     `gorm:"type:text" json:"rationale,omitempty"`
	Status                    TaskStatus  `gorm:"type:varchar(20);not null;default:'to_do';index" json:"status"`
	Order                     int         `gorm:"not null;index" json:"order"`
	EstimatedEffortHours      *float64    `gorm:"type:float" json:"estimated_effort_hours,omitempty"`
	ActualCompletionTimeHours *float64    `gorm:"type:float" json:"actual_completion_time_hours,omitempty"`
	CompletedAt               *time.Time  `gorm:"type:timestamp" json:"completed_at,omitempty"`
	DependsOnTaskID           *uuid.UUID  `gorm:"type:uuid;index" json:"depends_on_task_id,omitempty"`
	CreatedAt                 time.Time   `json:"created_at"`
	UpdatedAt                 time.Time   `json:"updated_at"`

	// Relationships
	Plan          interface{} `gorm:"-" json:"plan,omitempty"`
	DependsOnTask interface{} `gorm:"-" json:"depends_on_task,omitempty"`
}

// BeforeCreate hook to generate UUID
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (Task) TableName() string {
	return "tasks"
}

