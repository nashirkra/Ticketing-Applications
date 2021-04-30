package entity

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

/**
 * This is a Event's entity class and is not intended for modification.
 */
type Event struct {
	ID                int            `gorm:"primary_key:auto_increment" json:"id,string"`
	CreatorId         int            `gorm:"not null" json:"creator_id,string"`
	TitleEvent        string         `gorm:"type:varchar(255);not null" json:"title_event"`
	LinkWebinar       string         `gorm:"not null" json:"link_webinar"`
	Description       string         `gorm:"not null" json:"description"`
	TypeEvent         string         `gorm:"type:enum('Online','Offline');not null" json:"type_event"`
	Banner            string         `gorm:"not null" json:"banner"`
	Price             float64        `gorm:"type:double" json:"price,string"`
	Quantity          int            `gorm:"not null" json:"quantity,string"`
	Status            string         `gorm:"type:enum('Draft','Scheduled','Canceled','In Progress','Stopped','Completed','Completed and Verified','Closed');not null" json:"status"`
	EventStartDate    *time.Time     `gorm:"type:date;not null" json:"event_start_date" time_format:"2006-01-02"`
	EventEndDate      *time.Time     `gorm:"type:date;not null" json:"event_end_date" time_format:"2006-01-02"`
	CampaignStartDate *time.Time     `gorm:"type:date;not null" json:"campaign_start_date" time_format:"2006-01-02"`
	CampaignEndDate   *time.Time     `gorm:"type:date;not null" json:"campaign_end_date" time_format:"2006-01-02"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" time_format:"unixNano"`
	CreatedAt         time.Time      `gorm:"<-:create;not null" json:"created_at" time_format:"unixNano"`
	UpdatedAt         time.Time      `gorm:"not null" json:"updated_at" time_format:"unixNano"`
	Creator           User           `gorm:"foreignkey:CreatorId;constraint:onUpdate:CASCADE;onDelete:CASCADE" json:"creator,omitempty"`
}

func (ev *Event) MarshalBinary() ([]byte, error) {
	return json.Marshal(ev)
}

func (ev *Event) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &ev); err != nil {
		return err
	}

	return nil
}

/*
func (u *Event) UnmarshalJSON(b []byte) error {
	u.Creator = User{}
	if err := json.Unmarshal(b, &u); err != nil {
		return err
	}
	return nil
} */
