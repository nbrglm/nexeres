package contracts

// BaseResponse defines the standard structure for API responses.
// It includes a success flag and a message.
// This will be embedded in all non-error response structures.
type BaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
