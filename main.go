package main

import (
	"crypto/subtle" // Added for secure API key comparison
	// Added for parsing JSON tags
	"errors" // Added for gorm.ErrRecordNotFound
	"io"     // Added for io.MultiWriter
	"log"
	"os"      // Added for sorting tags, os.Stdout, os.MkdirAll, os.OpenFile
	"strconv" // Added for pagination
	"strings" // Added for tag processing
	"time"    // Added for rate limiter

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter" // Added for rate limiting
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/gorm" // Added for gorm.ErrRecordNotFound
	// GORM is already used by database.go, no need for direct import here unless using DB functions directly
)

// Define global DB variables for English and Russian databases
var DB_EN *gorm.DB
var DB_RU *gorm.DB

// Helper function to get the correct DB instance based on language
func getDBInstance(langCode string) *gorm.DB {
	if strings.ToLower(langCode) == "ru" {
		return DB_RU
	}
	return DB_EN // Default to English
}

// Define a response structure for paginated events, matching your spec
type PaginatedEventsResponse struct {
	Events     []Event     `json:"events"`     // Changed from Data json:"data"
	Pagination interface{} `json:"pagination"` // Using interface{} for flexibility initially
}

type PaginationData struct {
	CurrentPage int   `json:"current_page"` // Was Page       int   `json:"page"`
	PerPage     int   `json:"per_page"`     // Was Limit      int   `json:"limit"`
	Total       int64 `json:"total"`        // GORM Count returns int64
	LastPage    int   `json:"last_page"`    // Was TotalPages int   `json:"total_pages"`
}

// var expectedAPIKey []byte // Old: single API key
var validAPIKeys [][]byte // New: slice to hold multiple valid API keys

// authMiddleware checks for a valid API key
func authMiddleware(c *fiber.Ctx) error {
	providedKey := c.Get("X-API-KEY")
	if providedKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "API key required"})
	}

	providedKeyBytes := []byte(providedKey)
	for _, expectedKey := range validAPIKeys {
		// Securely compare the provided key with each of the expected keys
		if subtle.ConstantTimeCompare(providedKeyBytes, expectedKey) == 1 {
			return c.Next()
		}
	}

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid API key"})
}

// New handler function for getting a single event
func getEventHandler(c *fiber.Ctx) error {
	lang := c.Query("lang", "en") // Default to 'en' if not specified
	db := getDBInstance(lang)
	id := c.Params("id")

	log.Printf("[INFO] getEventHandler: Called with ID '%s', Lang '%s'", id, lang)

	if id == "" {
		log.Printf("[WARN] getEventHandler: Event ID is required. Lang '%s'", lang)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Event ID is required",
		})
	}

	eventID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		log.Printf("[WARN] getEventHandler: Invalid Event ID format for ID '%s', Lang '%s'. Error: %v", id, lang, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid Event ID format",
		})
	}

	var event Event
	result := db.First(&event, uint(eventID))

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("[WARN] getEventHandler: Event not found for ID '%s' (parsed as %d), Lang '%s'. Error: %v", id, eventID, lang, result.Error)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Event not found",
			})
		}
		log.Printf("[ERROR] getEventHandler: Failed to retrieve event for ID '%s' (parsed as %d), Lang '%s'. Error: %v", id, eventID, lang, result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve event",
		})
	}
	log.Printf("[INFO] getEventHandler: Successfully retrieved event ID '%s' (parsed as %d), Lang '%s'", id, eventID, lang)
	return c.JSON(fiber.Map{"data": event})
}

// Structure for the /api/tags response
type TagInfo struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// Handler for /api/tags
func getTagsHandler(c *fiber.Ctx) error {
	lang := c.Query("lang", "en") // Default to 'en' if not specified
	db := getDBInstance(lang)

	log.Printf("[INFO] getTagsHandler: Called for Lang '%s'", lang)

	var result []TagInfo
	// SQL query to extract, count, and lowercase tags directly from JSON arrays in the 'tags' column.
	// This approach assumes tags are stored as valid JSON arrays (e.g., ["tag1", "tag2"]).
	// It replaces the previous Go-based parsing and aggregation logic.
	// Note: Fallback for comma-separated tags is removed with this SQL-native approach.
	// If tags are not valid JSON arrays, or if individual tags within the array are empty/whitespace-only,
	// they will be ignored by this query.
	sqlQuery := `
SELECT
    LOWER(j.value) AS tag,
    COUNT(*) AS count
FROM
    events e,
    json_each(e.tags) j
WHERE
    e.tags IS NOT NULL
    AND e.tags != ''        -- Not an empty string literal
    AND e.tags != '[]'      -- Not an empty JSON array string literal
    AND json_valid(e.tags) = 1 -- Ensures the string is valid JSON
    AND json_type(e.tags) = 'array' -- Ensures it's specifically a JSON array
    AND j.value IS NOT NULL
    AND TRIM(CAST(j.value AS TEXT)) != '' -- Ensures the extracted tag is not an empty or whitespace-only string
GROUP BY
    LOWER(j.value) -- Group by the lowercased tag for case-insensitive counting
ORDER BY
    tag ASC; -- Order alphabetically by the (now lowercased) tag
`
	if err := db.Raw(sqlQuery).Scan(&result).Error; err != nil {
		log.Printf("[ERROR] getTagsHandler: Error executing raw SQL for tags (Lang '%s'). Error: %v", lang, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve tags from database",
		})
	}

	// Sorting is now handled by the SQL query's "ORDER BY tag ASC".
	// The result slice is already in the correct []TagInfo format.
	log.Printf("[INFO] getTagsHandler: Successfully retrieved %d tags for Lang '%s'", len(result), lang)
	return c.JSON(fiber.Map{"data": result})
}

// Handler for /api/events/tags/{tag}
func getEventsByTagHandler(c *fiber.Ctx) error {
	lang := c.Query("lang", "en") // Default to 'en' if not specified
	db := getDBInstance(lang)
	tagParam := c.Params("tag")
	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "20")

	log.Printf("[INFO] getEventsByTagHandler: Called with Tag '%s', Lang '%s', Page '%s', Limit '%s'", tagParam, lang, pageStr, limitStr)

	if tagParam == "" {
		log.Printf("[WARN] getEventsByTagHandler: Tag parameter is required. Lang '%s'", lang)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tag parameter is required",
		})
	}

	// Pagination parameters
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		log.Printf("[WARN] getEventsByTagHandler: Invalid page parameter '%s'. Using default 1. Lang '%s', Tag '%s'. Error: %v", pageStr, lang, tagParam, err)
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		log.Printf("[WARN] getEventsByTagHandler: Invalid limit parameter '%s'. Using default 20. Lang '%s', Tag '%s'. Error: %v", limitStr, lang, tagParam, err)
		limit = 20
	}
	offset := (page - 1) * limit

	var events []Event
	var totalEvents int64

	// Prepare the search term for LIKE query, expecting tags like ["tag1","searchtag","tag2"]
	// This will search for the tag as a whole word within the JSON array string.
	searchTerm := "%\"" + strings.ToLower(tagParam) + "\"%"

	// Get total count of events matching the tag
	// We need to apply the Where condition for Count as well.
	countQuery := db.Model(&Event{}).Where("LOWER(tags) LIKE ?", searchTerm)
	if err := countQuery.Count(&totalEvents).Error; err != nil {
		log.Printf("[ERROR] getEventsByTagHandler: Failed to count events by tag '%s', Lang '%s'. Error: %v", tagParam, lang, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count events by tag",
		})
	}

	// Get paginated events matching the tag
	// Default sort by date descending
	dataQuery := db.Model(&Event{}).Order("date desc").Limit(limit).Offset(offset).Where("LOWER(tags) LIKE ?", searchTerm)
	if err := dataQuery.Find(&events).Error; err != nil {
		log.Printf("[ERROR] getEventsByTagHandler: Failed to retrieve events by tag '%s', Lang '%s', Page %d, Limit %d. Error: %v", tagParam, lang, page, limit, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve events by tag",
		})
	}

	totalPages := (totalEvents + int64(limit) - 1) / int64(limit)
	log.Printf("[INFO] getEventsByTagHandler: Successfully retrieved %d events for Tag '%s', Lang '%s', Page %d, Limit %d. Total matching: %d", len(events), tagParam, lang, page, limit, totalEvents)

	return c.JSON(PaginatedEventsResponse{
		Events: events,
		Pagination: PaginationData{
			CurrentPage: page,
			LastPage:    int(totalPages),
			PerPage:     limit,
			Total:       totalEvents,
		},
	})
}

func main() {
	// --- Log File Setup ---
	logDir := "./logs"
	if _, statErr := os.Stat(logDir); os.IsNotExist(statErr) {
		if mkdirErr := os.MkdirAll(logDir, 0755); mkdirErr != nil {
			log.Fatalf("Failed to create log directory %s: %v", logDir, mkdirErr)
		}
	}

	// Setup for general application logs (api.log)
	apiLogFile, openErr := os.OpenFile(logDir+"/api.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if openErr != nil {
		log.Fatalf("Failed to open log file %s/api.log: %v", logDir, openErr)
	}
	// defer apiLogFile.Close() // Closing handled by application lifecycle

	// MultiWriter for standard log package (writes to stdout and api.log)
	mw := io.MultiWriter(os.Stdout, apiLogFile)
	log.SetOutput(mw)
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds) // Optional: more detailed log flags

	log.Println("Starting API server...") // This will now go to stdout and api.log

	// Setup for Fiber request logs (requests.log)
	// It's often good to have request logs separate for clarity
	reqLogFile, reqOpenErr := os.OpenFile(logDir+"/requests.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if reqOpenErr != nil {
		log.Fatalf("Failed to open request log file %s/requests.log: %v", logDir, reqOpenErr) // Use standard log which is now configured
	}
	// defer reqLogFile.Close() // Closing handled by application lifecycle

	// --- API Key Setup ---
	// apiKeyFromEnv := os.Getenv("API_KEY") // Old: single API key
	// if apiKeyFromEnv == "" {
	// 	log.Fatal("API_KEY environment variable is not set. Authentication is required.")
	// }
	// expectedAPIKey = []byte(apiKeyFromEnv) // Old: single API key

	apiKeysStr := os.Getenv("API_KEYS")
	if apiKeysStr == "" {
		log.Fatal("API_KEYS environment variable is not set. Authentication is required.")
	}

	keys := strings.Split(apiKeysStr, ",")
	if len(keys) == 0 || (len(keys) == 1 && keys[0] == "") {
		log.Fatal("API_KEYS environment variable is empty or not properly formatted (comma-separated).")
	}

	for _, k := range keys {
		trimmedKey := strings.TrimSpace(k)
		if trimmedKey != "" {
			validAPIKeys = append(validAPIKeys, []byte(trimmedKey))
		}
	}

	if len(validAPIKeys) == 0 {
		log.Fatal("No valid API keys found after parsing API_KEYS. Please check the format.")
	}
	log.Printf("Loaded %d API key(s)", len(validAPIKeys))

	// --- Database Initialization for API ---
	dbPathEN := os.Getenv("DB_PATH_EN")
	if dbPathEN == "" {
		dbPathEN = "./data/events.db" // Default EN path
	}
	dbPathRU := os.Getenv("DB_PATH_RU")
	if dbPathRU == "" {
		dbPathRU = "./data/events_ru.db" // Default RU path
	}

	// Ensure data directory exists (checking based on one of the paths, e.g., English DB path)
	// This simple check assumes both DBs are in the same parent data directory.
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		if mkdirErr := os.MkdirAll("./data", 0755); mkdirErr != nil {
			log.Fatalf("Failed to create data directory: %v", mkdirErr)
		}
	}

	var err error
	DB_EN, err = InitDB(dbPathEN)
	if err != nil {
		log.Fatalf("Failed to initialize English database ('%s'): %v", dbPathEN, err)
	}
	log.Printf("English database ('%s') initialized successfully for API.", dbPathEN)

	DB_RU, err = InitDB(dbPathRU)
	if err != nil {
		log.Fatalf("Failed to initialize Russian database ('%s'): %v", dbPathRU, err)
	}
	log.Printf("Russian database ('%s') initialized successfully for API.", dbPathRU)

	// --- Fiber App Initialization ---
	app := fiber.New()

	// Middleware
	// Configure Fiber logger to write to requests.log and os.Stdout
	// To write to os.Stdout as well for Fiber logger, we can pass a MultiWriter to it.
	// However, Fiber's logger.New() takes a single io.Writer for its Output.
	// If we want Fiber logs in BOTH file and console, we'd need a custom setup or pass our main `mw`.
	// For simplicity, let's make Fiber logger write to its own file AND stdout.
	fiberLoggerMw := io.MultiWriter(os.Stdout, reqLogFile)
	app.Use(logger.New(logger.Config{
		Output: fiberLoggerMw,
		// Optional: Customize format, time zone, etc.
		// Format: "${time} ${status} - ${latency} ${method} ${path}\\n",
	}))

	// Apply Rate Limiter to the /api group
	// Example: 100 requests per 1 minute per IP
	api := app.Group("/api")
	api.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Use IP address as the key for rate limiting
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded, please try again later.",
			})
		},
	}))

	// Apply Auth Middleware to the /api group
	api.Use(authMiddleware)

	// --- Routes ---
	api.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Bitcoin Events API!")
	})

	// Get All Events
	api.Get("/events", func(c *fiber.Ctx) error {
		lang := c.Query("lang", "en") // Default to 'en' if not specified
		db := getDBInstance(lang)
		pageStr := c.Query("page", "1")
		limitStr := c.Query("limit", "20")
		yearStr := c.Query("year")
		monthStr := c.Query("month")
		dayStr := c.Query("day")

		log.Printf("[INFO] /api/events: Called. Lang '%s', Page '%s', Limit '%s', Year '%s', Month '%s', Day '%s'", lang, pageStr, limitStr, yearStr, monthStr, dayStr)

		// Pagination parameters
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			log.Printf("[WARN] /api/events: Invalid page parameter '%s'. Using default 1. Lang '%s'. Error: %v", pageStr, lang, err)
			page = 1
		}
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			log.Printf("[WARN] /api/events: Invalid limit parameter '%s'. Using default 20. Lang '%s'. Error: %v", limitStr, lang, err)
			limit = 20
		}
		offset := (page - 1) * limit

		var events []Event
		var totalEvents int64

		query := db.Model(&Event{})
		conditions := []string{}
		params := []interface{}{}

		if yearStr != "" {
			_, err := strconv.Atoi(yearStr) // Validate if it's a number, actual length/range validation can be more specific if needed
			if err == nil {                 // Basic validation, ensuring it's a number. Assumes 4-digit year is typical but doesn't strictly enforce.
				log.Printf("[INFO] /api/events: Applying year filter. Lang '%s', Year '%s'", lang, yearStr)
				conditions = append(conditions, "strftime('%Y', date) = ?")
				params = append(params, yearStr)
			} else {
				log.Printf("[WARN] /api/events: Invalid year parameter format '%s' (Lang '%s'). Ignoring year filter. Error: %v", yearStr, lang, err)
			}
		}

		if monthStr != "" {
			month, err := strconv.Atoi(monthStr)
			if err == nil && month >= 1 && month <= 12 {
				log.Printf("[INFO] /api/events: Applying month filter. Lang '%s', Month '%s'", lang, monthStr)
				formattedMonth := monthStr
				if month < 10 && !strings.HasPrefix(monthStr, "0") {
					formattedMonth = "0" + monthStr
				}
				conditions = append(conditions, "strftime('%m', date) = ?")
				params = append(params, formattedMonth)
			} else {
				log.Printf("[WARN] /api/events: Invalid month parameter '%s' (Lang '%s'). Ignoring month filter. Error: %v", monthStr, lang, err)
			}
		}

		if dayStr != "" {
			day, err := strconv.Atoi(dayStr)
			if err == nil && day >= 1 && day <= 31 {
				log.Printf("[INFO] /api/events: Applying day filter. Lang '%s', Day '%s'", lang, dayStr)
				formattedDay := dayStr
				if day < 10 && !strings.HasPrefix(dayStr, "0") {
					formattedDay = "0" + dayStr
				}
				conditions = append(conditions, "strftime('%d', date) = ?")
				params = append(params, formattedDay)
			} else {
				log.Printf("[WARN] /api/events: Invalid day parameter '%s' (Lang '%s'). Ignoring day filter. Error: %v", dayStr, lang, err)
			}
		}

		if len(conditions) > 0 {
			query = query.Where(strings.Join(conditions, " AND "), params...)
		}

		// Get total count of events (with potential date filter)
		if err := query.Count(&totalEvents).Error; err != nil {
			log.Printf("[ERROR] /api/events: Failed to count events. Lang '%s', Year '%s', Month '%s', Day '%s'. Error: %v", lang, yearStr, monthStr, dayStr, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to count events",
			})
		}

		// Get paginated events (with potential date filter)
		// Default sort by date descending
		if err := query.Order("date desc").Limit(limit).Offset(offset).Find(&events).Error; err != nil {
			log.Printf("[ERROR] /api/events: Failed to retrieve events. Lang '%s', Page %d, Limit %d, Year '%s', Month '%s', Day '%s'. Error: %v", lang, page, limit, yearStr, monthStr, dayStr, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve events",
			})
		}

		totalPages := 0
		if limit > 0 { // Avoid division by zero if limit is 0 for some reason
			totalPages = int((totalEvents + int64(limit) - 1) / int64(limit))
		}
		log.Printf("[INFO] /api/events: Successfully retrieved %d events. Lang '%s', Page %d, Limit %d, Year '%s', Month '%s', Day '%s'. Total matching: %d", len(events), lang, page, limit, yearStr, monthStr, dayStr, totalEvents)

		return c.JSON(PaginatedEventsResponse{
			Events: events,
			Pagination: PaginationData{
				CurrentPage: page,
				LastPage:    totalPages,
				PerPage:     limit,
				Total:       totalEvents,
			},
		})
	})

	// Get Single Event by ID
	api.Get("/events/:id", getEventHandler) // Registering the new handler

	// Add the new route for getting all tags
	api.Get("/tags", getTagsHandler)

	// Add the new route for getting events by tag
	api.Get("/events/tags/:tag", getEventsByTagHandler)

	// --- Start Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // Default port
	}
	log.Printf("Starting server on port %s", port) // This will also go to stdout and api.log
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err) // This will also go to stdout and api.log
	}
}
