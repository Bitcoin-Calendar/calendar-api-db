[![⚡️zapmeacoffee](https://img.shields.io/badge/⚡️zap_-me_a_coffee-violet?style=plastic)](https://zapmeacoffee.com/npub1tcalvjvswjh5rwhr3gywmfjzghthexjpddzvlxre9wxfqz4euqys0309hn)

# Docker Deployment Guide for Bitcoin Historical Events API

This guide provides instructions on how to build, run, and manage the Bitcoin Historical Events API using Docker and Docker Compose.

## Prerequisites

- Docker Engine: [Install Docker](https://docs.docker.com/engine/install/)
- Docker Compose: [Install Docker Compose](https://docs.docker.com/compose/install/)

## Project Structure (`calendar-api-db/` directory)

-   `main.go`: The main Go application file for the API server. Handles API requests, database connections, and language selection.
-   `database.go`: Go source file defining the `Event` struct and the `InitDB` function for database initialization and schema migration.
-   `go.mod`, `go.sum`: Go module files for dependency management.
-   `Dockerfile`: Defines the multi-stage Docker build for creating a lean production image of the API server.
-   `docker-compose.yml`: Configures the API service for Docker Compose, including build context, port mapping, volume mounts for data persistence, and environment variable settings.
-   `data/`:
    -   `events.db`: SQLite database for English events (default).
    -   `events_ru.db`: SQLite database for Russian events (default).
-   `docs/`:
    -   `APIDocumentation.md`: Detailed information about API endpoints, authentication, and usage.
    -   `DatabaseDocumentation.md`: Schema details for the SQLite databases and instructions for manual data population.
    -   `Deployment.md`: This file.
-   `README.md`: Overview of the project, features, and pointers to documentation.

## Building and Running the API

1.  **Navigate to the Project Directory:**
    Open a terminal and change to the `calendar-api-db` directory (where `docker-compose.yml` is located):
    ```bash
    cd /path/to/your/project/calendar-api-db
    ```

2.  **Set Environment Variables:**
    The `docker-compose.yml` file is configured to use environment variables. Ensure you have `API_KEYS` (comma-separated if multiple) set in your environment or in a `.env` file in the `calendar-api-db` directory. Example `.env` file content:
    ```
    API_KEYS=your_secret_api_key1,another_key2
    DB_PATH_EN=./data/events.db
    DB_PATH_RU=./data/events_ru.db
    PORT=3001
    ```
    The `DB_PATH_EN`, `DB_PATH_RU`, and `PORT` variables in the `.env` file will override the defaults in `main.go` if set.

3.  **Build and Run with Docker Compose:**
    Use the following command to build the Docker image (if changed) and start the API service in detached mode:
    ```bash
    docker-compose up --build -d
    ```
    - `--build`: Forces Docker Compose to rebuild the image if the `Dockerfile` or application source code has changed.
    - `-d`: Runs the containers in detached mode.

4.  **Verify the Container is Running:**
    ```bash
    docker-compose ps
    # OR, to see logs:
    docker-compose logs -f api
    ```
    You should see the `calendar-api-db_api_1` (or similar, depending on your directory name) container running and the port (default `3000`) mapped.

5.  **Accessing the API:**
    The API will be accessible at `http://<your_host_ip>:<PORT>/api` (e.g., `http://localhost:3001/api` or `http://213.176.74.147:3001/api`). Refer to `docs/APIDocumentation.md` for endpoint details.

## Stopping the API

To stop the API service and remove the containers:
```bash
docker-compose down
```
To only stop the service:
```bash
docker-compose stop api
```

## Data Persistence

The SQLite database files (`events.db`, `events_ru.db`) are persisted on the host machine via volume mounts defined in `docker-compose.yml`:

```yaml
volumes:
  - ./data:/app/data
```
This maps the host's `./data` directory (relative to `docker-compose.yml`) to `/app/data` inside the container. The Go application uses these paths (configurable via `DB_PATH_EN` and `DB_PATH_RU` environment variables) to access the databases.

## Environment Variables for the API Server

The API server (`main.go`) can be configured using the following environment variables. These can be set in your shell, a `.env` file in the `calendar-api-db` directory (which `docker-compose` automatically loads), or directly in the `docker-compose.yml`.

-   `API_KEYS`: (Required) Comma-separated list of secret keys for API authentication (e.g., `key1,key2`).
-   `DB_PATH_EN`: Path to the English SQLite database, relative to the app's working directory inside the container (`/app`). Defaults to `./data/events.db` (effectively `/app/data/events.db`).
-   `DB_PATH_RU`: Path to the Russian SQLite database, relative to the app's working directory inside the container (`/app`). Defaults to `./data/events_ru.db` (effectively `/app/data/events_ru.db`).
-   `PORT`: Port for the API server. Defaults to `3000`.

**Example `docker-compose.yml` section for environment variables:**
```yaml
services:
  api:
    # ... other configs
    build: .
    ports:
      - "${PORT:-3000}:3000" # Uses PORT from .env or defaults to 3000 for host
    volumes:
      - ./data:/app/data
    environment:
      - API_KEYS=${API_KEYS} # Must be set in .env or shell
      - DB_PATH_EN=${DB_PATH_EN:-./data/events.db} # Uses DB_PATH_EN from .env or defaults
      - DB_PATH_RU=${DB_PATH_RU:-./data/events_ru.db} # Uses DB_PATH_RU from .env or defaults
      - PORT=${PORT:-3000} # Sets PORT inside container, default used by app is 3000
    # For the API server to listen on the PORT var inside the container, 
    # main.go would need to be modified to use os.Getenv("PORT") for app.Listen.
    # The current main.go hardcodes "3000" or uses PORT env var if set, which is fine.
```
*(Note: The `docker-compose.yml` shown here is an illustrative example of how env vars can be handled; your actual file might differ slightly but should use these variables.)*

## Troubleshooting

- **Check Container Logs:** `docker-compose logs -f api`
- **Port Conflicts:** Ensure the host port (e.g., `3000`) is not in use.
- **File Permissions:** For the `./data` directory on the host, ensure the user running Docker has permissions.
- **API Key:** Double-check the `API_KEYS` is set and correctly passed in the `X-API-KEY` header. 

[![⚡️zapmeacoffee](https://img.shields.io/badge/⚡️zap_-me_a_coffee-violet?style=plastic)](https://zapmeacoffee.com/npub1tcalvjvswjh5rwhr3gywmfjzghthexjpddzvlxre9wxfqz4euqys0309hn)