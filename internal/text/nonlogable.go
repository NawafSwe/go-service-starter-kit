package text

import (
	"encoding/json"
	"fmt"
)

// NonLoggable wraps a string that must never appear in logs or JSON output.
type NonLoggable string

var (
	_ fmt.Stringer   = NonLoggable("")
	_ fmt.GoStringer = NonLoggable("")
	_ json.Marshaler = NonLoggable("")
)

func (nl NonLoggable) String() string               { return "redacted" }
func (nl NonLoggable) GoString() string             { return "redacted" }
func (nl NonLoggable) MarshalJSON() ([]byte, error) { return []byte(`"redacted"`), nil }
func (nl NonLoggable) GetValue() string             { return string(nl) }
