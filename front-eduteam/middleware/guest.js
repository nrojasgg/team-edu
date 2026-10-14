// Middleware para redirigir usuarios ya autenticados
export default defineNuxtRouteMiddleware(async (to, from) => {
  // Solo ejecutar en el cliente
  if (import.meta.server) return

  // Verificar si hay un token de autenticación
  const token = localStorage.getItem('auth_token')
  const userInfo = localStorage.getItem('user_info')

  // Si ya está autenticado, redirigir a home
  if (token && userInfo) {
    try {
      const user = JSON.parse(userInfo)
      if (user.id && user.email && user.role_name) {
        // Verificar si el token ha expirado
        try {
          const payload = JSON.parse(atob(token.split('.')[1]))
          const currentTime = Math.floor(Date.now() / 1000)
          if (payload.exp >= currentTime) {
            await navigateTo('/home')
            return
          } else {
            // Token expirado, limpiar datos
            localStorage.removeItem('auth_token')
            localStorage.removeItem('user_info')
          }
        } catch (error) {
          // Error al decodificar el token, limpiar datos
          console.error('Error decoding token in guest middleware:', error)
          localStorage.removeItem('auth_token')
          localStorage.removeItem('user_info')
        }
      }
    } catch (error) {
      // Si hay error en los datos, limpiar
      console.error('Error parsing user info in guest middleware:', error)
      localStorage.removeItem('auth_token')
      localStorage.removeItem('user_info')
    }
  }
})
