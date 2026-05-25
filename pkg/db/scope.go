package db

import "gorm.io/gorm"

// NotDeleted is a global scope that filters out soft-deleted records.
func NotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("deleted_at IS NULL")
}
