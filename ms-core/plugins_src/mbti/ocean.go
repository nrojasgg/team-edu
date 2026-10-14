package main

import (
	"math"
	"strings"

	"ms-core/internal/core/domain"
)

var mbtiToOCEAN = map[string]domain.OCEANVector{
	"INTJ": {Openness: 0.8, Conscientiousness: 0.7, Extraversion: 0.2, Agreeableness: 0.4, Neuroticism: 0.4},
	"INTP": {Openness: 0.9, Conscientiousness: 0.4, Extraversion: 0.2, Agreeableness: 0.5, Neuroticism: 0.5},
	"ENTJ": {Openness: 0.7, Conscientiousness: 0.8, Extraversion: 0.8, Agreeableness: 0.3, Neuroticism: 0.3},
	"ENTP": {Openness: 0.9, Conscientiousness: 0.4, Extraversion: 0.8, Agreeableness: 0.5, Neuroticism: 0.4},
	"INFJ": {Openness: 0.8, Conscientiousness: 0.6, Extraversion: 0.3, Agreeableness: 0.8, Neuroticism: 0.5},
	"INFP": {Openness: 0.9, Conscientiousness: 0.4, Extraversion: 0.3, Agreeableness: 0.8, Neuroticism: 0.6},
	"ENFJ": {Openness: 0.7, Conscientiousness: 0.7, Extraversion: 0.8, Agreeableness: 0.9, Neuroticism: 0.4},
	"ENFP": {Openness: 0.9, Conscientiousness: 0.4, Extraversion: 0.9, Agreeableness: 0.7, Neuroticism: 0.5},
	"ISTJ": {Openness: 0.3, Conscientiousness: 0.9, Extraversion: 0.3, Agreeableness: 0.6, Neuroticism: 0.4},
	"ISFJ": {Openness: 0.4, Conscientiousness: 0.8, Extraversion: 0.3, Agreeableness: 0.9, Neuroticism: 0.5},
	"ESTJ": {Openness: 0.3, Conscientiousness: 0.9, Extraversion: 0.8, Agreeableness: 0.5, Neuroticism: 0.3},
	"ESFJ": {Openness: 0.4, Conscientiousness: 0.8, Extraversion: 0.9, Agreeableness: 0.9, Neuroticism: 0.4},
	"ISTP": {Openness: 0.6, Conscientiousness: 0.5, Extraversion: 0.3, Agreeableness: 0.4, Neuroticism: 0.3},
	"ISFP": {Openness: 0.7, Conscientiousness: 0.5, Extraversion: 0.4, Agreeableness: 0.8, Neuroticism: 0.5},
	"ESTP": {Openness: 0.6, Conscientiousness: 0.4, Extraversion: 0.9, Agreeableness: 0.5, Neuroticism: 0.3},
	"ESFP": {Openness: 0.6, Conscientiousness: 0.4, Extraversion: 0.9, Agreeableness: 0.7, Neuroticism: 0.4},
}

func getOCEANFromMBTI(mbtiType string) (domain.OCEANVector, bool) {
	mbtiType = strings.ToUpper(strings.TrimSpace(mbtiType))
	v, ok := mbtiToOCEAN[mbtiType]
	return v, ok
}

func isValidMBTIType(mbtiType string) bool {
	mbtiType = strings.ToUpper(strings.TrimSpace(mbtiType))
	_, ok := mbtiToOCEAN[mbtiType]
	return ok
}

func mbtiDimensionIndex(questionID int) int {
	switch {
	case questionID >= 1 && questionID <= 10:
		return 0
	case questionID >= 11 && questionID <= 20:
		return 1
	case questionID >= 21 && questionID <= 30:
		return 2
	case questionID >= 31 && questionID <= 40:
		return 3
	default:
		return -1
	}
}

func tabulateAnswers(answers []MBTIAnswer, result *MBTIResult) {
	for _, answer := range answers {
		idx := mbtiDimensionIndex(answer.QuestionID)
		if idx < 0 {
			continue
		}
		switch answer.Selection {
		case "A":
			result.Dimensions[idx].ScoreA++
		case "B":
			result.Dimensions[idx].ScoreB++
		}
	}
}

func resolveDimensionPreference(dim *MBTIDimensionResult, poleA, poleB string) {
	if dim.ScoreA > dim.ScoreB {
		dim.Preference = poleA
	} else {
		dim.Preference = poleB
	}
	total := dim.ScoreA + dim.ScoreB
	if total > 0 {
		diff := math.Abs(float64(dim.ScoreA - dim.ScoreB))
		dim.PCI = (diff / float64(total)) * 100
	}
}

func defaultOCEAN() domain.OCEANVector {
	return domain.OCEANVector{
		Openness:          0.5,
		Conscientiousness: 0.5,
		Extraversion:      0.5,
		Agreeableness:     0.5,
		Neuroticism:       0.5,
	}
}

func CalculateMBTI(answers []MBTIAnswer) MBTIResult {
	result := MBTIResult{
		Dimensions: make([]MBTIDimensionResult, 4),
		TotalScore: len(answers),
	}

	dims := []MBTIDimension{DimensionEI, DimensionSN, DimensionTF, DimensionJP}
	for i, dim := range dims {
		result.Dimensions[i] = MBTIDimensionResult{Name: dim}
	}

	tabulateAnswers(answers, &result)

	poloA := []string{"E", "S", "T", "J"}
	poloB := []string{"I", "N", "F", "P"}

	mbtiType := ""
	for i := range result.Dimensions {
		resolveDimensionPreference(&result.Dimensions[i], poloA[i], poloB[i])
		mbtiType += result.Dimensions[i].Preference
	}

	result.MBTIType = mbtiType

	if oceanVector, found := getOCEANFromMBTI(mbtiType); found {
		result.OCEANVector = oceanVector
	} else {
		result.OCEANVector = defaultOCEAN()
	}

	return result
}

func CalculateMBTIFromMap(answersMap map[int]string) MBTIResult {
	answers := make([]MBTIAnswer, 0, len(answersMap))
	for qID, selection := range answersMap {
		answers = append(answers, MBTIAnswer{
			QuestionID: qID,
			Selection:  selection,
		})
	}
	return CalculateMBTI(answers)
}

func GetMBTIQuestions() []MBTIQuestion {
	return []MBTIQuestion{
		{ID: 1, Dimension: DimensionEI, Statement: "En reuniones sociales...", OptionA: "Me relaciono fácilmente con personas que no conozco.", OptionB: "Prefiero quedarme con las personas que ya conozco.", ValueA: "E", ValueB: "I", Required: true},
		{ID: 2, Dimension: DimensionEI, Statement: "Cuando estoy con otras personas...", OptionA: "Me motivo y recargo energías.", OptionB: "Me siento agotado/a y necesito tiempo solo/a.", ValueA: "E", ValueB: "I", Required: true},
		{ID: 3, Dimension: DimensionEI, Statement: "Para tomar decisiones prefiero...", OptionA: "Hablar el tema con otros antes de decidir.", OptionB: "Pensar en silencio antes de hablar.", ValueA: "E", ValueB: "I", Required: true},
		{ID: 4, Dimension: DimensionEI, Statement: "Me describiría como...", OptionA: "Abierto/a y comunicativo/a.", OptionB: "Reservado/a y reflexivo/a.", ValueA: "E", ValueB: "I", Required: true},
		{ID: 5, Dimension: DimensionEI, Statement: "Cuando trabajo en un proyecto...", OptionA: "Prefiero colaborar con otras personas.", OptionB: "Prefiero trabajar de manera independiente.", ValueA: "E", ValueB: "I", Required: true},
		{ID: 6, Dimension: DimensionEI, Statement: "En clase o en el trabajo...", OptionA: "Suelo intervenir y expresar mis ideas con frecuencia.", OptionB: "Prefiero escuchar y hablar solo cuando tengo algo concreto que aportar.", ValueA: "E", ValueB: "I", Required: true},
		{ID: 7, Dimension: DimensionEI, Statement: "Para recargar energías prefiero...", OptionA: "Salir y estar con amigos.", OptionB: "Quedarme en casa descansando.", ValueA: "E", ValueB: "I", Required: true},
		{ID: 8, Dimension: DimensionEI, Statement: "Al conocer personas nuevas...", OptionA: "Me acerco a ellas sin dudarlo.", OptionB: "Espero a que me presenten o que ellas den el primer paso.", ValueA: "E", ValueB: "I", Required: true},
		{ID: 9, Dimension: DimensionEI, Statement: "Me siento más cómodo/a cuando...", OptionA: "Hay mucha actividad y conversación a mi alrededor.", OptionB: "El ambiente es tranquilo y con pocas distracciones.", ValueA: "E", ValueB: "I", Required: true},
		{ID: 10, Dimension: DimensionEI, Statement: "Si me dan tiempo libre no estructurado...", OptionA: "Lo aprovecharía para estar con otras personas.", OptionB: "Lo aprovecharía para hacer algo por mi cuenta.", ValueA: "E", ValueB: "I", Required: true},

		{ID: 11, Dimension: DimensionSN, Statement: "Cuando aprendo algo nuevo, prefiero...", OptionA: "Hechos concretos y ejemplos prácticos.", OptionB: "Conceptos generales y sus posibilidades.", ValueA: "S", ValueB: "N", Required: true},
		{ID: 12, Dimension: DimensionSN, Statement: "Me fijo más en...", OptionA: "Los detalles y lo que está pasando ahora.", OptionB: "El panorama general y lo que podría ser.", ValueA: "S", ValueB: "N", Required: true},
		{ID: 13, Dimension: DimensionSN, Statement: "Cuando enfrento un problema, confío más en...", OptionA: "Mi experiencia y lo que ha funcionado antes.", OptionB: "Mi intuición y nuevas ideas.", ValueA: "S", ValueB: "N", Required: true},
		{ID: 14, Dimension: DimensionSN, Statement: "Me resulta más interesante...", OptionA: "Lo práctico y aplicable.", OptionB: "Lo teórico e imaginativo.", ValueA: "S", ValueB: "N", Required: true},
		{ID: 15, Dimension: DimensionSN, Statement: "Al describir una situación, tiendo a...", OptionA: "Dar detalles específicos y concretos.", OptionB: "Dar una visión general con sus implicaciones.", ValueA: "S", ValueB: "N", Required: true},
		{ID: 16, Dimension: DimensionSN, Statement: "Soy más del tipo de persona que...", OptionA: "Confía en lo que puede ver, oír o tocar.", OptionB: "Confía en corazonadas y presentimientos.", ValueA: "S", ValueB: "N", Required: true},
		{ID: 17, Dimension: DimensionSN, Statement: "Prefiero leer libros que...", OptionA: "Expliquen cómo funciona algo paso a paso.", OptionB: "Me den nuevas formas de ver las cosas.", ValueA: "S", ValueB: "N", Required: true},
		{ID: 18, Dimension: DimensionSN, Statement: "El futuro me parece...", OptionA: "Algo en lo que pienso cuando es necesario.", OptionB: "Algo que siempre tengo presente e imagino.", ValueA: "S", ValueB: "N", Required: true},
		{ID: 19, Dimension: DimensionSN, Statement: "Cuando trabajo en grupo, aporto más...", OptionA: "Ideas prácticas y realistas.", OptionB: "Ideas creativas e innovadoras.", ValueA: "S", ValueB: "N", Required: true},
		{ID: 20, Dimension: DimensionSN, Statement: "Tiendo a pensar más en...", OptionA: "Lo que ya ha ocurrido y lo actual.", OptionB: "Lo que podría ocurrir y las posibilidades.", ValueA: "S", ValueB: "N", Required: true},

		{ID: 21, Dimension: DimensionTF, Statement: "Al tomar decisiones, me baso más en...", OptionA: "La lógica y el análisis objetivo.", OptionB: "Los valores y el impacto en las personas.", ValueA: "T", ValueB: "F", Required: true},
		{ID: 22, Dimension: DimensionTF, Statement: "Cuando alguien tiene un problema, tiendo a...", OptionA: "Buscar la solución más lógica y eficaz.", OptionB: "Ofrecer apoyo emocional antes que soluciones.", ValueA: "T", ValueB: "F", Required: true},
		{ID: 23, Dimension: DimensionTF, Statement: "Me resulta más difícil...", OptionA: "Considerar el lado emocional de un asunto.", OptionB: "Ser directo/a cuando alguien podría molestarse.", ValueA: "T", ValueB: "F", Required: true},
		{ID: 24, Dimension: DimensionTF, Statement: "Cuando evalúo una situación...", OptionA: "Busco coherencia y objetividad.", OptionB: "Busco armonía y bienestar grupal.", ValueA: "T", ValueB: "F", Required: true},
		{ID: 25, Dimension: DimensionTF, Statement: "Creo que tomar una buena decisión implica...", OptionA: "Basarse en datos y análisis.", OptionB: "Tener en cuenta los sentimientos de todos.", ValueA: "T", ValueB: "F", Required: true},
		{ID: 26, Dimension: DimensionTF, Statement: "Cuando alguien hace algo incorrecto...", OptionA: "Lo señalo directamente, aunque incomode.", OptionB: "Busco la manera de decirlo sin herir.", ValueA: "T", ValueB: "F", Required: true},
		{ID: 27, Dimension: DimensionTF, Statement: "Me identifico más con...", OptionA: "Las personas que son lógicas y racionales.", OptionB: "Las personas que son empáticas y compasivas.", ValueA: "T", ValueB: "F", Required: true},
		{ID: 28, Dimension: DimensionTF, Statement: "En un conflicto prefiero...", OptionA: "Resolver el problema de forma justa, aunque alguien se moleste.", OptionB: "Mantener la paz y buscar un acuerdo consensuado.", ValueA: "T", ValueB: "F", Required: true},
		{ID: 29, Dimension: DimensionTF, Statement: "Me incomoda más...", OptionA: "Tomar decisiones basadas solo en emociones.", OptionB: "Tomar decisiones que van en contra de mis valores.", ValueA: "T", ValueB: "F", Required: true},
		{ID: 30, Dimension: DimensionTF, Statement: "Para criticar a alguien, prefiero...", OptionA: "Ser directo/a y preciso/a.", OptionB: "Ser cuidadoso/a y considerar cómo se sentirá.", ValueA: "T", ValueB: "F", Required: true},

		{ID: 31, Dimension: DimensionJP, Statement: "Prefiero vivir de manera...", OptionA: "Planificada y organizada.", OptionB: "Flexible y espontánea.", ValueA: "J", ValueB: "P", Required: true},
		{ID: 32, Dimension: DimensionJP, Statement: "Cuando tengo una fecha límite...", OptionA: "Termino la tarea con anticipación.", OptionB: "Trabajo mejor con la presión del plazo.", ValueA: "J", ValueB: "P", Required: true},
		{ID: 33, Dimension: DimensionJP, Statement: "Prefiero un entorno de trabajo...", OptionA: "Con rutinas y estructura definida.", OptionB: "Con variedad y posibilidad de improvisar.", ValueA: "J", ValueB: "P", Required: true},
		{ID: 34, Dimension: DimensionJP, Statement: "Me incomoda más...", OptionA: "Que un plan cambie de última hora.", OptionB: "Quedarme comprometido/a con un plan cuando hay mejores opciones.", ValueA: "J", ValueB: "P", Required: true},
		{ID: 35, Dimension: DimensionJP, Statement: "Cuando inicio un proyecto, suelo...", OptionA: "Organizarlo todo antes de comenzar.", OptionB: "Empezar y ver cómo avanza.", ValueA: "J", ValueB: "P", Required: true},
		{ID: 36, Dimension: DimensionJP, Statement: "Tomar decisiones me resulta...", OptionA: "Liberador, me gusta tener todo resuelto.", OptionB: "Difícil, siempre podría surgir algo nuevo.", ValueA: "J", ValueB: "P", Required: true},
		{ID: 37, Dimension: DimensionJP, Statement: "Mi escritorio o espacio de trabajo suele estar...", OptionA: "Ordenado y cada cosa en su lugar.", OptionB: "Con algunas cosas dispersas, pero sé dónde está todo.", ValueA: "J", ValueB: "P", Required: true},
		{ID: 38, Dimension: DimensionJP, Statement: "Cuando viajo, prefiero...", OptionA: "Tener el itinerario planeado de antemano.", OptionB: "Decidir qué hacer en el momento.", ValueA: "J", ValueB: "P", Required: true},
		{ID: 39, Dimension: DimensionJP, Statement: "Me siento mejor cuando...", OptionA: "He tomado una decisión y puedo avanzar.", OptionB: "Tengo opciones abiertas y flexibilidad.", ValueA: "J", ValueB: "P", Required: true},
		{ID: 40, Dimension: DimensionJP, Statement: "Las reglas y procedimientos me parecen...", OptionA: "Útiles: dan estructura y claridad.", OptionB: "Limitantes: prefiero adaptarme según el contexto.", ValueA: "J", ValueB: "P", Required: true},
	}
}

func (p *MBTIPlugin) GetMBTICompatibility(typeA, typeB string) float64 {
	if len(typeA) != 4 || len(typeB) != 4 {
		return 0.5
	}

	if typeA == typeB {
		return 0.7
	}

	score := 0.0

	for i := 0; i < 4; i++ {
		if typeA[i] == typeB[i] {
			score += 0.15
		}
	}

	score += getMBTIGroupBonus(typeA, typeB)

	if score > 1.0 {
		score = 1.0
	}

	return score
}

func getMBTIGroupBonus(typeA, typeB string) float64 {
	score := 0.0
	if isNT(typeA) && isNT(typeB) {
		score += 0.2
	}
	if isNF(typeA) && isNF(typeB) {
		score += 0.2
	}
	if isSJ(typeA) && isSJ(typeB) {
		score += 0.15
	}
	if isSP(typeA) && isSP(typeB) {
		score += 0.15
	}
	if isComplementary(typeA, typeB) {
		score += 0.25
	}
	return score
}

func isNT(mbtiType string) bool {
	return len(mbtiType) == 4 && mbtiType[1] == 'N' && mbtiType[2] == 'T'
}

func isNF(mbtiType string) bool {
	return len(mbtiType) == 4 && mbtiType[1] == 'N' && mbtiType[2] == 'F'
}

func isSJ(mbtiType string) bool {
	return len(mbtiType) == 4 && mbtiType[1] == 'S' && mbtiType[3] == 'J'
}

func isSP(mbtiType string) bool {
	return len(mbtiType) == 4 && mbtiType[1] == 'S' && mbtiType[3] == 'P'
}

func isComplementary(typeA, typeB string) bool {
	pairs := map[string]string{
		"INTJ": "ENTP", "ENTP": "INTJ",
		"INFJ": "ENFP", "ENFP": "INFJ",
		"INTP": "ENTJ", "ENTJ": "INTP",
		"INFP": "ENFJ", "ENFJ": "INFP",
		"ISTJ": "ESTP", "ESTP": "ISTJ",
		"ISFJ": "ESFP", "ESFP": "ISFJ",
		"ISTP": "ESTJ", "ESTJ": "ISTP",
		"ISFP": "ESFJ", "ESFJ": "ISFP",
	}

	if pair, exists := pairs[typeA]; exists && pair == typeB {
		return true
	}
	return false
}
