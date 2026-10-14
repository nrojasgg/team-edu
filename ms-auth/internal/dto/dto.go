package dto

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type GoogleAuthRequest struct {
	Code string `json:"code" validate:"required"`
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type RegisterRequest struct {
	Nombres    *string `json:"nombres,omitempty"`
	Paterno    *string `json:"paterno,omitempty"`
	Materno    *string `json:"materno,omitempty"`
	RUT        *string `json:"rut,omitempty"`
	Email      string  `json:"email" validate:"required,email"`
	Password   string  `json:"password" validate:"required,min=6"`
	RoleID     *int64  `json:"role_id,omitempty"`
	CareerCode *string `json:"career_code,omitempty"`
}

type UserResponse struct {
	ID         int64   `json:"id"`
	Nombres    string  `json:"nombres"`
	Paterno    string  `json:"paterno"`
	Materno    string  `json:"materno"`
	FullName   string  `json:"full_name"` // Campo calculado para compatibilidad
	RUT        string  `json:"rut"`
	Email      string  `json:"email"`
	PictureURL string  `json:"picture_url"`
	RoleID     int64   `json:"role_id"`
	CareerCode *string `json:"career_code"`
	RoleName   string  `json:"role_name,omitempty"`
}

type CreateUserRequest struct {
	Nombres    string  `json:"nombres" validate:"required"`
	Paterno    string  `json:"paterno" validate:"required"`
	Materno    string  `json:"materno"`
	RUT        string  `json:"rut" validate:"required"`
	Email      string  `json:"email" validate:"required,email"`
	Password   string  `json:"password" validate:"required,min=6"`
	RoleID     int64   `json:"role_id" validate:"required"`
	CareerCode *string `json:"career_code"`
}

type UpdateUserRequest struct {
	Nombres    string  `json:"nombres" validate:"required"`
	Paterno    string  `json:"paterno" validate:"required"`
	Materno    string  `json:"materno"`
	RUT        string  `json:"rut" validate:"required"`
	Email      string  `json:"email" validate:"required,email"`
	RoleID     int64   `json:"role_id" validate:"required"`
	CareerCode *string `json:"career_code"`
}

type CreateRoleRequest struct {
	Name string `json:"name" validate:"required"`
}

type CreatePermissionRequest struct {
	Name string `json:"name" validate:"required"`
}

type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type ForgotPasswordRequest struct {
	Email    string `json:"email" validate:"required,email"`
	ResetURL string `json:"reset_url" validate:"required"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}
