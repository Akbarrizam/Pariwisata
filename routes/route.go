package routes

import (
	"net/http"
	"pariwisata/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Setup HTML Templates
	r.LoadHTMLGlob("templates/*")
	
	// Setup Static Files (CSS/Images)
	r.Static("/uploads", "./uploads")

	// --- PUBLIC ROUTES (Siapapun bisa akses) ---
	r.GET("/", controllers.Home)
	r.GET("/destination/:id", controllers.DetailDestination)
	r.POST("/review", controllers.PostReview)
	
	// --- AUTH MEMBER ROUTES (Login/Register) ---
	r.GET("/register", controllers.ShowRegister)
	r.POST("/register", controllers.ProcessRegister)
	r.GET("/login-member", controllers.ShowMemberLogin)
	r.POST("/login-member", controllers.ProcessMemberLogin)
	r.GET("/logout-member", controllers.MemberLogout)

	// --- MEMBER AREA (Harus Login) ---
	r.GET("/user/dashboard", controllers.UserDashboard)
	
	// >>> FITUR TIKET & PEMBAYARAN <<<
	r.POST("/checkout", controllers.ShowCheckout)          // Halaman Bayar
	r.POST("/process-payment", controllers.ProcessPayment) // Proses Bayar
	r.GET("/user/tickets", controllers.MyTickets)          // Tiket Saya (INI YANG TADI MISSING)
	// --------------------------------

	// Helper untuk QR Code
	r.GET("/qrcode", controllers.GenerateQRCode) 

	// --- ADMIN AUTH ---
	r.GET("/login", controllers.ShowLogin)
	r.POST("/login", controllers.ProcessLogin)
	r.GET("/logout", controllers.Logout)

	// --- ADMIN DASHBOARD (Protected) ---
	admin := r.Group("/admin")
	admin.Use(AuthMiddleware())
	{
		// 1. Dashboard Utama
		admin.GET("/dashboard", controllers.Dashboard)
		
		// 2. Create (Simpan Wisata & Kuliner Baru)
		// Pastikan baris ini ada dan tidak error
		admin.POST("/destination", controllers.CreateDestination) 

		// 3. Fitur Scanner
		admin.GET("/scanner", controllers.ShowScanner)
		admin.POST("/validate-ticket", controllers.ApiValidateTicket)

		// 4. Fitur Edit & Update
		admin.GET("/destination/edit/:id", controllers.ShowEditDestination)
		admin.POST("/destination/update/:id", controllers.UpdateDestination)

		// 5. Fitur Delete
		admin.GET("/destination/delete/:id", controllers.DeleteDestination)
	}
	// API Routes (JSON)
	api := r.Group("/api/v1")
	{
		api.GET("/destinations", controllers.ApiGetDestinations)
	}

	return r
}

// Middleware Sederhana Admin
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := c.Cookie("user_id")
		if err != nil {
			c.Redirect(http.StatusFound, "/login") 
			c.Abort()
			return
		}
		c.Next()
	}
}