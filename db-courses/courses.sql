-- Base de datos para cursos y secciones
CREATE TABLE IF NOT EXISTS courses (
  id bigint primary key generated always as identity,
  code text not null unique,
  name text not null
);

CREATE TABLE IF NOT EXISTS semesters (
  id bigint primary key generated always as identity,
  year int not null,
  term int not null,
  is_active boolean not null default false,
  unique (year, term)
);

CREATE TABLE IF NOT EXISTS course_offerings (
  id bigint primary key generated always as identity,
  course_id bigint references courses (id),
  semester_id bigint references semesters (id),
  unique (course_id, semester_id)
);

CREATE TABLE IF NOT EXISTS sections (
  id bigint primary key generated always as identity,
  name text not null,
  course_offering_id bigint references course_offerings (id),
  professor_id bigint,
  assistant_id bigint
);

CREATE TABLE IF NOT EXISTS section_professors (
  section_id bigint NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
  professor_id bigint NOT NULL,
  PRIMARY KEY (section_id, professor_id)
);

CREATE INDEX IF NOT EXISTS idx_section_professors_section ON section_professors(section_id);
CREATE INDEX IF NOT EXISTS idx_section_professors_prof ON section_professors(professor_id);

CREATE TABLE IF NOT EXISTS section_assistants (
  section_id bigint NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
  assistant_id bigint NOT NULL,
  PRIMARY KEY (section_id, assistant_id)
);

CREATE INDEX IF NOT EXISTS idx_section_assistants_section ON section_assistants(section_id);
CREATE INDEX IF NOT EXISTS idx_section_assistants_ast ON section_assistants(assistant_id);

CREATE TABLE IF NOT EXISTS student_sections (
  student_id bigint not null, -- Referencia al usuario en la DB de auth
  section_id bigint references sections (id),
  primary key (student_id, section_id)
);

-- =====================================================
-- Códigos de invitación para secciones
-- =====================================================
CREATE TABLE IF NOT EXISTS invitation_codes (
    id          SERIAL PRIMARY KEY,
    section_id  INT NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    code        VARCHAR(8) UNIQUE NOT NULL,
    expires_at  TIMESTAMP NOT NULL,
    max_uses    INT DEFAULT NULL,          -- NULL = sin límite de usos
    uses_count  INT DEFAULT 0,
    is_active   BOOLEAN DEFAULT true,
    created_at  TIMESTAMP DEFAULT NOW(),
    created_by  INT NOT NULL
);

-- Índice para buscar por código rápidamente
CREATE INDEX IF NOT EXISTS idx_invitation_codes_code ON invitation_codes(code);

-- Índice para buscar códigos activos de una sección
CREATE INDEX IF NOT EXISTS idx_invitation_codes_section_active ON invitation_codes(section_id, is_active) WHERE is_active = true;

-- =====================================================
-- Registro de uso de códigos de invitación
-- =====================================================
CREATE TABLE IF NOT EXISTS invitation_code_uses (
    id                  SERIAL PRIMARY KEY,
    invitation_code_id INT NOT NULL REFERENCES invitation_codes(id) ON DELETE CASCADE,
    student_id          INT NOT NULL,
    used_at             TIMESTAMP DEFAULT NOW(),
    UNIQUE(invitation_code_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_invitation_code_uses_code ON invitation_code_uses(invitation_code_id);
