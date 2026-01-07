package main

import (
	"database/sql"
	"fmt"
	_ "lizobly/ctc-db-api/docs"
	"lizobly/ctc-db-api/user"

	postgresRepo "lizobly/ctc-db-api/internal/repository/postgres"
	"lizobly/ctc-db-api/internal/rest"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	pkgMiddleware "lizobly/ctc-db-api/pkg/middleware"
	"lizobly/ctc-db-api/pkg/validator"
	"lizobly/ctc-db-api/traveller"
	"log"
	"os"

	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//	@title			COTC DB API
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
		zap.String("service.name", "cotc-db-api"),
		zap.String("environment", env),
	)

	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbUser := os.Getenv("DATABASE_USER")
	dbPass := os.Getenv("DATABASE_PASS")
	dbName := os.Getenv("DATABASE_NAME")
	dsn := fmt.Sprintf("sslmode=disable host=%s port=%s user=%s password='%s' dbname=%s timezone=%s", dbHost, dbPort, dbUser, dbPass, dbName, "Asia/Jakarta")

	dbConn, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("failed open database ", err)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: dbConn,
	}), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to open gorm ", err)
	}

	err = dbConn.Ping()
	if err != nil {
		log.Fatal("failed to ping database ", err)
	}

	logger.Info("database connected",
		zap.String("db.system", "postgres"),
		zap.String("db.host", dbHost),
	)

	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Fatal("got error when closing the DB connection", err)
		}
	}()

	addr := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	e := echo.New()

	// Add request ID middleware
	e.Use(pkgMiddleware.RequestIDMiddleware(logger))

	e.GET("/swagger/*", echoSwagger.WrapHandler)
	// Validator
	validator := validator.NewValidator()
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
	userRepo := postgresRepo.NewUserRepository(db, logger)

	// Service
	travellerService := traveller.NewTravellerService(travellerRepo, logger)
	userService := user.NewUserService(userRepo, logger)

	v1 := e.Group("/api/v1")
	// JWT Middleware Flag
	if helpers.EnvWithDefaultBool("AUTH_IS_ENABLED", false) {
		v1.Use(jwtMiddleware)
	}

	// Handler
	rest.NewTravellerHandler(v1, travellerService)
	rest.NewUserHandler(v1, userService)

	logger.Info("starting server",
		zap.String("service.name", "cotc-db-api"),
		zap.String("environment", env),
		zap.String("address", addr),
	)

	e.Logger.Fatal(e.Start(addr))
}
