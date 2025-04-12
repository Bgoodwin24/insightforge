# InsightForge

## Project Overview
A web application that allows users to upload datasets and explore them through visualizations and analysis tools.

## Architecture
```mermaid
graph TD
    A[React Components] --> B[State Management]
    B --> C[Chart.js/D3.js]
    D[API Routes] --> E[Service Layer]
    E --> F[Data Access Layer]
    F --> G[PostgreSQL]
    G --> H[File Storage]
```

## Dependencies
### Backend
- Go (v1.21.5 or later)
- PostgreSQL (v17.4 or later)
- Gin Web Framework (v1.10.0 or later)
- GORM (v1.25.12 or later)
- Godotenv (v1.5.1 or later)

### Frontend
- Node.js (v21.7.0 or later)
- NPM (v10.5.0 or later)
- Chart.js (for data visualization) (v4.4.8 or later)

## Development Tools
- Air (live reloading) (v1.61.7 or later)
- Goose (database migrations) (v3.24.2 or later)
- Testify (testing framework) (v1.10.0 or later)
- HTTPTest (v2.2.0 or later)

## Features
- Data upload and storage
- Interactive visualizations
- Dataset analysis capabilities
- User accounts and saved projects
- Export options for visualizations

## Setup Instructions
### Clone the repository
``` bash
git clone github.com/Bgoodwin24/insightforge.git
```

### Install dependencies
- Golang
``` bash
sudo apt update
sudo apt install golang-go
```
- Gin Web Framework
``` bash
go get -u github.com/gin-gonic/gin
```
- GORM
``` bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
```
- PostgreSQL
``` bash
sudo apt update
sudo apt install postgresql postgresql-contrib
```
- PostgreSQL Driver
``` bash
go get -u github.com/lib/pq
```
- Goose
``` bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```
- Node.js/NPM
``` bash
sudo apt update
sudo apt install nodejs npm
```
- Air
``` bash
go install github.com/air-verse/air@latest
```
- Chart.js
``` bash
npm install chart.js
```
- Godotenv
``` bash
go get -u github.com/joho/godotenv
```
- Testify
``` bash
go get -u github.com/stretchr/testify
```
- HTTPTest
``` bash
go get -u github.com/appleboy/gofight/v2
```

## Environment Setup
### Create a .env file in the project root
``` bash
touch .env
```

### Add the following variables to your .env file

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password_here
DB_NAME=insightforge
API_PORT=8080
FRONT_END_PORT=3000
API_URL=http://localhost:8080
ENV=development
JWT_SECRET=replace_with_secure_random_string
UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE=52428800  # 50MB in bytes (this limit could be adjusted in a production environment)
```

**Where the values match your local implementation**

## Database Setup
### Create the database
``` bash
sudo -u postgres createdb your_database_name
```
**Where your_database_name is replaced by your actual database name**

### Run migrations

``` bash
goose up
```

## Running the application
### Start the backend with Air
``` bash
air
```

### In another terminal, start the frontend
``` bash
cd frontend
npm start
```

... more to instructions to come

## Screenshots
[Will be added when application UI is developed]

## License
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
This project is licensed under the terms of the MIT license.

## Author
Bgoodwin24
