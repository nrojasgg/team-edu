// Middleware para verificar autenticación
export default defineNuxtRouteMiddleware(async (to, from) => {
  // Solo ejecutar en el cliente
  if (import.meta.server) return

  // Verificar si hay un token de autenticación
  const token = localStorage.getItem('auth_token')
  const userInfo = localStorage.getItem('user_info')

  // Si no hay token o información de usuario, redirigir al login
  if (!token || !userInfo) {
    await navigateTo('/')
    return
  }

  // Verificar que la información del usuario sea válida
  try {
    const user = JSON.parse(userInfo)
    if (!user.id || !user.email || !user.role_name) {
      // Información de usuario inválida, limpiar y redirigir
      localStorage.removeItem('auth_token')
      localStorage.removeItem('user_info')
      await navigateTo('/')
      return
    }
  } catch (error) {
    // Error al parsear la información del usuario
    console.error('Error parsing user info in auth middleware:', error)
    localStorage.removeItem('auth_token')
    localStorage.removeItem('user_info')
    await navigateTo('/')
    return
  }

  // Verificar si el token ha expirado
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    const currentTime = Math.floor(Date.now() / 1000)
    if (payload.exp < currentTime) {
      // Token expirado, limpiar datos y redirigir
      localStorage.removeItem('auth_token')
      localStorage.removeItem('user_info')
      await navigateTo('/')
      return
    }
  } catch (error) {
    // Error al decodificar el token, asumir expirado
    console.error('Error decoding token in auth middleware:', error)
    localStorage.removeItem('auth_token')
    localStorage.removeItem('user_info')
    await navigateTo('/')
    return
  }
})
