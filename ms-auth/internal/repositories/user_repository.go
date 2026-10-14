package repositories

import (
	"database/sql"
	"fmt"
	"ms-registro-curricular/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id int64) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByRUT(rut string) (*models.User, error)
	Update(user *models.User) error
	Delete(id int64) error
	List(limit, offset int) ([]*models.User, error)
	GetUserPermissions(userID int64) ([]*models.Permission, error)
	GetUsersBatch(ids []int64) ([]*models.User, error)
	Search(query string, roleID int, limit, offset int) ([]*models.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (nombres, paterno, materno, rut, password, email, picture_url, role_id, career_code)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	err := r.db.QueryRow(query, user.Nombres, user.Paterno, user.Materno, user.RUT, user.Password, user.Email, user.PictureURL, user.RoleID, user.CareerCode).Scan(&user.ID)
	return err
}

func (r *userRepository) GetByID(id int64) (*models.User, error) {
	query := `
		SELECT u.id, u.nombres, u.paterno, u.materno, u.rut, u.password, u.email, u.picture_url, u.role_id, u.career_code, 
		       COALESCE(r.name, '') as role_name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1`

	user := &models.User{Role: &models.Role{}}
	var roleName string
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Nombres, &user.Paterno, &user.Materno, &user.RUT, &user.Password,
		&user.Email, &user.PictureURL, &user.RoleID, &user.CareerCode, &roleName,
	)

	if err != nil {
		return nil, err
	}

	user.Role.Name = roleName
	return user, nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT u.id, u.nombres, u.paterno, u.materno, u.rut, u.password, u.email, u.picture_url, u.role_id, u.career_code,
		       COALESCE(r.name, '') as role_name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1`

	user := &models.User{Role: &models.Role{}}
	var roleName string
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Nombres, &user.Paterno, &user.Materno, &user.RUT, &user.Password,
		&user.Email, &user.PictureURL, &user.RoleID, &user.CareerCode, &roleName,
	)

	if err != nil {
		return nil, err
	}

	user.Role.Name = roleName
	return user, nil
}

func (r *userRepository) GetByRUT(rut string) (*models.User, error) {
	query := `
		SELECT u.id, u.nombres, u.paterno, u.materno, u.rut, u.password, u.email, u.picture_url, u.role_id, u.career_code,
		       COALESCE(r.name, '') as role_name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.rut = $1`

	user := &models.User{Role: &models.Role{}}
	var roleName string
	err := r.db.QueryRow(query, rut).Scan(
		&user.ID, &user.Nombres, &user.Paterno, &user.Materno, &user.RUT, &user.Password,
		&user.Email, &user.PictureURL, &user.RoleID, &user.CareerCode, &roleName,
	)

	if err != nil {
		return nil, err
	}

	user.Role.Name = roleName
	return user, nil
}

func (r *userRepository) Update(user *models.User) error {
	query := `
		UPDATE users 
		SET nombres = $1, paterno = $2, materno = $3, rut = $4, password = $5, email = $6, picture_url = $7, role_id = $8, career_code = $9
		WHERE id = $10`

	_, err := r.db.Exec(query, user.Nombres, user.Paterno, user.Materno, user.RUT, user.Password, user.Email, user.PictureURL, user.RoleID, user.CareerCode, user.ID)
	return err
}

func (r *userRepository) Delete(id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *userRepository) List(limit, offset int) ([]*models.User, error) {
	query := `
		SELECT u.id, u.nombres, u.paterno, u.materno, u.rut, u.password, u.email, u.picture_url, u.role_id, u.career_code,
		       COALESCE(r.name, '') as role_name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		ORDER BY u.id
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{Role: &models.Role{}}
		var roleName string
		var pictureURL sql.NullString
		err := rows.Scan(
			&user.ID, &user.Nombres, &user.Paterno, &user.Materno, &user.RUT, &user.Password,
			&user.Email, &pictureURL, &user.RoleID, &user.CareerCode, &roleName,
		)
		if err != nil {
			return nil, err
		}
		if pictureURL.Valid {
			user.PictureURL = pictureURL.String
		} else {
			user.PictureURL = ""
		}
		user.Role.Name = roleName
		users = append(users, user)
	}

	return users, nil
}

func (r *userRepository) GetUserPermissions(userID int64) ([]*models.Permission, error) {
	query := `
		SELECT p.id, p.name
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN users u ON rp.role_id = u.role_id
		WHERE u.id = $1
		ORDER BY p.name`

	rows, err := r.db.Query(query, userID)
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

func (r *userRepository) GetUsersBatch(ids []int64) ([]*models.User, error) {
	if len(ids) == 0 {
		return []*models.User{}, nil
	}

	// Create placeholders for the IN clause
	placeholders := ""
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "$" + fmt.Sprintf("%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT u.id, u.nombres, u.paterno, u.materno, u.rut, u.email, u.picture_url, u.role_id, u.career_code,
		       COALESCE(r.name, '') as role_name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id IN (%s)
		ORDER BY u.id`, placeholders)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{Role: &models.Role{}}
		var roleName string
		var pictureURL sql.NullString
		var rut sql.NullString
		err := rows.Scan(
			&user.ID, &user.Nombres, &user.Paterno, &user.Materno, &rut, &user.Email,
			&pictureURL, &user.RoleID, &user.CareerCode, &roleName,
		)
		if err != nil {
			return nil, err
		}
		if pictureURL.Valid {
			user.PictureURL = pictureURL.String
		} else {
			user.PictureURL = ""
		}
		if rut.Valid {
			user.RUT = rut.String
		} else {
			user.RUT = ""
		}
		user.Role.Name = roleName
		users = append(users, user)
	}

	return users, nil
}

func (r *userRepository) Search(searchTerm string, roleID int, limit, offset int) ([]*models.User, error) {
	whereClause := "1=1"
	var args []interface{}
	argCount := 1

	if searchTerm != "" {
		searchTerm = "%" + searchTerm + "%"
		whereClause += fmt.Sprintf(" AND (u.nombres ILIKE $%d OR u.paterno ILIKE $%d OR u.materno ILIKE $%d OR u.email ILIKE $%d OR u.rut ILIKE $%d OR u.career_code ILIKE $%d)",
			argCount, argCount, argCount, argCount, argCount, argCount)
		args = append(args, searchTerm)
		argCount++
	}

	if roleID > 0 {
		whereClause += fmt.Sprintf(" AND u.role_id = $%d", argCount)
		args = append(args, roleID)
		argCount++
	}

	query := fmt.Sprintf(`
		SELECT u.id, u.nombres, u.paterno, u.materno, u.rut, u.email, u.picture_url, u.role_id, u.career_code,
		       COALESCE(r.name, '') as role_name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE %s
		ORDER BY u.paterno, u.nombres
		LIMIT $%d OFFSET $%d`, whereClause, argCount, argCount+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{Role: &models.Role{}}
		var roleName string
		var pictureURL sql.NullString
		var rut sql.NullString
		err := rows.Scan(
			&user.ID, &user.Nombres, &user.Paterno, &user.Materno, &rut, &user.Email,
			&pictureURL, &user.RoleID, &user.CareerCode, &roleName,
		)
		if err != nil {
			return nil, err
		}
		if pictureURL.Valid {
			user.PictureURL = pictureURL.String
		}
		if rut.Valid {
			user.RUT = rut.String
		}
		user.Role.Name = roleName
		users = append(users, user)
	}

	return users, nil
}
