<template>
  <div v-if="isOpen && questionnaire" class="form-modal-overlay" @click.self="$emit('close')">
    <div class="form-modal">
      <div class="form-modal-header">
        <div class="header-content">
          <h2>{{ questionnaire.title }}</h2>
          <p class="question-counter">Pregunta {{ currentIndex + 1 }} de {{ questions.length }}</p>
        </div>
        <button @click="$emit('close')" class="close-btn">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <div class="form-modal-body">
        <div v-if="loading" class="loading-state">
          <div class="spinner"></div>
          <p>Cargando...</p>
        </div>

        <div v-else-if="error" class="error-state">
          <p>{{ error }}</p>
          <button @click="$emit('close')" class="btn-secondary">Cerrar</button>
        </div>

        <div v-else-if="isCompleted" class="completed-state">
          <div class="success-icon">
            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
              <polyline points="22 4 12 14.01 9 11.01"/>
            </svg>
          </div>
          <h3>¡Respuestas guardadas!</h3>
          <p>Tus respuestas han sido registradas correctamente.</p>
          <button @click="$emit('close')" class="btn-primary">Cerrar</button>
        </div>

        <div v-else-if="currentQuestion" class="question-container">
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: progressPercentage + '%' }"></div>
          </div>

          <div class="question-text">
            <h3>{{ currentQuestion.text }}</h3>
            <span v-if="currentQuestion.required" class="required-badge">Obligatorio</span>
          </div>

          <div class="input-container">
            <!-- Number input -->
            <input
              v-if="currentQuestion.type === 'number'"
              type="number"
              v-model.number="answers[currentQuestion.field_id]"
              :placeholder="currentQuestion.placeholder || 'Ingresa un número'"
              class="form-input"
            />

            <!-- Text input -->
            <input
              v-else-if="currentQuestion.type === 'text' || currentQuestion.type === 'text_field'"
              type="text"
              v-model="answers[currentQuestion.field_id]"
              :placeholder="currentQuestion.placeholder || 'Ingresa tu respuesta'"
              class="form-input"
            />

            <!-- Date input -->
            <input
              v-else-if="currentQuestion.type === 'date'"
              type="date"
              v-model="answers[currentQuestion.field_id]"
              class="form-input"
            />

            <!-- Select dropdown -->
            <select
              v-else-if="currentQuestion.type === 'select' || currentQuestion.type === 'radio'"
              v-model="answers[currentQuestion.field_id]"
              class="form-input"
            >
              <option value="" disabled>Selecciona una opción</option>
              <option 
                v-for="option in currentQuestion.options" 
                :key="option.value || option"
                :value="option.value || option"
              >
                {{ option.label || option }}
              </option>
            </select>

            <!-- Textarea -->
            <textarea
              v-else-if="currentQuestion.type === 'textarea' || currentQuestion.type === 'paragraph'"
              v-model="answers[currentQuestion.field_id]"
              :placeholder="currentQuestion.placeholder || 'Escribe tu respuesta'"
              class="form-input textarea"
              rows="4"
            ></textarea>

            <!-- Default fallback -->
            <input
              v-else
              type="text"
              v-model="answers[currentQuestion.field_id]"
              :placeholder="currentQuestion.placeholder || 'Ingresa tu respuesta'"
              class="form-input"
            />
          </div>

          <div class="navigation-buttons">
            <button 
              @click="previousQuestion" 
              class="btn-secondary"
              :disabled="currentIndex === 0"
            >
              Anterior
            </button>
            <span class="question-indicator">{{ currentIndex + 1 }} / {{ questions.length }}</span>
            <button 
              v-if="currentIndex < questions.length - 1"
              @click="nextQuestion"
              class="btn-primary"
              :disabled="isCurrentQuestionEmpty"
            >
              Siguiente
            </button>
            <button 
              v-else
              @click="submitTest"
              class="btn-primary"
              :disabled="!isComplete || submitting"
            >
              {{ submitting ? 'Enviando...' : 'Enviar' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { usePluginApi } from '~/composables/usePluginApi'

const props = defineProps({
  isOpen: {
    type: Boolean,
    default: false
  },
  questionnaire: {
    type: Object,
    default: null
  },
  studentId: {
    type: String,
    default: ''
  },
  sectionId: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['close', 'submitted'])

const pluginApi = usePluginApi()

const loading = ref(false)
const error = ref('')
const questions = ref([])
const currentIndex = ref(0)
const answers = ref({})
const isCompleted = ref(false)
const submitting = ref(false)

const currentQuestion = computed(() => questions.value[currentIndex.value] || null)

const progressPercentage = computed(() => {
  if (questions.value.length === 0) return 0
  return ((currentIndex.value + 1) / questions.value.length) * 100
})

const isCurrentQuestionEmpty = computed(() => {
  if (!currentQuestion.value) return true
  if (!currentQuestion.value.required) return false
  const value = answers.value[currentQuestion.value.field_id]
  return value === undefined || value === null || value === ''
})

const isComplete = computed(() => {
  return questions.value.every(q => {
    if (!q.required) return true
    const value = answers.value[q.field_id]
    return value !== undefined && value !== null && value !== ''
  })
})

const nextQuestion = () => {
  if (currentIndex.value < questions.value.length - 1) {
    currentIndex.value++
  }
}

const previousQuestion = () => {
  if (currentIndex.value > 0) {
    currentIndex.value--
  }
}

const loadQuestions = async () => {
  try {
    loading.value = true
    error.value = ''

    if (!props.questionnaire) {
      throw new Error('No se ha seleccionado un cuestionario')
    }

    // Extract questions from questionnaire
    // Demographic questions can be in various formats
    let questionList = []
    
    if (props.questionnaire.questions && Array.isArray(props.questionnaire.questions)) {
      questionList = props.questionnaire.questions.map((q, idx) => ({
        field_id: q.field_id || q.id || `field_${idx}`,
        text: q.text || q.label || q.question || '',
        type: q.type || 'text',
        required: q.required ?? true,
        options: q.options || [],
        placeholder: q.placeholder || ''
      }))
    } else if (props.questionnaire.fields && Array.isArray(props.questionnaire.fields)) {
      questionList = props.questionnaire.fields.map((f, idx) => ({
        field_id: f.field_id || f.id || `field_${idx}`,
        text: f.text || f.label || f.name || '',
        type: f.type || 'text',
        required: f.required ?? true,
        options: f.options || [],
        placeholder: f.placeholder || ''
      }))
    }

    if (questionList.length === 0) {
      throw new Error('No se encontraron preguntas en el cuestionario')
    }

    questions.value = questionList

    // Initialize answers
    questions.value.forEach(q => {
      if (answers.value[q.field_id] === undefined) {
        answers.value[q.field_id] = ''
      }
    })

  } catch (err) {
    console.error('[FormTest] Error loading:', err)
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const submitTest = async () => {
  if (!isComplete.value || !props.questionnaire?.id) return

  try {
    submitting.value = true

    // Transform answers to expected format: { answers: [{field_id, numeric_value/choice_value}] }
    const answersList = Object.entries(answers.value).map(([field_id, value]) => {
      const answer = { field_id }
      
      // Determine if numeric or choice
      if (typeof value === 'number') {
        answer.numeric_value = value
      } else {
        answer.choice_value = value
      }
      
      return answer
    })

    const submissionData = { answers: answersList }

    console.log('[FormTest] Submitting:', submissionData)

    const { success, data, error: apiError } = await pluginApi.submitResponse(
      'demographic',
      props.questionnaire.id,
      submissionData,
      props.studentId
    )

    if (!success) {
      throw new Error(apiError || 'Error al enviar las respuestas')
    }

    isCompleted.value = true
    emit('submitted', data)
  } catch (err) {
    console.error('[FormTest] Error submitting:', err)
    error.value = err.message
  } finally {
    submitting.value = false
  }
}

const resetState = () => {
  currentIndex.value = 0
  answers.value = {}
  questions.value = []
  loading.value = false
  error.value = ''
  isCompleted.value = false
  submitting.value = false
}

watch(() => props.isOpen, (newValue) => {
  if (newValue) {
    resetState()
    loadQuestions()
  }
})
</script>

<style scoped>
.form-modal-overlay {
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

.form-modal {
  background: white;
  border-radius: 16px;
  max-width: 600px;
  width: 100%;
  max-height: 90vh;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
}

.form-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem 2rem;
  border-bottom: 1px solid #dfe4ea;
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
  color: white;
}

.form-modal-header h2 {
  margin: 0;
  font-size: 1.25rem;
}

.question-counter {
  margin: 0.25rem 0 0 0;
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
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.form-modal-body {
  padding: 1.5rem;
  overflow-y: auto;
  flex: 1;
}

.loading-state, .error-state, .completed-state {
  text-align: center;
  padding: 3rem;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #8b5cf6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.success-icon {
  color: #10b981;
  margin-bottom: 1rem;
}

.completed-state h3 {
  margin: 0 0 0.5rem 0;
  color: #065f46;
}

.completed-state p {
  margin: 0 0 1.5rem 0;
  color: #047857;
}

.question-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.progress-bar {
  height: 4px;
  background: #e5e7eb;
  border-radius: 2px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #8b5cf6, #7c3aed);
  transition: width 0.3s ease;
}

.question-text {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.question-text h3 {
  margin: 0;
  color: #1f2937;
  font-size: 1.125rem;
}

.required-badge {
  font-size: 0.75rem;
  padding: 0.25rem 0.5rem;
  background: #fee2e2;
  color: #dc2626;
  border-radius: 4px;
  font-weight: 500;
}

.input-container {
  margin-top: 0.5rem;
}

.form-input {
  width: 100%;
  padding: 0.875rem 1rem;
  font-size: 1rem;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: #8b5cf6;
}

.form-input.textarea {
  resize: vertical;
  min-height: 100px;
}

select.form-input {
  cursor: pointer;
  background: white;
}

.navigation-buttons {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin-top: 1.5rem;
}

.question-indicator {
  font-weight: 500;
  color: #6b7280;
}

.btn-primary {
  padding: 0.75rem 1.5rem;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  padding: 0.75rem 1.5rem;
  background: #e5e7eb;
  color: #374151;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
}

.btn-secondary:hover:not(:disabled) {
  background: #d1d5db;
}

.btn-secondary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 480px) {
  .form-modal-header {
    padding: 1rem 1.25rem;
  }
  
  .form-modal-body {
    padding: 1rem;
  }
}
</style>
