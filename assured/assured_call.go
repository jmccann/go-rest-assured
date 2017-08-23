package assured

import (
	"fmt"
)

// Call is a structure containing a request that is stubbed or made
type Call struct {
	Path       string
	Method     string
	StatusCode int
	Response   []byte
}

// ID is used as a key when managing stubbed and made calls
func (c Call) ID() string {
	return fmt.Sprintf("%s:%s", c.Method, c.Path)
}

// String converts a Call's Response into a string
func (c Call) String() string {
	rawString := string(c.Response)

	// TODO: implement string replacements for special cases
	return rawString
}