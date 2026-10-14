package models

type Role struct {
	ID              int64  `json:"id" db:"id"`
	Name            string `json:"name" db:"name" validate:"required"`
	PermissionCount int    `json:"permission_count,omitempty" db:"permission_count"`
}

type Permission struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name" validate:"required"`
}

type RolePermission struct {
	RoleID       int64 `json:"role_id" db:"role_id"`
	PermissionID int64 `json:"permission_id" db:"permission_id"`
}

type User struct {
	ID         int64   `json:"id" db:"id"`
	Nombres    string  `json:"nombres" db:"nombres"`
	Paterno    string  `json:"paterno" db:"paterno"`
	Materno    string  `json:"materno" db:"materno"`
	RUT        string  `json:"rut" db:"rut"`
	Password   string  `json:"-" db:"password" validate:"required,min=6"`
	Email      string  `json:"email" db:"email" validate:"required,email"`
	PictureURL string  `json:"picture_url" db:"picture_url"`
	RoleID     int64   `json:"role_id" db:"role_id"`
	CareerCode *string `json:"career_code" db:"career_code"`
	Role       *Role   `json:"role,omitempty"`
}
