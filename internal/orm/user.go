package orm

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Login  string
	OtpKey string
	Role   UserRole
}
