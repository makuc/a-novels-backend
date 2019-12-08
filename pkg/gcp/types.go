package gcp

import "time"

// UnknownValues is a helper type for figuring out structure of parsed data (by logging it)
type UnknownValues struct {
	Fields map[string]interface{} `json:"fields"`
}

// FieldPaths is a type for parsing `fieldPaths` data from Firestore Events
type FieldPaths struct {
	FieldPaths []string `json:"fieldPaths"`
}

// StringValue is a type for parsing `string` type from Firestore Events
type StringValue struct {
	StringValue string `json:"stringValue"`
}

// IntegerValue is a type for parsing `integer` type from Firestore Events
type IntegerValue struct {
	IntegerValue int64 `json:"integerValue"`
}

// TimestampValue is a type for parsing `Timestamp` type from Firestore Events
type TimestampValue struct {
	TimestampValue time.Time `json:"timestampValue"`
}

// BooleanValue is a type for parsing `boolean` type from Firestore Events
type BooleanValue struct {
	BooleanValue bool `json:"booleanValue"`
}
