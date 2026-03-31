package contracts

type VerifyEmailTokenRequest struct {
	Token string `json:"token"`
}

type VerifyEmailTokenResponse struct {
	BaseResponse
}
