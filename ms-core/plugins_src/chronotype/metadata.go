package main

import (
	"ms-core/internal/core"
)

func NewChronotypePlugin() *ChronotypePlugin {
	return &ChronotypePlugin{
		name:    "chronotype",
		version: "1.0.0",

		metadata: &core.PluginMetadata{
			NombreInterno:  "chronotype",
			Nombre:         "Cuestionario de Matutinidad-Vespertinidad (MEQ)",
			Version:        "1.0.0",
			Autor:          "TeamEdu",
			Descripcion:    "Test MEQ-SA para determinar el cronotipo de matutinidad-vespertinidad",
			TipoPlugin:     core.PluginTypeTeamFormation,
			TipoRespuestas: core.ResponseTypeLikert,
			Preguntas:      buildChronotypeQuestions(),
			Activo:         true,
			Resultados: []core.ResultDefinition{
				{
					Nombre:      "total_score",
					Tipo:        core.ResultTypeStatistical,
					Descripcion: "Puntuacion total del test (16-86)",
				},
				{
					Nombre:      "category",
					Tipo:        core.ResultTypeCategorical,
					Descripcion: "Categoria de cronotipo",
				},
			},
		},
	}
}

func buildChronotypeQuestions() core.Questions {
	q := make(core.Questions)

	q["1"] = "Si solo pensaras en cuando te sentirias mejor y fueras totalmente libre de planificarte el dia. A que hora te levantarias?"
	q["2"] = "Si solo pensaras en cuando te sentirias mejor y fueras totalmente libre de planificarte el dia. A que hora te acostarias?"
	q["3"] = "Para levantarte por la manana a una hora especifica. Hasta que punto necesitas que te avise el despertador?"
	q["4"] = "Te resulta facil levantarte por las mananas?"
	q["5"] = "Una vez levantado por las mananas. Que tal te encuentras durante la primera media hora?"
	q["6"] = "Una vez levantado por las mananas. Como es tu apetito durante la primera media hora?"
	q["7"] = "Una vez levantado por las mananas. Que tal te sientes durante la primera media hora?"
	q["8"] = "Cuando no tienes compromisos al dia siguiente. A que hora te acuestas en relacion con tu hora habitual?"
	q["9"] = "Has decidido hacer un poco de ejercicio fisico. Un amigo te propone hacerlo una hora dos veces por semana y segun el, la mejor hora seria de 7 a 8 de la manana. Como crees que te encontrarias?"
	q["10"] = "A que hora aproximada de la noche te sientes cansado y como consecuencia necesitas dormir?"
	q["11"] = "Quieres estar en tu punto maximo de rendimiento para una prueba de dos horas mentalmente agotadora. Que horario elegirias?"
	q["12"] = "Si te acostaras a las 11 de la noche. Que nivel de cansancio noterias?"
	q["13"] = "Por algun motivo te has acostado varias horas mas tarde de lo habitual. Cuando crees que te despertarias?"
	q["14"] = "Una noche tienes que permanecer despierto de 4 a 6 de la madrugada. Sin tener ningun compromiso al dia siguiente, que preferiras?"
	q["15"] = "Tienes que hacer dos horas de trabajo fisico pesado. Pensando solo en cuando te sentirias mejor, que horario escogeras?"
	q["16"] = "Has decidido hacer ejercicio fisico intenso. Un amigo te sugiere hacerlo de 10 a 11 de la noche. Como crees que te sentaria?"
	q["17"] = "Imaginate que puedes escoger tu horario de trabajo. Que CINCO HORAS CONSECUTIVAS seleccionarias?"
	q["18"] = "A que hora del dia crees que alcanzas tu maximo bienestar?"
	q["19"] = "Se habla de personas de tipo matutino y vespertino. Cual de estos tipos te consideras ser?"

	return q
}
