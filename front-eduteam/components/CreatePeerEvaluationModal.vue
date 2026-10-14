<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="closeModal">
    <div class="modal-content">
      <div class="modal-header">
        <h2>Nueva evaluación entre pares</h2>
        <button @click="closeModal" class="close-btn">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <div class="modal-body">
        <div class="survey-info">
          <div class="info-box">
            <h3>Información de la Evaluación</h3>
            <p>Esta evaluación permite que cada estudiante evalúe a todos los miembros de su grupo, incluyendo una autoevaluación. Las preguntas están predefinidas y cubren aspectos como tareas, participación, organización, compromiso y prosocialidad.</p>
          </div>
        </div>

        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="survey-title">Título *</label>
            <input
              id="survey-title"
              v-model="formData.title"
              type="text"
              required
              class="form-input"
              placeholder="Ej: Evaluación entre Pares - Grupo A"
            />
          </div>

          <div class="form-group">
            <label for="survey-description">Descripción</label>
            <textarea
              id="survey-description"
              v-model="formData.description"
              class="form-textarea"
              rows="4"
              placeholder="Describe el propósito de esta evaluación entre pares..."
            ></textarea>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="survey-start-date">Fecha de Inicio *</label>
              <DatetimeInput
                id="survey-start-date"
                v-model="formData.start_date"
                required
              />
            </div>

            <div class="form-group">
              <label for="survey-end-date">Fecha de Término *</label>
              <DatetimeInput
                id="survey-end-date"
                v-model="formData.end_date"
                required
              />
            </div>
          </div>

          <div class="questions-preview">
            <h4>Preguntas incluidas en la evaluación:</h4>
            <ul class="questions-list">
              <li v-for="question in questions" :key="question.id">
                <strong>{{ question.category }}:</strong> {{ question.text }}
              </li>
            </ul>
            <p class="questions-note">Cada pregunta se responde en una escala de 1 a 4 puntos.</p>
          </div>

          <div class="modal-footer">
            <button type="button" @click="closeModal" class="btn btn-cancel">
              Cancelar
            </button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              {{ submitting ? 'Creando...' : 'Crear Evaluación' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  isOpen: Boolean,
  submitting: Boolean
})

const emit = defineEmits(['close', 'submit'])

const formData = ref({
  title: '',
  description: '',
  start_date: '',
  end_date: ''
})

// Predefined questions for peer evaluation
const questions = ref([
  {
    id: 1,
    text: "Realiza las tareas que le son asignadas dentro del grupo en los plazos requeridos.",
    category: "Tareas"
  },
  {
    id: 2,
    text: "Participa de forma activa en los espacios de encuentro del equipo, compartiendo la información, conocimientos y las experiencias.",
    category: "Participación"
  },
  {
    id: 3,
    text: "Colabora en la definición, organización y distribución de las tareas del equipo.",
    category: "Organización"
  },
  {
    id: 4,
    text: "Se orienta a la consecución de acuerdos y objetivos comunes y se compromete con ellos.",
    category: "Compromiso"
  },
  {
    id: 5,
    text: "Toma en cuenta los puntos de vista de los demás y retroalimenta constructivamente.",
    category: "Prosocialidad"
  }
])

// Reset form when modal opens
watch(() => props.isOpen, (isOpen) => {
  if (isOpen) {
    // Set default dates: start now, end in 7 days
    const now = new Date()
    const endDate = new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000)

    formData.value = {
      title: '',
      description: '',
      start_date: now.toISOString(),
      end_date: endDate.toISOString()
    }
  }
})

const closeModal = () => {
  emit('close')
}

const handleSubmit = () => {
  // Validate dates
  if (new Date(formData.value.end_date) <= new Date(formData.value.start_date)) {
    alert('La fecha de término debe ser posterior a la fecha de inicio')
    return
  }

  emit('submit', {
    title: formData.value.title,
    description: formData.value.description,
    start_date: formData.value.start_date,
    end_date: formData.value.end_date
  })
}
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
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal-content {
  background: white;
  border-radius: 16px;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.15);
  width: 90vw;
  max-width: 700px;
  max-height: 90vh;
  overflow-y: auto;
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem 2rem;
  border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.modal-header h2 {
  margin: 0;
  color: var(--color-dark, #394049);
  font-size: 1.5rem;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 8px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  background: var(--color-background, #f4f6f9);
  color: var(--color-dark, #394049);
}

.modal-body {
  padding: 2rem;
}

.survey-info {
  margin-bottom: 2rem;
}

.info-box {
  background: var(--color-background, #f4f6f9);
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: 8px;
  padding: 1rem;
}

.info-box h3 {
  margin: 0 0 0.5rem 0;
  color: var(--color-dark, #394049);
  font-size: 1.1rem;
  font-weight: 600;
}

.info-box p {
  margin: 0;
  color: var(--color-text-secondary, #64748b);
  font-size: 0.9rem;
  line-height: 1.5;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: var(--color-dark, #394049);
  font-weight: 500;
  font-size: 0.95rem;
}

.form-input,
.form-textarea {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 1px solid #cbd5e0;
  border-radius: 8px;
  font-size: 1rem;
  transition: all 0.2s;
  font-family: inherit;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.1);
}

.form-textarea {
  resize: vertical;
  min-height: 100px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

@media (max-width: 640px) {
  .form-row {
    grid-template-columns: 1fr;
  }
}

.questions-preview {
  margin-top: 2rem;
  padding: 1.5rem;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e9ecef;
}

.questions-preview h4 {
  margin: 0 0 1rem 0;
  color: var(--color-dark, #394049);
  font-size: 1.1rem;
  font-weight: 600;
}

.questions-list {
  list-style: none;
  padding: 0;
  margin: 0 0 1rem 0;
}

.questions-list li {
  padding: 0.5rem 0;
  border-bottom: 1px solid #e9ecef;
  color: var(--color-text-secondary, #64748b);
  font-size: 0.9rem;
  line-height: 1.4;
}

.questions-list li:last-child {
  border-bottom: none;
}

.questions-note {
  margin: 0;
  color: #94a3b8;
  font-size: 0.85rem;
  font-style: italic;
}

.modal-footer {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  padding-top: 1.5rem;
  border-top: 1px solid var(--color-border, #dfe4ea);
  margin-top: 1rem;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-cancel {
  background: var(--color-background, #f4f6f9);
  color: var(--color-text-secondary, #64748b);
}

.btn-cancel:hover:not(:disabled) {
  background: var(--color-border, #dfe4ea);
}

.btn-primary {
  background: #006b64;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #008f85;
  box-shadow: 0 4px 12px rgba(0, 164, 153, 0.3);
}

/* Responsive Design */
@media (max-width: 640px) {
  .modal-content {
    width: 95vw;
    max-height: 95vh;
  }

  .modal-header,
  .modal-body {
    padding: 1rem 1.5rem;
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .modal-footer {
    flex-direction: column;
  }

  .btn {
    width: 100%;
    justify-content: center;
  }

  .questions-preview {
    padding: 1rem;
  }
}
</style>