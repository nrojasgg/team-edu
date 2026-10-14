<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="closeModal">
    <div class="modal-content">
      <div class="modal-header">
        <h2>Crear Código de Invitación</h2>
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
            <label for="expires-date">Fecha límite de expiración: *</label>
            <div class="datetime-inputs">
              <input
                id="expires-date"
                v-model="formData.expiresDate"
                type="date"
                required
                class="form-input"
                :min="minDate"
              />
              <input
                id="expires-time"
                v-model="formData.expiresTime"
                type="time"
                required
                class="form-input"
              />
            </div>
          </div>

          <div class="form-group">
            <label>Límite de usos:</label>
            <div class="radio-options">
              <label class="radio-label">
                <input
                  type="radio"
                  v-model="formData.hasUsageLimit"
                  :value="false"
                />
                <span class="radio-custom"></span>
                Sin límite
              </label>
              <label class="radio-label">
                <input
                  type="radio"
                  v-model="formData.hasUsageLimit"
                  :value="true"
                />
                <span class="radio-custom"></span>
                Limitar a
              </label>
              <input
                v-if="formData.hasUsageLimit"
                v-model.number="formData.maxUses"
                type="number"
                min="1"
                required
                class="form-input small-input"
                placeholder="Ej: 10"
              />
              <span v-if="formData.hasUsageLimit">usos</span>
            </div>
          </div>

          <div v-if="submitError" class="error-message">
            {{ submitError }}
          </div>

          <div class="modal-footer">
            <button type="button" @click="closeModal" class="btn btn-cancel">
              Cancelar
            </button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              {{ submitting ? 'Creando...' : 'Crear código' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'

const props = defineProps({
  isOpen: Boolean,
  sectionId: {
    type: [Number, String],
    required: true
  },
  submitting: Boolean
})

const emit = defineEmits(['close', 'submit'])

const API_BASE_URL = 'http://127.0.0.1:3000/api/v1'

const formData = ref({
  expiresDate: '',
  expiresTime: '23:59',
  hasUsageLimit: false,
  maxUses: null
})

const submitError = ref('')

const minDate = computed(() => {
  const today = new Date()
  return today.toISOString().split('T')[0]
})

const closeModal = () => {
  emit('close')
}

const handleSubmit = () => {
  submitError.value = ''

  if (!formData.value.expiresDate || !formData.value.expiresTime) {
    submitError.value = 'Por favor completa la fecha y hora de expiración'
    return
  }

  if (formData.value.hasUsageLimit && (!formData.value.maxUses || formData.value.maxUses < 1)) {
    submitError.value = 'Por favor ingresa un número válido de usos'
    return
  }

  const expiresAt = new Date(`${formData.value.expiresDate}T${formData.value.expiresTime}:00`)
  if (expiresAt <= new Date()) {
    submitError.value = 'La fecha de expiración debe ser futura'
    return
  }

  emit('submit', {
    section_id: props.sectionId,
    expires_at: expiresAt.toISOString(),
    max_uses: formData.value.hasUsageLimit ? formData.value.maxUses : null
  })
}

watch(() => props.isOpen, (isOpen) => {
  if (isOpen) {
    const defaultDate = new Date()
    defaultDate.setDate(defaultDate.getDate() + 7)
    formData.value = {
      expiresDate: defaultDate.toISOString().split('T')[0],
      expiresTime: '23:59',
      hasUsageLimit: false,
      maxUses: null
    }
    submitError.value = ''
  }
})
</script>

<style scoped>
.modal-overlay {
  position: fixed !important;
  top: 0 !important;
  left: 0 !important;
  right: 0 !important;
  bottom: 0 !important;
  background: rgba(0, 0, 0, 0.5) !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  z-index: 1200 !important;
  backdrop-filter: blur(4px) !important;
}

.modal-content {
  background: white;
  border-radius: 16px;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.15);
  width: 90vw;
  max-width: 480px;
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
  font-size: 1.35rem;
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

.datetime-inputs {
  display: flex;
  gap: 0.75rem;
}

.form-input {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 1px solid #cbd5e0;
  border-radius: 8px;
  font-size: 1rem;
  transition: all 0.2s;
  font-family: inherit;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.1);
}

.small-input {
  width: 100px;
}

.radio-options {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.radio-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  color: var(--color-dark, #394049);
  font-size: 0.95rem;
}

.radio-label input[type="radio"] {
  display: none;
}

.radio-custom {
  width: 18px;
  height: 18px;
  border: 2px solid #cbd5e0;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.radio-label input[type="radio"]:checked + .radio-custom {
  border-color: var(--color-primary, #00A499);
}

.radio-label input[type="radio"]:checked + .radio-custom::after {
  content: '';
  width: 10px;
  height: 10px;
  background: var(--color-primary, #00A499);
  border-radius: 50%;
}

.error-message {
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #991b1b;
  padding: 0.75rem 1rem;
  border-radius: 8px;
  margin-bottom: 1rem;
  font-size: 0.9rem;
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

@media (max-width: 640px) {
  .modal-content {
    width: 95vw;
    max-height: 95vh;
  }

  .modal-header,
  .modal-body {
    padding: 1rem 1.5rem;
  }

  .datetime-inputs {
    flex-direction: column;
  }

  .modal-footer {
    flex-direction: column;
  }

  .btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
