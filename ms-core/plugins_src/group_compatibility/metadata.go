package main

import (
	"ms-core/internal/core"
)

func NewGroupCompatibilityPlugin() *GroupCompatibilityPlugin {
	return &GroupCompatibilityPlugin{
		name:    "group_compatibility",
		version: "1.0.0",
		metadata: &core.PluginMetadata{
			NombreInterno:  "group_compatibility",
			Nombre:         "Afinidad Social",
			Version:        "1.0.0",
			Autor:          "Equipo EduTeam",
			Descripcion:    "Encuesta sociométrica de afinidad social entre estudiantes",
			TipoPlugin:     core.PluginTypeTeamFormation,
			TipoRespuestas: core.ResponseTypeMultipleSelection,
			Preguntas:      buildSociometricQuestions(),
			Activo:         true,
			Resultados: []core.ResultDefinition{
				{
					Nombre:      "sociogram",
					Tipo:        core.ResultTypeRelational,
					Descripcion: "Matriz de afinidad social (sociograma)",
				},
			},
		},
	}
}

func buildSociometricQuestions() core.Questions {
	q := make(core.Questions)
	q["1"] = "¿A quién(es) elegirías para trabajar en equipo?"
	q["2"] = "¿Quién(es) crees que te elegirían para trabajar en equipo?"
	q["3"] = "¿A quién(es) NO elegirías para trabajar en equipo?"
	q["4"] = "¿Quién(es) crees que NO te elegirían para trabajar en equipo?"
	return q
}
