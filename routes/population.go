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
}
