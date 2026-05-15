package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"DT-backend/db"
	"DT-backend/handlers"
	"DT-backend/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	serverAddr := ":8081"
	log.Printf("[STARTUP] Agriculture Digital Twin backend starting on %s", serverAddr)
	log.Println("[STARTUP] Mode: READ-ONLY — all endpoints use SELECT queries only")

	conn := db.Connect()
	defer conn.Close()

	// Detect optional columns in FAMILY table (SELECT on INFORMATION_SCHEMA — read-only)
	cc := handlers.NewColumnChecker(conn)

	cropHandler := &handlers.CropHandler{DB: conn}
	landHandler := &handlers.LandHandler{DB: conn}
	irrigationHandler := &handlers.IrrigationHandler{DB: conn}
	farmerHandler := &handlers.FarmerHandler{DB: conn}
	houseHandler := &handlers.HouseHandler{DB: conn, CC: cc}
	if err := houseHandler.PreloadHouseCache(); err != nil {
		log.Fatalf("[FATAL] failed to preload house cache: %v", err)
	}
	insightHandler := &handlers.InsightHandler{DB: conn, CC: cc}
	dashboardSummaryHandler := handlers.NewDashboardSummaryHandler(conn, cc)
	locationHandler := &handlers.LocationHandler{DB: conn}
	districtSurveyCountHandler := &handlers.DistrictSurveyCountHandler{DB: conn}
	districtCentroidsHandler := &handlers.DistrictCentroidsHandler{DB: conn}
	pdfHandler := &handlers.PDFHandler{DB: conn, CC: cc}
	pdfTwinHandler := &handlers.TwinPDFHandler{DB: conn, CC: cc}
	populationHandler := &handlers.PopulationHandler{DB: conn}
	unifiedRegistryHandler := handlers.NewUnifiedRegistryHandler(conn)
	schemeRecommendHandler := handlers.NewSchemeRecommendHandler(conn)
	advisoryHandler := handlers.NewAdvisoryHandler(conn)
	clusterAdvisoryHandler := handlers.NewClusterAdvisoryHandler(conn)
	viewOptionsHandler := handlers.NewViewOptionsHandler(conn, cc)

	authHandler := handlers.NewAuthHandler()

	r := gin.New()
	r.Use(gin.Recovery())

	// ── Request logger ───────────────────────────────────────────────────────
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("[REQUEST] %s %s → %d (%s)",
			c.Request.Method, c.Request.URL.RequestURI(),
			c.Writer.Status(), time.Since(start))
	})

	// ── CORS ─────────────────────────────────────────────────────────────────
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// ── Read-only enforcement ─────────────────────────────────────────────────
	// Allow POST only for: PDF generation, auth login/logout.
	// Everything else must be GET.
	r.Use(func(c *gin.Context) {
		method := strings.ToUpper(c.Request.Method)
		path := c.Request.URL.Path
		if method == http.MethodPost && (strings.HasPrefix(path, "/pdf/") ||
			path == "/auth/login" || path == "/auth/logout") {
			c.Next()
			return
		}
		if method != http.MethodGet && method != http.MethodOptions {
			log.Printf("[BLOCKED] %s %s — rejected (read-only mode)", c.Request.Method, path)
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{
				"error": "this server is read-only — only GET requests are permitted",
			})
			return
		}
		c.Next()
	})

	// ── Health check (public) ─────────────────────────────────────────────────
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong", "mode": "read-only"})
	})
	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong", "mode": "read-only"})
	})

	// ── Auth endpoints (public — no login required) ───────────────────────────
	r.POST("/auth/login",       authHandler.Login)
	r.POST("/auth/logout",      authHandler.Logout)
	r.GET("/auth/sso-verify",   authHandler.SSOVerify) // SSO: validates an IVDP bearer token

	// ── All data routes require a valid session ───────────────────────────────
	protected := r.Group("/")
	protected.Use(handlers.AuthMiddleware())
	protected.GET("/auth/me", authHandler.Me)

	// ── Digital Twin APIs (new) ───────────────────────────────────────────────
	protected.GET("/view-options", viewOptionsHandler.GetViewOptions)
	protected.GET("/houses", houseHandler.GetHouses)
	protected.GET("/houses/map-points", houseHandler.GetHousesMapPoints)
	protected.GET("/houses/summary", houseHandler.GetHousesSummary)
	protected.GET("/houses/batch-members", houseHandler.GetBatchMemberStats)
	protected.GET("/house/:id", houseHandler.GetHouseByID)
	protected.GET("/districts", locationHandler.GetDistricts)
	protected.GET("/location-options", locationHandler.GetLocationOptions)
	protected.GET("/map/district-survey-counts", districtSurveyCountHandler.GetDistrictSurveyCounts)
	protected.GET("/map/district-centroids", districtCentroidsHandler.GetDistrictCentroids)

	// ── PDF report (POST — reads DB, streams PDF; no DB writes) ──────────────
	protected.POST("/pdf/report", pdfHandler.GeneratePDF)
	protected.POST("/pdf/population-report", populationHandler.GeneratePopulationPDF)

	// 3D Twin summary PDF (GET with query params) — streams generated PDF
	protected.GET("/twin/export-pdf", pdfTwinHandler.GenerateTwinPDF)

	// ── Insights Engine (new) ─────────────────────────────────────────────────
	protected.GET("/insights/governance", insightHandler.GetGovernanceInsights)
	protected.GET("/insights/agriculture", insightHandler.GetAgricultureInsights)
	protected.GET("/insights/welfare", insightHandler.GetWelfareInsights)
	protected.GET("/dashboard/summary", dashboardSummaryHandler.GetDashboardSummary)
	routes.RegisterPopulationRoutes(protected, populationHandler)

	// ── Existing agri data endpoints ──────────────────────────────────────────
	protected.GET("/crops", cropHandler.GetCrops)
	protected.GET("/land", landHandler.GetLand)
	protected.GET("/irrigation", irrigationHandler.GetIrrigation)
	protected.GET("/citizens", farmerHandler.GetFarmers)
	protected.GET("/farmers", farmerHandler.GetFarmers)
	protected.GET("/unified-registry", unifiedRegistryHandler.GetUnifiedRegistry)

	// ── Static/in-memory modules ──────────────────────────────────────────────
	protected.GET("/soil", handlers.GetSoil)
	protected.GET("/schemes", handlers.GetSchemes)
	protected.GET("/schemes/recommend", schemeRecommendHandler.GetRecommendations)
	protected.GET("/advisory", advisoryHandler.GetAdvisory)
	protected.GET("/advisory/cluster", clusterAdvisoryHandler.GetClusterAdvisory)
	protected.GET("/market", handlers.GetMarket)

	log.Println("[STARTUP] Routes registered:")
	log.Println("  GET /view-options        — dynamic VIEW BY dropdown options (schema-driven)")
	log.Println("  GET /ping")
	log.Println("  GET /houses           — geo-mapped household data (2D map + 3D twin)")
	log.Println("  GET /houses/map-points — lightweight map coordinates for client clustering")
	log.Println("  GET /houses/summary   — grid-aggregated household counts (viewport clusters)")
	log.Println("  GET /house/:id        — single household detail + family members")
	log.Println("  GET /insights/governance  — sanitation, lighting, geo coverage")
	log.Println("  GET /insights/agriculture — land distribution, irrigation, crops")
	log.Println("  GET /insights/welfare     — BPL households, ration card data")
	log.Println("  GET /population/dashboard — population top-card metrics")
	log.Println("  GET /crops            — kharif/rabi cultivation data")
	log.Println("  GET /land             — land area records")
	log.Println("  GET /irrigation       — water source records")
	log.Println("  GET /citizens         — citizen registry (legacy)")
	log.Println("  GET /farmers          — farmer registry (legacy)")
	log.Println("  GET /unified-registry — unified master registry (merged population + agri)")
	log.Println("  GET /soil /schemes /market  — static reference data")
	log.Println("  POST /pdf/report            — generate PDF report (DB read-only)")
	log.Println("  POST /pdf/population-report — generate population PDF report (DB read-only)")

	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("[FATAL] Server failed to start on %s: %v", serverAddr, err)
	}
}
