package main

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Event matches the schema defined in Calendar API Spec.md
type Event struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Date        time.Time `json:"date" gorm:"type:date;not null"`
	Title       string    `json:"title" gorm:"size:255;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Tags        string    `json:"tags" gorm:"size:500"`        // JSON array as string
	Media       string    `json:"media" gorm:"type:text"`      // Link to media file
	References  string    `json:"references" gorm:"type:text"` // JSON array as string
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// InitDB initializes the database connection and migrates the schema.
// It now returns the DB instance or an error.
func InitDB(dbPath string) (*gorm.DB, error) {
	var err error
	var localDB *gorm.DB // Use a local variable for the DB instance
	localDB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Or logger.Info for more logs
	})
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	err = localDB.AutoMigrate(&Event{})
	if err != nil {
		return nil, err
	}

	// Create indexes
	if !localDB.Migrator().HasIndex(&Event{}, "idx_events_date") {
		err = localDB.Exec("CREATE INDEX IF NOT EXISTS idx_events_date ON events(date)").Error
		if err != nil {
			return nil, err
		}
	}
	if !localDB.Migrator().HasIndex(&Event{}, "idx_events_tags") { // Uncommenting Tags index
		err = localDB.Exec("CREATE INDEX IF NOT EXISTS idx_events_tags ON events(tags)").Error
		if err != nil {
			return nil, err
		}
	}

	return localDB, nil // Return the initialized DB instance
}
