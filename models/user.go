package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)


type User struct{
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
    Username  string         `gorm:"uniqueIndex;not null" json:"username"`
    Email     string         `gorm:"uniqueIndex;not null" json:"email"`
    Password  string         `gorm:"not null" json:"-"` // json:"-" hides password in responses
    Images    []Image        `gorm:"foreignKey:UserID" json:"images,omitempty"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // soft delete support

}

// BeforeCreate hook — runs automatically before INSERT.
// We generate a UUID here so we never get a blank ID.
func (u *User) BeforeCreate(tx *gorm.DB) error{
	u.ID=uuid.New()
	return nil
}

