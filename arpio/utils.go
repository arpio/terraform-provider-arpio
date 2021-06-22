package arpio

// StringHolder holds a string value.  Passing a holder struct into functions
// is often more explicit than passing pointers to pointers.
type StringHolder struct {
	S string
}
