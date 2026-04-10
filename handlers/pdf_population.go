package handlers

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
)

type populationPDFRow struct {
	ExternalFamilyID  string
	HouseNo           string
	HeadName          string
	TotalMembers      int
	MaleMembers       int
	FemaleMembers     int
	WorkingMembers    int
	LiterateMembers   int
	IlliterateMembers int
	StudentMembers    int
	DropoutMembers    int
	DisabledPersons   int
	HasDisability     int
	OccupationList    string
	BPLStatus         string
}

type occupationStat struct {
	Name  string
	Count int
}

type populationPDFStats struct {
	TotalHouseholds     int
	TotalPopulation     int
	MalePopulation      int
	FemalePopulation    int
	WorkingMembers      int
	DependentMembers    int
	WorkingHouseholds   int
	DependentHouseholds int
	LiteratePersons     int
	IlliteratePersons   int
	Students            int
	Dropouts            int
	BPLHouseholds       int
	NonBPLHouseholds    int
	DivyangHouseholds   int
	DisabledPersons     int
	Family1to2          int
	Family3to5          int
	Family6Plus         int
	TopOccupations      []occupationStat
}

type populationMetricBox struct {
	label string
	value string
	note  string
	color rgbColor
}

func (h *PopulationHandler) GeneratePopulationPDF(c *gin.Context) {
	var req PDFRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "detail": err.Error()})
		return
	}

	rows, err := h.queryPopulationRowsForPDF(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database query failed", "detail": err.Error()})
		return
	}

	stats := computePopulationPDFStats(rows)
	doc := buildPopulationPDF(req, rows, stats)

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF rendering failed", "detail": err.Error()})
		return
	}

	fname := populationPDFFilename(req)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fname))
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

func (h *PopulationHandler) queryPopulationRowsForPDF(req PDFRequest) ([]populationPDFRow, error) {
	clauses := []string{
		"f.LATITUDE IS NOT NULL",
		"f.LONGITUDE IS NOT NULL",
	}
	args := []interface{}{}

	if strings.TrimSpace(req.DistrictID) != "" {
		clauses = append(clauses, "CAST(f.DISTRICT_ID AS CHAR) = ?")
		args = append(args, strings.TrimSpace(req.DistrictID))
	}
	if strings.TrimSpace(req.TalukaID) != "" {
		clauses = append(clauses, "CAST(f.TALUKA_ID AS CHAR) = ?")
		args = append(args, strings.TrimSpace(req.TalukaID))
	}
	if strings.TrimSpace(req.VillageID) != "" {
		clauses = append(clauses, "CAST(f.VILLAGE_ID AS CHAR) = ?")
		args = append(args, strings.TrimSpace(req.VillageID))
	}

	where := strings.Join(clauses, " AND ")

	query := fmt.Sprintf(`
    SELECT
      COALESCE(CAST(f.EXTERNAL_FAMILY_ID AS CHAR), '') AS external_family_id,
      COALESCE(CAST(f.HOUSE_NO AS CHAR), '') AS house_no,
      COALESCE(TRIM(CONCAT(
        COALESCE(f.FIRST_NAME_HOUSEHOLD_HEAD, ''), ' ',
        COALESCE(f.MIDDLE_NAME_HOUSEHOLD_HEAD, ''), ' ',
        COALESCE(f.LAST_NAME_HOUSEHOLD_HEAD, '')
      )), '') AS head_name,
      COALESCE(fm_agg.total_members, 0) AS total_members,
      COALESCE(fm_agg.male_members, 0) AS male_members,
      COALESCE(fm_agg.female_members, 0) AS female_members,
      COALESCE(fm_agg.working_members, 0) AS working_members,
      COALESCE(fm_agg.literate_members, 0) AS literate_members,
      COALESCE(fm_agg.illiterate_members, 0) AS illiterate_members,
      COALESCE(fm_agg.student_members, 0) AS student_members,
      COALESCE(fm_agg.dropout_members, 0) AS dropout_members,
      COALESCE(fm_agg.disabled_persons, 0) AS disabled_persons,
      COALESCE(fm_agg.has_disability, 0) AS has_disability,
      COALESCE(fm_agg.occupation_list, '') AS occupation_list,
      COALESCE(TRIM(COALESCE(f.FAMILY_BELONG_BPL_CATEGORY, '')), '') AS bpl_status
    FROM FAMILY f
    LEFT JOIN (
      SELECT
        fm.EXTERNAL_FAMILY_ID,
        COUNT(fm.FAMILY_MEMBER_ID) AS total_members,
        SUM(CASE WHEN LOWER(TRIM(COALESCE(fm.GENDER, ''))) = 'male' THEN 1 ELSE 0 END) AS male_members,
        SUM(CASE WHEN LOWER(TRIM(COALESCE(fm.GENDER, ''))) = 'female' THEN 1 ELSE 0 END) AS female_members,
        SUM(CASE
          WHEN fm.OCCUPATION IS NOT NULL
            AND TRIM(fm.OCCUPATION) != ''
            AND LOWER(TRIM(fm.OCCUPATION)) NOT IN (
              'housewife',
              'homemaker',
              'student',
              'studying',
              'not applicable',
              'na',
              'n/a',
              'none',
              'unemployed',
              'child',
              'nil',
              'no work'
            )
          THEN 1
          ELSE 0
        END) AS working_members,
        SUM(CASE
          WHEN LOWER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL, ''))) = 'yes'
            OR (
              NULLIF(TRIM(COALESCE(fm.EDUCATION_STATUS, '')), '') IS NOT NULL
              AND LOWER(TRIM(COALESCE(fm.EDUCATION_STATUS, ''))) NOT IN ('illiterate', 'cannot read', 'no schooling')
            )
          THEN 1
          ELSE 0
        END) AS literate_members,
        SUM(CASE
          WHEN LOWER(TRIM(COALESCE(fm.EVER_ATTENDED_SCHOOL, ''))) = 'no'
            OR LOWER(TRIM(COALESCE(fm.EDUCATION_STATUS, ''))) IN ('illiterate', 'cannot read', 'no schooling')
          THEN 1
          ELSE 0
        END) AS illiterate_members,
        SUM(CASE
          WHEN UPPER(TRIM(COALESCE(fm.CURRENTLY_PURSUING_EDUCATION, ''))) = 'YES'
          THEN 1
          ELSE 0
        END) AS student_members,
        SUM(CASE
          WHEN UPPER(TRIM(COALESCE(fm.DROP_OUT, ''))) = 'YES'
            OR TRIM(COALESCE(fm.DROP_OUT, '')) IN ('1','2','3','4','5','6','7','8','9','10')
          THEN 1
          ELSE 0
        END) AS dropout_members,
        SUM(CASE
          WHEN UPPER(TRIM(COALESCE(fm.DIVYANG, ''))) = 'YES'
            OR NULLIF(TRIM(COALESCE(fm.DISABILITY, '')), '') IS NOT NULL
            OR NULLIF(TRIM(COALESCE(CAST(fm.DISABILITY_PERCENTAGE AS CHAR), '')), '') IS NOT NULL
          THEN 1
          ELSE 0
        END) AS disabled_persons,
        MAX(CASE
          WHEN UPPER(TRIM(COALESCE(fm.DIVYANG, ''))) = 'YES'
            OR NULLIF(TRIM(COALESCE(fm.DISABILITY, '')), '') IS NOT NULL
            OR NULLIF(TRIM(COALESCE(CAST(fm.DISABILITY_PERCENTAGE AS CHAR), '')), '') IS NOT NULL
          THEN 1
          ELSE 0
        END) AS has_disability,
        COALESCE(GROUP_CONCAT(
          DISTINCT CASE
            WHEN fm.OCCUPATION IS NOT NULL
              AND TRIM(fm.OCCUPATION) != ''
              AND LOWER(TRIM(fm.OCCUPATION)) NOT IN (
                'housewife',
                'homemaker',
                'student',
                'studying',
                'not applicable',
                'na',
                'n/a',
                'none',
                'unemployed',
                'child',
                'nil',
                'no work'
              )
            THEN TRIM(fm.OCCUPATION)
          END
          SEPARATOR '|'
        ), '') AS occupation_list
      FROM FAMILY_MEMBER fm
      GROUP BY fm.EXTERNAL_FAMILY_ID
    ) fm_agg ON fm_agg.EXTERNAL_FAMILY_ID = f.EXTERNAL_FAMILY_ID
    WHERE %s
    ORDER BY f.DISTRICT_ID, f.TALUKA_ID, f.VILLAGE_ID, f.HOUSE_NO, f.EXTERNAL_FAMILY_ID
  `, where)

	qrows, err := h.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer qrows.Close()

	out := []populationPDFRow{}
	for qrows.Next() {
		var r populationPDFRow
		if err := qrows.Scan(
			&r.ExternalFamilyID,
			&r.HouseNo,
			&r.HeadName,
			&r.TotalMembers,
			&r.MaleMembers,
			&r.FemaleMembers,
			&r.WorkingMembers,
			&r.LiterateMembers,
			&r.IlliterateMembers,
			&r.StudentMembers,
			&r.DropoutMembers,
			&r.DisabledPersons,
			&r.HasDisability,
			&r.OccupationList,
			&r.BPLStatus,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}

	return out, nil
}

func computePopulationPDFStats(rows []populationPDFRow) populationPDFStats {
	s := populationPDFStats{TotalHouseholds: len(rows)}
	occMap := map[string]int{}

	for _, row := range rows {
		s.TotalPopulation += row.TotalMembers
		s.MalePopulation += row.MaleMembers
		s.FemalePopulation += row.FemaleMembers
		s.WorkingMembers += row.WorkingMembers
		s.LiteratePersons += row.LiterateMembers
		s.IlliteratePersons += row.IlliterateMembers
		s.Students += row.StudentMembers
		s.Dropouts += row.DropoutMembers
		s.DisabledPersons += row.DisabledPersons

		if row.WorkingMembers > 0 {
			s.WorkingHouseholds++
		} else {
			s.DependentHouseholds++
		}

		if row.HasDisability == 1 {
			s.DivyangHouseholds++
		}

		members := row.TotalMembers
		if members <= 2 {
			s.Family1to2++
		} else if members <= 5 {
			s.Family3to5++
		} else {
			s.Family6Plus++
		}

		if isBPLPopulationHouse(row.BPLStatus) {
			s.BPLHouseholds++
		}

		for _, occ := range strings.Split(row.OccupationList, "|") {
			name := strings.TrimSpace(occ)
			if name == "" {
				continue
			}
			occMap[name]++
		}
	}

	s.DependentMembers = s.TotalPopulation - s.WorkingMembers
	if s.DependentMembers < 0 {
		s.DependentMembers = 0
	}
	s.NonBPLHouseholds = s.TotalHouseholds - s.BPLHouseholds
	if s.NonBPLHouseholds < 0 {
		s.NonBPLHouseholds = 0
	}

	occList := make([]occupationStat, 0, len(occMap))
	for name, count := range occMap {
		occList = append(occList, occupationStat{Name: name, Count: count})
	}
	sort.Slice(occList, func(i, j int) bool {
		if occList[i].Count == occList[j].Count {
			return strings.ToLower(occList[i].Name) < strings.ToLower(occList[j].Name)
		}
		return occList[i].Count > occList[j].Count
	})
	if len(occList) > 10 {
		occList = occList[:10]
	}
	s.TopOccupations = occList

	return s
}

func buildPopulationPDF(req PDFRequest, rows []populationPDFRow, s populationPDFStats) *fpdf.Fpdf {
	doc := fpdf.New("P", "mm", "A4", "")
	doc.SetMargins(14, 14, 14)
	doc.SetAutoPageBreak(true, 16)

	pageW, _ := doc.GetPageSize()
	ml, _, mr, _ := doc.GetMargins()
	uw := pageW - ml - mr

	generatedAt := time.Now().Format("02 Jan 2006  15:04")
	doc.SetFooterFunc(func() {
		doc.SetY(-12)
		doc.SetDrawColor(pBorder.r, pBorder.g, pBorder.b)
		doc.SetLineWidth(0.2)
		doc.Line(ml, doc.GetY(), ml+uw, doc.GetY())
		doc.SetFont("Helvetica", "I", 6.5)
		doc.SetTextColor(pGray.r, pGray.g, pGray.b)
		doc.CellFormat(uw/2, 5, "PopTwin - Population Digital Twin Report", "", 0, "L", false, 0, "")
		doc.CellFormat(uw/2, 5,
			fmt.Sprintf("Page %d  |  Generated %s", doc.PageNo(), generatedAt),
			"", 0, "R", false, 0, "")
	})

	doc.AddPage()

	doc.SetFillColor(pGreen.r, pGreen.g, pGreen.b)
	doc.Rect(ml, 14, uw, 8, "F")

	doc.SetXY(ml+2, 14)
	doc.SetFont("Helvetica", "B", 10)
	doc.SetTextColor(pWhite.r, pWhite.g, pWhite.b)
	doc.CellFormat(uw-55, 8, "PopTwin  |  Population Digital Twin", "", 0, "L", false, 0, "")

	doc.SetFont("Helvetica", "", 8)
	doc.CellFormat(53, 8, time.Now().Format("02 January 2006"), "", 0, "R", false, 0, "")

	doc.SetXY(ml, 24)
	doc.SetFont("Helvetica", "B", 17)
	doc.SetTextColor(pDark.r, pDark.g, pDark.b)
	doc.CellFormat(uw, 9, "Village Population Report", "", 1, "L", false, 0, "")

	districtName := strings.TrimSpace(req.DistrictName)
	if districtName == "" {
		districtName = "All"
	}
	talukaName := strings.TrimSpace(req.TalukaName)
	if talukaName == "" {
		talukaName = "All"
	}
	villageName := strings.TrimSpace(req.VillageName)
	if villageName == "" {
		villageName = "All"
	}

	doc.SetFont("Helvetica", "", 9)
	doc.SetTextColor(pTextGray.r, pTextGray.g, pTextGray.b)
	doc.CellFormat(uw, 5, fmt.Sprintf("District: %s   Taluka: %s   Village: %s", districtName, talukaName, villageName), "", 1, "L", false, 0, "")
	doc.CellFormat(uw, 5, fmt.Sprintf("Generated On: %s", time.Now().Format("02 Jan 2006 15:04")), "", 1, "L", false, 0, "")

	doc.SetDrawColor(pGreen.r, pGreen.g, pGreen.b)
	doc.SetLineWidth(0.4)
	divY := doc.GetY() + 2
	doc.Line(ml, divY, ml+uw, divY)
	doc.SetY(divY + 4)

	ensureSpace(doc, 40)
	pdfSectionTitle(doc, "SUMMARY STATISTICS", ml, uw)
	doc.Ln(6)

	type statBox struct {
		label string
		value string
		note  string
		color rgbColor
	}

	boxes := []statBox{
		{"Total Households", fmt.Sprintf("%d", s.TotalHouseholds), "survey completed", pGreen},
		{"Total Population", fmt.Sprintf("%d", s.TotalPopulation), "all members", pBlue},
		{"Male Population", fmt.Sprintf("%d", s.MalePopulation), fmt.Sprintf("%d%%", ppct(s.MalePopulation, s.TotalPopulation)), pTeal},
		{"Female Population", fmt.Sprintf("%d", s.FemalePopulation), fmt.Sprintf("%d%%", ppct(s.FemalePopulation, s.TotalPopulation)), pPurple},
		{"Working Members", fmt.Sprintf("%d", s.WorkingMembers), fmt.Sprintf("%d%%", ppct(s.WorkingMembers, s.TotalPopulation)), pAmber},
		{"Dependent Members", fmt.Sprintf("%d", s.DependentMembers), fmt.Sprintf("%d%%", ppct(s.DependentMembers, s.TotalPopulation)), pRed},
	}

	bw := (uw - 2*3) / 3
	y0 := doc.GetY()
	for i, box := range boxes {
		col := float64(i % 3)
		row := float64(i / 3)
		bx := ml + col*(bw+3)
		by := y0 + row*23

		doc.SetFillColor(pLightGray.r, pLightGray.g, pLightGray.b)
		doc.Rect(bx, by, bw, 21, "F")
		doc.SetFillColor(box.color.r, box.color.g, box.color.b)
		doc.Rect(bx, by, bw, 1.5, "F")

		doc.SetXY(bx, by+3)
		doc.SetFont("Helvetica", "B", 13)
		doc.SetTextColor(pDark.r, pDark.g, pDark.b)
		doc.CellFormat(bw, 7, box.value, "", 1, "C", false, 0, "")

		doc.SetX(bx)
		doc.SetFont("Helvetica", "B", 7)
		doc.SetTextColor(pTextGray.r, pTextGray.g, pTextGray.b)
		doc.CellFormat(bw, 4, box.label, "", 1, "C", false, 0, "")

		doc.SetX(bx)
		doc.SetFont("Helvetica", "", 6.5)
		doc.SetTextColor(box.color.r, box.color.g, box.color.b)
		doc.CellFormat(bw, 4, box.note, "", 1, "C", false, 0, "")
	}
	doc.Ln(10)

	ensureSpace(doc, 30)
	pdfSectionTitle(doc, "COLOUR LEGEND", ml, uw)
	doc.Ln(6)

	type legendItem struct {
		label string
		color rgbColor
	}

	legend := []legendItem{
		{"Population Density: 1-2 members", rgbColor{34, 197, 94}},
		{"Population Density: 3-5 members", rgbColor{245, 158, 11}},
		{"Population Density: 6+ members", rgbColor{239, 68, 68}},
		{"BPL household", rgbColor{239, 68, 68}},
		{"Non-BPL household", rgbColor{22, 163, 74}},
		{"Working household", rgbColor{22, 163, 74}},
		{"No earning member", rgbColor{245, 158, 11}},
		{"Disability present", rgbColor{123, 31, 162}},
		{"No disability", rgbColor{22, 163, 74}},
		{"Literate", rgbColor{59, 130, 246}},
		{"Illiterate", pGray},
	}

	ly0 := doc.GetY()
	colW := uw / 2
	for i, leg := range legend {
		col := i % 2
		row := i / 2
		lx := ml + float64(col)*colW
		ly := ly0 + float64(row)*5.5
		doc.SetFillColor(leg.color.r, leg.color.g, leg.color.b)
		doc.Rect(lx, ly+1, 3, 3, "F")
		doc.SetFont("Helvetica", "", 7.3)
		doc.SetTextColor(pDark.r, pDark.g, pDark.b)
		doc.SetXY(lx+4.5, ly)
		doc.CellFormat(colW-5, 5, leg.label, "", 0, "L", false, 0, "")
	}
	doc.SetY(ly0 + float64((len(legend)+1)/2)*5.5 + 10)

	ensureSpace(doc, 30)
	pdfSectionTitle(doc, "POPULATION OVERVIEW", ml, uw)
	doc.Ln(6)
	drawRatioBar(doc, ml, uw, "Gender Ratio", s.MalePopulation, s.FemalePopulation, "Male", "Female", pBlue, rgbColor{236, 72, 153})
	drawRatioBar(doc, ml, uw, "Family Size Distribution", s.Family1to2, s.Family3to5+s.Family6Plus, "1-2 members", "3+ members", rgbColor{34, 197, 94}, rgbColor{245, 158, 11})
	drawRatioBar(doc, ml, uw, "Working vs Dependent", s.WorkingMembers, s.DependentMembers, "Working", "Dependent", rgbColor{22, 163, 74}, rgbColor{245, 158, 11})

	doc.AddPage()
	pdfSectionTitle(doc, "EDUCATION OVERVIEW", ml, uw)
	doc.Ln(8)
	cardWidth := (uw - 10) / 2
	cardHeight := 18.0
	startX := ml
	startY := doc.GetY()

	drawPopulationMetricCard(doc, startX, startY, cardWidth, cardHeight, "Literate", fmt.Sprintf("%d", s.LiteratePersons), fmt.Sprintf("%d%%", ppct(s.LiteratePersons, s.TotalPopulation)), rgbColor{59, 130, 246})
	drawPopulationMetricCard(doc, startX+cardWidth+10, startY, cardWidth, cardHeight, "Illiterate", fmt.Sprintf("%d", s.IlliteratePersons), fmt.Sprintf("%d%%", ppct(s.IlliteratePersons, s.TotalPopulation)), pGray)

	doc.SetY(startY + cardHeight + 8)
	startY = doc.GetY()
	drawPopulationMetricCard(doc, startX, startY, cardWidth, cardHeight, "Students", fmt.Sprintf("%d", s.Students), "currently pursuing", rgbColor{5, 150, 105})
	drawPopulationMetricCard(doc, startX+cardWidth+10, startY, cardWidth, cardHeight, "Dropouts", fmt.Sprintf("%d", s.Dropouts), "attention needed", pRed)
	doc.Ln(cardHeight + 12)

	ensureSpace(doc, 22)
	pdfSectionTitle(doc, "EMPLOYMENT OVERVIEW", ml, uw)
	doc.Ln(6)
	drawKVMetrics(doc, ml, uw, []populationMetricBox{
		{"Working Households", fmt.Sprintf("%d", s.WorkingHouseholds), fmt.Sprintf("%d%%", ppct(s.WorkingHouseholds, s.TotalHouseholds)), rgbColor{22, 163, 74}},
		{"Dependent Households", fmt.Sprintf("%d", s.DependentHouseholds), fmt.Sprintf("%d%%", ppct(s.DependentHouseholds, s.TotalHouseholds)), rgbColor{245, 158, 11}},
	})

	doc.SetFont("Helvetica", "B", 7.2)
	doc.SetTextColor(pTextGray.r, pTextGray.g, pTextGray.b)
	doc.CellFormat(uw, 5, "Top Occupations", "", 1, "L", false, 0, "")
	doc.SetFont("Helvetica", "", 7)
	if len(s.TopOccupations) == 0 {
		doc.SetTextColor(pGray.r, pGray.g, pGray.b)
		doc.CellFormat(uw, 4.5, "No occupation data available", "", 1, "L", false, 0, "")
	} else {
		for _, occ := range s.TopOccupations {
			doc.SetTextColor(pDark.r, pDark.g, pDark.b)
			doc.CellFormat(uw, 4.5, fmt.Sprintf("- %s (%d)", occ.Name, occ.Count), "", 1, "L", false, 0, "")
		}
	}
	doc.Ln(10)

	ensureSpace(doc, 28)
	pdfSectionTitle(doc, "BPL DISTRIBUTION", ml, uw)
	doc.Ln(6)
	drawRatioBar(doc, ml, uw, "BPL vs Non-BPL Households", s.BPLHouseholds, s.NonBPLHouseholds, "BPL", "Non-BPL", rgbColor{239, 68, 68}, rgbColor{22, 163, 74})

	ensureSpace(doc, 28)
	pdfSectionTitle(doc, "DIVYANG SUMMARY", ml, uw)
	doc.Ln(6)
	drawKVMetrics(doc, ml, uw, []populationMetricBox{
		{"Households with Disability", fmt.Sprintf("%d", s.DivyangHouseholds), fmt.Sprintf("%d%%", ppct(s.DivyangHouseholds, s.TotalHouseholds)), rgbColor{123, 31, 162}},
		{"Total Disabled Persons", fmt.Sprintf("%d", s.DisabledPersons), fmt.Sprintf("%d%%", ppct(s.DisabledPersons, s.TotalPopulation)), rgbColor{123, 31, 162}},
	})

	// Keep charts on a dedicated page to avoid bottom-of-page collisions.
	ensureSpace(doc, 40)
	doc.AddPage()
	pdfSectionTitle(doc, "POPULATION CHARTS", ml, uw)
	doc.Ln(8)

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
	}

	if len(validCharts) > 0 {
		chartsByTitle := map[string]PDFChartImage{}
		for _, ch := range validCharts {
			chartsByTitle[ch.Title] = ch
		}

		desiredOrder := []string{
			"Gender Ratio Chart",
			"Employment Distribution Chart",
			"BPL Distribution Chart",
			"Family Size Distribution Chart",
		}

		orderedCharts := make([]PDFChartImage, 0, 4)
		for _, title := range desiredOrder {
			if ch, ok := chartsByTitle[title]; ok {
				orderedCharts = append(orderedCharts, ch)
			}
		}
		if len(orderedCharts) == 0 {
			orderedCharts = validCharts
			if len(orderedCharts) > 4 {
				orderedCharts = orderedCharts[:4]
			}
		}

		chartSize := 60.0
		startX := 20.0
		startY := doc.GetY()
		rowGap := chartSize + 22

		for i, ch := range orderedCharts {
			col := i % 2
			row := i / 2
			x := startX + float64(col)*90
			y := startY + float64(row)*rowGap

			imgBytes, _ := base64.StdEncoding.DecodeString(ch.Image)
			key := fmt.Sprintf("_pc_%d", i)
			doc.RegisterImageOptionsReader(
				key,
				fpdf.ImageOptions{ImageType: "PNG"},
				bytes.NewReader(imgBytes),
			)
			if !doc.Ok() {
				break
			}

			doc.SetFont("Helvetica", "B", 8.5)
			doc.SetTextColor(pDark.r, pDark.g, pDark.b)
			doc.SetXY(x, y)
			doc.CellFormat(chartSize+30, 5, ch.Title, "", 0, "C", false, 0, "")

			imgX := x + 15
			doc.ImageOptions(key, imgX, y+6, chartSize, chartSize, false,
				fpdf.ImageOptions{ImageType: "PNG"}, 0, "")

			legendText := chartPercentCaption(ch.Title, s)
			doc.SetFont("Helvetica", "", 7)
			doc.SetTextColor(pTextGray.r, pTextGray.g, pTextGray.b)
			doc.SetXY(x, y+chartSize+9)
			doc.MultiCell(chartSize+30, 4, legendText, "", "C", false)
		}

		rowsUsed := (len(orderedCharts) + 1) / 2
		doc.SetY(startY + float64(rowsUsed)*rowGap + 12)
	}

	ensureSpace(doc, 40)
	doc.AddPage()
	pdfSectionTitle(doc, fmt.Sprintf("HOUSEHOLD DATA  (%d records)", s.TotalHouseholds), ml, uw)
	doc.Ln(6)

	type tcol struct {
		hdr   string
		width float64
		align string
	}
	tcols := []tcol{
		{"House No", 18, "L"},
		{"Head Name", 62, "L"},
		{"Total", 16, "C"},
		{"Male", 16, "C"},
		{"Female", 16, "C"},
		{"Working", 18, "C"},
		{"BPL", 18, "L"},
		{"Disability", 18, "L"},
	}

	drawHeader := func() {
		doc.SetFillColor(pDarkGreen.r, pDarkGreen.g, pDarkGreen.b)
		doc.SetTextColor(pWhite.r, pWhite.g, pWhite.b)
		doc.SetFont("Helvetica", "B", 7)
		for _, tc := range tcols {
			doc.CellFormat(tc.width, 6, "  "+tc.hdr, "0", 0, tc.align, true, 0, "")
		}
		doc.Ln(6)
	}

	drawHeader()
	doc.SetFont("Helvetica", "", 6.8)
	headColWidth := tcols[1].width

	for _, row := range rows {
		name := strings.TrimSpace(row.HeadName)
		if name == "" {
			name = "Household"
		}

		nameLines := doc.SplitLines([]byte(name), headColWidth-2)
		rowH := float64(len(nameLines))*3.8 + 1.6
		if rowH < 5 {
			rowH = 5
		}

		ensureSpace(doc, rowH+3)
		if doc.GetY()+rowH > 275 {
			doc.AddPage()
			pdfSectionTitle(doc, "HOUSEHOLD DATA (continued)", ml, uw)
			doc.Ln(6)
			drawHeader()
			doc.SetFont("Helvetica", "", 6.8)
		}

		y := doc.GetY()
		if (int(y*10) % 2) == 0 {
			doc.SetFillColor(pLightGray.r, pLightGray.g, pLightGray.b)
		} else {
			doc.SetFillColor(pWhite.r, pWhite.g, pWhite.b)
		}
		doc.Rect(ml, y, uw, rowH, "F")

		x := ml
		doc.SetTextColor(pDark.r, pDark.g, pDark.b)
		doc.SetXY(x, y)
		doc.CellFormat(tcols[0].width, rowH, "  "+strings.TrimSpace(row.HouseNo), "0", 0, tcols[0].align, false, 0, "")
		x += tcols[0].width

		doc.SetXY(x+1, y+0.8)
		doc.MultiCell(tcols[1].width-2, 3.8, name, "", "L", false)
		x += tcols[1].width

		doc.SetXY(x, y)
		doc.CellFormat(tcols[2].width, rowH, fmt.Sprintf("  %d", row.TotalMembers), "0", 0, tcols[2].align, false, 0, "")
		x += tcols[2].width
		doc.CellFormat(tcols[3].width, rowH, fmt.Sprintf("  %d", row.MaleMembers), "0", 0, tcols[3].align, false, 0, "")
		x += tcols[3].width
		doc.CellFormat(tcols[4].width, rowH, fmt.Sprintf("  %d", row.FemaleMembers), "0", 0, tcols[4].align, false, 0, "")
		x += tcols[4].width

		if row.WorkingMembers > 0 {
			doc.SetTextColor(pGreen.r, pGreen.g, pGreen.b)
		} else {
			doc.SetTextColor(pAmber.r, pAmber.g, pAmber.b)
		}
		doc.CellFormat(tcols[5].width, rowH, fmt.Sprintf("  %d", row.WorkingMembers), "0", 0, tcols[5].align, false, 0, "")
		x += tcols[5].width

		bplLabel := populationBPLLabel(row.BPLStatus)
		if bplLabel == "BPL" {
			doc.SetTextColor(pRed.r, pRed.g, pRed.b)
		} else {
			doc.SetTextColor(pGreen.r, pGreen.g, pGreen.b)
		}
		doc.CellFormat(tcols[6].width, rowH, "  "+bplLabel, "0", 0, tcols[6].align, false, 0, "")
		x += tcols[6].width

		disability := "No"
		if row.HasDisability == 1 {
			disability = "Yes"
			doc.SetTextColor(123, 31, 162)
		} else {
			doc.SetTextColor(pGreen.r, pGreen.g, pGreen.b)
		}
		doc.CellFormat(tcols[7].width, rowH, "  "+disability, "0", 0, tcols[7].align, false, 0, "")

		doc.SetY(y + rowH)
	}

	return doc
}

func drawKVMetrics(doc *fpdf.Fpdf, ml, uw float64, boxes []populationMetricBox) {
	if len(boxes) == 0 {
		return
	}

	cols := 2
	if len(boxes) == 1 {
		cols = 1
	}
	bw := (uw - float64(cols-1)*3) / float64(cols)
	y0 := doc.GetY()

	for i, box := range boxes {
		col := float64(i % cols)
		row := float64(i / cols)
		bx := ml + col*(bw+3)
		by := y0 + row*18

		doc.SetFillColor(pLightGray.r, pLightGray.g, pLightGray.b)
		doc.Rect(bx, by, bw, 16, "F")
		doc.SetFillColor(box.color.r, box.color.g, box.color.b)
		doc.Rect(bx, by, bw, 1.5, "F")

		doc.SetXY(bx, by+2.5)
		doc.SetFont("Helvetica", "B", 11)
		doc.SetTextColor(pDark.r, pDark.g, pDark.b)
		doc.CellFormat(bw, 6, box.value, "", 1, "C", false, 0, "")

		doc.SetX(bx)
		doc.SetFont("Helvetica", "B", 7)
		doc.SetTextColor(pTextGray.r, pTextGray.g, pTextGray.b)
		doc.CellFormat(bw, 3.5, box.label, "", 1, "C", false, 0, "")

		doc.SetX(bx)
		doc.SetFont("Helvetica", "", 6.3)
		doc.SetTextColor(box.color.r, box.color.g, box.color.b)
		doc.CellFormat(bw, 3.5, box.note, "", 1, "C", false, 0, "")
	}

	rows := (len(boxes) + cols - 1) / cols
	doc.SetY(y0 + float64(rows)*18)
}

func drawPopulationMetricCard(doc *fpdf.Fpdf, x, y, w, h float64, label, value, note string, color rgbColor) {
	doc.SetFillColor(pLightGray.r, pLightGray.g, pLightGray.b)
	doc.Rect(x, y, w, h, "F")
	doc.SetFillColor(color.r, color.g, color.b)
	doc.Rect(x, y, w, 1.5, "F")

	doc.SetXY(x, y+2.5)
	doc.SetFont("Helvetica", "B", 11)
	doc.SetTextColor(pDark.r, pDark.g, pDark.b)
	doc.CellFormat(w, 6, value, "", 1, "C", false, 0, "")

	doc.SetX(x)
	doc.SetFont("Helvetica", "B", 7)
	doc.SetTextColor(pTextGray.r, pTextGray.g, pTextGray.b)
	doc.CellFormat(w, 3.5, label, "", 1, "C", false, 0, "")

	doc.SetX(x)
	doc.SetFont("Helvetica", "", 6.3)
	doc.SetTextColor(color.r, color.g, color.b)
	doc.CellFormat(w, 3.5, note, "", 1, "C", false, 0, "")
}

func drawRatioBar(doc *fpdf.Fpdf, ml, uw float64, title string, leftCount, rightCount int, leftLabel, rightLabel string, leftColor, rightColor rgbColor) {
	total := leftCount + rightCount

	doc.SetFont("Helvetica", "B", 7.5)
	doc.SetTextColor(pDark.r, pDark.g, pDark.b)
	doc.CellFormat(uw, 4.5, title, "", 1, "L", false, 0, "")

	barH := 5.2
	barX := ml
	barY := doc.GetY()
	doc.SetFillColor(pBorder.r, pBorder.g, pBorder.b)
	doc.Rect(barX, barY, uw, barH, "F")

	if total > 0 {
		leftW := (float64(leftCount) / float64(total)) * uw
		rightW := uw - leftW

		doc.SetFillColor(leftColor.r, leftColor.g, leftColor.b)
		doc.Rect(barX, barY, leftW, barH, "F")
		doc.SetFillColor(rightColor.r, rightColor.g, rightColor.b)
		doc.Rect(barX+leftW, barY, rightW, barH, "F")
	}

	doc.Ln(barH + 2)

	doc.SetFont("Helvetica", "", 6.8)
	doc.SetFillColor(leftColor.r, leftColor.g, leftColor.b)
	doc.Rect(ml, doc.GetY()+1, 2.5, 2.5, "F")
	doc.SetTextColor(pDark.r, pDark.g, pDark.b)
	doc.SetXY(ml+3.5, doc.GetY())
	doc.CellFormat(uw/2-4, 4.5, fmt.Sprintf("%s: %d (%d%%)", leftLabel, leftCount, ppct(leftCount, total)), "", 0, "L", false, 0, "")

	rx := ml + uw/2
	doc.SetFillColor(rightColor.r, rightColor.g, rightColor.b)
	doc.Rect(rx, doc.GetY()+1, 2.5, 2.5, "F")
	doc.SetXY(rx+3.5, doc.GetY())
	doc.CellFormat(uw/2-4, 4.5, fmt.Sprintf("%s: %d (%d%%)", rightLabel, rightCount, ppct(rightCount, total)), "", 1, "L", false, 0, "")
	doc.Ln(4)
}

func ensureSpace(doc *fpdf.Fpdf, neededHeight float64) {
	_, pageHeight := doc.GetPageSize()
	_, y := doc.GetXY()
	if y+neededHeight > pageHeight-20 {
		doc.AddPage()
	}
}

func chartPercentCaption(title string, s populationPDFStats) string {
	switch title {
	case "Gender Ratio Chart":
		return fmt.Sprintf("Male %d%%   Female %d%%", ppct(s.MalePopulation, s.TotalPopulation), ppct(s.FemalePopulation, s.TotalPopulation))
	case "Employment Distribution Chart":
		return fmt.Sprintf("Working %d%%   Dependent %d%%", ppct(s.WorkingMembers, s.TotalPopulation), ppct(s.DependentMembers, s.TotalPopulation))
	case "BPL Distribution Chart":
		return fmt.Sprintf("BPL %d%%   Non-BPL %d%%", ppct(s.BPLHouseholds, s.TotalHouseholds), ppct(s.NonBPLHouseholds, s.TotalHouseholds))
	case "Family Size Distribution Chart":
		return fmt.Sprintf("1-2 %d%%   3-5 %d%%   6+ %d%%", ppct(s.Family1to2, s.TotalHouseholds), ppct(s.Family3to5, s.TotalHouseholds), ppct(s.Family6Plus, s.TotalHouseholds))
	default:
		return ""
	}
}

func isBPLPopulationHouse(raw string) bool {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" {
		return false
	}
	if strings.Contains(v, "non-bpl") || v == "apl" || v == "no" {
		return false
	}
	if strings.Contains(v, "bpl") || v == "yes" {
		return true
	}
	return false
}

func populationBPLLabel(raw string) string {
	if isBPLPopulationHouse(raw) {
		return "BPL"
	}
	return "Non-BPL"
}

func populationPDFFilename(req PDFRequest) string {
	parts := []string{"PopTwin"}
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

var _ = sql.ErrNoRows
