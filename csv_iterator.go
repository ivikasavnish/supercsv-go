// Package supercsv provides a flexible, type-safe CSV parser for Go using struct annotations.
//
// This package allows you to parse CSV data into any struct type using Go generics.
// Simply define your struct with csv:"column_name" tags and the iterator will
// automatically map CSV columns to struct fields with full type safety.
//
// Example usage:
//
//	type Person struct {
//	    Name   string   `csv:"name,required"`
//	    Age    int      `csv:"age"`
//	    Email  string   `csv:"email,required"`
//	    Salary *float64 `csv:"salary"`        // Optional field
//	    Active bool     `csv:"active"`
//	}
//
//	// From file
//	iterator, err := supercsv.NewFromFile[Person]("data.csv")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer iterator.Close()
//
//	// Iterate through rows
//	for {
//	    person, err := iterator.Next()
//	    if err == io.EOF {
//	        break
//	    }
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Printf("Name: %s, Age: %d\n", person.Name, person.Age)
//	}
package supercsv

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// CSVIterator provides generic CSV parsing with struct annotations
type CSVIterator[T any] struct {
	reader     *csv.Reader
	closer     io.Closer
	headers    []string
	fieldMap   map[string]int
	structType reflect.Type
	fieldInfo  []fieldInfo
}

type fieldInfo struct {
	fieldIndex int
	csvColumn  string
	fieldType  reflect.Type
	required   bool
}

// NewFromFile creates a CSV iterator from a file path
func NewFromFile[T any](filepath string) (*CSVIterator[T], error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return newIterator[T](file, file, ',')
}

// NewFromFileWithDelimiter creates a CSV iterator from a file path with custom delimiter
func NewFromFileWithDelimiter[T any](filepath string, delimiter rune) (*CSVIterator[T], error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return newIterator[T](file, file, delimiter)
}

// NewFromURL creates a CSV iterator from a URL
func NewFromURL[T any](url string) (*CSVIterator[T], error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	return newIterator[T](resp.Body, resp.Body, ',')
}

// NewFromURLWithDelimiter creates a CSV iterator from a URL with custom delimiter
func NewFromURLWithDelimiter[T any](url string, delimiter rune) (*CSVIterator[T], error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	return newIterator[T](resp.Body, resp.Body, delimiter)
}

// NewFromReader creates a CSV iterator from an io.Reader
func NewFromReader[T any](reader io.Reader) (*CSVIterator[T], error) {
	return newIterator[T](reader, nil, ',')
}

// NewFromReaderWithDelimiter creates a CSV iterator from an io.Reader with custom delimiter
func NewFromReaderWithDelimiter[T any](reader io.Reader, delimiter rune) (*CSVIterator[T], error) {
	return newIterator[T](reader, nil, delimiter)
}

func newIterator[T any](reader io.Reader, closer io.Closer, delimiter rune) (*CSVIterator[T], error) {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = delimiter           // Set custom delimiter
	csvReader.FieldsPerRecord = -1        // Allow variable number of fields

	// Read headers
	headers, err := csvReader.Read()
	if err != nil {
		if closer != nil {
			closer.Close()
		}
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}

	// Create field mapping
	fieldMap := make(map[string]int)
	for i, header := range headers {
		fieldMap[strings.TrimSpace(header)] = i
	}

	// Analyze struct type and build field info
	var zero T
	structType := reflect.TypeOf(zero)
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		if closer != nil {
			closer.Close()
		}
		return nil, fmt.Errorf("type parameter must be a struct, got %s", structType.Kind())
	}

	fieldInfo, err := buildFieldInfo(structType, fieldMap)
	if err != nil {
		if closer != nil {
			closer.Close()
		}
		return nil, err
	}

	return &CSVIterator[T]{
		reader:     csvReader,
		closer:     closer,
		headers:    headers,
		fieldMap:   fieldMap,
		structType: structType,
		fieldInfo:  fieldInfo,
	}, nil
}

func buildFieldInfo(structType reflect.Type, fieldMap map[string]int) ([]fieldInfo, error) {
	var fields []fieldInfo

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		csvTag := field.Tag.Get("csv")
		if csvTag == "" {
			return nil, fmt.Errorf("field %s missing required 'csv' annotation", field.Name)
		}

		// Parse csv tag (format: "column_name" or "column_name,required")
		parts := strings.Split(csvTag, ",")
		csvColumn := strings.TrimSpace(parts[0])
		required := false

		for _, part := range parts[1:] {
			if strings.TrimSpace(part) == "required" {
				required = true
			}
		}

		// Check if column exists in CSV
		_, exists := fieldMap[csvColumn]
		if !exists {
			if required {
				return nil, fmt.Errorf("required CSV column '%s' not found for field %s", csvColumn, field.Name)
			}
			continue // Skip optional missing columns
		}

		fields = append(fields, fieldInfo{
			fieldIndex: i,
			csvColumn:  csvColumn,
			fieldType:  field.Type,
			required:   required,
		})
	}

	return fields, nil
}

// Next reads and parses the next CSV row into the struct type
func (it *CSVIterator[T]) Next() (*T, error) {
	record, err := it.reader.Read()
	if err != nil {
		return nil, err // This includes io.EOF
	}

	// Create new instance
	var result T
	resultValue := reflect.ValueOf(&result).Elem()

	// Handle pointer types
	if it.structType.Kind() == reflect.Ptr {
		newStruct := reflect.New(it.structType.Elem())
		resultValue.Set(newStruct)
		resultValue = newStruct.Elem()
	}

	// Parse each field
	for _, field := range it.fieldInfo {
		columnIndex := it.fieldMap[field.csvColumn]
		
		// Check if we have enough columns
		if columnIndex >= len(record) {
			if field.required {
				return nil, fmt.Errorf("missing required column '%s' in CSV row", field.csvColumn)
			}
			continue
		}

		value := strings.TrimSpace(record[columnIndex])
		
		// Skip empty values for non-required fields
		if value == "" && !field.required {
			continue
		}

		if err := setFieldValue(resultValue.Field(field.fieldIndex), value, field.fieldType); err != nil {
			return nil, fmt.Errorf("failed to parse field %s (column %s): %w", 
				it.structType.Field(field.fieldIndex).Name, field.csvColumn, err)
		}
	}

	return &result, nil
}

func setFieldValue(fieldValue reflect.Value, strValue string, fieldType reflect.Type) error {
	if !fieldValue.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	switch fieldType.Kind() {
	case reflect.String:
		fieldValue.SetString(strValue)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if strValue == "" {
			return nil // Leave zero value
		}
		intVal, err := strconv.ParseInt(strValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer: %s", strValue)
		}
		fieldValue.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if strValue == "" {
			return nil // Leave zero value
		}
		uintVal, err := strconv.ParseUint(strValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer: %s", strValue)
		}
		fieldValue.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		if strValue == "" {
			return nil // Leave zero value
		}
		floatVal, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return fmt.Errorf("invalid float: %s", strValue)
		}
		fieldValue.SetFloat(floatVal)

	case reflect.Bool:
		if strValue == "" {
			return nil // Leave zero value
		}
		boolVal, err := strconv.ParseBool(strValue)
		if err != nil {
			return fmt.Errorf("invalid boolean: %s", strValue)
		}
		fieldValue.SetBool(boolVal)

	case reflect.Struct:
		// Handle time.Time specifically
		if fieldType == reflect.TypeOf(time.Time{}) {
			if strValue == "" {
				return nil // Leave zero value
			}
			
			// Try common time formats in order of preference
			formats := []string{
				time.RFC3339,     // "2006-01-02T15:04:05Z07:00"
				time.RFC3339Nano, // "2006-01-02T15:04:05.999999999Z07:00"
				"2006-01-02 15:04:05", // Common SQL datetime format
				"2006-01-02",     // Date only
				"15:04:05",       // Time only
				"01/02/2006",     // US date format
				"01/02/2006 15:04:05", // US datetime format
				"02/01/2006",     // European date format
				"02/01/2006 15:04:05", // European datetime format
			}
			
			var timeVal time.Time
			var err error
			for _, format := range formats {
				// Use UTC for formats without explicit timezone to ensure cross-platform consistency
				if format == time.RFC3339 || format == time.RFC3339Nano {
					timeVal, err = time.Parse(format, strValue)
				} else {
					timeVal, err = time.ParseInLocation(format, strValue, time.UTC)
				}
				if err == nil {
					fieldValue.Set(reflect.ValueOf(timeVal))
					return nil
				}
			}
			return fmt.Errorf("invalid time format: %s (supported formats: RFC3339, YYYY-MM-DD, YYYY-MM-DD HH:MM:SS, etc.)", strValue)
		}
		return fmt.Errorf("unsupported struct type: %s", fieldType)

	case reflect.Ptr:
		if strValue == "" {
			return nil // Leave nil
		}
		// Create new instance of the pointed-to type
		newVal := reflect.New(fieldType.Elem())
		if err := setFieldValue(newVal.Elem(), strValue, fieldType.Elem()); err != nil {
			return err
		}
		fieldValue.Set(newVal)

	default:
		return fmt.Errorf("unsupported field type: %s", fieldType.Kind())
	}

	return nil
}

// ForEach iterates through all CSV rows and calls the provided function
func (it *CSVIterator[T]) ForEach(fn func(*T, error) bool) {
	for {
		item, err := it.Next()
		if err == io.EOF {
			break
		}
		if !fn(item, err) {
			break
		}
	}
}

// ToSlice reads all remaining CSV rows into a slice
func (it *CSVIterator[T]) ToSlice() ([]*T, error) {
	var results []*T
	
	for {
		item, err := it.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	
	return results, nil
}

// Headers returns the CSV column headers
func (it *CSVIterator[T]) Headers() []string {
	return it.headers
}

// Close closes the underlying reader if it implements io.Closer
func (it *CSVIterator[T]) Close() error {
	if it.closer != nil {
		return it.closer.Close()
	}
	return nil
}