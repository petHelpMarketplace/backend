package domain

// LoginReq represents the request body for user login.
// @Description User login request payload
type LoginReq struct {
	// Email address of the user.
	// required: true
	// format: email
	// example: user@example.com
	Email string `json:"email" binding:"required,email" example:"user@example.com"`

	// Password for the user account.
	// required: true
	// example: MySecretPassword123!
	Password string `json:"password" binding:"required" example:"MySecretPassword123!"`
}

// ChangePassReq represents the request body for changing a specialist's password.
// @Description Change password request payload
type ChangePassReq struct {
	// The specialist's current password.
	// required: true
	// example: MySecretPassword123!
	CurrentPass string `json:"current_password" binding:"required" validate:"required" example:"MySecretPassword123!"`

	// The new password for the specialist account.
	// Must be at least 12 characters long.
	// Must contain at least one uppercase letter, one lowercase letter, one number, and one special character from @$!%*?&.
	// required: true
	// minLength: 12
	// maxLength: 255
	// pattern: "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]{12,255}$"
	// example: NewStr0ngP@ssw0rd!
	NewPass string `json:"new_password" binding:"required,min=12" validate:"required,min=12" example:"NewStr0ngP@ssw0rd!"`
}

// SuccessResponse is a generic success response structure.
type SuccessResponse struct {
	Code    int    `json:"code" example:"200"`
	Message string `json:"message" example:"Operation was successful"`
}
