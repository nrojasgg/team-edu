package main

import (
	"ms-core/internal/core"
)

func NewMyPlugin() *MyPlugin {
	return &MyPlugin{
		name:    "my_plugin",
		version: "1.0.0",

		metadata: &core.PluginMetadata{
			NombreInterno:  "my_plugin",
			Nombre:         "Mi Plugin",
			Version:        "1.0.0",
			Autor:          "Tu Nombre o Equipo",
			Descripcion:    "Descripción breve del plugin y su propósito",
			TipoPlugin:     core.PluginTypeTeamFormation,
			TipoRespuestas: core.ResponseTypeLikert,
			Preguntas:      buildMyQuestions(),
			Activo:         true,
			Resultados: []core.ResultDefinition{
				{
					Nombre:      "vector_ejemplo",
					Tipo:        core.ResultTypeStatistical,
					Descripcion: "Vector de ejemplo",
				},
			},
		},
	}
}

func buildMyQuestions() core.Questions {
	q := make(core.Questions)

	q["1"] = "¿Pregunta 1 del cuestionario?"
	q["2"] = "¿Pregunta 2 del cuestionario?"
	q["3"] = "¿Pregunta 3 del cuestionario?"

	return q
}
