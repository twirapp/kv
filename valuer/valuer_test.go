package kvvaluer

import (
	"reflect"
	"testing"
)

func TestValuer_Bool(t *testing.T) {
	t.Parallel()

	type fields struct {
		value []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{
			name:    "valid true",
			fields:  fields{value: []byte("true")},
			want:    true,
			wantErr: false,
		},
		{
			name:    "valid false",
			fields:  fields{value: []byte("false")},
			want:    false,
			wantErr: false,
		},
		{
			name:    "invalid value",
			fields:  fields{value: []byte("notabool")},
			want:    false,
			wantErr: true,
		},
		{
			name:    "empty slice",
			fields:  fields{value: []byte("")},
			want:    false,
			wantErr: true,
		},
		{
			name: "valid 1",
			fields: fields{
				value: []byte("1"),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "valid 0",
			fields: fields{
				value: []byte("0"),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "valid 2",
			fields: fields{
				value: []byte("2"),
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				v := &Valuer{
					Value: tt.fields.value,
				}
				got, err := v.Bool()
				if (err != nil) != tt.wantErr {
					t.Errorf("Bool() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("Bool() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestValuer_Bytes(t *testing.T) {
	t.Parallel()

	type fields struct {
		value []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name:    "valid bytes",
			fields:  fields{value: []byte("hello")},
			want:    []byte("hello"),
			wantErr: false,
		},
		{
			name:    "empty bytes",
			fields:  fields{value: []byte("")},
			want:    []byte(""),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				v := &Valuer{
					Value: tt.fields.value,
				}
				got, err := v.Bytes()
				if (err != nil) != tt.wantErr {
					t.Errorf("Bytes() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Bytes() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestValuer_Float(t *testing.T) {
	t.Parallel()

	type fields struct {
		value []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    float64
		wantErr bool
	}{
		{
			name:    "valid float",
			fields:  fields{value: []byte("3.14")},
			want:    3.14,
			wantErr: false,
		},
		{
			name:    "invalid float",
			fields:  fields{value: []byte("notafloat")},
			want:    0,
			wantErr: true,
		},
		{
			name:    "empty slice",
			fields:  fields{value: []byte("")},
			want:    0,
			wantErr: true,
		},
		{
			name:    "valid integer as float",
			fields:  fields{value: []byte("42")},
			want:    42.0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				v := &Valuer{
					Value: tt.fields.value,
				}
				got, err := v.Float()
				if (err != nil) != tt.wantErr {
					t.Errorf("Float() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("Float() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestValuer_Int(t *testing.T) {
	t.Parallel()

	type fields struct {
		value []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		{
			name:    "valid int",
			fields:  fields{value: []byte("42")},
			want:    42,
			wantErr: false,
		},
		{
			name:    "invalid int",
			fields:  fields{value: []byte("notanint")},
			want:    0,
			wantErr: true,
		},
		{
			name:    "empty slice",
			fields:  fields{value: []byte("")},
			want:    0,
			wantErr: true,
		},
		{
			name:    "int out of range",
			fields:  fields{value: []byte("9223372036854775808")}, // math.MaxInt64 + 1
			want:    0,
			wantErr: true,
		},
		{
			name: "float as int",
			fields: fields{
				value: []byte("3.14"),
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				v := &Valuer{
					Value: tt.fields.value,
				}
				got, err := v.Int()
				if (err != nil) != tt.wantErr {
					t.Errorf("Int() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("Int() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestValuer_String(t *testing.T) {
	t.Parallel()

	type fields struct {
		value []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name:    "valid string",
			fields:  fields{value: []byte("hello")},
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "empty string",
			fields:  fields{value: []byte("")},
			want:    "",
			wantErr: false,
		},
		{
			name:    "string with spaces",
			fields:  fields{value: []byte("  hello world  ")},
			want:    "  hello world  ",
			wantErr: false,
		},
		{
			name:    "string with special characters",
			fields:  fields{value: []byte("!@#$%^&*()")},
			want:    "!@#$%^&*()",
			wantErr: false,
		},
		{
			name:    "string with numbers",
			fields:  fields{value: []byte("12345")},
			want:    "12345",
			wantErr: false,
		},
		{
			name:    "string with newline",
			fields:  fields{value: []byte("hello\nworld")},
			want:    "hello\nworld",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				v := &Valuer{
					Value: tt.fields.value,
				}
				got, err := v.String()
				if (err != nil) != tt.wantErr {
					t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("String() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestValuer_Scan(t *testing.T) {
	t.Parallel()

	type SimpleStruct struct {
		Name     string
		Age      int
		UserName string `kv:"user_name"`
	}

	// with test on exact values match
	type fields struct {
		value []byte
	}

	tests := []struct {
		name     string
		fields   fields
		dest     interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "valid struct",
			fields:   fields{value: []byte(`{"Name":"John","Age":30,"user_name":"johndoe"}`)},
			dest:     &SimpleStruct{},
			expected: &SimpleStruct{Name: "John", Age: 30, UserName: "johndoe"},
			wantErr:  false,
		},
		{
			name:     "invalid JSON",
			fields:   fields{value: []byte(`{"Name":"John","Age":30,`)},
			dest:     &SimpleStruct{},
			expected: &SimpleStruct{},
			wantErr:  true,
		},
		{
			name:     "type mismatch",
			fields:   fields{value: []byte(`{"Name":"John","Age":"thirty","user_name":"johndoe"}`)},
			dest:     &SimpleStruct{},
			expected: &SimpleStruct{},
			wantErr:  true,
		},
		{
			name:     "empty JSON",
			fields:   fields{value: []byte(`{}`)},
			dest:     &SimpleStruct{},
			expected: &SimpleStruct{},
			wantErr:  false,
		},
		{
			name:     "extra fields in JSON",
			fields:   fields{value: []byte(`{"Name":"John","Age":30,"user_name":"johndoe","ExtraField":"extra"}`)},
			dest:     &SimpleStruct{},
			expected: &SimpleStruct{Name: "John", Age: 30, UserName: "johndoe"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				v := &Valuer{
					Value: tt.fields.value,
				}
				err := v.Scan(tt.dest)
				if (err != nil) != tt.wantErr {
					t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && !reflect.DeepEqual(tt.dest, tt.expected) {
					t.Errorf("Scan() got = %v, want %v", tt.dest, tt.expected)
				}
			},
		)
	}
}
