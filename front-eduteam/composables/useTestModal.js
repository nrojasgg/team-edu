import { ref, computed } from 'vue'
import { usePluginApi } from '~/composables/usePluginApi'

/**
 * Shared composable for psychological test modals (BFI-44, MBTI, etc.)
 *
 * Encapsulates:
 *  - JWT auth helpers (getUserIdFromToken, isAuthenticated)
 *  - Question navigation state and helpers
 *  - Generic checkExistingResponse / loadQuestions / submitTest flow
 *  - closeModal reset helper
 *
 * @param {object} options
 * @param {import('vue').Ref}  options.questionnaire - reactive questionnaire prop
 * @param {string}  options.apiPath       - API path segment, e.g. 'bfi44' or 'mbti'
 * @param {Function} options.getQuestions  - () => Question[] — returns the static question list
 * @param {Function} options.mapAnswer     - (questionId, value) => object — formats one answer entry for the POST body
 * @param {Function} options.onResult      - (result, { isCompleted }) => void — called after successful submit/check
 * @param {Function} options.onClose       - () => void — called by closeModal to reset result-specific state
 * @param {import('vue').Emit} options.emit - component's emit function
 */
export function useTestModal({ questionnaire, apiPath, getQuestions, mapAnswer, onResult, onClose, emit }) {
    const pluginApi = usePluginApi()
    // ── State ────────────────────────────────────────────────────────────────
    const loading = ref(false)
    const error = ref('')
    const questions = ref([])
    const currentQuestionIndex = ref(0)
    const answers = ref({})
    const selectedAnswer = ref(null)
    const isCompleted = ref(false)
    const submitting = ref(false)

    // ── Computed ─────────────────────────────────────────────────────────────
    const currentQuestion = computed(() => questions.value[currentQuestionIndex.value] || {})

    // ── Auth helpers ─────────────────────────────────────────────────────────

    /**
     * Decodes a JWT stored in localStorage and returns the user id.
     * Returns null if the token is missing or malformed.
     */
    const getUserIdFromToken = () => {
        try {
            const token = localStorage.getItem('auth_token')
            if (!token) return null

            const payload = token.split('.')[1]
            const decodedPayload = JSON.parse(atob(payload))
            const userId = decodedPayload.user_id || decodedPayload.id || decodedPayload.sub

            return userId ? userId.toString() : null
        } catch (err) {
            console.error('Error decodificando token JWT:', err)
            return null
        }
    }

    /**
     * Returns true if there is a non-expired auth_token in localStorage.
     */
    const isAuthenticated = () => {
        const token = localStorage.getItem('auth_token')
        const userId = getUserIdFromToken()
        if (!token || !userId) return false

        try {
            const payload = JSON.parse(atob(token.split('.')[1]))
            const currentTime = Math.floor(Date.now() / 1000)
            return payload.exp >= currentTime
        } catch (err) {
            console.error('Error en autenticación:', err)
            return false
        }
    }

    // ── Navigation helpers ────────────────────────────────────────────────────

    const saveCurrentAnswer = () => {
        if (selectedAnswer.value !== null) {
            answers.value[currentQuestion.value.id] = selectedAnswer.value
        }
    }

    const loadCurrentAnswer = () => {
        selectedAnswer.value = answers.value[currentQuestion.value.id] ?? null
    }

    const selectAnswer = (value) => {
        selectedAnswer.value = value
        saveCurrentAnswer()
    }

    const previousQuestion = () => {
        if (currentQuestionIndex.value > 0) {
            saveCurrentAnswer()
            currentQuestionIndex.value--
            loadCurrentAnswer()
        }
    }

    const nextQuestion = () => {
        if (selectedAnswer.value !== null && currentQuestionIndex.value < questions.value.length - 1) {
            saveCurrentAnswer()
            currentQuestionIndex.value++
            loadCurrentAnswer()
        }
    }

    // ── Core async logic ──────────────────────────────────────────────────────

    /**
     * Checks if the current student already has a complete response for this test.
     * If a complete response is found, sets isCompleted and calls onResult with the
     * existing result data.
     * @returns {boolean} true if already completed
     */
    const checkExistingResponse = async () => {
        if (!questionnaire.value?.id) return false

        if (!isAuthenticated()) {
            console.warn('Usuario no autenticado')
            return false
        }

        try {
            const token = localStorage.getItem('auth_token')
            const studentId = getUserIdFromToken()

            if (!studentId) return false

            const { success, data } = await pluginApi.getSectionResponses(apiPath, questionnaire.value.section_id)

            if (success && Array.isArray(data)) {
                const userResponse = data.find(r => 
                    r.student_id?.toString() === studentId.toString() &&
                    r.test_id === questionnaire.value.id
                )
                if (userResponse?.is_complete) {
                    isCompleted.value = true
                    onResult(userResponse.results, { isCompleted })
                    console.log(`Usuario ya completó el test (${apiPath})`)
                    return true
                }
            }
        } catch (err) {
            console.error('Error verificando respuesta existente:', err)
        }

        return false
    }

    /**
     * Loads the static question list for this test after verifying auth and
     * checking for an existing completion.
     */
    const loadQuestions = async () => {
        try {
            loading.value = true
            error.value = ''

            if (!questionnaire.value) {
                throw new Error('No se ha seleccionado un cuestionario')
            }

            if (!isAuthenticated()) {
                throw new Error('Debes estar autenticado para acceder al test')
            }

            if (await Promise.resolve(checkExistingResponse())) {
                loading.value = false
                return
            }

            questions.value = getQuestions()
            console.log(`Preguntas ${apiPath} cargadas:`, questions.value.length)
        } catch (err) {
            console.error('Error en loadQuestions:', err)
            error.value = err.message
            questions.value = []
        } finally {
            loading.value = false
        }
    }

    /**
     * Submits all answers to the API endpoint for this test type.
     */
    const submitTest = async () => {
        if (selectedAnswer.value === null || !questionnaire.value?.id) return

        if (!isAuthenticated()) {
            error.value = 'Debes estar autenticado para enviar el test'
            return
        }

        saveCurrentAnswer()

        const totalQuestions = questions.value.length
        const answeredQuestions = Object.keys(answers.value).length

        if (answeredQuestions < totalQuestions) {
            error.value = `Por favor responde todas las preguntas. Has respondido ${answeredQuestions} de ${totalQuestions}.`
            return
        }

        try {
            submitting.value = true
            const token = localStorage.getItem('auth_token')
            const studentId = getUserIdFromToken()

            if (!studentId) {
                throw new Error('No se pudo obtener el ID del usuario')
            }

            // mapAnswer is provided by each component — formats each answer correctly
            const answersArray = Object.entries(answers.value).map(([questionId, value]) =>
                mapAnswer(Number.parseInt(questionId), value)
            )

            const requestData = {
                student_id: studentId,
                answers: answersArray
            }

            console.log(`Enviando respuestas ${apiPath}:`, requestData)

            const { success, data, error: apiError } = await pluginApi.submitResponse(
                apiPath,
                questionnaire.value.id,
                answersArray,
                studentId
            )

            if (!success) {
                throw new Error(apiError || 'Error al enviar las respuestas')
            }

            console.log('Respuesta del servidor:', data)

            onResult(data, { isCompleted })
            emit('completed', data)
            isCompleted.value = true
        } catch (err) {
            console.error('Error enviando test:', err)
            error.value = err.message
        } finally {
            submitting.value = false
        }
    }

    /**
     * Resets all shared state and emits 'close'. Individual components can
     * use the onClose callback to reset their own result refs.
     */
    const closeModal = () => {
        currentQuestionIndex.value = 0
        selectedAnswer.value = null
        answers.value = {}
        error.value = ''
        questions.value = []
        loading.value = false
        submitting.value = false
        isCompleted.value = false

        onClose?.()
        emit('close')
    }

    return {
        // state
        loading,
        error,
        questions,
        currentQuestionIndex,
        answers,
        selectedAnswer,
        isCompleted,
        submitting,
        // computed
        currentQuestion,
        // auth
        getUserIdFromToken,
        isAuthenticated,
        // navigation
        selectAnswer,
        previousQuestion,
        nextQuestion,
        saveCurrentAnswer,
        loadCurrentAnswer,
        // async
        checkExistingResponse,
        loadQuestions,
        submitTest,
        closeModal
    }
}
