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
//   - Support for various Go types: string, int, uint, float, bool, time.Time, and pointers
//
// # Basic Usage
//
// Define a struct with csv tags:
//
//	type Person struct {
//	    Name      string     `csv:"name,required"`
//	    Age       int        `csv:"age"`
//	    Email     string     `csv:"email,required"`
//	    Salary    *float64   `csv:"salary"`        // Optional field (pointer)
//	    Active    bool       `csv:"active"`
//	    BirthDate time.Time  `csv:"birth_date"`    // Date field
//	    CreatedAt *time.Time `csv:"created_at"`    // Optional timestamp
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
//   - string: Direct string values, no conversion needed
//   - Integers: int, int8, int16, int32, int64 (decimal format)
//   - Unsigned integers: uint, uint8, uint16, uint32, uint64 (decimal format)
//   - Floating point: float32, float64 (decimal format, scientific notation supported)
//   - Boolean: bool (accepts: true/false, 1/0, yes/no, on/off, case insensitive)
//   - Time: time.Time (multiple format auto-detection, see Time Parsing section)
//   - Pointers: *T where T is any supported type above (for optional/nullable fields)
//
// Empty CSV values are handled as follows:
//   - Regular fields: Set to their zero value (0, "", false, time.Time{})
//   - Pointer fields: Set to nil
//   - Required fields: Generate an error if empty
//
// # Time Parsing
//
// The time.Time type supports automatic format detection for common date/time formats:
//
//   - RFC3339: "2006-01-02T15:04:05Z07:00"
//   - RFC3339Nano: "2006-01-02T15:04:05.999999999Z07:00" 
//   - SQL DateTime: "2006-01-02 15:04:05"
//   - Date only: "2006-01-02"
//   - Time only: "15:04:05"
//   - US date: "01/02/2006"
//   - US datetime: "01/02/2006 15:04:05"
//   - European date: "02/01/2006"
//   - European datetime: "02/01/2006 15:04:05"
//
// For cross-platform consistency, dates without explicit timezone information
// are parsed as UTC. Only RFC3339 formats with explicit timezone are preserved.
//
// Example with time fields:
//
//	type Event struct {
//	    Name      string    `csv:"event_name,required"`
//	    StartDate time.Time `csv:"start_date"`
//	    EndTime   *time.Time `csv:"end_time"` // Optional
//	}
//
//	// CSV data with various time formats:
//	// event_name,start_date,end_time
//	// "Conference",2024-03-15,2024-03-15T18:00:00Z
//	// "Workshop","03/20/2024 09:00:00",
//	// "Meeting",2024-03-22 14:30:00,2024-03-22T16:00:00Z
//	// "Webinar",01/15/2024,01/15/2024 16:30:00
//
// Time Parsing Notes:
//   - Formats are tried in order until one succeeds
//   - All formats without explicit timezone are parsed as UTC for cross-platform consistency
//   - RFC3339 formats with timezone information preserve the original timezone
//   - Empty time fields result in time.Time{} (zero value) or nil for pointer fields
//   - Invalid time formats return descriptive error messages with supported formats
//
// # Time Format Troubleshooting
//
// If you encounter time parsing errors, check the following:
//
//   - Ensure your date format matches one of the supported patterns
//   - For custom formats, consider preprocessing your CSV data
//   - Use RFC3339 format ("2006-01-02T15:04:05Z07:00") for maximum compatibility
//   - Check for extra whitespace around date values in your CSV
//   - Verify that date separators match expected patterns (/ vs - vs space)
//
// Common time format examples:
//   - "2024-03-15" → March 15, 2024 (date only)
//   - "2024-03-15 14:30:00" → March 15, 2024 at 2:30 PM
//   - "2024-03-15T14:30:00Z" → March 15, 2024 at 2:30 PM UTC
//   - "03/15/2024" → March 15, 2024 (US format)
//   - "15/03/2024" → March 15, 2024 (European format)
//
// # Error Handling
//
// The package provides detailed error messages for:
//
//   - Missing required CSV columns
//   - Missing struct annotations
//   - Type conversion errors (including time parsing failures)
//   - Invalid CSV format
//   - Network errors (for URL sources)
//   - Time format mismatches with helpful format suggestions
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
// # Advanced Example with Time Fields
//
// Here's a comprehensive example showing time.Time usage with different formats:
//
//	type LogEntry struct {
//	    ID        int        `csv:"id,required"`
//	    Message   string     `csv:"message,required"`
//	    Timestamp time.Time  `csv:"timestamp"`        // Required time field
//	    ExpiresAt *time.Time `csv:"expires_at"`       // Optional time field
//	    Level     string     `csv:"level"`
//	}
//
//	// Sample CSV content:
//	// id,message,timestamp,expires_at,level
//	// 1,"System started","2024-03-15T10:30:00Z","2024-12-31T23:59:59Z","INFO"
//	// 2,"User login","2024-03-15 10:31:25",,DEBUG
//	// 3,"Error occurred","03/15/2024 10:32:00","12/31/2024","ERROR"
//
//	iterator, err := supercsv.NewFromFile[LogEntry]("logs.csv")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer iterator.Close()
//
//	for {
//	    entry, err := iterator.Next()
//	    if err == io.EOF {
//	        break
//	    }
//	    if err != nil {
//	        log.Printf("Parse error: %v", err)
//	        continue
//	    }
//	    
//	    fmt.Printf("Log %d: %s at %s\n", 
//	        entry.ID, entry.Message, entry.Timestamp.Format("2006-01-02 15:04:05"))
//	    
//	    if entry.ExpiresAt != nil {
//	        fmt.Printf("  Expires: %s\n", entry.ExpiresAt.Format("2006-01-02"))
//	    }
//	}
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