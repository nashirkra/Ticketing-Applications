package entity

import (
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	ID            int            `gorm:"primary_key:auto_increment" json:"id,string"`
	ParticipantId int            `gorm:"not null" json:"participant_id,string"`
	CreatorId     int            `gorm:"not null" json:"creator_id,string"`
	EventId       int            `gorm:"not null" json:"event_id,string"`
	Amount        float64        `gorm:"type:double" json:"amount,string"`
	StatusPayment string         `gorm:"type:enum('Pending','Processing','Completed','Refund & Cancelled','Cancelled');not null" json:"status_payment"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" time_format:"unixNano"`
	CreatedAt     time.Time      `gorm:"<-:create;not null" json:"created_at" time_format:"unixNano"`
	UpdatedAt     time.Time      `gorm:"not null" json:"updated_at" time_format:"unixNano"`
	Participant   User           `gorm:"foreignkey:ParticipantId;constraint:onUpdate:CASCADE;onDelete:CASCADE" json:"participant,omitempty"`
	Event         Event          `gorm:"foreignkey:EventId;constraint:onUpdate:CASCADE;onDelete:CASCADE" json:"event,omitempty"`
}
