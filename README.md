# URL Shortener

This is a simple URL shortener service written in Go. It provides an API to create short URLs and redirect to the original URL.

## Features

-   Create short URLs with custom aliases.
-   Generate random short aliases if no custom alias is provided.
-   Redirect to the original URL using the short alias.
-   Delete short URLs.
-   Rate limiting to prevent abuse.
-   Caching with Redis to improve performance.

## Technologies Used

-   **Language:** Go
-   **API Framework:** [chi](https://github.com/go-chi/chi)
-   **Database:** PostgreSQL
-   **Cache:** Redis
-   **Logging:** [zap](https://github.com/uber-go/zap)
-   **Containerization:** Docker, Docker Compose

## Getting Started

### Prerequisites

-   [Docker](https://www.docker.com/get-started)
-   [Docker Compose](https://docs.docker.com/compose/install/)

### Installation

1.  Clone the repository:
    ```sh
    git clone https://github.com/Pshimaf-Git/url-shortener.git
    ```
2.  Navigate to the project directory:
    ```sh
    cd url-shortener
    ```
3.  Start the services using Docker Compose:
    ```sh
    docker-compose up --build
    ```

The application will be running at `http://localhost:8080`.

## API Endpoints

| Method | Endpoint        | Description                  |
| ------ | --------------- | ---------------------------- |
| `POST` | `/api/v1/url`   | Create a new short URL.      |
| `GET`  | `/api/v1/url`   | Redirect to the original URL.|
| `DELETE`| `/api/v1/url`   | Delete a short URL.          |
| `GET`  | `/helthy`       | Health check endpoint.       |

### Create a new short URL

**Request:**

```http
POST /api/v1/url
Content-Type: application/json

{
  "url": "https://www.google.com",
  "alias": "google"
}
```

**Response:**

```json
{
  "status": "OK",
  "alias": "google"
}
```

### Redirect to the original URL

**Request:**

```http
GET /api/v1/url?alias=google
```

**Response:**

Redirects to `https://www.google.com`.

### Delete a short URL

**Request:**

```http
DELETE /api/v1/url?alias=google
```

**Response:**

```json
{
  "status": "OK",
  "alias": "google"
}
```
