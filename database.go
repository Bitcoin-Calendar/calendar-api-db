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
	Media       string    `json:"media" gorm:"type:text"`      // Link to media file(s), stored as a JSON array string e.g., ["url1", "url2"]
	References  string    `json:"references" gorm:"type:text"` // JSON array as string
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Rank        float64   `json:"-" gorm:"-"` // Omit from JSON and DB schema
}

// InitDB initializes the database connection and migrates the schema.
// It now returns the DB instance or an error.
func InitDB(dbPath string) (*gorm.DB, error) {
	var err error
	var localDB *gorm.DB // Use a local variable for the DB instance
	localDB, err = gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000"), &gorm.Config{
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

	// Create FTS5 virtual table
	if err := localDB.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS events_fts USING fts5(
			title,
			description,
			tags,
			content='events',
			content_rowid='id'
		);
	`).Error; err != nil {
		return nil, err
	}

	// Triggers to keep FTS table synchronized with events table
	if err := localDB.Exec(`
		CREATE TRIGGER IF NOT EXISTS events_after_insert
		AFTER INSERT ON events
		BEGIN
			INSERT INTO events_fts(rowid, title, description, tags)
			VALUES (new.id, new.title, new.description, new.tags);
		END;
	`).Error; err != nil {
		return nil, err
	}

	if err := localDB.Exec(`
		CREATE TRIGGER IF NOT EXISTS events_after_delete
		AFTER DELETE ON events
		BEGIN
			INSERT INTO events_fts(events_fts, rowid, title, description, tags)
			VALUES ('delete', old.id, old.title, old.description, old.tags);
		END;
	`).Error; err != nil {
		return nil, err
	}

	if err := localDB.Exec(`
		CREATE TRIGGER IF NOT EXISTS events_after_update
		AFTER UPDATE ON events
		BEGIN
			INSERT INTO events_fts(events_fts, rowid, title, description, tags)
			VALUES ('delete', old.id, old.title, old.description, old.tags);
			INSERT INTO events_fts(rowid, title, description, tags)
			VALUES (new.id, new.title, new.description, new.tags);
		END;
	`).Error; err != nil {
		return nil, err
	}

	// Initial population of FTS table
	var count int64
	localDB.Model(&Event{}).Count(&count)
	var ftsCount int64
	localDB.Table("events_fts").Count(&ftsCount)

	if count > 0 && ftsCount == 0 {
		if err := localDB.Exec(`
			INSERT INTO events_fts(rowid, title, description, tags)
			SELECT id, title, description, tags FROM events;
		`).Error; err != nil {
			return nil, err
		}
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
