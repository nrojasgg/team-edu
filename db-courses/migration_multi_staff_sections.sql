-- Migration: Support multiple professors and assistants per section
-- Date: 2026-10-14

-- 1. Tabla para múltiples profesores por sección
CREATE TABLE IF NOT EXISTS section_professors (
    section_id bigint NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    professor_id bigint NOT NULL,
    PRIMARY KEY (section_id, professor_id)
);

CREATE INDEX IF NOT EXISTS idx_section_professors_section ON section_professors(section_id);
CREATE INDEX IF NOT EXISTS idx_section_professors_prof ON section_professors(professor_id);

-- 2. Tabla para múltiples ayudantes por sección
CREATE TABLE IF NOT EXISTS section_assistants (
    section_id bigint NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    assistant_id bigint NOT NULL,
    PRIMARY KEY (section_id, assistant_id)
);

CREATE INDEX IF NOT EXISTS idx_section_assistants_section ON section_assistants(section_id);
CREATE INDEX IF NOT EXISTS idx_section_assistants_ast ON section_assistants(assistant_id);

-- 3. Migrar asignaciones existentes de la tabla sections
INSERT INTO section_professors (section_id, professor_id)
SELECT id, professor_id FROM sections WHERE professor_id IS NOT NULL
ON CONFLICT DO NOTHING;

INSERT INTO section_assistants (section_id, assistant_id)
SELECT id, assistant_id FROM sections WHERE assistant_id IS NOT NULL
ON CONFLICT DO NOTHING;
