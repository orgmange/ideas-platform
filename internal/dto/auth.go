package dto

type VerifyOTPRequest struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
	Name  string `json:"name,omitempty"`
}
