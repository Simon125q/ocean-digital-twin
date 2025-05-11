# Adding a New Data Source

This document outlines the step-by-step process for integrating a new external data source into the application.

## Prerequisites

- Go 1.24+
- Docker and Docker Compose
- Make
- Goose migration tool installed (`go install github.com/pressly/goose/v3/cmd/goose@latest`)

## Steps

### Step 1: Define the Database Schema

Create table(s) in the database for the new data source.

1.  **While in the root directory of the project create a new migration file:**

    ```bash
    goose -dir backend/internal/database/migrations create add_source_name_tables sql
    ```

    Replace `source_name` with a descriptive name for your data source (snake_case).

2.  **Implement the table creation:**
    Edit the generated SQL file (`backend/internal/database/migrations/..._add_source_name_tables.sql`) to define the table schema. Include columns for timestamp, location (using PostGIS `GEOGRAPHY` is recommended), and the specific data variables. Add indexes for querying.

3.  **Apply the migration:**
    Ensure your database is running (`make docker-run`), then run:
    ```bash
    goose -dir backend/internal/database/migrations up
    ```

### Step 2: Set Up Automatic Data Cleanup

Implement a SQL function to remove old data and set it up with pg_cron.

1.  **Create a new migration file:**

    ```bash
    goose -dir backend/internal/database/migrations create shedule_source_name_cleanup sql
    ```

2.  **Implement the cleanup function:**
    Edit the generated SQL file to define a PostgreSQL function that deletes data from your new table older than your desired window (e.g., 30 days).

3.  **Apply the migration:**
    ```bash
    goose -dir backend/internal/database/migrations up
    ```

### Step 3: Define the Data Model

Create a Go struct to represent the data structure.

1.  **Create a new file:**
    Create `backend/internal/models/source_name.go`.

2.  **Define the struct:**
    Define a Go struct (`SourceNameData`) with fields matching your database table columns. Include a function to convert data to GeoJSON if it's spatial.

### Step 4: Implement Database Queries

Add functions to interact with the new data table.

1.  **Create a new file:**
    Create `backend/internal/database/queries-source_name.go`.

2.  **Add methods to the `Service` interface:**
    In `backend/internal/database/database.go`, add method signatures for saving data (`SaveSourceNameData`), retrieving data (`GetSourceNameData`), getting the latest timestamp (`GetLatestSourceNameTimestamp`), and calling the cleanup function (`CleanupOldSourceNameData`).

3.  **Implement the methods:**
    In `backend/internal/database/queries-source_name.go`, implement the methods for the `service` struct using SQL queries.

### Step 5: Implement Data Downloading and Processing

Write code to fetch data from the external source and convert it to your Go model.

1.  **Create a new file:**
    Create a file in `backend/internal/utils/erddap/source_name.go`

2.  **Implement the data download function:**
    Implement function `DownloadDataSourceNameData()` to download data, and process the raw data into a slice of `[]models.SourceNameData`.

### Step 6: Implement Data Update Scheduling

Create a scheduler component to periodically fetch and save data.

1.  **Create a new file:**
    Create `backend/internal/utils/scheduler/source_name.go`.

2.  **Implement the updater:**
    Create an function `updateSourceNameData()` function. Inside implement the logic to:

    - Call the database cleanup function.
    - Get the latest timestamp from the database to determine the data fetching start time.
    - Download the data using your downloader.
    - Save the processed data to the database.
    - Include error handling for edge cases and robust logging.

3.  **Call the update function**
    Call the implemented function in `update()` function in file `scheduler/updater.go`

### Step 7: Implement API Handlers

Create handlers to expose the new data source through the API.

1.  **Create a new file:**
    Create `backend/internal/server/handlers-source_name.go`.

2.  **Implement the handlers:**
    Create handler functions for your endpoints (e.g., `getSourceNameData`). These handlers will:
    - Parse request parameters (time range, bounding box).
    - Call the appropriate method(s) on the database `Service` to retrieve data.
    - Convert the retrieved data to a suitable response format (e.g., GeoJSON using your model's helper function).
    - Write the response with the correct headers and status code.

### Step 8: Register the New Routes

Add the routes for your new data source to the main router in `backend/internal/server/routes.go`.

---

Information on data interpolation and how to add new data to be interpolated can be found in `docs/interpolation.md`
