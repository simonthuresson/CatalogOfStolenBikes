# Catalog of Stolen Bikes

## Overview
A simple Go server for cataloging stolen bikes. This application is deployed with Docker.

## Requirements
- Docker Compose

## Architectural Overview
![Architecture diagram](Overview.png)

### Description of Architecture
The entry point (main) begins by initializing the database. Once initialization is complete, the server's endpoints are registered. The functions associated with each endpoint are fetched from the utils directory, which contains several utility files organized by specific domains. Each utility file uses the initialized database to perform queries. After this setup process is complete, the main file starts the server.

## Quick Start
From root directory:

```bash
./start-server.sh
```
or
```bash
docker compose -f docker-compose.prod.yml up -d --build
```
This starts the server and database. The endpoints are accessible from localhost:8080

## API Documentation

### Base URL
```
/api
```

## Endpoints

### Police
**Middleware**: Requires police authentication, except for police creation (POST) for simplification
Base URL: `/api/police`

- **GET /** - Retrieve all police records
- **POST /** - Create a new police record
  ```json
  {
    "email": "user@example.com",
    "name": "policeName",
    "password": "password123"
  }
  ```
- **PATCH /:id** - Update a police record by ID
  ```json
  {
    "name": "policeName",
    "password": "password123"
  }
  ```
- **DELETE /:id** - Delete a police record by ID

### Citizen
Base URL: `/api/citizen`

- **GET /** - Retrieve all citizen records
- **POST /** - Create a new citizen record
  ```json
  {
    "email": "user@example.com",
    "name": "citizenName",
    "password": "password123"
  }
  ```

### Bike
Base URL: `/api/bike`

**Middleware**: Requires citizen authentication

- **GET /** - Retrieve all bikes
- **POST /** - Report a stolen bike
  ```json
  {
    "description": "21-speed mountain bike with front suspension"
  }
  ```
- **GET /found/:id** - Mark a bike as found by ID and unassign the police officer

### Login
Base URL: `/login`

- **POST /citizen** - Authenticate a citizen and return a session/token
- **POST /police** - Authenticate a police and return a session/token
- Both accepts JSON:
  ```json
  {
    "email": "user@example.com",
    "password": "password123"
  }
  ```

## Authentication

 - Uses JWT cookies
 - One type of cookie for police authentication and one for citizen
 - Valid for 24 h
 - Http-only cookie

