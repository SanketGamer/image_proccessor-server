package config

import (
	"fmt"
	"image_service/models"
	"log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)


func ConnectDB(cfg *Config) *gorm.DB{
	dsn:=fmt.Sprintf(    "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	
	  db,err:=gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), //print sql queries in terminal
		DisableForeignKeyConstraintWhenMigrating: true, 
	  })
	  if err!=nil{
	 	log.Fatalf("failed to connect to database: %v",err)
      }
	
	// Migrate User FIRST — images depends on users
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("User migration failed: %v", err)
	}

	// THEN migrate Image
	if err := db.AutoMigrate(&models.Image{}); err != nil {
		log.Fatalf("Image migration failed: %v", err)
	}
	 log.Println("Database connected and migrated")
	 return db
}