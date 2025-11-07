package reserr

import (
	"fmt"
	"net/http"

	"github.com/nbrglm/nexeres/opts"
)

// BadRequest creates a new ErrorResponse instance with HTTP 400 Bad Request status code.
//
// Note: If 1 msg is provided, it is used as the debug message.
// If 2 msgs are provided, the first is used as the main message and the second as the debug message.
func BadRequest(msgs ...string) *ErrorResponse {
	main := "Bad Request"
	debug := "The request could not be understood or was missing required parameters."
	if len(msgs) == 1 {
		debug = msgs[0]
	} else if len(msgs) >= 2 {
		main = msgs[0]
		debug = msgs[1]
	}
	return New(main, debug, http.StatusBadRequest, nil)
}

// Unauthorized creates a new ErrorResponse instance with HTTP 401 Unauthorized status code.
//
// Note: If 1 msg is provided, it is used as the debug message.
// If 2 msgs are provided, the first is used as the main message and the second as the debug message.
func Unauthorized(msgs ...string) *ErrorResponse {
	main := "Unauthorized"
	debug := "Authentication is required and has failed or has not yet been provided."
	if len(msgs) == 1 {
		debug = msgs[0]
	} else if len(msgs) >= 2 {
		main = msgs[0]
		debug = msgs[1]
	}
	return New(main, debug, http.StatusUnauthorized, nil)
}

// Forbidden creates a new ErrorResponse instance with HTTP 403 Forbidden status code.
//
// Note: If 1 msg is provided, it is used as the debug message.
// If 2 msgs are provided, the first is used as the main message and the second as the debug message.
func Forbidden(msgs ...string) *ErrorResponse {
	main := "Forbidden"
	debug := "You do not have permission to access this resource."
	if len(msgs) == 1 {
		debug = msgs[0]
	} else if len(msgs) >= 2 {
		main = msgs[0]
		debug = msgs[1]
	}
	return New(main, debug, http.StatusForbidden, nil)
}

// InternalServerError creates a new ErrorResponse instance with HTTP 500 Internal Server Error code and attaches the debug message and underlying error.
//
// You need to invoke .Filter() before you return an error response.
func InternalServerError(underlying error, debug string) *ErrorResponse {
	return New("Internal Server Error", debug, http.StatusInternalServerError, underlying)
}

const GenericMessage = "An error occurred while processing your request. Please try again later."

// NewGeneric creates a new ErrorResponse instance with a generic error message.
func NewGeneric(opName string, err error) *ErrorResponse {
	return &ErrorResponse{
		Message:         GenericMessage,
		DebugMessage:    "Failed to handle " + opName,
		Code:            http.StatusInternalServerError,
		UnderlyingError: err,
	}
}

// New creates a new ErrorResponse instance.
//
// It takes a user-friendly message, a debug message for developers, and an error code.
//
// The debug message is intended for internal use and should not be exposed to end users in production environments.
// For doing that, use the Filter method on the ErrorResponse instance when passing it to gin or the client.
//
// Also, the "Debug" Message is used to record things in span context and/or logs for observability purposes,
// so do not include sensitive information in the debug message.
//
// If underlying error is not nil, it will be logged at ERROR level for debugging purposes,
// If underlying error is nil, no error will be logged.
//
// Note: To specify the `RetryUrl`, `RedirectUrl`, and `RetryButtonText` fields, use the `errResponse.WithUI()` function.
func New(message, debug string, code int, underlying error) *ErrorResponse {
	return &ErrorResponse{
		Message:         message,
		DebugMessage:    debug,
		Code:            code,
		UnderlyingError: underlying,
	}
}

type ErrorResponse struct {
	// Message is a user-friendly message that can be displayed to the end user.
	Message string `json:"message"`
	// DebugMessage is a technical message that can be used for debugging.
	DebugMessage string `json:"debug"`

	// HTTP Status code associated with this error.
	// It is not serialized to JSON. This field is useful for setting the HTTP status code in the response.
	Code int `json:"-"`

	// UnderlyingError is an optional field that can hold the original error
	// that caused this error response. It is not serialized to JSON.
	// This field is useful for logging and debugging purposes.
	UnderlyingError error `json:"-"`
}

func (e *ErrorResponse) Error() string {
	if opts.Debug {
		return fmt.Sprintf("Error %d: %s (Debug: %s)", e.Code, e.Message, e.DebugMessage)
	}
	return e.Message
}

// Filter filters the ErrorResponse based on the debug mode.
// If debug mode is enabled, it returns the full error response including the debug message.
// If debug mode is not enabled, it returns a filtered error response without the debug message.
// This is useful for controlling the visibility of debug information in a production environment.
func (e *ErrorResponse) Filter() *ErrorResponse {
	if opts.Debug || e.DebugMessage == "" {
		return e
	}

	// Otherwise, return a filtered error response without the debug message
	return &ErrorResponse{
		Message:      e.Message,
		DebugMessage: "",
		Code:         e.Code,
	}
}
