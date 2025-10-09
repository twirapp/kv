package kvscanner_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	kvscanner "github.com/twirapp/kv/scanner"
)

// Test structs
type SimpleStructWithTags struct {
	Name string `kv:"name"`
	Age  int    `kv:"age"`
}

type SimpleStructWithExactlyDifferentKvTag struct {
	Name string `kv:"username"`
}

type SimpleStructWithoutTags struct {
	Name     string
	Age      int
	UserName string
}

type ComplexStruct struct {
	Name     string   `kv:"name"`
	Age      int      `kv:"age"`
	Balance  float64  `kv:"balance"`
	IsActive bool     `kv:"is_active"`
	Tags     []string `kv:"tags"`
	Address  Address  `kv:"address"`
}

type Address struct {
	Street string `kv:"street"`
	City   string `kv:"city"`
}

type SnakeCaseStruct struct {
	UserName  string `kv:"user_name"`
	UserEmail string `kv:"user_email"`
}

type PascalCaseStruct struct {
	UserName  string `kv:"UserName"`
	UserEmail string `kv:"UserEmail"`
}

type DuplicateFieldStruct struct {
	Name string `kv:"name"`
	NaMe string `kv:"name"`
}

type StructWithAndWithoutTags struct {
	Name     string `kv:"name"`
	Age      int
	UserName string `kv:"user_name"`
	LastName string
}

// Test cases
func TestScan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []byte
		dest     interface{}
		expected interface{}
		wantErr  bool
		errMsg   string
	}{
		{
			name:  "Simple struct with kv tags",
			input: []byte(`{"name":"John","age":30}`),
			dest:  &SimpleStructWithTags{},
			expected: &SimpleStructWithTags{
				Name: "John",
				Age:  30,
			},
			wantErr: false,
		},
		{
			name:  "Simple struct without kv tags",
			input: []byte(`{"name":"Jane","age":28,"userName":"jane_doe"}`),
			dest:  &SimpleStructWithoutTags{},
			expected: &SimpleStructWithoutTags{
				Name:     "Jane",
				Age:      28,
				UserName: "jane_doe",
			},
			wantErr: false,
		},
		{
			name:  "Struct with and without tags",
			input: []byte(`{"name":"Mike","age":35,"user_name":"mike123","lastName":"Smith"}`),
			dest:  &StructWithAndWithoutTags{},
			expected: &StructWithAndWithoutTags{
				Name:     "Mike",
				Age:      35,
				UserName: "mike123",
				LastName: "Smith",
			},
			wantErr: false,
		},
		{
			name:  "Struct with exactly different kv tag",
			input: []byte(`{"username":"alice"}`),
			dest:  &SimpleStructWithExactlyDifferentKvTag{},
			expected: &SimpleStructWithExactlyDifferentKvTag{
				Name: "alice",
			},
			wantErr: false,
		},
		{
			name:  "Struct with exactly different kv tag 2",
			input: []byte(`{"name":"alice"}`),
			dest:  &SimpleStructWithExactlyDifferentKvTag{},
			expected: &SimpleStructWithExactlyDifferentKvTag{
				Name: "",
			},
			wantErr: false,
		},
		{
			name:  "Struct with and without tags with different cases",
			input: []byte(`{"name":"Mike","age":35,"user_name":"mike123","last_name":"Smith"}`),
			dest:  &StructWithAndWithoutTags{},
			expected: &StructWithAndWithoutTags{
				Name:     "Mike",
				Age:      35,
				UserName: "mike123",
				LastName: "Smith",
			},
			wantErr: false,
		},
		{
			name:  "Complex struct with various types",
			input: []byte(`{"name":"Alice","age":25,"balance":1234.56,"is_active":true,"tags":["student","active"],"address":{"street":"123 Main St","city":"New York"}}`),
			dest:  &ComplexStruct{},
			expected: &ComplexStruct{
				Name:     "Alice",
				Age:      25,
				Balance:  1234.56,
				IsActive: true,
				Tags:     []string{"student", "active"},
				Address: Address{
					Street: "123 Main St",
					City:   "New York",
				},
			},
			wantErr: false,
		},
		{
			name:  "Snake case  keys",
			input: []byte(`{"user_name":"Bob","user_email":"bob@example.com"}`),
			dest:  &SnakeCaseStruct{},
			expected: &SnakeCaseStruct{
				UserName:  "Bob",
				UserEmail: "bob@example.com",
			},
			wantErr: false,
		},
		{
			name:  "Pascal case  keys",
			input: []byte(`{"UserName":"Charlie","UserEmail":"charlie@example.com"}`),
			dest:  &PascalCaseStruct{},
			expected: &PascalCaseStruct{
				UserName:  "Charlie",
				UserEmail: "charlie@example.com",
			},
			wantErr: false,
		},
		{
			name:    "Invalid ",
			input:   []byte(`{"name":"John",age:30}`),
			dest:    &SimpleStructWithTags{},
			wantErr: true,
			errMsg:  "failed to unmarshal ",
		},
		{
			name:    "Type mismatch (string to int)",
			input:   []byte(`{"name":"John","age":"thirty"}`),
			dest:    &SimpleStructWithTags{},
			wantErr: true,
			errMsg:  "expected number, got string",
		},
		{
			name:    "Nil destination",
			input:   []byte(`{"name":"John","age":30}`),
			dest:    nil,
			wantErr: true,
			errMsg:  "destination must be a non-nil pointer",
		},
		{
			name:    "Non-struct destination",
			input:   []byte(`{"name":"John","age":30}`),
			dest:    new(string),
			wantErr: true,
			errMsg:  "destination must be a pointer to a struct",
		},
		{
			name:    "Duplicate field names",
			input:   []byte(`{"name":"John"}`),
			dest:    &DuplicateFieldStruct{},
			wantErr: true,
			errMsg:  "duplicate field name detected",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Handle expected panic for duplicate field names
				var err error
				defer func() {
					if r := recover(); r != nil {
						if tt.wantErr && strings.Contains(fmt.Sprint(r), tt.errMsg) {
							return
						}
						t.Errorf("unexpected panic: %v", r)
					}
				}()

				// Run Scan
				err = kvscanner.Scan(tt.input, tt.dest)
				if (err != nil) != tt.wantErr {
					t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Scan() error = %v, wantErrMsg %v", err, tt.errMsg)
					return
				}

				// Check field values if no error expected
				if !tt.wantErr {
					if !reflect.DeepEqual(tt.dest, tt.expected) {
						t.Errorf("Scan() got = %+v, want = %+v", tt.dest, tt.expected)
					}
				}
			},
		)
	}
}

// Test partial  input
func TestScan_PartialInput(t *testing.T) {
	t.Parallel()

	input := []byte(`{"name":"John"}`)
	dest := &SimpleStructWithTags{}
	expected := &SimpleStructWithTags{Name: "John", Age: 0}

	err := kvscanner.Scan(input, dest)
	if err != nil {
		t.Errorf("Scan() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(dest, expected) {
		t.Errorf("Scan() got = %+v, want = %+v", dest, expected)
	}
}

// Test naming convention fallback
func TestScan_NamingFallback(t *testing.T) {
	t.Parallel()

	input := []byte(`{"userName":"John","userEmail":"john@example.com"}`)
	dest := &SnakeCaseStruct{}
	expected := &SnakeCaseStruct{UserName: "", UserEmail: ""}

	err := kvscanner.Scan(input, dest)
	if err != nil {
		t.Errorf("Scan() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(dest, expected) {
		t.Errorf("Scan() got = %+v, want = %+v", dest, expected)
	}
}
