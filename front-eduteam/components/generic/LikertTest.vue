<template>
  <div v-if="isOpen && questionnaire" ref="modalOverlay" class="likert-modal-overlay" @click="closeModal" @keydown="handleKeydown" tabindex="0">
    <div class="likert-modal" @click.stop>
      <div class="likert-modal-header">
        <div class="header-content">
          <h2>{{ questionnaire.title }}</h2>
          <p v-if="questionnaire.description" class="subtitle">{{ questionnaire.description }}</p>
        </div>
        <button @click="closeModal" class="close-btn">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <div class="likert-modal-body">
        <div v-if="loading" class="loading-state">
          <div class="spinner"></div>
          <p>Cargando cuestionario...</p>
        </div>

        <div v-else-if="error" class="error-state">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="8" x2="12" y2="12"/>
            <line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
          <p>{{ error }}</p>
          <button @click="closeModal" class="btn-secondary">Cerrar</button>
        </div>

        <div v-else-if="isCompleted" class="completed-state">
          <div class="success-icon">
            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
              <polyline points="22 4 12 14.01 9 11.01"/>
            </svg>
          </div>
          <h3>Test Completado</h3>
          <p class="completed-message">Tus respuestas han sido guardadas correctamente.</p>
          
          <div v-if="resultData" class="results-summary">
            <p class="result-detail">{{ resultMessage }}</p>
          </div>

          <button @click="$emit('view-results', resultData)" class="btn-secondary">
            Ver Resultados Detallados
          </button>
          <button @click="closeModal" class="btn-primary">Cerrar</button>
        </div>

        <div v-else-if="questions.length > 0" class="questions-container">
          <div class="progress-indicator">
            <span>Pregunta {{ currentIndex + 1 }} de {{ questions.length }}</span>
            <div class="progress-bar">
              <div
                class="progress-fill"
                :style="{ width: `${((currentIndex + 1) / questions.length) * 100}%` }"
              ></div>
            </div>
          </div>

          <div class="question-card">
            <p class="question-text">{{ currentQuestion.text }}</p>
            <p class="keyboard-hint">Usa los números 1-5 para seleccionar</p>

            <div class="likert-scale">
              <div class="likert-options">
                <div
                  v-for="(option, index) in currentQuestion.options"
                  :key="option.score"
                  class="likert-option"
                  :class="{ 
                    'selected': answers[currentQuestion.id] === option.score,
                    'focused': focusedOptionIndex === index
                  }"
                  @click="selectAnswer(option.score)"
                >
                  <div class="likert-circle">
                    <span class="likert-value">{{ option.score }}</span>
                  </div>
                  <span class="likert-label">{{ option.text }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="navigation-buttons">
            <button
              v-if="currentIndex > 0"
              @click="previousQuestion"
              class="btn-secondary"
            >
              Anterior
            </button>

            <div class="spacer"></div>

            <button
              v-if="currentIndex < questions.length - 1"
              @click="nextQuestion"
              class="btn-primary"
              :disabled="answers[currentQuestion.id] === undefined"
            >
              Siguiente
            </button>

            <button
              v-else
              @click="submitTest"
              class="btn-primary"
              :disabled="!isComplete || submitting"
            >
              {{ submitting ? 'Enviando...' : 'Finalizar Test' }}
            </button>
          </div>
        </div>

        <div v-else class="no-questions">
          <p>No se encontraron preguntas para este test.</p>
          <button @click="closeModal" class="btn-secondary">Cerrar</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { usePluginApi } from '~/composables/usePluginApi'

const props = defineProps({
  isOpen: {
    type: Boolean,
    default: false
  },
  questionnaire: {
    type: Object,
    required: true
  },
  metadata: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['close', 'submitted', 'view-results'])

const pluginApi = usePluginApi()
const loading = ref(false)
const error = ref('')
const questions = ref([])
const currentIndex = ref(0)
const answers = ref({})
const isCompleted = ref(false)
const submitting = ref(false)
const resultData = ref(null)
const focusedOptionIndex = ref(-1)
const modalOverlay = ref(null)

const currentQuestion = computed(() => questions.value[currentIndex.value] || {})

const isComplete = computed(() => {
  return questions.value.every(q => answers.value[q.id] !== undefined)
})

const resultMessage = computed(() => {
  if (!resultData.value) return ''
  if (resultData.value.message) return resultData.value.message
  if (resultData.value.dominant_type) return `Tu tipo dominante es: ${resultData.value.dominant_type}`
  return 'Test completado exitosamente'
})

// Also try to load questions on mount (in case watch missed the initial change)
onMounted(() => {
  console.log('[LikertTest onMounted] Component mounted, isOpen:', props.isOpen, {
    hasQuestionnaire: !!props.questionnaire,
    hasMetadata: !!props.metadata,
    metadataFullQuestions: props.metadata?.full_questions?.length
  });
  if (props.isOpen && props.questionnaire && props.metadata) {
    loadQuestions()
  }
})

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

const selectAnswer = (score) => {
  answers.value[currentQuestion.value.id] = score
}

const previousQuestion = () => {
  if (currentIndex.value > 0) {
    currentIndex.value--
  }
}

const nextQuestion = () => {
  if (currentIndex.value < questions.value.length - 1 && answers.value[currentQuestion.value.id] !== undefined) {
    currentIndex.value++
  }
}

const selectByIndex = (index) => {
  const options = currentQuestion.value.options
  if (options && options[index]) {
    selectAnswer(options[index].score)
  }
}

const handleKeydown = (event) => {
  const options = currentQuestion.value.options
  if (!options) return

  // Number keys 1-5 to select option
  if (event.key >= '1' && event.key <= '5') {
    const index = parseInt(event.key) - 1
    if (index < options.length) {
      selectByIndex(index)
      focusedOptionIndex.value = index
      event.preventDefault()
    }
    return
  }

  // Arrow left/right to navigate options
  if (event.key === 'ArrowLeft') {
    if (focusedOptionIndex.value > 0) {
      focusedOptionIndex.value--
    } else if (focusedOptionIndex.value === -1 && options.length > 0) {
      focusedOptionIndex.value = 0
    }
    event.preventDefault()
    return
  }

  if (event.key === 'ArrowRight') {
    if (focusedOptionIndex.value < options.length - 1) {
      focusedOptionIndex.value++
    }
    event.preventDefault()
    return
  }

  // Enter to go next or submit
  if (event.key === 'Enter') {
    if (currentIndex.value < questions.value.length - 1) {
      if (answers.value[currentQuestion.value.id] !== undefined) {
        nextQuestion()
        focusedOptionIndex.value = -1
      }
    } else {
      if (isComplete.value) {
        submitTest()
      }
    }
    event.preventDefault()
    return
  }

  // Backspace or Left arrow on first question could go back
  if (event.key === 'Backspace' && currentIndex.value > 0) {
    previousQuestion()
    focusedOptionIndex.value = -1
    event.preventDefault()
  }
}

const getPluginName = () => props.questionnaire?.plugin?.toLowerCase() || ''

const checkExistingResponse = async () => {
  if (!props.questionnaire?.id) return false

  if (!isAuthenticated()) {
    console.warn('Usuario no autenticado')
    return false
  }

  try {
    const studentId = getUserIdFromToken()
    if (!studentId) return false

    const pluginName = getPluginName()
    const sectionId = props.questionnaire.section_id

    // Get all responses for the section and find this student's response for this test
    const { success, data } = await pluginApi.getSectionResponses(pluginName, sectionId)

    if (success && Array.isArray(data)) {
      const studentResponse = data.find(r =>
        r.student_id?.toString() === studentId.toString() &&
        r.test_id === props.questionnaire.id
      )

      if (studentResponse?.is_complete) {
        isCompleted.value = true
        resultData.value = studentResponse.results || studentResponse
        console.log('Usuario ya completó el test')
        return true
      }
    }
  } catch (err) {
    console.error('Error verificando respuesta existente:', err)
  }

  return false
}

const loadQuestions = async () => {
  try {
    loading.value = true
    error.value = ''

    if (!props.questionnaire) {
      throw new Error('No se ha seleccionado un cuestionario')
    }

    if (!isAuthenticated()) {
      throw new Error('Debes estar autenticado para acceder al test')
    }

    if (!props.metadata) {
      throw new Error('Metadatos del plugin no disponibles')
    }

    const alreadyCompleted = await checkExistingResponse()
    if (alreadyCompleted) {
      loading.value = false
      return
    }

    // Try full_questions first (array of objects with options)
    let rawQuestions = props.metadata?.full_questions;
    
    // If full_questions is not available or empty, try preguntas (map of id -> text)
    if (!rawQuestions || rawQuestions.length === 0) {
      rawQuestions = props.metadata?.preguntas;
    }
    
    if (!rawQuestions || rawQuestions.length === 0) {
      console.error('[loadQuestions] No questions found in metadata:', {
        full_questions: props.metadata?.full_questions,
        preguntas: props.metadata?.preguntas
      });
      throw new Error('No se encontraron preguntas en los metadatos')
    }

    // Handle preguntas (map format) - convert to array format
    let questionsData;
    if (typeof rawQuestions === 'object' && !Array.isArray(rawQuestions)) {
      // rawQuestions is a map like {"1": "question text", ...}
      // We need to convert it to array format but we don't have options
      // This shouldn't happen if full_questions is properly populated
      console.warn('[loadQuestions] preguntas is a map, not an array. This plugin may not support Likert tests properly.');
      throw new Error('El formato de preguntas no es compatible con este componente')
    } else {
      // rawQuestions is an array
      questionsData = rawQuestions.map(q => ({
        ...q,
        options: (q.options || []).map(opt => ({
          score: opt.score ?? opt.Score,
          text: opt.text ?? opt.Text
        }))
      }));
    }

    if (questionsData.length > 0) {
      questions.value = questionsData
      console.log('[loadQuestions] Preguntas cargadas:', questions.value.length, questionsData[0])
    } else {
      throw new Error('No se encontraron preguntas en los metadatos')
    }
  } catch (err) {
    console.error('[loadQuestions] Error:', err)
    error.value = err.message
    questions.value = []
  } finally {
    loading.value = false
  }
}

const submitTest = async () => {
  if (!isComplete.value || !props.questionnaire?.id) return

  if (!isAuthenticated()) {
    error.value = 'Debes estar autenticado para enviar el test'
    return
  }

  try {
    submitting.value = true
    const studentId = getUserIdFromToken()

    if (!studentId) {
      throw new Error('No se pudo obtener el ID del usuario')
    }

    const answersArray = Object.entries(answers.value).map(([questionId, score]) => ({
      question_id: Number.parseInt(questionId),
      score: score
    }))

    const pluginName = getPluginName()
    const { success, data, error: apiError } = await pluginApi.submitResponse(
      pluginName,
      props.questionnaire.id,
      answersArray,
      studentId
    )

    if (!success) {
      throw new Error(apiError || 'Error al enviar las respuestas')
    }

    console.log('Respuesta del servidor:', data)

    resultData.value = data?.results || data
    isCompleted.value = true
    emit('submitted', data)
  } catch (err) {
    console.error('Error enviando test:', err)
    error.value = err.message
  } finally {
    submitting.value = false
  }
}

const closeModal = () => {
  currentIndex.value = 0
  answers.value = {}
  error.value = ''
  questions.value = []
  loading.value = false
  submitting.value = false
  resultData.value = null
  focusedOptionIndex.value = -1

  emit('close')
}

watch(() => props.isOpen, (newValue) => {
  console.log('[LikertTest watch] props.isOpen changed to:', newValue, {
    hasQuestionnaire: !!props.questionnaire,
    hasMetadata: !!props.metadata,
    metadataFullQuestions: props.metadata?.full_questions?.length
  });
  if (newValue && props.questionnaire && props.metadata) {
    loadQuestions()
    nextTick(() => {
      modalOverlay.value?.focus()
    })
  }
}, { immediate: true })
</script>

<style scoped>
.likert-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.likert-modal {
  background: white;
  border-radius: 16px;
  max-width: 800px;
  width: 100%;
  max-height: 90vh;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
}

.likert-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem 2rem;
  border-bottom: 1px solid var(--color-border, #dfe4ea);
  background: linear-gradient(135deg, #4338ca 0%, #6d28d9 100%);
  color: white;
}

.header-content {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.likert-modal-header h2 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
}

.subtitle {
  margin: 0;
  font-size: 0.875rem;
  opacity: 0.9;
}

.close-btn {
  background: none;
  border: none;
  color: white;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 8px;
  transition: background 0.2s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.likert-modal-body {
  padding: 2rem;
  overflow-y: auto;
  flex: 1;
}

.loading-state, .error-state, .no-questions {
  text-align: center;
  padding: 3rem;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error-state svg {
  color: #ef4444;
  margin-bottom: 1rem;
}

.error-state p {
  color: #6b7280;
  margin-bottom: 1.5rem;
}

.success-icon {
  margin: 0 auto 1rem;
  color: #10b981;
}

.completed-state {
  text-align: center;
  padding: 3rem 2rem;
}

.completed-state h3 {
  color: #1f2937;
  margin-bottom: 0.5rem;
  font-size: 1.5rem;
}

.completed-message {
  color: #6b7280;
  margin-bottom: 2rem;
}

.results-summary {
  background: #f9fafb;
  border-radius: 12px;
  padding: 1.5rem;
  margin-bottom: 2rem;
  max-width: 500px;
  margin-left: auto;
  margin-right: auto;
}

.result-detail {
  color: #374151;
  font-size: 1rem;
  margin: 0;
}

.progress-indicator {
  margin-bottom: 2rem;
}

.progress-indicator span {
  display: block;
  text-align: center;
  margin-bottom: 0.5rem;
  color: #6b7280;
  font-weight: 500;
}

.progress-bar {
  width: 100%;
  height: 8px;
  background: var(--color-border, #dfe4ea);
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #667eea, #764ba2);
  transition: width 0.3s ease;
}

.question-card {
  background: #f9fafb;
  border-radius: 12px;
  padding: 2rem;
  margin-bottom: 2rem;
}

.question-text {
  color: #1f2937;
  font-size: 1.25rem;
  font-weight: 500;
  margin: 0 0 0.5rem 0;
  text-align: center;
}

.keyboard-hint {
  color: #9ca3af;
  font-size: 0.75rem;
  margin: 0 0 1.5rem 0;
  text-align: center;
}

.likert-scale {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.likert-options {
  display: flex;
  justify-content: space-between;
  gap: 0.5rem;
}

.likert-option {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  transition: transform 0.2s;
}

.likert-option:hover {
  transform: translateY(-2px);
}

.likert-option.focused .likert-circle {
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.3);
}

.likert-option.selected .likert-circle {
  background: linear-gradient(135deg, #4338ca 0%, #6d28d9 100%);
  color: white;
  transform: scale(1.1);
  box-shadow: 0 4px 12px rgba(67, 56, 202, 0.4);
}

.likert-circle {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  background: white;
  border: 2px solid var(--color-border, #dfe4ea);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.likert-value {
  font-size: 1.25rem;
  font-weight: 600;
}

.likert-label {
  font-size: 0.75rem;
  color: #6b7280;
  text-align: center;
  line-height: 1.2;
}

.navigation-buttons {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
}

.spacer {
  flex: 1;
}

.btn-primary, .btn-secondary {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
  font-size: 1rem;
}

.btn-primary {
  background: linear-gradient(135deg, #4338ca 0%, #6d28d9 100%);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(67, 56, 202, 0.4);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: #e5e7eb;
  color: #374151;
}

.btn-secondary:hover {
  background: #d1d5db;
}

@media (max-width: 640px) {
  .likert-options {
    flex-direction: column;
  }
  
  .likert-option {
    flex-direction: row;
    justify-content: flex-start;
  }
  
  .likert-label {
    text-align: left;
  }
}
</style>
