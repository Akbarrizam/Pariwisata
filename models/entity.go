package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique;size:191" json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`      // 'admin' atau 'user'
	FullName string `json:"full_name"` // Nama Lengkap Member
}

type Category struct {
	gorm.Model
	Name string `json:"name"`
}

type Destination struct {
	gorm.Model
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Price       float64   `json:"price"`
	CategoryID  uint      `json:"category_id"`
	
	// Koordinat Peta
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`

	Category    Category  `json:"category"`
	Galleries   []Gallery `json:"galleries"`
	Reviews     []Review  `json:"reviews"`
}

type Gallery struct {
	gorm.Model
	DestinationID uint   `json:"destination_id"`
	ImageURL      string `json:"image_url"`
}

type Review struct {
	gorm.Model
	DestinationID uint   `json:"destination_id"`
	VisitorName   string `json:"visitor_name"`
	Comment       string `json:"comment"`
	Rating        int    `json:"rating"`
}

// --- INI YANG TADI HILANG (TRANSACTION) ---
type Transaction struct {
	gorm.Model
	UserID        uint
	DestinationID uint
	TotalAmount   float64
	Status        string // 'paid', 'used', 'pending'
	BookingCode   string `gorm:"unique"` // Kode Unik Tiket

	User          User
	Destination   Destination
}