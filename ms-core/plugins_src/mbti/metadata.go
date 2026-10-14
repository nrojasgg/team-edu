package main

import (
	"ms-core/internal/core"
)

func NewMBTIPlugin() *MBTIPlugin {
	return &MBTIPlugin{
		name:    "mbti",
		version: "1.0.0",
		metadata: &core.PluginMetadata{
			NombreInterno:  "mbti",
			Nombre:         "MBTI",
			Version:        "1.0.0",
			Autor:          "Equipo EduTeam",
			Descripcion:    "Test de personalidad Myers-Briggs Type Indicator basado en Francis Psychological Type Scales",
			TipoPlugin:     core.PluginTypeTeamFormation,
			TipoRespuestas: core.ResponseTypeDichotomous,
			Preguntas:      buildMBTIQuestions(),
			Activo:         true,
			Resultados: []core.ResultDefinition{
				{
					Nombre:      "mbti_type",
					Tipo:        core.ResultTypeCategorical,
					Descripcion: "Tipo MBTI de 4 letras (ej: INTJ, ESFP)",
				},
				{
					Nombre:      "ocean_vector",
					Tipo:        core.ResultTypeStatistical,
					Descripcion: "Vector OCEAN de personalidad (Big Five)",
				},
			},
		},
	}
}

func buildMBTIQuestions() core.Questions {
	questions := make(core.Questions)

	questions["1"] = "En reuniones sociales..."
	questions["2"] = "Cuando estoy con otras personas..."
	questions["3"] = "Para tomar decisiones prefiero..."
	questions["4"] = "Me describiría como..."
	questions["5"] = "Cuando trabajo en un proyecto..."
	questions["6"] = "En clase o en el trabajo..."
	questions["7"] = "Para recargar energías prefiero..."
	questions["8"] = "Al conocer personas nuevas..."
	questions["9"] = "Me siento más cómodo/a cuando..."
	questions["10"] = "Si me dan tiempo libre no estructurado..."

	questions["11"] = "Cuando aprendo algo nuevo, prefiero..."
	questions["12"] = "Me fijo más en..."
	questions["13"] = "Cuando enfrento un problema, confío más en..."
	questions["14"] = "Me resulta más interesante..."
	questions["15"] = "Al describir una situación, tiendo a..."
	questions["16"] = "Soy más del tipo de persona que..."
	questions["17"] = "Prefiero leer libros que..."
	questions["18"] = "El futuro me parece..."
	questions["19"] = "Cuando trabajo en grupo, aporto más..."
	questions["20"] = "Tiendo a pensar más en..."

	questions["21"] = "Al tomar decisiones, me baso más en..."
	questions["22"] = "Cuando alguien tiene un problema, tiendo a..."
	questions["23"] = "Me resulta más difícil..."
	questions["24"] = "Cuando evalúo una situación..."
	questions["25"] = "Creo que tomar una buena decisión implica..."
	questions["26"] = "Cuando alguien hace algo incorrecto..."
	questions["27"] = "Me identifico más con..."
	questions["28"] = "En un conflicto prefiero..."
	questions["29"] = "Me incomoda más..."
	questions["30"] = "Para criticar a alguien, prefiero..."

	questions["31"] = "Prefiero vivir de manera..."
	questions["32"] = "Cuando tengo una fecha límite..."
	questions["33"] = "Prefiero un entorno de trabajo..."
	questions["34"] = "Me incomoda más..."
	questions["35"] = "Cuando inicio un proyecto, suelo..."
	questions["36"] = "Tomar decisiones me resulta..."
	questions["37"] = "Mi escritorio o espacio de trabajo suele estar..."
	questions["38"] = "Cuando viajo, prefiero..."
	questions["39"] = "Me siento mejor cuando..."
	questions["40"] = "Las reglas y procedimientos me parecen..."

	return questions
}
