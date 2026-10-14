<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="closeModal">
    <div class="modal-content">
      <div class="modal-header">
        <h2>Editar Cuestionario</h2>
        <button @click="closeModal" class="close-btn">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <div class="modal-body">
        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="edit-title">Título *</label>
            <input
              id="edit-title"
              v-model="formData.title"
              type="text"
              required
              class="form-input"
              placeholder="Ej: Test de Eneagrama - Primer Semestre"
            />
          </div>

          <div class="form-group">
            <label for="edit-description">Descripción</label>
            <textarea
              id="edit-description"
              v-model="formData.description"
              class="form-textarea"
              rows="4"
              placeholder="Describe el propósito y objetivos del cuestionario..."
            ></textarea>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="edit-start-date">Fecha de Inicio *</label>
              <DatetimeInput
                id="edit-start-date"
                v-model="formData.start_date"
                required
              />
            </div>

            <div class="form-group">
              <label for="edit-end-date">Fecha de Término *</label>
              <DatetimeInput
                id="edit-end-date"
                v-model="formData.end_date"
                required
              />
            </div>
          </div>

          <div class="modal-footer">
            <button type="button" @click="closeModal" class="btn btn-cancel">
              Cancelar
            </button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              {{ submitting ? 'Guardando...' : 'Guardar Cambios' }}
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
  questionnaire: Object
})

const emit = defineEmits(['close', 'save'])

const formData = ref({
  title: '',
  description: '',
  start_date: '',
  end_date: ''
})

const submitting = ref(false)

// Watch for questionnaire changes to populate form
watch(() => props.questionnaire, (newQuestionnaire) => {
  if (newQuestionnaire) {
    formData.value = {
      title: newQuestionnaire.title || '',
      description: newQuestionnaire.description || '',
      start_date: newQuestionnaire.start_date || '',
      end_date: newQuestionnaire.end_date || ''
    }
  }
}, { immediate: true })

const closeModal = () => {
  emit('close')
}

const handleSubmit = async () => {
  if (submitting.value) return
  
  // Validate dates
  if (new Date(formData.value.end_date) <= new Date(formData.value.start_date)) {
    alert('La fecha de término debe ser posterior a la fecha de inicio')
    return
  }
  
  submitting.value = true
  
  try {
    await emit('save', {
      id: props.questionnaire.id || props.questionnaire._id,
      ...formData.value
    })
  } finally {
    submitting.value = false
  }
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
  padding: 1rem;
}

.modal-content {
  background: white;
  border-radius: 12px;
  max-width: 600px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.modal-header h2 {
  margin: 0;
  color: var(--color-dark, #394049);
  font-size: 1.5rem;
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: #94a3b8;
  padding: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: var(--color-background, #f4f6f9);
  color: var(--color-dark, #394049);
}

.modal-body {
  padding: 1.5rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: var(--color-dark, #394049);
  font-weight: 500;
  font-size: 0.875rem;
}

.form-input,
.form-textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: 6px;
  font-size: 0.875rem;
  transition: all 0.2s;
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
  font-family: inherit;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.modal-footer {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  padding-top: 1.5rem;
  border-top: 1px solid var(--color-border, #dfe4ea);
}

.btn {
  padding: 0.75rem 1.5rem;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.btn-cancel {
  background: var(--color-background, #f4f6f9);
  color: var(--color-dark, #394049);
}

.btn-cancel:hover {
  background: var(--color-border, #dfe4ea);
}

.btn-primary {
  background: var(--color-primary, #006b64);
  color: white;
}

.btn-primary:hover {
  background: var(--color-primary-dark, #008a7f);
}

.btn-primary:disabled {
  background: #cbd5e0;
  cursor: not-allowed;
}

@media (max-width: 640px) {
  .form-row {
    grid-template-columns: 1fr;
  }
  
  .modal-content {
    max-height: 95vh;
  }
  
  .modal-footer {
    flex-direction: column-reverse;
  }
  
  .btn {
    width: 100%;
  }
}
</style>
