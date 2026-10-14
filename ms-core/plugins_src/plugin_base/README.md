# Plugin Base - Estructura Multiarquivo

Este plugin ha sido refactorizado a una estructura de múltiples archivos para mejor organización y mantenimiento.

## Estructura de Archivos

```
plugins_src/plugin_base/
├── main.go       # Punto de entrada: shared.Serve(NewPlugin())
├── plugin.go     # Struct del plugin + interface implementation
├── metadata.go   # Constructor NewPlugin() + buildQuestions()
├── types.go      # Tipos de datos (Test, Question, Response, Results, etc.)
├── handlers.go   # CRUD de surveys/tests
├── responses.go  # Manejo de respuestas
├── helpers.go    # Funciones auxiliares
└── README.md
```

## Descripción de Archivos

### main.go
Punto de entrada del plugin. Llama a `shared.Serve(NewMyPlugin())`.

### plugin.go
- Struct `MyPlugin` con campos: name, version, collection, metadata
- Interface methods: `Name()`, `Version()`, `Initialize()`, `Shutdown()`, `Execute()`
- `GetMetadata()` para acceder a los metadatos

### metadata.go
- `NewMyPlugin()` - Constructor con todos los `PluginMetadata`
- `buildMyQuestions()` - Mapa de preguntas del plugin

### types.go
- `MyTest` - Test/encuesta
- `MyQuestion` - Pregunta del cuestionario
- `MyResponse` - Respuesta de estudiante
- `MyAnswer` - Respuesta a pregunta específica
- `MyResults` - Resultados calculados (OCEAN)

### handlers.go
CRUD de surveys/tests:
- `createTest`
- `getTests`
- `getTest`
- `updateTest`
- `deleteTest`

### responses.go
Manejo de respuestas:
- `submitResponse`
- `getResponses`
- `getStudentResponse`
- `getSectionResponses`

### helpers.go
Funciones auxiliares:
- `parseOptionalDate` - Parsea fechas opcionales
- `parseObjectID` - Parsea strings a ObjectID
- `currentTime` - Retorna tiempo actual UTC
- `CalculateLikertDimension` - Calcula promedio de dimensión Likert
- `NormalizeToOcean` - Normaliza puntuaciones a escala 0-1

## Compilación

```bash
go build -o ../../plugins_dist/plugin_base .
```

## Original Template Structure

```
plugins_src/
├── shared/
│   └── bootstrap.go        # Función Serve() que inicia el plugin
├── plugin_base/
│   ├── main.go            # Plantilla base del plugin
│   └── README.md          # Este archivo
├── mbti/
├── bfi44/
├── enneagram/
├── group_compatibility/
└── peer_evaluation/
```
