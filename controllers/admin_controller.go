package controllers

import (
	"net/http"
	"path/filepath"
	"pariwisata/config"
	"pariwisata/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// --- Auth ---

func ShowLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func ProcessLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	var user models.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{"error": "User tidak ditemukan"})
		return
	}

	// Cek Password Biasa (Tanpa Bcrypt)
	if user.Password != password {
		c.HTML(http.StatusOK, "login.html", gin.H{"error": "Password salah"})
		return
	}

	// Set Cookie Sederhana
	c.SetCookie("user_id", strconv.Itoa(int(user.ID)), 3600, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/admin/dashboard")
}

func Logout(c *gin.Context) {
	c.SetCookie("user_id", "", -1, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/login")
}

// --- Dashboard ---

func Dashboard(c *gin.Context) {
	var destinations []models.Destination
	config.DB.Preload("Category").Find(&destinations)
	
	var categories []models.Category
	config.DB.Find(&categories)

	var members []models.User
	config.DB.Where("role = ?", "user").Order("created_at desc").Find(&members)

	var totalRevenue float64
	config.DB.Model(&models.Transaction{}).Where("status = ?", "paid").Select("sum(total_amount)").Scan(&totalRevenue)
	
    // Data dummy untuk chart (bisa diganti real data nanti)
    visitorStats := []int{120, 150, 180, 220, 300, 450, 400, 380, 420, 500, 550, 600}

	// Ambil Data Transaksi Terbaru
	var transactions []models.Transaction
	config.DB.Preload("User").Preload("Destination").Order("updated_at desc").Limit(10).Find(&transactions)

	c.HTML(http.StatusOK, "admin_dashboard.html", gin.H{
		"destinations": destinations,
		"categories":   categories,
		"members":      members,
		"totalRevenue": totalRevenue,
		"visitorStats": visitorStats,
		"transactions": transactions,
	})
}

// --- CRUD DESTINASI ---

func ShowEditDestination(c *gin.Context) {
	id := c.Param("id")

	var dest models.Destination
	if err := config.DB.Preload("Category").First(&dest, id).Error; err != nil {
		c.String(http.StatusNotFound, "Data tidak ditemukan")
		return
	}

	var categories []models.Category
	config.DB.Find(&categories)

	c.HTML(http.StatusOK, "admin_edit.html", gin.H{
		"destination": dest,
		"categories":  categories,
	})
}

// --- CREATE / TAMBAH DESTINASI BARU ---
func CreateDestination(c *gin.Context) {
	// 1. Ambil data form text
	name := c.PostForm("name")
	desc := c.PostForm("description")
	loc := c.PostForm("location")
	catID, _ := strconv.Atoi(c.PostForm("category_id"))
	price, _ := strconv.ParseFloat(c.PostForm("price"), 64)
	
	lat, _ := strconv.ParseFloat(c.PostForm("latitude"), 64)
	long, _ := strconv.ParseFloat(c.PostForm("longitude"), 64)

	// Data Kuliner (Jam Buka & Menu)
	openHours := c.PostForm("opening_hours")
	menu := c.PostForm("highlight_menu")

	// 2. Simpan Data ke Struct Destination
	dest := models.Destination{
		Name:          name, 
		Description:   desc, 
		Location:      loc, 
		CategoryID:    uint(catID), 
		Price:         price,
		Latitude:      lat,
		Longitude:     long,
		OpeningHours:  openHours, 
		HighlightMenu: menu,      
	}
	
	// Simpan ke DB
	if err := config.DB.Create(&dest).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal simpan data: "+err.Error())
		return
	}

	// 3. Logic Multi Upload Foto
	form, err := c.MultipartForm()
	if err == nil {
		files := form.File["images"]
		for _, file := range files {
			filename := filepath.Base(file.Filename)
			dst := "./uploads/" + filename

			if err := c.SaveUploadedFile(file, dst); err == nil {
				gallery := models.Gallery{
					DestinationID: dest.ID,
					ImageURL:      filename,
				}
				config.DB.Create(&gallery)
			}
		}
	}

    // --- PERBAIKAN: TAMBAHKAN INI DI BAGIAN PALING BAWAH ---
    // Agar setelah simpan, browser kembali ke Dashboard dan muncul Pop Up
	c.Redirect(http.StatusFound, "/admin/dashboard?status=success")
}

func UpdateDestination(c *gin.Context) {
	id := c.Param("id")
	
	var dest models.Destination
	if err := config.DB.First(&dest, id).Error; err != nil {
		c.String(http.StatusNotFound, "Data tidak ditemukan")
		return
	}

	// Update Fields
	dest.Name = c.PostForm("name")
	dest.Description = c.PostForm("description")
	dest.Location = c.PostForm("location")
	dest.OpeningHours = c.PostForm("opening_hours")
	dest.HighlightMenu = c.PostForm("highlight_menu")
	
	catID, _ := strconv.Atoi(c.PostForm("category_id"))
	dest.CategoryID = uint(catID)
	
	dest.Price, _ = strconv.ParseFloat(c.PostForm("price"), 64)
	dest.Latitude, _ = strconv.ParseFloat(c.PostForm("latitude"), 64)
	dest.Longitude, _ = strconv.ParseFloat(c.PostForm("longitude"), 64)

	// Update Image (Single for simplicity in Edit)
	file, err := c.FormFile("image")
	if err == nil {
		filename := filepath.Base(file.Filename)
		dst := "./uploads/" + filename
		if err := c.SaveUploadedFile(file, dst); err == nil {
            // Logic update gallery (optional/simplified)
		}
	}

	config.DB.Save(&dest)
    
    // --- TAMBAHAN: Status Updated ---
	c.Redirect(http.StatusFound, "/admin/dashboard?status=updated")
}

func DeleteDestination(c *gin.Context) {
	id := c.Param("id")
	
	config.DB.Delete(&models.Destination{}, id)

    // --- TAMBAHAN: Status Deleted ---
	c.Redirect(http.StatusFound, "/admin/dashboard?status=deleted")
}