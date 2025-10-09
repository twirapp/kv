package kvvaluer

import (
	"strconv"

	kv "github.com/twirapp/kv"
	kvscanner "github.com/twirapp/kv/scanner"
)

var _ kv.Valuer = (*Valuer)(nil)

type Valuer struct {
	Value []byte
	Error error
}

func (v *Valuer) Err() error {
	return v.Error
}

func (v *Valuer) Int() (int64, error) {
	if v.Error != nil {
		return 0, v.Error
	}

	if len(v.Value) == 0 {
		return 0, kv.ErrValuerEmptySlice
	}
	n, err := strconv.Atoi(string(v.Value))
	if err != nil {
		return 0, err
	}
	result := int64(n)

	return result, nil
}

func (v *Valuer) String() (string, error) {
	if v.Error != nil {
		return "", v.Error
	}

	return string(v.Value), nil
}

func (v *Valuer) Bytes() ([]byte, error) {
	if v.Error != nil {
		return nil, v.Error
	}

	return v.Value, nil
}

func (v *Valuer) Bool() (bool, error) {
	if v.Error != nil {
		return false, v.Error
	}

	if len(v.Value) == 0 {
		return false, kv.ErrValuerEmptySlice
	}

	return strconv.ParseBool(string(v.Value))
}

func (v *Valuer) Float() (float64, error) {
	if v.Error != nil {
		return 0, v.Error
	}

	if len(v.Value) == 0 {
		return 0, kv.ErrValuerEmptySlice
	}

	return strconv.ParseFloat(string(v.Value), 64)
}

func (v *Valuer) Scan(dest any) error {
	if v.Error != nil {
		return v.Error
	}

	return kvscanner.Scan(v.Value, dest)
}
