# Group Compatibility Plugin

Plugin de encuestas sociométricas para ms-core.

## Estructura

- `main.go` - Punto de entrada
- `plugin.go` - Struct e implementación de interface
- `metadata.go` - Constructor y metadatos
- `types.go` - Tipos de datos
- `handlers.go` - CRUD de encuestas
- `responses.go` - Manejo de respuestas

## Acciones

### Surveys
- `create_survey` - Crear encuesta
- `get_surveys` - Listar encuestas por sección
- `get_survey` - Obtener una encuesta
- `update_survey` - Actualizar encuesta
- `delete_survey` - Eliminar encuesta

### Responses
- `submit_response` - Submitir respuesta
- `get_responses` - Obtener respuestas de encuesta
- `get_student_response` - Obtener respuesta de estudiante
- `get_section_responses` - Obtener todas las respuestas de sección

## Compilar

```bash
go build -o ../../plugins_dist/group_compatibility .
```
