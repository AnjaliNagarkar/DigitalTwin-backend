package main

import (
	"net/http"

	"DT-backend/db"
	"DT-backend/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	conn := db.Connect()
	defer conn.Close()

	cropHandler := &handlers.CropHandler{DB: conn}
	landHandler := &handlers.LandHandler{DB: conn}
	irrigationHandler := &handlers.IrrigationHandler{DB: conn}
	farmerHandler := &handlers.FarmerHandler{DB: conn}

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Crops — FAMILY table (kharif/rabi cultivation)
	r.GET("/crops", cropHandler.GetCrops)

	// Land — FAMILY table (land area)
	r.GET("/land", landHandler.GetLand)

	// Irrigation — FAMILY table (water source, pump)
	r.GET("/irrigation", irrigationHandler.GetIrrigation)

	// Farmers — FAMILY JOIN FAMILY_MEMBER
	r.GET("/farmers", farmerHandler.GetFarmers)

	// Remaining in-memory modules
	r.GET("/soil", handlers.GetSoil)
	r.GET("/schemes", handlers.GetSchemes)
	r.GET("/market", handlers.GetMarket)

	r.Run(":8080")
}
