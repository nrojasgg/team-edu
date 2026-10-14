package main

import (
	"ms-core/internal/core"
)

func NewBFI44Plugin() *BFI44Plugin {
	return &BFI44Plugin{
		name:    "bfi44",
		version: "1.0.0",
		metadata: &core.PluginMetadata{
			NombreInterno:  "bfi44",
			Nombre:         "BFI-44",
			Version:        "1.0.0",
			Autor:          "Equipo EduTeam",
			Descripcion:    "Big Five Inventory-44: Test de personalidad de los Cinco Grandes factores",
			TipoPlugin:     core.PluginTypeTeamFormation,
			TipoRespuestas: core.ResponseTypeLikert,
			Preguntas:      buildBFI44Questions(),
			Activo:         true,
			Resultados: []core.ResultDefinition{
				{
					Nombre:      "ocean_vector",
					Tipo:        core.ResultTypeStatistical,
					Descripcion: "Vector OCEAN de personalidad (Big Five)",
				},
			},
		},
	}
}

func buildBFI44Questions() core.Questions {
	q := make(core.Questions)

	q["1"] = "Me veo como alguien que es hablador/a."
	q["2"] = "Me veo como alguien que tiende a ser callado/a."
	q["3"] = "Me veo como alguien que es lleno/a de energía."
	q["4"] = "Me veo como alguien que genera mucho entusiasmo."
	q["5"] = "Me veo como alguien que tiene una personalidad asertiva."
	q["6"] = "Me veo como alguien que es tímido/a e inhibido/a."
	q["7"] = "Me veo como alguien que es sociable y extrovertido/a."
	q["8"] = "Me veo como alguien reservado/a."

	q["9"] = "Me veo como alguien que tiene una naturaleza indulgente."
	q["10"] = "Me veo como alguien que es de ayuda y sin egoísmo con los demás."
	q["11"] = "Me veo como alguien que empieza peleas con otros."
	q["12"] = "Me veo como alguien que tiene un corazón blando."
	q["13"] = "Me veo como alguien que es confiable y de fiar."
	q["14"] = "Me veo como alguien que puede ser frío/a y distante."
	q["15"] = "Me veo como alguien que es considerado/a y amable con casi todo el mundo."
	q["16"] = "Me veo como alguien que a veces es rudo/a con los demás."
	q["42"] = "Me veo como alguien que le gusta cooperar con otros."
	q["43"] = "Me veo como alguien que es a veces descortés con los demás."

	q["17"] = "Me veo como alguien que hace un trabajo minucioso."
	q["18"] = "Me veo como alguien que tiende a ser desorganizado/a."
	q["19"] = "Me veo como alguien que puede ser algo descuidado/a."
	q["20"] = "Me veo como alguien que es un/a trabajador/a confiable."
	q["21"] = "Me veo como alguien que persevera hasta que la tarea esté terminada."
	q["22"] = "Me veo como alguien que tiende a ser perezoso/a."
	q["23"] = "Me veo como alguien que hace planes y los lleva a cabo."
	q["24"] = "Me veo como alguien que se distrae fácilmente."
	q["25"] = "Me veo como alguien que es eficiente en el trabajo."

	q["26"] = "Me veo como alguien que está deprimido/a, triste."
	q["27"] = "Me veo como alguien que es relajado/a, maneja el estrés bien."
	q["28"] = "Me veo como alguien que puede ser tenso/a."
	q["29"] = "Me veo como alguien que se preocupa mucho."
	q["30"] = "Me veo como alguien que es emocionalmente estable y no se molesta fácilmente."
	q["31"] = "Me veo como alguien que puede ser malhumorado/a."
	q["32"] = "Me veo como alguien que sigue siendo calmado/a en situaciones tensas."
	q["33"] = "Me veo como alguien que se pone nervioso/a con facilidad."

	q["34"] = "Me veo como alguien que es original, que tiene nuevas ideas."
	q["35"] = "Me veo como alguien que es curioso/a de muchas cosas diferentes."
	q["36"] = "Me veo como alguien que es ingenioso/a, un/a pensador/a profundo/a."
	q["37"] = "Me veo como alguien que tiene una imaginación activa."
	q["38"] = "Me veo como alguien que valora las experiencias artísticas y estéticas."
	q["39"] = "Me veo como alguien que prefiere el trabajo rutinario."
	q["40"] = "Me veo como alguien que le gusta reflexionar y jugar con las ideas."
	q["41"] = "Me veo como alguien que tiene poca interés en las artes."
	q["44"] = "Me veo como alguien sofisticado/a en arte, música o literatura."

	return q
}
