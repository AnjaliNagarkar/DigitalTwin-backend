package handlers

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
)

// TwinPDFHandler handles PDF export for the Digital Twin view.
type TwinPDFHandler struct {
	DB *sql.DB
	CC *ColumnChecker
}

// GenerateTwinPDF handles GET /api/twin/export-pdf.
// Query params: district_id, taluka_id, village_id
func (h *TwinPDFHandler) GenerateTwinPDF(c *gin.Context) {
	districtID := c.Query("district_id")
	talukaID := c.Query("taluka_id")
	villageID := c.Query("village_id")

	// Resolve location names from master tables
	districtName := h.resolveName("district_master", "pklDistrictId", "vsDistrictName", districtID)
	talukaName := h.resolveName("taluka_master", "pklTalukaId", "vsTalukaName", talukaID)
	villageName := h.resolveName("village_master", "pklVillageId", "vsVillageName", villageID)

	agri, err := h.queryTwinStats(districtID, talukaID, villageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database query failed", "detail": err.Error()})
		return
	}

	summary, err := h.queryVillageSummaryStats(districtID, talukaID, villageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "village summary query failed", "detail": err.Error()})
		return
	}

	doc := h.buildTwinPDF(districtName, talukaName, villageName, agri, summary)

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF rendering failed", "detail": err.Error()})
		return
	}

	parts := []string{"AgriTwin"}
	if districtName != "" {
		parts = append(parts, districtName)
	}
	if talukaName != "" {
		parts = append(parts, talukaName)
	}
	if villageName != "" {
		parts = append(parts, villageName)
	}
	parts = append(parts, time.Now().Format("2006-01-02"))
	fname := joinUnderscore(parts) + ".pdf"

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fname))
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

// twinStats holds agriculture-focused stats from the FAMILY table.
type twinStats struct {
	TotalFamilies     int
	TotalAgriLand     float64
	IrrigatedFamilies int
	KharifFamilies    int
	RabiFamilies      int
	ElectricFamilies  int
}

// villageSummaryStats holds population/demographic stats from FAMILY + FAMILY_MEMBER.
type villageSummaryStats struct {
	TotalHouseholds int
	TotalPopulation int
	MaleCount       int
	FemaleCount     int
	OtherCount      int
	TotalFarmers    int
	BPLHouseholds   int
}

func (h *TwinPDFHandler) buildWhereArgs(districtID, talukaID, villageID string) (string, []interface{}) {
	where := "WHERE 1=1"
	args := []interface{}{}
	if districtID != "" {
		where += " AND CAST(DISTRICT_ID AS CHAR) = ?"
		args = append(args, districtID)
	}
	if talukaID != "" {
		where += " AND CAST(TALUKA_ID AS CHAR) = ?"
		args = append(args, talukaID)
	}
	if villageID != "" {
		where += " AND CAST(VILLAGE_ID AS CHAR) = ?"
		args = append(args, villageID)
	}
	return where, args
}

func (h *TwinPDFHandler) queryTwinStats(districtID, talukaID, villageID string) (twinStats, error) {
	where, args := h.buildWhereArgs(districtID, talukaID, villageID)

	query := fmt.Sprintf(`
		SELECT
			COUNT(*),
			COALESCE(SUM(CAST(NULLIF(AREA_AGRICULTURE_LAND_ACRES, '') AS DECIMAL(10,2))), 0),
			SUM(CASE WHEN SOURCE_WATER_IRRIGATION IS NOT NULL AND SOURCE_WATER_IRRIGATION != '' AND SOURCE_WATER_IRRIGATION != 'No' THEN 1 ELSE 0 END),
			SUM(CASE WHEN CULTIVATING_DURING_KHARIF_SEASON = 'Yes' THEN 1 ELSE 0 END),
			SUM(CASE WHEN TAKING_CROPS_RABI_SEASON = 'Yes' THEN 1 ELSE 0 END),
			SUM(CASE WHEN ELECTRICITY_CONNECTION = 'Yes' THEN 1 ELSE 0 END)
		FROM FAMILY
		%s
	`, where)

	var s twinStats
	err := h.DB.QueryRow(query, args...).Scan(
		&s.TotalFamilies,
		&s.TotalAgriLand,
		&s.IrrigatedFamilies,
		&s.KharifFamilies,
		&s.RabiFamilies,
		&s.ElectricFamilies,
	)
	return s, err
}

func (h *TwinPDFHandler) queryVillageSummaryStats(districtID, talukaID, villageID string) (villageSummaryStats, error) {
	fWhere, fArgs := h.buildWhereArgs(districtID, talukaID, villageID)

	// Build FAMILY_MEMBER join-based WHERE (aliases f for FAMILY, fm for FAMILY_MEMBER)
	fmWhere := "WHERE 1=1"
	fmArgs := []interface{}{}
	if districtID != "" {
		fmWhere += " AND CAST(f.DISTRICT_ID AS CHAR) = ?"
		fmArgs = append(fmArgs, districtID)
	}
	if talukaID != "" {
		fmWhere += " AND CAST(f.TALUKA_ID AS CHAR) = ?"
		fmArgs = append(fmArgs, talukaID)
	}
	if villageID != "" {
		fmWhere += " AND CAST(f.VILLAGE_ID AS CHAR) = ?"
		fmArgs = append(fmArgs, villageID)
	}

	var s villageSummaryStats

	// Total households
	if err := h.DB.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM FAMILY %s", fWhere), fArgs...,
	).Scan(&s.TotalHouseholds); err != nil {
		return s, fmt.Errorf("households query: %w", err)
	}

	// Total population + gender breakdown
	popQuery := fmt.Sprintf(`
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN LOWER(TRIM(fm.GENDER)) = 'male'   THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN LOWER(TRIM(fm.GENDER)) = 'female' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN LOWER(TRIM(COALESCE(fm.GENDER,''))) NOT IN ('male','female') THEN 1 ELSE 0 END), 0)
		FROM FAMILY_MEMBER fm
		JOIN FAMILY f ON f.EXTERNAL_FAMILY_ID = fm.EXTERNAL_FAMILY_ID
		%s
	`, fmWhere)
	if err := h.DB.QueryRow(popQuery, fmArgs...).Scan(
		&s.TotalPopulation, &s.MaleCount, &s.FemaleCount, &s.OtherCount,
	); err != nil {
		return s, fmt.Errorf("population query: %w", err)
	}

	// Farmers: families that cultivate land in either season
	farmerQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM FAMILY
		%s
		  AND (CULTIVATING_DURING_KHARIF_SEASON = 'Yes' OR TAKING_CROPS_RABI_SEASON = 'Yes')
	`, fWhere)
	if err := h.DB.QueryRow(farmerQuery, fArgs...).Scan(&s.TotalFarmers); err != nil {
		return s, fmt.Errorf("farmers query: %w", err)
	}

	// BPL households (optional columns — skip gracefully if absent)
	if h.CC != nil {
		hasBPL := h.CC.Has("FAMILY_BELONG_BPL_CATEGORY")
		hasRation := h.CC.Has("RATION_CARD_TYPE")
		if hasBPL || hasRation {
			conds := ""
			if hasBPL {
				conds = "UPPER(TRIM(COALESCE(FAMILY_BELONG_BPL_CATEGORY,''))) = 'YES'"
			}
			if hasRation {
				if conds != "" {
					conds += " OR "
				}
				conds += "UPPER(TRIM(COALESCE(RATION_CARD_TYPE,''))) IN ('BPL','AAY')"
			}
			bplQuery := fmt.Sprintf("SELECT COUNT(*) FROM FAMILY %s AND (%s)", fWhere, conds)
			// intentionally ignore error — BPL is best-effort
			_ = h.DB.QueryRow(bplQuery, fArgs...).Scan(&s.BPLHouseholds)
		}
	}

	return s, nil
}

func (h *TwinPDFHandler) resolveName(table, idCol, nameCol, id string) string {
	if id == "" {
		return ""
	}
	var name string
	q := fmt.Sprintf("SELECT COALESCE(%s, '') FROM %s WHERE CAST(%s AS CHAR) = ? LIMIT 1", nameCol, table, idCol)
	_ = h.DB.QueryRow(q, id).Scan(&name)
	return name
}

func (h *TwinPDFHandler) buildTwinPDF(district, taluka, village string, agri twinStats, sum villageSummaryStats) *fpdf.Fpdf {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// ── Title ────────────────────────────────────────────────────────────────
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(22, 101, 52)
	pdf.CellFormat(0, 12, "Village Summary Report", "", 1, "C", false, 0, "")

	// Location subtitle
	loc := joinComma([]string{village, taluka, district})
	if loc == "" {
		loc = "All Locations"
	}
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(0, 7, loc, "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(0, 6, "Generated: "+time.Now().Format("02 Jan 2006"), "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// ── Section: Population Overview ─────────────────────────────────────────
	h.renderSectionHeader(pdf, "Population Overview")

	popRows := [][]string{
		{"Total Households", fmt.Sprintf("%d", sum.TotalHouseholds)},
		{"Total Population", fmt.Sprintf("%d", sum.TotalPopulation)},
		{"Male Count", fmt.Sprintf("%d", sum.MaleCount)},
		{"Female Count", fmt.Sprintf("%d", sum.FemaleCount)},
		{"Farmers (Kharif or Rabi)", fmt.Sprintf("%d", sum.TotalFarmers)},
		{"BPL Households", fmt.Sprintf("%d", sum.BPLHouseholds)},
	}
	h.renderTable(pdf, popRows)
	pdf.Ln(6)

	// ── Section: Agriculture Overview ────────────────────────────────────────
	h.renderSectionHeader(pdf, "Agriculture Overview")

	agriRows := [][]string{
		{"Total Families Surveyed", fmt.Sprintf("%d", agri.TotalFamilies)},
		{"Total Agricultural Land (acres)", fmt.Sprintf("%.2f", agri.TotalAgriLand)},
		{"Families with Irrigation", fmt.Sprintf("%d", agri.IrrigatedFamilies)},
		{"Kharif Season Cultivators", fmt.Sprintf("%d", agri.KharifFamilies)},
		{"Rabi Season Cultivators", fmt.Sprintf("%d", agri.RabiFamilies)},
		{"Families with Electricity", fmt.Sprintf("%d", agri.ElectricFamilies)},
	}
	h.renderTable(pdf, agriRows)

	// ── Footer ────────────────────────────────────────────────────────────────
	pdf.Ln(8)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(0, 5, "Agriculture Digital Twin — Read-only data export", "", 1, "C", false, 0, "")

	return pdf
}

func (h *TwinPDFHandler) renderSectionHeader(pdf *fpdf.Fpdf, title string) {
	pdf.SetFillColor(22, 101, 52)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 8, "  "+title, "", 1, "L", true, 0, "")
	pdf.Ln(1)
}

func (h *TwinPDFHandler) renderTable(pdf *fpdf.Fpdf, rows [][]string) {
	for i, row := range rows {
		if i%2 == 0 {
			pdf.SetFillColor(240, 248, 240)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.SetTextColor(30, 30, 30)
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(120, 8, row[0], "1", 0, "L", true, 0, "")
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(60, 8, row[1], "1", 1, "C", true, 0, "")
	}
}

func joinUnderscore(parts []string) string {
	result := ""
	for _, p := range parts {
		if p == "" {
			continue
		}
		if result != "" {
			result += "_"
		}
		for _, ch := range p {
			if ch == ' ' {
				result += "_"
			} else {
				result += string(ch)
			}
		}
	}
	return result
}

func joinComma(parts []string) string {
	result := ""
	for _, p := range parts {
		if p == "" {
			continue
		}
		if result != "" {
			result += ", "
		}
		result += p
	}
	return result
}
