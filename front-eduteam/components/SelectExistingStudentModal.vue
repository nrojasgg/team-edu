<template>
  <div v-if="isOpen" class="modal-overlay" @click="closeModal">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>Seleccionar Estudiante Existente</h3>
        <button @click="closeModal" class="close-btn">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="m18 6-12 12"/>
            <path d="m6 6 12 12"/>
          </svg>
        </button>
      </div>
      
      <div class="modal-body">
        <div class="search-container">
          <input 
            v-model="searchQuery" 
            type="text" 
            placeholder="Buscar por nombre, RUT o email..." 
            class="search-input"
            @input="handleSearch"
          />
          <div v-if="searching" class="search-spinner"></div>
        </div>

        <div class="results-container">
          <div v-if="results.length > 0" class="results-list">
            <div 
              v-for="student in results" 
              :key="student.id" 
              class="student-item"
              :class="{ 'selected': selectedStudent?.id === student.id }"
              @click="selectedStudent = student"
            >
              <div class="student-avatar" :style="{ backgroundColor: getAvatarColor(student.nombres) }">
                {{ student.nombres.charAt(0) }}{{ student.paterno.charAt(0) }}
              </div>
              <div class="student-info">
                <p class="student-name">{{ student.full_name || `${student.nombres} ${student.paterno}` }}</p>
                <p class="student-meta">{{ student.rut }} • {{ student.email }}</p>
              </div>
            </div>
          </div>
          <div v-else-if="searchQuery.length >= 2" class="no-results">
            No se encontraron estudiantes para "{{ searchQuery }}"
          </div>
          <div v-else class="search-placeholder">
            Ingresa al menos 2 caracteres para buscar
          </div>
        </div>

        <div class="modal-actions">
          <button @click="closeModal" class="cancel-btn">Cancelar</button>
          <button 
            :disabled="!selectedStudent || enrolling" 
            class="confirm-btn"
            @click="handleEnroll"
          >
            <span v-if="enrolling">Agregando...</span>
            <span v-else>Agregar a la Sección</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  isOpen: Boolean,
  sectionId: [String, Number]
})

const emit = defineEmits(['close', 'student-selected'])

const searchQuery = ref('')
const results = ref([])
const searching = ref(false)
const selectedStudent = ref(null)
const enrolling = ref(false)

const { getToken } = useAuth()
const studentsModule = useSectionStudents()

let searchTimeout = null

const handleSearch = () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  
  if (searchQuery.value.length < 2) {
    results.value = []
    return
  }

  searching.value = true
  searchTimeout = setTimeout(async () => {
    try {
      const token = getToken()
      results.value = await studentsModule.searchUsers(searchQuery.value, 3, token) // 3 is student role
    } catch (error) {
      console.error('Error searching students:', error)
    } finally {
      searching.value = false
    }
  }, 300)
}

const handleEnroll = async () => {
  if (!selectedStudent.value || enrolling.value) return
  
  enrolling.value = true
  try {
    const token = getToken()
    await studentsModule.enrollStudent(selectedStudent.value.id, props.sectionId, token)
    emit('student-selected', selectedStudent.value)
    closeModal()
  } catch (error) {
    console.error('Error enrolling student:', error)
    // Here we could emit an error event to show a notification
  } finally {
    enrolling.value = false
  }
}

const closeModal = () => {
  searchQuery.value = ''
  results.value = []
  selectedStudent.value = null
  emit('close')
}

const getAvatarColor = (name) => {
  const colors = ['#4299e1', '#48bb78', '#ed8936', '#9f7aea', '#f56565', '#38b2ac']
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.codePointAt(i) + ((hash << 5) - hash)
  }
  return colors[Math.abs(hash) % colors.length]
}

// Reset when opening
watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    searchQuery.value = ''
    results.value = []
    selectedStudent.value = null
  }
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.modal-header h3 {
  margin: 0;
  color: var(--color-dark, #394049);
  font-size: 1.5rem;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: var(--color-background, #f4f6f9);
  color: var(--color-text-secondary, #64748b);
}

.modal-body {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.search-container {
  position: relative;
  margin-bottom: 1rem;
}

.search-input {
  width: 100%;
  padding: 0.75rem 1rem;
  padding-right: 3rem;
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: 8px;
  font-size: 1rem;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.1);
}

.search-spinner {
  position: absolute;
  right: 1rem;
  top: 50%;
  transform: translateY(-50%);
  width: 1.25rem;
  height: 1.25rem;
  border: 2px solid var(--color-border, #dfe4ea);
  border-top-color: var(--color-primary, #00A499);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: translateY(-50%) rotate(360deg); }
}

.results-container {
  flex: 1;
  overflow-y: auto;
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: 8px;
  min-height: 250px;
  background: var(--color-background, #f4f6f9);
}

.results-list {
  padding: 0.5rem;
}

.student-item {
  display: flex;
  align-items: center;
  padding: 0.75rem;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.1s;
  background: white;
  margin-bottom: 0.5rem;
  border: 1px solid transparent;
}

.student-item:hover {
  background-color: #f1f5f9;
}

.student-item.selected {
  border-color: var(--color-primary, #00A499);
  background-color: #ebf8ff;
}

.student-avatar {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 0.875rem;
  margin-right: 1rem;
}

.student-info {
  flex: 1;
}

.student-name {
  margin: 0;
  font-weight: 500;
  color: var(--color-dark, #394049);
}

.student-meta {
  margin: 0;
  font-size: 0.75rem;
  color: #94a3b8;
}

.no-results, .search-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 250px;
  color: #94a3b8;
  font-style: italic;
  text-align: center;
  padding: 2rem;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1.5rem;
}

.cancel-btn, .confirm-btn {
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.cancel-btn {
  background: var(--color-background, #f4f6f9);
  color: var(--color-text-secondary, #64748b);
  border: 1px solid var(--color-border, #dfe4ea);
}

.cancel-btn:hover {
  background: var(--color-background, #f4f6f9);
}

.confirm-btn {
  background: var(--color-primary, #006b64);
  color: white;
}

.confirm-btn:hover:not(:disabled) {
  background: var(--color-primary-dark, #008a7f);
}

.confirm-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
