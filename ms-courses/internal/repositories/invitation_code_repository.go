package repositories

import (
	"database/sql"
	"math/rand"
	"ms-courses/internal/models"
	"time"
)

type InvitationCodeRepository struct {
	db *sql.DB
}

func NewInvitationCodeRepository(db *sql.DB) *InvitationCodeRepository {
	return &InvitationCodeRepository{db: db}
}

const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateCode(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(code)
}

func (r *InvitationCodeRepository) Create(req *models.CreateInvitationCodeRequest, professorID int64) (*models.InvitationCode, error) {
	code := generateCode(8)

	query := `
		INSERT INTO invitation_codes (section_id, code, expires_at, max_uses, uses_count, is_active, created_by)
		VALUES ($1, $2, $3, $4, 0, true, $5)
		RETURNING id, section_id, code, expires_at, max_uses, uses_count, is_active, created_at, created_by
	`

	var ic models.InvitationCode
	err := r.db.QueryRow(query, req.SectionID, code, req.ExpiresAt, req.MaxUses, professorID).Scan(
		&ic.ID, &ic.SectionID, &ic.Code, &ic.ExpiresAt, &ic.MaxUses, &ic.UsesCount, &ic.IsActive, &ic.CreatedAt, &ic.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	return &ic, nil
}

func (r *InvitationCodeRepository) GetByID(id int64) (*models.InvitationCode, error) {
	query := `
		SELECT id, section_id, code, expires_at, max_uses, uses_count, is_active, created_at, created_by
		FROM invitation_codes
		WHERE id = $1
	`

	var ic models.InvitationCode
	err := r.db.QueryRow(query, id).Scan(
		&ic.ID, &ic.SectionID, &ic.Code, &ic.ExpiresAt, &ic.MaxUses, &ic.UsesCount, &ic.IsActive, &ic.CreatedAt, &ic.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	return &ic, nil
}

func (r *InvitationCodeRepository) GetByCode(code string) (*models.InvitationCode, error) {
	query := `
		SELECT id, section_id, code, expires_at, max_uses, uses_count, is_active, created_at, created_by
		FROM invitation_codes
		WHERE code = $1
	`

	var ic models.InvitationCode
	err := r.db.QueryRow(query, code).Scan(
		&ic.ID, &ic.SectionID, &ic.Code, &ic.ExpiresAt, &ic.MaxUses, &ic.UsesCount, &ic.IsActive, &ic.CreatedAt, &ic.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	return &ic, nil
}

func (r *InvitationCodeRepository) GetActiveBySection(sectionID int64) ([]*models.InvitationCode, error) {
	query := `
		SELECT id, section_id, code, expires_at, max_uses, uses_count, is_active, created_at, created_by
		FROM invitation_codes
		WHERE section_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, sectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	codes := make([]*models.InvitationCode, 0)
	for rows.Next() {
		var ic models.InvitationCode
		err := rows.Scan(
			&ic.ID, &ic.SectionID, &ic.Code, &ic.ExpiresAt, &ic.MaxUses, &ic.UsesCount, &ic.IsActive, &ic.CreatedAt, &ic.CreatedBy,
		)
		if err != nil {
			return nil, err
		}
		codes = append(codes, &ic)
	}

	return codes, nil
}

func (r *InvitationCodeRepository) Deactivate(id int64) error {
	query := `UPDATE invitation_codes SET is_active = false WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *InvitationCodeRepository) IncrementUses(id int64) error {
	query := `UPDATE invitation_codes SET uses_count = uses_count + 1 WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *InvitationCodeRepository) GetUsesByCodeID(codeID int64) ([]*models.InvitationCodeUse, error) {
	query := `
		SELECT id, invitation_code_id, student_id, student_name, student_email, student_rut, used_at
		FROM invitation_code_uses
		WHERE invitation_code_id = $1
		ORDER BY used_at DESC
	`

	rows, err := r.db.Query(query, codeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uses []*models.InvitationCodeUse
	for rows.Next() {
		var u models.InvitationCodeUse
		err := rows.Scan(&u.ID, &u.InvitationCodeID, &u.StudentID, &u.StudentName, &u.StudentEmail, &u.StudentRUT, &u.UsedAt)
		if err != nil {
			return nil, err
		}
		uses = append(uses, &u)
	}

	return uses, nil
}

func (r *InvitationCodeRepository) AddStudentToSection(studentID, sectionID int64) error {
	query := `
		INSERT INTO student_sections (student_id, section_id)
		VALUES ($1, $2)
		ON CONFLICT (student_id, section_id) DO NOTHING
	`
	_, err := r.db.Exec(query, studentID, sectionID)
	return err
}

func (r *InvitationCodeRepository) RecordUse(codeID, studentID int64, studentName, studentEmail, studentRUT string) error {
	query := `
		INSERT INTO invitation_code_uses (invitation_code_id, student_id, student_name, student_email, student_rut)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(query, codeID, studentID, studentName, studentEmail, studentRUT)
	return err
}
