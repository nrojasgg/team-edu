package core

// Tipo de plugin según su uso en el sistema
type PluginType string

const (
	PluginTypeTeamFormation PluginType = "conformacion_equipos"
	PluginTypePeerEval      PluginType = "evaluacion_pares"
)

// Tipo de respuesta que acepta el plugin
type ResponseType string

const (
	ResponseTypeDichotomous       ResponseType = "dicotomico"
	ResponseTypeLikert            ResponseType = "likert"
	ResponseTypeMultipleSelection ResponseType = "seleccion_multiple"
)

// Tipo de resultado que entrega el plugin (cómo el core lo procesa)
type ResultType string

const (
	ResultTypeStatistical ResultType = "estadistico"
	ResultTypeRelational  ResultType = "relacional"
	ResultTypeCategorical ResultType = "categorico"
)

// ResultDefinition define qué resultados entrega un plugin y cómo el core los interpreta
type ResultDefinition struct {
	Nombre      string     `json:"nombre"`
	Tipo        ResultType `json:"tipo"`
	Descripcion string     `json:"descripcion"`
}

// Questions define las preguntas del plugin como un map de ID -> texto
type Questions map[string]string

// QuestionOption represents a single option in a question with score
type QuestionOption struct {
	Score int    `bson:"score" json:"score"`
	Text  string `bson:"text" json:"text"`
}

// FullQuestion represents a complete question with all its options
type FullQuestion struct {
	ID       int              `bson:"id" json:"id"`
	Text     string           `bson:"text" json:"text"`
	Options  []QuestionOption `bson:"options" json:"options"`
	Required bool             `bson:"required" json:"required"`
}

// PluginMetadataExtended extends PluginMetadata with full question details
type PluginMetadataExtended struct {
	PluginMetadata
	FullQuestions []FullQuestion `bson:"full_questions" json:"full_questions"`
}

// PluginMetadata contiene la información unificada de un plugin
// que el frontend usa para mostrar selector de criterios y el core
// para entender cómo procesar los resultados.
type PluginMetadata struct {
	NombreInterno  string             `json:"nombre_interno"`
	Nombre         string             `json:"nombre"`
	Version        string             `json:"version"`
	Autor          string             `json:"autor"`
	Descripcion    string             `json:"descripcion"`
	TipoPlugin     PluginType         `json:"tipo_plugin"`
	TipoRespuestas ResponseType       `json:"tipo_respuestas"`
	Preguntas      Questions          `json:"preguntas"`
	Activo         bool               `json:"activo"`
	Resultados     []ResultDefinition `json:"resultados"`
}
