import { ref, reactive } from 'vue'

// State moved outside to be shared across all instances
const sectionData = ref(null)
const courseData = ref(null)
const semesterInfo = ref('')
const professors = ref([])
const assistants = ref([])
const loading = ref(true)
const error = ref(null)

const assignments = reactive({
  professor_id: '',
  assistant_id: '',
  professor_ids: [],
  assistant_ids: []
})

export const useSectionGeneral = () => {
  const API_BASE_URL = 'http://127.0.0.1:3000/api/v1'
  const { authenticatedFetch } = useAuth()

  /**
   * Load section data including course and semester information
   * @param {string} sectionId - The section ID
   * @param {string} token - Authentication token
   */
  const loadSectionData = async (sectionId, token) => {
    try {
      loading.value = true
      error.value = null

      const response = await authenticatedFetch(`${API_BASE_URL}/sections/${sectionId}/complete`)

      if (!response) return // Token expirado, ya fue manejado por authenticatedFetch

      if (!response.ok) {
        throw new Error('No se pudo cargar la información de la sección')
      }

      const data = await response.json()

      // Asignar los datos que vienen pre-agregados
      sectionData.value = data.section || null
      courseData.value = data.course || null
      semesterInfo.value = data.semester_info || ''

      // Poblar asignaciones si existen
      if (data.section) {
        assignments.professor_id = data.section.professor_id || ''
        assignments.assistant_id = data.section.assistant_id || ''
        assignments.professor_ids = Array.isArray(data.section.professor_ids)
          ? [...data.section.professor_ids]
          : (data.section.professor_id ? [data.section.professor_id] : [])
        assignments.assistant_ids = Array.isArray(data.section.assistant_ids)
          ? [...data.section.assistant_ids]
          : (data.section.assistant_id ? [data.section.assistant_id] : [])
      }

    } catch (err) {
      console.error('Error loading section data:', err)
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  /**
   * Load professors and assistants from users API
   * Professors are filtered by role (role_id=2), assistants can be any user
   * @param {string} token - Authentication token
   */
  const loadProfessorsAndAssistants = async (token) => {
    try {
      // ID del rol profesor en la base de datos
      const PROFESSOR_ROLE_ID = 2

      // Obtener solo profesores (filtrado por role_id)
      const professorsResponse = await authenticatedFetch(
        `${API_BASE_URL}/users?role_id=${PROFESSOR_ROLE_ID}&limit=1000`
      )

      if (professorsResponse?.ok) {
        const data = await professorsResponse.json()
        professors.value = data.data || []
      }

      // Obtener todos los usuarios para ayudantes (sin filtro)
      const allUsersResponse = await authenticatedFetch(`${API_BASE_URL}/users?limit=1000`)

      if (allUsersResponse?.ok) {
        const data = await allUsersResponse.json()
        assistants.value = data.data || []
      }
    } catch (err) {
      console.error('Error loading professors and assistants:', err)
    }
  }

  /**
   * Update assignment for professors and assistants (supports multiple)
   * @param {string} sectionId - The section ID
   * @param {string|object} type - 'professor', 'assistant', 'staff' or { type, action }
   * @param {string} [action] - 'add', 'remove', 'update'
   * @returns {Promise<object>}
   */
  const updateAssignment = async (sectionId, type = 'staff', action = 'update') => {
    try {
      if (!sectionId) {
        throw new Error('Section ID is required')
      }

      let targetType = 'staff'
      let actionType = action || 'update'

      if (typeof type === 'object' && type !== null) {
        targetType = type.type || 'staff'
        actionType = type.action || actionType
      } else if (typeof type === 'string') {
        targetType = type
      }

      const pIds = assignments.professor_ids || []
      const aIds = assignments.assistant_ids || []

      const payload = {
        professor_id: pIds.length > 0 ? pIds[0] : (assignments.professor_id || null),
        assistant_id: aIds.length > 0 ? aIds[0] : (assignments.assistant_id || null),
        professor_ids: pIds,
        assistant_ids: aIds
      }

      const response = await authenticatedFetch(
        `${API_BASE_URL}/sections/${sectionId}/staff`,
        {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(payload)
        }
      )

      if (!response?.ok) {
        throw new Error('Failed to update staff assignment')
      }

      const resData = await response.json()
      if (resData.data) {
        assignments.professor_ids = resData.data.professor_ids || []
        assignments.assistant_ids = resData.data.assistant_ids || []
        assignments.professor_id = resData.data.professor_id || ''
        assignments.assistant_id = resData.data.assistant_id || ''
      }

      let label = 'Personal docente'
      if (targetType === 'professor') label = 'Profesor'
      else if (targetType === 'assistant') label = 'Ayudante'

      let message = `${label} actualizado correctamente`
      if (actionType === 'add') {
        message = `${label} asignado correctamente`
      } else if (actionType === 'remove') {
        message = `${label} quitado correctamente`
      }

      return {
        success: true,
        message
      }
    } catch (err) {
      console.error(`Error updating ${type} assignment:`, err)
      throw new Error(`Error al actualizar asignaciones del equipo docente`)
    }
  }

  return {
    // State
    sectionData,
    courseData,
    semesterInfo,
    professors,
    assistants,
    assignments,
    loading,
    error,

    // Methods
    loadSectionData,
    loadProfessorsAndAssistants,
    updateAssignment
  }
}
