package main

import (
	"fmt"
)

func getChronotypeQuestionsFull() []ChronotypeQuestion {
	return []ChronotypeQuestion{
		{
			ID:   1,
			Text: "Si solo pensaras en cuando te sentirias mejor y fueras totalmente libre de planificarte el dia. A que hora te levantarias?",
			Options: []ChronotypeOption{
				{Score: 5, Text: "05:00-06:30"},
				{Score: 4, Text: "06:30-07:45"},
				{Score: 3, Text: "07:45-09:45"},
				{Score: 2, Text: "09:45-11:00"},
				{Score: 1, Text: "11:00-12:00"},
			},
			Required: true,
		},
		{
			ID:   2,
			Text: "Si solo pensaras en cuando te sentirias mejor y fueras totalmente libre de planificarte el dia. A que hora te acostarias?",
			Options: []ChronotypeOption{
				{Score: 5, Text: "20:00-21:00"},
				{Score: 4, Text: "21:00-22:15"},
				{Score: 3, Text: "22:15-00:30"},
				{Score: 2, Text: "00:30-01:45"},
				{Score: 1, Text: "01:45-03:00"},
			},
			Required: true,
		},
		{
			ID:   3,
			Text: "Para levantarte por la manana a una hora especifica. Hasta que punto necesitas que te avise el despertador?",
			Options: []ChronotypeOption{
				{Score: 4, Text: "No lo necesito"},
				{Score: 3, Text: "Lo necesito poco"},
				{Score: 2, Text: "Lo necesito bastante"},
				{Score: 1, Text: "Lo necesito mucho"},
			},
			Required: true,
		},
		{
			ID:   4,
			Text: "Te resulta facil levantarte por las mananas? (cuando no te despiertan de forma inesperada)",
			Options: []ChronotypeOption{
				{Score: 1, Text: "Nada facil"},
				{Score: 2, Text: "No muy facil"},
				{Score: 3, Text: "Bastante facil"},
				{Score: 4, Text: "Muy facil"},
			},
			Required: true,
		},
		{
			ID:   5,
			Text: "Una vez levantado por las mananas. Que tal te encuentras durante la primera media hora?",
			Options: []ChronotypeOption{
				{Score: 1, Text: "Nada alerta"},
				{Score: 2, Text: "Poco alerta"},
				{Score: 3, Text: "Bastante alerta"},
				{Score: 4, Text: "Muy alerta"},
			},
			Required: true,
		},
		{
			ID:   6,
			Text: "Una vez levantado por las mananas. Como es tu apetito durante la primera media hora?",
			Options: []ChronotypeOption{
				{Score: 1, Text: "Muy escaso"},
				{Score: 2, Text: "Bastante escaso"},
				{Score: 3, Text: "Bastante bueno"},
				{Score: 4, Text: "Muy bueno"},
			},
			Required: true,
		},
		{
			ID:   7,
			Text: "Una vez levantado por las mananas. Que tal te sientes durante la primera media hora?",
			Options: []ChronotypeOption{
				{Score: 1, Text: "Muy cansado"},
				{Score: 2, Text: "Bastante cansado"},
				{Score: 3, Text: "Bastante descansado"},
				{Score: 4, Text: "Muy descansado"},
			},
			Required: true,
		},
		{
			ID:   8,
			Text: "Cuando no tienes compromisos al dia siguiente. A que hora te acuestas en relacion con tu hora habitual?",
			Options: []ChronotypeOption{
				{Score: 4, Text: "Nunca o raramente mas tarde"},
				{Score: 3, Text: "Menos de 1 hora mas tarde"},
				{Score: 2, Text: "De 1 a 2 horas mas tarde"},
				{Score: 1, Text: "Mas de 2 horas mas tarde"},
			},
			Required: true,
		},
		{
			ID:   9,
			Text: "Has decidido hacer un poco de ejercicio fisico. Un amigo te propone hacerlo una hora dos veces por semana y segun el, la mejor hora seria de 7 a 8 de la manana. Como crees que te encontrarias?",
			Options: []ChronotypeOption{
				{Score: 4, Text: "Estaria en buena forma"},
				{Score: 3, Text: "Estaria en forma aceptable"},
				{Score: 2, Text: "Me resultaria dificil"},
				{Score: 1, Text: "Me resultaria muy dificil"},
			},
			Required: true,
		},
		{
			ID:   10,
			Text: "A que hora aproximada de la noche te sientes cansado y como consecuencia necesitas dormir?",
			Options: []ChronotypeOption{
				{Score: 5, Text: "20:00-21:00"},
				{Score: 4, Text: "21:00-22:15"},
				{Score: 3, Text: "22:15-00:45"},
				{Score: 2, Text: "00:45-02:00"},
				{Score: 1, Text: "02:00-03:00"},
			},
			Required: true,
		},
		{
			ID:   11,
			Text: "Quieres estar en tu punto maximo de rendimiento para una prueba de dos horas mentalmente agotadora. Que horario elegirias?",
			Options: []ChronotypeOption{
				{Score: 6, Text: "08:00-10:00"},
				{Score: 4, Text: "11:00-13:00"},
				{Score: 2, Text: "13:00-17:00"},
				{Score: 0, Text: "19:00-21:00"},
			},
			Required: true,
		},
		{
			ID:   12,
			Text: "Si te acostaras a las 11 de la noche. Que nivel de cansancio noterias?",
			Options: []ChronotypeOption{
				{Score: 0, Text: "Ningun cansancio"},
				{Score: 2, Text: "Algun cansancio"},
				{Score: 3, Text: "Bastante cansancio"},
				{Score: 5, Text: "Mucho cansancio"},
			},
			Required: true,
		},
		{
			ID:   13,
			Text: "Por algun motivo te has acostado varias horas mas tarde de lo habitual, aunque al dia siguiente no has de levantarte a ninguna hora en particular. Cuando crees que te despertarias?",
			Options: []ChronotypeOption{
				{Score: 4, Text: "A la hora habitual y ya no dormiria mas"},
				{Score: 3, Text: "A la hora habitual y luego dormitaria"},
				{Score: 2, Text: "A la hora habitual y volveria a dormirme"},
				{Score: 1, Text: "Mas tarde de lo habitual"},
			},
			Required: true,
		},
		{
			ID:   14,
			Text: "Una noche tienes que permanecer despierto de 4 a 6 de la madrugada. Sin tener ningun compromiso al dia siguiente, que preferiras?",
			Options: []ChronotypeOption{
				{Score: 1, Text: "No acostarme hasta pasada la guardia"},
				{Score: 2, Text: "Echar una siesta antes y dormir despues"},
				{Score: 3, Text: "Echar un buen sueno antes y una siesta despues"},
				{Score: 4, Text: "Dormir solo antes de la guardia"},
			},
			Required: true,
		},
		{
			ID:   15,
			Text: "Tienes que hacer dos horas de trabajo fisico pesado. Pensando solo en cuando te sentirias mejor, que horario escogeras?",
			Options: []ChronotypeOption{
				{Score: 4, Text: "08:00-10:00"},
				{Score: 3, Text: "11:00-13:00"},
				{Score: 2, Text: "13:00-17:00"},
				{Score: 1, Text: "19:00-21:00"},
			},
			Required: true,
		},
		{
			ID:   16,
			Text: "Has decidido hacer ejercicio fisico intenso. Un amigo te sugiere hacerlo de 10 a 11 de la noche. Como crees que te sentaria?",
			Options: []ChronotypeOption{
				{Score: 1, Text: "Estaria en buena forma"},
				{Score: 2, Text: "Estaria en forma aceptable"},
				{Score: 3, Text: "Me resultaria dificil"},
				{Score: 4, Text: "Me resultaria muy dificil"},
			},
			Required: true,
		},
		{
			ID:   17,
			Text: "Imaginate que puedes escoger tu horario de trabajo. Que CINCO HORAS CONSECUTIVAS seleccionarias? Empezando en que hora?",
			Options: []ChronotypeOption{
				{Score: 5, Text: "04:00-08:00"},
				{Score: 4, Text: "08:00-09:00"},
				{Score: 3, Text: "09:00-14:00"},
				{Score: 2, Text: "14:00-17:00"},
				{Score: 1, Text: "17:00-04:00"},
			},
			Required: true,
		},
		{
			ID:   18,
			Text: "A que hora del dia crees que alcanzas tu maximo bienestar?",
			Options: []ChronotypeOption{
				{Score: 5, Text: "05:00-08:00"},
				{Score: 4, Text: "08:00-10:00"},
				{Score: 3, Text: "10:00-17:00"},
				{Score: 2, Text: "17:00-22:00"},
				{Score: 1, Text: "22:00-05:00"},
			},
			Required: true,
		},
		{
			ID:   19,
			Text: "Se habla de personas de tipo matutino y vespertino. Cual de estos tipos te consideras ser?",
			Options: []ChronotypeOption{
				{Score: 6, Text: "Claramente matutino"},
				{Score: 4, Text: "Mas matutino que vespertino"},
				{Score: 2, Text: "Mas vespertino que matutino"},
				{Score: 0, Text: "Claramente vespertino"},
			},
			Required: true,
		},
	}
}

func (p *ChronotypePlugin) Calculate(inputs map[int]int) (ChronotypeResults, error) {
	questions := getChronotypeQuestionsFull()

	if err := validateAnswers(questions, inputs); err != nil {
		return ChronotypeResults{}, err
	}

	totalScore := calculateTotalScore(inputs)
	category := categorize(totalScore)

	return ChronotypeResults{
		TotalScore: totalScore,
		Category:   category,
	}, nil
}

func validateAnswers(questions []ChronotypeQuestion, inputs map[int]int) error {
	for _, q := range questions {
		score, exists := inputs[q.ID]
		if !exists {
			return fmt.Errorf("missing answer for question %d", q.ID)
		}
		if !isValidScore(score, q.Options) {
			return fmt.Errorf("invalid score %d for question %d", score, q.ID)
		}
	}
	return nil
}

func isValidScore(score int, options []ChronotypeOption) bool {
	for _, opt := range options {
		if score == opt.Score {
			return true
		}
	}
	return false
}

func calculateTotalScore(inputs map[int]int) int {
	total := 0
	for _, score := range inputs {
		total += score
	}
	return total
}

func categorize(score int) string {
	switch {
	case score >= 70:
		return "Matutino extremo"
	case score >= 59:
		return "Matutino moderado"
	case score >= 42:
		return "Intermedio"
	case score >= 31:
		return "Vespertino moderado"
	default:
		return "Vespertino extremo"
	}
}
