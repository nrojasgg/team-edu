<template>
  <div v-if="isOpen && questionnaire" class="selection-modal-overlay" @click.self="$emit('close')">
    <div class="selection-modal">
      <div class="selection-modal-header">
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

      <div class="selection-modal-body">
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
          </div>

          <div class="students-grid">
            <div 
              v-for="student in otherStudents" 
              :key="student.id || student.student_id"
              class="student-card"
              :class="{ 
                'selected-positive': answers[currentQuestion.id]?.[student.id] === 1,
                'selected-negative': answers[currentQuestion.id]?.[student.id] === -1,
                'selected-neutral': answers[currentQuestion.id]?.[student.id] === 0
              }"
              @click="selectStudent(student)"
            >
              <span class="student-name">{{ student.name || student.student_name || `Estudiante ${student.id || student.student_id}` }}</span>
              <span class="student-selection">
                {{ getSelectionLabel(answers[currentQuestion.id]?.[student.id]) }}
              </span>
            </div>
          </div>

          <div class="quick-select">
            <button @click="selectAllPositive" class="btn-quick positive">+1 Todos</button>
            <button @click="selectAllNeutral" class="btn-quick neutral">0 Todos</button>
            <button @click="selectAllNegative" class="btn-quick negative">-1 Todos</button>
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
              :disabled="!isQuestionAnswered"
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
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { usePluginApi } from '~/composables/usePluginApi'
import { useSectionGroups } from '~/composables/useSectionGroups'

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
const groupsModule = useSectionGroups()

const loading = ref(false)
const error = ref('')
const questions = ref([])
const currentIndex = ref(0)
const answers = ref({})
const isCompleted = ref(false)
const submitting = ref(false)
const otherStudents = ref([])

const currentQuestion = computed(() => questions.value[currentIndex.value] || null)

const progressPercentage = computed(() => {
  if (questions.value.length === 0) return 0
  return ((currentIndex.value + 1) / questions.value.length) * 100
})

const isQuestionAnswered = computed(() => {
  if (!currentQuestion.value) return false
  const questionAnswers = answers.value[currentQuestion.value.id]
  return questionAnswers && Object.keys(questionAnswers).length > 0
})

const isComplete = computed(() => {
  return questions.value.every(q => {
    const qAnswers = answers.value[q.id]
    return qAnswers && Object.keys(qAnswers).length > 0
  })
})

const getSelectionLabel = (value) => {
  if (value === 1) return '+1'
  if (value === -1) return '-1'
  return '0'
}

const selectStudent = (student) => {
  if (!currentQuestion.value) return
  
  const studentId = student.id || student.student_id
  if (!answers.value[currentQuestion.value.id]) {
    answers.value[currentQuestion.value.id] = {}
  }
  
  const currentValue = answers.value[currentQuestion.value.id][studentId] || 0
  // Cycle: 0 -> 1 -> -1 -> 0
  const newValue = currentValue === 0 ? 1 : currentValue === 1 ? -1 : 0
  answers.value[currentQuestion.value.id][studentId] = newValue
}

const selectAllPositive = () => {
  if (!currentQuestion.value) return
  answers.value[currentQuestion.value.id] = {}
  otherStudents.value.forEach(s => {
    const sid = s.id || s.student_id
    answers.value[currentQuestion.value.id][sid] = 1
  })
}

const selectAllNeutral = () => {
  if (!currentQuestion.value) return
  answers.value[currentQuestion.value.id] = {}
  otherStudents.value.forEach(s => {
    const sid = s.id || s.student_id
    answers.value[currentQuestion.value.id][sid] = 0
  })
}

const selectAllNegative = () => {
  if (!currentQuestion.value) return
  answers.value[currentQuestion.value.id] = {}
  otherStudents.value.forEach(s => {
    const sid = s.id || s.student_id
    answers.value[currentQuestion.value.id][sid] = -1
  })
}

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

    // Load students in the section
    const allStudents = await groupsModule.loadSectionStudents(props.sectionId)
    
    // Filter out current student
    otherStudents.value = allStudents.filter(s => {
      const sid = s.id?.toString() || s.student_id?.toString()
      return sid !== props.studentId?.toString()
    })

    // Extract questions from questionnaire metadata
    if (props.questionnaire.questions && Array.isArray(props.questionnaire.questions)) {
      questions.value = props.questionnaire.questions
    } else {
      throw new Error('No se encontraron preguntas en el cuestionario')
    }

    // Initialize answers object
    questions.value.forEach(q => {
      answers.value[q.id] = {}
    })

  } catch (err) {
    console.error('[SelectionTest] Error loading:', err)
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const submitTest = async () => {
  if (!isComplete.value || !props.questionnaire?.id) return

  try {
    submitting.value = true

    // Transform answers to expected format: { question1: [{student_id, response}, ...] }
    const submissionData = {}
    Object.entries(answers.value).forEach(([questionId, studentResponses]) => {
      submissionData[`question${questionId}`] = Object.entries(studentResponses).map(([studentId, response]) => ({
        student_id: studentId,
        response: response
      }))
    })

    console.log('[SelectionTest] Submitting:', submissionData)

    const { success, data, error: apiError } = await pluginApi.submitResponse(
      'group_compatibility',
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
    console.error('[SelectionTest] Error submitting:', err)
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
  otherStudents.value = []
}

watch(() => props.isOpen, (newValue) => {
  if (newValue) {
    resetState()
    loadQuestions()
  }
})
</script>

<style scoped>
.selection-modal-overlay {
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

.selection-modal {
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

.selection-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem 2rem;
  border-bottom: 1px solid #dfe4ea;
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: white;
}

.selection-modal-header h2 {
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

.selection-modal-body {
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
  border-top: 4px solid #3b82f6;
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
  background: linear-gradient(90deg, #3b82f6, #1d4ed8);
  transition: width 0.3s ease;
}

.question-text {
  text-align: center;
}

.question-text h3 {
  margin: 0;
  color: #1f2937;
  font-size: 1.25rem;
}

.students-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 0.75rem;
  max-height: 300px;
  overflow-y: auto;
}

.student-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0.75rem;
  background: #f9fafb;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.student-card:hover {
  transform: translateY(-2px);
  border-color: #d1d5db;
}

.student-card.selected-positive {
  background: #dcfce7;
  border-color: #10b981;
}

.student-card.selected-negative {
  background: #fee2e2;
  border-color: #ef4444;
}

.student-card.selected-neutral {
  background: #f3f4f6;
  border-color: #6b7280;
}

.student-name {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
  text-align: center;
  word-break: break-word;
}

.student-selection {
  font-size: 1rem;
  font-weight: 700;
  margin-top: 0.25rem;
  color: #6b7280;
}

.student-card.selected-positive .student-selection {
  color: #10b981;
}

.student-card.selected-negative .student-selection {
  color: #ef4444;
}

.student-card.selected-neutral .student-selection {
  color: #6b7280;
}

.quick-select {
  display: flex;
  justify-content: center;
  gap: 0.75rem;
}

.btn-quick {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-quick.positive {
  background: #dcfce7;
  color: #10b981;
}

.btn-quick.positive:hover {
  background: #bbf7d0;
}

.btn-quick.neutral {
  background: #f3f4f6;
  color: #6b7280;
}

.btn-quick.neutral:hover {
  background: #e5e7eb;
}

.btn-quick.negative {
  background: #fee2e2;
  color: #ef4444;
}

.btn-quick.negative:hover {
  background: #fecaca;
}

.navigation-buttons {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin-top: 1rem;
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
  .students-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .quick-select {
    flex-wrap: wrap;
  }
}
</style>
