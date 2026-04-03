package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type MarketPrice struct {
	ID          int     `json:"id"`
	CropName    string  `json:"cropName"`
	Market      string  `json:"market"`
	PricePerQtl float64 `json:"pricePerQtl"`
	Date        string  `json:"date"`
}

var marketPrices = []MarketPrice{
	{ID: 1, CropName: "Rice", Market: "Pune APMC", PricePerQtl: 2100.00, Date: "2026-04-03"},
	{ID: 2, CropName: "Wheat", Market: "Nashik APMC", PricePerQtl: 2400.00, Date: "2026-04-03"},
	{ID: 3, CropName: "Soybean", Market: "Aurangabad APMC", PricePerQtl: 4500.00, Date: "2026-04-03"},
}

func GetMarket(c *gin.Context) {
	c.JSON(http.StatusOK, marketPrices)
}
