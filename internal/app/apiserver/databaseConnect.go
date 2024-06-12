package apiserver

import (
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"petition_api/internal/app/models"
)

func configureDB(s *ApiServer) error {
	// Mysql in openserver 6.0.0
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		s.config.Database.Username,
		s.config.Database.Password,
		s.config.Database.Host,
		s.config.Database.Port,
	)
	db, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	if err != nil {
		s.logger.Errorf("failed to connect to MySQL server: %v", err)
		return err
	}

	// Создаем базу данных, если ее нет
	gormDb, err := createDatabaseIfNotExistsAndConnect(db, s.config.Database)
	if err != nil {
		s.logger.Errorf("failed to connect to database: %v", err)
		return err
	}

	s.db = gormDb
	return nil
}

// createDatabaseIfNotExistsAndConnect Создает базу если нет в сервере базы
func createDatabaseIfNotExistsAndConnect(db *gorm.DB, config DatabaseConfig) (*gorm.DB, error) {
	// Проверяем, существует ли уже база данных
	var result sql.NullString
	err := db.Raw("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", config.DatabaseName).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	// Закрываем подключение для проверки существования базы
	sqlDb, _ := db.DB()
	defer func(sqlDb *sql.DB) {
		err := sqlDb.Close()
		if err != nil {
			panic(fmt.Sprintf("failed to close DB: %v", err))
		}
	}(sqlDb)

	// Тут два сценария:
	// 1) Если есть база (Valid) то подключаемся и возвращаем gormDb
	// 2) Если нет, то создаем базу и таблицу, подключаемся к этой базе и вернем подключение к базе
	if result.Valid {
		// Если база существует подключемся к нему
		connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.DatabaseName,
		)
		gormDB, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
		if err != nil {
			return nil, err
		}

		return gormDB, nil // База данных уже существует
	} else {
		// Тут происходит если нет базы данных и нужно создать и вернуть
		err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", config.DatabaseName)).Error
		if err != nil {
			return nil, err
		}

		// Если база существует подключемся к нему
		connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.DatabaseName,
		)
		gormDb, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
		if err != nil {
			return nil, err
		}

		// Создаем необходимые таблицы
		err = gormDb.AutoMigrate(
			models.UserModel{},
			models.RefreshSession{},
			models.Petition{},
			models.Comment{},
			models.Vote{},
		)
		if err != nil {
			return nil, err
		}

		// Создать уникальный индекс чтобы не было дважды голосовать в одну петицию
		err = gormDb.Exec("CREATE UNIQUE INDEX idx_user_petition ON votes(user_id, petition_id)").Error
		if err != nil {
			return nil, err
		}

		return gormDb, nil
	}
}
