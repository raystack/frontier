package deleter

import "fmt"

var (
	ErrDeleteNotAllowed = fmt.Errorf("deletion not allowed for billed accounts")
)
