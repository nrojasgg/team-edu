# Enneagram Plugin

Plugin de subprocess para el test de Eneagrama dentro del sistema ms-core.

## Estructura

```
enneagram/
├── main.go              # Punto de entrada
├── plugin.go            # EnneagramPlugin struct + interface
├── metadata.go          # NewEnneagramPlugin() + buildEnneagramQuestions()
├── types.go             # Tipos de datos
├── handlers.go          # CRUD de tests
├── responses.go         # Manejo de respuestas
├── enneagram_ocean.go   # Conversión Enneagram->OCEAN
└── README.md
```

## Compilación

```bash
go build -o ../../plugins_dist/enneagram .
```

## Acciones disponibles

- `create_test` - Crear un nuevo test
- `get_tests` - Listar tests
- `get_test` - Obtener un test específico
- `update_test` - Actualizar un test
- `delete_test` - Eliminar un test
- `submit_response` - Enviar respuestas de un estudiante
- `get_responses` - Obtener respuestas de un test
- `get_results` - Obtener resultados de un estudiante
- `get_section_responses` - Obtener todas las respuestas de una sección
- `get_section_results` - Obtener resultados de una sección
- `get_student_eneatypes` - Obtener tipos eneatipos de estudiantes
- `calculate_group_compatibility` - Calcular compatibilidad grupal
- `get_pairwise_compatibility` - Calcular compatibilidad entre dos tipos
- `get_ocean_from_type` - Convertir tipo Enneagram a vector OCEAN
- `get_metadata` - Obtener metadatos del plugin

## Colecciones MongoDB

- `enneagram_tests` - Tests creados
- `enneagram_responses` - Respuestas de estudiantes
