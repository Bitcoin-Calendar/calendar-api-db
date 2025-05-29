[![⚡️zapmeacoffee](https://img.shields.io/badge/⚡️zap_-me_a_coffee-violet?style=plastic)](https://zapmeacoffee.com/npub1tcalvjvswjh5rwhr3gywmfjzghthexjpddzvlxre9wxfqz4euqys0309hn)

# Bitcoin Historical Events API Documentation

This document provides details for interacting with the Bitcoin Historical Events API.

## Base URL

The API is served from the root of the application. If running on a server with IP `213.176.74.147` and port `3001`, the base URL for all API endpoints will be:

`http://213.176.74.147:3001/api`

(Replace with `http://localhost:3001/api` if running locally).

## Authentication

The API requires an API key to be passed in the `X-API-KEY` header for all endpoints under `/api`. The server can be configured with one or more comma-separated keys via the `API_KEYS` environment variable.

## Rate Limiting

Rate limiting is applied per IP address. The current limit is 100 requests per minute.

## Overview of Endpoints

The API provides the following main functionalities:

*   **`/`**: Health check for the API.
*   **`/events`**: Retrieve a paginated list of all events, with powerful filtering by date (year, month, day, or combinations) and language.
*   **`/events/:id`**: Fetch a single event by its unique ID.
*   **`/tags`**: Get a list of all unique event tags and their usage counts.
*   **`/events/tags/:tag`**: Retrieve a paginated list of events associated with a specific tag.

Detailed information for each endpoint is provided below.

## Available Tags for Querying

The following tags can be used with the `/events/tags/:tag` endpoint to find relevant events. Note that tag searching is case-insensitive.

*   `clownworld`: Content related to traditional finance, banking sector announcements and activities, mainstream media narratives, and so on.
*   `cypherpunks`: Events, discussions, quotations, and news pertaining to the cypherpunk movement, including key figures like Satoshi Nakamoto, Hal Finney, David Chaum, and other early adopters, along with significant related activities.
*   `quotes`: Direct quotations featured within the event descriptions.
*   `bitcointalk`: Events or discussions from the BitcoinTalk forum.
*   `ecash`: Content pertaining to DigiCash, e-cash, or similar digital cash systems.
*   `lightning`: Events, developments, or discussions related to the Lightning Network or Lightning payments.
*   `onchain`: Events and discussions concerning on-chain Bitcoin transactions.
*   `obituaries`: Bitcoin obituaries events.
*   `trading`: Content related to Bitcoin trading, market analysis, or price charts.
*   `first`: Milestone events marking a 'first' occurrence in the Bitcoin ecosystem.
*   `scam`: Incidents involving scams, financial losses, or theft within the Bitcoin space.
*   `hack`: Incidents involving hacks or theft within the Bitcoin space.
*   `mustread`: Important documents, articles and books deemed essential reading.
*   `econ`: Discussions and events related to economic theories, principles, or impacts concerning Bitcoin.
*   `privacy`: Content focusing on privacy aspects, technologies, or discussions within the Bitcoin context.
*   `development`: Events related to Bitcoin Core releases, protocol upgrades, and technical proposals.
*   `adoption`: Events highlighting instances of countries, governments, businesses, individuals, or organizations starting to accept or use Bitcoin.
*   `legal`: For events involving lawsuits, regulations, and government legal actions.
*   `media`: To specifically categorize mentions of Bitcoin in articles, TV shows, and other media.

## Language Support

Most data-retrieving endpoints support a `lang` query parameter to specify the language of the events to be queried. This feature is confirmed to be working.

*   `lang=en` (Default): Retrieves events from the English database (`events.db`).
*   `lang=ru`: Retrieves events from the Russian database (`events_ru.db`).

If the `lang` parameter is omitted or an unsupported value is provided, it defaults to `en`.

## Error Responses

Standard HTTP status codes are used. Common error responses include:

*   `400 Bad Request`: The request was malformed (e.g., missing required parameters, invalid parameter format).
*   `401 Unauthorized`: The API key is missing or invalid.
*   `404 Not Found`: The requested resource (e.g., a specific event) could not be found.
*   `429 Too Many Requests`: Rate limit exceeded.
*   `500 Internal ServerError`: An unexpected error occurred on the server.

Error responses will typically be in JSON format, like:

```json
{
  "error": "Descriptive error message"
}
```

## Endpoints

### 1. Health Check

*   **Endpoint:** `/`
*   **Method:** `GET`
*   **Description:** A simple health check endpoint. (Note: This endpoint is authenticated)
*   **Query Parameters:** None
*   **Request Body:** None
*   **Success Response (200 OK):**
    *   **Content-Type:** `text/plain`
    *   **Body:** `Hello, Bitcoin Events API!`
*   **Example:**
    ```bash
    curl -H "X-API-KEY: your_api_key" http://213.176.74.147:3001/api/
    ```

### 2. Get All Events (Paginated)

*   **Endpoint:** `/events`
*   **Method:** `GET`
*   **Description:** Retrieves a paginated list of all historical Bitcoin events, sorted by date in descending order by default. Supports language selection and flexible date filtering by year, month, and/or day. Filters can be combined (e.g., year and month, or year, month, and day).
*   **Query Parameters:**
    *   `page` (optional, integer): The page number to retrieve. Defaults to `1`.
    *   `limit` (optional, integer): The number of events per page. Defaults to `20`.
    *   `lang` (optional, string): Language for the events. `en` for English (default), `ru` for Russian.
    *   `year` (optional, string, format: `YYYY` e.g., "2022"): Year for filtering events.
    *   `month` (optional, string, format: `MM` or `M` e.g., "05" or "5"): Month for filtering events.
    *   `day` (optional, string, format: `DD` or `D` e.g., "27" or "7"): Day for filtering events.
*   **Request Body:** None
*   **Success Response (200 OK):**
    *   **Content-Type:** `application/json`
    *   **Body:**
        ```json
        {
          "events": [ // Note: Changed from "data" to "events"
            {
              "id": 1,
              "date": "2008-11-01T00:00:00Z",
              "title": "📜 Bitcoin Whitepaper Published",
              "description": "Satoshi Nakamoto publishes the Bitcoin whitepaper...",
              "tags": "["bitcoin","whitepaper","satoshi"]",
              "media": "https://bitcoin.org/img/icons/opengraph.png",
              "references": "["https://bitcoin.org/bitcoin.pdf"]",
              "created_at": "2025-05-26T17:00:00Z", // Example timestamp
              "updated_at": "2025-05-26T17:00:00Z"  // Example timestamp
            }
            // ... more events
          ],
          "pagination": {
            "current_page": 1, // Note: field names changed
            "per_page": 20,    // Note: field names changed
            "total": 230,
            "last_page": 12    // Note: field names changed
          }
        }
        ```
*   **Example:**
    ```bash
    # Get first 5 English events
    curl -H "X-API-KEY: your_api_key" "http://213.176.74.147:3001/api/events?page=1&limit=5&lang=en"

    # Get Russian events for December 2023
    curl -H "X-API-KEY: your_api_key" "http://213.176.74.147:3001/api/events?year=2023&month=12&lang=ru"

    # Get English events for the 15th of any month/year
    curl -H "X-API-KEY: your_api_key" "http://213.176.74.147:3001/api/events?day=15&lang=en"

    # Get English events for any day in May of any year
    curl -H "X-API-KEY: your_api_key" "http://213.176.74.147:3001/api/events?month=05&lang=en"
    
    # Get English events for the year 2021
    curl -H "X-API-KEY: your_api_key" "http://213.176.74.147:3001/api/events?year=2021&lang=en"

    # Get Russian events for May 27th, 2020
    curl -H "X-API-KEY: your_api_key" "http://213.176.74.147:3001/api/events?year=2020&month=05&day=27&lang=ru"
    ```

### 3. Get Single Event by ID

*   **Endpoint:** `/events/:id`
*   **Method:** `GET`
*   **Description:** Retrieves a single historical Bitcoin event by its unique ID. Supports language selection.
*   **Path Parameters:**
    *   `id` (required, integer): The unique identifier of the event.
*   **Query Parameters:**
    *   `lang` (optional, string): Language for the event. `en` for English (default), `ru` for Russian.
*   **Request Body:** None
*   **Success Response (200 OK):**
    *   **Content-Type:** `application/json`
    *   **Body:**
        ```json
        {
          "data": { // Note: This endpoint's response structure was not specified as changed in the original issue, keeping "data" wrapper for now.
            "id": 1,
            "date": "2008-11-01T00:00:00Z",
            "title": "📜 Bitcoin Whitepaper Published",
            "description": "Satoshi Nakamoto publishes the Bitcoin whitepaper...",
            "tags": "["bitcoin","whitepaper","satoshi"]",
            "media": "https://bitcoin.org/img/icons/opengraph.png",
            "references": "["https://bitcoin.org/bitcoin.pdf"]",
            "created_at": "2025-05-26T17:00:00Z",
            "updated_at": "2025-05-26T17:00:00Z"
          }
        }
        ```
*   **Error Responses:**
    *   `400 Bad Request`: If `id` is not a valid integer.
        ```json
        { "error": "Invalid Event ID format" }
        ```
    *   `404 Not Found`: If an event with the given `id` does not exist.
        ```json
        { "error": "Event not found" }
        ```
*   **Example:**
    ```bash
    # Get English event with ID 1
    curl -H "X-API-KEY: your_api_key" "http://213.176.74.147:3001/api/events/1?lang=en"

    # Get Russian event with ID 20 (ID for the May 27th event in Russian DB)
    curl -H "X-API-KEY: your_api_key" "http://213.176.74.147:3001/api/events/20?lang=ru"
    ```

### 4. Get All Unique Tags

*   **Endpoint:** `/tags`
*   **Method:** `GET`
*   **Description:** Retrieves a list of all unique tags found across all events, along with the count of events associated with each tag. Tags are returned in alphabetical order. Supports language selection.
*   **Query Parameters:**
    *   `lang` (optional, string): Language for the tags. `en` for English (default), `ru` for Russian.
*   **Request Body:** None
*   **Success Response (200 OK):**
    *   **Content-Type:** `application/json`
    *   **Body:**
        ```json
        {
          "data": [ // Note: This endpoint's response structure was not specified as changed, keeping "data" wrapper for now.
            {
              "tag": "adoption",
              "count": 72
            },
            {
              "tag": "bitcoin",
              "count": 1
            }
            // ... more tags
          ]
        }
        ```
*   **Example:**
    ```bash
    # Get English tags
    curl -H "X-API-KEY: your_api_key" "http://213.176.74.147:3001/api/tags?lang=en"

    # Get Russian tags
    curl -H "X-API-KEY: your_api_key" "http://213.176.74.147:3001/api/tags?lang=ru"
    ```

### 5. Get Events by Tag (Paginated)

*   **Endpoint:** `/events/tags/:tag`
*   **Method:** `GET`
*   **Description:** Retrieves a paginated list of historical Bitcoin events associated with a specific tag. Events are sorted by date in descending order by default. The tag search is case-insensitive. Supports language selection.
*   **Path Parameters:**
    *   `tag` (required, string): The tag to filter events by.
*   **Query Parameters:**
    *   `page` (optional, integer): The page number to retrieve. Defaults to `1`.
    *   `limit` (optional, integer): The number of events per page. Defaults to `20`.
    *   `lang` (optional, string): Language for the events. `en` for English (default), `ru` for Russian.
*   **Request Body:** None
*   **Success Response (200 OK):**
    *   **Content-Type:** `application/json`
    *   **Body (follows the same structure as Get All Events `/events`):**
        ```json
        {
          "events": [ // Note: Changed from "data" to "events"
            {
              "id": 10,
              "date": "2010-05-22T00:00:00Z",
              "title": "🍕 Bitcoin Pizza Day",
              "description": "Laszlo Hanyecz made the first purchase...",
              "tags": "["first","adoption","bitcointalk"]",
              "media": "https://example.com/pizza.webp",
              "references": "["https://bitcointalk.org/..."]",
              "created_at": "2025-05-26T17:00:00Z",
              "updated_at": "2025-05-26T17:00:00Z"
            }
            // ... more events with the specified tag
          ],
          "pagination": {
            "current_page": 1,
            "per_page": 20,
            "total": 5, // Total events matching the tag
            "last_page": 1
          }
        }
        ```
*   **Error Responses:**
    *   `400 Bad Request`: If `tag` parameter is missing.
        ```json
        { "error": "Tag parameter is required" }
        ```
*   **Example:**
    ```bash
    # Get first 2 English events tagged with 'adoption'
    curl -H "X-API-KEY: your_api_key" "http://213.176.74.147:3001/api/events/tags/adoption?limit=2&lang=en"

    # Get first 2 Russian events tagged with 'adoption'
    curl -H "X-API-KEY: your_api_key" "http://213.176.74.147:3001/api/events/tags/adoption?limit=2&lang=ru"
    ```

[![⚡️zapmeacoffee](https://img.shields.io/badge/⚡️zap_-me_a_coffee-violet?style=plastic)](https://zapmeacoffee.com/npub1tcalvjvswjh5rwhr3gywmfjzghthexjpddzvlxre9wxfqz4euqys0309hn) 