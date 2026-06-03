package sqlinit

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Charset  string
}

func NewConfigFromEnv() DBConfig {
	log.Println("[DB][S1] load env files")
	// Try to load .env from current directory or parent directory
	envFiles := []string{".env", "../.env"}
	for _, f := range envFiles {
		if err := godotenv.Load(f); err == nil {
			log.Printf("[DB][S2] loaded environment from %s", f)
		}
	}

	host := getEnv("MYSQL_HOST", "127.0.0.1")
	port := getEnv("MYSQL_PORT", "3306")
	user := getEnv("MYSQL_USER", "root")
	password := getEnv("MYSQL_PASSWORD", "422624")
	database := getEnv("MYSQL_DATABASE", "ai_agent")
	charset := getEnv("MYSQL_CHARSET", "utf8mb4")

	// Critical check: if port contains "tcp/", strip it.
	// This fixes the "dial tcp: lookup tcp/3306" error.
	if len(port) > 4 && port[:4] == "tcp/" {
		port = port[4:]
	}
	// If host contains "tcp/", strip it too
	if len(host) > 4 && host[:4] == "tcp/" {
		host = host[4:]
	}

	log.Printf("[DB][S3] config host=%s port=%s user=%s database=%s", host, port, user, database)

	return DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
		Charset:  charset,
	}
}

func InitDB(config DBConfig) error {
	log.Println("[DB][S4] ensure database exists")
	err := ensureDatabase(config)
	if err != nil {
		log.Printf("[DB][E_DB_ENSURE_DATABASE] ensure database failed: %v", err)
		return err
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Charset,
	)
	log.Println("[DB][S5] open gorm connection")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("[DB][E_DB_GORM_OPEN] gorm open failed: %v", err)
		return err
	}

	log.Println("[DB][S6] auto migrate models")
	err = db.AutoMigrate(
		&User{},
		&UserProfile{},
		&PlanRecord{},
		&DailyTaskStatus{},
		&ChatMessageRecord{},
		&UserWeightParam{},
		&UserWeightRecord{},
		&UserMemoryRecord{},
		&UserAttendanceRecord{},
		&UserAttendanceWeek{},
		&Product{},
	)
	if err != nil {
		log.Printf("[DB][E_DB_MIGRATE] auto migrate failed: %v", err)
		return err
	}

	DB = db
	log.Println("[DB][S7] db init complete")
	return nil
}

func ensureDatabase(config DBConfig) error {
	log.Println("[DB][S8] open sql connection for ensure database")
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/?charset=%s&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Charset,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("[DB][E_DB_SQL_OPEN] sql open failed: %v", err)
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Printf("[DB][E_DB_SQL_PING] sql ping failed: %v", err)
		return err
	}

	log.Printf("[DB][S9] create database if not exists: %s", config.Database)
	_, err = db.Exec(
		fmt.Sprintf(
			"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET %s",
			config.Database,
			config.Charset,
		),
	)
	if err != nil {
		log.Printf("[DB][E_DB_CREATE_DATABASE] create database failed: %v", err)
		return err
	}
	log.Println("[DB][S10] ensure database complete")
	return err
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
