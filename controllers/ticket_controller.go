package controllers

import (
	"fmt"
	"math/rand"
	"net/http"
	"pariwisata/config"
	"pariwisata/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

// 1. TAMPILKAN HALAMAN PEMBAYARAN (CHECKOUT)
func ShowCheckout(c *gin.Context) {
	// Ambil data yang dikirim dari halaman detail
	destID := c.PostForm("destination_id")
	price := c.PostForm("price")
	
	// Kita butuh nama wisata untuk ditampilkan, jadi query dulu
	var dest models.Destination
	config.DB.First(&dest, destID)

	c.HTML(http.StatusOK, "payment.html", gin.H{
		"dest_id":   destID,
		"dest_name": dest.Name,
		"price":     price,
	})
}

// 2. PROSES PEMBAYARAN & BUAT TIKET
func ProcessPayment(c *gin.Context) {
	// Ambil User ID dari Cookie
	userIDStr, err := c.Cookie("member_id")
	if err != nil {
		c.Redirect(http.StatusFound, "/login-member")
		return
	}
	userID, _ := strconv.Atoi(userIDStr)

	// Ambil data dari form payment.html
	destID, _ := strconv.Atoi(c.PostForm("destination_id"))
	price, _ := strconv.ParseFloat(c.PostForm("price"), 64)

	// Generate Kode Unik Tiket
	rand.Seed(time.Now().UnixNano())
	randomCode := rand.Intn(99999)
	bookingCode := fmt.Sprintf("TIX-%d-%d", randomCode, destID)

	// Simpan Transaksi ke Database
	trx := models.Transaction{
		UserID:        uint(userID),
		DestinationID: uint(destID),
		TotalAmount:   price,
		Status:        "paid", // Simulasi: Langsung sukses
		BookingCode:   bookingCode,
	}
	
	if err := config.DB.Create(&trx).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal membuat tiket: "+err.Error())
		return
	}

	// Sukses -> Arahkan ke tiket saya
	c.Redirect(http.StatusFound, "/user/tickets")
}

// 2. HALAMAN TIKET SAYA (User)
func MyTickets(c *gin.Context) {
	userIDStr, _ := c.Cookie("member_id")
	
	var transactions []models.Transaction
	// Ambil tiket milik user ini
	config.DB.Preload("Destination").Where("user_id = ?", userIDStr).Order("created_at desc").Find(&transactions)

	c.HTML(http.StatusOK, "my_tickets.html", gin.H{
		"tickets": transactions,
	})
}

// 3. GENERATOR GAMBAR QR CODE (Dipanggil via <img> tag)
func GenerateQRCode(c *gin.Context) {
	code := c.Query("code") // Ambil parameter ?code=TIX-123...
	
	// Buat QR Code jadi bytes (Format PNG, Size 256x256)
	png, err := qrcode.Encode(code, qrcode.Medium, 256)
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal generate QR")
		return
	}

	// Tampilkan sebagai gambar
	c.Data(http.StatusOK, "image/png", png)
}

// 4. ADMIN: HALAMAN SCANNER
func ShowScanner(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_scanner.html", nil)
}

// 5. ADMIN: PROSES VALIDASI TIKET (API)
func ApiValidateTicket(c *gin.Context) {
	code := c.PostForm("code")

	var trx models.Transaction
	// Cari tiket berdasarkan kode
	if err := config.DB.Preload("User").Preload("Destination").Where("booking_code = ?", code).First(&trx).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Tiket TIDAK Ditemukan!"})
		return
	}

	// Cek Status
	if trx.Status == "used" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Tiket SUDAH DIPAKAI sebelumnya!"})
		return
	}

	// Update jadi Used
	config.DB.Model(&trx).Update("status", "used")

	c.JSON(http.StatusOK, gin.H{
		"status": "success", 
		"message": "Tiket VALID! Silakan Masuk.",
		"data": gin.H{
			"visitor": trx.User.FullName,
			"dest": trx.Destination.Name,
		},
	})
}