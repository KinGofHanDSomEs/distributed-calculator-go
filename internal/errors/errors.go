package errors

import "errors"

var (
	ErrInvalidData    = errors.New("invalid data")
	ErrServerSide     = errors.New("something went wrong")
	ErrNotFound       = errors.New("there is no such expression")
	ErrInvalidSymbol  = errors.New("invalid symbol")
	ErrOpeningBracket = errors.New("mismatched opening bracket")
	ErrClosingBracket = errors.New("mismatched closing bracket")
	ErrVariableValue  = errors.New("invalid environment variable value")
	ErrDivisionByZero = errors.New("division by zero is prohibited")
)
