import { ref } from 'vue'
import { useAuth } from '~/composables/useAuth'

export const useSectionStudents = () => {
  const API_BASE_URL = 'http://127.0.0.1:3000/api/v1'
  const { authenticatedFetch } = useAuth()

  // State
  const students = ref([])
  const loadingStudents = ref(false)
  const studentToEdit = ref(null)
  const studentToDelete = ref(null)

  // Modal states
  const showEditStudentModal = ref(false)
  const showDeleteConfirmModal = ref(false)

  /**
   * Load students for a section
   * @param {string} sectionId - The section ID
   * @param {string} token - Authentication token
   */
  const loadStudents = async (sectionId, token) => {
    try {
      loadingStudents.value = true
      console.log('Loading students for section:', sectionId)

      // Nueva llamada unificada que reemplaza las 3 llamadas anteriores
      const response = await authenticatedFetch(`${API_BASE_URL}/sections/${sectionId}/students-complete`)

      if (!response) return // Token expirado, ya fue manejado por authenticatedFetch

      console.log('Students API response status:', response.status)

      if (!response.ok) {
        if (response.status === 404) {
          students.value = []
          console.log('No students found for section')
          return
        }
        throw new Error('Error al cargar los estudiantes de la sección')
      }

      const data = await response.json()
      console.log('Students API response data:', data)

      if (!data.success || !data.students || data.students.length === 0) {
        students.value = []
        console.log('No students in response')
        return
      }

      // Mapear los datos de estudiantes que ya vienen con la información completa
      students.value = data.students.map(userData => ({
        id: userData.id ?? null,
        rut: userData.rut ?? 'N/A',
        nombres: userData.nombres ?? '',
        paterno: userData.paterno ?? '',
        materno: userData.materno ?? '',
        carrera: userData.career_code ?? 'N/A',
        full_name: userData.full_name ?? `${userData.nombres ?? ''} ${userData.paterno ?? ''} ${userData.materno ?? ''}`.trim(),
        first_name: userData.nombres ?? (userData.full_name ? userData.full_name.split(' ')[0] : 'N/A'),
        last_name: `${userData.paterno ?? ''} ${userData.materno ?? ''}`.trim() || '',
        email: userData.email ?? 'N/A',
        eneatype: userData.eneatype ?? null,
        ocean_vector: userData.ocean_vector ?? null
      }))

      console.log('Loaded students:', students.value.length)
    } catch (err) {
      console.error('Error loading students:', err)
      students.value = []
    } finally {
      loadingStudents.value = false
    }
  }

  /**
   * Get user by email
   * @param {string} email - User email
   * @param {string} token - Authentication token
   * @returns {Promise<Object|null>} User data or null if not found
   */
  const getUserByEmailAPI = async (email, token) => {
    try {
      const response = await authenticatedFetch(`${API_BASE_URL}/users/by-email?email=${encodeURIComponent(email)}`)

      if (!response) return // Token expirado, ya fue manejado por authenticatedFetch

      if (response.ok) {
        const result = await response.json()
        return result.data
      } else if (response.status === 404) {
        return null
      } else {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Error al buscar usuario por email')
      }
    } catch (error) {
      console.error('Error buscando usuario por email:', error)
      if (error.message.includes('404') || error.message.includes('not found')) {
        return null
      }
      throw error
    }
  }

  /**
   * Get user by RUT
   * @param {string} rut - User RUT
   * @returns {Promise<Object|null>} User data or null if not found
   */
  const getUserByRutAPI = async (rut) => {
    try {
      const response = await authenticatedFetch(`${API_BASE_URL}/users/by-rut?rut=${encodeURIComponent(rut)}`)

      if (!response) return null // Token expirado

      if (response.ok) {
        const result = await response.json()
        return result.data
      } else if (response.status === 404) {
        return null
      } else {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Error al buscar usuario por RUT')
      }
    } catch (error) {
      console.error('Error buscando usuario por RUT:', error)
      if (error.message.includes('404') || error.message.includes('not found')) {
        return null
      }
      throw error
    }
  }

  /**
   * Create a new user in the auth system
   * @param {Object} studentData - Student data
   * @param {string} token - Authentication token
   * @returns {Promise<Object>} Created user data
   */
  const createUserInAuth = async (studentData, token) => {
    try {
      const studentRoleId = 3
      // Generate a random temporary password (user must reset on first login)
      const tempPassword = Array.from(
        crypto.getRandomValues(new Uint8Array(16)),
        b => b.toString(16).padStart(2, '0')
      ).join('')

      const requestBody = {
        nombres: studentData.nombres || '',
        paterno: studentData.paterno || '',
        materno: studentData.materno || '',
        rut: studentData.rut,
        email: studentData.email,
        password: tempPassword,
        role_id: studentRoleId,
        career_code: studentData.carrera || studentData.career_code || null
      }

      const response = await authenticatedFetch(`${API_BASE_URL}/users`, {
        method: 'POST',
        body: JSON.stringify(requestBody)
      })

      if (!response) return // Token expirado, ya fue manejado por authenticatedFetch

      if (!response.ok) {
        let errorData
        try {
          errorData = await response.json()
        } catch (parseError) {
          console.error('Error parsing response body:', parseError)
          throw new Error(`HTTP ${response.status}: Failed to create user`)
        }

        throw new Error(errorData.message || errorData.error || `HTTP ${response.status}: Failed to create user`)
      }

      const result = await response.json()
      return result.data
    } catch (error) {
      console.error('Error creando usuario:', error)
      throw error
    }
  }

  

  /**
   * Enroll student in a section
   * @param {number} studentId - Student ID
   * @param {number} sectionId - Section ID
   * @param {string} token - Authentication token
   * @returns {Promise<Object>} Enrollment result
   */
  const enrollStudentInSection = async (studentId, sectionId, token) => {
    try {
      const response = await authenticatedFetch(`${API_BASE_URL}/student-sections`, {
        method: 'POST',
        body: JSON.stringify({
          student_id: Number.parseInt(studentId),
          section_id: Number.parseInt(sectionId)
        })
      })

      if (!response) return // Token expirado, ya fue manejado por authenticatedFetch

      if (!response.ok) {
        const errorData = await response.json()

        if (response.status === 409 || errorData.error?.includes('duplicate') || errorData.error?.includes('already')) {
          return { message: 'Student already enrolled' }
        }

        throw new Error(errorData.error || 'Error al inscribir estudiante en la sección')
      }

      return await response.json()
    } catch (error) {
      console.error('Error inscribiendo estudiante:', error)
      throw error
    }
  }

  /**
   * Handle adding a single student
   * @param {Object} studentData - Student data
   * @param {number} sectionId - Section ID
   * @param {string} token - Authentication token
   */
  const handleStudentAdded = async (studentData, sectionId, token) => {
    try {
      if (!studentData.email || !studentData.nombres || !studentData.paterno) {
        throw new Error('Faltan datos obligatorios: email, nombres y apellido paterno son requeridos')
      }

      // Buscar primero por RUT si está disponible, de lo contrario por email
      let user = null
      if (studentData.rut) {
        user = await getUserByRutAPI(studentData.rut)
      }
      if (!user) {
        user = await getUserByEmailAPI(studentData.email, token)
      }

      if (user) {
        // El usuario ya existe en la BD: usamos sus datos tal cual están
        // No intentamos actualizar para evitar conflictos con campos únicos (email, rut)
        console.log(`Usuario encontrado en BD (id=${user.id}), usando datos existentes`)
      } else {
        const createData = {
          email: studentData.email,
          fullName: `${studentData.nombres} ${studentData.paterno} ${studentData.materno || ''}`.trim(),
          nombres: studentData.nombres,
          paterno: studentData.paterno,
          materno: studentData.materno || '',
          rut: studentData.rut,
          career_code: studentData.carrera || null
        }

        user = await createUserInAuth(createData, token)
      }

      await enrollStudentInSection(user.id, sectionId, token)
      await loadStudents(sectionId, token)

      return { success: true, message: `Estudiante ${studentData.nombres} ${studentData.paterno} añadido exitosamente` }

    } catch (error) {
      console.error('Error al añadir estudiante:', error)
      throw error
    }
  }

  /**
   * Handle uploading multiple students
   * @param {Array} studentsData - Array of student data
   * @param {number} sectionId - Section ID
   * @param {string} token - Authentication token
   */

  // Helper: look up a student by RUT first, then email as fallback
  const findExistingUser = async (studentData, token) => {
    let user = await getUserByRutAPI(studentData.rut)
    if (!user) {
      user = await getUserByEmailAPI(studentData.email, token)
    }
    return user
  }

  // Helper: enroll an existing user in a section and track result in counts
  const processExistingUserEnroll = async (userId, sectionId, counts) => {
    try {
      const enrollResult = await enrollStudentInSection(userId, sectionId)
      if (enrollResult?.message === 'Student already enrolled') {
        counts.skippedCount++
      } else {
        counts.successCount++
      }
    } catch (enrollError) {
      if (enrollError.message.includes('duplicate') || enrollError.message.includes('already')) {
        counts.skippedCount++
      } else {
        throw enrollError
      }
    }
  }

  // Helper: create new user and enroll during upload
  const processNewUserUpload = async (studentData, sectionId, token, counts, errors) => {
    try {
      const newUser = await createUserInAuth(studentData, token)
      counts.createdCount++
      try {
        await enrollStudentInSection(newUser.id, sectionId, token)
        counts.successCount++
      } catch (enrollError) {
        counts.errorCount++
        errors.push(`${studentData.fullName} (${studentData.email}): Usuario creado pero error al inscribir - ${enrollError.message}`)
      }
    } catch (createError) {
      counts.errorCount++
      errors.push(`${studentData.fullName} (${studentData.email}): Error al crear usuario - ${createError.message}`)
    }
  }

  const handleStudentsUploaded = async (studentsData, sectionId, token) => {
    const counts = { successCount: 0, errorCount: 0, skippedCount: 0, createdCount: 0, updatedCount: 0 }
    const errors = []

    for (const studentData of studentsData) {
      try {
        if (!studentData.email || !studentData.fullName || !studentData.rut) {
          counts.errorCount++
          errors.push(`Estudiante con datos incompletos: ${JSON.stringify(studentData)}`)
          continue
        }

        // Buscar primero por RUT (identificador único oficial del estudiante),
        // con fallback a email si no se encuentra
        const user = await findExistingUser(studentData, token)

        if (user) {
          // El usuario ya existe: no actualizar para evitar conflictos con campos únicos
          // Simplemente inscribirlo en la sección con los datos que ya tiene en la BD
          console.log(`Usuario ya existe (rut=${studentData.rut}, id=${user.id}), usando datos de BD`)
          await processExistingUserEnroll(user.id, sectionId, counts)
        } else {
          await processNewUserUpload(studentData, sectionId, token, counts, errors)
        }
      } catch (error) {
        counts.errorCount++
        const errorMessage = error.message || 'Error desconocido'
        errors.push(`${studentData.fullName || 'Estudiante desconocido'} (${studentData.email || studentData.rut}): ${errorMessage}`)
      }
    }

    await loadStudents(sectionId, token)

    return { ...counts, errors }
  }

  /**
   * Open edit student modal
   * @param {Object} student - Student to edit
   */
  const openEditStudent = (student) => {
    studentToEdit.value = {
      id: student.id,
      nombres: student.nombres || student.first_name || '',
      paterno: student.paterno || student.last_name || '',
      materno: student.materno || '',
      rut: student.rut || '',
      email: student.email || '',
      carrera: student.carrera || student.career_code || '',
      role_id: student.role_id || student.roleId || 3
    }
    showEditStudentModal.value = true
  }

  /**
   * Close edit student modal
   */
  const closeEditStudent = () => {
    showEditStudentModal.value = false
    studentToEdit.value = null
  }

  /**
   * Save edited student
   * @param {number} sectionId - Section ID
   * @param {string} token - Authentication token
   */
  const saveEditStudent = async (sectionId, token) => {
    try {
      const response = await authenticatedFetch(`${API_BASE_URL}/users/${studentToEdit.value.id}`, {
        method: 'PUT',
        body: JSON.stringify({
          nombres: studentToEdit.value.nombres,
          paterno: studentToEdit.value.paterno,
          materno: studentToEdit.value.materno,
          rut: studentToEdit.value.rut,
          email: studentToEdit.value.email,
          role_id: studentToEdit.value.role_id,
          career_code: studentToEdit.value.carrera
        })
      })

      if (!response) return // Token expirado, ya fue manejado por authenticatedFetch

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.message || 'Error al actualizar el estudiante')
      }

      await loadStudents(sectionId, token)
      closeEditStudent()

      return { success: true, message: 'Estudiante actualizado correctamente' }
    } catch (error) {
      console.error('Error updating student:', error)
      throw error
    }
  }

  /**
   * Open delete confirmation modal
   * @param {Object} student - Student to delete
   */
  const openDeleteConfirm = (student) => {
    studentToDelete.value = student
    showDeleteConfirmModal.value = true
  }

  /**
   * Close delete confirmation modal
   */
  const closeDeleteConfirm = () => {
    showDeleteConfirmModal.value = false
    studentToDelete.value = null
  }

  /**
   * Confirm and delete student
   * @param {number} sectionId - Section ID
   * @param {string} token - Authentication token
   */
  const confirmDeleteStudent = async (sectionId, token) => {
    try {
      const studentId = studentToDelete.value.id

      const response = await authenticatedFetch(`${API_BASE_URL}/sections/${sectionId}/students/${studentId}`, {
        method: 'DELETE'
      })

      if (!response) return // Token expirado, ya fue manejado por authenticatedFetch

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || 'Error al eliminar el estudiante de la sección')
      }

      await loadStudents(sectionId, token)
      closeDeleteConfirm()

      return { success: true, message: 'Estudiante eliminado de la sección correctamente' }
    } catch (error) {
      console.error('Error removing student from section:', error)
      throw error
    }
  }

  /**
   * Search for users
   * @param {string} query - Search query
   * @param {number} roleId - Role filter
   * @param {string} token - Auth token
   */
  const searchUsers = async (query, roleId, token) => {
    try {
      const response = await authenticatedFetch(`${API_BASE_URL}/users?q=${encodeURIComponent(query)}&role_id=${roleId}`)
      if (!response?.ok) return []
      const result = await response.json()
      return result.data || []
    } catch (error) {
      console.error('Error searching users:', error)
      return []
    }
  }

  /**
   * Enroll a student in a section
   * @param {number} studentId - Student ID
   * @param {number} sectionId - Section ID
   * @param {string} token - Auth token
   */
  const enrollStudent = async (studentId, sectionId, token) => {
    try {
      const result = await enrollStudentInSection(studentId, sectionId, token)
      await loadStudents(sectionId, token)
      return result
    } catch (error) {
      console.error('Error enrolling student:', error)
      throw error
    }
  }

  return {
    // State
    students,
    loadingStudents,
    studentToEdit,
    studentToDelete,
    showEditStudentModal,
    showDeleteConfirmModal,

    // Methods
    loadStudents,
    handleStudentAdded,
    handleStudentsUploaded,
    searchUsers,
    enrollStudent,
    openEditStudent,
    closeEditStudent,
    saveEditStudent,
    openDeleteConfirm,
    closeDeleteConfirm,
    confirmDeleteStudent
  }
}
