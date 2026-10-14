package models

import "time"

type InvitationCode struct {
	ID        int64     `json:"id"`
	SectionID int64     `json:"section_id"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	MaxUses   *int      `json:"max_uses,omitempty"`
	UsesCount int       `json:"uses_count"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy int64     `json:"created_by"`
}

type CreateInvitationCodeRequest struct {
	SectionID int64     `json:"section_id"`
	ExpiresAt time.Time `json:"expires_at"`
	MaxUses   *int      `json:"max_uses,omitempty"`
}

type UseInvitationCodeRequest struct {
	Code string `json:"code"`
}

type InvitationCodeUse struct {
	ID               int64     `json:"id"`
	InvitationCodeID int64     `json:"invitation_code_id"`
	StudentID        int64     `json:"student_id"`
	StudentName      string    `json:"student_name,omitempty"`
	StudentEmail     string    `json:"student_email,omitempty"`
	StudentRUT       string    `json:"student_rut,omitempty"`
	UsedAt           time.Time `json:"used_at"`
}
