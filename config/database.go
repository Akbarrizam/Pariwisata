package config

import (
	"fmt"
	"pariwisata/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	// Ganti user:password sesuai setting MySQL lokal Anda
	dsn := "root:@tcp(127.0.0.1:3306)/db_pariwisata?charset=utf8mb4&parseTime=True&loc=Local"
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("Gagal koneksi ke database!")
	}

	// Auto Migrate untuk memastikan struct sesuai tabel
	database.AutoMigrate(&models.User{}, &models.Category{}, &models.Destination{}, &models.Gallery{}, &models.Review{}, &models.Transaction{})

	DB = database
	fmt.Println("Database terhubung!")
}