package main

import (
	"bytes"
	"fmt"
	"github.com/jung-kurt/gofpdf"
)

func GenerateReportPDF(tasks []*Task) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Заголовок
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Link status report:")
	pdf.Ln(12)

	// Информация по генерации
	pdf.SetFont("Arial", "I", 10)
	pdf.Cell(0, 8, "automatic generation")
	pdf.Ln(15)

	// Для каждой задачи
	for _, task := range tasks {
		// Название задачи
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, fmt.Sprintf("Task pool #%d", task.ID))
		pdf.Ln(8)

		// Ссылки
		pdf.SetFont("Arial", "", 10)
		for link, status := range task.Links {
			pdf.Cell(0, 6, fmt.Sprintf("- %s - %s", link, status))
			pdf.Ln(6)
		}

		pdf.Ln(10)
	}

	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 10, fmt.Sprintf("total tasks: %d", len(tasks)))

	// Создание буфера и запись ПДФ
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("PDF generation error: %v", err)
	}

	return buf.Bytes(), nil
}
