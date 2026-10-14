// Composable para manejo de semestres
export const useSemester = () => {
  const currentSemester = ref(null)
  const loading = ref(false)
  const error = ref(null)

  // Función auxiliar para obtener el token
  const getAuthToken = () => {
    if (import.meta.client) {
      return localStorage.getItem('auth_token')
    }
    return null
  }

  const getCurrentSemester = async () => {
    loading.value = true
    error.value = null
    
    try {
      const token = getAuthToken()
      console.log('Token obtenido para getCurrentSemester:', token) // Debug
      if (!token) {
        throw new Error('No authentication token found')
      }

      const response = await $fetch('http://127.0.0.1:3000/api/v1/semesters/current', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      
      if (response.success) {
        currentSemester.value = response.data
      } else {
        throw new Error(response.message || 'Error al obtener el semestre actual')
      }
    } catch (err) {
      error.value = err.message || 'Error de conexión'
      console.error('Error getting current semester:', err)
      
      // Fallback data
      currentSemester.value = {
        year: 2025,
        term: 2,
        is_active: true
      }
    } finally {
      loading.value = false
    }
  }

  const createOrActivateSemester = async (year, term) => {
    loading.value = true
    error.value = null

    try {
      const token = getAuthToken()
      console.log('Token obtenido para createOrActivateSemester:', token) // Debug
      if (!token) {
        throw new Error('No authentication token found')
      }

      const response = await $fetch('http://127.0.0.1:3000/api/v1/semesters/create-or-activate', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ year, term })
      })

      if (response.success) {
        currentSemester.value = response.data
        return response.data
      } else {
        throw new Error(response.message || 'Error al crear o activar el semestre')
      }
    } catch (err) {
      error.value = err.message || 'Error de conexión'
      console.error('Error creating or activating semester:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  const getAllSemesters = async () => {
    try {
      const token = getAuthToken()
      console.log('Token obtenido para getAllSemesters:', token) // Debug
      if (!token) {
        throw new Error('No authentication token found')
      }

      const response = await $fetch('http://127.0.0.1:3000/api/v1/semesters/', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      
      return response || []
    } catch (err) {
      console.error('Error getting all semesters:', err)
      return []
    }
  }

  const formatSemesterName = computed(() => {
    if (!currentSemester.value) return 'N/A'
    return `${currentSemester.value.year}-${currentSemester.value.term}`
  })

  const isRegistrationOpen = computed(() => {
    if (!currentSemester.value) return false
    
    const now = new Date()
    const registrationStart = new Date(currentSemester.value.registration_start)
    const registrationEnd = new Date(currentSemester.value.registration_end)
    
    return now >= registrationStart && now <= registrationEnd
  })

  const formatRegistrationDeadline = computed(() => {
    if (!currentSemester.value?.registration_end) return 'N/A'
    
    const date = new Date(currentSemester.value.registration_end)
    return date.toLocaleDateString('es-CL', {
      day: 'numeric',
      month: 'long',
      year: 'numeric'
    })
  })

  return {
    currentSemester: readonly(currentSemester),
    loading: readonly(loading),
    error: readonly(error),
    formatSemesterName,
    isRegistrationOpen,
    formatRegistrationDeadline,
    getCurrentSemester,
    createOrActivateSemester,
    getAllSemesters
  }
}
