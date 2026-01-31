package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "lizobly/ctc-db-api/docs"
	"lizobly/ctc-db-api/internal/accessory"
	internalJWT "lizobly/ctc-db-api/internal/jwt"
	"lizobly/ctc-db-api/internal/traveller"
	"lizobly/ctc-db-api/internal/user"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	pkgMiddleware "lizobly/ctc-db-api/pkg/middleware"
	"lizobly/ctc-db-api/pkg/telemetry"
	"lizobly/ctc-db-api/pkg/validator"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//	@title			CTC DB API
//	@version		1.0
//	@description	REST API for managing CTC game database including travellers (characters), accessories (equipment), and user authentication.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Liz
//	@contact.email	j2qgehn84@mozmail.com

//	@BasePath	/api/v1

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer " followed by your JWT token (include the word Bearer and a space before the token)
func main() {
	// Load environment variables
	if err := godotenv.Load("config.env"); err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	// Initialize logger
	env := helpers.EnvWithDefault("ENVIRONMENT", "development")
	logger := initLogger(env)
	defer logger.Sync()

	// Initialize tracer
	tracerProvider := initTracer(logger)
	defer shutdownTracer(tracerProvider, logger)

	// Initialize database
	db, dbConn := initDatabase(logger)
	defer closeDatabase(dbConn, logger)

	// Initialize application
	app := initApplication(db, logger)

	// Configure server with timeouts
	addr := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	requestTimeoutStr := helpers.EnvWithDefault("REQUEST_TIMEOUT", "30s")
	requestTimeout, _ := time.ParseDuration(requestTimeoutStr)
	writeTimeout := requestTimeout + (5 * time.Second) // Add buffer for response writing

	server := &http.Server{
		Addr:         addr,
		Handler:      app,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: writeTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	logger.Info("starting server",
		zap.String("service.name", "ctc-db-api"),
		zap.String("environment", env),
		zap.String("address", addr),
		zap.Duration("request.timeout", requestTimeout),
		zap.Duration("write.timeout", writeTimeout),
	)
	app.Logger.Fatal(server.ListenAndServe())
}

func initLogger(env string) *logging.Logger {
	logger, err := logging.NewLogger(env)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	zap.ReplaceGlobals(logger.Logger)

	logger.Info("logger initialized",
		zap.String("service.name", "ctc-db-api"),
		zap.String("environment", env),
	)
	return logger
}

func initTracer(logger *logging.Logger) *telemetry.TracerProvider {
	tracerProvider, err := telemetry.InitTracer(logger.Logger)
	if err != nil {
		logger.Fatal("Failed to initialize tracer", zap.Error(err))
	}
	return tracerProvider
}

func shutdownTracer(tracerProvider *telemetry.TracerProvider, logger *logging.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := tracerProvider.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown tracer", zap.Error(err))
	}
}

func initDatabase(logger *logging.Logger) (*gorm.DB, *sql.DB) {
	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbUser := os.Getenv("DATABASE_USER")
	dbPass := os.Getenv("DATABASE_PASS")
	dbName := os.Getenv("DATABASE_NAME")
	dbSSLMode := helpers.EnvWithDefault("DATABASE_SSLMODE", "disable")

	dsn := fmt.Sprintf("sslmode=%s host=%s port=%s user=%s password='%s' dbname=%s timezone=%s",
		dbSSLMode, dbHost, dbPort, dbUser, dbPass, dbName, "Asia/Jakarta")

	dbConn, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Fatal("Failed to open database connection",
			zap.Error(err),
			zap.String("db.host", dbHost),
			zap.String("db.port", dbPort))
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: dbConn,
	}), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		logger.Fatal("Failed to initialize GORM",
			zap.Error(err),
			zap.String("db.system", "postgres"))
	}

	if err = dbConn.Ping(); err != nil {
		logger.Fatal("Failed to ping database",
			zap.Error(err),
			zap.String("db.host", dbHost))
	}

	// Configure connection pool
	dbConn.SetMaxIdleConns(10)
	dbConn.SetMaxOpenConns(100)
	dbConn.SetConnMaxLifetime(time.Hour)
	dbConn.SetConnMaxIdleTime(10 * time.Minute)

	logger.Info("database connected",
		zap.String("db.system", "postgres"),
		zap.String("db.host", dbHost),
	)

	return db, dbConn
}

func closeDatabase(dbConn *sql.DB, logger *logging.Logger) {
	if err := dbConn.Close(); err != nil {
		logger.Error("Failed to close database connection", zap.Error(err))
	}
}

func initApplication(db *gorm.DB, logger *logging.Logger) *echo.Echo {
	e := echo.New()

	// Load request timeout configuration
	requestTimeoutStr := helpers.EnvWithDefault("REQUEST_TIMEOUT", "30s")
	requestTimeout, err := time.ParseDuration(requestTimeoutStr)
	if err != nil {
		logger.Fatal("Invalid REQUEST_TIMEOUT format",
			zap.String("request.timeout", requestTimeoutStr),
			zap.Error(err))
	}

	// Setup middleware
	e.Use(pkgMiddleware.TracingMiddleware(logger))
	e.Use(pkgMiddleware.RequestIDMiddleware())
	e.Use(pkgMiddleware.TimeoutMiddleware(requestTimeout, logger))
	e.Use(pkgMiddleware.RequestBodyLoggingMiddleware(logger))

	// Setup Swagger
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Setup validator
	setupValidator(e, logger)

	// Setup repositories, services, and handlers
	setupRoutes(e, db, logger)

	return e
}

func setupValidator(e *echo.Echo, logger *logging.Logger) {
	v, err := validator.NewValidator()
	if err != nil {
		logger.Fatal("Failed to initialize validator", zap.Error(err))
	}
	e.Validator = v
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set("validator", v)
			return next(ctx)
		}
	})
}

func setupRoutes(e *echo.Echo, db *gorm.DB, logger *logging.Logger) {
	// Initialize token service
	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	if jwtSecretKey == "" {
		logger.Fatal("JWT_SECRET_KEY environment variable must be set")
	}
	jwtTimeoutStr := helpers.EnvWithDefault("JWT_TIMEOUT", "10m")
	jwtTimeout, err := time.ParseDuration(jwtTimeoutStr)
	if err != nil {
		logger.Fatal("Invalid JWT_TIMEOUT format",
			zap.String("jwt.timeout", jwtTimeoutStr),
			zap.Error(err))
	}
	tokenService := internalJWT.NewTokenService(jwtSecretKey, jwtTimeout, logger)

	// Initialize repositories
	travellerRepo := traveller.NewTravellerRepository(db, logger)
	accessoryRepo := accessory.NewAccessoryRepository(db, logger)
	userRepo := user.NewUserRepository(db, logger)

	// Initialize services
	travellerService := traveller.NewTravellerService(travellerRepo, logger)
	userService := user.NewUserService(userRepo, tokenService, logger)
	accessoryService := accessory.NewAccessoryService(accessoryRepo, logger)

	// Setup API group with optional JWT middleware
	v1 := e.Group("/api/v1")
	if helpers.EnvWithDefaultBool("AUTH_IS_ENABLED", false) {
		jwtMiddleware := pkgMiddleware.NewJWTMiddleware()
		v1.Use(jwtMiddleware)
	}

	// Register handlers
	traveller.NewTravellerHandler(v1, travellerService, logger)
	user.NewUserHandler(v1, userService, logger)
	accessory.NewAccessoryHandler(v1, accessoryService, logger)
}
