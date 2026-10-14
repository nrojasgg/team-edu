# Peer Evaluation Plugin

Sistema de evaluación entre compañeros de grupo para la plataforma EduTeam.

## Estructura

```
peer_evaluation/
├── main.go        # Punto de entrada
├── plugin.go      # Struct + interface implementation
├── metadata.go    # NewPeerEvaluationPlugin() + buildQuestions()
├── types.go       # Tipos de datos
├── handlers.go    # CRUD de evaluaciones
├── responses.go   # Manejo de respuestas
└── README.md
```

## Acciones disponibles

### Surveys
- `create_survey` - Crear nueva encuesta de evaluación
- `get_surveys` - Obtener todas las encuestas de una sección
- `get_survey` - Obtener una encuesta específica
- `update_survey` - Actualizar una encuesta
- `delete_survey` - Eliminar una encuesta
- `get_survey_peers` - Obtener compañeros de grupo

### Respuestas
- `submit_response` - Enviar una respuesta de evaluación
- `get_responses` - Obtener respuestas de una encuesta
- `get_results` - Obtener resultados agregados
- `get_section_responses` - Obtener todas las respuestas de una sección
- `get_section_responses_summary` - Resumen de respuestas por estudiante
- `get_section_group_responses` - Respuestas de un grupo específico
- `export_to_excel` - Exportar resultados a Excel

## Compilación

```bash
go build -o ../../plugins_dist/peer_evaluation .
```
