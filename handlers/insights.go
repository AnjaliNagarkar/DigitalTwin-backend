package handlers

// insights.go — Governance, Agriculture, and Welfare analytics
// ALL queries are SELECT-only. No INSERT/UPDATE/DELETE/DROP/ALTER anywhere.

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type InsightHandler struct {
	DB *sql.DB
	CC *ColumnChecker
}

type CountItem struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// GetGovernanceInsights — GET /insights/governance
// Returns household counts, sanitation coverage, and lighting coverage.
func (h *InsightHandler) GetGovernanceInsights(c *gin.Context) {
	log.Println("[SELECT] GET /insights/governance")
	result := gin.H{}

	var totalHouseholds int
	h.DB.QueryRow("SELECT COUNT(*) FROM FAMILY").Scan(&totalHouseholds)
	result["totalHouseholds"] = totalHouseholds

	var withGeo int
	h.DB.QueryRow(`
		SELECT COUNT(*) FROM FAMILY
		WHERE LATITUDE IS NOT NULL AND LONGITUDE IS NOT NULL
		  AND LATITUDE != 0 AND LONGITUDE != 0
	`).Scan(&withGeo)
	result["householdsWithGeoData"] = withGeo

	if h.CC.Has("TYPE_OF_LATRINE") {
		var noToilet int
		h.DB.QueryRow(`
			SELECT COUNT(*) FROM FAMILY
			WHERE TYPE_OF_LATRINE IS NULL
			   OR TYPE_OF_LATRINE = ''
			   OR TYPE_OF_LATRINE = 'No Latrine'
			   OR TYPE_OF_LATRINE = 'None'
		`).Scan(&noToilet)
		result["householdsWithoutToilet"] = noToilet

		latrineRows, err := h.DB.Query(`
			SELECT COALESCE(TYPE_OF_LATRINE, 'Unknown') as label, COUNT(*) as cnt
			FROM FAMILY GROUP BY TYPE_OF_LATRINE ORDER BY cnt DESC LIMIT 10
		`)
		if err == nil {
			defer latrineRows.Close()
			var items []CountItem
			for latrineRows.Next() {
				var item CountItem
				latrineRows.Scan(&item.Label, &item.Count)
				items = append(items, item)
			}
			result["latrineDistribution"] = items
		}
	} else {
		result["householdsWithoutToilet"] = 0
		result["_note_latrine"] = "TYPE_OF_LATRINE column not present in this database schema"
	}

	if h.CC.Has("SOURCE_OF_LIGHTING") {
		var noElec int
		h.DB.QueryRow(`
			SELECT COUNT(*) FROM FAMILY
			WHERE SOURCE_OF_LIGHTING IS NULL
			   OR SOURCE_OF_LIGHTING = ''
			   OR SOURCE_OF_LIGHTING = 'Kerosene'
			   OR SOURCE_OF_LIGHTING = 'None'
			   OR SOURCE_OF_LIGHTING = 'No Lighting'
		`).Scan(&noElec)
		result["householdsWithoutElectricity"] = noElec

		lightRows, err := h.DB.Query(`
			SELECT COALESCE(SOURCE_OF_LIGHTING, 'Unknown') as label, COUNT(*) as cnt
			FROM FAMILY GROUP BY SOURCE_OF_LIGHTING ORDER BY cnt DESC LIMIT 10
		`)
		if err == nil {
			defer lightRows.Close()
			var items []CountItem
			for lightRows.Next() {
				var item CountItem
				lightRows.Scan(&item.Label, &item.Count)
				items = append(items, item)
			}
			result["lightingDistribution"] = items
		}
	} else {
		result["householdsWithoutElectricity"] = 0
		result["_note_lighting"] = "SOURCE_OF_LIGHTING column not present in this database schema"
	}

	c.JSON(http.StatusOK, result)
}

// GetAgricultureInsights — GET /insights/agriculture
// Returns land distribution, irrigation coverage, and crop season data.
func (h *InsightHandler) GetAgricultureInsights(c *gin.Context) {
	log.Println("[SELECT] GET /insights/agriculture")
	result := gin.H{}

	// ── DB-wide totals (all households, not just GPS-tagged ones) ─────────────
	var totalHouseholds int
	h.DB.QueryRow("SELECT COUNT(*) FROM FAMILY").Scan(&totalHouseholds)
	result["totalHouseholds"] = totalHouseholds

	var totalPopulation int
	h.DB.QueryRow("SELECT COUNT(*) FROM FAMILY_MEMBER").Scan(&totalPopulation)
	result["totalPopulation"] = totalPopulation

	var totalMale int
	h.DB.QueryRow(`SELECT COUNT(*) FROM FAMILY_MEMBER
		WHERE LOWER(TRIM(COALESCE(GENDER,''))) IN ('male','m')`).Scan(&totalMale)
	result["totalMale"] = totalMale

	var totalFemale int
	h.DB.QueryRow(`SELECT COUNT(*) FROM FAMILY_MEMBER
		WHERE LOWER(TRIM(COALESCE(GENDER,''))) IN ('female','f')`).Scan(&totalFemale)
	result["totalFemale"] = totalFemale

	var totalFarmers int
	h.DB.QueryRow("SELECT COUNT(*) FROM FAMILY WHERE OWN_AGRICULTURE_LAND = 'Yes'").Scan(&totalFarmers)
	result["totalFarmers"] = totalFarmers

	var noIrrigation int
	h.DB.QueryRow(`
		SELECT COUNT(*) FROM FAMILY
		WHERE OWN_AGRICULTURE_LAND = 'Yes'
		  AND (SOURCE_WATER_IRRIGATION IS NULL
		    OR SOURCE_WATER_IRRIGATION = ''
		    OR SOURCE_WATER_IRRIGATION = 'None'
		    OR SOURCE_WATER_IRRIGATION = 'Rain Fed')
	`).Scan(&noIrrigation)
	result["farmersWithoutIrrigation"] = noIrrigation

	// Land size distribution (SELECT with CASE bucketing — read-only)
	landRows, err := h.DB.Query(`
		SELECT
			CASE
				WHEN CAST(AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(10,2)) = 0 THEN 'Landless'
				WHEN CAST(AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(10,2)) <= 1 THEN 'Marginal (0-1 acre)'
				WHEN CAST(AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(10,2)) <= 2.5 THEN 'Small (1-2.5 acres)'
				WHEN CAST(AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(10,2)) <= 5 THEN 'Semi-Medium (2.5-5 acres)'
				WHEN CAST(AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(10,2)) <= 10 THEN 'Medium (5-10 acres)'
				ELSE 'Large (>10 acres)'
			END as category,
			COUNT(*) as cnt
		FROM FAMILY
		WHERE OWN_AGRICULTURE_LAND = 'Yes'
		GROUP BY category
		ORDER BY cnt DESC
	`)
	if err == nil {
		defer landRows.Close()
		var items []CountItem
		for landRows.Next() {
			var item CountItem
			landRows.Scan(&item.Label, &item.Count)
			items = append(items, item)
		}
		result["landDistribution"] = items
	}

	waterRows, err := h.DB.Query(`
		SELECT COALESCE(SOURCE_WATER_IRRIGATION, 'Unknown') as label, COUNT(*) as cnt
		FROM FAMILY WHERE OWN_AGRICULTURE_LAND = 'Yes'
		GROUP BY SOURCE_WATER_IRRIGATION ORDER BY cnt DESC LIMIT 10
	`)
	if err == nil {
		defer waterRows.Close()
		var items []CountItem
		for waterRows.Next() {
			var item CountItem
			waterRows.Scan(&item.Label, &item.Count)
			items = append(items, item)
		}
		result["waterSourceDistribution"] = items
	}

	var kharifCount int
	h.DB.QueryRow(`
		SELECT COUNT(*) FROM FAMILY
		WHERE CULTIVATING_DURING_KHARIF_SEASON IS NOT NULL
		  AND CULTIVATING_DURING_KHARIF_SEASON != ''
		  AND CULTIVATING_DURING_KHARIF_SEASON != 'No'
	`).Scan(&kharifCount)
	result["kharifFarmers"] = kharifCount

	var rabiCount int
	h.DB.QueryRow(`
		SELECT COUNT(*) FROM FAMILY
		WHERE TAKING_CROPS_RABI_SEASON IS NOT NULL
		  AND TAKING_CROPS_RABI_SEASON != ''
		  AND TAKING_CROPS_RABI_SEASON != 'No'
	`).Scan(&rabiCount)
	result["rabiFarmers"] = rabiCount

	c.JSON(http.StatusOK, result)
}

// GetWelfareInsights — GET /insights/welfare
// Returns BPL household data, ration card distribution, and vulnerability metrics.
func (h *InsightHandler) GetWelfareInsights(c *gin.Context) {
	log.Println("[SELECT] GET /insights/welfare")
	result := gin.H{}

	hasRation := h.CC.Has("TYPE_OF_RATION_CARD")
	hasLatrine := h.CC.Has("TYPE_OF_LATRINE")
	hasLighting := h.CC.Has("SOURCE_OF_LIGHTING")

	if hasRation {
		rationRows, err := h.DB.Query(`
			SELECT COALESCE(TYPE_OF_RATION_CARD, 'Unknown') as label, COUNT(*) as cnt
			FROM FAMILY GROUP BY TYPE_OF_RATION_CARD ORDER BY cnt DESC
		`)
		if err == nil {
			defer rationRows.Close()
			var items []CountItem
			for rationRows.Next() {
				var item CountItem
				rationRows.Scan(&item.Label, &item.Count)
				items = append(items, item)
			}
			result["rationCardDistribution"] = items
		}

		var bplCount int
		h.DB.QueryRow(`
			SELECT COUNT(*) FROM FAMILY
			WHERE TYPE_OF_RATION_CARD LIKE '%BPL%' OR TYPE_OF_RATION_CARD LIKE '%Antyodaya%'
		`).Scan(&bplCount)
		result["bplHouseholds"] = bplCount

		if hasLatrine {
			var bplNoToilet int
			h.DB.QueryRow(`
				SELECT COUNT(*) FROM FAMILY
				WHERE (TYPE_OF_RATION_CARD LIKE '%BPL%' OR TYPE_OF_RATION_CARD LIKE '%Antyodaya%')
				  AND (TYPE_OF_LATRINE IS NULL OR TYPE_OF_LATRINE = '' OR TYPE_OF_LATRINE = 'No Latrine' OR TYPE_OF_LATRINE = 'None')
			`).Scan(&bplNoToilet)
			result["bplWithoutToilet"] = bplNoToilet
		}

		if hasLighting {
			var bplNoElec int
			h.DB.QueryRow(`
				SELECT COUNT(*) FROM FAMILY
				WHERE (TYPE_OF_RATION_CARD LIKE '%BPL%' OR TYPE_OF_RATION_CARD LIKE '%Antyodaya%')
				  AND (SOURCE_OF_LIGHTING IS NULL OR SOURCE_OF_LIGHTING = '' OR SOURCE_OF_LIGHTING = 'Kerosene' OR SOURCE_OF_LIGHTING = 'None')
			`).Scan(&bplNoElec)
			result["bplWithoutElectricity"] = bplNoElec
		}
	} else {
		result["_note"] = "TYPE_OF_RATION_CARD column not present — welfare insights limited"
	}

	var smallNoIrrigation int
	h.DB.QueryRow(`
		SELECT COUNT(*) FROM FAMILY
		WHERE OWN_AGRICULTURE_LAND = 'Yes'
		  AND CAST(AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(10,2)) <= 2.5
		  AND (SOURCE_WATER_IRRIGATION IS NULL OR SOURCE_WATER_IRRIGATION = '' OR SOURCE_WATER_IRRIGATION = 'Rain Fed' OR SOURCE_WATER_IRRIGATION = 'None')
	`).Scan(&smallNoIrrigation)
	result["smallFarmersWithoutIrrigation"] = smallNoIrrigation

	c.JSON(http.StatusOK, result)
}
