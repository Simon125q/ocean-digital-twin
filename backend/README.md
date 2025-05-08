# Ocean Digital Twin Backend

This is the backend service for the Ocean Digital Twin project. It's built with Go and utilizes a set of modern technologies to provide a robust and scalable foundation.

## Technologies Used

- **Go:** The primary programming language for the backend logic.
- **Chi:** A lightweight, idiomatic, and composable router for building HTTP services in Go.
- **PostgreSQL:** A powerful open-source relational database for data storage.
- **PostGIS:** A spatial database extender for PostgreSQL, enabling geographic object storage and queries.
- **Docker:** Used for containerizing the PostgreSQL database, ensuring a consistent development environment.
- **Docker Compose:** A tool for defining and running multi-container Docker applications (specifically for managing the database container).
- **Goose:** A database migration tool for managing schema changes in a version-controlled manner.
- **Air:** A live-reloading tool for Go applications, streamlining the development workflow.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

Ensure you have the following installed:

- **Go:** [https://golang.org/doc/install](https://golang.org/doc/install)
- **Docker & Docker Compose:** [https://www.docker.com/get-started](https://www.docker.com/get-started)
- **Make:** Usually pre-installed on macOS and Linux. For Windows, consider using WSL or installing a Make equivalent.
- **Goose (Database Migration Tool):**
  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```
- **Air (Live Realoading for go apps):**
  ```bash
  go install github.com/air-verse/air@latest
  ```

### Installation

1.  Clone the repository:

    ```bash
    git clone <repository_url>
    cd ocean-digital-twin/backend
    ```

2.  Install Go dependencies:

    ```bash
    go mod tidy
    ```

3.  Set up environment variables:
    Create a `.env` file in the project root based on the example or documentation provided. This file will contain database credentials and other configuration.

    ```
    PORT=3000
    APP_ENV=local
    BLUEPRINT_DB_HOST=localhost
    BLUEPRINT_DB_PORT=5432
    BLUEPRINT_DB_DATABASE=dt_database
    BLUEPRINT_DB_USERNAME=postgres
    BLUEPRINT_DB_PASSWORD=password
    BLUEPRINT_DB_SCHEMA=public
    ```

4.  Start the database container:

    ```bash
    make docker-run
    ```

    This will use Docker Compose to start a PostgreSQL container with the specified configuration and persistent volume.

5.  Run database migrations:
    Migrations are used to set up and update the database schema.

    - **Creating a new migration:**
      To create a new SQL migration file (e.g., for adding a new table or altering an existing one), run:

      ```bash
      goose -dir backend/internal/database/migrations create <migration_name> sql
      ```

      Replace `<migration_name>` with a descriptive name (e.g., `create_users_table`, `add_product_price_column`). Edit the generated SQL file to define your schema changes in the `-- +goose Up` and `-- +goose Down` sections.

    - **Running pending migrations:**
      To apply all pending migrations to your database, run:
      ```bash
      goose -dir backend/internal/database/migrations up
      ```

### Running the Backend

Once the prerequisites are met, including the running database container and applied migrations, you can run the backend application:

- **Run the application (standard):**

  ```bash
  make run
  ```

- **Run the application (with live reload for development):**
  ```bash
  make watch
  ```
  Requires `air` to be installed (the `make watch` command should guide you through this if it's not found).

### MakeFile Commands

The `Makefile` provides convenient shortcuts for common tasks:

```bash
make all            # Build the application and run tests
make build          # Build the application binary
make run            # Run the compiled application
make docker-run     # Create and start the database container using Docker Compose
make docker-down    # Stop and remove the database container using Docker Compose
make itest          # Run database integration tests
make test           # Run the full test suite
make watch          # Run the application with live reloading (for development)
make clean          # Clean up the generated binary
```

---
