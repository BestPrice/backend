package bp

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"

	uuid "github.com/nu7hatch/gouuid"
)

type ID struct {
	*uuid.UUID
}

func (x *ID) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(b, "\"")
	uuid, err := uuid.Parse(b)
	if err != nil {
		return err
	}
	x.UUID = uuid
	return nil
}

func (x *ID) MarshalJSON() ([]byte, error) {
	if x.UUID == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(x.String())
}

func (x *ID) Scan(v interface{}) error {
	switch v.(type) {
	case nil:
		return nil
	case []byte:
		uuid, err := uuid.ParseHex(string(v.([]byte)))
		if err != nil {
			return err
		}
		x.UUID = uuid
		return nil
	default:
		return errors.New("unsuported type")
	}
}

func (x *ID) Null() bool {
	return x.UUID == nil
}

type JsonNullInt64 struct {
	sql.NullInt64
}

func (v JsonNullInt64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int64)
	} else {
		return json.Marshal(nil)
	}
}

func (v *JsonNullInt64) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *int64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Int64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

type JsonNullString struct {
	sql.NullString
}

func (v JsonNullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	} else {
		return json.Marshal(nil)
	}
}

func (v *JsonNullString) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *string
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.String = *x
	} else {
		v.Valid = false
	}
	return nil
}

type JsonNullBool struct {
	sql.NullBool
}

func (v JsonNullBool) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Bool)
	} else {
		return json.Marshal(nil)
	}
}

func (v *JsonNullBool) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *bool
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Bool = *x
	} else {
		v.Valid = false
	}
	return nil
}

type JsonNullFloat64 struct {
	sql.NullFloat64
}

func (v JsonNullFloat64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Float64)
	} else {
		return json.Marshal(nil)
	}
}

func (v *JsonNullFloat64) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *float64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Float64 = *x
	} else {
		v.Valid = false
	}
	return nil
}
