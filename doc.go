// Package supercsv provides a flexible, type-safe CSV parser for Go using struct annotations.
//
// # Overview
//
// SuperCSV allows you to parse CSV data into any struct type using Go generics and struct tags.
// It supports reading from files, URLs, or any io.Reader with memory-efficient iteration.
//
// # Key Features
//
//   - Generic struct parsing using Go generics
//   - CSV column mapping via struct tags
//   - Support for required and optional fields
//   - Multiple data sources (files, URLs, readers)
//   - Memory-efficient row-by-row processing
//   - Type-safe parsing with comprehensive error handling
//   - Support for various Go types: string, int, uint, float, bool, and pointers
//
// # Basic Usage
//
// Define a struct with csv tags:
//
//	type Person struct {
//	    Name   string   `csv:"name,required"`
//	    Age    int      `csv:"age"`
//	    Email  string   `csv:"email,required"`
//	    Salary *float64 `csv:"salary"`        // Optional field (pointer)
//	    Active bool     `csv:"active"`
//	}
//
// Create an iterator and process data:
//
//	// From file
//	iterator, err := supercsv.NewFromFile[Person]("data.csv")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer iterator.Close()
//
//	// Process row by row
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
//
// # CSV Tag Format
//
// The csv tag supports the following format:
//
//	`csv:"column_name"`          // Maps to CSV column, optional field
//	`csv:"column_name,required"` // Maps to CSV column, required field
//
// All struct fields that should be parsed MUST have a csv tag. Fields without
// csv tags are ignored and will cause an error.
//
// # Supported Types
//
//   - string
//   - int, int8, int16, int32, int64
//   - uint, uint8, uint16, uint32, uint64
//   - float32, float64
//   - bool
//   - Pointers to any of the above (for optional fields)
//
// # Error Handling
//
// The package provides detailed error messages for:
//
//   - Missing required CSV columns
//   - Missing struct annotations
//   - Type conversion errors
//   - Invalid CSV format
//   - Network errors (for URL sources)
//
// # Data Sources
//
// SuperCSV supports multiple data sources:
//
//	// From local file
//	iterator, err := supercsv.NewFromFile[Person]("data.csv")
//
//	// From URL
//	iterator, err := supercsv.NewFromURL[Person]("https://example.com/data.csv")
//
//	// From any io.Reader
//	iterator, err := supercsv.NewFromReader[Person](reader)
//
// # Batch Operations
//
// For convenience, SuperCSV provides batch processing methods:
//
//	// Read all rows into a slice
//	people, err := iterator.ToSlice()
//
//	// Process with callback
//	iterator.ForEach(func(person *Person, err error) bool {
//	    if err != nil {
//	        log.Printf("Error: %v", err)
//	        return false // Stop processing
//	    }
//	    processPerson(person)
//	    return true // Continue
//	})
//
// # Memory Efficiency
//
// The iterator processes CSV data row-by-row, making it suitable for large files
// without loading everything into memory at once. Only the current row and struct
// metadata are kept in memory during iteration.
//
// # Thread Safety
//
// CSVIterator instances are not thread-safe. Each goroutine should use its own
// iterator instance.
package supercsv