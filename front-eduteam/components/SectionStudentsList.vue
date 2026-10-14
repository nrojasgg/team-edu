<template>
  <div class="students-list">
    <div class="students-header">
      <h3>Lista de estudiantes</h3>
      <div v-if="!readOnly" class="students-actions">
        <button @click="$emit('upload-students')" class="upload-students-btn">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="7,10 12,15 17,10"/>
            <line x1="12" y1="15" x2="12" y2="3"/>
          </svg>
          Cargar Estudiantes
        </button>
        <button @click="$emit('add-student')" class="add-student-btn">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <path d="M12 8v8"/>
            <path d="M8 12h8"/>
          </svg>
          Agregar Estudiante
        </button>
        <button @click="$emit('generate-invitation-code')" class="invitation-code-btn">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
            <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
          </svg>
          Código de Invitación
        </button>
      </div>
    </div>

    <div v-if="!loadingStudents && students.length > 0" class="students-search-bar">
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8"/>
        <path d="m21 21-4.35-4.35"/>
      </svg>
      <input
        v-model="searchQuery"
        type="text"
        placeholder="Buscar por nombre o RUT..."
        class="search-input"
      />
      <button v-if="searchQuery" @click="searchQuery = ''" class="search-clear-btn">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M18 6 6 18M6 6l12 12"/>
        </svg>
      </button>
    </div>

    <div v-if="loadingStudents" class="loading-students">
      Cargando estudiantes...
    </div>
    
    <div v-else-if="students.length === 0" class="no-students">
      <p>No hay estudiantes inscritos en esta coordinación</p>
    </div>

    <div v-else-if="filteredStudents.length === 0" class="no-students">
      <p>No se encontraron estudiantes para "{{ searchQuery }}"</p>
    </div>
    
    <div v-else class="students-table-container">
      <table class="students-table">
        <thead>
          <tr>
            <th>RUT</th>
            <th>Nombres</th>
            <th>Paterno</th>
            <th>Materno</th>
            <th>Carrera</th>
            <th>Correo Electrónico</th>
            <th v-if="!readOnly">Acciones</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="student in filteredStudents" :key="student.id">
            <td>{{ student.rut || 'N/A' }}</td>
            <td>{{ student.nombres || student.first_name || 'N/A' }}</td>
            <td>{{ student.paterno || student.last_name || '' }}</td>
            <td>{{ student.materno || '' }}</td>
            <td>{{ student.carrera || 'N/A' }}</td>
            <td>{{ student.email || 'N/A' }}</td>
            <td v-if="!readOnly" class="student-actions">
              <button class="action-btn edit-btn" @click="$emit('edit-student', student)" title="Editar estudiante">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M12 20h9"/>
                  <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/>
                </svg>
              </button>
              <button class="action-btn delete-btn" @click="$emit('delete-student', student)" title="Eliminar estudiante">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="3,6 5,6 21,6"/>
                  <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2 2h4a2 2 0 0 1 2 2v2"/>
                  <path d="M10 11V17"/>
                  <path d="M14 11V17"/>
                </svg>
              </button>
              <button v-if="isAdmin" @click="$emit('impersonate-user', student.id)" class="impersonate-btn">Simular</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal de edición de estudiante -->
    <div v-if="showEditStudentModal" class="modal-overlay">
      <div class="modal">
        <h3>Editar Estudiante</h3>
        <form @submit.prevent="saveEditStudent">
          <div class="form-group">
            <label for="ssl-nombres">Nombres:</label>
            <input id="ssl-nombres" v-model="studentToEdit.nombres" type="text" required />
          </div>
          <div class="form-group">
            <label for="ssl-paterno">Apellido Paterno:</label>
            <input id="ssl-paterno" v-model="studentToEdit.paterno" type="text" required />
          </div>
          <div class="form-group">
            <label for="ssl-materno">Apellido Materno:</label>
            <input id="ssl-materno" v-model="studentToEdit.materno" type="text" />
          </div>
          <div class="form-group">
            <label for="ssl-rut">RUT:</label>
            <input id="ssl-rut" v-model="studentToEdit.rut" type="text" required />
          </div>
          <div class="form-group">
            <label for="ssl-email">Correo Electrónico:</label>
            <input id="ssl-email" v-model="studentToEdit.email" type="email" required />
          </div>
          <div class="form-group">
            <label for="ssl-carrera">Carrera:</label>
            <input id="ssl-carrera" v-model="studentToEdit.carrera" type="text" />
          </div>
          <div class="modal-actions">
            <button type="submit" class="action-btn save-btn">Guardar Cambios</button>
            <button type="button" class="action-btn cancel-btn" @click="closeEditStudent">Cancelar</button>
          </div>
        </form>
      </div>
    </div>

    <!-- Modal de confirmación de eliminación -->
    <div v-if="showDeleteConfirmModal" class="modal-overlay">
      <div class="modal">
        <h3>¿Confirmar eliminación?</h3>
        <p>¿Estás seguro de que deseas eliminar al estudiante <strong>{{ studentToDelete?.nombres }} {{ studentToDelete?.paterno }}</strong> de esta coordinación ?</p>
        <p class="warning-text">Esta acción no se puede deshacer.</p>
        <div class="modal-actions">
          <button class="action-btn delete-btn" @click="handleDeleteStudent">Sí, Eliminar</button>
          <button class="action-btn cancel-btn" @click="closeDeleteConfirm">Cancelar</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  students: Array,
  loadingStudents: Boolean,
  isAdmin: Boolean,
  readOnly: {
    type: Boolean,
    default: false
  }
})

const searchQuery = ref('')

const filteredStudents = computed(() => {
  if (!searchQuery.value.trim()) return props.students
  const q = searchQuery.value.toLowerCase().trim()
  return props.students.filter(s => {
    const nombre = `${s.nombres || s.first_name || ''} ${s.paterno || s.last_name || ''} ${s.materno || ''}`.toLowerCase()
    const rut = (s.rut || '').toLowerCase()
    return nombre.includes(q) || rut.includes(q)
  })
})

const emit = defineEmits([
  'upload-students',
  'add-student', 
  'edit-student',
  'delete-student',
  'impersonate-user',
  'save-edit-student',
  'delete-student-confirmed',
  'generate-invitation-code'
])

// Estados para edición y eliminación de estudiantes
const showEditStudentModal = ref(false)
const showDeleteConfirmModal = ref(false)
const studentToEdit = ref(null)
const studentToDelete = ref(null)

// Funciones para manejar la edición de estudiantes
const openEditStudent = (student) => {
  studentToEdit.value = {
    id: student.id,
    nombres: student.nombres || student.first_name || '',
    paterno: student.paterno || student.last_name || '',
    materno: student.materno || '',
    rut: student.rut || '',
    email: student.email || '',
    carrera: student.carrera || student.career_code || ''
  }
  showEditStudentModal.value = true
}

const closeEditStudent = () => {
  showEditStudentModal.value = false
  studentToEdit.value = null
}

const saveEditStudent = async () => {
  emit('save-edit-student', studentToEdit.value)
  closeEditStudent()
}

// Funciones para manejar la eliminación de estudiantes
const confirmDeleteStudent = (student) => {
  studentToDelete.value = student
  showDeleteConfirmModal.value = true
}

const closeDeleteConfirm = () => {
  showDeleteConfirmModal.value = false
  studentToDelete.value = null
}

const handleDeleteStudent = async () => {
  emit('delete-student-confirmed', studentToDelete.value)
  closeDeleteConfirm()
}

// Exponer funciones para uso desde el componente padre
defineExpose({
  openEditStudent,
  confirmDeleteStudent
})
</script>

<style scoped>
/* Lista de estudiantes */
.students-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.students-header h3 {
  margin: 0;
  color: var(--color-dark, #394049);
  font-size: 1.5rem;
}

.students-actions {
  display: flex;
  gap: 1rem;
}

/* Responsive para móvil */
@media (max-width: 768px) {
  .students-header {
    flex-direction: column;
    gap: 1rem;
    text-align: center;
  }
  
  .students-actions {
    justify-content: center;
  }
}

.upload-students-btn,
.add-student-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  border: none;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s;
}

.upload-students-btn {
  background-color: var(--color-primary, #00A499);
  color: white;
}

.upload-students-btn:hover {
  background-color: var(--color-primary-dark, #008a7f);
  box-shadow: 0 4px 12px rgba(0, 164, 153, 0.3);
}

.add-student-btn {
  background-color: var(--color-accent, #EA7600);
  color: white;
}

.add-student-btn:hover {
  background-color: #d06a00;
  box-shadow: 0 4px 12px rgba(234, 118, 0, 0.3);
}

.invitation-code-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  border: none;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s;
  background-color: var(--color-secondary, #7C3AED);
  color: white;
}

.invitation-code-btn:hover {
  background-color: var(--color-secondary-dark, #6D28D9);
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.3);
}

.loading-students {
  text-align: center;
  padding: 3rem;
  color: #94a3b8;
}

.students-search-bar {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.6rem 1rem;
  background: var(--color-surface, #ffffff);
  border: 1.5px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-md, 10px);
  margin-bottom: 1rem;
  color: #94a3b8;
}

.students-search-bar:focus-within {
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.1);
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 0.9rem;
  color: var(--color-dark, #394049);
  background: transparent;
}

.search-input::placeholder {
  color: #94a3b8;
}

.search-clear-btn {
  border: none;
  background: none;
  cursor: pointer;
  color: #94a3b8;
  padding: 0.2rem;
  display: flex;
  align-items: center;
  transition: color 0.15s;
}

.search-clear-btn:hover {
  color: var(--color-dark, #394049);
}

.no-students {
  text-align: center;
  padding: 3rem;
  color: #94a3b8;
}

.students-table-container {
  background: var(--color-surface, #ffffff);
  border-radius: var(--radius-md, 10px);
  overflow: hidden;
  box-shadow: var(--shadow-sm, 0 1px 3px rgba(0,0,0,0.08));
}

.students-table {
  width: 100%;
  border-collapse: collapse;
}

.students-table th,
.students-table td {
  padding: 1rem;
  text-align: left;
  border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.students-table th {
  background-color: var(--color-background, #f4f6f9);
  font-weight: 600;
  color: var(--color-dark, #394049);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.students-table tr:hover {
  background-color: var(--color-background, #f4f6f9);
}

.student-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.875rem;
}

.action-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.edit-btn {
  background-color: var(--color-primary, #00A499);
  color: white;
}

.edit-btn:hover {
  background-color: var(--color-primary-dark, #008a7f);
}

.delete-btn {
  background-color: var(--color-danger, #d7373f);
  color: white;
}

.delete-btn:hover {
  background-color: #b82e35;
}

.save-btn {
  background-color: var(--color-primary, #00A499);
  color: white;
  width: auto;
  padding: 0.5rem 1rem;
}

.save-btn:hover {
  background-color: var(--color-primary-dark, #008a7f);
}

.cancel-btn {
  background-color: #475569;
  color: white;
  width: auto;
  padding: 0.5rem 1rem;
}

.cancel-btn:hover {
  background-color: #475569;
}

.impersonate-btn {
  background: var(--color-accent, #b55600);
  color: white;
  border: none;
  border-radius: var(--radius-sm, 6px);
  padding: 0.5rem 1rem;
  cursor: pointer;
  font-size: 0.9rem;
  margin-left: 0.5rem;
  transition: background 0.2s;
}

.impersonate-btn:hover {
  background: #d06a00;
}

/* Modal styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--color-surface, #ffffff);
  border-radius: var(--radius-lg, 14px);
  box-shadow: var(--shadow-xl, 0 20px 25px rgba(0,0,0,0.12));
  width: 90vw;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
  padding: 2rem;
}

.modal h3 {
  margin: 0 0 1.5rem 0;
  color: var(--color-dark, #394049);
  font-size: 1.25rem;
  font-weight: 600;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: var(--color-text-secondary, #64748b);
  font-weight: 500;
}

.form-group input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-sm, 6px);
  font-size: 1rem;
  transition: border-color 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.1);
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 2rem;
}

.warning-text {
  color: var(--color-danger, #d7373f);
  font-size: 0.875rem;
  margin-top: 0.5rem;
}
</style>
