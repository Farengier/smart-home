package orm

import "gorm.io/gorm"

type UserRole struct {
	gorm.Model
	UserID int
	Role   string
}
