<template>
  <div class="callback-container">
    <div class="callback-content">
      <div v-if="loading" class="loading">
        <div class="spinner"></div>
        <h2>Procesando autenticación...</h2>
        <p>Por favor espera mientras verificamos tu cuenta con Google.</p>
      </div>
      
      <div v-if="error" class="error">
        <div class="error-icon">⚠️</div>
        <h2>Error de Autenticación</h2>
        <p>{{ error }}</p>
        <button @click="goToHome" class="back-btn">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M19 12H5"/>
            <path d="M12 19l-7-7 7-7"/>
          </svg>
          Volver al Inicio
        </button>
      </div>
      
      <div v-if="success" class="success">
        <div class="success-icon">✅</div>
        <h2>¡Autenticación Exitosa!</h2>
        <p>Bienvenido {{ user?.fullname || user?.nombres }}. Serás redirigido en breve...</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()
const { handleGoogleCallback, saveToken, saveUser } = useAuth()

const loading = ref(true)
const error = ref('')
const success = ref(false)
const user = ref(null)

const processGoogleCallback = async () => {
  try {
    // Obtener el código de la URL
    const code = route.query.code
    const errorParam = route.query.error
    
    if (errorParam) {
      throw new Error('Usuario canceló la autenticación con Google')
    }
    
    if (!code) {
      throw new Error('No se recibió el código de autorización de Google')
    }

    // Procesar el callback con el backend
    const authData = await handleGoogleCallback(code)
    
    console.log('Datos de autenticación recibidos:', authData) // Debug
    
    // Guardar token y información del usuario
    saveToken(authData.token)
    saveUser(authData.user)
    
    user.value = authData.user
    success.value = true
    
    // Mostrar mensaje de éxito por 2 segundos y luego redirigir
    setTimeout(() => {
      // Redirigir a home para todos los usuarios
      router.push('/home')
    }, 2000)
    
  } catch (err) {
    console.error('Error processing Google callback:', err)
    error.value = err.message || 'Error desconocido en la autenticación'
  } finally {
    loading.value = false
  }
}

const goToHome = () => {
  router.push('/')
}

onMounted(() => {
  processGoogleCallback()
})

// Meta información para SEO
useHead({
  title: 'Autenticación - TeamEdu',
  meta: [
    { name: 'description', content: 'Procesando autenticación con Google' }
  ]
})
</script>

<style scoped>
.callback-container {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: #00A499;
  padding: 1rem;
}

.callback-content {
  background: white;
  border-radius: 12px;
  padding: 3rem;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  text-align: center;
  max-width: 500px;
  width: 100%;
}

.loading {
  color: #333;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #00A499;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 1.5rem;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error {
  color: #721c24;
}

.error-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.success {
  color: #155724;
}

.success-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

h2 {
  margin: 1rem 0;
  font-size: 1.5rem;
}

p {
  margin: 0.5rem 0;
  color: #666;
  line-height: 1.5;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1.5rem;
  background: #a85500;
  border: 1px solid #a85500;
  border-radius: 12px;
  color: white;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 500;
  font-size: 1rem;
  margin-top: 1rem;
}

.back-btn:hover {
  background: #edf2f7;
  border-color: #cbd5e0;
  color: #a85500;
}

@media (max-width: 768px) {
  .back-btn {
    padding: 0.5rem 1rem;
    font-size: 0.9rem;
    gap: 0.5rem;
    margin-top: 0.75rem;
  }
  
  .back-btn svg {
    width: 18px;
    height: 18px;
  }
}
</style>
