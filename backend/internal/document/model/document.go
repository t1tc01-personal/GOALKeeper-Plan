package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DocumentUploadStatus represents the upload status of a document
type DocumentUploadStatus string

const (
	DocumentStatusUploading  DocumentUploadStatus = "uploading"
	DocumentStatusProcessing DocumentUploadStatus = "processing"
	DocumentStatusCompleted  DocumentUploadStatus = "completed"
	DocumentStatusFailed     DocumentUploadStatus = "failed"
)

// Document represents user-uploaded files that enhance AI knowledge
type Document struct {
	ID               uuid.UUID          `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID           `gorm:"type:uuid;not null;index" json:"user_id"`
	Name             string              `gorm:"type:varchar(255);not null" json:"name"`
	FileType         string              `gorm:"type:varchar(100);not null" json:"file_type"`
	FileSizeBytes    int64               `gorm:"not null" json:"file_size_bytes"`
	R2Key            string              `gorm:"type:varchar(500);not null;uniqueIndex" json:"r2_key"`
	ProcessedContent *string             `gorm:"type:text" json:"processed_content,omitempty"`
	ExtractedInsights *string             `gorm:"type:jsonb" json:"extracted_insights,omitempty"`
	UploadStatus     DocumentUploadStatus `gorm:"type:varchar(20);not null;default:'uploading';index" json:"upload_status"`
	UploadedAt       time.Time           `gorm:"not null" json:"uploaded_at"`
	ProcessedAt      *time.Time          `gorm:"type:timestamp" json:"processed_at,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`

	// Relationships
	User interface{} `gorm:"-" json:"user,omitempty"`
}

// BeforeCreate hook to generate UUID
func (d *Document) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (Document) TableName() string {
	return "documents"
}

