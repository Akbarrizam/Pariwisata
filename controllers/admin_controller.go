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

// INI FUNGSI YANG TADI ERROR (HILANG)
func Logout(c *gin.Context) {
	c.SetCookie("user_id", "", -1, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/login")
}

// --- Dashboard ---

func Dashboard(c *gin.Context) {
    // ... (Kode query destinations, categories, members tetap sama) ...
	var destinations []models.Destination
	config.DB.Preload("Category").Find(&destinations)
	
	var categories []models.Category
	config.DB.Find(&categories)

	var members []models.User
	config.DB.Where("role = ?", "user").Order("created_at desc").Find(&members)

    // ... (Hitungan Revenue tetap sama) ...
    var totalRevenue float64
	config.DB.Model(&models.Transaction{}).Where("status = ?", "paid").Select("sum(total_amount)").Scan(&totalRevenue)
	visitorStats := []int{120, 150, 180, 220, 300, 450, 400, 380, 420, 500, 550, 600}


    // --- TAMBAHAN BARU: AMBIL DATA TRANSAKSI ---
    var transactions []models.Transaction
    // Preload User & Destination agar nama muncul, urutkan dari terbaru
    config.DB.Preload("User").Preload("Destination").Order("updated_at desc").Limit(10).Find(&transactions)
    // -------------------------------------------

	c.HTML(http.StatusOK, "admin_dashboard.html", gin.H{
		"destinations":  destinations,
		"categories":    categories,
		"members":       members,
		"totalRevenue":  totalRevenue,
		"visitorStats":  visitorStats,
        "transactions":  transactions, // Kirim ke HTML
	})
}

// --- CRUD Destinasi ---

func CreateDestination(c *gin.Context) {
	// 1. Ambil data form text
	name := c.PostForm("name")
	desc := c.PostForm("description")
	loc := c.PostForm("location")
	catID, _ := strconv.Atoi(c.PostForm("category_id"))
	price, _ := strconv.ParseFloat(c.PostForm("price"), 64)
	
	lat, _ := strconv.ParseFloat(c.PostForm("latitude"), 64)
	long, _ := strconv.ParseFloat(c.PostForm("longitude"), 64)

	// 2. Simpan Data Utama Destinasi Dulu
	dest := models.Destination{
		Name:        name, 
		Description: desc, 
		Location:    loc, 
		CategoryID:  uint(catID), 
		Price:       price,
		Latitude:    lat,
		Longitude:   long,
	}
	
	// Simpan ke DB untuk mendapatkan ID Destinasi
	if err := config.DB.Create(&dest).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal simpan data: "+err.Error())
		return
	}

	// 3. --- LOGIC BARU: MULTI UPLOAD ---
	// Ambil Multipart Form
	form, err := c.MultipartForm()
	if err == nil {
		// Ambil semua file dari input dengan name="images"
		files := form.File["images"]

		// Looping setiap file yang diupload
		for _, file := range files {
			// Buat nama file unik (tambah timestamp biar gak bentrok)
			filename := filepath.Base(file.Filename)
			dst := "./uploads/" + filename

			// Simpan fisik file ke folder
			if err := c.SaveUploadedFile(file, dst); err == nil {
				// Simpan info file ke Tabel Galleries
				gallery := models.Gallery{
					DestinationID: dest.ID, // Pakai ID dari destinasi yang baru dibuat
					ImageURL:      filename,
				}
				config.DB.Create(&gallery)
			}
		}
	}

	c.Redirect(http.StatusFound, "/admin/dashboard")
}