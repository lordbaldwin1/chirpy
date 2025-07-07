# chirpy

A production-like HTTP server written with the Go standard library, SQLC, Goose, and Postgres.

Includes auth with JWT token creation and validation as well as CRUD operations with query parameters.

## Motivation

My start in writing production Go HTTP servers starts with mastering the fundamentals. This project is to aid in that by concretely mastering the fundamentals of writing HTTP servers in Go without the aid of outside frameworks or libraries.

## Goal

The goal with `chirpy` is to be a small-scale HTTP server that I can use as reference in future work.

## Installation

Inside a Go module run:
```bash
go get github.com/lordbaldwin1/chirpy
```

## Chirpy API Documentation

## Base URL
`/api`

## Authentication

*   **JWT Access Tokens**: Used for most authenticated endpoints. Valid for 1 hour.
    *   Sent in the `Authorization` header as `Bearer <token>`.
*   **Refresh Tokens**: Used to obtain new JWT access tokens. Valid for 60 days.
    *   Sent in the `Authorization` header as `Bearer <token>`.
*   **Polka API Key**: Used for webhook authentication.
    *   Sent in the `Authorization` header as `Apikey <key>`.

---

## Endpoints

### 1. Health Check

**GET** `/api/healthz`

*   **Description**: Checks the health of the API server.
*   **Response**:
    *   `200 OK`: `text/plain` body with "OK".

### 2. Chirps

#### Create Chirp

**POST** `/api/chirps`

*   **Description**: Creates a new chirp. Profanity is filtered. Max 140 characters.
*   **Authentication**: Required (JWT Access Token)
*   **Request Body**: `application/json`
    ```json
    {
      "body": "This is my new chirp!"
    }
    ```
*   **Responses**:
    *   `201 Created`: `application/json`
        ```json
        {
          "id": "uuid",
          "created_at": "timestamp",
          "updated_at": "timestamp",
          "body": "string",
          "user_id": "uuid"
        }
        ```
    *   `400 Bad Request`: If chirp body is too long.
    *   `401 Unauthorized`: If JWT is missing or invalid.
    *   `500 Internal Server Error`: For other server issues.

#### Get All Chirps

**GET** `/api/chirps`

*   **Description**: Retrieves all chirps or filters by author.
*   **Query Parameters**:
    *   `author_id` (optional): `uuid` - Filters chirps by the specified user ID.
    *   `sort` (optional): `string` - Sorts chirps by `created_at`. Accepts `asc` (ascending) or `desc` (descending).
*   **Response**:
    *   `200 OK`: `application/json` - An array of chirp objects.
        ```json
        [
          {
            "id": "uuid",
            "created_at": "timestamp",
            "updated_at": "timestamp",
            "body": "string",
            "user_id": "uuid"
          }
        ]
        ```
    *   `400 Bad Request`: If `author_id` is invalid.
    *   `500 Internal Server Error`: For database retrieval issues.

#### Get Chirp by ID

**GET** `/api/chirps/{chirpID}`

*   **Description**: Retrieves a single chirp by its ID.
*   **Path Parameters**:
    *   `chirpID`: `uuid` - The ID of the chirp to retrieve.
*   **Response**:
    *   `200 OK`: `application/json` - A single chirp object.
        ```json
        {
          "id": "uuid",
          "created_at": "timestamp",
          "updated_at": "timestamp",
          "body": "string",
          "user_id": "uuid"
        }
        ```
    *   `404 Not Found`: If the chirp does not exist.
    *   `500 Internal Server Error`: For database retrieval issues.

#### Delete Chirp

**DELETE** `/api/chirps/{chirpID}`

*   **Description**: Deletes a chirp if the authenticated user is the owner.
*   **Authentication**: Required (JWT Access Token)
*   **Path Parameters**:
    *   `chirpID`: `uuid` - The ID of the chirp to delete.
*   **Responses**:
    *   `204 No Content`: If the chirp was successfully deleted.
    *   `400 Bad Request`: If `chirpID` is invalid.
    *   `401 Unauthorized`: If JWT is missing or invalid.
    *   `403 Forbidden`: If the user is not the owner of the chirp.
    *   `404 Not Found`: If the chirp does not exist.
    *   `500 Internal Server Error`: For database deletion issues.

### 3. Users

#### Create User (Register)

**POST** `/api/users`

*   **Description**: Registers a new user.
*   **Request Body**: `application/json`
    ```json
    {
      "email": "user@example.com",
      "password": "mySecurePassword123"
    }
    ```
*   **Responses**:
    *   `201 Created`: `application/json`
        ```json
        {
          "id": "uuid",
          "created_at": "timestamp",
          "updated_at": "timestamp",
          "email": "user@example.com",
          "is_chirpy_red": false
        }
        ```
    *   `500 Internal Server Error`: For database or password hashing issues.

#### User Login

**POST** `/api/login`

*   **Description**: Authenticates a user and returns JWT access and refresh tokens.
*   **Request Body**: `application/json`
    ```json
    {
      "email": "user@example.com",
      "password": "mySecurePassword123"
    }
    ```
*   **Responses**:
    *   `200 OK`: `application/json`
        ```json
        {
          "id": "uuid",
          "created_at": "timestamp",
          "updated_at": "timestamp",
          "email": "user@example.com",
          "is_chirpy_red": false,
          "token": "jwt_access_token_string",
          "refresh_token": "refresh_token_string"
        }
        ```
    *   `401 Unauthorized`: If password is incorrect.
    *   `404 Not Found`: If user email does not exist.
    *   `500 Internal Server Error`: For token generation or database issues.

#### Update User Profile

**PUT** `/api/users`

*   **Description**: Updates the authenticated user's email and password.
*   **Authentication**: Required (JWT Access Token)
*   **Request Body**: `application/json`
    ```json
    {
      "email": "newemail@example.com",
      "password": "newSecurePassword123"
    }
    ```
*   **Responses**:
    *   `200 OK`: `application/json`
        ```json
        {
          "id": "uuid",
          "created_at": "timestamp",
          "updated_at": "timestamp",
          "email": "newemail@example.com",
          "is_chirpy_red": false
        }
        ```
    *   `401 Unauthorized`: If JWT is missing or invalid.
    *   `500 Internal Server Error`: For database or password hashing issues.

### 4. Token Management

#### Refresh Token

**POST** `/api/refresh`

*   **Description**: Exchanges a valid refresh token for a new JWT access token.
*   **Authentication**: Required (Refresh Token)
*   **Request Body**: None
*   **Responses**:
    *   `200 OK`: `application/json`
        ```json
        {
          "token": "new_jwt_access_token_string"
        }
        ```
    *   `401 Unauthorized`: If refresh token is missing, invalid, or expired.
    *   `500 Internal Server Error`: For token generation or database issues.

#### Revoke Token

**POST** `/api/revoke`

*   **Description**: Revokes a refresh token, invalidating it immediately.
*   **Authentication**: Required (Refresh Token)
*   **Request Body**: None
*   **Responses**:
    *   `204 No Content`: If the refresh token was successfully revoked.
    *   `400 Bad Request`: If refresh token is missing.
    *   `500 Internal Server Error`: If token could not be revoked in the database.

### 5. Webhooks

#### Polka Webhook

**POST** `/api/polka/webhooks`

*   **Description**: Endpoint for Polka webhooks to signal user upgrades.
*   **Authentication**: Required (Polka API Key)
*   **Request Body**: `application/json`
    ```json
    {
      "event": "user.upgraded",
      "data": {
        "user_id": "uuid_of_user_to_upgrade"
      }
    }
    ```
*   **Responses**:
    *   `204 No Content`: If the event was processed successfully (user upgraded or event ignored).
    *   `401 Unauthorized`: If Polka API Key is missing or incorrect.
    *   `500 Internal Server Error`: If user ID cannot be parsed or user cannot be found/upgraded in DB.

### 6. Admin Endpoints

#### File Server Metrics

**GET** `/admin/metrics`

*   **Description**: Displays the number of times the file server has been hit.
*   **Response**:
    *   `200 OK`: `text/html` - HTML page displaying the hit count.

#### Reset Metrics and Database

**POST** `/admin/reset`

*   **Description**: Resets the file server hit count to 0 and clears all user data from the database. **Only available in `dev` environment.**
*   **Response**:
    *   `200 OK`: `text/plain` - Confirmation message.
    *   `403 Forbidden`: If not in `dev` environment.
    *   `500 Internal Server Error`: If database reset fails.

---

