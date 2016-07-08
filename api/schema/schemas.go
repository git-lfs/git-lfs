// schema provides a testing utility for testing API types against a predefined
// JSON schema.
//
// The core philosophy for this package is as follows: when a new API is
// accepted, JSON Schema files should be added to document the types that are
// exchanged over this new API. Those files are placed in the `/api/schema`
// directory, and are used by the schema.Validate function to test that
// particular instances of these types as represented in Go match the predefined
// schema that was proposed as a part of the API.
//
// For ease of use, this file defines several constants, one for each schema
// file's name, to easily pass around during tests.
//
// As briefly described above, to validate that a Go type matches the schema for
// a particular API call, one should use the schema.Validate() function.
package schema

const (
	LockListSchema       = "lock_list_schema.json"
	LockRequestSchema    = "lock_request_schema.json"
	LockResponseSchema   = "lock_response_schema.json"
	UnlockRequestSchema  = "unlock_request_schema.json"
	UnlockResponseSchema = "unlock_response_schema.json"
)
