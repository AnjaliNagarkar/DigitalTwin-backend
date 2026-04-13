package handlers

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
)

// ─── Public types ─────────────────────────────────────────────────────────────

// PDFRequest is sent by the frontend as a JSON body.
type PDFRequest struct {
	DistrictID   string          `json:"districtId"`
	DistrictName string          `json:"districtName"`
	TalukaID     string          `json:"talukaId"`
	TalukaName   string          `json:"talukaName"`
	VillageID    string          `json:"villageId"`
	VillageName  string          `json:"villageName"`
	// Charts holds optional pre-rendered donut chart images from the frontend
	// (base64-encoded PNG, no data-URL prefix).
	Charts []PDFChartImage `json:"charts"`
	// ProblemFilters holds the sidebar problem-filter counts from the frontend.
	ProblemFilters   []PDFProblemFilter `json:"problemFilters"`
	ProblemMatchTotal int               `json:"problemMatchTotal"`
}

// PDFChartImage is one rendered chart sent from the frontend.
type PDFChartImage struct {
	Title    string            `json:"title"`
	Image    string            `json:"image"`    // base64 PNG, without the "data:image/png;base64," prefix
	Segments []PDFChartSegment `json:"segments"` // colour legend data
}

// PDFChartSegment is one slice of a pie/donut chart.
type PDFChartSegment struct {
	Label string `json:"label"`
	Pct   int    `json:"pct"`
	Color string `json:"color"` // CSS hex, e.g. "#16a34a"
}

// PDFProblemFilter holds a single problem-filter entry from the frontend sidebar.
type PDFProblemFilter struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Count  int    `json:"count"`
	Total  int    `json:"total"`
	Active bool   `json:"active"`
}

// PDFHandler handles PDF generation (POST /pdf/report).
type PDFHandler struct {
	DB *sql.DB
	CC *ColumnChecker
}

// ─── Internal types ───────────────────────────────────────────────────────────

type pdfStats struct {
	Total                      int
	Irrigated                  int
	NoLatrine                  int
	NoElec                     int
	BPL                        int
	Both, KharifOnly, RabiOnly int
	NoCrop                     int
	Marginal, Small, MedLarge  int
}

// rgbColor is a simple RGB triple for gofpdf colour helpers.
type rgbColor struct{ r, g, b int }

// ─── Colour palette ───────────────────────────────────────────────────────────

var (
	pGreen     = rgbColor{22, 163, 74}
	pDarkGreen = rgbColor{14, 103, 48}
	pRed       = rgbColor{220, 38, 38}
	pAmber     = rgbColor{217, 119, 6}
	pBlue      = rgbColor{59, 130, 246}
	pPurple    = rgbColor{139, 92, 246}
	pGray      = rgbColor{107, 114, 128}
	pLightGray = rgbColor{243, 244, 246}
	pWhite     = rgbColor{255, 255, 255}
	pDark      = rgbColor{17, 24, 39}
	pTextGray  = rgbColor{75, 85, 99}
	pBorder    = rgbColor{229, 231, 235}
	pTeal      = rgbColor{16, 185, 129}
)

// ─── Handler entry point ──────────────────────────────────────────────────────

// GeneratePDF handles POST /pdf/report.
// It queries the DB, computes stats, and streams a PDF to the client.
func (h *PDFHandler) GeneratePDF(c *gin.Context) {
	var req PDFRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid request body",
			"detail": err.Error(),
		})
		return
	}

	houses, err := h.queryHousesForPDF(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "database query failed",
			"detail": err.Error(),
		})
		return
	}

	stats := computePDFStats(houses)
	doc := buildAgriPDF(req, houses, stats)

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "PDF rendering failed",
			"detail": err.Error(),
		})
		return
	}

	fname := pdfFilename(req)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fname))
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

// ─── DB query ─────────────────────────────────────────────────────────────────

func (h *PDFHandler) queryHousesForPDF(req PDFRequest) ([]HouseRecord, error) {
	latCol, lngCol := h.CC.LatCol, h.CC.LngCol
	if latCol == "" {
		latCol = "LATITUDE"
	}
	if lngCol == "" {
		lngCol = "LONGITUDE"
	}

	where := fmt.Sprintf(
		"WHERE f.%s IS NOT NULL AND f.%s IS NOT NULL AND f.%s != 0 AND f.%s != 0",
		latCol, lngCol, latCol, lngCol,
	)
	args := []interface{}{}

	if req.DistrictID != "" {
		where += " AND CAST(f.DISTRICT_ID AS CHAR) = ?"
		args = append(args, req.DistrictID)
	}
	if req.TalukaID != "" {
		where += " AND CAST(f.TALUKA_ID AS CHAR) = ?"
		args = append(args, req.TalukaID)
	}
	if req.VillageID != "" {
		where += " AND CAST(f.VILLAGE_ID AS CHAR) = ?"
		args = append(args, req.VillageID)
	}

	query := fmt.Sprintf(`
		SELECT
			f.FAMILY_ID,
			COALESCE(CAST(f.DISTRICT_ID AS CHAR), ''),
			COALESCE(dm.vsDisplayName, dm.vsDistrictName, ''),
			COALESCE(CAST(f.TALUKA_ID AS CHAR), ''),
			COALESCE(tm.vsDisplayName, tm.vsTalukaName, ''),
			COALESCE(CAST(f.VILLAGE_ID AS CHAR), ''),
			COALESCE(vm.vsDisplayName, vm.vsVillageName, ''),
			COALESCE(f.%s, 0),
			COALESCE(f.%s, 0),
			COALESCE(f.AREA_AGRICULTURE_LAND_ACRES, ''),
			COALESCE(f.LAND_UNDER_CULTIVATION_ACRES, ''),
			COALESCE(f.OWN_AGRICULTURE_LAND, ''),
			COALESCE(f.SOURCE_WATER_IRRIGATION, ''),
			COALESCE(f.CULTIVATING_DURING_KHARIF_SEASON, ''),
			COALESCE(f.TAKING_CROPS_RABI_SEASON, ''),
			COALESCE(f.SANITATION_TOILET_FACILITY, ''),
			COALESCE(f.ELECTRICITY_CONNECTION, ''),
			COALESCE(f.RATION_CARD_TYPE, ''),
			COALESCE(TRIM(CONCAT(
				COALESCE(f.FIRST_NAME_HOUSEHOLD_HEAD, ''), ' ',
				COALESCE(f.MIDDLE_NAME_HOUSEHOLD_HEAD, ''), ' ',
				COALESCE(f.LAST_NAME_HOUSEHOLD_HEAD, '')
			)), '')
		FROM FAMILY f
		LEFT JOIN district_master dm ON dm.pklDistrictId = f.DISTRICT_ID
		LEFT JOIN taluka_master tm   ON tm.pklTalukaId   = f.TALUKA_ID
		LEFT JOIN village_master vm  ON vm.pklVillageId  = f.VILLAGE_ID
		%s
		ORDER BY f.FAMILY_ID
		LIMIT 5000
	`, latCol, lngCol, where)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var houses []HouseRecord
	for rows.Next() {
		var house HouseRecord
		if err := rows.Scan(
			&house.FamilyID,
			&house.DistrictID, &house.DistrictName,
			&house.TalukaID, &house.TalukaName,
			&house.VillageID, &house.VillageName,
			&house.Latitude, &house.Longitude,
			&house.TotalLand, &house.CultivatedLand, &house.OwnLand,
			&house.WaterSource, &house.Kharif, &house.Rabi,
			&house.Latrine, &house.Lighting, &house.RationCard,
			&house.HeadName,
		); err != nil {
			return nil, err
		}
		houses = append(houses, house)
	}
	return houses, nil
}

// ─── Statistics ───────────────────────────────────────────────────────────────

func computePDFStats(houses []HouseRecord) pdfStats {
	s := pdfStats{Total: len(houses)}
	for _, h := range houses {
		ws := strings.ToLower(strings.TrimSpace(h.WaterSource))
		if ws != "" && ws != "none" && ws != "rain fed" && !strings.Contains(ws, "no source") {
			s.Irrigated++
		}

		lat := strings.ToLower(strings.TrimSpace(h.Latrine))
		if lat == "" || lat == "no latrine" || lat == "none" {
			s.NoLatrine++
		}

		lit := strings.ToLower(strings.TrimSpace(h.Lighting))
		if lit == "" || lit == "kerosene" || lit == "none" {
			s.NoElec++
		}

		rc := strings.ToLower(h.RationCard)
		if strings.Contains(rc, "bpl") || strings.Contains(rc, "antyodaya") {
			s.BPL++
		}

		k := strings.ToLower(strings.TrimSpace(h.Kharif)) == "yes"
		r := strings.ToLower(strings.TrimSpace(h.Rabi)) == "yes"
		switch {
		case k && r:
			s.Both++
		case k:
			s.KharifOnly++
		case r:
			s.RabiOnly++
		default:
			s.NoCrop++
		}

		land, _ := strconv.ParseFloat(strings.TrimSpace(h.TotalLand), 64)
		if land <= 1 {
			s.Marginal++
		} else if land <= 2.5 {
			s.Small++
		} else {
			s.MedLarge++
		}
	}
	return s
}

// ─── PDF layout ───────────────────────────────────────────────────────────────

func buildAgriPDF(req PDFRequest, houses []HouseRecord, s pdfStats) *fpdf.Fpdf {
	doc := fpdf.New("P", "mm", "A4", "")
	doc.SetMargins(14, 14, 14)
	doc.SetAutoPageBreak(true, 16)

	pageW, _ := doc.GetPageSize()
	ml, _, mr, _ := doc.GetMargins()
	uw := pageW - ml - mr // usable width (182 mm for A4 with 14 mm margins)

	// ── Footer (runs automatically on every page) ─────────────────────────
	generatedAt := time.Now().Format("02 Jan 2006  15:04")
	doc.SetFooterFunc(func() {
		doc.SetY(-12)
		doc.SetDrawColor(pBorder.r, pBorder.g, pBorder.b)
		doc.SetLineWidth(0.2)
		doc.Line(ml, doc.GetY(), ml+uw, doc.GetY())
		doc.SetFont("Helvetica", "I", 6.5)
		doc.SetTextColor(pGray.r, pGray.g, pGray.b)
		doc.CellFormat(uw/2, 5, "AgriTwin — Agriculture Digital Twin Report", "", 0, "L", false, 0, "")
		doc.CellFormat(uw/2, 5,
			fmt.Sprintf("Page %d  |  Generated %s", doc.PageNo(), generatedAt),
			"", 0, "R", false, 0, "")
	})

	doc.AddPage()

	// ── 1. HEADER BAR ────────────────────────────────────────────────────────
	doc.SetFillColor(pGreen.r, pGreen.g, pGreen.b)
	doc.Rect(ml, 14, uw, 8, "F")

	doc.SetXY(ml+2, 14)
	doc.SetFont("Helvetica", "B", 10)
	doc.SetTextColor(pWhite.r, pWhite.g, pWhite.b)
	doc.CellFormat(uw-55, 8, "AgriTwin  |  Agriculture Digital Twin", "", 0, "L", false, 0, "")

	doc.SetFont("Helvetica", "", 8)
	doc.CellFormat(53, 8, time.Now().Format("02 January 2006"), "", 0, "R", false, 0, "")

	// ── 2. TITLE + LOCATION ─────────────────────────────────────────────────
	doc.SetXY(ml, 24)
	doc.SetFont("Helvetica", "B", 17)
	doc.SetTextColor(pDark.r, pDark.g, pDark.b)
	doc.CellFormat(uw, 9, "Village Agriculture Report", "", 1, "L", false, 0, "")

	crumbParts := []string{}
	if req.DistrictName != "" {
		crumbParts = append(crumbParts, req.DistrictName)
	}
	if req.TalukaName != "" {
		crumbParts = append(crumbParts, req.TalukaName)
	}
	if req.VillageName != "" {
		crumbParts = append(crumbParts, req.VillageName)
	}
	crumb := strings.Join(crumbParts, "  \u203a  ")
	if crumb == "" {
		crumb = "All Households — Maharashtra"
	}

	doc.SetFont("Helvetica", "", 9)
	doc.SetTextColor(pTextGray.r, pTextGray.g, pTextGray.b)
	doc.CellFormat(uw, 5, crumb, "", 1, "L", false, 0, "")

	// Green divider
	doc.SetDrawColor(pGreen.r, pGreen.g, pGreen.b)
	doc.SetLineWidth(0.4)
	divY := doc.GetY() + 2
	doc.Line(ml, divY, ml+uw, divY)
	doc.SetY(divY + 4)

	// ── 3. SUMMARY STAT BOXES ───────────────────────────────────────────────
	pdfSectionTitle(doc, "SUMMARY STATISTICS", ml, uw)
	doc.Ln(3)

	type statBox struct {
		label string
		value string
		note  string
		color rgbColor
	}
	boxes := []statBox{
		{
			"Total Households",
			fmt.Sprintf("%d", s.Total),
			"survey area",
			pGreen,
		},
		{
			"Irrigated Farms",
			fmt.Sprintf("%d", s.Irrigated),
			fmt.Sprintf("%d%% of total", ppct(s.Irrigated, s.Total)),
			pBlue,
		},
		{
			"Have Sanitation",
			fmt.Sprintf("%d", s.Total-s.NoLatrine),
			fmt.Sprintf("%d%% coverage", ppct(s.Total-s.NoLatrine, s.Total)),
			pTeal,
		},
		{
			"BPL Households",
			fmt.Sprintf("%d", s.BPL),
			fmt.Sprintf("%d%% of total", ppct(s.BPL, s.Total)),
			pAmber,
		},
	}

	bw := (uw - 3*3) / 4 // 4 boxes, 3 gaps of 3 mm
	y0 := doc.GetY()
	for i, box := range boxes {
		bx := ml + float64(i)*(bw+3)

		doc.SetFillColor(pLightGray.r, pLightGray.g, pLightGray.b)
		doc.Rect(bx, y0, bw, 21, "F")
		doc.SetFillColor(box.color.r, box.color.g, box.color.b)
		doc.Rect(bx, y0, bw, 1.5, "F")

		doc.SetXY(bx, y0+3)
		doc.SetFont("Helvetica", "B", 14)
		doc.SetTextColor(pDark.r, pDark.g, pDark.b)
		doc.CellFormat(bw, 8, box.value, "", 1, "C", false, 0, "")

		doc.SetX(bx)
		doc.SetFont("Helvetica", "B", 7)
		doc.SetTextColor(pTextGray.r, pTextGray.g, pTextGray.b)
		doc.CellFormat(bw, 4, box.label, "", 1, "C", false, 0, "")

		doc.SetX(bx)
		doc.SetFont("Helvetica", "", 6.5)
		doc.SetTextColor(box.color.r, box.color.g, box.color.b)
		doc.CellFormat(bw, 4, box.note, "", 1, "C", false, 0, "")
	}
	doc.SetY(y0 + 25)

	// ── 4. PROBLEM FILTER SUMMARY ───────────────────────────────────────────
	if len(req.ProblemFilters) > 0 {
		pdfSectionTitle(doc, "PROBLEM FILTER SUMMARY", ml, uw)
		doc.Ln(3)

		// Three filter boxes side by side
		pfW := (uw - float64(len(req.ProblemFilters)-1)*4) / float64(len(req.ProblemFilters))
		pfY := doc.GetY()

		// Colour mapping for each known filter key
		filterColors := map[string]rgbColor{
			"noRationCard": pAmber,
			"noIrrigation": pPurple,
			"noLand":       pRed,
		}

		for i, pf := range req.ProblemFilters {
			pfX := ml + float64(i)*(pfW+4)
			col := filterColors[pf.Key]
			if col == (rgbColor{}) {
				col = pGray
			}

			// Box background
			doc.SetFillColor(pLightGray.r, pLightGray.g, pLightGray.b)
			doc.Rect(pfX, pfY, pfW, 22, "F")

			// Top accent stripe (thicker if active)
			stripeH := 1.5
			if pf.Active {
				stripeH = 3.0
			}
			doc.SetFillColor(col.r, col.g, col.b)
			doc.Rect(pfX, pfY, pfW, stripeH, "F")

			// Count value
			doc.SetXY(pfX, pfY+stripeH+2)
			doc.SetFont("Helvetica", "B", 14)
			doc.SetTextColor(pDark.r, pDark.g, pDark.b)
			doc.CellFormat(pfW, 8, fmt.Sprintf("%d", pf.Count), "", 1, "C", false, 0, "")

			// Label
			doc.SetX(pfX)
			doc.SetFont("Helvetica", "B", 6.5)
			doc.SetTextColor(pTextGray.r, pTextGray.g, pTextGray.b)
			doc.CellFormat(pfW, 4, pf.Label, "", 1, "C", false, 0, "")

			// Percentage note
			pct := 0
			if pf.Total > 0 {
				pct = int(float64(pf.Count)/float64(pf.Total)*100 + 0.5)
			}
			activeTag := ""
			if pf.Active {
				activeTag = " ★ filtered"
			}
			doc.SetX(pfX)
			doc.SetFont("Helvetica", "", 6)
			doc.SetTextColor(col.r, col.g, col.b)
			doc.CellFormat(pfW, 4, fmt.Sprintf("%d%% of total%s", pct, activeTag), "", 1, "C", false, 0, "")
		}
		doc.SetY(pfY + 26)

		// If multiple filters were active, show the combined match count
		if req.ProblemMatchTotal > 0 {
			doc.SetFont("Helvetica", "I", 7)
			doc.SetTextColor(pTextGray.r, pTextGray.g, pTextGray.b)
			activeLbls := []string{}
			for _, pf := range req.ProblemFilters {
				if pf.Active {
					activeLbls = append(activeLbls, pf.Label)
				}
			}
			if len(activeLbls) > 1 {
				doc.CellFormat(uw, 4.5,
					fmt.Sprintf("Combined match (%s): %d households", strings.Join(activeLbls, " + "), req.ProblemMatchTotal),
					"", 1, "L", false, 0, "")
				doc.Ln(2)
			}
		}
	}

	// ── 5. COLOUR LEGEND ────────────────────────────────────────────────────
	pdfSectionTitle(doc, "COLOUR LEGEND", ml, uw)
	doc.Ln(2)

	type legendItem struct {
		label string
		color rgbColor
	}
	legItems := []legendItem{
		{"Irrigated farm", pGreen},
		{"Rain-fed / no water source", pRed},
		{"Has sanitation facility", pTeal},
		{"No sanitation (open defecation)", pRed},
		{"BPL / Antyodaya household", pAmber},
		{"No electricity connection", pPurple},
	}

	ly0 := doc.GetY()
	colW := uw / 3.0
	rows3 := (len(legItems) + 2) / 3
	for i, leg := range legItems {
		col := i % 3
		row := i / 3
		lx := ml + float64(col)*colW
		ly := ly0 + float64(row)*6
		doc.SetFillColor(leg.color.r, leg.color.g, leg.color.b)
		doc.Rect(lx, ly+1.2, 3, 3, "F")
		doc.SetFont("Helvetica", "", 7.5)
		doc.SetTextColor(pDark.r, pDark.g, pDark.b)
		doc.SetXY(lx+4.5, ly)
		doc.CellFormat(colW-5, 5, leg.label, "", 0, "L", false, 0, "")
	}
	doc.SetY(ly0 + float64(rows3)*6 + 4)

	// ── 6. AGRICULTURE STACKED BAR CHARTS ───────────────────────────────────
	pdfSectionTitle(doc, "AGRICULTURE OVERVIEW", ml, uw)
	doc.Ln(3)

	type chartSeg struct {
		label string
		count int
		color rgbColor
	}
	type barChart struct {
		title string
		segs  []chartSeg
	}

	bcharts := []barChart{
		{
			"Irrigation Coverage",
			[]chartSeg{
				{"Irrigated", s.Irrigated, pGreen},
				{"Rain-fed", s.Total - s.Irrigated, pRed},
			},
		},
		{
			"Crop Seasons",
			[]chartSeg{
				{"Both seasons", s.Both, pGreen},
				{"Kharif only", s.KharifOnly, pAmber},
				{"Rabi only", s.RabiOnly, pBlue},
				{"No crop data", s.NoCrop, pGray},
			},
		},
		{
			"Land Holdings",
			[]chartSeg{
				{"Marginal <=1 ac", s.Marginal, pRed},
				{"Small 1-2.5 ac", s.Small, pAmber},
				{"Med/Large >2.5 ac", s.MedLarge, pGreen},
			},
		},
	}

	labelW := 46.0
	barAreaW := uw - labelW
	barH := 5.0

	for _, ch := range bcharts {
		cy := doc.GetY()

		doc.SetFont("Helvetica", "B", 8)
		doc.SetTextColor(pDark.r, pDark.g, pDark.b)
		doc.SetXY(ml, cy)
		doc.CellFormat(labelW, barH, ch.title, "", 0, "L", false, 0, "")

		barX := ml + labelW
		doc.SetFillColor(229, 231, 235)
		doc.Rect(barX, cy+0.5, barAreaW, barH-1, "F")

		total := 0
		for _, seg := range ch.segs {
			total += seg.count
		}
		bx := barX
		for _, seg := range ch.segs {
			if total == 0 || seg.count == 0 {
				continue
			}
			sw := float64(seg.count) / float64(total) * barAreaW
			doc.SetFillColor(seg.color.r, seg.color.g, seg.color.b)
			doc.Rect(bx, cy+0.5, sw, barH-1, "F")
			bx += sw
		}

		doc.Ln(barH + 1)

		// Mini legend
		lx2 := barX
		doc.SetFont("Helvetica", "", 6.5)
		for _, seg := range ch.segs {
			p := ppct(seg.count, total)
			doc.SetFillColor(seg.color.r, seg.color.g, seg.color.b)
			doc.Rect(lx2, doc.GetY()+1, 2.5, 2.5, "F")
			doc.SetTextColor(pDark.r, pDark.g, pDark.b)
			doc.SetX(lx2 + 3.5)
			doc.CellFormat(38, 4.5, fmt.Sprintf("%s %d%%", seg.label, p), "", 0, "L", false, 0, "")
			lx2 += 40
			if lx2 > ml+uw-38 {
				doc.Ln(4.5)
				lx2 = barX
			}
		}
		doc.Ln(7)
	}

	// ── 7. HOUSEHOLD TABLE ───────────────────────────────────────────────────
	doc.Ln(2)

	tableHouses := houses
	limited := false
	if len(tableHouses) > 500 {
		tableHouses = tableHouses[:500]
		limited = true
	}

	pdfSectionTitle(doc,
		fmt.Sprintf("HOUSEHOLD DATA  (%d records)", s.Total),
		ml, uw)
	doc.Ln(2)

	if limited {
		doc.SetFont("Helvetica", "I", 7)
		doc.SetTextColor(pAmber.r, pAmber.g, pAmber.b)
		doc.CellFormat(uw, 4.5,
			fmt.Sprintf(
				"Note: Table shows first 500 of %d records. Use CSV / JSON export for the full dataset.",
				s.Total,
			), "", 1, "L", false, 0, "")
		doc.Ln(1)
	}

	// Column definitions — widths must sum to uw (182 mm)
	type tcol struct {
		hdr   string
		width float64
		align string
	}
	tcols := []tcol{
		{"#", 7, "C"},
		{"Head Name", 45, "L"},
		{"Land (ac)", 16, "C"},
		{"Water Source", 30, "L"},
		{"Kharif", 13, "C"},
		{"Rabi", 13, "C"},
		{"Sanitation", 28, "L"},
		{"Lighting", 30, "L"},
	} // total: 7+45+16+30+13+13+28+30 = 182 ✓

	// Table header row
	doc.SetFillColor(pDarkGreen.r, pDarkGreen.g, pDarkGreen.b)
	doc.SetTextColor(pWhite.r, pWhite.g, pWhite.b)
	doc.SetFont("Helvetica", "B", 7)
	for _, tc := range tcols {
		doc.CellFormat(tc.width, 6, "  "+tc.hdr, "0", 0, tc.align, true, 0, "")
	}
	doc.Ln(6)

	// Data rows
	doc.SetFont("Helvetica", "", 7)
	for i, h := range tableHouses {
		if i%2 == 0 {
			doc.SetFillColor(pLightGray.r, pLightGray.g, pLightGray.b)
		} else {
			doc.SetFillColor(pWhite.r, pWhite.g, pWhite.b)
		}
		fill := true

		name := pdfTrunc(strings.TrimSpace(h.HeadName), 22)
		if name == "" {
			name = fmt.Sprintf("Family #%d", h.FamilyID)
		}

		ws := pdfTrunc(strings.TrimSpace(h.WaterSource), 16)
		wsRF := ws == "" ||
			strings.Contains(strings.ToLower(ws), "rain") ||
			strings.ToLower(ws) == "none"

		san := strings.TrimSpace(h.Latrine)
		if san == "" {
			san = "None"
		}
		san = pdfTrunc(san, 16)
		sanBad := strings.ToLower(san) == "none" || strings.ToLower(san) == "no latrine"

		lit := strings.TrimSpace(h.Lighting)
		if lit == "" {
			lit = "None"
		}
		lit = pdfTrunc(lit, 16)

		kh := "N"
		if strings.ToLower(strings.TrimSpace(h.Kharif)) == "yes" {
			kh = "Y"
		}
		rb := "N"
		if strings.ToLower(strings.TrimSpace(h.Rabi)) == "yes" {
			rb = "Y"
		}

		// Col 0 — Sr. No (gray)
		doc.SetTextColor(pGray.r, pGray.g, pGray.b)
		doc.CellFormat(tcols[0].width, 5, fmt.Sprintf("  %d", i+1), "0", 0, tcols[0].align, fill, 0, "")
		// Col 1 — Head Name
		doc.SetTextColor(pDark.r, pDark.g, pDark.b)
		doc.CellFormat(tcols[1].width, 5, "  "+name, "0", 0, tcols[1].align, fill, 0, "")
		// Col 2 — Land
		doc.CellFormat(tcols[2].width, 5, "  "+h.TotalLand, "0", 0, tcols[2].align, fill, 0, "")
		// Col 3 — Water Source (green = irrigated, red = rain-fed)
		if wsRF {
			doc.SetTextColor(pRed.r, pRed.g, pRed.b)
		} else {
			doc.SetTextColor(pGreen.r, pGreen.g, pGreen.b)
		}
		doc.CellFormat(tcols[3].width, 5, "  "+ws, "0", 0, tcols[3].align, fill, 0, "")
		// Col 4 — Kharif
		if kh == "Y" {
			doc.SetTextColor(pGreen.r, pGreen.g, pGreen.b)
		} else {
			doc.SetTextColor(pGray.r, pGray.g, pGray.b)
		}
		doc.CellFormat(tcols[4].width, 5, "  "+kh, "0", 0, tcols[4].align, fill, 0, "")
		// Col 5 — Rabi
		if rb == "Y" {
			doc.SetTextColor(pGreen.r, pGreen.g, pGreen.b)
		} else {
			doc.SetTextColor(pGray.r, pGray.g, pGray.b)
		}
		doc.CellFormat(tcols[5].width, 5, "  "+rb, "0", 0, tcols[5].align, fill, 0, "")
		// Col 6 — Sanitation
		if sanBad {
			doc.SetTextColor(pRed.r, pRed.g, pRed.b)
		} else {
			doc.SetTextColor(pDark.r, pDark.g, pDark.b)
		}
		doc.CellFormat(tcols[6].width, 5, "  "+san, "0", 0, tcols[6].align, fill, 0, "")
		// Col 7 — Lighting
		switch strings.ToLower(lit) {
		case "electricity":
			doc.SetTextColor(pGreen.r, pGreen.g, pGreen.b)
		case "kerosene", "none":
			doc.SetTextColor(pAmber.r, pAmber.g, pAmber.b)
		default:
			doc.SetTextColor(pDark.r, pDark.g, pDark.b)
		}
		doc.CellFormat(tcols[7].width, 5, "  "+lit, "0", 0, tcols[7].align, fill, 0, "")

		doc.Ln(5)
	}

	// ── 8. FRONTEND CHART IMAGES (optional, second page) ─────────────────────
	validCharts := make([]PDFChartImage, 0, len(req.Charts))
	for _, ch := range req.Charts {
		if ch.Image == "" {
			continue
		}
		imgBytes, err := base64.StdEncoding.DecodeString(ch.Image)
		if err != nil || !isValidPNG(imgBytes) {
			continue
		}
		validCharts = append(validCharts, PDFChartImage{Title: ch.Title, Image: ch.Image})
		_ = imgBytes
	}

	if len(validCharts) > 0 {
		doc.AddPage()
		pdfSectionTitle(doc, "AGRICULTURE CHARTS", ml, uw)
		doc.Ln(5)

		imgW := (uw - 6) / 2 // two charts per row, 6 mm gap
		imgH := 52.0          // fixed image height so legend placement is predictable
		cix := ml
		ciy := doc.GetY()

		for i, ch := range validCharts {
			imgBytes, _ := base64.StdEncoding.DecodeString(ch.Image)
			key := fmt.Sprintf("_fc_%d", i)
			doc.RegisterImageOptionsReader(
				key,
				fpdf.ImageOptions{ImageType: "PNG"},
				bytes.NewReader(imgBytes),
			)
			if !doc.Ok() {
				break
			}

			// ── Chart title ───────────────────────────────────────────────
			doc.SetFont("Helvetica", "B", 8.5)
			doc.SetTextColor(pDark.r, pDark.g, pDark.b)
			doc.SetXY(cix, ciy)
			doc.CellFormat(imgW, 5, ch.Title, "", 0, "C", false, 0, "")

			// ── Donut image ───────────────────────────────────────────────
			doc.ImageOptions(key, cix+(imgW-imgH)/2, ciy+6, imgH, imgH, false,
				fpdf.ImageOptions{ImageType: "PNG"}, 0, "")

			// ── Colour legend below the image ─────────────────────────────
			legendY := ciy + 6 + imgH + 3
			dotSize := 3.0
			rowH := 5.0

			// Two-column layout inside the chart cell
			colW2 := imgW / 2
			for j, seg := range ch.Segments {
				col := j % 2
				row := j / 2
				lx := cix + float64(col)*colW2 + 1
				ly := legendY + float64(row)*rowH

				col2 := hexToRGB(seg.Color)
				doc.SetFillColor(col2.r, col2.g, col2.b)
				doc.Rect(lx, ly+0.9, dotSize, dotSize, "F")

				doc.SetFont("Helvetica", "", 6.5)
				doc.SetTextColor(pDark.r, pDark.g, pDark.b)
				doc.SetXY(lx+dotSize+1.5, ly)
				label := seg.Label
				if seg.Pct > 0 {
					label = fmt.Sprintf("%s  %d%%", seg.Label, seg.Pct)
				}
				doc.CellFormat(colW2-dotSize-3, rowH, label, "", 0, "L", false, 0, "")
			}

			// Height used by this chart cell = title + image + legend rows
			legendRows := (len(ch.Segments) + 1) / 2
			cellH := 5 + imgH + 3 + float64(legendRows)*rowH + 6

			if i%2 == 0 {
				cix += imgW + 6
			} else {
				ciy += cellH
				cix = ml
				// Ensure we have enough space for the next row; add page if needed
				if ciy+cellH > 270 {
					doc.AddPage()
					ciy = doc.GetY()
				}
			}
		}
	}

	return doc
}

// ─── Small helpers ────────────────────────────────────────────────────────────

// pdfSectionTitle draws a labelled grey band (used as section separator).
func pdfSectionTitle(doc *fpdf.Fpdf, title string, x, w float64) {
	y := doc.GetY()
	doc.SetFillColor(pLightGray.r, pLightGray.g, pLightGray.b)
	doc.Rect(x, y, w, 6, "F")
	doc.SetDrawColor(pBorder.r, pBorder.g, pBorder.b)
	doc.SetLineWidth(0.15)
	doc.Rect(x, y, w, 6, "D")
	doc.SetFont("Helvetica", "B", 7.5)
	doc.SetTextColor(pTextGray.r, pTextGray.g, pTextGray.b)
	doc.SetXY(x+2, y)
	doc.CellFormat(w-4, 6, title, "", 1, "L", false, 0, "")
}

// ppct returns an integer percentage (rounds to nearest).
func ppct(count, total int) int {
	if total == 0 {
		return 0
	}
	return int(float64(count)/float64(total)*100 + 0.5)
}

// pdfTrunc truncates a string to maxRunes runes, appending ".." if cut.
func pdfTrunc(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes-2]) + ".."
}

// hexToRGB parses a CSS hex colour (#rrggbb or #rgb) into an rgbColor.
// Returns pGray on any parse failure.
func hexToRGB(h string) rgbColor {
	h = strings.TrimPrefix(h, "#")
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	if len(h) != 6 {
		return pGray
	}
	parse := func(s string) int {
		n := 0
		for _, c := range s {
			n <<= 4
			switch {
			case c >= '0' && c <= '9':
				n |= int(c - '0')
			case c >= 'a' && c <= 'f':
				n |= int(c-'a') + 10
			case c >= 'A' && c <= 'F':
				n |= int(c-'A') + 10
			}
		}
		return n
	}
	return rgbColor{parse(h[0:2]), parse(h[2:4]), parse(h[4:6])}
}

// isValidPNG checks the 8-byte PNG magic header.
func isValidPNG(data []byte) bool {
	if len(data) < 8 {
		return false
	}
	return data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G' &&
		data[4] == '\r' && data[5] == '\n' && data[6] == 0x1a && data[7] == '\n'
}

// pdfFilename builds a descriptive filename like AgriTwin_Solapur_Mohol_2025-06-10.pdf
func pdfFilename(req PDFRequest) string {
	parts := []string{"AgriTwin"}
	if req.DistrictName != "" {
		parts = append(parts, strings.ReplaceAll(req.DistrictName, " ", "_"))
	}
	if req.TalukaName != "" {
		parts = append(parts, strings.ReplaceAll(req.TalukaName, " ", "_"))
	}
	if req.VillageName != "" {
		parts = append(parts, strings.ReplaceAll(req.VillageName, " ", "_"))
	}
	parts = append(parts, time.Now().Format("2006-01-02"))
	return strings.Join(parts, "_") + ".pdf"
}
