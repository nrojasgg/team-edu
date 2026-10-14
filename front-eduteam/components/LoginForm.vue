<template>
  <div class="auth-container">
    <!-- Vista de Login -->
    <div v-if="!isRegisterMode" class="auth-card">

      <div class="card-header">
        <h2>Bienvenido</h2>
        <p>Ingresa tus credenciales para acceder al sistema</p>
      </div>

      <!-- Google OAuth (destacado al inicio) -->
      <button
        type="button"
        @click="handleGoogleLogin"
        :disabled="isLoading"
        class="google-btn"
      >
        <svg viewBox="0 0 24 24" class="google-icon">
          <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
          <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
          <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
          <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
        </svg>
        Continuar con Google (@usach.cl)
      </button>

      <div class="divider"><span>o ingresa con tu cuenta</span></div>

      <form @submit.prevent="handleLogin" class="auth-form">
        <div class="form-group">
          <label for="email">Usuario USACH</label>
          <div class="input-suffix-wrapper">
            <div class="input-wrapper input-suffix-field">
              <svg class="input-icon" viewBox="0 0 20 20" fill="currentColor">
                <path d="M2.003 5.884L10 9.882l7.997-3.998A2 2 0 0016 4H4a2 2 0 00-1.997 1.884z"/>
                <path d="M18 8.118l-8 4-8-4V14a2 2 0 002 2h12a2 2 0 002-2V8.118z"/>
              </svg>
              <input
                id="email"
                v-model="loginForm.username"
                type="text"
                required
                placeholder="nombre.apellido"
                class="form-input"
                autocomplete="username"
              />
            </div>
            <span class="input-suffix">@usach.cl</span>
          </div>
        </div>

        <div class="form-group">
          <div class="label-row">
            <label for="password">Contraseña</label>
            <button type="button" @click="showForgotPassword = true" class="link-btn">¿Olvidaste tu contraseña?</button>
          </div>
          <div class="input-wrapper">
            <svg class="input-icon" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd"/>
            </svg>
            <input
              id="password"
              v-model="loginForm.password"
              type="password"
              required
              placeholder="Tu contraseña"
              class="form-input"
            />
          </div>
        </div>

        <button type="submit" :disabled="isLoading" class="primary-btn">
          <span v-if="isLoading" class="spinner"></span>
          {{ isLoading ? 'Verificando...' : 'Iniciar sesión' }}
        </button>
      </form>

      <p class="register-prompt">
        ¿No tienes cuenta?
        <button type="button" @click="isRegisterMode = true" class="link-btn--inline">Regístrate aquí</button>
      </p>
    </div>

    <!-- Vista de Registro -->
    <div v-else class="auth-card">
      <div class="card-header">
        <h2>Crear cuenta</h2>
        <p>Completa los datos para registrarte en el sistema</p>
      </div>

      <form @submit.prevent="handleRegister" class="auth-form">
        <div class="form-row">
          <div class="form-group">
            <label for="register-nombres">Nombres</label>
            <input
              id="register-nombres"
              v-model="registerForm.nombres"
              type="text"
              placeholder="Tus nombres"
              class="form-input"
            />
          </div>

          <div class="form-group">
            <label for="register-paterno">Apellido Paterno</label>
            <input
              id="register-paterno"
              v-model="registerForm.paterno"
              type="text"
              placeholder="Apellido paterno"
              class="form-input"
            />
          </div>
        </div>

        <div class="form-group">
          <label for="register-materno">Apellido Materno</label>
          <input
            id="register-materno"
            v-model="registerForm.materno"
            type="text"
            placeholder="Apellido materno"
            class="form-input"
          />
        </div>

        <div class="form-group">
          <label for="register-rut">RUT</label>
          <input
            id="register-rut"
            v-model="registerForm.rut"
            type="text"
            placeholder="12.345.678-9"
            class="form-input"
          />
        </div>

        <div class="form-group">
          <label for="register-email">Usuario USACH *</label>
          <div class="input-suffix-wrapper">
            <div class="input-wrapper input-suffix-field">
              <svg class="input-icon" viewBox="0 0 20 20" fill="currentColor">
                <path d="M2.003 5.884L10 9.882l7.997-3.998A2 2 0 0016 4H4a2 2 0 00-1.997 1.884z"/>
                <path d="M18 8.118l-8 4-8-4V14a2 2 0 002 2h12a2 2 0 002-2V8.118z"/>
              </svg>
              <input
                id="register-email"
                v-model="registerForm.username"
                type="text"
                required
                placeholder="nombre.apellido"
                class="form-input"
                autocomplete="username"
              />
            </div>
            <span class="input-suffix">@usach.cl</span>
          </div>
        </div>

        <div class="form-group">
          <label for="register-password">Contraseña *</label>
          <input
            id="register-password"
            v-model="registerForm.password"
            type="password"
            required
            minlength="6"
            placeholder="Mínimo 6 caracteres"
            class="form-input"
          />
        </div>

        <div class="form-group">
          <label for="register-confirm-password">Confirmar contraseña *</label>
          <input
            id="register-confirm-password"
            v-model="registerForm.confirmPassword"
            type="password"
            required
            minlength="6"
            placeholder="Repite tu contraseña"
            class="form-input"
          />
        </div>

        <button
          type="submit"
          :disabled="isLoading || !passwordsMatch"
          class="primary-btn"
        >
          <span v-if="isLoading" class="spinner"></span>
          {{ isLoading ? 'Creando cuenta...' : 'Crear cuenta' }}
        </button>
      </form>

      <p class="register-prompt">
        ¿Ya tienes cuenta?
        <button type="button" @click="isRegisterMode = false" class="link-btn--inline">Inicia sesión</button>
      </p>
    </div>

    <!-- Modal de Olvidé mi contraseña -->
    <div v-if="showForgotPassword" class="modal-overlay" @click="showForgotPassword = false">
      <div class="modal-content" @click.stop>
        <div class="modal-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" />
          </svg>
        </div>
        <h3>Recuperar contraseña</h3>
        <p>Ingresa tu correo institucional y te enviaremos las instrucciones para restablecer tu contraseña.</p>

        <form @submit.prevent="handleForgotPassword" class="auth-form">
          <div class="form-group">
            <label for="forgot-email">Correo electrónico</label>
            <input
              id="forgot-email"
              v-model="forgotEmail"
              type="email"
              required
              placeholder="tu@email.com"
              class="form-input"
            />
          </div>

          <div class="modal-actions">
            <button type="button" @click="showForgotPassword = false" class="modal-cancel-btn">
              Cancelar
            </button>
            <button type="submit" :disabled="isLoading" class="primary-btn">
              <span v-if="isLoading" class="spinner"></span>
              {{ isLoading ? 'Enviando...' : 'Enviar instrucciones' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <div v-if="successMessage" class="success-message">
      {{ successMessage }}
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const emit = defineEmits(['login-success'])
const { authenticatedFetch, initiateGoogleAuth } = useAuth()

// Estados
const isLoading = ref(false)
const error = ref('')
const successMessage = ref('')
const isRegisterMode = ref(false)
const showForgotPassword = ref(false)

// Formularios
const loginForm = ref({
  username: '',
  password: ''
})

const registerForm = ref({
  nombres: '',
  paterno: '',
  materno: '',
  rut: '',
  username: '',
  password: '',
  confirmPassword: ''
})

const forgotEmail = ref('')

// Computed
const passwordsMatch = computed(() => {
  return registerForm.value.password === registerForm.value.confirmPassword
})

// Funciones
const handleLogin = async () => {
  isLoading.value = true
  error.value = ''
  successMessage.value = ''

  try {
    const response = await fetch('http://127.0.0.1:3001/api/v1/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        email: loginForm.value.username + '@usach.cl',
        password: loginForm.value.password
      })
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.message || 'Error al iniciar sesión')
    }

    // Guardar token y usuario
    const userData = data.data.user
    // Aplanar la información del rol para compatibilidad con el middleware
    const flattenedUser = {
      ...userData,
      role_name: userData.role?.name || 'estudiante'
    }

    localStorage.setItem('auth_token', data.data.token)
    localStorage.setItem('user_info', JSON.stringify(flattenedUser))

    emit('login-success', flattenedUser)
  } catch (err) {
    error.value = err.message || 'Error al iniciar sesión'
    console.error('Login error:', err)
  } finally {
    isLoading.value = false
  }
}

const handleRegister = async () => {
  if (!passwordsMatch.value) {
    error.value = 'Las contraseñas no coinciden'
    return
  }

  isLoading.value = true
  error.value = ''
  successMessage.value = ''

  try {
    const registerData = {
      nombres: registerForm.value.nombres || null,
      paterno: registerForm.value.paterno || null,
      materno: registerForm.value.materno || null,
      rut: registerForm.value.rut || null,
      email: registerForm.value.username + '@usach.cl',
      password: registerForm.value.password,
      role_id: 3, // Rol de estudiante por defecto
      career_code: null
    }

    const response = await fetch('http://127.0.0.1:3001/api/v1/auth/register', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(registerData)
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.message || 'Error al crear la cuenta')
    }

    successMessage.value = 'Cuenta creada exitosamente. Ahora puedes iniciar sesión.'
    isRegisterMode.value = false

    // Limpiar formulario
    registerForm.value = {
      nombres: '',
      paterno: '',
      materno: '',
      rut: '',
      username: '',
      password: '',
      confirmPassword: ''
    }
  } catch (err) {
    error.value = err.message || 'Error al crear la cuenta'
    console.error('Register error:', err)
  } finally {
    isLoading.value = false
  }
}

const handleGoogleLogin = async () => {
  isLoading.value = true
  error.value = ''

  try {
    await initiateGoogleAuth()
  } catch (err) {
    error.value = err.message || 'Error al conectar con Google'
    console.error('Error initiating Google auth:', err)
  } finally {
    isLoading.value = false
  }
}

const handleForgotPassword = async () => {
  isLoading.value = true
  error.value = ''
  successMessage.value = ''

  try {
    const resetURL = `${globalThis.location.origin}/reset-password`
    const { forgotPassword } = useAuth()
    const response = await forgotPassword(forgotEmail.value, resetURL)

    successMessage.value = response.message || 'Se han enviado las instrucciones de recuperación a tu correo electrónico.'
    showForgotPassword.value = false
    forgotEmail.value = ''
  } catch (err) {
    error.value = err.message || 'Error al enviar las instrucciones de recuperación'
    console.error('Forgot password error:', err)
  } finally {
    isLoading.value = false
  }
}
</script>

<style scoped>
* {
  box-sizing: border-box;
}

.auth-container {
  width: 100%;
  max-width: 420px;
}

/* ── Card principal ── */
.auth-card {
  background: white;
  border-radius: 16px;
  padding: 2.25rem 2rem;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08), 0 1px 4px rgba(0, 0, 0, 0.04);
  border: 1px solid #e8edf2;
}

.card-header {
  margin-bottom: 1.75rem;
  text-align: center;
}

.card-header h2 {
  font-size: 1.5rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 0.35rem 0;
  letter-spacing: -0.3px;
}

.card-header p {
  font-size: 0.875rem;
  color: #64748b;
  margin: 0;
}

/* ── Google button ── */
.google-btn {
  width: 100%;
  padding: 0.75rem 1rem;
  background: white;
  color: #1e293b;
  border: 1.5px solid #cbd5e1;
  border-radius: 10px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  margin-bottom: 1.25rem;
}

.google-btn:hover:not(:disabled) {
  border-color: #4285F4;
  background: #f8fbff;
  box-shadow: 0 2px 8px rgba(66, 133, 244, 0.15);
}

.google-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.google-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

/* ── Divider ── */
.divider {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
}

.divider::before,
.divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--color-border, #dfe4ea);
}

.divider span {
  font-size: 0.78rem;
  color: #94a3b8;
  white-space: nowrap;
}

/* ── Formulario ── */
.auth-form {
  margin-bottom: 1.25rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.4rem;
  font-size: 0.82rem;
  font-weight: 600;
  color: #374151;
  letter-spacing: 0.01em;
}

.label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.4rem;
}

.label-row label {
  margin-bottom: 0;
}

.input-wrapper {
  position: relative;
}

.input-icon {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  color: #94a3b8;
  pointer-events: none;
}

.form-input {
  width: 100%;
  padding: 0.65rem 0.875rem 0.65rem 2.4rem;
  border: 1.5px solid #d1d5db;
  border-radius: 8px;
  font-size: 0.9rem;
  color: #1e293b;
  background: #fafafa;
  transition: border-color 0.2s, box-shadow 0.2s, background 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: #00A499;
  background: white;
  box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.12);
}

.form-input::placeholder {
  color: #b0bec5;
}

/* Sin icono (registro) */
.form-group:not(:has(.input-wrapper)) .form-input,
.form-group .form-input:not(.input-wrapper > .form-input) {
  padding-left: 0.875rem;
}

/* ── Input con sufijo @usach.cl ── */
.input-suffix-wrapper {
  display: flex;
  align-items: stretch;
  border: 1.5px solid #d1d5db;
  border-radius: 8px;
  overflow: hidden;
  background: #fafafa;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.input-suffix-wrapper:focus-within {
  border-color: #00A499;
  background: white;
  box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.12);
}

.input-suffix-field {
  flex: 1;
  min-width: 0;
}

.input-suffix-field .form-input {
  border: none;
  border-radius: 0;
  background: transparent;
  box-shadow: none !important;
  padding-right: 0.5rem;
}

.input-suffix-field .form-input:focus {
  border: none;
  box-shadow: none;
  outline: none;
}

.input-suffix {
  display: flex;
  align-items: center;
  padding: 0 0.75rem;
  background: #f1f5f9;
  border-left: 1.5px solid #d1d5db;
  color: #4b5563;
  font-size: 0.875rem;
  font-weight: 500;
  white-space: nowrap;
  user-select: none;
  flex-shrink: 0;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

/* ── Botón primario ── */
.primary-btn {
  width: 100%;
  padding: 0.75rem;
  background: linear-gradient(135deg, #006b64, #007a71);
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  letter-spacing: 0.01em;
}

.primary-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #009188, #006860);
  box-shadow: 0 4px 12px rgba(0, 164, 153, 0.35);
  transform: translateY(-1px);
}

.primary-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

/* ── Spinner ── */
.spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255,255,255,0.35);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
  flex-shrink: 0;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ── Links ── */
.link-btn {
  background: none;
  border: none;
  color: #00A499;
  cursor: pointer;
  font-size: 0.78rem;
  padding: 0;
  font-weight: 500;
}

.link-btn:hover {
  color: #007a71;
  text-decoration: underline;
}

.link-btn--inline {
  background: none;
  border: none;
  color: #00A499;
  cursor: pointer;
  font-size: 0.875rem;
  padding: 0;
  font-weight: 600;
}

.link-btn--inline:hover {
  color: #007a71;
  text-decoration: underline;
}

.register-prompt {
  text-align: center;
  font-size: 0.875rem;
  color: #64748b;
  margin: 0;
}

/* ── Modal ── */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(2px);
}

.modal-content {
  background: white;
  border-radius: 16px;
  padding: 2rem;
  max-width: 400px;
  width: 90%;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.modal-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: rgba(0, 164, 153, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1rem;
}

.modal-icon svg {
  width: 24px;
  height: 24px;
  color: #00A499;
}

.modal-content h3 {
  margin: 0 0 0.5rem 0;
  color: #0f172a;
  font-size: 1.25rem;
  font-weight: 700;
  text-align: center;
}

.modal-content p {
  margin: 0 0 1.5rem 0;
  color: #64748b;
  font-size: 0.875rem;
  line-height: 1.6;
  text-align: center;
}

.modal-actions {
  display: flex;
  gap: 0.75rem;
  margin-top: 1.25rem;
}

.modal-actions .primary-btn {
  flex: 1;
}

.modal-cancel-btn {
  flex: 1;
  padding: 0.75rem;
  background: #f1f5f9;
  color: #475569;
  border: none;
  border-radius: 10px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.modal-cancel-btn:hover {
  background: var(--color-border, #dfe4ea);
}

/* ── Alertas ── */
.error-message {
  display: flex;
  align-items: flex-start;
  gap: 0.6rem;
  padding: 0.75rem 1rem;
  background: #fef2f2;
  color: #b91c1c;
  border: 1px solid #fecaca;
  border-radius: 8px;
  margin-top: 1rem;
  font-size: 0.85rem;
  line-height: 1.4;
}

.error-message::before {
  content: '⚠';
  flex-shrink: 0;
  font-size: 0.9rem;
}

.success-message {
  display: flex;
  align-items: flex-start;
  gap: 0.6rem;
  padding: 0.75rem 1rem;
  background: #f0fdf4;
  color: #15803d;
  border: 1px solid #bbf7d0;
  border-radius: 8px;
  margin-top: 1rem;
  font-size: 0.85rem;
  line-height: 1.4;
}

.success-message::before {
  content: '✓';
  flex-shrink: 0;
  font-size: 0.9rem;
  font-weight: 700;
}

/* ── Responsive ── */
@media (max-width: 480px) {
  .auth-card {
    padding: 1.5rem 1.25rem;
    border-radius: 12px;
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .modal-content {
    margin: 1rem;
    padding: 1.5rem;
  }

  .modal-actions {
    flex-direction: column;
  }
}
</style>
