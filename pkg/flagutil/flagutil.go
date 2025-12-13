package flagutil

import "strings"

type Optional struct {
	Name        string
	Description string
}

// NewOptional creates a new Optional with the specified name and description.
func NewOptional(name, desc string) *Optional {
	return &Optional{
		Name:        name,
		Description: desc,
	}
}

// Set creates a string set of multiple optionals with the given separator.
func Set(sep string, o ...*Optional) string {
	n := make([]string, len(o))
	for i, o := range o {
		if o != nil {
			n[i] = o.Name
		}
	}

	return strings.Join(n, sep)
}

// IsIn checks if the receiver name exists within the flag string.
func (o *Optional) IsIn(flag string) bool {
	return strings.Contains(flag, o.Name)
}

// Format returns a string representation of the receiver.
func (o *Optional) Format() string {
	return o.Name + ": " + o.Description
}
