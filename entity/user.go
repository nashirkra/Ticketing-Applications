package entity

/**
 * This is a User's entity class and is not intended for modification.
 */
type User struct {
	ID       int    `gorm:"primary_key:auto_increment" json:"id,string"`
	Username string `gorm:"uniqueIndex;type:varchar(255);not null" json:"username"`
	Fullname string `gorm:"type:varchar(255);not null" json:"fullname"`
	Email    string `gorm:"uniqueIndex;type:varchar(255);not null" json:"email"`
	// EmailVerifiedAt time.Time `gorm:"<-:create"`
	Password  string `gorm:"->;<-;not null" json:"-"`
	Role      string `gorm:"type:enum('admin','creator','participant');not null" json:"role"`
	DeletedAt int64  `json:"deleted_at,string"`
	CreatedAt int64  `gorm:"<-:create;not null" json:"created_at,string"`
	UpdatedAt int64  `gorm:"not null" json:"updated_at,string"`
	Token     string `json:"token,omitempty"`
}
