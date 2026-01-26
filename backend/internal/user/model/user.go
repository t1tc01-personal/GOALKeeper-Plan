package model

import (
	"time"

	rbacModel "goalkeeper-plan/internal/rbac/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents an application user
type User struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email           string     `gorm:"uniqueIndex;not null" json:"email"`
	Name            string     `gorm:"not null;size:100" json:"name"`
	PasswordHash    string     `gorm:"size:255" json:"-"`                                   // Hidden from JSON
	OAuthProvider   string     `gorm:"column:oauth_provider;size:50" json:"oauth_provider"` // github, google, or empty
	OAuthProviderID string     `gorm:"column:oauth_provider_id;size:255" json:"-"`          // OAuth provider user ID
	IsEmailVerified bool       `gorm:"default:false" json:"is_email_verified"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`

	// RBAC relationships
	Roles []rbacModel.Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// BeforeCreate hook to generate UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (User) TableName() string {
	return "users"
}
