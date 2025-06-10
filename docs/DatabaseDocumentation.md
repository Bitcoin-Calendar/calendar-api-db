[![⚡️zapmeacoffee](https://img.shields.io/badge/⚡️zap_-me_a_coffee-violet?style=plastic)](https://zapmeacoffee.com/npub1tcalvjvswjh5rwhr3gywmfjzghthexjpddzvlxre9wxfqz4euqys0309hn)

# Bitcoin Historical Events SQLite Database Documentation

This document provides details about the SQLite database schema used by the Bitcoin Historical Events API.

## Database Files

The API utilizes two separate SQLite database files to store event data for different languages:

*   **English Events:**
    *   **Location:** `calendar-api-db/data/events.db` (default when `DB_PATH_EN` is not set)
    *   **Environment Variable for Path:** `DB_PATH_EN`
*   **Russian Events:**
    *   **Location:** `calendar-api-db/data/events_ru.db` (default when `DB_PATH_RU` is not set)
    *   **Environment Variable for Path:** `DB_PATH_RU`

Both databases are of type SQLite 3 and share the same table schema described below.

## Table: `events`

This is the primary table storing all historical event data in *both* database files (`events.db` and `events_ru.db`).

### Columns

| Column Name   | Data Type         | Constraints                | Description                                                                 |
|---------------|-------------------|----------------------------|-----------------------------------------------------------------------------|
| `id`          | `INTEGER`         | `PRIMARY KEY AUTOINCREMENT`| Unique identifier for the event.                                            |
| `date`        | `DATE`            | `NOT NULL`                 | The date of the event (YYYY-MM-DD format).                                  |
| `title`       | `VARCHAR(255)`    | `NOT NULL`                 | The title or headline of the event.                                         |
| `description` | `TEXT`            |                            | A detailed description of the event.                                        |
| `tags`        | `VARCHAR(500)`    |                            | A JSON array of strings representing tags associated with the event. E.g., `["bitcoin", "whitepaper"]`. |
| `media`       | `TEXT`            |                            | A URL pointing to a relevant media file (image, video, etc.).             |
| `references`  | `TEXT`            |                            | A JSON array of strings representing URLs for source/reference links. E.g., `["http://example.com/source1", "http://example.com/source2"]`. |
| `created_at`  | `DATETIME`        |                            | Timestamp of when the record was created in the database.                   |
| `updated_at`  | `DATETIME`        |                            | Timestamp of when the record was last updated in the database.                |

### Indexes

*   **`idx_events_date`**: Index on the `date` column to speed up date-based queries.
*   **`idx_events_tags`**: Index on the `tags` column to speed up tag-based searches.
*   An implicit index is created on `id` as it is the `PRIMARY KEY`.

## Full-Text Search Table: `events_fts`

To enable efficient full-text searching, the database utilizes an FTS5 virtual table named `events_fts`.

*   **Purpose:** This table provides fast, indexed searching on the text content of the `events` table. It is used by the `/api/search` endpoint.
*   **Indexed Columns:** The `events_fts` table indexes the content from the following columns of the `events` table:
    *   `title`
    *   `description`
    *   `tags`
*   **Synchronization:** The `events_fts` table is kept automatically synchronized with the main `events` table using database triggers. Any `INSERT`, `UPDATE`, or `DELETE` operation on `events` is automatically reflected in `events_fts`. This means no manual intervention is required to keep the search index up-to-date.
*   **Creation:** The table and its triggers are created automatically by the `InitDB` function in `database.go` when the API server starts.

## Database Initialization and Migration (Schema)

The database schema is managed by GORM (the Go ORM used in this project) via the `AutoMigrate` feature.

The `InitDB` function in `calendar-api-db/database.go` handles:
1.  Connecting to a specified SQLite database file (path provided as an argument).
2.  Automatically migrating the `Event` struct to the `events` table, creating or updating columns as necessary.
3.  Ensuring the specified indexes (`idx_events_date`, `idx_events_tags`) exist.

In `calendar-api-db/main.go`, `InitDB` is called twice during API server startup to initialize connections to both the English and Russian database files, using paths specified by `DB_PATH_EN` and `DB_PATH_RU` environment variables (or their defaults if the variables are not set).

## Data Population (Manual Migration from CSV)

Data is populated into the `events` table using the `calendar-api-db/migrate.go` script. This script is **intended for manual execution** if you need to (re)populate the database from CSV files.

**To run the migration script:**

1.  **Ensure Go is installed** on the machine where you intend to run the script.
2.  **Navigate to the `calendar-api-db` directory** in your terminal:
    ```bash
    cd /path/to/your/project/calendar-api-db
    ```
3.  **Prepare your CSV file(s).** The script expects CSV files with columns: `date,title,description,tags,media,references` (though `tags`, `media`, and `references` are optional).
4.  **Modify `migrate.go` (Temporary Change):**
    *   Open `calendar-api-db/migrate.go` in a text editor.
    *   Find the function definition: `func runMigration(csvFilePath string, dbPath string) { ... }`
    *   Temporarily rename this function to `main`: `func main() { ... }`
    *   **Important**: The `migrate.go` script, when its main function is named `main`, uses its own command-line flag parsing for `-csv` and `-db` paths. These are distinct from the (now removed) flags that were in `main.go`.
5.  **Run the script using `go run`:**
    Execute the script, providing the paths to your input CSV and the target database file. You will need to define the `csvPathPtr` and `dbPathPtr` inside the now `main` function in `migrate.go` or modify it to accept them as arguments if you prefer a different execution method.
    A more direct way, assuming `migrate.go` is set up to parse flags when `runMigration` is `main`:
    ```bash
    # Example for English events:
go run migrate.go -csv=./data/events.csv -db=./data/events.db

    # Example for Russian events:
go run migrate.go -csv=./data/events_ru.csv -db=./data/events_ru.db
    ```
    *(Adjust file paths as necessary.)*

6.  **Revert Changes to `migrate.go`:**
    *   After the script finishes, open `calendar-api-db/migrate.go` again.
    *   Rename the `main` function back to `runMigration`.
    *   This is crucial because the `main` package for the API server (`main.go`) expects `runMigration` to be a regular function, not the entry point of the package.

The script will:
1.  Read event data from the specified CSV file.
2.  Parse each row.
3.  Transform `tags` and `references` from comma-separated strings into JSON array strings.
4.  Insert each event into the `events` table of the specified database file.

## Notes on `tags` and `references` storage

*   **JSON Array as String:** Storing tags and references as JSON array strings is a denormalized approach. It simplifies the current application structure.
*   **Searching:** Searching for specific tags or references currently involves using `LIKE` queries on these JSON strings (e.g., `WHERE tags LIKE '%"mytag"%'`).
*   **Alternatives for Larger Scale:** For more complex querying needs or larger datasets, consider normalizing these into separate tables (e.g., `event_tags`, `tag_list`, `event_references`) with many-to-many relationships.

[![⚡️zapmeacoffee](https://img.shields.io/badge/⚡️zap_-me_a_coffee-violet?style=plastic)](https://zapmeacoffee.com/npub1tcalvjvswjh5rwhr3gywmfjzghthexjpddzvlxre9wxfqz4euqys0309hn) 