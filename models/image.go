package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)



type Image struct {
    ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
    UserID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
    OriginalURL     string         `gorm:"not null" json:"original_url"`
    TransformedURL  string         `json:"transformed_url,omitempty"`
    Filename        string         `gorm:"not null" json:"filename"`
    Format          string         `json:"format"`          // jpeg, png, webp
    Width           int            `json:"width"`
    Height          int            `json:"height"`
    Size            int64          `json:"size"`            // bytes
    Status          string         `gorm:"default:'uploaded'" json:"status"` // uploaded | processing | done
    CreatedAt       time.Time      `json:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at"`
    DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (i *Image) BeforeCreate(tx *gorm.DB) error{
	i.ID=uuid.New()
	return nil
}