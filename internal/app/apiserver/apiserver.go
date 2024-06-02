package apiserver

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"os"
	"petition_api/internal/app/handlers/httpHandlers"
	repository "petition_api/internal/app/repositories"
	"petition_api/utils/logger"
)

type ApiServer struct {
	config *Config
	logger *logrus.Logger
	db     *gorm.DB
	router *gin.Engine
}

// NewApiServer Создает новый сервер
func NewApiServer(config *Config) *ApiServer {
	return &ApiServer{
		config: config,
		logger: logrus.New(),
		db:     nil,
	}
}

// Start Запуск сервера
func (s *ApiServer) Start() error {
	if err := s.configureLogger(); err != nil {
		return err
	}
	s.logger.Info("Starting API Server...")

	if err := configureDB(s); err != nil {
		return err
	}
	s.logger.Info("Connect to database successfully")

	// Создание роутера
	s.router = gin.Default()
	// Настройка CORS
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200", "http://localhost:8080"},
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Для проверки работы сервера
	s.router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Создание роутов для юзера
	userRoutes := httpHandlers.NewUserModelRoute(
		repository.NewUserRepository(s.db, s.logger),
		repository.NewSessionRepo(s.db, s.logger),
		s.logger)

	userRoutes.BindUserToRoute(s.router.Group("/user"))

	s.logger.Info("API Server started!")
	// Запуск сервера
	if err := s.router.Run(":8080"); err != nil {
		panic(err)
	}
	s.logger.Info("API Server stopped!")
	return nil
}

// configureLogger конфигурирует логгер
func (s *ApiServer) configureLogger() error {
	level, err := logrus.ParseLevel(s.config.App.LogLevel)
	if err != nil {
		return err
	}
	s.logger = &logrus.Logger{
		Out:       os.Stderr,
		Level:     level,
		Formatter: &logger.CustomFormatter{},
	}
	return nil
}
