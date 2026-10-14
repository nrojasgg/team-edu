<template>
  <div>
    <SectionStudentsList 
      :students="students"
      :loading-students="loadingStudents"
      :read-only="user?.role_name === 'estudiante'"
      @add-student="openSelectionModal"
      @upload-students="openUploadStudentsModal"
      @edit-student="handleEditStudent"
      @delete-student="handleDeleteStudent"
      @generate-invitation-code="handleGenerateInvitationCode"
    />

    <!-- Modales -->
    <AddStudentSelectionModal
      :is-open="showSelectionModal"
      @close="showSelectionModal = false"
      @select-existing="openSelectExistingModal"
      @select-new="openAddStudentModal"
    />

    <SelectExistingStudentModal
      :is-open="showSelectExistingModal"
      :section-id="route.params.id"
      @close="showSelectExistingModal = false"
      @student-selected="handleExistingStudentSelected"
    />

    <AddStudentModal 
      :is-open="showAddModal"
      @close="closeAddStudentModal"
      @student-added="handleStudentAdded"
    />
    
    <UploadStudentsModal 
      :is-open="showUploadModal"
      @close="closeUploadStudentsModal"
      @students-uploaded="handleStudentsUploaded"
    />

    <!-- Modal de edición de estudiante -->
    <div v-if="showEditStudentModal" class="modal-overlay">
      <div class="modal">
        <h3>Editar Estudiante</h3>
        <form @submit.prevent="saveEditStudent">
          <div class="form-group">
            <label for="list-nombres">Nombres:</label>
            <input id="list-nombres" v-model="studentToEdit.nombres" type="text" required />
          </div>
          <div class="form-group">
            <label for="list-paterno">Apellido Paterno:</label>
            <input id="list-paterno" v-model="studentToEdit.paterno" type="text" required />
          </div>
          <div class="form-group">
            <label for="list-materno">Apellido Materno:</label>
            <input id="list-materno" v-model="studentToEdit.materno" type="text" />
          </div>
          <div class="form-group">
            <label for="list-rut">RUT:</label>
            <input id="list-rut" v-model="studentToEdit.rut" type="text" required />
          </div>
          <div class="form-group">
            <label for="list-email">Correo Electrónico:</label>
            <input id="list-email" v-model="studentToEdit.email" type="email" required />
          </div>
          <div class="form-group">
            <label for="list-carrera">Carrera:</label>
            <input id="list-carrera" v-model="studentToEdit.carrera" type="text" />
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
          <button class="action-btn delete-btn" @click="confirmDeleteStudent">Sí, Eliminar</button>
          <button class="action-btn cancel-btn" @click="closeDeleteConfirm">Cancelar</button>
        </div>
      </div>
    </div>
  </div>

  <!-- Info Modal -->
  <InfoModal
    :show="showInfoModal"
    :title="infoTitle"
    :message="infoMessage"
    @close="showInfoModal = false"
  />

  <div v-if="showInvitationCodesPanel" class="codes-panel-overlay" @click.self="showInvitationCodesPanel = false">
    <InvitationCodesPanel 
      ref="invitationCodesPanelRef"
      :sectionId="route.params.id" 
      :isAdmin="user?.role_name !== 'estudiante'"
      @close="showInvitationCodesPanel = false"
      @create="handleOpenCreateInvitationModal"
    />
  </div>

  <CreateInvitationCodeModal
    v-if="showCreateInvitationCodeModal"
    :is-open="showCreateInvitationCodeModal"
    :section-id="route.params.id"
    :submitting="createInvitationCodeSubmitting"
    @close="showCreateInvitationCodeModal = false"
    @submit="handleCreateInvitationCode"
  />
</template>

<script setup>
import { ref, onMounted } from 'vue'
import AddStudentModal from '~/components/AddStudentModal.vue'
import AddStudentSelectionModal from '~/components/AddStudentSelectionModal.vue'
import SelectExistingStudentModal from '~/components/SelectExistingStudentModal.vue'
import UploadStudentsModal from '~/components/UploadStudentsModal.vue'
import SectionStudentsList from '~/components/SectionStudentsList.vue'
import InvitationCodesPanel from '~/components/InvitationCodesPanel.vue'
import CreateInvitationCodeModal from '~/components/CreateInvitationCodeModal.vue'
import InfoModal from '~/components/InfoModal.vue'
import { useSectionStudents } from '~/composables/useSectionStudents'

const props = defineProps({
  sectionData: Object,
  courseData: Object,
  semesterInfo: String,
  user: Object
})

const route = useRoute()
const { getToken } = useAuth()

// Initialize composable
const studentsModule = useSectionStudents()

// Local state
const showSelectionModal = ref(false)
const showSelectExistingModal = ref(false)
const showAddModal = ref(false)
const showUploadModal = ref(false)

// Info modal state
const showInfoModal = ref(false)
const infoTitle = ref('')
const infoMessage = ref('')

// Invitation codes state
const showInvitationCodesPanel = ref(false)
const showCreateInvitationCodeModal = ref(false)
const createInvitationCodeSubmitting = ref(false)
const invitationCodesPanelRef = ref(null)

// Helper function to show info modal
const showInfo = (title, message) => {
  infoTitle.value = title
  infoMessage.value = message
  showInfoModal.value = true
}

// Destructure composable states
const {
  students,
  loadingStudents,
  studentToEdit,
  studentToDelete,
  showEditStudentModal,
  showDeleteConfirmModal
} = studentsModule

// Handlers
const openSelectionModal = () => {
  showSelectionModal.value = true
}

const openSelectExistingModal = () => {
  showSelectExistingModal.value = true
}

const openAddStudentModal = () => {
  showAddModal.value = true
}

const closeAddStudentModal = () => {
  showAddModal.value = false
}

const openUploadStudentsModal = () => {
  showUploadModal.value = true
}

const closeUploadStudentsModal = () => {
  showUploadModal.value = false
}

const handleEditStudent = (student) => {
  studentsModule.openEditStudent(student)
}

const handleDeleteStudent = (student) => {
  studentsModule.openDeleteConfirm(student)
}

const closeEditStudent = () => {
  studentsModule.closeEditStudent()
}

const saveEditStudent = async () => {
  try {
    const sectionId = route.params.id
    const token = getToken()
    const result = await studentsModule.saveEditStudent(sectionId, token)
    showInfo('Éxito', result.message)
  } catch (error) {
    showInfo('Error', 'Error al actualizar el estudiante: ' + error.message)
  }
}

const closeDeleteConfirm = () => {
  studentsModule.closeDeleteConfirm()
}

const confirmDeleteStudent = async () => {
  try {
    const sectionId = route.params.id
    const token = getToken()
    const result = await studentsModule.confirmDeleteStudent(sectionId, token)
    showInfo('Éxito', result.message)
  } catch (error) {
    showInfo('Error', 'Error al eliminar el estudiante: ' + error.message)
  }
}

const handleStudentAdded = async (studentData) => {
  try {
    const sectionId = Number.parseInt(route.params.id)
    const token = getToken()
    const result = await studentsModule.handleStudentAdded(studentData, sectionId, token)
    showInfo('Éxito', result.message)
  } catch (error) {
    showInfo('Error', `Error al añadir estudiante: ${error.message}`)
  }
}

const handleExistingStudentSelected = (student) => {
  showInfo('Éxito', `Estudiante ${student.nombres} ${student.paterno} añadido exitosamente a la sección`)
}

const handleStudentsUploaded = async (studentsData) => {
  try {
    const sectionId = Number.parseInt(route.params.id)
    const token = getToken()
    const result = await studentsModule.handleStudentsUploaded(studentsData, sectionId, token)
    
    let message = `Proceso completado:\n${result.successCount} estudiantes inscritos exitosamente`
    if (result.createdCount > 0) {
      message += `\n${result.createdCount} nuevos usuarios creados`
    }
    if (result.updatedCount > 0) {
      message += `\n${result.updatedCount} usuarios actualizados con datos del CSV`
    }
    if (result.skippedCount > 0) {
      message += `\n${result.skippedCount} estudiantes ya estaban inscritos en esta coordinación  (omitidos)`
    }
    if (result.errorCount > 0) {
      message += `\n${result.errorCount} errores encontrados:`
      message += '\n' + result.errors.slice(0, 5).join('\n')
      if (result.errors.length > 5) {
        message += `\n... y ${result.errors.length - 5} errores más`
      }
    }
    
    showInfo('Éxito', message)
  } catch (error) {
    showInfo('Error', `Error al procesar estudiantes: ${error.message}`)
  }
}

const handleGenerateInvitationCode = () => {
  showInvitationCodesPanel.value = true
}

const handleOpenCreateInvitationModal = () => {
  showCreateInvitationCodeModal.value = true
}

const handleCreateInvitationCode = async (codeData) => {
  try {
    createInvitationCodeSubmitting.value = true
    const token = getToken()
    const response = await fetch('http://127.0.0.1:3000/api/v1/invitation-codes', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(codeData)
    })
    
    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.message || 'Error al crear el código de invitación')
    }
    
    showCreateInvitationCodeModal.value = false
    if (invitationCodesPanelRef.value) {
      invitationCodesPanelRef.value.loadCodes()
    }
    showInfo('Éxito', 'Código de invitación creado exitosamente')
  } catch (error) {
    showInfo('Error', error.message)
  } finally {
    createInvitationCodeSubmitting.value = false
  }
}

// Load students on mount
onMounted(async () => {
  const sectionId = route.params.id
  const token = getToken()
  await studentsModule.loadStudents(sectionId, token)
})
</script>

<style scoped>
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

.codes-panel-overlay {
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
  z-index: 1100;
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

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-sm, 6px);
  font-size: 0.875rem;
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
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

.action-btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
}

.save-btn {
  background-color: var(--color-primary, #00A499);
  color: white;
}

.save-btn:hover {
  background-color: var(--color-primary-dark, #008a7f);
}

.cancel-btn {
  background-color: #4b5563;
  color: white;
}

.cancel-btn:hover {
  background-color: #64748b;
}

.delete-btn {
  background-color: var(--color-danger, #d7373f);
  color: white;
}

.delete-btn:hover {
  background-color: #b82e35;
}

.warning-text {
  color: var(--color-danger, #d7373f);
  font-size: 0.875rem;
  font-style: italic;
  margin-top: 0.5rem;
}
</style>
