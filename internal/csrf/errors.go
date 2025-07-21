package csrf

const (
	ErrCSRFValidation = 1018
)

var ErrorMessages = map[int]string{
	ErrCSRFValidation: "CSRF validation failed",
}