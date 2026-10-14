## Proceso de Formación de Equipos en `ms-core`

El sistema de formación de equipos es el componente más complejo de `ms-core`. Su arquitectura sigue un patrón **microkernel + plugin** donde múltiples capas colaboran para agrupar estudiantes en equipos óptimos.

---

### 1. Visión General del Flujo

```
┌─────────────────────────────────────────────────────────────────┐
│  PROFESOR (Frontend)                                            │
│  POST /api/v1/team-formation/form-flexible                      │
│  → envía: section_id, candidates, team_size, criteria_parameters│
└──────────────────────────┬──────────────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  HANDLER (team_formation_handler.go)                            │
│  1. Parsea la request HTTP                                      │
│  2. Enriquece datos OCEAN desde plugins de personalidad         │
│  3. Decide modo: "legacy" (provider_config) o "criteria"        │
│  4. Si hay plugin team_formation → lo ejecuta como subprocesso  │
│     Si no → usa el engine integrado (fallback)                  │
└──────────────────────────┬──────────────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  PLUGIN team_formation (subproceso aislado via go-plugin)       │
│  Dos modos:                                                     │
│  • "affinity": recibe matriz pre-calculada → algoritmo genético │
│  • "criteria": recibe datos demográficos + criterios dinámicos  │
│    → algoritmo genético con fitness multiobjetivo (NSGA-II)     │
└──────────────────────────┬──────────────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  RESULTADO → TeamFormationResult                                │
│  → equipos con miembros, scores, metadata, breakdown por criterio│
│  → guardado en groups_historydata (MongoDB)                     │
└─────────────────────────────────────────────────────────────────┘
```

---

### 2. Inicialización (Bootstrap)

En `bootstrap/team_formation.go`, al arrancar el servidor:

1. **Se crea el `TeamFormationEngine`** con la configuración del archivo `.env` (algoritmo, pesos, parámetros genéticos).
2. **Se crea un `ProviderRegistry`** que gestiona los proveedores de afinidad.
3. **Se registran 3 AffinityProviders** (si sus plugins están cargados):
   - `enneagram` → `EnneagramAffinityProvider` (personalidad, intencionalidad = harmony)
   - `group_compatibility` → `SociogramAffinityProvider` (sociometría, intencionalidad = controlled friction)
   - `mbti` → `MBTIAffinityProvider` (personalidad)
4. **Se filtran providers elegibles** según la configuración (pesos > 0, no excluidos).
5. **Se adjunta el plugin `team_formation`** si existe en `plugins_dist/`. Su ausencia no es fatal: el handler usa el engine integrado como fallback.

---

### 3. Interfaz `AffinityProvider` (el contrato clave)

Cada fuente de datos de compatibilidad implementa:

```go
type AffinityProvider interface {
    ProviderID() string                                    // "enneagram", "mbti", etc.
    ProviderName() string                                  // Nombre legible
    CalculatePairwiseAffinity(ctx, personA, personB) (float64, error)  // [-1.0, 1.0]
    SupportsContext(ctx) bool                              // ¿Tiene datos para esta sección?
    GetMetadata() ProviderMetadata                         // Categoría, intencionalidad (S_p)
}
```

El campo `SocialIntentionality` en el metadata es clave para la fórmula de la tesis:
- `S_p = 1` → **armonía** (compatibilidad positiva deseada)
- `S_p = -1` → **fricción controlada** (diversidad deseada)

---

### 4. Dos Modos de Formación

#### Modo Legacy (`provider_config`)
- El profesor elige qué tests usar y sus pesos (ej: enneagram 0.4, sociogram 0.6).
- El engine calcula una **matriz de afinidad combinada** ponderando los scores de cada provider.
- Ejecuta el algoritmo genético NSGA-II con dos objetivos: **psicométrico** (OCEAN) y **sociométrico** (afinidad).

#### Modo Criteria (`criteria_parameters`) ← **el modo principal**
- El profesor define **criterios dinámicos** con target, weight y goal:
  - Ejemplo: `{"sex": {target: 0.5, weight: 0.4, goal: "target"}, "career_code": {target: 0.7, weight: 0.6, goal: "heterogeneity"}}`
  - También puede incluir traits OCEAN: `openness`, `conscientiousness`, `extraversion`, `agreeableness`, `neuroticism`
- El handler **enriquece automáticamente** los datos OCEAN de los estudiantes consultando los plugins de personalidad (enneagram → mbti → bfi44, en ese orden de prioridad).
- Se mapea a un dominio con `CriterionFactory` que crea un evaluador por cada criterio.
- **Detección automática de multiobjetivo**: si hay criterios psicométricos Y sociales mezclados, se usa **NSGA-II multiobjetivo** (dos funciones de fitness separadas). Si todos son de un solo tipo, se usa algoritmo genético escalar.

---

### 5. El Algoritmo Genético

#### Algoritmo Genético Escalar (modo criteria simple)
Ubicado en `genetic/engine.go`:

1. **Población inicial**: 150 cromosomas (permutaciones aleatorias de estudiantes).
2. **Fitness**: cada cromosoma se evalúa contra los criterios configurados. El fitness es la suma ponderada de scores por criterio. Cada criterio evalúa qué tan cerca está un grupo del target deseado.
3. **Operadores genéticos**:
   - **Selección por torneo** (tournament size = 5)
   - **Order Crossover** (OX) para permutaciones
   - **Mutación por swap** (tasa = 0.2)
   - **Elitismo** (15% de la población)
4. **Parada**: convergencia (stagnation de 40 generaciones), fitness objetivo alcanzado, o máximo 100 generaciones.
5. **Multi-run**: hasta 3 ejecuciones con diferentes semillas, se conserva el mejor resultado.

#### NSGA-II Multiobjetivo (modo criteria mixto)
Ubicado en `genetic/mo_engine.go` y `genetic/nsga2.go`:

1. Separa los criterios en dos categorías: **psicométricos** (OCEAN traits) y **sociales** (demográficos/sociométricos).
2. Ejecuta NSGA-II con **dos funciones de fitness independientes**: `f1 = fitness_psicométrico` y `f2 = fitness_social`.
3. Usa **Fast Non-Dominated Sorting** + **Crowding Distance** para mantener diversidad en el Frente de Pareto.
4. Selecciona la **solución knee-point** (la más cercana al ideal `(1,1)` en el espacio de objetivos).

---

### 6. Fórmula de la Tesis (Costo Relacional)

Para el modo affinity, se implementa la fórmula de costo relacional de la tesis:

```
C'_ij = Σ_p ω_p × (1 - S_p × T_p(i,j)) / 2
```

Donde:
- `T_p(i,j)` ∈ [-1, 1] es la compatibilidad cruda del provider p para el par (i,j)
- `ω_p` es el peso normalizado del provider
- `S_p` ∈ {-1, 1} es la intencionalidad social (armonía o fricción controlada)

La cohesión relacional de un grupo es:
```
S_rel(G) = min(1, Σ_ij C'_ij / (|G|(|G|-1)/2))
```

Y el fitness global: `f_2(x) = (1/m) × Σ_G S_rel(G)`

---

### 7. Estrategia de Equidad (Minimax)

El campo `equity_strategy` permite activar **agregación minimax**:
- En lugar de promediar los scores de todos los equipos, el fitness es el **score del peor equipo**.
- Esto penaliza soluciones donde un equipo es excelente pero otro es terrible.
- Los umbrales de convergencia se reducen (target: 0.75 en vez de 0.98).

---

### 8. Endpoints HTTP Disponibles

| Endpoint | Método | Descripción |
|----------|--------|-------------|
| `/api/v1/team-formation/form` | POST | Formación básica (legacy) |
| `/api/v1/team-formation/form-flexible` | POST | **Formación principal** con criterios dinámicos |
| `/api/v1/team-formation/form-flexible/async` | POST | Versión asíncrona (cola Redis) |
| `/api/v1/team-formation/evaluate-groups` | POST | Evalúa grupos ya existentes |
| `/api/v1/team-formation/providers` | GET | Lista providers disponibles |
| `/api/v1/team-formation/options` | GET | Opciones disponibles para el frontend |
| `/api/v1/team-formation/affinity` | GET | Calcula afinidad entre estudiantes |
| `/api/v1/team-formation/weights` | PUT | Actualiza pesos de providers |
| `/api/v1/jobs/:id` | GET | Consulta resultado de job asíncrono |

---

### 9. Persistencia

Cada formación exitosa se guarda en la colección MongoDB `groups_historydata` como `GroupHistoryRecord`, que incluye: sección, algoritmo usado, providers, pesos, targets, estrategia de equidad, score global, equipos con sus miembros, y el desglose detallado por criterio de cada equipo.