package parser

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"

	"github.com/stretchr/testify/mock"
	"github.com/xuri/excelize/v2"
)

// Parser defines the interface for parsing Excel files.
type Parser interface {
	ParseXlsxToJson(file *multipart.FileHeader, columns []string) ([]map[string]interface{}, error)
}

// DefaultParser is the default implementation of the Parser interface.
type DefaultParser struct{}

// parseXlsx reads and parses the Excel file
func (p *DefaultParser) ParseXlsxToJson(
	file *multipart.FileHeader,
	columns []string,
) ([]map[string]interface{}, error) {
	// Buka file
	rows, err := getExcelRows(file)
	if err != nil {
		return nil, err
	}

	// Cari header
	header, headerRowIndex := findHeader(rows, columns)
	if headerRowIndex == -1 {
		return []map[string]interface{}{}, nil // Return array kosong jika tidak ada header
	}

	// Konversi data menjadi slice JSON
	return parseRows(rows, header, headerRowIndex, columns), nil
}

// getExcelRows membuka file dan membaca semua baris dari sheet pertama
func getExcelRows(file *multipart.FileHeader) ([][]string, error) {
	fileContent, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer fileContent.Close()

	f, err := excelize.OpenReader(fileContent)
	if err != nil {
		return nil, fmt.Errorf("failed to read Excel file: %v", err)
	}
	defer f.Close()

	sheetNames := f.GetSheetList()
	if len(sheetNames) == 0 {
		return nil, errors.New("no sheets found in Excel file")
	}

	rows, err := f.GetRows(sheetNames[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get rows from sheet %s: %v", sheetNames[0], err)
	}
	return rows, nil
}

// findHeader mencari baris header yang cocok dengan kolom
func findHeader(rows [][]string, columns []string) ([]string, int) {
	for i, row := range rows {
		if containsAny(row, columns) {
			return row, i
		}
	}
	return nil, -1
}

// parseRows mengonversi baris menjadi slice JSON berdasarkan header
func parseRows(
	rows [][]string,
	header []string,
	headerRowIndex int,
	columns []string,
) []map[string]interface{} {
	var result []map[string]interface{}
	for _, row := range rows[headerRowIndex+1:] { // Mulai dari baris setelah header
		rowData := make(map[string]interface{})
		for i, cell := range row {
			if i < len(header) && contains(columns, header[i]) {
				rowData[header[i]] = parseValue(cell)
			}
		}
		if len(rowData) > 0 {
			result = append(result, rowData)
		}
	}
	return result
}

// parseValue menentukan tipe data berdasarkan nilai string
func parseValue(value string) interface{} {
	if boolValue, err := strconv.ParseBool(value); err == nil {
		return boolValue
	}
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}
	return value
}

// containsAny memeriksa apakah salah satu item dalam `columns` ada di baris
func containsAny(row []string, columns []string) bool {
	for _, col := range row {
		for _, target := range columns {
			if col == target {
				return true
			}
		}
	}
	return false
}

// contains memeriksa apakah string ada di dalam slice
func contains(slice []string, item string) bool {
	for _, str := range slice {
		if str == item {
			return true
		}
	}
	return false
}

type MockParser struct {
	mock.Mock
}

func (m *MockParser) ParseXlsxToJson(
	file *multipart.FileHeader,
	columns []string,
) ([]map[string]interface{}, error) {
	args := m.Called(file, columns)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}
