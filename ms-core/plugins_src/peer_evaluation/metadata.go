package main

import (
	"ms-core/internal/core"
)

func NewPeerEvaluationPlugin() *PeerEvaluationPlugin {
	return &PeerEvaluationPlugin{
		name:    "peer_evaluation",
		version: "1.0.0",
		metadata: &core.PluginMetadata{
			NombreInterno:  "peer_evaluation",
			Nombre:         "Evaluación entre Pares",
			Version:        "1.0.0",
			Autor:          "Equipo EduTeam",
			Descripcion:    "Sistema de evaluación entre compañeros de grupo",
			TipoPlugin:     core.PluginTypePeerEval,
			TipoRespuestas: core.ResponseTypeLikert,
			Preguntas:      buildPeerEvaluationQuestions(),
			Activo:         true,
			Resultados: []core.ResultDefinition{
				{
					Nombre:      "evaluations",
					Tipo:        core.ResultTypeStatistical,
					Descripcion: "Puntuaciones promediadas de evaluación entre pares",
				},
			},
		},
	}
}

func buildPeerEvaluationQuestions() core.Questions {
	q := make(core.Questions)
	q["1"] = "Realiza las tareas que le son asignadas dentro del grupo en los plazos requeridos."
	q["2"] = "Participa de forma activa en los espacios de encuentro del equipo, compartiendo la información, conocimientos y las experiencias."
	q["3"] = "Colabora en la definición, organización y distribución de las tareas del equipo."
	q["4"] = "Se orienta a la consecución de acuerdos y objetivos comunes y se compromete con ellos."
	q["5"] = "Toma en cuenta los puntos de vista de los demás y retroalimenta constructivamente."
	return q
}
