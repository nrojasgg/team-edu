<template>
  <div class="reset-password-page">
    <div class="container">
      <div class="reset-form-container">
        <h1>Restablecer Contraseña</h1>
        
        <div v-if="!isSuccess">
          <p class="description">Ingresa tu nueva contraseña para actualizar tu cuenta.</p>
          
          <form @submit.prevent="handleResetPassword" class="reset-form">
            <div class="form-group">
              <label for="new-password">Nueva Contraseña</label>
              <input
                id="new-password"
                v-model="formData.password"
                type="password"
                required
                minlength="6"
                placeholder="Mínimo 6 caracteres"
                class="form-input"
              />
            </div>
            
            <div class="form-group">
              <label for="confirm-password">Confirmar Contraseña</label>
              <input
                id="confirm-password"
                v-model="formData.confirmPassword"
                type="password"
                required
                minlength="6"
                placeholder="Repite tu contraseña"
                class="form-input"
              />
            </div>
            
            <div v-if="error" class="error-message">
              {{ error }}
            </div>
            
            <button
              type="submit"
              :disabled="isLoading"
              class="primary-btn"
            >
              {{ isLoading ? 'Restableciendo...' : 'Restablecer Contraseña' }}
            </button>
          </form>
        </div>
        
        <div v-else class="success-content">
          <div class="success-icon">✅</div>
          <h2>¡Éxito!</h2>
          <p>Tu contraseña ha sido restablecida correctamente.</p>
          <button @click="goToLogin" class="primary-btn">
            Ir a Iniciar Sesión
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
const { resetPassword } = useAuth()
const route = useRoute()

const isLoading = ref(false)
const isSuccess = ref(false)
const error = ref('')
const token = ref('')

const formData = ref({
  password: '',
  confirmPassword: ''
})

onMounted(() => {
  // Obtener token de la URL
  token.value = route.query.token
  if (!token.value) {
    error.value = 'Enlace de restablecimiento inválido o expirado.'
  }
})

const handleResetPassword = async () => {
  if (formData.value.password !== formData.value.confirmPassword) {
    error.value = 'Las contraseñas no coinciden'
    return
  }
  
  if (!token.value) {
    error.value = 'No se encontró un token válido. Por favor, solicita de nuevo el restablecimiento.'
    return
  }

  isLoading.value = true
  error.value = ''

  try {
    await resetPassword(token.value, formData.value.password)
    isSuccess.value = true
  } catch (err) {
    error.value = err.message || 'Ocurrió un error al restablecer la contraseña'
  } finally {
    isLoading.value = false
  }
}

const goToLogin = () => {
  navigateTo('/')
}

// Metadatos de la página
useHead({
  title: 'Restablecer Contraseña - TeamEdu',
})
</script>

<style scoped>
.reset-password-page {
  min-height: 100vh;
  background: #00A499;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.container {
  width: 100%;
  max-width: 500px;
}

.reset-form-container {
  background: white;
  padding: 2.5rem;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
}

h1 {
  color: #333;
  margin-bottom: 0.5rem;
  text-align: center;
  font-size: 1.8rem;
}

.description {
  color: #666;
  text-align: center;
  margin-bottom: 2rem;
  font-size: 0.95rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: #555;
  font-weight: 500;
  font-size: 0.9rem;
}

.form-input {
  width: 100%;
  padding: 0.8rem;
  border: 2px solid var(--color-border, #dfe4ea);
  border-radius: 8px;
  font-size: 1rem;
  box-sizing: border-box;
  transition: border-color 0.3s;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary, #00A499);
}

.primary-btn {
  width: 100%;
  padding: 0.9rem;
  background: var(--color-primary, #006b64);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.3s;
}

.primary-btn:hover:not(:disabled) {
  background: var(--color-primary-dark, #008a7f);
}

.primary-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-message {
  background: #fed7d7;
  color: #b82e35;
  padding: 0.75rem;
  border-radius: 6px;
  margin-bottom: 1.5rem;
  font-size: 0.9rem;
  text-align: center;
}

.success-content {
  text-align: center;
}

.success-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.success-content h2 {
  color: var(--color-primary-dark, #008a7f);
  margin-bottom: 1rem;
}

.success-content p {
  color: var(--color-text-secondary, #64748b);
  margin-bottom: 2rem;
}
</style>
