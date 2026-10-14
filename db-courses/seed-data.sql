-- Datos iniciales para la base de datos de cursos

-- Insertar semestres por defecto
INSERT INTO semesters (year, term, is_active) VALUES 
  (2024, 1, false),
  (2024, 2, false),
  (2025, 1, false),
  (2025, 2, true)
ON CONFLICT (year, term) DO UPDATE SET 
  is_active = EXCLUDED.is_active;

-- Insertar cursos de ejemplo
INSERT INTO courses (code, name) VALUES ('INF-101', 'Introducción a la Programación') ON CONFLICT (code) DO NOTHING;
INSERT INTO courses (code, name) VALUES ('INF-201', 'Programación Orientada a Objetos') ON CONFLICT (code) DO NOTHING;
INSERT INTO courses (code, name) VALUES ('INF-301', 'Base de Datos') ON CONFLICT (code) DO NOTHING;
INSERT INTO courses (code, name) VALUES ('INF-401', 'Ingeniería de Software') ON CONFLICT (code) DO NOTHING;
INSERT INTO courses (code, name) VALUES ('MAT-101', 'Matemáticas Discretas') ON CONFLICT (code) DO NOTHING;

-- Crear ofertas de cursos para el semestre 2025-1
INSERT INTO course_offerings (course_id, semester_id) 
SELECT c.id, s.id 
FROM courses c, semesters s 
WHERE s.year = 2025 AND s.term = 1
ON CONFLICT (course_id, semester_id) DO NOTHING;

-- Crear algunas secciones de ejemplo
INSERT INTO sections (name, course_offering_id) 
SELECT 'Sección A', co.id 
FROM course_offerings co 
JOIN courses c ON co.course_id = c.id 
JOIN semesters s ON co.semester_id = s.id 
WHERE c.code = 'INF-101' AND s.year = 2025 AND s.term = 1;

INSERT INTO sections (name, course_offering_id) 
SELECT 'Sección B', co.id 
FROM course_offerings co 
JOIN courses c ON co.course_id = c.id 
JOIN semesters s ON co.semester_id = s.id 
WHERE c.code = 'INF-101' AND s.year = 2025 AND s.term = 1;

INSERT INTO sections (name, course_offering_id) 
SELECT 'Sección A', co.id 
FROM course_offerings co 
JOIN courses c ON co.course_id = c.id 
JOIN semesters s ON co.semester_id = s.id 
WHERE c.code = 'INF-201' AND s.year = 2025 AND s.term = 1;

-- Insertar estudiantes de ejemplo en las secciones
-- Nota: Los student_id deben corresponder a usuarios con rol 'estudiante' en la base de datos de autenticación
-- Asumiendo que existen usuarios con IDs 1, 2, 3, etc. que tienen rol de estudiante

-- Estudiantes para INF-101 Sección A
INSERT INTO student_sections (student_id, section_id) 
SELECT 1, s.id 
FROM sections s 
JOIN course_offerings co ON s.course_offering_id = co.id 
JOIN courses c ON co.course_id = c.id 
JOIN semesters sem ON co.semester_id = sem.id 
WHERE c.code = 'INF-101' AND s.name = 'Sección A' AND sem.year = 2025 AND sem.term = 1
ON CONFLICT (student_id, section_id) DO NOTHING;

INSERT INTO student_sections (student_id, section_id) 
SELECT 2, s.id 
FROM sections s 
JOIN course_offerings co ON s.course_offering_id = co.id 
JOIN courses c ON co.course_id = c.id 
JOIN semesters sem ON co.semester_id = sem.id 
WHERE c.code = 'INF-101' AND s.name = 'Sección A' AND sem.year = 2025 AND sem.term = 1
ON CONFLICT (student_id, section_id) DO NOTHING;

INSERT INTO student_sections (student_id, section_id) 
SELECT 3, s.id 
FROM sections s 
JOIN course_offerings co ON s.course_offering_id = co.id 
JOIN courses c ON co.course_id = c.id 
JOIN semesters sem ON co.semester_id = sem.id 
WHERE c.code = 'INF-101' AND s.name = 'Sección A' AND sem.year = 2025 AND sem.term = 1
ON CONFLICT (student_id, section_id) DO NOTHING;

-- Estudiantes para INF-101 Sección B
INSERT INTO student_sections (student_id, section_id) 
SELECT 4, s.id 
FROM sections s 
JOIN course_offerings co ON s.course_offering_id = co.id 
JOIN courses c ON co.course_id = c.id 
JOIN semesters sem ON co.semester_id = sem.id 
WHERE c.code = 'INF-101' AND s.name = 'Sección B' AND sem.year = 2025 AND sem.term = 1
ON CONFLICT (student_id, section_id) DO NOTHING;

INSERT INTO student_sections (student_id, section_id) 
SELECT 5, s.id 
FROM sections s 
JOIN course_offerings co ON s.course_offering_id = co.id 
JOIN courses c ON co.course_id = c.id 
JOIN semesters sem ON co.semester_id = sem.id 
WHERE c.code = 'INF-101' AND s.name = 'Sección B' AND sem.year = 2025 AND sem.term = 1
ON CONFLICT (student_id, section_id) DO NOTHING;
