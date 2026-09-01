package service

// ValidationError señala que una regla de negocio (no de persistencia)
// rechazó la operación. handlers la mapea a HTTP 400, igual que
// store.StoreError se mapea según su Typ.
type ValidationError struct {
	msg string
}

func NewValidationError(msg string) ValidationError {
	return ValidationError{msg: msg}
}

func (e ValidationError) Error() string {
	return e.msg
}
