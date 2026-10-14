/**
 * BFI-44 - Big Five Inventory (44 items)
 * Cuestionario de Personalidad con las 5 dimensiones OCEAN.
 *
 * Cada pregunta tiene:
 *   - id: identificador único (1–44)
 *   - text: enunciado en español
 *   - dimension: dimensión OCEAN ('O', 'C', 'E', 'A', 'N')
 *   - reversed: true si el ítem es de puntuación invertida
 */
export const bfi44Questions = [
    // Extraversión (E)
    { id: 1, text: 'Me veo como alguien que es hablador/a.', dimension: 'E', reversed: false },
    { id: 2, text: 'Me veo como alguien que tiende a ser callado/a.', dimension: 'E', reversed: true },
    { id: 3, text: 'Me veo como alguien que es lleno/a de energía.', dimension: 'E', reversed: false },
    { id: 4, text: 'Me veo como alguien que genera mucho entusiasmo.', dimension: 'E', reversed: false },
    { id: 5, text: 'Me veo como alguien que tiene una personalidad asertiva.', dimension: 'E', reversed: false },
    { id: 6, text: 'Me veo como alguien que es tímido/a e inhibido/a.', dimension: 'E', reversed: true },
    { id: 7, text: 'Me veo como alguien que es sociable y extrovertido/a.', dimension: 'E', reversed: false },
    { id: 8, text: 'Me veo como alguien reservado/a.', dimension: 'E', reversed: true },

    // Amabilidad (A)
    { id: 9, text: 'Me veo como alguien que tiene una naturaleza indulgente.', dimension: 'A', reversed: false },
    { id: 10, text: 'Me veo como alguien que es de ayuda y sin egoísmo con los demás.', dimension: 'A', reversed: false },
    { id: 11, text: 'Me veo como alguien que empieza peleas con otros.', dimension: 'A', reversed: true },
    { id: 12, text: 'Me veo como alguien que tiene un corazón blando.', dimension: 'A', reversed: false },
    { id: 13, text: 'Me veo como alguien que es confiable y de fiar.', dimension: 'A', reversed: false },
    { id: 14, text: 'Me veo como alguien que puede ser frío/a y distante.', dimension: 'A', reversed: true },
    { id: 15, text: 'Me veo como alguien que es considerado/a y amable con casi todo el mundo.', dimension: 'A', reversed: false },
    { id: 16, text: 'Me veo como alguien que a veces es rudo/a con los demás.', dimension: 'A', reversed: true },

    // Responsabilidad (C)
    { id: 17, text: 'Me veo como alguien que hace un trabajo minucioso.', dimension: 'C', reversed: false },
    { id: 18, text: 'Me veo como alguien que tiende a ser desorganizado/a.', dimension: 'C', reversed: true },
    { id: 19, text: 'Me veo como alguien que puede ser algo descuidado/a.', dimension: 'C', reversed: true },
    { id: 20, text: 'Me veo como alguien que es un/a trabajador/a confiable.', dimension: 'C', reversed: false },
    { id: 21, text: 'Me veo como alguien que persevera hasta que la tarea esté terminada.', dimension: 'C', reversed: false },
    { id: 22, text: 'Me veo como alguien que tiende a ser perezoso/a.', dimension: 'C', reversed: true },
    { id: 23, text: 'Me veo como alguien que hace planes y los lleva a cabo.', dimension: 'C', reversed: false },
    { id: 24, text: 'Me veo como alguien que se distrae fácilmente.', dimension: 'C', reversed: true },
    { id: 25, text: 'Me veo como alguien que es eficiente en el trabajo.', dimension: 'C', reversed: false },

    // Neuroticismo (N)
    { id: 26, text: 'Me veo como alguien que está deprimido/a, triste.', dimension: 'N', reversed: false },
    { id: 27, text: 'Me veo como alguien que es relajado/a, maneja el estrés bien.', dimension: 'N', reversed: true },
    { id: 28, text: 'Me veo como alguien que puede ser tenso/a.', dimension: 'N', reversed: false },
    { id: 29, text: 'Me veo como alguien que se preocupa mucho.', dimension: 'N', reversed: false },
    { id: 30, text: 'Me veo como alguien que es emocionalmente estable y no se molesta fácilmente.', dimension: 'N', reversed: true },
    { id: 31, text: 'Me veo como alguien que puede ser malhumorado/a.', dimension: 'N', reversed: false },
    { id: 32, text: 'Me veo como alguien que sigue siendo calmado/a en situaciones tensas.', dimension: 'N', reversed: true },
    { id: 33, text: 'Me veo como alguien que se pone nervioso/a con facilidad.', dimension: 'N', reversed: false },

    // Apertura a la Experiencia (O)
    { id: 34, text: 'Me veo como alguien que es original, que tiene nuevas ideas.', dimension: 'O', reversed: false },
    { id: 35, text: 'Me veo como alguien que es curioso/a de muchas cosas diferentes.', dimension: 'O', reversed: false },
    { id: 36, text: 'Me veo como alguien que es ingenioso/a, un/a pensador/a profundo/a.', dimension: 'O', reversed: false },
    { id: 37, text: 'Me veo como alguien que tiene una imaginación activa.', dimension: 'O', reversed: false },
    { id: 38, text: 'Me veo como alguien que valora las experiencias artísticas y estéticas.', dimension: 'O', reversed: false },
    { id: 39, text: 'Me veo como alguien que prefiere el trabajo rutinario.', dimension: 'O', reversed: true },
    { id: 40, text: 'Me veo como alguien que le gusta reflexionar y jugar con las ideas.', dimension: 'O', reversed: false },
    { id: 41, text: 'Me veo como alguien que tiene poca interés en las artes.', dimension: 'O', reversed: true },
    { id: 42, text: 'Me veo como alguien que le gusta cooperar con otros.', dimension: 'A', reversed: false },
    { id: 43, text: 'Me veo como alguien que es a veces descortés con los demás.', dimension: 'A', reversed: true },
    { id: 44, text: 'Me veo como alguien sofisticado/a en arte, música o literatura.', dimension: 'O', reversed: false },
]

/**
 * Etiquetas de la escala Likert (1–5) para el BFI-44.
 * Clave: número entero del 1 al 5.
 */
export const likertLabels = {
    1: 'Muy en desacuerdo',
    2: 'En desacuerdo',
    3: 'Ni acuerdo ni desacuerdo',
    4: 'De acuerdo',
    5: 'Muy de acuerdo',
}
