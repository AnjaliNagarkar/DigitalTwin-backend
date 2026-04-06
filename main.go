package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"DT-backend/db"
	"DT-backend/handlers"

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
	locationHandler := &handlers.LocationHandler{DB: conn}

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
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// ── Read-only enforcement ─────────────────────────────────────────────────
	// Block any non-GET/OPTIONS request with a 405. This is a hard safeguard
	// in addition to having only SELECT queries in all handlers.
	r.Use(func(c *gin.Context) {
		method := strings.ToUpper(c.Request.Method)
		if method != http.MethodGet && method != http.MethodOptions {
			log.Printf("[BLOCKED] Write attempt: %s %s — rejected (read-only mode)",
				c.Request.Method, c.Request.URL.Path)
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
	r.GET("/house/:id", houseHandler.GetHouseByID)
	r.GET("/location-options", locationHandler.GetLocationOptions)

	// ── Insights Engine (new) ─────────────────────────────────────────────────
	r.GET("/insights/governance", insightHandler.GetGovernanceInsights)
	r.GET("/insights/agriculture", insightHandler.GetAgricultureInsights)
	r.GET("/insights/welfare", insightHandler.GetWelfareInsights)

	// ── Existing agri data endpoints ──────────────────────────────────────────
	r.GET("/crops", cropHandler.GetCrops)
	r.GET("/land", landHandler.GetLand)
	r.GET("/irrigation", irrigationHandler.GetIrrigation)
	r.GET("/farmers", farmerHandler.GetFarmers)

	// ── Static/in-memory modules ──────────────────────────────────────────────
	r.GET("/soil", handlers.GetSoil)
	r.GET("/schemes", handlers.GetSchemes)
	r.GET("/market", handlers.GetMarket)

	log.Println("[STARTUP] Routes registered:")
	log.Println("  GET /ping")
	log.Println("  GET /houses           — geo-mapped household data (2D map + 3D twin)")
	log.Println("  GET /house/:id        — single household detail + family members")
	log.Println("  GET /insights/governance  — sanitation, lighting, geo coverage")
	log.Println("  GET /insights/agriculture — land distribution, irrigation, crops")
	log.Println("  GET /insights/welfare     — BPL households, ration card data")
	log.Println("  GET /crops            — kharif/rabi cultivation data")
	log.Println("  GET /land             — land area records")
	log.Println("  GET /irrigation       — water source records")
	log.Println("  GET /farmers          — farmer registry")
	log.Println("  GET /soil /schemes /market  — static reference data")

	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("[FATAL] Server failed to start on %s: %v", serverAddr, err)
	}
}
