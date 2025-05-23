package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"ocean-digital-twin/internal/database/models"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/paulmach/orb"
	"github.com/pressly/goose/v3"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Operations
	SaveChlorophyllData(ctx context.Context, data []models.ChlorophyllData) error
	SaveChlorophyllDataRaw(ctx context.Context, data []models.ChlorophyllData) error
	GetChlorophyllData(ctx context.Context, startTime, endTime time.Time, minLat, minLon, maxLat, maxLon float64, rawData bool) ([]models.ChlorophyllData, error)
	GetLatestChlorophyllTimestamp(ctx context.Context) (time.Time, error)
	GetAllChlorophyllLocations(ctx context.Context) ([]orb.Point, error)
	GetChlorophyllDataAtLocation(ctx context.Context, point orb.Point) ([]models.ChlorophyllData, error)
	GetChlorophyllDataAtTimestamp(ctx context.Context, timestamp time.Time) ([][]models.ChlorophyllData, error)
	GetAllChlorophyllTimestamps(ctx context.Context) ([]time.Time, error)
	UpdateChlorophyllData(ctx context.Context, data []models.ChlorophyllData) error

	SaveCurrentsData(ctx context.Context, data []models.CurrentsData) error
	SaveCurrentsDataRaw(ctx context.Context, data []models.CurrentsData) error
	GetLatestCurrentsTimestamp(ctx context.Context) (time.Time, error)
	GetCurrentsData(ctx context.Context, startTime, endTime time.Time, minLat, minLon, maxLat, maxLon float64, rawData bool) ([]models.CurrentsData, error)
	GetAllCurrentsLocations(ctx context.Context) ([]orb.Point, error)
	GetUCurrentsDataAtLocation(ctx context.Context, point orb.Point) ([]models.UCurrentsData, error)
	GetVCurrentsDataAtLocation(ctx context.Context, point orb.Point) ([]models.VCurrentsData, error)
	UpdateUCurrentsData(ctx context.Context, data []models.UCurrentsData) error
	UpdateVCurrentsData(ctx context.Context, data []models.VCurrentsData) error
	GetAllCurrentsTimestamps(ctx context.Context) ([]time.Time, error)
	GetVCurrentDataAtTimestamp(ctx context.Context, timestamp time.Time) ([][]models.VCurrentsData, error)
	GetUCurrentDataAtTimestamp(ctx context.Context, timestamp time.Time) ([][]models.UCurrentsData, error)

	GetCount() int
	UpdateCount(int) error
	NewCount() (int, error)

	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string
	// Migrate database up
	Up() error
	Close() error
}

type service struct {
	db *sql.DB
}

var (
	database   = os.Getenv("BLUEPRINT_DB_DATABASE")
	password   = os.Getenv("BLUEPRINT_DB_PASSWORD")
	username   = os.Getenv("BLUEPRINT_DB_USERNAME")
	port       = os.Getenv("BLUEPRINT_DB_PORT")
	host       = os.Getenv("BLUEPRINT_DB_HOST")
	schema     = os.Getenv("BLUEPRINT_DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func (s *service) Up() error {
	if err := goose.SetDialect("postgres"); err != nil {
		slog.Error("Failed to select dialect", "err", err)
		return err
	}

	if err := goose.Up(s.db, "internal/database/migrations"); err != nil {
		slog.Error("Failed run migrations", "err", err)
		return err
	}

	slog.Info("Migrations completed successfully")
	return nil
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}
