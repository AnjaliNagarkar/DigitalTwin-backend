package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Soil struct {
	Type           string `json:"type"`
	FertilityLevel string `json:"fertilityLevel"`
	PHLevel        string `json:"phLevel"`
	Moisture       string `json:"moisture"`
}

var soilData = Soil{
	Type:           "Black Soil",
	FertilityLevel: "High",
	PHLevel:        "6.5",
	Moisture:       "45%",
}

func GetSoil(c *gin.Context) {
	c.JSON(http.StatusOK, soilData)
}
