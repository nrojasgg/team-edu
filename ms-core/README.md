# MS-Core - Microservicio Central

Microservicio Go que implementa la lógica central del sistema TeamEdu, incluyendo encuestas, evaluaciones de pares, tests de personalidad y gestión de compatibilidad grupal a través de una arquitectura de plugins.

## Características

- **Arquitectura de Plugins**: Sistema modular que permite cargar diferentes tipos de encuestas y evaluaciones
- **Evaluaciones de Pares**: Implementación completa con snapshots de grupos basada en encuesta_pares.txt con 5 categorías
- **Test de Eneagrama**: Plugin para test de personalidad con mapeo de tipos de eneagrama
- **Compatibilidad Grupal**: Plugin para mapear preferencias de trabajo en equipo (sí/no a compañeros)
- **Gestión de Grupos**: Formación automática de grupos y manejo de compatibilidad
- **MongoDB**: Base de datos NoSQL para almacenamiento flexible de datos de encuestas
- **Go Fiber 2.52.0**: Framework web rápido y moderno
- **JWT Authentication**: Autenticación segura con tokens JWT

## Estructura del Proyecto

```
ms-core/
├── main.go                          # Punto de entrada de la aplicación
├── go.mod                           # Dependencias de Go
├── go.sum                           # Checksums de dependencias
├── .env.example                     # Variables de entorno de ejemplo
├── README.md                        # Este archivo
├── GROUP_COMPATIBILITY_API.md       # Documentación del plugin de compatibilidad grupal
├── PEER_EVALUATION_PLUGIN.md        # Documentación del plugin de evaluación de pares
├── ENNEAGRAM_PLUGIN.md              # Documentación del plugin de eneagrama
└── internal/
    ├── config/
    │   └── config.go               # Configuración de la aplicación
    ├── database/
    │   └── connection.go           # Conexión a MongoDB
    ├── core/
    │   └── plugin_manager.go       # Gestor de plugins
    ├── handlers/
    │   ├── peer_evaluation_handler.go # Handlers para evaluaciones de pares
    │   ├── enneagram_handler.go    # Handlers para test de eneagrama
    │   ├── group_compatibility_handler.go # Handlers para compatibilidad grupal
    │   └── group_handler.go        # Handlers para gestión de grupos
    ├── middleware/
    │   └── auth.go                 # Middleware de autenticación JWT
    └── routes/
        └── routes.go               # Configuración de rutas

plugins/
├── peer_evaluation.go              # Plugin de evaluación entre pares
├── enneagram.go                    # Plugin de test de eneagrama
└── group_compatibility.go          # Plugin de compatibilidad grupal
```

## Instalación y Configuración

### Prerequisitos

- Go 1.21 o superior
- MongoDB 4.4 o superior

### Configuración

1. Clonar el repositorio y navegar al directorio ms-core
2. Copiar el archivo de configuración:
   ```powershell
   cp .env.example .env
   ```
3. Editar el archivo `.env` con tus configuraciones específicas
4. Instalar dependencias:
   ```bash
   go mod tidy
   ```

### Ejecutar el servicio

```bash
go run main.go
```

El servicio estará disponible en `http://localhost:3003`

## API Endpoints

### Health Check
- `GET /health` - Verificar estado del servicio y plugins cargados

### Evaluación de Pares
- `POST /api/v1/peer-evaluation/surveys` - Crear survey de evaluación de pares
- `GET /api/v1/peer-evaluation/surveys` - Obtener surveys de evaluación
- `GET /api/v1/peer-evaluation/surveys/:id` - Obtener survey específico
- `PUT /api/v1/peer-evaluation/surveys/:id` - Actualizar survey
- `DELETE /api/v1/peer-evaluation/surveys/:id` - Eliminar survey
- `POST /api/v1/peer-evaluation/survey-peers` - Obtener pares a evaluar del survey
- `POST /api/v1/peer-evaluation/surveys/:survey_id/responses` - Enviar respuesta
- `GET /api/v1/peer-evaluation/surveys/:survey_id/responses` - Obtener respuestas
- `GET /api/v1/peer-evaluation/surveys/:survey_id/results` - Obtener resultados

### Secciones (Responses)
- `GET /api/v1/sections/:section_id/responses/summary` - Resumen de respuestas
- `GET /api/v1/sections/:section_id/responses/:group_id` - Respuestas por grupo
- `GET /api/v1/sections/:section_id/responses` - Todas las respuestas

### Eneagrama
- `POST /api/v1/enneagram/questionnaires` - Crear test
- `GET /api/v1/enneagram/questionnaires` - Obtener tests
- `POST /api/v1/enneagram/questionnaires/:test_id/responses` - Enviar respuesta
- `GET /api/v1/enneagram/questionnaires/:test_id/results` - Obtener resultados

### Compatibilidad Grupal
- `POST /api/v1/group-compatibility/surveys` - Crear survey de compatibilidad
- `GET /api/v1/group-compatibility/surveys/section/:sectionId` - Obtener surveys
- `GET /api/v1/group-compatibility/surveys/:id` - Obtener survey específico
- `PUT /api/v1/group-compatibility/surveys/:id` - Actualizar survey
- `DELETE /api/v1/group-compatibility/surveys/:id` - Eliminar survey
- `POST /api/v1/group-compatibility/surveys/:surveyId/responses` - Enviar respuesta
- `GET /api/v1/group-compatibility/surveys/:surveyId/responses` - Obtener respuestas
- `GET /api/v1/group-compatibility/surveys/:surveyId/responses/:studentId` - Respuesta de estudiante

### Grupos
- `GET /api/v1/groups/section/:sectionId` - Obtener grupos de una sección
- `POST /api/v1/groups` - Crear o actualizar grupo
- `DELETE /api/v1/groups/section/:sectionId/student/:studentId` - Remover estudiante de grupo

### Plugins
- `GET /api/v1/plugins` - Listar plugins cargados

## Arquitectura de Plugins

### Plugin Interface

Todos los plugins deben implementar la interfaz `Plugin`:

```go
type Plugin interface {
    Name() string
    Version() string
    Initialize() error
    Execute(ctx context.Context, data interface{}) (interface{}, error)
    Shutdown() error
}
```

### Plugins Disponibles

#### 1. Plugin de Evaluación de Pares

El plugin `peer_evaluation` implementa evaluaciones de pares entre miembros de grupos con snapshots de grupos.

**Características:**
- Snapshots de Grupos: Captura el estado de los grupos en el momento de crear la survey
- Validación: Valida que evaluador y evaluado pertenezcan al mismo grupo
- 5 Categorías basadas en encuesta_pares.txt con escala 1-4

**Categorías de Evaluación:**
1. **Tareas** (25%): Realización de tareas asignadas
2. **Participación** (25%): Participación activa en reuniones
3. **Organización** (25%): Colaboración en organización del trabajo
4. **Compromiso** (25%): Orientación a objetivos comunes
5. **Prosocialidad** (25%): Consideración de puntos de vista ajenos

Para más detalles, ver [PEER_EVALUATION_PLUGIN.md](./PEER_EVALUATION_PLUGIN.md)

#### 2. Plugin de Eneagrama

El plugin `enneagram` proporciona tests de personalidad con 9 tipos de eneagrama para mapeo de compatibilidad de grupos.

**Características:**
- Test de personalidad con 36 preguntas
- Mapeo automático a uno de 9 tipos de eneagrama
- Cálculo de compatibilidad grupal basado en tipos
- Análisis de resultados por sección

Para más detalles, ver [ENNEAGRAM_PLUGIN.md](./ENNEAGRAM_PLUGIN.md)

#### 3. Plugin de Compatibilidad Grupal

El plugin `group_compatibility` permite a estudiantes indicar sus preferencias sobre con quién trabajar y con quién no.

**Características:**
- 4 preguntas sobre preferencias de trabajo en equipo
- Mapeo sí/no para cada compañero
- Dos modos: análisis a todo el curso o dentro de grupos formados
- Estructura flexible para análisis de afinidad

**Preguntas:**
1. ¿A quién(es) elegirías para trabajar en equipo?
2. ¿Quién(es) crees que te elegirían para trabajar en equipo?
3. ¿A quién(es) NO elegirías para trabajar en equipo?
4. ¿Quién(es) crees que NO te elegirían para trabajar en equipo?

Para más detalles, ver [GROUP_COMPATIBILITY_API.md](./GROUP_COMPATIBILITY_API.md)

## Modelos de Datos

### Colecciones MongoDB

#### peer_evaluation_surveys
Almacena las surveys de evaluación entre pares con snapshots de grupos

#### peer_evaluation_responses
Almacena las respuestas de estudiantes a surveys de evaluación

#### enneagram_tests
Almacena los tests de eneagrama disponibles

#### enneagram_responses
Almacena las respuestas y resultados de tests de eneagrama

#### group_compatibility_surveys
Almacena las surveys de compatibilidad grupal

#### group_compatibility_responses
Almacena las respuestas de estudiantes sobre preferencias de compañeros

#### groups
Almacena información de grupos de estudiantes

## Compilación y Despliegue

### Compilar

```bash
go build -o bin/ms-core main.go
```

### Ejecutar

```bash
./bin/ms-core
```

### Docker

```bash
docker build -t ms-core .
docker run -p 3003:3003 -e MONGODB_URI=mongodb://localhost:27017 ms-core
```

## Variables de Entorno

```
MONGODB_URI=mongodb://localhost:27017
DB_NAME=TeamEdu
JWT_SECRET=tu_secret_muy_seguro
PORT=3003
NODE_ENV=development
```

## Troubleshooting

### Plugin no se carga
- Verificar que el plugin implementa la interfaz correctamente
- Revisar logs de inicialización del plugin
- Verificar conexión a MongoDB

### Error de conexión a MongoDB
- Verificar que MongoDB está en ejecución
- Verificar la URI de conexión en `.env`
- Verificar permisos de acceso a la base de datos

## Licencia

MIT

