<template>
  <div class="auth-page">
    <div class="container">
      <div class="welcome-section">
        <h1>TeamEdu</h1>
        <p class="subtitle">Sistema de gestión de equipos de trabajo USACH</p>
      </div>
      
      <div class="auth-content">
        <LoginForm @login-success="handleLoginSuccess" />
      </div>
      
    </div>
  </div>
</template>

<script setup>
// Aplicar middleware para usuarios no autenticados
definePageMeta({
  middleware: 'guest'
})

const handleLoginSuccess = (userData) => {
  console.log('Login exitoso:', userData)
  // Redirigir a home después del login exitoso
  navigateTo('/home')
}

// Metadatos de la página
useHead({
  title: 'TeamEdu - Sistema de gestión de equipos de trabajo',
  meta: [
    {
      name: 'description',
      content: 'Accede al sistema de gestión de equipos de trabajo de la Universidad de Santiago de Chile (USACH)'
    }
  ]
})
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  background: #00A499;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.container {
  width: 100%;
  max-width: 1200px;
  display: grid;
  grid-template-columns: 1fr;
  gap: 3rem;
  align-items: center;
  min-height: 80vh;
}

/* Desktop layout - two columns */
@media (min-width: 769px) {
  .container {
    grid-template-columns: 1fr 1fr;
    max-width: 1200px;
  }
  
  .welcome-section {
    margin-bottom: 0;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    height: 100%;
  }
  
  .auth-content {
    margin-bottom: 0;
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100%;
  }
}

.welcome-section {
  text-align: center;
}

.welcome-section h1 {
  color: white;
  font-size: 4rem;
  font-weight: 300;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

.subtitle {
  color: rgba(255, 255, 255, 0.9);
  font-size: 1.7rem;
  margin: 0;
  font-weight: 300;
}

.auth-content {
  margin-bottom: 3rem;
}

.info-item {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  text-align: center;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.info-item h3 {
  color: white;
  font-size: 1.1rem;
  margin: 0 0 0.75rem 0;
  font-weight: 600;
}

.info-item p {
  color: rgba(255, 255, 255, 0.8);
  font-size: 0.9rem;
  margin: 0;
  line-height: 1.4;
}

/* Responsive */
@media (max-width: 768px) {
  .container {
    grid-template-columns: 1fr;
  }
  
  .welcome-section h1 {
    font-size: 2.5rem;
  }
  
  .subtitle {
    font-size: 1.1rem;
  }
  
  .info-footer {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
  
  .info-item {
    padding: 1.25rem;
  }
}

@media (max-width: 480px) {
  .auth-page {
    padding: 0.5rem;
  }
  
  .welcome-section h1 {
    font-size: 2rem;
  }
  
  .subtitle {
    font-size: 1rem;
  }
  
  .container {
    max-width: 100%;
  }
  
  .welcome-section {
    margin-bottom: 2rem;
  }
  
  .auth-content {
    margin-bottom: 2rem;
  }
}
</style>
