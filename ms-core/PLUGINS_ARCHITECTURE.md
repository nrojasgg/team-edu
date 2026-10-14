# Arquitectura de Plugins — ms-core

## Sistema de Plugins

`ms-core` utiliza una **arquitectura microkernel** donde el núcleo (kernel) es el servidor HTTP con Fiber, y cada plugin corre como un **subproceso independiente** comunicado mediante `net/rpc` a través de `hashicorp/go-plugin`.

### Interfaz del Plugin

Todos los plugins deben implementar la interfaz definida en `internal/core/plugin_manager.go`:

```go
type Plugin interface {
    Name() string
    Version() string
    Initialize() error
    Execute(ctx context.Context, data interface{}) (interface{}, error)
    Shutdown() error
}
```

| Método | Propósito |
|--------|-----------|
| `Name()` | Identificador único del plugin |
| `Version()` | Versión actual del plugin |
| `Initialize()` | Configuración inicial (índices MongoDB, validaciones) |
| `Execute()` | Punto de entrada para todas las operaciones del plugin |
| `Shutdown()` | Limpieza de recursos al apagar el servidor |

### Flujo de Carga

```
bootstrap.LoadPlugins()
    │
    ▼
Escanea directorio plugins_dist/ (configurable vía PLUGIN_DIR)
    │
    ▼
Por cada binario → PluginManager.LoadPlugin(binaryPath)
    │
    ├─ Lanza subproceso (exec.Command)
    ├─ Handshake via hashicorp/go-plugin
    ├─ Dispensa el plugin vía net/rpc
    ├─ Llama Initialize()
    └─ Almacena con enabled=true
```

### PluginManager

Ubicado en `internal/core/plugin_manager.go`, gestiona el ciclo de vida completo:

| Método | Descripción |
|--------|-------------|
| `RegisterPlugin(plugin)` | Registra y activa un plugin (llama `Initialize()` fuera del lock) |
| `LoadPlugin(binaryPath)` | Lanza un binario como subproceso y lo registra |
| `EnablePlugin(name)` | Reactiva un plugin deshabilitado |
| `DisablePlugin(name)` | Desactiva un plugin sin descargarlo de memoria |
| `IsEnabled(name)` | Consulta el estado activo/inactivo |
| `GetPlugin(name)` | Obtiene un plugin por nombre (retorna error si deshabilitado) |
| `ExecutePlugin(name, data)` | Ejecuta un plugin con timeout de 30 segundos |
| `GetLoadedPlugins()` | Lista todos los plugins con su estado |
| `Shutdown()` | Cierra todos los plugins y termina sus subprocesos |

**Seguridad de concurrencia:** Usa `sync.RWMutex` — `RLock` para lecturas y `Lock` para escrituras.

### Código Compartido

`plugins_src/shared/bootstrap.go` proporciona la función `Serve()` que todos los plugins usan para:
1. Conectarse a MongoDB (usando `MONGO_URI` del entorno heredado del proceso padre)
2. Registrarse vía `core.ServePlugin(plugin)` para exponer el plugin por RPC

---

## Plugins Compilados en Docker

El Dockerfile compila **7 plugins** como binarios independientes:

| # | Plugin | Directorio | Colecciones MongoDB |
|---|--------|------------|---------------------|
| 1 | `enneagram` | `plugins_src/enneagram/` | `enneagram_responses` |
| 2 | `mbti` | `plugins_src/mbti/` | `mbti_tests`, `mbti_responses` |
| 3 | `bfi44` | `plugins_src/bfi44/` | (psicometría) |
| 4 | `peer_evaluation` | `plugins_src/peer_evaluation/` | `peer_evaluation_surveys`, `peer_evaluation_responses` |
| 5 | `sociogram` | `plugins_src/sociogram/` | `group_compatibility_surveys`, `group_compatibility_responses` |
| 6 | `team_formation` | `plugins_src/team_formation/` | (motor de formación) |
| 7 | `group_compatibility` | `plugins_src/group_compatibility/` | (preferencias de grupo) |

---

### 1. Enneagram (`enneagram/`)

**Archivos:** `main.go`, `plugin.go`, `metadata.go`, `types.go`, `handlers.go`, `responses.go`, `enneagram_ocean.go`, `README.md`

**Descripción:** Test de personalidad basado en los **9 tipos de eneagrama**:
1. Reformador
2. Ayudador
3. Triunfador
4. Individualista
5. Investigador
6. Leal
7. Entusiasta
8. Desafiador
9. Pacificador

**Datos:** 45 preguntas con escala Likert.

**Funcionalidades:**
- CRUD de tests (crear, listar, obtener, actualizar, eliminar)
- Envío de respuestas y cálculo automático del tipo de eneagrama
- Mapeo OCEAN (Big Five) por tipo de eneagrama (`enneagram_ocean.go`)
- Matriz de compatibilidad 9×9 entre tipos
- Cálculo de compatibilidad grupal
- Resultados por sección y por estudiante

---

### 2. MBTI (`mbti/`)

**Archivos:** `main.go`, `plugin.go`, `metadata.go`, `types.go`, `handlers.go`, `responses.go`, `ocean.go`, `mbti_test.go`

**Descripción:** Test **Myers-Briggs Type Indicator** con 4 dimensiones dicotómicas:
- **E**xtraversión / **I**ntroversión
- **S**ensación / i**N**tuición
- **T**hinking (Pensamiento) / **F**eeling (Sentimiento)
- **J**uicio / **P**ercepción

**Datos:** 40 preguntas en español, responseType `dichotomous`.

**Funcionalidades:**
- CRUD de tests
- Cálculo automático del tipo MBTI (16 tipos posibles) con score por dimensión (PCI)
- Vector OCEAN derivado del tipo MBTI (`ocean.go`)
- Compatibilidad entre tipos MBTI
- Respuestas por sección y por estudiante
- Tests unitarios (`mbti_test.go`)

**16 tipos resultantes:** ISTJ, ISFJ, INFJ, INTJ, ISTP, ISFP, INFP, INTP, ESTP, ESFP, ENFP, ENTP, ESTJ, ESFJ, ENFJ, ENTJ

---

### 3. BFI-44 (`bfi44/`)

**Archivos:** `main.go`, `plugin.go`, `metadata.go`, `types.go`, `handlers.go`, `responses.go`, `calculator.go`, `bfi44_test.go`, `example_bfi44_main.go`, `README.md`

**Descripción:** Test **Big Five Inventory** de 44 ítems — el modelo de personalidad de 5 factores ampliamente usado en psicometría:

| Factor | Sigla | Descripción |
|--------|-------|-------------|
| Apertura a la experiencia | O | Curiosidad intelectual, creatividad |
| Consciencia | C | Organización, responsabilidad, disciplina |
| Extraversión | E | Sociabilidad, energía, asertividad |
| Amabilidad | A | Cooperación, empatía, confianza |
| Neuroticismo | N | Inestabilidad emocional, ansiedad |

**Datos:** 44 preguntas con escala Likert de 5 puntos.

**Funcionalidades:**
- CRUD de tests psicométricos
- Cálculo de scores por factor OCEAN (`calculator.go`)
- Respuestas por sección y resultados por estudiante
- Tests unitarios (`bfi44_test.go`)
- Ejemplo de uso (`example_bfi44_main.go`)

---

### 4. Peer Evaluation (`peer_evaluation/`)

**Archivos:** `main.go`, `plugin.go`, `metadata.go`, `types.go`, `handlers.go`, `responses.go`, `README.md`

**Descripción:** Evaluación entre pares — permite a estudiantes evaluar a sus compañeros de equipo.

**5 Categorías de evaluación** (peso igualitario 25% cada una):

| # | Categoría | Qué mide |
|---|-----------|----------|
| 1 | Tareas | Realización de tareas asignadas |
| 2 | Participación | Participación activa en reuniones |
| 3 | Organización | Colaboración en organización del trabajo |
| 4 | Compromiso | Orientación a objetivos comunes |
| 5 | Prosocialidad | Consideración de puntos de vista ajenos |

**Escala:** 1-4

**Características:**
- Snapshots de grupos al momento de crear la survey (captura el estado del grupo en ese momento)
- Validación de que evaluador y evaluado pertenezcan al mismo grupo
- Resultados agregados por grupo y por sección

---

### 5. Sociogram (`sociogram/`)

**Archivo:** `main.go` (~602 líneas — implementación monolítica)

**Descripción:** Encuesta sociométrica de compatibilidad grupal. Mapea preferencias de trabajo en equipo entre estudiantes.

**Preguntas clave:**
1. ¿A quién(es) elegirías para trabajar en equipo?
2. ¿Quién(es) crees que te elegirían para trabajar en equipo?
3. ¿A quién(es) NO elegirías para trabajar en equipo?
4. ¿Quién(es) crees que NO te elegirían para trabajar en equipo?

**Escala de respuesta:** `-1` (rechazo), `0` (indiferente), `1` (elección positiva)

**Tipos de datos:**
- `CompatibilityQuestion` — pregunta con ID, texto, categoría (`positive_choice`/`negative_choice`)
- `StudentAnswer` — student_id + response
- `QuestionResponse` — mapea question_id a lista de StudentAnswer
- `GroupCompatibilityStudentResponse` — respuesta completa de un estudiante

---

### 6. Team Formation (`team_formation/`)

**Archivo:** `main.go` (~736 líneas — implementación monolítica)

**Descripción:** Motor de **formación automática de equipos**. Es el plugin más complejo del sistema — orquesta la lógica para agrupar estudiantes en equipos basándose en los datos recopilados por los demás plugins.

**Integración especial:** En `main.go` del servidor, `bootstrap.InitializeTeamFormation()` inicializa un `TeamFormationEngine` con un `Registry` y este plugin. Se exponen rutas dedicadas via `routes.SetupTeamFormationRoutes()`. Además, se configura un worker de cola de tareas (asynq) para procesamiento asíncrono.

**Datos de entrada que consume:**
- Resultados de tests de personalidad (BFI-44, Eneagrama, MBTI)
- Evaluaciones de pares
- Preferencias de compatibilidad/sociograma

---

### 7. Group Compatibility (`group_compatibility/`)

**Archivos:** `main.go`, `plugin.go`, `metadata.go`, `types.go`, `handlers.go`, `responses.go`, `README.md`

**Descripción:** Plugin para que los estudiantes mapeen sus preferencias de trabajo en equipo. Similar al sociograma pero con un enfoque complementario.

**4 preguntas sobre preferencias:**
1. ¿A quién(es) elegirías para trabajar en equipo?
2. ¿Quién(es) crees que te elegirían para trabajar en equipo?
3. ¿A quién(es) NO elegirías para trabajar en equipo?
4. ¿Quién(es) crees que NO te elegirían para trabajar en equipo?

**Dos modos de análisis:**
- **A todo el curso** — análisis de afinidad entre todos los estudiantes de una sección
- **Dentro de grupos formados** — análisis de compatibilidad intra-grupo

---

## Plugins en Desarrollo (no compilados en Docker)

### Chronotype (`chronotype/`)

**Archivos:** `main.go`, `plugin.go`, `metadata.go`, `types.go`, `handlers.go`, `responses.go`, `calculator.go`

**Descripción:** Plugin para clasificar el **cronotipo** (tipo de ritmo circadiano) de los estudiantes. No está incluido en el Dockerfile actual, por lo que está en fase de desarrollo.

---

### Plugin Base (`plugin_base/`)

**Archivos:** `main.go`, `plugin.go`, `metadata.go`, `types.go`, `handlers.go`, `responses.go`, `helpers.go`, `README.md`

**Descripción:** Plantilla/base para crear nuevos plugins. Contiene la estructura genérica con boilerplate de handlers, responses, types y metadata que puede usarse como punto de partida para desarrollar plugins adicionales.

---

## Flujo Funcional del Sistema

```
┌─────────────────────────────────────────────────────────┐
│              TESTS PSICOMÉTRICOS                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │  BFI-44  │  │ Eneagrama│  │   MBTI   │              │
│  │ (OCEAN)  │  │ (9 tipos)│  │ (16 tipos│              │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘              │
│       └──────────────┼─────────────┘                    │
│                      ▼                                   │
│         Perfil de personalidad                           │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│              EVALUACIONES SOCIALES                       │
│  ┌────────────────┐  ┌───────────┐  ┌────────────────┐ │
│  │    Peer Eval   │  │ Sociogram │  │    Group       │ │
│  │ (5 categorías) │  │ (sí/no)   │  │ Compatibility  │ │
│  └───────┬────────┘  └─────┬─────┘  └───────┬────────┘ │
│          └─────────────────┼─────────────────┘          │
│                            ▼                             │
│              Dinámicas sociales del grupo                │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│           TEAM FORMATION ENGINE                          │
│                                                          │
│  Combina personalidad + dinámicas sociales               │
│  para formar equipos óptimos                             │
│                                                          │
│  ┌──────────────────────────────────────────────┐       │
│  │  Grupos de trabajo equilibrados               │       │
│  └──────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────┘
```

## Compilación

### Todos los plugins (Docker — recomendado)
```bash
make plugins-docker
```

### Plugin individual (Docker)
```bash
make plugin-docker-enneagram
```

### Todos los plugins (Linux local)
```bash
make plugins
```

### Todos los plugins (Windows — desarrollo)
```bash
make plugins-windows
```

### Servidor + plugins en un paso
```bash
make build-all