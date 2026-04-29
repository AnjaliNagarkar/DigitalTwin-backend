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
	insightHandler := &handlers.InsightHandler{DB: conn, CC: cc}
	dashboardSummaryHandler := handlers.NewDashboardSummaryHandler(conn, cc)
	locationHandler := &handlers.LocationHandler{DB: conn}
	districtSurveyCountHandler := &handlers.DistrictSurveyCountHandler{DB: conn}
	districtCentroidsHandler := &handlers.DistrictCentroidsHandler{DB: conn}
	districtPopulationHandler := &handlers.DistrictPopulationHandler{DB: conn}
	districtBplHandler := &handlers.DistrictBplHandler{DB: conn}
	pdfHandler := &handlers.PDFHandler{DB: conn, CC: cc}
	populationHandler := &handlers.PopulationHandler{DB: conn}
	unifiedRegistryHandler := handlers.NewUnifiedRegistryHandler(conn)
	schemeRecommendHandler := handlers.NewSchemeRecommendHandler(conn)
	advisoryHandler := handlers.NewAdvisoryHandler(conn)
	clusterAdvisoryHandler := handlers.NewClusterAdvisoryHandler(conn)

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
		// POST is allowed for PDF generation; all other mutations remain blocked.
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// ── Read-only enforcement ─────────────────────────────────────────────────
	// Block any non-GET/OPTIONS request with a 405, EXCEPT POST /pdf/* which
	// only reads from the DB and streams a generated PDF (no writes at all).
	r.Use(func(c *gin.Context) {
		method := strings.ToUpper(c.Request.Method)
		path := c.Request.URL.Path // reliable in middleware (set by Vite proxy rewrite)
		isPDFPost := method == http.MethodPost && strings.HasPrefix(path, "/pdf/")
		if isPDFPost {
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

	// ── Health check ──────────────────────────────────────────────────────────
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong", "mode": "read-only"})
	})

	// ── Digital Twin APIs (new) ───────────────────────────────────────────────
	r.GET("/houses", houseHandler.GetHouses)
	r.GET("/houses/map-points", houseHandler.GetHousesMapPoints)
	r.GET("/houses/summary", houseHandler.GetHousesSummary)
	r.GET("/house/:id", houseHandler.GetHouseByID)
	r.GET("/districts", locationHandler.GetDistricts)
	r.GET("/location-options", locationHandler.GetLocationOptions)
	r.GET("/map/district-survey-counts", districtSurveyCountHandler.GetDistrictSurveyCounts)
	r.GET("/map/district-centroids", districtCentroidsHandler.GetDistrictCentroids)
	r.GET("/map/district-population", districtPopulationHandler.GetDistrictPopulation)
	r.GET("/map/district-bpl", districtBplHandler.GetDistrictBpl)

	// ── PDF report (POST — reads DB, streams PDF; no DB writes) ──────────────
	r.POST("/pdf/report", pdfHandler.GeneratePDF)
	r.POST("/pdf/population-report", populationHandler.GeneratePopulationPDF)

	// ── Insights Engine (new) ─────────────────────────────────────────────────
	r.GET("/insights/governance", insightHandler.GetGovernanceInsights)
	r.GET("/insights/agriculture", insightHandler.GetAgricultureInsights)
	r.GET("/insights/welfare", insightHandler.GetWelfareInsights)
	r.GET("/dashboard/summary", dashboardSummaryHandler.GetDashboardSummary)
	routes.RegisterPopulationRoutes(r, populationHandler)

	// ── Existing agri data endpoints ──────────────────────────────────────────
	r.GET("/crops", cropHandler.GetCrops)
	r.GET("/land", landHandler.GetLand)
	r.GET("/irrigation", irrigationHandler.GetIrrigation)
	r.GET("/citizens", farmerHandler.GetFarmers)
	r.GET("/farmers", farmerHandler.GetFarmers)
	r.GET("/unified-registry", unifiedRegistryHandler.GetUnifiedRegistry)

	// ── Static/in-memory modules ──────────────────────────────────────────────
	r.GET("/soil", handlers.GetSoil)
	r.GET("/schemes", handlers.GetSchemes)
	r.GET("/schemes/recommend", schemeRecommendHandler.GetRecommendations)
	r.GET("/advisory", advisoryHandler.GetAdvisory)
	r.GET("/advisory/cluster", clusterAdvisoryHandler.GetClusterAdvisory)
	r.GET("/market", handlers.GetMarket)

	log.Println("[STARTUP] Routes registered:")
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
