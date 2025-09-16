package supercsv

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

// Example structs with CSV annotations
type Person struct {
	Name   string  `csv:"name,required"`
	Age    int     `csv:"age"`
	Email  string  `csv:"email,required"`
	Salary *float64 `csv:"salary"`
	Active bool    `csv:"active"`
}

type Product struct {
	ID          int     `csv:"id,required"`
	Name        string  `csv:"product_name,required"`
	Price       float64 `csv:"price"`
	InStock     bool    `csv:"in_stock"`
	Description string  `csv:"description"`
}

func TestCSVIterator_Basic(t *testing.T) {
	csvData := `name,age,email,salary,active
John Doe,30,john@example.com,50000.50,true
Jane Smith,25,jane@example.com,,false
Bob Johnson,35,bob@example.com,75000.00,true`

	reader := strings.NewReader(csvData)
	iterator, err := NewFromReader[Person](reader)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}
	defer iterator.Close()

	var people []*Person
	for {
		person, err := iterator.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read person: %v", err)
		}
		people = append(people, person)
	}

	if len(people) != 3 {
		t.Fatalf("Expected 3 people, got %d", len(people))
	}

	// Verify first person
	if people[0].Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", people[0].Name)
	}
	if people[0].Age != 30 {
		t.Errorf("Expected age 30, got %d", people[0].Age)
	}
	if people[0].Salary == nil || *people[0].Salary != 50000.50 {
		t.Errorf("Expected salary 50000.50, got %v", people[0].Salary)
	}
	if !people[0].Active {
		t.Errorf("Expected active true, got %t", people[0].Active)
	}

	// Verify second person (with nil salary)
	if people[1].Salary != nil {
		t.Errorf("Expected nil salary, got %v", people[1].Salary)
	}
}

func TestCSVIterator_ToSlice(t *testing.T) {
	csvData := `id,product_name,price,in_stock,description
1,Laptop,999.99,true,High-performance laptop
2,Mouse,29.99,true,Wireless mouse
3,Keyboard,79.99,false,Mechanical keyboard`

	reader := strings.NewReader(csvData)
	iterator, err := NewFromReader[Product](reader)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}
	defer iterator.Close()

	products, err := iterator.ToSlice()
	if err != nil {
		t.Fatalf("Failed to read products: %v", err)
	}

	if len(products) != 3 {
		t.Fatalf("Expected 3 products, got %d", len(products))
	}

	// Verify first product
	if products[0].ID != 1 {
		t.Errorf("Expected ID 1, got %d", products[0].ID)
	}
	if products[0].Name != "Laptop" {
		t.Errorf("Expected name 'Laptop', got '%s'", products[0].Name)
	}
	if products[0].Price != 999.99 {
		t.Errorf("Expected price 999.99, got %f", products[0].Price)
	}
}

func TestCSVIterator_ForEach(t *testing.T) {
	csvData := `name,age,email,salary,active
Alice,28,alice@example.com,60000,true
Charlie,32,charlie@example.com,80000,true`

	reader := strings.NewReader(csvData)
	iterator, err := NewFromReader[Person](reader)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}
	defer iterator.Close()

	count := 0
	iterator.ForEach(func(person *Person, err error) bool {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return false
		}
		count++
		return true
	})

	if count != 2 {
		t.Errorf("Expected 2 iterations, got %d", count)
	}
}

func TestCSVIterator_MissingRequiredColumn(t *testing.T) {
	csvData := `name,age,salary,active
John Doe,30,50000.50,true`

	reader := strings.NewReader(csvData)
	_, err := NewFromReader[Person](reader)
	if err == nil {
		t.Fatal("Expected error for missing required column 'email'")
	}
	if !strings.Contains(err.Error(), "required CSV column 'email' not found") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestCSVIterator_MissingAnnotation(t *testing.T) {
	type BadStruct struct {
		Name string // Missing csv annotation
		Age  int    `csv:"age"`
	}

	csvData := `name,age
John,30`

	reader := strings.NewReader(csvData)
	_, err := NewFromReader[BadStruct](reader)
	if err == nil {
		t.Fatal("Expected error for missing csv annotation")
	}
	if !strings.Contains(err.Error(), "missing required 'csv' annotation") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// Example usage functions
func ExampleNewFromFile() {
	// Create iterator from file
	iterator, err := NewFromFile[Person]("/path/to/people.csv")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer iterator.Close()

	// Read all data
	people, err := iterator.ToSlice()
	if err != nil {
		fmt.Printf("Error reading data: %v\n", err)
		return
	}

	for _, person := range people {
		fmt.Printf("Name: %s, Age: %d, Email: %s\n", 
			person.Name, person.Age, person.Email)
	}
}

func ExampleNewFromURL() {
	// Create iterator from URL
	iterator, err := NewFromURL[Product]("https://example.com/products.csv")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer iterator.Close()

	// Iterate one by one
	for {
		product, err := iterator.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading product: %v\n", err)
			break
		}
		
		fmt.Printf("Product: %s, Price: $%.2f\n", 
			product.Name, product.Price)
	}
}

func ExampleCSVIterator_ForEach() {
	csvData := `name,age,email,salary,active
John Doe,30,john@example.com,50000,true
Jane Smith,25,jane@example.com,60000,false`

	reader := strings.NewReader(csvData)
	iterator, err := NewFromReader[Person](reader)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer iterator.Close()

	// Use ForEach for processing
	iterator.ForEach(func(person *Person, err error) bool {
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return false
		}
		
		fmt.Printf("Processing: %s (Age: %d)\n", person.Name, person.Age)
		return true // Continue iteration
	})
}