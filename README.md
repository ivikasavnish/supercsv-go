# SuperCSV - Generic CSV Iterator for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/ivikasavnish/supercsv-go.svg)](https://pkg.go.dev/github.com/ivikasavnish/supercsv-go)
[![Test](https://github.com/ivikasavnish/supercsv-go/workflows/Test/badge.svg)](https://github.com/ivikasavnish/supercsv-go/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/ivikasavnish/supercsv-go)](https://goreportcard.com/report/github.com/ivikasavnish/supercsv-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A flexible, type-safe CSV parser for Go that uses struct annotations to map CSV columns to struct fields.

## Features

🎯 **Generic Struct Parsing** - Works with any struct type using Go generics
📝 **CSV Annotations** - Simple `csv:"column_name"` tag system
🔧 **Multiple Sources** - Read from files, URLs, or any `io.Reader`
✅ **Required Fields** - Mark fields as required with `csv:"column,required"`
🚀 **Iterator Pattern** - Memory efficient row-by-row processing
📦 **Batch Operations** - Convert to slice or use ForEach for bulk processing
🛡️ **Type Safety** - Full compile-time type checking
🔧 **Custom Delimiters** - Support for comma, semicolon, tab, and any custom delimiter

## Quick Start

### 1. Define your struct with CSV annotations

```go
type Person struct {
    Name   string   `csv:"name,required"`
    Age    int      `csv:"age"`
    Email  string   `csv:"email,required"`
    Salary *float64 `csv:"salary"`        // Pointer for optional fields
    Active bool     `csv:"active"`
}
```

### 2. Create an iterator

```go
import "github.com/ivikasavnish/supercsv-go"

// From file (default comma delimiter)
iterator, err := supercsv.NewFromFile[Person]("data.csv")

// From file with custom delimiter (semicolon)
iterator, err := supercsv.NewFromFileWithDelimiter[Person]("data.csv", ';')

// From URL (default comma delimiter)
iterator, err := supercsv.NewFromURL[Person]("https://example.com/data.csv")

// From URL with custom delimiter (semicolon)
iterator, err := supercsv.NewFromURLWithDelimiter[Person]("https://example.com/data.csv", ';')

// From reader (default comma delimiter)
iterator, err := supercsv.NewFromReader[Person](reader)

// From reader with custom delimiter (semicolon) 
iterator, err := supercsv.NewFromReaderWithDelimiter[Person](reader, ';')

defer iterator.Close()
```

### 3. Process the data

```go
// One by one
for {
    person, err := iterator.Next()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Name: %s, Age: %d\n", person.Name, person.Age)
}

// All at once
people, err := iterator.ToSlice()

// With callback
iterator.ForEach(func(person *Person, err error) bool {
    if err != nil {
        log.Printf("Error: %v", err)
        return false
    }
    
    processPerson(person)
    return true // Continue
})
```

## CSV Annotation Rules

- **Required**: All struct fields must have `csv:"column_name"` annotation
- **Column Mapping**: `csv:"column_name"` maps to CSV header
- **Required Fields**: `csv:"column_name,required"` - fails if column missing
- **Optional Fields**: Use pointers for optional fields that can be nil

## Supported Types

- `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `bool`
- Pointers to any of the above (for optional fields)

## Example CSV Data

```csv
name,age,email,salary,active
John Doe,30,john@example.com,50000.50,true
Jane Smith,25,jane@example.com,,false
Bob Johnson,35,bob@example.com,75000.00,true
```

## Error Handling

The iterator provides detailed error messages for:
- Missing required CSV columns
- Missing struct annotations
- Type conversion errors
- Invalid CSV format

## Memory Efficiency

The iterator processes CSV data row-by-row, making it suitable for large files without loading everything into memory at once.

## Installation

```bash
go get github.com/ivikasavnish/supercsv-go
```

## Testing

Run the tests:

```bash
go test -v
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.