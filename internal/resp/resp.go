package resp

type BaseResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}
