package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Scheme struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Benefit     string `json:"benefit"`
	Eligibility string `json:"eligibility"`
}

var schemes = []Scheme{
	{
		ID:          1,
		Name:        "PM-KISAN",
		Description: "Direct income support to farmers",
		Benefit:     "₹6000 per year",
		Eligibility: "Small and marginal farmers",
	},
	{
		ID:          2,
		Name:        "Pradhan Mantri Fasal Bima Yojana",
		Description: "Crop insurance scheme",
		Benefit:     "Insurance coverage for crop loss",
		Eligibility: "All farmers growing notified crops",
	},
	{
		ID:          3,
		Name:        "Kisan Credit Card",
		Description: "Credit facility for agricultural needs",
		Benefit:     "Low interest credit up to ₹3 lakh",
		Eligibility: "Farmers, sharecroppers, and tenant farmers",
	},
}

func GetSchemes(c *gin.Context) {
	c.JSON(http.StatusOK, schemes)
}
