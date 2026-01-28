package auth

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	MFARequired  bool   `json:"mfaRequired"`
	PreAuthToken string `json:"preAuthToken"`
}

type RefreshTokensRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type RefreshTokensResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type ResetPasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=6"`
}

type SetupMFAResponse struct {
	Secret     string `json:"secret"`
	OtpAuthURL string `json:"otpAuthUrl"`
}

type EnableMFARequest struct {
	Code string `json:"code" binding:"required"`
}

type EnableMFAResponse struct {
	BackupCodes []string `json:"backupCodes"`
}

type VerifyMFARequest struct {
	Code         string `json:"code" binding:"required"`
	PreAuthToken string `json:"preAuthToken" binding:"required"`
}

type VerifyMFAResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type DisableMFARequest struct {
	Password string `json:"password" binding:"required"`
}
