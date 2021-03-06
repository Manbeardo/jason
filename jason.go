package jason

import (
	"encoding/json"
	"io"
)

type Jason struct {
	data   interface{}
	exists bool // Used to separate nil and non-existing values
	root   bool // whether it is the root struct
}

// Private array
type jArray struct {
	Slice []*Jason
	Valid bool
}

// Private bool
type jBool struct {
	Bool  bool
	Valid bool
}

type jNull struct {
	Valid bool
}

type jNumber struct {
	Float64 float64
	Valid   bool
}

type jObject struct {
	Map   map[string]*Jason
	Valid bool
}

type jString struct {
	String string
	Valid  bool
}

// Create a new instance from a io.reader.
// Useful for parsing the body of a net/http response.
// Example: NewFromReader(res.Body)
func NewFromReader(reader io.Reader) (*Jason, error) {
	j := new(Jason)
	j.root = true
	d := json.NewDecoder(reader)
	err := d.Decode(&j.data)
	return j, err
}

// Create a new instance from bytes
// Returns an error if the bytes couldn't be parsed.
func NewFromBytes(b []byte) (*Jason, error) {
	j := new(Jason)
	j.root = true
	err := json.Unmarshal(b, &j.data)
	return j, err
}

// Create a new instance from a string
// Returns an error if the string couldn't be parsed.
func NewFromString(s string) (*Jason, error) {
	b := []byte(s)
	return NewFromBytes(b)
}

// Returns true if this key exists
// Example: j.Get("address").Exists()
func (j *Jason) Exists() bool {
	return j.exists
}

// Marshal into bytes
func (j *Jason) Marshal() ([]byte, error) {
	return json.Marshal(j.data)
}

// Private Get
func (j *Jason) get(key string) *Jason {

	// Assume this is an object
	obj := j.object()

	// Only continue if it really is an object
	if obj.Valid {
		child, ok := obj.Map[key]
		if ok {
			return child
		}
	}

	return &Jason{nil, false, false}

}

// Private to get path
func (j *Jason) getPath(keys []string) *Jason {
	current := j
	for _, key := range keys {
		current = current.get(key)
	}
	return current
}

// Get key or key path. Returns a new Jason instance.
// Example: Get("address", "street")
func (j *Jason) Get(keys ...string) *Jason {
	return j.getPath(keys)
}

// Determine if key path exists
func (j *Jason) Has(keys ...string) bool {
	return j.getPath(keys).Exists()
}

func (j *Jason) null() *jNull {

	var valid bool

	// Check the type of this data
	switch j.data.(type) {
	case nil:
		valid = true
		break
	}

	n := new(jNull)
	n.Valid = valid && j.exists // We also need to check that it actually exists here to separate nil and non-existing values

	return n
}

// Returns true if the instance is actually a JSON null object.
func (j *Jason) IsNull() bool {
	n := j.null()
	return n.Valid
}

func (j *Jason) array() *jArray {

	var valid bool

	// Check the type of this data
	switch j.data.(type) {
	case []interface{}:
		valid = true
		break
	}

	a := new(jArray)
	a.Valid = valid

	// Unsure if this is a good way to use slices, it's probably not
	var slice []*Jason

	if valid {

		for _, element := range j.data.([]interface{}) {
			child := Jason{element, true, false}
			slice = append(slice, &child)
		}
	}

	a.Slice = slice

	return a
}

// Returns the current data as an array of Jason values.
// Fallbacks on empty array
// Check IsArray() before using if you want to know.
func (j *Jason) Array() []*Jason {
	a := j.array()
	return a.Slice
}

// Returns true if the instance is actually a JSON array.
func (j *Jason) IsArray() bool {
	a := j.array()
	return a.Valid
}

func (j *Jason) number() *jNumber {

	var valid bool

	// Check the type of this data
	switch j.data.(type) {
	case float64:
		valid = true
		break
	}

	n := new(jNumber)
	n.Valid = valid

	if valid {
		n.Float64 = j.data.(float64)
	}

	return n
}

func (j *Jason) Number() float64 {
	n := j.number()
	return n.Float64
}

// Returns the same as Number()
func (j *Jason) Float64() float64 {
	return j.Number()
}

// Returns the Number() converted to an int64
func (j *Jason) Int64() int64 {
	return int64(j.Number())
}

// Returns true if the instance is actually a JSON number.
func (j *Jason) IsNumber() bool {
	n := j.number()
	return n.Valid
}

// Private
func (j *Jason) boolean() *jBool {

	var valid bool

	// Check the type of this data
	switch j.data.(type) {
	case bool:
		valid = true
		break
	}

	b := new(jBool)
	b.Valid = valid

	if valid {
		b.Bool = j.data.(bool)
	}

	return b
}

// Returns true if the instance is actually a JSON bool.
func (j *Jason) IsBoolean() bool {
	b := j.boolean()
	return b.Valid
}

// Returns true if the instance is actually a JSON bool.
func (j *Jason) Boolean() bool {
	b := j.boolean()
	return b.Bool
}

// Private object
func (j *Jason) object() *jObject {

	var valid bool

	// Check the type of this data
	switch j.data.(type) {
	case map[string]interface{}:
		valid = true
		break
	}

	obj := new(jObject)
	obj.Valid = valid

	m := make(map[string]*Jason)

	if valid {
		//obj.Map = j.data.(map[string]interface{})

		for key, element := range j.data.(map[string]interface{}) {
			m[key] = &Jason{element, true, false}
		}
	}

	obj.Map = m

	return obj
}

// Returns the current data as objects with string keys and Jason values.
// Fallbacks on empty map if invalid.
// Check IsObject() before using if you want to know.
func (j *Jason) Object() map[string]*Jason {
	obj := j.object()
	return obj.Map
}

// Returns true if the instance is actually a JSON object
func (j *Jason) IsObject() bool {
	obj := j.object()
	return obj.Valid
}

func (j *Jason) sstring() *jString {

	var valid bool

	// Check the type of this data
	switch j.data.(type) {
	case string:
		valid = true
		break
	}

	s := new(jString)
	s.Valid = valid

	if valid {
		s.String = j.data.(string)
	}

	return s
}

// Returns the current data as string. Fallbacks on empty string if invalid.
// Check IsString() before using if you want to know.
// Note: This is also the method used by log to print contents,
// so that's why you need to use Log() instead when printing
func (j *Jason) String() string {

	// If j is the root node, it can never be a string
	// Since log and fmt uses this method to log value, we should return something nice in those cases
	// Give kind reminder if this is the root node.
	if j.root {
		return "Note: Jason instances cannot be printed due to String() method name already being used by this library. Use Log() instead."
	} else {
		s := j.sstring()
		return s.String
	}

}

// Use this method when logging.
//
// The second version below will not work since log uses String() method that we are already using.
// DO: log.Println("root: ", root.Log())
// DO NOT: log.Println("root: ", root)

func (j *Jason) Log() string {
	f, err := json.Marshal(j.data)

	if err != nil {
		return err.Error()
	} else {
		return string(f)
	}
}

// Returns true if the object is actually an object
func (j *Jason) IsString() bool {
	s := j.sstring()
	return s.Valid
}
