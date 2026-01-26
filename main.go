package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"lizobly/ctc-db-api/accessory"
	_ "lizobly/ctc-db-api/docs"
	postgresRepo "lizobly/ctc-db-api/internal/repository/postgres"
	"lizobly/ctc-db-api/internal/rest"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	pkgMiddleware "lizobly/ctc-db-api/pkg/middleware"
	"lizobly/ctc-db-api/pkg/telemetry"
	"lizobly/ctc-db-api/pkg/validator"
	"lizobly/ctc-db-api/traveller"
	"lizobly/ctc-db-api/user"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//	@title			CTC DB API
//	@version		1.0
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Liz
//	@contact.email	j2qgehn84@mozmail.com

// @BasePath	/api/v1
func main() {

	err := godotenv.Load("config.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	// Initialize logger
	env := helpers.EnvWithDefault("ENVIRONMENT", "development")
	logger, err := logging.NewLogger(env)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger.Logger)

	logger.Info("logger initialized",
		zap.String("service.name", "ctc-db-api"),
		zap.String("environment", env),
	)

	// Initialize OpenTelemetry tracer
	tracerProvider, err := telemetry.InitTracer(logger.Logger)
	if err != nil {
		logger.Fatal("Failed to initialize tracer", zap.Error(err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracerProvider.Shutdown(ctx); err != nil {
			logger.Error("Failed to shutdown tracer", zap.Error(err))
		}
	}()

	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbUser := os.Getenv("DATABASE_USER")
	dbPass := os.Getenv("DATABASE_PASS")
	dbName := os.Getenv("DATABASE_NAME")
	dbSSLMode := helpers.EnvWithDefault("DATABASE_SSLMODE", "disable")
	dsn := fmt.Sprintf("sslmode=%s host=%s port=%s user=%s password='%s' dbname=%s timezone=%s", dbSSLMode, dbHost, dbPort, dbUser, dbPass, dbName, "Asia/Jakarta")

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

	err = dbConn.Ping()
	if err != nil {
		logger.Fatal("Failed to ping database",
			zap.Error(err),
			zap.String("db.host", dbHost))
	}

	logger.Info("database connected",
		zap.String("db.system", "postgres"),
		zap.String("db.host", dbHost),
	)

	defer func() {
		if err := dbConn.Close(); err != nil {
			logger.Error("Failed to close database connection",
				zap.Error(err))
		}
	}()

	addr := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	e := echo.New()

	// Add OTel tracing middleware FIRST (if enabled)
	// Must run before RequestIDMiddleware so trace_id is in context
	e.Use(pkgMiddleware.TracingMiddleware(logger))

	// Add request ID middleware
	e.Use(pkgMiddleware.RequestIDMiddleware(logger))

	e.GET("/swagger/*", echoSwagger.WrapHandler)
	// Validator
	validator, err := validator.NewValidator()
	if err != nil {
		logger.Fatal("Failed to initialize validator", zap.Error(err))
	}
	e.Validator = validator
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set("validator", validator)
			return next(ctx)
		}
	})

	// Middleware
	jwtMiddleware := pkgMiddleware.NewJWTMiddleware()

	// Repository
	travellerRepo := postgresRepo.NewTravellerRepository(db, logger)
	accessoryRepo := postgresRepo.NewAccessoryRepository(db, logger)
	userRepo := postgresRepo.NewUserRepository(db, logger)

	// Service
	travellerService := traveller.NewTravellerService(travellerRepo, logger)
	userService := user.NewUserService(userRepo, logger)
	accessoryService := accessory.NewAccessoryService(accessoryRepo, logger)

	v1 := e.Group("/api/v1")
	// JWT Middleware Flag
	if helpers.EnvWithDefaultBool("AUTH_IS_ENABLED", false) {
		v1.Use(jwtMiddleware)
	}

	// Handler
	rest.NewTravellerHandler(v1, travellerService)
	rest.NewUserHandler(v1, userService)
	rest.NewAccessoryHandler(v1, accessoryService)

	logger.Info("starting server",
		zap.String("service.name", "ctc-db-api"),
		zap.String("environment", env),
		zap.String("address", addr),
	)

	e.Logger.Fatal(e.Start(addr))
}
