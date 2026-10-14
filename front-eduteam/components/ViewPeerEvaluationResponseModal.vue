<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="closeModal">
    <div class="modal-content">
      <div class="modal-header">
        <div>
          <h2>Tu respuesta para {{ evaluatedName }}</h2>
          <p class="modal-subtitle">Responder por {{ evaluatorName }}</p>
        </div>
        <button @click="closeModal" class="close-btn">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <div class="modal-body">
        <div v-if="loading" class="loading-state">
          <div class="spinner"></div>
          <p>Cargando tu respuesta...</p>
        </div>

        <div v-else-if="error" class="error-state">
          <p>{{ error }}</p>
        </div>

        <div v-else-if="!response" class="no-data">
          <p>No se encontró tu respuesta para esta evaluación.</p>
        </div>

        <div v-else class="response-content">
          <div class="response-info">
            <div><strong>Fecha:</strong> {{ formatDate(response.submitted_at || response.created_at) }}</div>
            <div v-if="response.comment"><strong>Comentario:</strong> {{ response.comment }}</div>
            <div v-else><strong>Comentario:</strong> Sin comentario</div>
          </div>
          <div class="answers-list">
            <div v-for="ans in response.responses" :key="ans.question_id" class="ans-item">
              <div class="ans-question">{{ getQuestionText(ans.question_id) }}</div>
              <div class="ans-score">{{ ans.score }}/4</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useDatetimeInput } from '~/composables/useDatetimeInput'

const props = defineProps({
  isOpen: Boolean,
  response: Object,
  allStudents: Object
})

const emit = defineEmits(['close'])

const loading = ref(false)
const error = ref('')

const closeModal = () => {
  console.log('closeModal called')
  emit('close')
}

const { formatDatetimeDisplay } = useDatetimeInput()
const formatDate = (dateString) => formatDatetimeDisplay(dateString)

const getQuestionText = (questionId) => {
  const questions = [
    'Realiza las tareas que le son asignadas dentro del grupo en los plazos requeridos.',
    'Participa de forma activa en los espacios de encuentro del equipo, compartiendo la información, conocimientos y las experiencias.',
    'Colabora en la definición, organización y distribución de las tareas del equipo.',
    'Se orienta a la consecución de acuerdos y objetivos comunes y se compromete con ellos.',
    'Toma en cuenta los puntos de vista de los demás y retroalimenta constructivamente.'
  ]
  return questions[questionId - 1] || `Pregunta ${questionId}`
}

const evaluatorName = computed(() => {
  if (!props.response?.evaluator_id || !props.allStudents) return `ID: ${props.response?.evaluator_id || 'Desconocido'}`
  const ev = props.allStudents.value.find(s => String(s.id) === String(props.response.evaluator_id))
  return ev ? `${ev.nombres || ''} ${ev.paterno || ''}`.trim() : `ID: ${props.response.evaluator_id}`
})

const evaluatedName = computed(() => {
  if (!props.response?.evaluated_id || !props.allStudents) return `ID: ${props.response?.evaluated_id || 'Desconocido'}`
  const ev = props.allStudents.value.find(s => String(s.id) === String(props.response.evaluated_id))
  return ev ? `${ev.nombres || ''} ${ev.paterno || ''}`.trim() : `ID: ${props.response.evaluated_id}`
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 8px;
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 1.5rem;
  border-bottom: 1px solid var(--color-border, #dfe4ea);
  background: var(--color-background, #f4f6f9);
  border-radius: 8px 8px 0 0;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: #1e293b;
}

.modal-subtitle {
  font-size: 0.875rem;
  color: #64748b;
  margin-top: 0.25rem;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #64748b;
  padding: 0.25rem;
  border-radius: 4px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: var(--color-border, #dfe4ea);
  color: #334155;
}

.modal-body {
  padding: 1.5rem;
}

.loading-state, .error-state, .no-data {
  text-align: center;
  padding: 2rem;
}

.loading-state {
  color: #64748b;
}

.error-state {
  color: #dc2626;
}

.no-data {
  color: #64748b;
}

.response-info {
  background: var(--color-background, #f4f6f9);
  padding: 1rem;
  border-radius: 6px;
  margin-bottom: 1.5rem;
  border: 1px solid var(--color-border, #dfe4ea);
}

.response-info div {
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
}

.response-info div:last-child {
  margin-bottom: 0;
}

.response-info strong {
  color: #374151;
  font-weight: 600;
}

.answers-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.ans-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: white;
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: 6px;
  transition: all 0.2s;
}

.ans-item:hover {
  border-color: #cbd5e1;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.ans-question {
  flex: 1;
  font-size: 0.875rem;
  color: #374151;
  line-height: 1.4;
}

.ans-score {
  font-weight: 600;
  font-size: 0.875rem;
  color: #1e293b;
  background: #dbeafe;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  min-width: 3rem;
  text-align: center;
}

.spinner {
  width: 2rem;
  height: 2rem;
  border: 2px solid var(--color-border, #dfe4ea);
  border-top: 2px solid #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

@media (max-width: 640px) {
  .modal-content {
    margin: 1rem;
    max-height: calc(100vh - 2rem);
  }

  .modal-header, .modal-body {
    padding: 1rem;
  }

  .ans-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }

  .ans-score {
    align-self: flex-end;
  }
}
</style>

<style scoped>
.modal-overlay{position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,0.5);display:flex;align-items:center;justify-content:center;padding:1rem}
.modal-content{background:white;border-radius:8px;width:100%;max-width:640px;overflow:auto}
.modal-header{display:flex;justify-content:space-between;align-items:center;padding:1rem;border-bottom:1px solid var(--color-border, #dfe4ea)}
.modal-body{padding:1rem}
.ans-item{display:flex;justify-content:space-between;padding:1.25rem;background:linear-gradient(135deg,var(--color-background, #f4f6f9) 0%,#ffffff 100%);border:1px solid var(--color-border, #dfe4ea);border-radius:12px;transition:all 0.3s ease;box-shadow:0 1px 3px rgba(0,0,0,0.05);position:relative;overflow:hidden}
</style>
