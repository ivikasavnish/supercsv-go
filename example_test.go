package supercsv

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
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

type Event struct {
	Name      string     `csv:"event_name,required"`
	StartDate time.Time  `csv:"start_date"`
	EndTime   *time.Time `csv:"end_time"`
	Duration  int        `csv:"duration_minutes"`
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

func TestCSVIterator_CustomDelimiter(t *testing.T) {
	csvData := `name;age;email
John Doe;30;john@example.com
Jane Smith;25;jane@example.com`

	reader := strings.NewReader(csvData)
	iterator, err := NewFromReaderWithDelimiter[Person](reader, ';')
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

	if len(people) != 2 {
		t.Fatalf("Expected 2 people, got %d", len(people))
	}

	// Verify first person
	if people[0].Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", people[0].Name)
	}
	if people[0].Age != 30 {
		t.Errorf("Expected age 30, got %d", people[0].Age)
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

func TestCSVIterator_TimeSupport(t *testing.T) {
	csvData := `event_name,start_date,end_time,duration_minutes
Conference,2024-03-15,2024-03-15T18:00:00Z,480
Workshop,03/20/2024 09:00:00,,240
Meeting,2024-03-22 14:30:00,2024-03-22T16:00:00Z,90
Webinar,2024-04-01T10:00:00Z,2024-04-01T11:30:00Z,90`

	reader := strings.NewReader(csvData)
	iterator, err := NewFromReader[Event](reader)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}
	defer iterator.Close()

	events, err := iterator.ToSlice()
	if err != nil {
		t.Fatalf("Failed to read events: %v", err)
	}

	if len(events) != 4 {
		t.Fatalf("Expected 4 events, got %d", len(events))
	}

	// Verify first event (date only format)
	expectedDate1 := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	if !events[0].StartDate.Equal(expectedDate1) {
		t.Errorf("Expected start date %v, got %v", expectedDate1, events[0].StartDate)
	}
	if events[0].EndTime == nil {
		t.Error("Expected end time to be parsed, got nil")
	} else {
		expectedEndTime1 := time.Date(2024, 3, 15, 18, 0, 0, 0, time.UTC)
		if !events[0].EndTime.Equal(expectedEndTime1) {
			t.Errorf("Expected end time %v, got %v", expectedEndTime1, *events[0].EndTime)
		}
	}

	// Verify second event (US datetime format, nil end time)
	expectedDate2 := time.Date(2024, 3, 20, 9, 0, 0, 0, time.UTC)
	if !events[1].StartDate.Equal(expectedDate2) {
		t.Errorf("Expected start date %v, got %v", expectedDate2, events[1].StartDate)
	}
	if events[1].EndTime != nil {
		t.Errorf("Expected nil end time, got %v", events[1].EndTime)
	}

	// Verify third event (SQL datetime format)
	expectedDate3 := time.Date(2024, 3, 22, 14, 30, 0, 0, time.UTC)
	if !events[2].StartDate.Equal(expectedDate3) {
		t.Errorf("Expected start date %v, got %v", expectedDate3, events[2].StartDate)
	}
	if events[2].Duration != 90 {
		t.Errorf("Expected duration 90, got %d", events[2].Duration)
	}

	// Verify fourth event (RFC3339 format)
	expectedDate4 := time.Date(2024, 4, 1, 10, 0, 0, 0, time.UTC)
	if !events[3].StartDate.Equal(expectedDate4) {
		t.Errorf("Expected start date %v, got %v", expectedDate4, events[3].StartDate)
	}
}

func ExampleNewFromReader_timeSupport() {
	csvData := `event_name,start_date,end_time,duration_minutes
Conference,2024-03-15,2024-03-15T18:00:00Z,480
Workshop,03/20/2024 09:00:00,,240`

	reader := strings.NewReader(csvData)
	iterator, err := NewFromReader[Event](reader)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer iterator.Close()

	for {
		event, err := iterator.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading event: %v\n", err)
			break
		}
		
		fmt.Printf("Event: %s, Start: %s, Duration: %d mins\n", 
			event.Name, event.StartDate.Format("2006-01-02 15:04"), event.Duration)
		
		if event.EndTime != nil {
			fmt.Printf("  End: %s\n", event.EndTime.Format("2006-01-02 15:04"))
		}
	}
	// Output:
	// Event: Conference, Start: 2024-03-15 00:00, Duration: 480 mins
	//   End: 2024-03-15 18:00
	// Event: Workshop, Start: 2024-03-20 09:00, Duration: 240 mins
}