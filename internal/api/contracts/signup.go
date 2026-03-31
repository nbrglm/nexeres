package contracts

import "github.com/nbrglm/nexeres/internal/models"

// AuthSignupRequest represents the request payload for user signup
type AuthSignupRequest struct {
	// The email of the user signing up
	Email string `json:"email" validate:"required,email"`

	// The password of the user signing up, must meet complexity requirements:
	// - Minimum 8 characters
	// - At least one uppercase letter
	// - At least one lowercase letter
	// - At least one digit
	// - At least one special character, allowed ones: -_*@.
	Password string `json:"password" validate:"required"`

	// The confirmation of the password, must match the Password field
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=Password"`

	// Required full name of the user
	Name string `json:"name" validate:"required,min=2,max=512"`

	// Optional invite token if available, required in multi-tenant mode
	InviteToken *string `json:"inviteToken,omitempty"`
}

// AuthSignupResponse represents the response payload after a successful signup
type AuthSignupResponse struct {
	// Base response structure
	BaseResponse

	// ID of the newly created user
	UserID string `json:"userID"`

	// Backup codes generated for the user, if MFA is enabled during signup due to policy enforcement
	BackupCodes *models.BackupCodes `json:"backupCodes,omitempty"`
}
