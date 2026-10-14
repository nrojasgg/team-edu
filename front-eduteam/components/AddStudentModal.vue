<template>
  <div v-if="isOpen" class="modal-overlay" @click="closeModal">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>Agregar Estudiante</h3>
        <button @click="closeModal" class="close-btn">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="m18 6-12 12"/>
            <path d="m6 6 12 12"/>
          </svg>
        </button>
      </div>
      
      <form @submit.prevent="submitForm" class="modal-body">
        <div class="form-group">
          <label for="rut">RUT *</label>
          <input
            id="rut"
            v-model="formData.rut"
            type="text"
            placeholder="Ej: 12.345.678-9"
            required
            class="form-input"
            :class="{ 'error': errors.rut }"
          />
          <span v-if="errors.rut" class="error-message">{{ errors.rut }}</span>
        </div>
        
        <div class="form-group">
          <label for="nombres">Nombres *</label>
          <input
            id="nombres"
            v-model="formData.nombres"
            type="text"
            placeholder="Ej: Juan Carlos"
            required
            class="form-input"
            :class="{ 'error': errors.nombres }"
          />
          <span v-if="errors.nombres" class="error-message">{{ errors.nombres }}</span>
        </div>
        
        <div class="form-row">
          <div class="form-group">
            <label for="paterno">Apellido Paterno *</label>
            <input
              id="paterno"
              v-model="formData.paterno"
              type="text"
              placeholder="Ej: Pérez"
              required
              class="form-input"
              :class="{ 'error': errors.paterno }"
            />
            <span v-if="errors.paterno" class="error-message">{{ errors.paterno }}</span>
          </div>
          
          <div class="form-group">
            <label for="materno">Apellido Materno</label>
            <input
              id="materno"
              v-model="formData.materno"
              type="text"
              placeholder="Ej: González"
              class="form-input"
              :class="{ 'error': errors.materno }"
            />
            <span v-if="errors.materno" class="error-message">{{ errors.materno }}</span>
          </div>
        </div>
        
        <div class="form-group">
          <label for="carrera">Código de Carrera</label>
          <input
            id="carrera"
            v-model="formData.carrera"
            type="text"
            placeholder="Ej: 1883, 1353, etc."
            class="form-input"
            :class="{ 'error': errors.carrera }"
          />
          <span v-if="errors.carrera" class="error-message">{{ errors.carrera }}</span>
        </div>
        
        <div class="form-group">
          <label for="email">Correo Electrónico *</label>
          <input
            id="email"
            v-model="formData.email"
            type="email"
            placeholder="Ej: juan.perez@usach.cl"
            required
            class="form-input"
            :class="{ 'error': errors.email }"
          />
          <span v-if="errors.email" class="error-message">{{ errors.email }}</span>
        </div>
        
        <div class="modal-actions">
          <button type="button" @click="closeModal" class="cancel-btn">
            Cancelar
          </button>
          <button type="submit" :disabled="loading" class="submit-btn">
            <span v-if="loading">Agregando...</span>
            <span v-else>Agregar Estudiante</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { isValidEmail, validateRUT, addDotsToRUT } from '~/composables/useValidation'

const props = defineProps({
  isOpen: Boolean
})

const emit = defineEmits(['close', 'studentAdded'])

const loading = ref(false)
const formData = reactive({
  rut: '',
  nombres: '',
  paterno: '',
  materno: '',
  carrera: '',
  email: ''
})

const errors = reactive({
  rut: '',
  nombres: '',
  paterno: '',
  materno: '',
  carrera: '',
  email: ''
})

// Función para limpiar errores
const clearErrors = () => {
  errors.rut = ''
  errors.nombres = ''
  errors.paterno = ''
  errors.materno = ''
  errors.carrera = ''
  errors.email = ''
}

// RUT validation provided by composable `useValidation`

// Función para validar el formulario
const validateForm = () => {
  clearErrors()
  let isValid = true
  
  // Validar RUT
  if (!formData.rut.trim()) {
    errors.rut = 'El RUT es obligatorio'
    isValid = false
  } else if (!validateRUT(formData.rut)) {
    errors.rut = 'El RUT ingresado no es válido'
    isValid = false
  }
  
  // Validar nombres
  if (!formData.nombres.trim()) {
    errors.nombres = 'Los nombres son obligatorios'
    isValid = false
  }
  
  // Validar apellido paterno
  if (!formData.paterno.trim()) {
    errors.paterno = 'El apellido paterno es obligatorio'
    isValid = false
  }
  
  // Validar email (sin regex vulnerable a ReDoS)
  if (!formData.email.trim()) {
    errors.email = 'El correo electrónico es obligatorio'
    isValid = false
  } else if (!isValidEmail(formData.email)) {
    errors.email = 'El formato del correo electrónico no es válido'
    isValid = false
  }
  
  return isValid
}

// Función para enviar el formulario
const submitForm = async () => {
  if (!validateForm()) return
  
  loading.value = true
  
  try {
    // Emitir evento con los datos del estudiante
    emit('studentAdded', {
      rut: formData.rut.trim(),
      nombres: formData.nombres.trim(),
      paterno: formData.paterno.trim(),
      materno: formData.materno.trim(),
      carrera: formData.carrera.trim() || null,
      email: formData.email.trim().toLowerCase()
    })
    
    // Cerrar modal después de emitir el evento
    closeModal()
  } catch (error) {
    console.error('Error al agregar estudiante:', error)
  } finally {
    loading.value = false
  }
}

// Función para cerrar el modal
const closeModal = () => {
  // Limpiar formulario
  formData.rut = ''
  formData.nombres = ''
  formData.paterno = ''
  formData.materno = ''
  formData.carrera = ''
  formData.email = ''
  clearErrors()
  
  emit('close')
}

// Email validation provided by composable `useValidation`

// addDotsToRUT provided by composable `useValidation`

// Formatear RUT mientras se escribe
watch(() => formData.rut, (newValue) => {
  let cleanValue = ''
  for (const ch of newValue) {
    if ((ch >= '0' && ch <= '9') || ch === 'k' || ch === 'K') {
      cleanValue += ch
    }
  }

  if (cleanValue.length > 1) {
    const rutBody = cleanValue.slice(0, -1)
    const rutDV = cleanValue.slice(-1)
    formData.rut = addDotsToRUT(rutBody) + '-' + rutDV
  }
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
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.modal-header h3 {
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
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: var(--color-background, #f4f6f9);
  color: var(--color-text-secondary, #64748b);
}

.modal-body {
  padding: 1.5rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-row {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.form-row .form-group {
  flex: 1;
  margin-bottom: 0;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: var(--color-dark, #394049);
  font-weight: 500;
}

.form-input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: 8px;
  font-size: 1rem;
  transition: all 0.2s;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.1);
}

.form-input.error {
  border-color: var(--color-danger, #d7373f);
}

.error-message {
  display: block;
  margin-top: 0.5rem;
  color: var(--color-danger, #d7373f);
  font-size: 0.875rem;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 2rem;
}

.cancel-btn, .submit-btn {
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.cancel-btn {
  background: var(--color-background, #f4f6f9);
  color: var(--color-text-secondary, #64748b);
  border: 1px solid var(--color-border, #dfe4ea);
}

.cancel-btn:hover {
  background: var(--color-background, #f4f6f9);
  border-color: #cbd5e0;
}

.submit-btn {
  background: var(--color-primary, #006b64);
  color: white;
}

.submit-btn:hover:not(:disabled) {
  background: var(--color-primary-dark, #008a7f);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
