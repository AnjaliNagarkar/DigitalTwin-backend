package handlers

// insights.go — Governance, Agriculture, and Welfare analytics
// ALL queries are SELECT-only. No INSERT/UPDATE/DELETE/DROP/ALTER anywhere.

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

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

func (h *InsightHandler) buildInsightFamilyFilters(alias string, c *gin.Context) (string, []interface{}) {
	clauses := []string{"1=1"}
	args := []interface{}{}

	toNullable := func(value string) interface{} {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return nil
		}
		return trimmed
	}

	districtID := toNullable(c.Query("district_id"))
	talukaID := toNullable(c.Query("taluka_id"))
	villageID := toNullable(c.Query("village_id"))

	clauses = append(clauses, fmt.Sprintf("(? IS NULL OR ? = '' OR CAST(%s.DISTRICT_ID AS CHAR) = ?)", alias))
	args = append(args, districtID, districtID, districtID)

	clauses = append(clauses, fmt.Sprintf("(? IS NULL OR ? = '' OR CAST(%s.TALUKA_ID AS CHAR) = ?)", alias))
	args = append(args, talukaID, talukaID, talukaID)

	clauses = append(clauses, fmt.Sprintf("(? IS NULL OR ? = '' OR CAST(%s.VILLAGE_ID AS CHAR) = ?)", alias))
	args = append(args, villageID, villageID, villageID)

	return strings.Join(clauses, " AND "), args
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
	_, cancel, ok := ensureDBReady(c, h.DB, "/insights/agriculture")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection unavailable"})
		return
	}
	defer cancel()
	districtID := strings.TrimSpace(c.Query("district_id"))
	talukaID := strings.TrimSpace(c.Query("taluka_id"))
	villageID := strings.TrimSpace(c.Query("village_id"))
	log.Printf("[FILTER] /insights/agriculture district_id=%q taluka_id=%q village_id=%q", districtID, talukaID, villageID)
	where, args := h.buildInsightFamilyFilters("f", c)
	result := gin.H{}
	result["landUtilizationRows"] = []gin.H{}
	result["seasonCropRows"] = []gin.H{}

	var totalFarmers int
	h.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM FAMILY f WHERE f.OWN_AGRICULTURE_LAND = 'Yes' AND %s", where), args...).Scan(&totalFarmers)
	result["totalFarmers"] = totalFarmers

	var noIrrigation int
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) FROM FAMILY f
		WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
		  AND (f.SOURCE_WATER_IRRIGATION IS NULL
		    OR f.SOURCE_WATER_IRRIGATION = ''
		    OR f.SOURCE_WATER_IRRIGATION = 'None'
		    OR f.SOURCE_WATER_IRRIGATION = 'Rain Fed')
		  AND %s
	`, where), args...).Scan(&noIrrigation)
	result["farmersWithoutIrrigation"] = noIrrigation

	// Land size distribution (SELECT with CASE bucketing — read-only)
	landRows, err := h.DB.Query(fmt.Sprintf(`
		SELECT
			CASE
				WHEN CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(10,2)) = 0 THEN 'Landless'
				WHEN CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(10,2)) <= 1 THEN 'Marginal (0-1 acre)'
				WHEN CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(10,2)) <= 2.5 THEN 'Small (1-2.5 acres)'
				WHEN CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(10,2)) <= 5 THEN 'Semi-Medium (2.5-5 acres)'
				WHEN CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(10,2)) <= 10 THEN 'Medium (5-10 acres)'
				ELSE 'Large (>10 acres)'
			END as category,
			COUNT(*) as cnt
		FROM FAMILY f
		WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
		  AND %s
		GROUP BY category
		ORDER BY cnt DESC
	`, where), args...)
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

	waterRows, err := h.DB.Query(fmt.Sprintf(`
		SELECT COALESCE(f.SOURCE_WATER_IRRIGATION, 'Unknown') as label, COUNT(*) as cnt
		FROM FAMILY f WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
		  AND %s
		GROUP BY f.SOURCE_WATER_IRRIGATION ORDER BY cnt DESC LIMIT 10
	`, where), args...)
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

	if h.CC.Has("AREA_AGRICULTURE_LAND_ACRES") && h.CC.Has("LAND_UNDER_CULTIVATION_ACRES") {
		var totalLand sql.NullFloat64
		var cultivatedLand sql.NullFloat64
		var unusedLand sql.NullFloat64
		var validRecords int64
		var invalidRecords int64

		err := h.DB.QueryRow(fmt.Sprintf(`
			SELECT
				COALESCE(ROUND(SUM(total_land), 2), 0) AS total_land,
				COALESCE(ROUND(SUM(cultivated_land), 2), 0) AS cultivated_land,
				COALESCE(ROUND(SUM(total_land - cultivated_land), 2), 0) AS unused_land,
				COUNT(*) AS valid_records
			FROM (
				SELECT
					CAST(AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2)) AS total_land,
					CAST(LAND_UNDER_CULTIVATION_ACRES AS DECIMAL(12,2)) AS cultivated_land
				FROM FAMILY f
				WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
				  AND %s
				  AND f.AREA_AGRICULTURE_LAND_ACRES IS NOT NULL
				  AND f.LAND_UNDER_CULTIVATION_ACRES IS NOT NULL
				  AND TRIM(f.AREA_AGRICULTURE_LAND_ACRES) <> ''
				  AND TRIM(f.LAND_UNDER_CULTIVATION_ACRES) <> ''
				  AND f.AREA_AGRICULTURE_LAND_ACRES REGEXP '^[0-9]*\\.?[0-9]+$'
				  AND f.LAND_UNDER_CULTIVATION_ACRES REGEXP '^[0-9]*\\.?[0-9]+$'
				  AND CAST(f.LAND_UNDER_CULTIVATION_ACRES AS DECIMAL(12,2)) <= CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2))
				  AND CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2)) BETWEEN 0 AND 500
			) t
		`, where), args...).Scan(&totalLand, &cultivatedLand, &unusedLand, &validRecords)

		if err == nil {
			err = h.DB.QueryRow(fmt.Sprintf(`
				SELECT COUNT(*) AS invalid_records
				FROM FAMILY f
				WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
				  AND %s
				  AND (
					f.AREA_AGRICULTURE_LAND_ACRES IS NULL
					OR f.LAND_UNDER_CULTIVATION_ACRES IS NULL
					OR TRIM(f.AREA_AGRICULTURE_LAND_ACRES) = ''
					OR TRIM(f.LAND_UNDER_CULTIVATION_ACRES) = ''
					OR f.AREA_AGRICULTURE_LAND_ACRES NOT REGEXP '^[0-9]*\\.?[0-9]+$'
					OR f.LAND_UNDER_CULTIVATION_ACRES NOT REGEXP '^[0-9]*\\.?[0-9]+$'
					OR CAST(f.LAND_UNDER_CULTIVATION_ACRES AS DECIMAL(12,2)) > CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2))
					OR CAST(f.AREA_AGRICULTURE_LAND_ACRES AS DECIMAL(12,2)) > 500
				  )
			`, where), args...).Scan(&invalidRecords)
		}

		if err == nil {
			total := totalLand.Float64
			cultivated := cultivatedLand.Float64
			unused := unusedLand.Float64
			if unused < 0 {
				unused = 0
			}

			cultivatedPercent := 0.0
			unusedPercent := 0.0
			if total > 0 {
				cultivatedPercent = (cultivated * 100) / total
				unusedPercent = (unused * 100) / total
			}

			result["landUtilizationSummary"] = gin.H{
				"total_land":         total,
				"cultivated_land":    cultivated,
				"unused_land":        unused,
				"valid_records":      validRecords,
				"invalid_records":    invalidRecords,
				"cultivated_percent": cultivatedPercent,
				"unused_percent":     unusedPercent,
			}

			result["landUtilizationRows"] = []gin.H{
				{
					"AREA_AGRICULTURE_LAND_ACRES":  total,
					"LAND_UNDER_CULTIVATION_ACRES": cultivated,
					"unused_land_acres":            unused,
					"valid_records":                validRecords,
					"invalid_records":              invalidRecords,
				},
			}
		}
	}

	rabiColumn := ""
	if h.CC.Has("CULTIVATING_DURING_RABI_SEASON") {
		rabiColumn = "CULTIVATING_DURING_RABI_SEASON"
	} else if h.CC.Has("TAKING_CROPS_RABI_SEASON") {
		rabiColumn = "TAKING_CROPS_RABI_SEASON"
	}

	if h.CC.Has("CULTIVATING_DURING_KHARIF_SEASON") || rabiColumn != "" {
		queryParts := []string{}
		cropArgs := []interface{}{}

		if h.CC.Has("CULTIVATING_DURING_KHARIF_SEASON") {
			queryParts = append(queryParts, `
				SELECT 'Kharif' AS season, TRIM(f.CULTIVATING_DURING_KHARIF_SEASON) AS crop, COUNT(*) AS cnt
				FROM FAMILY f
				WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
				  AND `+where+`
				  AND f.CULTIVATING_DURING_KHARIF_SEASON IS NOT NULL
				  AND TRIM(f.CULTIVATING_DURING_KHARIF_SEASON) != ''
				GROUP BY TRIM(f.CULTIVATING_DURING_KHARIF_SEASON)
			`)
			cropArgs = append(cropArgs, args...)
		}

		if rabiColumn != "" {
			queryParts = append(queryParts, `
				SELECT 'Rabi' AS season, TRIM(f.`+rabiColumn+`) AS crop, COUNT(*) AS cnt
				FROM FAMILY f
				WHERE f.OWN_AGRICULTURE_LAND = 'Yes'
				  AND `+where+`
				  AND f.`+rabiColumn+` IS NOT NULL
				  AND TRIM(f.`+rabiColumn+`) != ''
				GROUP BY TRIM(f.`+rabiColumn+`)
			`)
			cropArgs = append(cropArgs, args...)
		}

		if len(queryParts) > 0 {
			cropRows, err := h.DB.Query(`
				SELECT season, crop, cnt
				FROM (
					`+queryParts[0]+func() string {
				if len(queryParts) > 1 {
					return ` UNION ALL ` + queryParts[1]
				}
				return ``
			}()+`
				) merged
				ORDER BY cnt DESC
			`, cropArgs...)

			if err == nil {
				defer cropRows.Close()
				items := []gin.H{}
				for cropRows.Next() {
					var season string
					var crop string
					var count int
					if scanErr := cropRows.Scan(&season, &crop, &count); scanErr == nil {
						items = append(items, gin.H{
							"season": season,
							"crop":   crop,
							"count":  count,
						})
					}
				}
				result["seasonCropRows"] = items
			}
		}
	}

	var kharifCount int
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) FROM FAMILY f
		WHERE f.CULTIVATING_DURING_KHARIF_SEASON IS NOT NULL
		  AND f.CULTIVATING_DURING_KHARIF_SEASON != ''
		  AND f.CULTIVATING_DURING_KHARIF_SEASON != 'No'
		  AND %s
	`, where), args...).Scan(&kharifCount)
	result["kharifFarmers"] = kharifCount

	var rabiCount int
	h.DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) FROM FAMILY f
		WHERE f.TAKING_CROPS_RABI_SEASON IS NOT NULL
		  AND f.TAKING_CROPS_RABI_SEASON != ''
		  AND f.TAKING_CROPS_RABI_SEASON != 'No'
		  AND %s
	`, where), args...).Scan(&rabiCount)
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
