<template>
  <div class="general-info">
    <div class="info-card">
      <h3>Información de la Coordinación (Sección)</h3>
      <div class="info-grid">
        <div class="info-item">
          <span class="info-label">Nombre del Curso:</span>
          <span>{{ courseData?.name || 'N/A' }}</span>
        </div>
        
        <div class="info-item">
          <span class="info-label">Coordinación (Sección):</span>
          <span>{{ sectionData?.name || 'N/A' }}</span>
        </div>
        
        <div class="info-item">
          <span class="info-label">Semestre:</span>
          <span>{{ semesterInfo || 'N/A' }}</span>
        </div>
      </div>

      <!-- Asignación de Personal Docente -->
      <div class="assignment-section">
        <h4>Equipo Docente de la Coordinación</h4>
        
        <div class="assignment-grid">
          <!-- Bloque Profesores -->
          <div class="form-group staff-card">
            <div class="staff-header">
              <span class="form-label">
                <span class="staff-icon">👤</span> Profesores Asignados
                <span class="staff-badge">{{ selectedProfessors.length }}</span>
              </span>
            </div>

            <!-- Lista de profesores asignados -->
            <div class="staff-chips-container">
              <div 
                v-for="p in selectedProfessors" 
                :key="p.id" 
                class="staff-chip"
              >
                <div class="staff-chip-info">
                  <span class="staff-chip-name">{{ p.full_name }}</span>
                  <span v-if="p.email" class="staff-chip-email">{{ p.email }}</span>
                </div>
                <button 
                  v-if="canManageProfessors"
                  type="button"
                  class="remove-staff-btn"
                  @click="confirmRemoveProfessor(p)"
                  title="Quitar profesor"
                >
                  &times;
                </button>
              </div>

              <div v-if="selectedProfessors.length === 0" class="no-staff-notice">
                <span>Sin profesores asignados a esta coordinación</span>
              </div>
            </div>

            <!-- Dropdown para agregar profesor -->
            <div 
              v-if="canManageProfessors"
              class="searchable-select" 
              :class="{ 'open': profOpen }"
            >
              <button 
                type="button" 
                class="add-staff-trigger" 
                @click="toggleProf"
              >
                <span>+ Agregar Profesor</span>
                <span class="chevron">▼</span>
              </button>
              
              <div v-show="profOpen" class="select-dropdown">
                <div class="search-wrapper">
                  <input 
                    type="text" 
                    v-model="profSearch" 
                    placeholder="Buscar profesor por nombre o correo..." 
                    class="dropdown-search"
                    @click.stop
                    ref="profSearchInput"
                  />
                </div>
                <div class="options-list">
                  <div 
                    v-for="p in availableProfessors" 
                    :key="p.id"
                    class="option-item"
                    @click="addProfessor(p.id)"
                  >
                    <div class="option-info">
                      <span class="option-name">{{ p.full_name }}</span>
                      <span class="option-email">{{ p.email }}</span>
                    </div>
                  </div>
                  <div v-if="availableProfessors.length === 0" class="no-results">
                    {{ profSearch ? 'No se encontraron resultados' : 'No hay más profesores disponibles' }}
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          <!-- Bloque Ayudantes -->
          <div class="form-group staff-card">
            <div class="staff-header">
              <span class="form-label">
                <span class="staff-icon">🎓</span> Ayudantes Asignados
                <span class="staff-badge">{{ selectedAssistants.length }}</span>
              </span>
            </div>

            <!-- Lista de ayudantes asignados -->
            <div class="staff-chips-container">
              <div 
                v-for="a in selectedAssistants" 
                :key="a.id" 
                class="staff-chip assistant"
              >
                <div class="staff-chip-info">
                  <span class="staff-chip-name">{{ a.full_name }}</span>
                  <span v-if="a.email" class="staff-chip-email">{{ a.email }}</span>
                </div>
                <button 
                  v-if="canManageAssistants"
                  type="button"
                  class="remove-staff-btn"
                  @click="confirmRemoveAssistant(a)"
                  title="Quitar ayudante"
                >
                  &times;
                </button>
              </div>

              <div v-if="selectedAssistants.length === 0" class="no-staff-notice">
                <span>Sin ayudantes asignados a esta coordinación</span>
              </div>
            </div>

            <!-- Dropdown para agregar ayudante -->
            <div 
              v-if="canManageAssistants"
              class="searchable-select" 
              :class="{ 'open': astOpen }"
            >
              <button 
                type="button" 
                class="add-staff-trigger" 
                @click="toggleAst"
              >
                <span>+ Agregar Ayudante</span>
                <span class="chevron">▼</span>
              </button>
              
              <div v-show="astOpen" class="select-dropdown">
                <div class="search-wrapper">
                  <input 
                    type="text" 
                    v-model="astSearch" 
                    placeholder="Buscar ayudante por nombre o correo..." 
                    class="dropdown-search"
                    @click.stop
                    ref="astSearchInput"
                  />
                </div>
                <div class="options-list">
                  <div 
                    v-for="a in availableAssistants" 
                    :key="a.id"
                    class="option-item"
                    @click="addAssistant(a.id)"
                  >
                    <div class="option-info">
                      <span class="option-name">{{ a.full_name }}</span>
                      <span class="option-email">{{ a.email }}</span>
                    </div>
                  </div>
                  <div v-if="availableAssistants.length === 0" class="no-results">
                    {{ astSearch ? 'No se encontraron resultados' : 'No hay más ayudantes disponibles' }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal de confirmación para desasignar docente o ayudante -->
    <ConfirmModal
      :show="showConfirmModal"
      :title="confirmTitle"
      :message="confirmMessage"
      confirm-text="Quitar"
      confirm-class="btn-danger"
      @confirm="executeRemove"
      @cancel="cancelRemove"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import ConfirmModal from '~/components/ConfirmModal.vue'

const props = defineProps({
  courseData: Object,
  sectionData: Object,
  semesterInfo: String,
  professors: {
    type: Array,
    default: () => []
  },
  assistants: {
    type: Array,
    default: () => []
  },
  assignments: {
    type: Object,
    default: () => ({
      professor_id: '',
      assistant_id: '',
      professor_ids: [],
      assistant_ids: []
    })
  },
  readOnly: {
    type: Boolean,
    default: false
  },
  userRole: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update-assignment'])

// Permisos de edición de staff
const canManageProfessors = computed(() => {
  return !props.readOnly && props.userRole !== 'profesor'
})

const canManageAssistants = computed(() => {
  return !props.readOnly
})

// Control de apertura de dropdowns
const profOpen = ref(false)
const astOpen = ref(false)
const profSearchInput = ref(null)
const astSearchInput = ref(null)

// Estados de búsqueda
const profSearch = ref('')
const astSearch = ref('')

// Normalizar IDs asignados
const assignedProfIds = computed(() => {
  const list = props.assignments.professor_ids || []
  if (list.length === 0 && props.assignments.professor_id) {
    return [Number(props.assignments.professor_id)]
  }
  return list.map(Number).filter(id => id > 0)
})

const assignedAstIds = computed(() => {
  const list = props.assignments.assistant_ids || []
  if (list.length === 0 && props.assignments.assistant_id) {
    return [Number(props.assignments.assistant_id)]
  }
  return list.map(Number).filter(id => id > 0)
})

// Objetos de profesores asignados
const selectedProfessors = computed(() => {
  const set = new Set(assignedProfIds.value)
  return props.professors.filter(p => set.has(Number(p.id)))
})

// Objetos de ayudantes asignados
const selectedAssistants = computed(() => {
  const set = new Set(assignedAstIds.value)
  return props.assistants.filter(a => set.has(Number(a.id)))
})

// Profesores disponibles para agregar
const availableProfessors = computed(() => {
  const set = new Set(assignedProfIds.value)
  let list = props.professors.filter(p => !set.has(Number(p.id)))
  if (profSearch.value.trim()) {
    const q = profSearch.value.toLowerCase().trim()
    list = list.filter(p => 
      p.full_name?.toLowerCase().includes(q) || 
      p.email?.toLowerCase().includes(q)
    )
  }
  return list
})

// Ayudantes disponibles para agregar
const availableAssistants = computed(() => {
  const set = new Set(assignedAstIds.value)
  let list = props.assistants.filter(a => !set.has(Number(a.id)))
  if (astSearch.value.trim()) {
    const q = astSearch.value.toLowerCase().trim()
    list = list.filter(a => 
      a.full_name?.toLowerCase().includes(q) || 
      a.email?.toLowerCase().includes(q)
    )
  }
  return list
})

// Toggle handlers
const toggleProf = () => {
  if (!canManageProfessors.value) return
  profOpen.value = !profOpen.value
  astOpen.value = false
  if (profOpen.value) {
    nextTick(() => profSearchInput.value?.focus())
  }
}

const toggleAst = () => {
  if (!canManageAssistants.value) return
  astOpen.value = !astOpen.value
  profOpen.value = false
  if (astOpen.value) {
    nextTick(() => astSearchInput.value?.focus())
  }
}

// Handlers para agregar/quitar
const addProfessor = (id) => {
  const numId = Number(id)
  if (!numId) return

  if (!props.assignments.professor_ids) {
    props.assignments.professor_ids = []
  }
  if (!props.assignments.professor_ids.map(Number).includes(numId)) {
    props.assignments.professor_ids.push(numId)
  }
  props.assignments.professor_id = props.assignments.professor_ids[0]

  profOpen.value = false
  profSearch.value = ''
  emit('update-assignment', { type: 'professor', action: 'add' })
}

const removeProfessor = (id) => {
  if (!canManageProfessors.value) return
  const numId = Number(id)

  props.assignments.professor_ids = (props.assignments.professor_ids || [])
    .map(Number)
    .filter(pid => pid !== numId)

  props.assignments.professor_id = props.assignments.professor_ids.length > 0 
    ? props.assignments.professor_ids[0] 
    : ''

  emit('update-assignment', { type: 'professor', action: 'remove' })
}

const addAssistant = (id) => {
  const numId = Number(id)
  if (!numId) return

  if (!props.assignments.assistant_ids) {
    props.assignments.assistant_ids = []
  }
  if (!props.assignments.assistant_ids.map(Number).includes(numId)) {
    props.assignments.assistant_ids.push(numId)
  }
  props.assignments.assistant_id = props.assignments.assistant_ids[0]

  astOpen.value = false
  astSearch.value = ''
  emit('update-assignment', { type: 'assistant', action: 'add' })
}

const removeAssistant = (id) => {
  if (!canManageAssistants.value) return
  const numId = Number(id)

  props.assignments.assistant_ids = (props.assignments.assistant_ids || [])
    .map(Number)
    .filter(aid => aid !== numId)

  props.assignments.assistant_id = props.assignments.assistant_ids.length > 0 
    ? props.assignments.assistant_ids[0] 
    : ''

  emit('update-assignment', { type: 'assistant', action: 'remove' })
}

// Modal de confirmación para desasignación
const showConfirmModal = ref(false)
const confirmTitle = ref('')
const confirmMessage = ref('')
const pendingRemove = ref(null)

const confirmRemoveProfessor = (prof) => {
  if (!canManageProfessors.value) return
  confirmTitle.value = 'Quitar Profesor'
  confirmMessage.value = `¿Estás seguro de que deseas quitar a "${prof.full_name}" del equipo docente de esta coordinación?`
  pendingRemove.value = () => removeProfessor(prof.id)
  showConfirmModal.value = true
}

const confirmRemoveAssistant = (ast) => {
  if (!canManageAssistants.value) return
  confirmTitle.value = 'Quitar Ayudante'
  confirmMessage.value = `¿Estás seguro de que deseas quitar a "${ast.full_name}" de los ayudantes de esta coordinación?`
  pendingRemove.value = () => removeAssistant(ast.id)
  showConfirmModal.value = true
}

const executeRemove = () => {
  if (pendingRemove.value) {
    pendingRemove.value()
    pendingRemove.value = null
  }
  showConfirmModal.value = false
}

const cancelRemove = () => {
  pendingRemove.value = null
  showConfirmModal.value = false
}

// Cierre al hacer click fuera
const handleClickOutside = (event) => {
  if (!event.target.closest('.searchable-select')) {
    profOpen.value = false
    astOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.general-info {
  padding: 2rem 0;
}

.info-card {
  background: var(--color-surface, #ffffff);
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-md, 10px);
  padding: 2rem;
}

.info-card h3 {
  margin: 0 0 1.5rem 0;
  color: var(--color-dark, #394049);
  font-size: 1.3rem;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.info-item .info-label {
  font-weight: 600;
  color: var(--color-dark, #394049);
  font-size: 0.875rem;
}

.info-item span {
  color: var(--color-text-secondary, #64748b);
  font-size: 1rem;
}

.assignment-section {
  border-top: 1px solid var(--color-border, #dfe4ea);
  padding-top: 2rem;
}

.assignment-section h4 {
  margin: 0 0 1.5rem 0;
  color: var(--color-dark, #394049);
  font-size: 1.1rem;
}

.assignment-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 1.5rem;
}

.staff-card {
  background: var(--color-background, #f8fafc);
  border: 1px solid var(--color-border, #e2e8f0);
  border-radius: var(--radius-md, 10px);
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.staff-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.staff-header .form-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 600;
  color: var(--color-dark, #394049);
  font-size: 0.95rem;
}

.staff-icon {
  font-size: 1.1rem;
}

.staff-badge {
  background: rgba(0, 164, 153, 0.12);
  color: var(--color-primary, #00A499);
  font-size: 0.75rem;
  padding: 0.15rem 0.5rem;
  border-radius: 9999px;
  font-weight: 700;
}

.staff-chips-container {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  min-height: 50px;
}

.staff-chip {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  border-left: 4px solid var(--color-primary, #00A499);
  border-radius: 6px;
  padding: 0.6rem 0.75rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.75rem;
  box-shadow: 0 1px 2px rgba(0,0,0,0.03);
}

.staff-chip.assistant {
  border-left-color: #6366f1;
}

.staff-chip-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.staff-chip-name {
  font-weight: 600;
  color: var(--color-dark, #1e293b);
  font-size: 0.9rem;
}

.staff-chip-email {
  font-size: 0.75rem;
  color: #64748b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.remove-staff-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  font-size: 1.3rem;
  line-height: 1;
  cursor: pointer;
  padding: 0.2rem 0.4rem;
  border-radius: 4px;
  transition: all 0.2s;
}

.remove-staff-btn:hover {
  background: #fee2e2;
  color: #dc2626;
}

.no-staff-notice {
  padding: 1rem;
  text-align: center;
  color: #94a3b8;
  font-size: 0.85rem;
  font-style: italic;
  background: #ffffff;
  border: 1px dashed #cbd5e1;
  border-radius: 6px;
}

.add-staff-trigger {
  width: 100%;
  padding: 0.6rem 1rem;
  border: 1px dashed var(--color-primary, #00A499);
  border-radius: 6px;
  background: #ffffff;
  color: var(--color-primary, #00A499);
  font-weight: 600;
  font-size: 0.875rem;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: all 0.2s;
}

.add-staff-trigger:hover {
  background: rgba(0, 164, 153, 0.05);
  border-style: solid;
}

.searchable-select {
  position: relative;
  width: 100%;
}

.chevron {
  font-size: 0.7rem;
  color: inherit;
  transition: transform 0.2s;
}

.searchable-select.open .chevron {
  transform: rotate(180deg);
}

.select-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  z-index: 100;
  margin-top: 0.4rem;
  background: var(--color-surface, #ffffff);
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-md, 10px);
  box-shadow: var(--shadow-lg, 0 10px 15px -3px rgba(0,0,0,0.1));
  overflow: hidden;
}

.search-wrapper {
  padding: 0.75rem;
  border-bottom: 1px solid var(--color-border, #dfe4ea);
  background: var(--color-background, #f4f6f9);
}

.dropdown-search {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-sm, 6px);
  font-size: 0.875rem;
  outline: none;
}

.dropdown-search:focus {
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 0 0 2px rgba(0, 164, 153, 0.1);
}

.options-list {
  max-height: 220px;
  overflow-y: auto;
}

.option-item {
  padding: 0.6rem 0.9rem;
  cursor: pointer;
  transition: background 0.2s;
  border-bottom: 1px solid #f1f5f9;
}

.option-item:hover {
  background: rgba(0, 164, 153, 0.08);
}

.option-info {
  display: flex;
  flex-direction: column;
}

.option-name {
  font-weight: 500;
  color: #1e293b;
  font-size: 0.875rem;
}

.option-email {
  font-size: 0.75rem;
  color: #64748b;
}

.no-results {
  padding: 1.5rem 1rem;
  text-align: center;
  color: #94a3b8;
  font-size: 0.85rem;
  font-style: italic;
}

@media (max-width: 640px) {
  .assignment-grid {
    grid-template-columns: 1fr;
  }
}
</style>
