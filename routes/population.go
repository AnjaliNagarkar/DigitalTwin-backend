package routes

import (
	"DT-backend/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterPopulationRoutes(r *gin.Engine, populationHandler *handlers.PopulationHandler) {
	r.GET("/population/dashboard", populationHandler.GetPopulationDashboard)
	r.GET("/population/demographics", populationHandler.GetPopulationDemographics)
	r.GET("/population/education", populationHandler.GetPopulationEducation)
	r.GET("/population/employment", populationHandler.GetPopulationEmployment)
	r.GET("/population/registry", populationHandler.GetPopulationRegistry)
	r.GET("/population/map-summary", populationHandler.GetPopulationMapSummary)
	r.GET("/population/map-data", populationHandler.GetPopulationMapData)
	r.GET("/population/map-insights", populationHandler.GetPopulationMapInsights)
}
