// Composable para manejo de autenticación
export const useAuth = () => {
  const API_BASE_URL = 'http://127.0.0.1:3000/api/v1'

  // Función para manejar errores de autenticación
  const handleAuthError = () => {
    if (import.meta.client) {
      // Limpiar datos de autenticación
      localStorage.removeItem('auth_token')
      localStorage.removeItem('user_info')

      // Redirigir a la página principal
      globalThis.location.href = '/'
    }
  }

  // Función para hacer llamadas HTTP con manejo automático de autenticación
  const authenticatedFetch = async (url, options = {}) => {
    const token = getToken()

    // Si no hay token, no continuar
    if (!token) {
      handleAuthError()
      return null
    }

    // Verificar si el token ha expirado antes de hacer la petición
    if (isTokenExpired()) {
      console.warn('Token expirado, redirigiendo al login')
      handleAuthError()
      return null
    }

    // Agregar token de autorización si no está presente
    if (!options.headers) {
      options.headers = {}
    }

    if (!options.headers['Authorization']) {
      options.headers['Authorization'] = `Bearer ${token}`
    }

    if (!options.headers['Content-Type']) {
      options.headers['Content-Type'] = 'application/json'
    }

    // Crear AbortController para timeout
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 15000) // 15 segundos timeout

    try {
      // Agregar signal al options
      options.signal = controller.signal

      // Debug: show request headers and body in development
      if (import.meta.client && globalThis.location?.hostname === 'localhost') {
        try {
          console.debug('authenticatedFetch =>', url, options)
        } catch (e) { console.debug('Debug logging skipped:', e) }
      }
      const response = await fetch(url, options)

      // Limpiar timeout
      clearTimeout(timeoutId)

      // Si la respuesta es 401 (Unauthorized) o 403 (Forbidden), token expirado
      if (response.status === 401 || response.status === 403) {
        console.warn('Token expirado o inválido, redirigiendo al login')
        handleAuthError()
        return null // No continuar con la ejecución
      }

      return response
    } catch (error) {
      // Limpiar timeout
      clearTimeout(timeoutId)

      // Si es abort, es timeout
      if (error.name === 'AbortError') {
        console.error('Request timeout after 15 seconds')
        throw new Error('La solicitud tardó demasiado tiempo. Verifica tu conexión a internet.')
      }

      // Si hay un error de red, no hacer logout automático
      console.error('Error de red:', error)
      throw error
    }
  }

  // Obtener URL de autenticación de Google
  const getGoogleAuthURL = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/google`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      })

      if (!response.ok) {
        throw new Error('Error al obtener la URL de autenticación de Google')
      }

      const data = await response.json()
      return data.data.auth_url
    } catch (error) {
      console.error('Error getting Google auth URL:', error)
      throw error
    }
  }

  // Procesar callback de Google OAuth
  const handleGoogleCallback = async (code) => {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/google/callback`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ code }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Error en la autenticación con Google')
      }

      const data = await response.json()
      return data.data
    } catch (error) {
      console.error('Error in Google callback:', error)
      throw error
    }
  }

  // Iniciar autenticación con Google
  const initiateGoogleAuth = async () => {
    try {
      const authUrl = await getGoogleAuthURL()
      globalThis.location.href = authUrl
    } catch (error) {
      throw new Error('Error al iniciar autenticación con Google: ' + error.message)
    }
  }

  // Guardar token en localStorage
  const saveToken = (token) => {
    if (import.meta.client) {
      localStorage.setItem('auth_token', token)
    }
  }

  // Obtener token de localStorage
  const getToken = () => {
    if (import.meta.client) {
      return localStorage.getItem('auth_token')
    }
    return null
  }

  // Eliminar token de localStorage
  const removeToken = () => {
    if (import.meta.client) {
      localStorage.removeItem('auth_token')
    }
  }

  // Guardar información del usuario
  const saveUser = (user) => {
    if (import.meta.client) {
      // Aplanar la información del rol para compatibilidad con el middleware
      const flattenedUser = {
        ...user,
        role_name: user.role?.name || user.role_name || 'estudiante'
      }
      localStorage.setItem('user_info', JSON.stringify(flattenedUser))
    }
  }

  // Obtener información del usuario
  const getUser = () => {
    if (import.meta.client) {
      const userInfo = localStorage.getItem('user_info')
      return userInfo ? JSON.parse(userInfo) : null
    }
    return null
  }

  // Eliminar información del usuario
  const removeUser = () => {
    if (import.meta.client) {
      localStorage.removeItem('user_info')
    }
  }

  // Cerrar sesión
  const logout = () => {
    removeToken()
    removeUser()
    if (import.meta.client) {
      globalThis.location.href = '/'
    }
  }

  // Verificar si el usuario está autenticado
  const isAuthenticated = () => {
    return !!getToken()
  }

  // Verificar si el token ha expirado
  const isTokenExpired = () => {
    const token = getToken()
    if (!token) return true

    try {
      // Decodificar JWT (sin librería, asumiendo formato estándar)
      const payload = JSON.parse(atob(token.split('.')[1]))
      const currentTime = Math.floor(Date.now() / 1000)
      return payload.exp < currentTime
    } catch (error) {
      console.error('Error decodificando token:', error)
      return true // Si no se puede decodificar, asumir expirado
    }
  }

  return {
    getGoogleAuthURL,
    handleGoogleCallback,
    initiateGoogleAuth,
    saveToken,
    getToken,
    removeToken,
    saveUser,
    getUser,
    removeUser,
    logout,
    isAuthenticated,
    isTokenExpired,
    authenticatedFetch,
    handleAuthError,

    // Password reset
    forgotPassword: async (email, resetURL) => {
      const response = await fetch(`${API_BASE_URL}/auth/forgot-password`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, reset_url: resetURL })
      })
      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.message || 'Error al procesar la solicitud')
      }
      return await response.json()
    },
    resetPassword: async (token, newPassword) => {
      const response = await fetch(`${API_BASE_URL}/auth/reset-password`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token, new_password: newPassword })
      })
      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.message || 'Error al restablecer la contraseña')
      }
      return await response.json()
    }
  }
}
