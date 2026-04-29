package handlers

import (
	"database/sql"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

type DistrictCropCountRow struct {
	DistrictID int64
	Crop       string
	CropCount  int64
}

type DistrictDominantCrop struct {
	DistrictID int64  `json:"district_id"`
	Crop       string `json:"crop"`
	Count      int64  `json:"count"`
}

type DistrictCropDominantHandler struct {
	DB *sql.DB
}

// GetDistrictDominantCrops returns the dominant crop (highest count) per district.
func (h *DistrictCropDominantHandler) GetDistrictDominantCrops(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT
			district_id,
			crop,
			COUNT(*) AS crop_count
		FROM (
			SELECT
				f.DISTRICT_ID AS district_id,
				TRIM(f.CULTIVATING_DURING_KHARIF_SEASON) AS crop
			FROM FAMILY f
			WHERE f.DISTRICT_ID IS NOT NULL
			  AND TRIM(COALESCE(f.CULTIVATING_DURING_KHARIF_SEASON, '')) <> ''

			UNION ALL

			SELECT
				f.DISTRICT_ID AS district_id,
				TRIM(f.CULTIVATING_DURING_RABI_SEASON) AS crop
			FROM FAMILY f
			WHERE f.DISTRICT_ID IS NOT NULL
			  AND TRIM(COALESCE(f.CULTIVATING_DURING_RABI_SEASON, '')) <> ''
		) crops
		GROUP BY district_id, crop
		ORDER BY district_id, crop_count DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch district crop counts", "detail": err.Error()})
		return
	}
	defer rows.Close()

	dominantByDistrict := map[int64]DistrictDominantCrop{}
	for rows.Next() {
		var row DistrictCropCountRow
		if scanErr := rows.Scan(&row.DistrictID, &row.Crop, &row.CropCount); scanErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse district crop counts", "detail": scanErr.Error()})
			return
		}

		current, exists := dominantByDistrict[row.DistrictID]
		if !exists || row.CropCount > current.Count {
			dominantByDistrict[row.DistrictID] = DistrictDominantCrop{
				DistrictID: row.DistrictID,
				Crop:       row.Crop,
				Count:      row.CropCount,
			}
		}
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read district crop counts", "detail": err.Error()})
		return
	}

	result := make([]DistrictDominantCrop, 0, len(dominantByDistrict))
	keys := make([]int64, 0, len(dominantByDistrict))
	for districtID := range dominantByDistrict {
		keys = append(keys, districtID)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, districtID := range keys {
		result = append(result, dominantByDistrict[districtID])
	}

	c.JSON(http.StatusOK, result)
}
