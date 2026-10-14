package repositories

import (
	"database/sql"
	"ms-registro-curricular/internal/models"
)

type PermissionRepository interface {
	Create(permission *models.Permission) error
	GetByID(id int64) (*models.Permission, error)
	GetByName(name string) (*models.Permission, error)
	Update(permission *models.Permission) error
	Delete(id int64) error
	List() ([]*models.Permission, error)
	GetByRoleID(roleID int64) ([]*models.Permission, error)
	AssignToRole(roleID, permissionID int64) error
	RemoveFromRole(roleID, permissionID int64) error
}

type permissionRepository struct {
	db *sql.DB
}

func NewPermissionRepository(db *sql.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(permission *models.Permission) error {
	query := `INSERT INTO permissions (name) VALUES ($1) RETURNING id`
	err := r.db.QueryRow(query, permission.Name).Scan(&permission.ID)
	return err
}

func (r *permissionRepository) GetByID(id int64) (*models.Permission, error) {
	query := `SELECT id, name FROM permissions WHERE id = $1`
	permission := &models.Permission{}
	err := r.db.QueryRow(query, id).Scan(&permission.ID, &permission.Name)
	return permission, err
}

func (r *permissionRepository) GetByName(name string) (*models.Permission, error) {
	query := `SELECT id, name FROM permissions WHERE name = $1`
	permission := &models.Permission{}
	err := r.db.QueryRow(query, name).Scan(&permission.ID, &permission.Name)
	return permission, err
}

func (r *permissionRepository) Update(permission *models.Permission) error {
	query := `UPDATE permissions SET name = $1 WHERE id = $2`
	_, err := r.db.Exec(query, permission.Name, permission.ID)
	return err
}

func (r *permissionRepository) Delete(id int64) error {
	query := `DELETE FROM permissions WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *permissionRepository) List() ([]*models.Permission, error) {
	query := `SELECT id, name FROM permissions ORDER BY name`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*models.Permission
	for rows.Next() {
		permission := &models.Permission{}
		err := rows.Scan(&permission.ID, &permission.Name)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (r *permissionRepository) GetByRoleID(roleID int64) ([]*models.Permission, error) {
	query := `
		SELECT p.id, p.name
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.name`

	rows, err := r.db.Query(query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*models.Permission
	for rows.Next() {
		permission := &models.Permission{}
		err := rows.Scan(&permission.ID, &permission.Name)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (r *permissionRepository) AssignToRole(roleID, permissionID int64) error {
	query := `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(query, roleID, permissionID)
	return err
}

func (r *permissionRepository) RemoveFromRole(roleID, permissionID int64) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`
	_, err := r.db.Exec(query, roleID, permissionID)
	return err
}
