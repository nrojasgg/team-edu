-- Datos iniciales para la base de datos de autenticación

-- Insertar roles por defecto
INSERT INTO roles (name) VALUES ('admin') ON CONFLICT (name) DO NOTHING;
INSERT INTO roles (name) VALUES ('coordinador') ON CONFLICT (name) DO NOTHING;
INSERT INTO roles (name) VALUES ('profesor') ON CONFLICT (name) DO NOTHING;
INSERT INTO roles (name) VALUES ('ayudante') ON CONFLICT (name) DO NOTHING;
INSERT INTO roles (name) VALUES ('estudiante') ON CONFLICT (name) DO NOTHING;

-- Permisos legacy (mantenidos por compatibilidad)
INSERT INTO permissions (name) VALUES ('read_users') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('write_users') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('delete_users') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('read_courses') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('write_courses') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('delete_courses') ON CONFLICT (name) DO NOTHING;

-- Permisos de usuarios
INSERT INTO permissions (name) VALUES ('Crear estudiante') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Crear profesor') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Crear coordinador') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Crear administrador') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Editar estudiante') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Editar profesor') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Editar coordinador') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Editar administrador') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Eliminar usuario') ON CONFLICT (name) DO NOTHING;

-- Permisos de semestres
INSERT INTO permissions (name) VALUES ('Editar semestre actual') ON CONFLICT (name) DO NOTHING;

-- Permisos de cursos
INSERT INTO permissions (name) VALUES ('Acceder a cursos vistos') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Ver todos los cursos') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Crear curso') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Editar curso') ON CONFLICT (name) DO NOTHING;

-- Permisos de secciones
INSERT INTO permissions (name) VALUES ('Crear sección') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Crear código de invitación') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Editar sección') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Ver secciones propias') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Ver todas las secciones') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Agregar estudiante a la sección') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Usar código de invitación') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Revocar código de invitación') ON CONFLICT (name) DO NOTHING;

-- Permisos de cuestionarios
INSERT INTO permissions (name) VALUES ('Crear cuestionario') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Acceder a cuestionario vistas') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Ver todas las cuestionario') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Editar cuestionario') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Crear cuestionario para curso') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Crear cuestionario para sección') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Crear cuestionario para grupo') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Ver todas las respuestas en cuestionario') ON CONFLICT (name) DO NOTHING;

-- Permisos de encuestas
INSERT INTO permissions (name) VALUES ('Acceder a encuestas vistas') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Ver todas las encuestas') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Editar encuesta') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Crear encuesta para curso') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Crear encuesta para sección') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Crear encuesta para grupo') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Ver todas las respuestas en encuesta') ON CONFLICT (name) DO NOTHING;

-- Permisos de grupos
INSERT INTO permissions (name) VALUES ('Ver grupos propios') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Editar grupos') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Crear grupos aleatorios') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name) VALUES ('Crear grupos con plugins') ON CONFLICT (name) DO NOTHING;

-- Insertar usuarios de ejemplo
-- Usuario administrador por defecto
INSERT INTO users (nombres, paterno, materno, email, password, rut, role_id, career_code)
SELECT 'Admin', 'Sistema', 'TeamEdu', 'admin@TeamEdu.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '12345678-9', r.id, NULL
FROM roles r
WHERE r.name = 'admin'
ON CONFLICT (email) DO NOTHING;

-- Usuarios estudiantes de ejemplo
INSERT INTO users (nombres, paterno, materno, email, password, rut, role_id, career_code)
SELECT 'Juan Carlos', 'Pérez', 'González', 'juan.perez@estudiante.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '19876543-2', r.id, '1353'
FROM roles r
WHERE r.name = 'estudiante'
ON CONFLICT (email) DO NOTHING;

INSERT INTO users (nombres, paterno, materno, email, password, rut, role_id, career_code)
SELECT 'María Elena', 'González', 'Rivera', 'maria.gonzalez@estudiante.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '18765432-1', r.id, '1353'
FROM roles r
WHERE r.name = 'estudiante'
ON CONFLICT (email) DO NOTHING;

INSERT INTO users (nombres, paterno, materno, email, password, rut, role_id, career_code)
SELECT 'Carlos Andrés', 'Rodríguez', 'Morales', 'carlos.rodriguez@estudiante.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '17654321-0', r.id, '1354'
FROM roles r
WHERE r.name = 'estudiante'
ON CONFLICT (email) DO NOTHING;

INSERT INTO users (nombres, paterno, materno, email, password, rut, role_id, career_code)
SELECT 'Ana Sofía', 'Martínez', 'Herrera', 'ana.martinez@estudiante.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '16543210-9', r.id, '1355'
FROM roles r
WHERE r.name = 'estudiante'
ON CONFLICT (email) DO NOTHING;

INSERT INTO users (nombres, paterno, materno, email, password, rut, role_id, career_code)
SELECT 'Luis Fernando', 'López', 'Castro', 'luis.lopez@estudiante.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '15432109-8', r.id, '1353'
FROM roles r
WHERE r.name = 'estudiante'
ON CONFLICT (email) DO NOTHING;

-- Usuarios profesores de ejemplo
INSERT INTO users (nombres, paterno, materno, email, password, rut, role_id, career_code)
SELECT 'Roberto', 'Silva', 'Vargas', 'roberto.silva@profesor.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '14321098-7', r.id, NULL
FROM roles r
WHERE r.name = 'profesor'
ON CONFLICT (email) DO NOTHING;

INSERT INTO users (nombres, paterno, materno, email, password, rut, role_id, career_code)
SELECT 'Carmen', 'Torres', 'Mendoza', 'carmen.torres@profesor.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '13210987-6', r.id, NULL
FROM roles r
WHERE r.name = 'profesor'
ON CONFLICT (email) DO NOTHING;

-- =============================================================
-- ASIGNACIÓN DE PERMISOS POR ROL
-- =============================================================

-- Admin: todos los permisos
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- Coordinador
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'coordinador'
  AND p.name IN (
    'Crear estudiante',
    'Crear profesor',
    'Editar estudiante',
    'Editar profesor',
    'Acceder a cursos vistos',
    'Crear curso',
    'Editar curso',
    'Crear sección',
    'Editar sección',
    'Ver secciones propias',
    'Agregar estudiante a la sección',
    'Crear cuestionario',
    'Acceder a cuestionario vistas',
    'Ver todas las cuestionario',
    'Editar cuestionario',
    'Crear cuestionario para curso',
    'Crear cuestionario para sección',
    'Crear cuestionario para grupo',
    'Ver todas las respuestas en cuestionario',
    'Acceder a encuestas vistas',
    'Ver todas las encuestas',
    'Editar encuesta',
    'Crear encuesta para curso',
    'Crear encuesta para sección',
    'Crear encuesta para grupo',
    'Ver todas las respuestas en encuesta',
    'Ver grupos propios',
    'Editar grupos',
    'Crear grupos aleatorios',
    'Crear grupos con plugins'
  )
ON CONFLICT DO NOTHING;

-- Profesor
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'profesor'
  AND p.name IN (
    'Crear estudiante',
    'Editar estudiante',
    'Acceder a cursos vistos',
    'Crear código de invitación',
    'Usar código de invitación',
    'Revocar código de invitación',
    'Ver secciones propias',
    'Agregar estudiante a la sección',
    'Crear cuestionario',
    'Acceder a cuestionario vistas',
    'Editar cuestionario',
    'Crear cuestionario para curso',
    'Crear cuestionario para sección',
    'Crear cuestionario para grupo',
    'Ver todas las respuestas en cuestionario',
    'Acceder a encuestas vistas',
    'Editar encuesta',
    'Crear encuesta para curso',
    'Crear encuesta para sección',
    'Crear encuesta para grupo',
    'Ver todas las respuestas en encuesta',
    'Ver grupos propios',
    'Editar grupos',
    'Crear grupos aleatorios',
    'Crear grupos con plugins'
  )
ON CONFLICT DO NOTHING;

-- Ayudante (igual que profesor, sin Crear/Editar estudiante)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'ayudante'
  AND p.name IN (
    'Acceder a cursos vistos',
    'Crear código de invitación',
    'Usar código de invitación',
    'Revocar código de invitación',
    'Ver secciones propias',
    'Agregar estudiante a la sección',
    'Crear cuestionario',
    'Acceder a cuestionario vistas',
    'Editar cuestionario',
    'Crear cuestionario para curso',
    'Crear cuestionario para sección',
    'Crear cuestionario para grupo',
    'Ver todas las respuestas en cuestionario',
    'Acceder a encuestas vistas',
    'Editar encuesta',
    'Crear encuesta para curso',
    'Crear encuesta para sección',
    'Crear encuesta para grupo',
    'Ver todas las respuestas en encuesta',
    'Ver grupos propios',
    'Editar grupos',
    'Crear grupos aleatorios',
    'Crear grupos con plugins'
  )
ON CONFLICT DO NOTHING;

-- Estudiante
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'estudiante'
  AND p.name IN (
    'Acceder a cursos vistas',
    'Ver secciones propias',
    'Acceder a cuestionario vistas',
    'Acceder a encuestas vistas',
    'Ver grupos propios',
    'Usar código de invitación'
  )
ON CONFLICT DO NOTHING;
