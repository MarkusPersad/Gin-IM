package database

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"github.com/valkey-io/valkey-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"os"
	"strconv"
	"strings"
	"time"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error
	// GetDB returns the database connection associated with the given context.
	GetDB(ctx context.Context) *gorm.DB
	// Transaction executes a function within a transaction.
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
	// 实现 Base64CaptCha 的 Store 接口
	Set(id string, value string) error

	Get(id string, clear bool) string

	Verify(id, answer string, clear bool) bool

	SetAndTime(ctx *gin.Context, key, value string, timeout int64) error

	GetValue(ctx *gin.Context, key string) string

	DelValue(ctx *gin.Context, key string) error
}

type service struct {
	db        *gorm.DB
	valClient valkey.Client
}

var (
	dbHost           = os.Getenv("DB_HOST")
	dbPort           = os.Getenv("DB_PORT")
	dbUser           = os.Getenv("DB_USERNAME")
	dbPass           = os.Getenv("DB_PASSWORD")
	dbName           = os.Getenv("DB_DATABASE")
	maxOpenConns, _  = strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
	maxIdleConns, _  = strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))
	dbMaxLifetime, _ = strconv.Atoi(os.Getenv("DB_CONN_MAX_LIFETIME"))
	valHost          = os.Getenv("VALKEY_CLIENT_HOST")
	valPort          = os.Getenv("VALKEY_CLIENT_PORT")
	valPass          = os.Getenv("VALKEY_CLIENT_PASSWORD")
	dbInstance       *service
	ctxTxKey         = "Tx"
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)
	isSingularTable, _ := strconv.ParseBool(os.Getenv("DB_SINGULAR_TABLE"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: isSingularTable,
		},
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to get sql db")
		os.Exit(2)
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(dbMaxLifetime) * time.Second)
	valClient, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{valHost + ":" + valPort},
		Password:    valPass,
	})
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to create valkey client")
		os.Exit(2)
	}
	dbInstance = &service{
		db:        db,
		valClient: valClient,
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stats := make(map[string]string)
	sqlDB, err := s.db.DB()
	if err != nil {
		stats["mariadb_status"] = "down"
		stats["mariadb_error"] = fmt.Sprintf("mariadb down: %v", err)
		log.Logger.Fatal().Err(err).Msg("mariadb down")
		return stats
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		stats["mariadb_status"] = "down"
		stats["mariadb_error"] = fmt.Sprintf("mariadb down: %v", err)
		log.Logger.Fatal().Err(err).Msg("mariadb down")
		return stats
	}
	stats["mariadb_status"] = "up"
	stats["mariadb_message"] = "It's healthy"

	dbStats := sqlDB.Stats()
	stats["mariadb_open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["mariadb_in_use"] = strconv.Itoa(dbStats.InUse)
	stats["mariadb_idle"] = strconv.Itoa(dbStats.Idle)
	stats["mariadb_wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["mariadb_wait_duration"] = dbStats.WaitDuration.String()
	stats["mariadb_max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["mariadb_max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["mariadb_message"] = "The database is experiencing heavy load."
	}
	if dbStats.WaitCount > 1000 {
		stats["mariadb_message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["mariadb_message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["mariadb_message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}
	valResult := s.valClient.Do(ctx, s.valClient.B().Ping().Build())
	if valResult.Error() != nil {
		stats["valkey_status"] = "down"
		stats["valkey_error"] = fmt.Sprintf("valkey down: %v", valResult.Error())
		log.Logger.Fatal().Err(valResult.Error()).Msg("valkey down")
		return stats
	}
	stats["valkey_status"] = "up"
	stats["valkey_message"] = "It's healthy"
	valStatus := parseValkeyInfo(valResult.String())
	for k, v := range valStatus {
		stats[k] = v
	}
	return stats
}
func parseValkeyInfo(info string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
	}
	return result
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	sqlDB, _ := s.db.DB()
	log.Logger.Info().Msgf("Disconnected from %s database", dbName)
	s.valClient.Close()
	log.Logger.Info().Msgf("Disconnected from %s valkey", valHost)
	return sqlDB.Close()
}

// GetDB return tx
// If you need to create a Transaction, you must call DB(ctx) and Transaction(ctx,fn)
func (s *service) GetDB(ctx context.Context) *gorm.DB {
	if ctx != nil {
		if tx, ok := ctx.Value(ctxTxKey).(*gorm.DB); ok {
			return tx
		}
		return s.db.WithContext(ctx)
	}
	return s.db
}

func (s *service) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, ctxTxKey, tx)
		return fn(ctx)
	})
}
