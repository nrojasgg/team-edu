package repositories

import (
	"database/sql"
	"ms-registro-curricular/internal/models"
)

type RoleRepository interface {
	Create(role *models.Role) error
	GetByID(id int64) (*models.Role, error)
	GetByName(name string) (*models.Role, error)
	Update(role *models.Role) error
	Delete(id int64) error
	List() ([]*models.Role, error)
}

type roleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(role *models.Role) error {
	query := `INSERT INTO roles (name) VALUES ($1) RETURNING id`
	err := r.db.QueryRow(query, role.Name).Scan(&role.ID)
	return err
}

func (r *roleRepository) GetByID(id int64) (*models.Role, error) {
	query := `
		SELECT r.id, r.name, COUNT(rp.permission_id) as permission_count
		FROM roles r
		LEFT JOIN role_permissions rp ON r.id = rp.role_id
		WHERE r.id = $1
		GROUP BY r.id, r.name
	`
	role := &models.Role{}
	err := r.db.QueryRow(query, id).Scan(&role.ID, &role.Name, &role.PermissionCount)
	return role, err
}

func (r *roleRepository) GetByName(name string) (*models.Role, error) {
	query := `
		SELECT r.id, r.name, COUNT(rp.permission_id) as permission_count
		FROM roles r
		LEFT JOIN role_permissions rp ON r.id = rp.role_id
		WHERE r.name = $1
		GROUP BY r.id, r.name
	`
	role := &models.Role{}
	err := r.db.QueryRow(query, name).Scan(&role.ID, &role.Name, &role.PermissionCount)
	return role, err
}

func (r *roleRepository) Update(role *models.Role) error {
	query := `UPDATE roles SET name = $1 WHERE id = $2`
	_, err := r.db.Exec(query, role.Name, role.ID)
	return err
}

func (r *roleRepository) Delete(id int64) error {
	query := `DELETE FROM roles WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *roleRepository) List() ([]*models.Role, error) {
	query := `
		SELECT r.id, r.name, COUNT(rp.permission_id) as permission_count
		FROM roles r
		LEFT JOIN role_permissions rp ON r.id = rp.role_id
		GROUP BY r.id, r.name
		ORDER BY r.name
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		err := rows.Scan(&role.ID, &role.Name, &role.PermissionCount)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}
