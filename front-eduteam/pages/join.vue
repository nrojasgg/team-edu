<template>
  <div class="join-page">
    <div class="join-container">
      <div class="join-card">
        <div class="join-header">
          <div class="logo">
            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
              <circle cx="9" cy="7" r="4"/>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
            </svg>
          </div>
          <h1>Únete a una coordinación</h1>
          <p>Ingresa el código de invitación que te proporcionó tu profesor</p>
        </div>

        <form @submit.prevent="handleSubmit" class="join-form">
          <div class="form-group">
            <label for="invitation-code">Código de invitación</label>
            <div class="input-wrapper" :class="{ valid: isValid, invalid: isInvalid }">
              <input
                id="invitation-code"
                v-model="invitationCode"
                type="text"
                required
                placeholder="ABC123XY"
                @input="validateCode"
                @blur="validateCode"
                autocomplete="off"
                spellcheck="false"
              />
              <div class="input-icon">
                <svg v-if="isValid" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="20 6 9 17 4 12"/>
                </svg>
                <svg v-else-if="isInvalid" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"/>
                  <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </div>
            </div>
            <span v-if="errorMessage" class="error-text">{{ errorMessage }}</span>
          </div>

          <button
            type="submit"
            class="submit-btn"
            :disabled="!invitationCode || submitting"
          >
            <span v-if="submitting">Verificando...</span>
            <span v-else>Unirme a la coordinación</span>
          </button>
        </form>

        <div class="join-footer">
          <div class="divider">
            <span>¿No tienes un código?</span>
          </div>
          <p class="help-text">Contacta a tu profesor para obtener uno.</p>
        </div>
      </div>

      <div v-if="success" class="success-overlay">
        <div class="success-card">
          <div class="success-icon">
            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
              <polyline points="22 4 12 14.01 9 11.01"/>
            </svg>
          </div>
          <h2>¡Te has unido correctamente!</h2>
          <p>Ahora puedes acceder a los recursos de tu coordinación.</p>
          <button @click="goHome" class="go-home-btn">
            Ir al inicio
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

definePageMeta({
  middleware: 'auth'
})

const { getToken } = useAuth()

const API_BASE_URL = 'http://127.0.0.1:3000/api/v1'

const invitationCode = ref('')
const submitting = ref(false)
const errorMessage = ref('')
const success = ref(false)
const isValid = ref(false)
const isInvalid = ref(false)

const validateCode = () => {
  const code = invitationCode.value.toUpperCase().trim()
  invitationCode.value = code

  if (code.length === 0) {
    isValid.value = false
    isInvalid.value = false
    errorMessage.value = ''
    return
  }

  if (code.length < 4) {
    isValid.value = false
    isInvalid.value = true
    errorMessage.value = ''
    return
  }

  isValid.value = false
  isInvalid.value = false
  errorMessage.value = ''
}

const handleSubmit = async () => {
  if (!invitationCode.value || submitting.value) return

  submitting.value = true
  errorMessage.value = ''

  try {
    const token = getToken()

    const response = await fetch(`${API_BASE_URL}/invitation-codes/use`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        code: invitationCode.value.toUpperCase().trim()
      })
    })

    const result = await response.json()

    if (!response.ok) {
      throw new Error(result.message || 'El código no es válido o ha expirado')
    }

    success.value = true
  } catch (err) {
    errorMessage.value = err.message || 'El código no es válido o ha expirado'
    isInvalid.value = true
    isValid.value = false
  } finally {
    submitting.value = false
  }
}

const goHome = () => {
  navigateTo('/home')
}
</script>

<style scoped>
.join-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #394049 0%, #2a3038 100%);
  padding: 1rem;
}

.join-container {
  width: 100%;
  max-width: 440px;
}

.join-card {
  background: white;
  border-radius: 20px;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.25);
  overflow: hidden;
}

.join-header {
  text-align: center;
  padding: 2.5rem 2rem 2rem;
  background: linear-gradient(135deg, #394049 0%, #2a3038 100%);
  color: white;
}

.logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  margin-bottom: 1.5rem;
}

.logo svg {
  color: var(--color-primary, #00A499);
}

.join-header h1 {
  margin: 0 0 0.5rem;
  font-size: 1.5rem;
  font-weight: 700;
}

.join-header p {
  margin: 0;
  font-size: 0.95rem;
  color: rgba(255, 255, 255, 0.7);
  line-height: 1.5;
}

.join-form {
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

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-wrapper input {
  width: 100%;
  padding: 1rem 3rem 1rem 1rem;
  border: 2px solid var(--color-border, #dfe4ea);
  border-radius: 12px;
  font-size: 1.1rem;
  font-weight: 600;
  letter-spacing: 3px;
  text-transform: uppercase;
  transition: all 0.2s;
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
}

.input-wrapper input:focus {
  outline: none;
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 0 0 4px rgba(0, 164, 153, 0.15);
}

.input-wrapper.valid input {
  border-color: #22c55e;
  background: #f0fdf4;
}

.input-wrapper.valid .input-icon {
  color: #22c55e;
}

.input-wrapper.invalid input {
  border-color: #ef4444;
  background: #fef2f2;
}

.input-wrapper.invalid .input-icon {
  color: #ef4444;
}

.input-icon {
  position: absolute;
  right: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
}

.error-text {
  display: block;
  margin-top: 0.5rem;
  color: #ef4444;
  font-size: 0.85rem;
}

.submit-btn {
  width: 100%;
  padding: 1rem;
  background: var(--color-primary, #006b64);
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.submit-btn:hover:not(:disabled) {
  background: #008a81;
  box-shadow: 0 4px 16px rgba(0, 164, 153, 0.35);
  transform: translateY(-1px);
}

.submit-btn:disabled {
  background: #9ca3af;
  cursor: not-allowed;
}

.join-footer {
  padding: 0 2rem 2rem;
  text-align: center;
}

.divider {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.divider::before,
.divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--color-border, #dfe4ea);
}

.divider span {
  font-size: 0.85rem;
  color: #9ca3af;
}

.help-text {
  margin: 0;
  font-size: 0.9rem;
  color: var(--color-text-secondary, #64707d);
}

.success-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.3s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.success-card {
  background: white;
  border-radius: 20px;
  padding: 2.5rem;
  text-align: center;
  max-width: 360px;
  margin: 1rem;
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

.success-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
  background: #dcfce7;
  border-radius: 50%;
  margin-bottom: 1.5rem;
}

.success-icon svg {
  color: #22c55e;
}

.success-card h2 {
  margin: 0 0 0.5rem;
  color: var(--color-dark, #394049);
  font-size: 1.35rem;
  font-weight: 700;
}

.success-card p {
  margin: 0 0 1.5rem;
  color: var(--color-text-secondary, #64707d);
  font-size: 0.95rem;
}

.go-home-btn {
  width: 100%;
  padding: 0.85rem 1.5rem;
  background: var(--color-primary, #006b64);
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.go-home-btn:hover {
  background: #008a81;
}

@media (max-width: 480px) {
  .join-header {
    padding: 2rem 1.5rem 1.5rem;
  }

  .join-header h1 {
    font-size: 1.25rem;
  }

  .join-form {
    padding: 1.5rem;
  }

  .input-wrapper input {
    font-size: 1rem;
    letter-spacing: 2px;
  }
}
</style>
