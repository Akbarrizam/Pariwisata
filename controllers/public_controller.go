package controllers

import (
	"net/http"
	"pariwisata/config"
	"pariwisata/models"
	"strconv" // Wajib ada untuk cookie

	"github.com/gin-gonic/gin"
)

// --- HALAMAN UTAMA (HOME) ---
func Home(c *gin.Context) {
	var destinations []models.Destination
	var categories []models.Category
	var user models.User // Variabel untuk menampung data user

	// 1. LOGIKA BARU: Cek apakah user sedang login?
	if cookie, err := c.Cookie("member_id"); err == nil {
		// Jika ada cookie, cari datanya di database
		config.DB.First(&user, cookie)
	}

	// 2. Logika Search & Filter (Tetap sama seperti sebelumnya)
	keyword := c.Query("keyword")
	catId := c.Query("category_id")

	query := config.DB.Preload("Category").Preload("Galleries")

	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	if catId != "" {
		query = query.Where("category_id = ?", catId)
	}

	query.Find(&destinations)
	config.DB.Find(&categories)

	// 3. Kirim data "user" ke HTML
	c.HTML(http.StatusOK, "index.html", gin.H{
		"destinations": destinations,
		"categories":   categories,
		"user":         user, // <--- INI PENTING: Data user dikirim ke sini
	})
}

// --- DETAIL WISATA ---
func DetailDestination(c *gin.Context) {
	id := c.Param("id")
	var dest models.Destination

	if err := config.DB.Preload("Category").Preload("Galleries").Preload("Reviews").First(&dest, id).Error; err != nil {
		c.String(http.StatusNotFound, "Wisata tidak ditemukan")
		return
	}

	c.HTML(http.StatusOK, "detail.html", gin.H{
		"destination": dest,
	})
}

// --- POST REVIEW ---
func PostReview(c *gin.Context) {
	var review models.Review
	destID := c.PostForm("destination_id")

	review.VisitorName = c.PostForm("visitor_name")
	review.Comment = c.PostForm("comment")
	rating, _ := strconv.Atoi(c.PostForm("rating"))

	// Convert string destID ke uint
	destIDUint, _ := strconv.ParseUint(destID, 10, 32)

	// Simpan review
	config.DB.Create(&models.Review{
		DestinationID: uint(destIDUint),
		VisitorName:   review.VisitorName,
		Comment:       review.Comment,
		Rating:        rating,
	})

	c.Redirect(http.StatusFound, "/destination/"+destID)
}

// ==========================================
// BAGIAN INI YANG TADI HILANG/UNDEFINED
// ==========================================

// --- REGISTRASI MEMBER ---
func ShowRegister(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", nil)
}

func ProcessRegister(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	fullName := c.PostForm("fullname")

	// Cek username
	var existingUser models.User
	if err := config.DB.Where("username = ?", username).First(&existingUser).Error; err == nil {
		c.HTML(http.StatusOK, "register.html", gin.H{"error": "Username sudah dipakai!"})
		return
	}

	// Buat User Baru
	newUser := models.User{
		Username: username,
		Password: password,
		FullName: fullName,
		Role:     "user",
	}
	config.DB.Create(&newUser)

	c.Redirect(http.StatusFound, "/login-member")
}

// --- LOGIN MEMBER ---
func ShowMemberLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login_member.html", nil)
}

func ProcessMemberLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	var user models.User
	// Cari user dengan role 'user'
	err := config.DB.Where("username = ? AND role = 'user'", username).First(&user).Error

	if err != nil || user.Password != password {
		c.HTML(http.StatusOK, "login_member.html", gin.H{"error": "Username atau Password salah!"})
		return
	}

	// Set Cookie Login
	c.SetCookie("member_id", strconv.Itoa(int(user.ID)), 3600, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/user/dashboard")
}

// --- LOGOUT MEMBER ---
// --- LOGOUT MEMBER ---
func MemberLogout(c *gin.Context) {
	// Hapus Cookie member
	c.SetCookie("member_id", "", -1, "/", "localhost", false, true)
	
	// Redirect ke halaman LOGIN MEMBER (Sesuai request Anda)
	c.Redirect(http.StatusFound, "/login-member")
}

// --- DASHBOARD USER ---
func UserDashboard(c *gin.Context) {
	// 1. Ambil User dari Cookie
	userIDStr, err := c.Cookie("member_id")
	if err != nil {
		c.Redirect(http.StatusFound, "/login-member")
		return
	}

	var user models.User
	if err := config.DB.First(&user, userIDStr).Error; err != nil {
		c.SetCookie("member_id", "", -1, "/", "localhost", false, true)
		c.Redirect(http.StatusFound, "/login-member")
		return
	}

	// 2. Ambil Tiket Aktif
	var tickets []models.Transaction
	config.DB.Preload("Destination").Where("user_id = ? AND status = ?", user.ID, "paid").Find(&tickets)

	// 3. AMBIL DESTINASI (DISINI PERBAIKANNYA)
	var destinations []models.Destination
	
	// TAMBAHKAN .Preload("Galleries") AGAR FOTO IKUT DIAMBIL
	config.DB.Preload("Category").Preload("Galleries").Find(&destinations) 

	c.HTML(http.StatusOK, "user_dashboard.html", gin.H{
		"user":         user,
		"tickets":      tickets,
		"destinations": destinations,
	})
}