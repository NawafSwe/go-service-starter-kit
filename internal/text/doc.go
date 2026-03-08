// Package text contains string utility types for safe logging and serialisation.
//
// # NonLoggable
//
// NonLoggable wraps sensitive string values (tokens, passwords, secrets) so
// they are automatically redacted in logs, fmt output, and JSON marshalling:
//
//	type User struct {
//	    ID        string
//	    AuthToken text.NonLoggable  // always logged/marshalled as "redacted"
//	}
//
// To retrieve the raw value (e.g. for signing or comparison), call GetValue:
//
//	raw := user.AuthToken.GetValue()
package text
