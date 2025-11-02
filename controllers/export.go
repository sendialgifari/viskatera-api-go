package controllers

import (
	"fmt"
	"net/http"
	"time"
	"viskatera-api-go/config"
	"viskatera-api-go/models"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

// ExportVisasExcel godoc
// @Summary Export visas to Excel
// @Description Export all visas filtered by customer role purchases to Excel format
// @Tags Export
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Success 200 {file} file
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /exports/visas/excel [get]
func ExportVisasExcel(c *gin.Context) {
	// Get all visas with purchases by customer role users
	var purchases []models.VisaPurchase
	query := config.DB.
		Preload("User", "role = ?", "customer").
		Preload("Visa").
		Preload("VisaOption").
		Joins("JOIN users ON visa_purchases.user_id = users.id").
		Where("users.role = ?", "customer")

	if err := query.Find(&purchases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to fetch visa data",
			"DATABASE_ERROR",
			err.Error(),
		))
		return
	}

	// Create Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetName := "Visas"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to create Excel sheet",
			"EXCEL_ERROR",
			err.Error(),
		))
		return
	}

	// Set active sheet
	f.SetActiveSheet(index)

	// Set headers
	headers := []string{"ID", "User Name", "User Email", "Country", "Visa Type", "Description", "Price", "Option", "Total Price", "Status", "Created At"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, getHeaderStyle(f))
	}

	// Add data
	for i, purchase := range purchases {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), purchase.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), purchase.User.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), purchase.User.Email)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), purchase.Visa.Country)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), purchase.Visa.Type)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), purchase.Visa.Description)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), purchase.Visa.Price)

		optionName := ""
		if purchase.VisaOption != nil {
			optionName = purchase.VisaOption.Name
		}
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), optionName)

		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), purchase.TotalPrice)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), purchase.Status)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), purchase.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Set column widths
	for i := range headers {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 20)
	}

	// Set filename
	filename := fmt.Sprintf("visas_export_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to write Excel file",
			"EXCEL_ERROR",
			err.Error(),
		))
		return
	}
}

// ExportPurchasesExcel godoc
// @Summary Export purchases to Excel
// @Description Export all purchases by customers to Excel format
// @Tags Export
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Success 200 {file} file
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /exports/purchases/excel [get]
func ExportPurchasesExcel(c *gin.Context) {
	// Get all purchases by customer role users
	var purchases []models.VisaPurchase
	query := config.DB.
		Preload("User", "role = ?", "customer").
		Preload("Visa").
		Preload("VisaOption").
		Joins("JOIN users ON visa_purchases.user_id = users.id").
		Where("users.role = ?", "customer")

	if err := query.Find(&purchases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to fetch purchase data",
			"DATABASE_ERROR",
			err.Error(),
		))
		return
	}

	// Create Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetName := "Purchases"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to create Excel sheet",
			"EXCEL_ERROR",
			err.Error(),
		))
		return
	}

	f.SetActiveSheet(index)

	// Set headers
	headers := []string{"Purchase ID", "User ID", "User Name", "User Email", "Visa ID", "Country", "Visa Type", "Total Price", "Status", "Created At"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, getHeaderStyle(f))
	}

	// Add data
	for i, purchase := range purchases {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), purchase.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), purchase.UserID)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), purchase.User.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), purchase.User.Email)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), purchase.VisaID)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), purchase.Visa.Country)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), purchase.Visa.Type)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), purchase.TotalPrice)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), purchase.Status)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), purchase.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Set column widths
	for i := range headers {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 20)
	}

	filename := fmt.Sprintf("purchases_export_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to write Excel file",
			"EXCEL_ERROR",
			err.Error(),
		))
		return
	}
}

// ExportVisasPDF godoc
// @Summary Export visas to PDF
// @Description Export all visas filtered by customer role purchases to PDF format
// @Tags Export
// @Accept json
// @Produce application/pdf
// @Security BearerAuth
// @Success 200 {file} file
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /exports/visas/pdf [get]
func ExportVisasPDF(c *gin.Context) {
	var purchases []models.VisaPurchase
	query := config.DB.
		Preload("User", "role = ?", "customer").
		Preload("Visa").
		Preload("VisaOption").
		Joins("JOIN users ON visa_purchases.user_id = users.id").
		Where("users.role = ?", "customer")

	if err := query.Find(&purchases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to fetch visa data",
			"DATABASE_ERROR",
			err.Error(),
		))
		return
	}

	// Create PDF
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(280, 10, "Visa Export Report")
	pdf.Ln(15)

	// Table header
	pdf.SetFont("Arial", "B", 8)
	headers := []string{"ID", "User", "Email", "Country", "Type", "Price", "Total", "Status", "Date"}
	widths := []float64{15, 40, 50, 35, 35, 25, 25, 20, 40}

	for i, header := range headers {
		pdf.CellFormat(widths[i], 10, header, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	// Table data
	pdf.SetFont("Arial", "", 7)
	for _, purchase := range purchases {
		row := []string{
			fmt.Sprintf("%d", purchase.ID),
			purchase.User.Name,
			purchase.User.Email,
			purchase.Visa.Country,
			purchase.Visa.Type,
			fmt.Sprintf("%.2f", purchase.Visa.Price),
			fmt.Sprintf("%.2f", purchase.TotalPrice),
			purchase.Status,
			purchase.CreatedAt.Format("2006-01-02"),
		}

		for i, data := range row {
			pdf.CellFormat(widths[i], 8, data, "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}

	filename := fmt.Sprintf("visas_export_%s.pdf", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to generate PDF",
			"PDF_ERROR",
			err.Error(),
		))
		return
	}
}

// ExportPurchasesPDF godoc
// @Summary Export purchases to PDF
// @Description Export all purchases by customers to PDF format
// @Tags Export
// @Accept json
// @Produce application/pdf
// @Security BearerAuth
// @Success 200 {file} file
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /exports/purchases/pdf [get]
func ExportPurchasesPDF(c *gin.Context) {
	var purchases []models.VisaPurchase
	query := config.DB.
		Preload("User", "role = ?", "customer").
		Preload("Visa").
		Preload("VisaOption").
		Joins("JOIN users ON visa_purchases.user_id = users.id").
		Where("users.role = ?", "customer")

	if err := query.Find(&purchases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to fetch purchase data",
			"DATABASE_ERROR",
			err.Error(),
		))
		return
	}

	// Create PDF
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(280, 10, "Purchase Export Report")
	pdf.Ln(15)

	// Table header
	pdf.SetFont("Arial", "B", 8)
	headers := []string{"ID", "User ID", "User Name", "Email", "Visa ID", "Country", "Type", "Total Price", "Status", "Date"}
	widths := []float64{12, 15, 40, 50, 12, 30, 30, 30, 25, 40}

	for i, header := range headers {
		pdf.CellFormat(widths[i], 10, header, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	// Table data
	pdf.SetFont("Arial", "", 7)
	for _, purchase := range purchases {
		row := []string{
			fmt.Sprintf("%d", purchase.ID),
			fmt.Sprintf("%d", purchase.UserID),
			purchase.User.Name,
			purchase.User.Email,
			fmt.Sprintf("%d", purchase.VisaID),
			purchase.Visa.Country,
			purchase.Visa.Type,
			fmt.Sprintf("%.2f", purchase.TotalPrice),
			purchase.Status,
			purchase.CreatedAt.Format("2006-01-02"),
		}

		for i, data := range row {
			pdf.CellFormat(widths[i], 8, data, "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}

	filename := fmt.Sprintf("purchases_export_%s.pdf", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to generate PDF",
			"PDF_ERROR",
			err.Error(),
		))
		return
	}
}

func getHeaderStyle(f *excelize.File) int {
	styleID, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D3D3D3"},
			Pattern: 1,
		},
	})
	if err != nil {
		return 0
	}
	return styleID
}
