package main

import (
	"ms-core/internal/core"
)

func NewEnneagramPlugin() *EnneagramPlugin {
	return &EnneagramPlugin{
		name:    "enneagram",
		version: "1.0.0",
		metadata: &core.PluginMetadata{
			NombreInterno:  "enneagram",
			Nombre:         "Eneagrama",
			Version:        "1.0.0",
			Autor:          "Equipo EduTeam",
			Descripcion:    "Test de Eneagrama: sistema de personalidad de 9 tipos",
			TipoPlugin:     core.PluginTypeTeamFormation,
			TipoRespuestas: core.ResponseTypeLikert,
			Preguntas:      buildEnneagramQuestions(),
			Activo:         true,
			Resultados: []core.ResultDefinition{
				{
					Nombre:      "dominant_type",
					Tipo:        core.ResultTypeCategorical,
					Descripcion: "Tipo dominante Eneagrama (1-9)",
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

func buildEnneagramQuestions() core.Questions {
	q := make(core.Questions)

	q["1"] = "Creo que he pagado un elevado precio por tratar de ser perfecto."
	q["2"] = "Mis principios e ideales me estimulan hacia mayores logros y dan sentido y valor a mi vida."
	q["3"] = "Depende tanto de mí que se hagan las cosas que tengo que ser más organizado y metódico que los demás."
	q["4"] = "Me imagino que tengo una misión personal, tal vez una vocación para algo superior, y creo que durante mi vida podría realizar algo extraordinario."
	q["5"] = "Detesto los errores, por lo tanto tiendo a ser muy minucioso para asegurarme de que las cosas se hagan bien."

	q["6"] = "Mi auténtico interés por los demás hace que me involucre profundamente con ellos, con sus esperanzas, sus sueños y necesidades."
	q["7"] = "No puedo ver un perro extraviado en la calle sin desejar llevármelo a casa."
	q["8"] = "Es cierto que suelo hacer más por los demás de lo que debiera; doy demasiado y no pienso en mí lo suficiente."
	q["9"] = "Hago un esfuerzo especial por saber qué les pasa a las personas que quiero."
	q["10"] = "Con frecuencia mi salud y mi economía han sufrido debido a que antepongo las necesidades e intereses de los demás a los míos."

	q["11"] = "Me considero una persona muy competente: en realidad, me molesta no ser eficaz y eficiente."
	q["12"] = "Es importante para mí sentir que tengo éxito, aunque aún no sea el éxito que deseo."
	q["13"] = "Para bien o para mal, soy bueno para encubrir mis inseguridades; nadie adivina jamás qué siento realmente."
	q["14"] = "Sé lo bien que les va a mis amigos y colegas, y tiendo a compararme con ellos."
	q["15"] = "Tengo una adicción al trabajo; me siento perdido si no estoy realizando cosas."

	q["16"] = "Muchos me consideran una persona enigmática, difícil y contradictoria, ¡y eso me gusta de mí!, soy bastante dramática y temperamental."
	q["17"] = "Tiendo a no seguir las reglas ni las expectativas porque quiero poner mi sello especial en todo lo que hago. Me cuesta participar en proyectos si no tengo el control creativo."
	q["18"] = "Tiendo a pasar bastante tiempo imaginando situaciones y conversaciones que no han ocurrido necesariamente."
	q["19"] = "Conocerme a mí mismo y ser fiel a mis necesidades emocionales han sido motivaciones muy importantes para mí."
	q["20"] = "Soy muy consciente de mis intuiciones, tenga o no el valor para seguirlas."

	q["21"] = "Me gusta profundizar en las cosas, y estudio detenidamente los detalles hasta que lo he entendido todo lo mejor posible."
	q["22"] = "Soy una persona muy reservada, no dejo entrar a muchas personas en mi mundo."
	q["23"] = "Solo teniendo la información correcta se puede tomar una decisión racional, pero, claro, la mayoría de la gente no es muy racional."
	q["24"] = "Mi familia me encuentra algo raro o excéntrico; certamente me han dicho que debo salir más."
	q["25"] = "Soy muy curioso y me gusta investigar por qué las cosas son como son; incluso las cosas obvias no son tan obvias si se miran bien."

	q["26"] = "Soy muy emotivo, pero no suele demostrar mis sentimientos, a no ser a mis íntimos, e incluso a ellos, no siempre."
	q["27"] = "Me siento más seguro haciendo lo que se espera de mí que obrando por mi cuenta."
	q["28"] = "Sé cuánto enredo armo yo, de modo que encuentro lógico desconfiar de lo que pretenden los demás."
	q["29"] = "Deseo fiarme de las personas, pero suelo sorprenderme desconfiando de sus motivos."
	q["30"] = "Cuando tengo que tomar una decisión importante, consulto la opinión de las personas en que confío."

	q["31"] = "Me encanta viajar y descubrir diferentes tipos de comidas, personas y experiencias, ¡todo el fabuloso torbellino de la vida!"
	q["32"] = "Una cosa que no soporto de ninguna manera es el aburrimiento, y procuro no aburrirme nunca, aunque me distraiga con facilidad y me disperse demasiado."
	q["33"] = "Me comprometo bastante cuando estoy en una relación, pero cuando se acaba, paso a otra cosa. De hecho, cuando ya no disfruto en una actividad dejo de hacerla."
	q["34"] = "Soy curioso y arriesgado; por lo general, soy el primero entre mis amigos en probar cualquier cosa nueva e interesante."
	q["35"] = "Tiendo a gastar más dinero del que provavelmente debería y me encanta estar con gente, siempre que quieran ir donde yo voy."

	q["36"] = "Soy extraordinariamente independiente y no me gusta depender de nadie para lo que necesito."
	q["37"] = "Cuando quiero a las personas las considero «mi gente» y siento la necesidad de cuidar de sus intereses, pero no siento mucha compasión por los débiles y los indecisos: la debilidad solo atrae problemas."
	q["38"] = "Tengo una voluntad fuerte y no renuncio ni me echo atrás fácilmente. Incluso, soy capaz de violentos estallidos de cólera, pero se me pasan."
	q["39"] = "Me considero un desafiador, una persona que saca a las personas de su zona de comodidad para que hagan todo lo que son capaces de hacer."
	q["40"] = "Mi sentido del humor es prosaico, incluso grosero a veces, aunque creo que la mayoría de las personas son demasiado remilgadas y sensibles."

	q["41"] = "Parece que lo que más gusta de mí a las personas es que se sienten seguras a mi lado."
	q["42"] = "No me importa estar con gente ni me importa estar solo; cualquiera de las dos cosas me va bien, siempre que esté en paz conmigo mismo."
	q["43"] = "He encontrado un cierto equilibrio en mi vida, y no veo ningún motivo para estropearlo. Estar «cómodo» en todo el sentido de la palabra me atrae muchísimo."
	q["44"] = "Prefiero ceder antes que armar una escena. Acentúo lo positivo antes que insistir en lo negativo."
	q["45"] = "Me es fácil comprender diferentes puntos de vista y tiendo más a estar de acuerdo que en desacuerdo con las personas."

	return q
}
