<template>
  <div class="courses-page">
    <!-- Header con información del usuario -->
    <TopBar :user="user" @logout="handleLogout" />

    <!-- Banner institucional -->
    <div class="page-banner">
      <div class="container">
        <div class="banner-inner">
          <div class="banner-left">
            <nav class="breadcrumb" aria-label="Navegación">
              <NuxtLink to="/home" class="breadcrumb-link">Inicio</NuxtLink>
              <span class="breadcrumb-sep">›</span>
              <span class="breadcrumb-item active">Cursos</span>
            </nav>
            <h1 class="banner-title">Gestión Curricular de Cursos</h1>
            <p class="banner-subtitle">Catálogo de cursos, coordinaciones y períodos académicos</p>
          </div>
        </div>
      </div>
    </div>

    <main class="main-content">
      <div class="container">
        <div class="courses-container">
          <div class="header-section">
            <div class="header-section-left">
              <h2 class="courses-list-heading">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/>
                  <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>
                </svg>
                Cursos Registrados
              </h2>
            </div>
            <button 
              v-if="user?.role_name === 'admin'" 
              @click="showCreateModal = true"
              class="create-btn"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <path d="M12 8v8"/>
                <path d="M8 12h8"/>
              </svg>
              Crear Curso
            </button>
          </div>
          
          <!-- Lista de cursos existentes -->
          <div v-if="user?.role_name === 'admin' || user?.role_name === 'coordinador' || user?.role_name === 'profesor' || user?.role_name === 'ayudante'" class="courses-list">
            <h3>Cursos Existentes</h3>
            
            <!-- Buscador de cursos -->
            <div class="search-box">
              <svg class="search-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="11" cy="11" r="8"/>
                <path d="m21 21-4.35-4.35"/>
              </svg>
              <input 
                v-model="searchQuery"
                type="text" 
                placeholder="Buscar por nombre o código del curso..."
                class="search-input"
              />
              <button 
                v-if="searchQuery" 
                @click="searchQuery = ''"
                class="clear-search-btn"
                title="Limpiar búsqueda"
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"/>
                  <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
            </div>
            
            <div v-if="loading" class="loading">
              Cargando cursos...
            </div>
            <div v-else-if="filteredCourses.length === 0 && searchQuery" class="no-courses">
              No se encontraron cursos que coincidan con "{{ searchQuery }}"
            </div>
            <div v-else-if="courses.length === 0" class="no-courses">
              No hay cursos disponibles
            </div>
            <div v-else class="courses-grid">
              <div v-for="course in filteredCourses" :key="course.id" class="course-card">
                <!-- Card Header -->
                <div class="course-card-header">
                  <div class="course-card-header-left">
                    <span class="course-code-badge">{{ course.code }}</span>
                    <span v-if="course.sections && course.sections.length > 0" class="course-sections-count">
                      {{ course.sections.length }} {{ course.sections.length === 1 ? 'coordinación' : 'coordinaciones' }}
                    </span>
                  </div>
                  <div v-if="user?.role_name === 'admin'" class="course-card-header-actions">
                    <button @click="editCourse(course)" class="card-icon-btn edit-icon-btn" title="Editar curso">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                        <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                      </svg>
                    </button>
                    <button @click="deleteCourse(course)" class="card-icon-btn delete-icon-btn" title="Eliminar curso">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="3,6 5,6 21,6"/>
                        <path d="M19,6v14a2,2,0,0,1-2,2H7a2,2,0,0,1-2-2V6m3,0V4a2,2,0,0,1,2-2h4a2,2,0,0,1,2,2v2"/>
                      </svg>
                    </button>
                  </div>
                </div>

                <!-- Course Name -->
                <p class="course-name">{{ course.name }}</p>

                <!-- Coordinaciones del curso -->
                <div v-if="course.sections && course.sections.length > 0" class="course-sections">
                  <div class="sections-header">
                    <h5>
                      <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="flex-shrink:0">
                        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                        <circle cx="9" cy="7" r="4"/>
                        <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
                        <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                      </svg>
                      Coordinaciones
                    </h5>
                    <div class="semester-filter">
                      <div class="semester-dropdown" :class="{ 'active': showSemesterDropdown[course.id] }">
                        <button 
                          @click="toggleSemesterDropdown(course.id)"
                          class="semester-dropdown-btn"
                          type="button"
                        >
                          {{ selectedSemesterFilter || 'Todos los semestres' }}
                          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <polyline points="6,9 12,15 18,9"/>
                          </svg>
                        </button>
                        
                        <div v-if="showSemesterDropdown[course.id]" class="semester-dropdown-menu">
                          <div class="semester-options">
                            <button
                              @click="selectSemester(course.id, { term: 'Todos', year: '' })"
                              class="semester-option"
                              :class="{ 'selected': !selectedSemesterFilter }"
                            >
                              Todos los semestres
                            </button>
                            <button
                              v-for="semester in getCourseSemesters(courses.find(c => c.id === course.id))"
                              :key="`${semester.term}_${semester.year}`"
                              @click="selectSemester(course.id, semester)"
                              class="semester-option"
                              :class="{ 'selected': selectedSemesterFilter === `${semester.term}_${semester.year}` }"
                            >
                              {{ semester.term }}_{{ semester.year }}
                              <span v-if="currentSemester && semester.id === currentSemester.id" class="current-badge">Actual</span>
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                  
                  <div class="sections-list">
                    <div v-for="section in getFilteredSections(course)" :key="section.id" class="section-item">
                      <span 
                        @click="viewSection(section)" 
                        class="section-name clickable"
                        :title="'Ver información de ' + section.name"
                      >
                        <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="section-arrow-icon">
                          <polyline points="9,18 15,12 9,6"/>
                        </svg>
                        {{ section.name }}
                      </span>
                      <div class="section-actions" v-if="user?.role_name === 'admin'">
                        <button @click="editSection(section)" class="section-edit-btn" title="Editar coordinación">
                          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                          </svg>
                        </button>
                        <button @click="deleteSection(section)" class="section-delete-btn" title="Eliminar coordinación">
                          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <polyline points="3,6 5,6 21,6"/>
                            <path d="M19,6v14a2,2,0,0,1-2,2H7a2,2,0,0,1-2-2V6m3,0V4a2,2,0,0,1,2-2h4a2,2,0,0,1,2,2v2"/>
                          </svg>
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
                <div v-else class="no-sections">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="no-sections-icon">
                    <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                    <circle cx="9" cy="7" r="4"/>
                    <line x1="19" y1="8" x2="19" y2="14"/>
                    <line x1="22" y1="11" x2="16" y2="11"/>
                  </svg>
                  <p class="no-sections-text">Sin coordinaciones aún</p>
                </div>

                <!-- Add section footer -->
                <div v-if="user?.role_name === 'admin'" class="course-card-footer">
                  <button @click="addSection(course)" class="add-section-btn">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                      <line x1="12" y1="5" x2="12" y2="19"/>
                      <line x1="5" y1="12" x2="19" y2="12"/>
                    </svg>
                    Añadir coordinación
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Información por rol (para no administradores) -->
          <div v-else class="role-info">
            
            <div v-if="user?.role_name === 'estudiante'" class="info-card">

              <div v-if="studentEnrollLoading" class="loading">Cargando cursos inscritos...</div>

              <div v-else-if="enrolledCourses.length === 0" class="student-no-courses">
                <svg class="student-no-courses-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M12 14l9-5-9-5-9 5 9 5z"/>
                  <path d="M12 14l6.16-3.422a12.083 12.083 0 01.665 6.479A11.952 11.952 0 0012 20.055a11.952 11.952 0 00-6.824-2.998 12.078 12.078 0 01.665-6.479L12 14z"/>
                  <path d="M12 14l9-5-9-5-9 5 9 5zm0 0l6.16-3.422a12.083 12.083 0 01.665 6.479A11.952 11.952 0 0012 20.055a11.952 11.952 0 00-6.824-2.998 12.078 12.078 0 01.665-6.479L12 14z"/>
                </svg>
                <h3>No tienes cursos inscritos</h3>
                <p>Contacta a tu profesor o administrador para inscribirte en cursos.</p>
              </div>

              <div v-else class="courses-grid student-courses-grid">
                <div v-for="ec in enrolledCourses" :key="ec.section.id" class="course-card" @click="viewStudentSection(ec.section)">
                  <div class="student-course-header">
                    <h3 class="student-course-title">{{ ec.course?.name || 'Curso desconocido' }}</h3>
                    <span class="student-course-code">{{ ec.course?.code || '—' }}</span>
                  </div>

                  <div class="student-course-info">
                    <div class="student-course-detail">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
                        <line x1="16" y1="2" x2="16" y2="6"/>
                        <line x1="8" y1="2" x2="8" y2="6"/>
                        <line x1="3" y1="10" x2="21" y2="10"/>
                      </svg>
                      <span>Coordinación : {{ ec.section.name }}</span>
                    </div>

                    <div class="student-course-detail">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
                        <line x1="16" y1="2" x2="16" y2="6"/>
                        <line x1="8" y1="2" x2="8" y2="6"/>
                        <line x1="3" y1="10" x2="21" y2="10"/>
                        <path d="M8 14l2 2 4-4"/>
                      </svg>
                      <span>Semestre: {{ ec.semester ? `${ec.semester.term}_${ec.semester.year}` : 'N/A' }}</span>
                    </div>

                    <div class="student-course-detail">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                        <circle cx="12" cy="7" r="4"/>
                      </svg>
                      <span>Profesor: {{ ec.section.teacher_name || 'Por asignar' }}</span>
                    </div>
                  </div>

                  <button class="student-course-action" @click.stop="viewStudentSection(ec.section)">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0z"/>
                      <path d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
                    </svg>
                    Ver Coordinación 
                  </button>
                </div>
              </div>
            </div>
            
            <div v-else-if="user?.role_name === 'profesor'" class="info-card">
              <h3>Como Profesor</h3>
              <ul>
                <li>Ver cursos asignados</li>
                <li>Gestionar estudiantes</li>
                <li>Ingresar calificaciones</li>
                <li>Ver listas de curso</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Modal para crear/editar curso -->
    <div v-if="showCreateModal || showEditModal" class="modal-overlay" @click="closeModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3>{{ editingCourse ? 'Editar Curso' : 'Crear Nuevo Curso' }}</h3>
          <button @click="closeModal" class="close-btn">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        
        <form @submit.prevent="submitCourse" class="modal-body">
          <div class="form-group">
            <label for="courseCode">Código del Curso *</label>
            <input
              id="courseCode"
              v-model="courseForm.code"
              type="text"
              placeholder="Ej: INF-101"
              required
              :disabled="editingCourse"
            />
            <small>El código debe ser único y seguir el formato departamento-número</small>
          </div>
          
          <div class="form-group">
            <label for="courseName">Nombre del Curso *</label>
            <input
              id="courseName"
              v-model="courseForm.name"
              type="text"
              placeholder="Ej: Introducción a la Programación"
              required
            />
          </div>

          <!-- Asignación opcional de profesor/ayudante (solo al crear) -->
          <template v-if="!editingCourse">
            <div v-if="staffUsers.length > 0" class="staff-assign-hint">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <line x1="12" y1="8" x2="12" y2="12"/>
                <line x1="12" y1="16" x2="12.01" y2="16"/>
              </svg>
              Al asignar un profesor o ayudante se creará automáticamente una «Sección A» en el semestre activo.
            </div>
            <div class="form-row">
              <div class="form-group">
                <label for="courseProfessor">Profesor <span class="optional-badge">opcional</span></label>
                <select id="courseProfessor" v-model="courseForm.professor_id" :disabled="staffLoading">
                  <option :value="null">{{ staffLoading ? 'Cargando...' : 'Sin asignar' }}</option>
                  <option v-for="u in staffUsers" :key="u.id" :value="u.id">{{ u.full_name }}</option>
                </select>
              </div>
              <div class="form-group">
                <label for="courseAssistant">Ayudante <span class="optional-badge">opcional</span></label>
                <select id="courseAssistant" v-model="courseForm.assistant_id" :disabled="staffLoading">
                  <option :value="null">{{ staffLoading ? 'Cargando...' : 'Sin asignar' }}</option>
                  <option v-for="u in staffUsers" :key="u.id" :value="u.id">{{ u.full_name }}</option>
                </select>
              </div>
            </div>
          </template>
          
          <div class="modal-actions">
            <button type="button" @click="closeModal" class="cancel-btn">
              Cancelar
            </button>
            <button type="submit" class="submit-btn" :disabled="submitting">
              {{ submitting ? 'Guardando...' : (editingCourse ? 'Actualizar' : 'Crear') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Modal de confirmación para eliminar -->
    <div v-if="showDeleteModal" class="modal-overlay" @click="showDeleteModal = false">
      <div class="modal confirm-modal" @click.stop>
        <div class="modal-header">
          <h3>Confirmar Eliminación</h3>
        </div>
        <div class="modal-body">
          <p>¿Estás seguro de que deseas eliminar el curso <strong>{{ courseToDelete?.code }} - {{ courseToDelete?.name }}</strong>?</p>
          <p class="warning">Esta acción no se puede deshacer.</p>
        </div>
        <div class="modal-actions">
          <button @click="showDeleteModal = false" class="cancel-btn">
            Cancelar
          </button>
          <button @click="confirmDelete" class="modal-delete-btn" :disabled="deleting">
            {{ deleting ? 'Eliminando...' : 'Eliminar' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal para añadir coordinación  -->
    <div v-if="showSectionModal" class="modal-overlay" @click="closeSectionModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3>Añadir Coordinación  - {{ selectedCourse?.code }}</h3>
          <button @click="closeSectionModal" class="close-btn">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        
        <form @submit.prevent="submitSection" class="modal-body">
          <div class="form-group">
            <label for="sectionName">Nombre de la Coordinación  *</label>
            <input
              id="sectionName"
              v-model="sectionForm.name"
              type="text"
              placeholder="Ej: Coordinación  A, Coordinación  B, Grupo 1"
              required
            />
            <small>Nombre identificativo de la coordinación  (ej: Coordinación  A, Grupo 1, etc.)</small>
          </div>
          
          <div class="info-section">
            <h4>Información del Curso</h4>
            <p><strong>Código:</strong> {{ selectedCourse?.code }}</p>
            <p><strong>Nombre:</strong> {{ selectedCourse?.name }}</p>
            <p><strong>Semestre:</strong> {{ currentSemester ? `${currentSemester.year}-${currentSemester.term}` : 'Cargando...' }}</p>
          </div>

          <!-- Asignación opcional de profesor/ayudante -->
          <div v-if="staffUsers.length > 0" class="form-row">
            <div class="form-group">
              <label for="sectionProfessor">Profesor <span class="optional-badge">opcional</span></label>
              <select id="sectionProfessor" v-model="sectionForm.professor_id" :disabled="staffLoading">
                <option :value="null">{{ staffLoading ? 'Cargando...' : 'Sin asignar' }}</option>
                <option v-for="u in staffUsers" :key="u.id" :value="u.id">{{ u.full_name }}</option>
              </select>
            </div>
            <div class="form-group">
              <label for="sectionAssistant">Ayudante <span class="optional-badge">opcional</span></label>
              <select id="sectionAssistant" v-model="sectionForm.assistant_id" :disabled="staffLoading">
                <option :value="null">{{ staffLoading ? 'Cargando...' : 'Sin asignar' }}</option>
                <option v-for="u in staffUsers" :key="u.id" :value="u.id">{{ u.full_name }}</option>
              </select>
            </div>
          </div>
          
          <div class="modal-actions">
            <button type="button" @click="closeSectionModal" class="cancel-btn">
              Cancelar
            </button>
            <button type="submit" class="submit-btn" :disabled="submittingSection">
              {{ submittingSection ? 'Creando...' : 'Crear Coordinación ' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Modal para editar coordinación -->
    <div v-if="showEditSectionModal" class="modal-overlay" @click="closeEditSectionModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3>Editar Coordinación</h3>
          <button @click="closeEditSectionModal" class="close-btn">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        
        <form @submit.prevent="submitEditSection" class="modal-body">
          <div class="form-group">
            <label for="editSectionNameInput">Nombre de la Coordinación *</label>
            <input
              id="editSectionNameInput"
              v-model="editSectionForm.name"
              type="text"
              placeholder="Ej: Coordinación A"
              required
            />
          </div>
          
          <div class="modal-actions">
            <button type="button" @click="closeEditSectionModal" class="cancel-btn">
              Cancelar
            </button>
            <button type="submit" class="submit-btn" :disabled="submittingEditSection">
              {{ submittingEditSection ? 'Guardando...' : 'Guardar Cambios' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Confirm Modal genérico para acciones de coordinación -->
    <ConfirmModal
      :show="showConfirmModal"
      :title="confirmTitle"
      :message="confirmMessage"
      :confirm-text="confirmBtnText"
      confirm-class="btn-danger"
      @confirm="handleConfirmAction"
      @cancel="showConfirmModal = false"
    />

    <!-- Info Modal para mensajes y alertas -->
    <InfoModal
      :show="showInfoModal"
      :title="infoTitle"
      :message="infoMessage"
      @close="showInfoModal = false"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, reactive, computed } from 'vue'
import ConfirmModal from '~/components/ConfirmModal.vue'
import InfoModal from '~/components/InfoModal.vue'

definePageMeta({
  middleware: 'auth'
})

const { getUser, getToken, logout } = useAuth()
const user = ref(null)
const courses = ref([])
const loading = ref(false)
const enrolledCourses = ref([])
const studentEnrollLoading = ref(false)

// Variable para la búsqueda
const searchQuery = ref('')

// Estados del modal
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteModal = ref(false)
const showSectionModal = ref(false)
const submitting = ref(false)
const deleting = ref(false)
const submittingSection = ref(false)

// Estados para ConfirmModal e InfoModal
const showConfirmModal = ref(false)
const confirmTitle = ref('Confirmar Acción')
const confirmMessage = ref('')
const confirmBtnText = ref('Eliminar')
const pendingConfirmAction = ref(null)

const handleConfirmAction = async () => {
  if (pendingConfirmAction.value) {
    const action = pendingConfirmAction.value
    pendingConfirmAction.value = null
    await action()
  }
  showConfirmModal.value = false
}

const showInfoModal = ref(false)
const infoTitle = ref('')
const infoMessage = ref('')

const showInfo = (title, message) => {
  infoTitle.value = title
  infoMessage.value = message
  showInfoModal.value = true
}

// Estados para editar coordinación
const showEditSectionModal = ref(false)
const sectionToEdit = ref(null)
const editSectionForm = reactive({
  name: ''
})
const submittingEditSection = ref(false)

// Datos del formulario
const courseForm = reactive({
  code: '',
  name: '',
  professor_id: null,
  assistant_id: null
})

// Datos del formulario de coordinación 
const sectionForm = reactive({
  name: '',
  professor_id: null,
  assistant_id: null
})

// Estados de edición y eliminación
const editingCourse = ref(null)
const courseToDelete = ref(null)
const selectedCourse = ref(null)
const currentSemester = ref(null)

// Variables para el filtro de semestres
const allSemesters = ref([])
const selectedSemesterFilter = ref(null)
const showSemesterDropdown = ref({})

// URL base de la API
const API_BASE_URL = 'http://127.0.0.1:3000/api/v1'

// Staff (profesores) disponibles para asignar a secciones
const staffUsers = ref([])
const staffLoading = ref(false)

// Computed property para filtrar cursos por búsqueda
// Computed property para filtrar cursos por búsqueda y rol
const filteredCourses = computed(() => {
  let result = courses.value

  // Si es profesor o ayudante, filtrar solo los cursos donde tiene al menos una sección
  if (user.value?.role_name === 'profesor') {
    result = result.filter(course => {
      return course.sections && course.sections.some(section => {
        const pIds = Array.isArray(section.professor_ids) ? section.professor_ids.map(Number) : []
        return Number(section.professor_id) === Number(user.value.id) || pIds.includes(Number(user.value.id))
      })
    })
  } else if (user.value?.role_name === 'ayudante') {
    result = result.filter(course => {
      return course.sections && course.sections.some(section => {
        const aIds = Array.isArray(section.assistant_ids) ? section.assistant_ids.map(Number) : []
        return Number(section.assistant_id) === Number(user.value.id) || aIds.includes(Number(user.value.id))
      })
    })
  }
  
  if (!searchQuery.value.trim()) {
    return result
  }
  
  const query = searchQuery.value.toLowerCase().trim()
  
  return result.filter(course => {
    const code = course.code?.toLowerCase() || ''
    const name = course.name?.toLowerCase() || ''
    
    return code.includes(query) || name.includes(query)
  })
})

onMounted(() => {
  user.value = getUser()
  
  if (!user.value) {
    navigateTo('/')
    return
  }

  // Cargar cursos si es administrador, coordinador, profesor o ayudante
  if (['admin', 'coordinador', 'profesor', 'ayudante'].includes(user.value.role_name)) {
    loadCoursesComplete()
  }

  // Cargar lista de profesores para selects de asignación
  if (user.value.role_name === 'admin') {
    loadStaffUsers()
  }

  // Si es estudiante, cargar sus cursos inscritos
  if (user.value.role_name === 'estudiante') {
    loadStudentEnrollments()
  }
  
  // Cerrar dropdowns al hacer clic fuera
  document.addEventListener('click', (event) => {
    if (!event.target.closest('.semester-dropdown')) {
      Object.keys(showSemesterDropdown.value).forEach(courseId => {
        showSemesterDropdown.value[courseId] = false
      })
    }
  })
})

// Cargar toda la información de cursos en una sola llamada
const loadCoursesComplete = async () => {
  try {
    loading.value = true
    const token = getToken()
    
    const response = await fetch(`${API_BASE_URL}/courses-complete`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })

    if (!response.ok) {
      throw new Error('Error al cargar los cursos')
    }

    const data = await response.json()
    
    // Asignar todos los datos de una vez
    courses.value = data.courses || []
    currentSemester.value = data.current_semester || null
    allSemesters.value = data.semesters || []
    
    // Establecer el semestre actual como filtro por defecto
    if (currentSemester.value) {
      selectedSemesterFilter.value = `${currentSemester.value.term}_${currentSemester.value.year}`
    }
    
  } catch (error) {
    console.error('Error loading courses complete:', error)
    showInfo('Error', 'Error al cargar los cursos')
  } finally {
    loading.value = false
  }
}

// Cargar lista de cursos (método legacy, mantener por compatibilidad)
const loadCourses = async () => {
  try {
    loading.value = true
    const token = getToken()
    
    const response = await fetch(`${API_BASE_URL}/courses`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })

    if (!response.ok) {
      throw new Error('Error al cargar los cursos')
    }

    const data = await response.json()
    courses.value = data || []
    
    // Cargar secciones para cada curso
    if (courses.value.length > 0) {
      await loadSectionsForCourses()
    }
  } catch (error) {
    console.error('Error loading courses:', error)
    showInfo('Error', 'Error al cargar los cursos')
  } finally {
    loading.value = false
  }
}

// Cargar cursos en los que el estudiante está inscrito
const loadStudentEnrollments = async () => {
  try {
    studentEnrollLoading.value = true
    const token = getToken()

    // user.id debería venir del token / user_info
    const studentId = user.value?.id
    if (!studentId) {
      enrolledCourses.value = []
      return
    }

    const resp = await fetch(`${API_BASE_URL}/students/${studentId}/sections`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })

    if (!resp.ok) {
      console.error('Failed to load student sections', resp.status)
      enrolledCourses.value = []
      return
    }

    const sections = await resp.json() // array of { student_id, section_id }

    // For each section, fetch section details (which include course_offering), then get courseOffering to get course and semester
    const tokenHeader = { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' }

    const results = []
    for (const ss of sections) {
      try {
        const sectionResp = await fetch(`${API_BASE_URL}/sections/${ss.section_id}`, { headers: tokenHeader })
        if (!sectionResp.ok) continue
        const section = await sectionResp.json()

        // section has course_offering_id -> fetch course-offering
        const coResp = await fetch(`${API_BASE_URL}/course-offerings/${section.course_offering_id}`, { headers: tokenHeader })
        let course = null
        let semester = null
        if (coResp.ok) {
          const co = await coResp.json()
          course = co.course || null
          semester = co.semester || null
        }

        results.push({ section, course, semester })
      } catch (e) {
        console.error('Error loading enrollment detail', e)
      }
    }

    enrolledCourses.value = results
  } catch (error) {
    console.error('Error loading student enrollments:', error)
    enrolledCourses.value = []
  } finally {
    studentEnrollLoading.value = false
  }
}

// Helper: build course object with its sections enriched with semester info
const buildCourseWithSections = (course, courseOfferings, allSections, allSemesters) => {
  const courseCourseOfferings = (courseOfferings || []).filter(co => co.course_id === course.id)
  const courseSections = []
  const semesterMap = new Map((allSemesters || []).map(s => [s.id, s]))

  courseCourseOfferings.forEach(co => {
    const coSections = (allSections || []).filter(section => section.course_offering_id === co.id)
    const semester = semesterMap.get(co.semester_id)
    const semesterLabel = semester ? `${semester.term}_${semester.year}` : 'N/A'
    coSections.forEach(section => {
      courseSections.push({ ...section, semester: semesterLabel })
    })
  })

  return { ...course, sections: courseSections }
}

// Cargar secciones para todos los cursos
const loadSectionsForCourses = async () => {
  try {
    const token = getToken()
    
    // Cargar todas las course offerings
    const courseOfferingsResponse = await fetch(`${API_BASE_URL}/course-offerings`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })
    
    const courseOfferings = courseOfferingsResponse.ok ? await courseOfferingsResponse.json() : []
    
    // Cargar todas las secciones
    const sectionsResponse = await fetch(`${API_BASE_URL}/sections`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })
    
    const allSections = sectionsResponse.ok ? await sectionsResponse.json() : []
    
    // Cargar todos los semestres para obtener la información completa
    const semestersResponse = await fetch(`${API_BASE_URL}/semesters`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })
    
    const allSemesters = semestersResponse.ok ? await semestersResponse.json() : []
    
    // Asignar secciones a cada curso
    courses.value = courses.value.map(course =>
      buildCourseWithSections(course, courseOfferings, allSections, allSemesters)
    )
  } catch (error) {
    console.error('Error loading sections for courses:', error)
    // No es crítico, simplemente no mostraremos secciones
  }
}

// Crear nuevo curso
const loadStaffUsers = async () => {
  try {
    staffLoading.value = true
    const token = getToken()
    const response = await fetch(`${API_BASE_URL}/users?role_id=2&limit=1000`, {
      headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' }
    })
    if (response.ok) {
      const result = await response.json()
      const raw = result.data || result.users || []
      staffUsers.value = raw.map(u => ({
        id: u.ID || u.id,
        full_name: (u.FullName || u.full_name || `${u.Nombres || ''} ${u.Paterno || ''}`).trim()
      }))
    }
  } catch (err) {
    console.error('Error loading staff users:', err)
  } finally {
    staffLoading.value = false
  }
}

// Crear nuevo curso
const createCourse = async () => {
  try {
    const token = getToken()
    
    const response = await fetch(`${API_BASE_URL}/courses`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        code: courseForm.code.trim(),
        name: courseForm.name.trim()
      })
    })

    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.error || 'Error al crear el curso')
    }

    const newCourse = await response.json()
    courses.value.push(newCourse)

    // Si se asignó profesor o ayudante, crear automáticamente una sección inicial
    if ((courseForm.professor_id || courseForm.assistant_id) && currentSemester.value) {
      try {
        const courseOfferingId = await getOrCreateCourseOffering(newCourse.id)
        const sectionToken = getToken()
        await fetch(`${API_BASE_URL}/sections`, {
          method: 'POST',
          headers: { 'Authorization': `Bearer ${sectionToken}`, 'Content-Type': 'application/json' },
          body: JSON.stringify({
            name: 'Sección A',
            course_offering_id: courseOfferingId,
            professor_id: courseForm.professor_id || null,
            assistant_id: courseForm.assistant_id || null
          })
        })
      } catch (e) {
        console.warn('No se pudo crear la sección inicial con staff:', e)
      }
    }
    
    return true
  } catch (error) {
    console.error('Error creating course:', error)
    showInfo('Error', error.message || 'Error al crear el curso')
    return false
  }
}

// Actualizar curso existente
const updateCourse = async () => {
  try {
    const token = getToken()
    
    const response = await fetch(`${API_BASE_URL}/courses/${editingCourse.value.id}`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        code: courseForm.code.trim(),
        name: courseForm.name.trim()
      })
    })

    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.error || 'Error al actualizar el curso')
    }

    const updatedCourse = await response.json()
    const index = courses.value.findIndex(c => c.id === editingCourse.value.id)
    if (index !== -1) {
      courses.value[index] = updatedCourse
    }
    
    return true
  } catch (error) {
    console.error('Error updating course:', error)
    showInfo('Error', error.message || 'Error al actualizar el curso')
    return false
  }
}

// Eliminar curso
const deleteCourseApi = async (courseId) => {
  try {
    const token = getToken()
    
    const response = await fetch(`${API_BASE_URL}/courses/${courseId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })

    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.error || 'Error al eliminar el curso')
    }

    courses.value = courses.value.filter(c => c.id !== courseId)
    return true
  } catch (error) {
    console.error('Error deleting course:', error)
    showInfo('Error', error.message || 'Error al eliminar el curso')
    return false
  }
}

// Manejar envío del formulario
const submitCourse = async () => {
  if (!courseForm.code.trim() || !courseForm.name.trim()) {
    showInfo('Validación', 'Por favor, completa todos los campos requeridos')
    return
  }

  submitting.value = true
  
  try {
    let success = false
    
    if (editingCourse.value) {
      success = await updateCourse()
    } else {
      success = await createCourse()
    }
    
    if (success) {
      closeModal()
      resetForm()
    }
  } finally {
    submitting.value = false
  }
}

// Abrir modal de edición
const editCourse = (course) => {
  editingCourse.value = course
  courseForm.code = course.code
  courseForm.name = course.name
  showEditModal.value = true
}

// Abrir modal de eliminación
const deleteCourse = (course) => {
  courseToDelete.value = course
  showDeleteModal.value = true
}

// Confirmar eliminación
const confirmDelete = async () => {
  if (!courseToDelete.value) return
  
  deleting.value = true
  
  try {
    const success = await deleteCourseApi(courseToDelete.value.id)
    if (success) {
      showDeleteModal.value = false
      courseToDelete.value = null
    }
  } finally {
    deleting.value = false
  }
}

// Cargar semestre actual
const loadCurrentSemester = async () => {
  try {
    const token = getToken()
    
    const response = await fetch(`${API_BASE_URL}/semesters/current`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })

    if (!response.ok) {
      if (response.status === 404) {
        console.log('No hay semestre activo configurado')
        currentSemester.value = null
        return
      }
      throw new Error('Error al cargar el semestre actual')
    }

    const responseData = await response.json()
    // El endpoint devuelve { success: true, data: semester }
    if (responseData.success && responseData.data) {
      currentSemester.value = responseData.data
      // Establecer el semestre actual como filtro por defecto
      selectedSemesterFilter.value = `${responseData.data.term}_${responseData.data.year}`
    } else {
      currentSemester.value = null
    }
  } catch (error) {
    console.error('Error loading current semester:', error)
    currentSemester.value = null
    // No mostrar alerta aquí ya que podría no haber semestre activo configurado
  }
}

// Cargar todos los semestres disponibles
const loadAllSemesters = async () => {
  try {
    const token = getToken()
    
    const response = await fetch(`${API_BASE_URL}/semesters`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })

    if (!response.ok) {
      throw new Error('Error al cargar los semestres')
    }

    const data = await response.json()
    allSemesters.value = data || []
  } catch (error) {
    console.error('Error loading all semesters:', error)
    allSemesters.value = []
  }
}

// Obtener o crear course offering para el semestre actual
const getOrCreateCourseOffering = async (courseId) => {
  try {
    const token = getToken()
    
    if (!currentSemester.value) {
      throw new Error('No hay semestre activo configurado')
    }

    // Primero intentar obtener las course offerings existentes
    const response = await fetch(`${API_BASE_URL}/course-offerings`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })

    let courseOfferings = []
    
    if (response.ok) {
      const data = await response.json()
      courseOfferings = data || []
      
      // Buscar si ya existe una course offering para este curso y semestre
      const existing = courseOfferings.find(co => 
        co.course_id === courseId && co.semester_id === currentSemester.value.id
      )
      
      if (existing) {
        return existing.id
      }
    } else {
      console.log('No se pudieron cargar las course offerings existentes, creando nueva...')
    }

    // Si no existe o no se pudieron cargar, crear una nueva course offering
    const createResponse = await fetch(`${API_BASE_URL}/course-offerings`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        course_id: courseId,
        semester_id: currentSemester.value.id
      })
    })

    if (!createResponse.ok) {
      const errorData = await createResponse.json()
      throw new Error(errorData.error || 'Error al crear la oferta de curso')
    }

    const newCourseOffering = await createResponse.json()
    return newCourseOffering.id
  } catch (error) {
    console.error('Error getting or creating course offering:', error)
    throw error
  }
}

// Crear nueva coordinación 
const createSection = async () => {
  try {
    const token = getToken()
    
    // Obtener o crear la course offering
    const courseOfferingId = await getOrCreateCourseOffering(selectedCourse.value.id)
    
    const response = await fetch(`${API_BASE_URL}/sections`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        name: sectionForm.name.trim(),
        course_offering_id: courseOfferingId,
        professor_id: sectionForm.professor_id || null,
        assistant_id: sectionForm.assistant_id || null
      })
    })

    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.error || 'Error al crear la coordinación ')
    }

    await response.json()
    return true
  } catch (error) {
    console.error('Error creating section:', error)
    showInfo('Error', error.message || 'Error al crear la coordinación')
    return false
  }
}

// Abrir modal de añadir coordinación 
const addSection = (course) => {
  if (!currentSemester.value) {
    showInfo('Atención', 'No hay semestre activo configurado. Por favor, configure un semestre antes de crear secciones.')
    return
  }
  
  selectedCourse.value = course
  sectionForm.name = ''
  showSectionModal.value = true
}

// Manejar envío del formulario de coordinación 
const submitSection = async () => {
  if (!sectionForm.name.trim()) {
    showInfo('Validación', 'Por favor, ingresa el nombre de la coordinación')
    return
  }

  submittingSection.value = true
  
  try {
    const success = await createSection()
    
    if (success) {
      const sectionName = sectionForm.name
      const courseCode = selectedCourse.value?.code || ''
      closeSectionModal()
      // Recargar cursos para mostrar la nueva coordinación 
      await loadCourses()
      showInfo('Éxito', `Coordinación "${sectionName}" creada exitosamente para el curso ${courseCode}`)
    }
  } finally {
    submittingSection.value = false
  }
}

// Cerrar modal de coordinación 
const closeSectionModal = () => {
  showSectionModal.value = false
  selectedCourse.value = null
  sectionForm.name = ''
  sectionForm.professor_id = null
  sectionForm.assistant_id = null
}

// Eliminar coordinación 
const deleteSection = (section) => {
  confirmTitle.value = 'Eliminar Coordinación'
  confirmMessage.value = `¿Estás seguro de que deseas eliminar la coordinación "${section.name}"?`
  confirmBtnText.value = 'Eliminar'
  pendingConfirmAction.value = () => performDeleteSection(section)
  showConfirmModal.value = true
}

const performDeleteSection = async (section) => {
  try {
    const token = getToken()
    
    const response = await fetch(`${API_BASE_URL}/sections/${section.id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })

    if (!response.ok) {
      throw new Error('Error al eliminar la coordinación')
    }

    // Recargar la lista de cursos para actualizar las secciones
    await loadCourses()
    showInfo('Éxito', 'Coordinación eliminada correctamente')
  } catch (error) {
    console.error('Error deleting section:', error)
    showInfo('Error', 'Error al eliminar la coordinación')
  }
}

// Editar coordinación 
const editSection = (section) => {
  sectionToEdit.value = section
  editSectionForm.name = section.name || ''
  showEditSectionModal.value = true
}

const closeEditSectionModal = () => {
  showEditSectionModal.value = false
  sectionToEdit.value = null
  editSectionForm.name = ''
}

const submitEditSection = async () => {
  if (!editSectionForm.name.trim()) {
    showInfo('Validación', 'Por favor, ingresa el nombre de la coordinación')
    return
  }

  submittingEditSection.value = true
  try {
    const token = getToken()
    
    const response = await fetch(`${API_BASE_URL}/sections/${sectionToEdit.value.id}`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        name: editSectionForm.name.trim(),
        course_offering_id: sectionToEdit.value.course_offering_id
      })
    })

    if (!response.ok) {
      throw new Error('Error al actualizar la coordinación')
    }

    closeEditSectionModal()
    // Recargar la lista de cursos para actualizar las secciones
    await loadCourses()
    showInfo('Éxito', 'Coordinación actualizada correctamente')
  } catch (error) {
    console.error('Error updating section:', error)
    showInfo('Error', 'Error al actualizar la coordinación')
  } finally {
    submittingEditSection.value = false
  }
}

// Cerrar modales
const closeModal = () => {
  showCreateModal.value = false
  showEditModal.value = false
  editingCourse.value = null
  resetForm()
}

// Resetear formulario
const resetForm = () => {
  courseForm.code = ''
  courseForm.name = ''
  courseForm.professor_id = null
  courseForm.assistant_id = null
}

// Función para manejar logout
const handleLogout = () => {
  logout()
}

// Filtrar secciones por semestre seleccionado
// Filtrar secciones por semestre seleccionado y rol
const getFilteredSections = (course) => {
  let sections = course.sections || []

  // Si es profesor o ayudante, mostrar solo sus secciones asignadas
  if (user.value?.role_name === 'profesor') {
    sections = sections.filter(section => {
      const pIds = Array.isArray(section.professor_ids) ? section.professor_ids.map(Number) : []
      return Number(section.professor_id) === Number(user.value.id) || pIds.includes(Number(user.value.id))
    })
  } else if (user.value?.role_name === 'ayudante') {
    sections = sections.filter(section => {
      const aIds = Array.isArray(section.assistant_ids) ? section.assistant_ids.map(Number) : []
      return Number(section.assistant_id) === Number(user.value.id) || aIds.includes(Number(user.value.id))
    })
  }

  if (!selectedSemesterFilter.value) {
    return sections
  }
  
  return sections.filter(section => section.semester === selectedSemesterFilter.value)
}

// Obtener semestres únicos de un curso
const getCourseSemesters = (course) => {
  if (!course.sections || course.sections.length === 0) {
    return []
  }
  
  const semesters = new Set(course.sections.map(section => section.semester))
  return allSemesters.value.filter(semester => 
    semesters.has(`${semester.term}_${semester.year}`)
  ).sort((a, b) => {
    if (a.year !== b.year) return b.year - a.year
    return b.term - a.term
  })
}

// Seleccionar semestre
const selectSemester = (courseId, semester) => {
  if (semester.term === 'Todos') {
    selectedSemesterFilter.value = null
  } else {
    selectedSemesterFilter.value = `${semester.term}_${semester.year}`
  }
  showSemesterDropdown.value[courseId] = false
}

// Alternar dropdown de semestres
const toggleSemesterDropdown = (courseId) => {
  showSemesterDropdown.value[courseId] = !showSemesterDropdown.value[courseId]
}

// Ver información de la coordinación 
const viewSection = (section) => {
  navigateTo(`/section/${section.id}`)
}

// Ver coordinación  como estudiante
const viewStudentSection = (section) => {
  navigateTo(`/section/${section.id}`)
}

useHead({
  title: 'TeamEdu - Cursos',
  meta: [
    { name: 'description', content: 'Gestión curricular de cursos en TeamEdu' }
  ]
})
</script>

<style scoped>
/* ════════════════════════════════════════════
   COURSES PAGE – Professional Academic Style
   ════════════════════════════════════════════ */
.courses-page {
  min-height: 100vh;
  background: var(--color-background, #f4f6f9);
}

/* ── PAGE BANNER ── */
.page-banner {
  background: linear-gradient(135deg, var(--color-dark, #394049) 0%, #2a3038 100%);
  padding: 1.8rem 0 1.6rem;
  position: relative;
  overflow: hidden;
}

.page-banner::before {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 35%;
  height: 100%;
  background: linear-gradient(135deg, transparent 0%, rgba(0, 164, 153, 0.06) 100%);
  pointer-events: none;
}

.page-banner::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, var(--color-primary, #00A499), var(--color-accent, #EA7600), var(--color-primary, #00A499));
}

.banner-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.banner-left { min-width: 0; }

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.breadcrumb-link {
  font-size: 0.72rem;
  color: rgba(255,255,255,0.5);
  text-transform: uppercase;
  letter-spacing: 1px;
  font-weight: 500;
  text-decoration: none;
  transition: color 0.18s;
}

.breadcrumb-link:hover {
  color: rgba(255,255,255,0.85);
}

.breadcrumb-sep {
  color: rgba(255,255,255,0.3);
  font-size: 0.72rem;
}

.breadcrumb-item {
  font-size: 0.72rem;
  color: rgba(255,255,255,0.5);
  text-transform: uppercase;
  letter-spacing: 1px;
  font-weight: 500;
}

.breadcrumb-item.active {
  color: rgba(255,255,255,0.8);
}

.banner-title {
  margin: 0;
  font-size: 1.65rem;
  font-weight: 700;
  color: white;
  letter-spacing: -0.3px;
  line-height: 1.2;
}

.banner-subtitle {
  margin: 0.35rem 0 0;
  font-size: 0.85rem;
  color: rgba(255,255,255,0.55);
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.6rem 1.2rem;
  background: var(--color-accent, #b85000);
  border: none;
  border-radius: var(--radius-sm, 6px);
  color: white;
  cursor: pointer;
  transition: all 0.18s;
  font-weight: 600;
  font-size: 0.88rem;
}

.back-btn:hover {
  background: var(--color-accent-dark, #c96500);
  box-shadow: 0 2px 8px rgba(234, 118, 0, 0.3);
}

/* ── MAIN CONTENT ── */
.main-content {
  padding: 2rem 0 3rem;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1.5rem;
}

.courses-container {
  background: var(--color-surface, white);
  border-radius: var(--radius-md, 10px);
  padding: 2rem;
  box-shadow: var(--shadow-md, 0 2px 8px rgba(57,64,73,0.08));
  border: 1px solid var(--color-border, #dfe4ea);
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.75rem;
  padding-bottom: 1rem;
  border-bottom: 2px solid var(--color-border-light, #eceff3);
}

.header-section-left { min-width: 0; }

.courses-list-heading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0;
  color: var(--color-dark, #394049);
  font-size: 0.95rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.courses-list-heading svg {
  color: var(--color-primary, #00A499);
}

.create-btn {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.55rem 1.1rem;
  background: var(--color-primary, #006b64);
  color: white;
  border: none;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  font-weight: 600;
  font-size: 0.85rem;
  transition: all 0.18s;
}

.create-btn:hover {
  background: var(--color-primary-dark, #008a81);
  box-shadow: 0 2px 8px rgba(0, 164, 153, 0.3);
}

.courses-list h3 { display: none; }

/* ── Search ── */
.search-box {
  position: relative;
  margin-bottom: 1.5rem;
  display: flex;
  align-items: center;
  background: var(--color-surface, white);
  border: 2px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-sm, 6px);
  padding: 0.65rem 1rem;
  transition: all 0.18s;
}

.search-box:focus-within {
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.1);
}

.search-icon {
  color: #a0a8b4;
  margin-right: 0.65rem;
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 0.9rem;
  color: var(--color-text, #2b2f36);
  background: transparent;
}

.search-input::placeholder {
  color: #a0a8b4;
}

.clear-search-btn {
  background: none;
  border: none;
  color: #a0a8b4;
  cursor: pointer;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.18s;
  margin-left: 0.5rem;
  flex-shrink: 0;
}

.clear-search-btn:hover {
  color: var(--color-text-secondary, #64707d);
  background: var(--color-background, #f4f6f9);
}

.loading {
  text-align: center;
  color: var(--color-text-secondary, #64707d);
  padding: 2rem;
  font-size: 0.9rem;
}

.no-courses {
  text-align: center;
  color: var(--color-text-secondary, #64707d);
  padding: 3rem;
  font-style: italic;
  font-size: 0.9rem;
}

/* ── Course Cards ── */
.courses-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.25rem;
}

.course-card {
  background: var(--color-surface, white);
  border: 1px solid var(--color-border, #dfe4ea);
  border-top: 4px solid var(--color-primary, #00A499);
  border-radius: var(--radius-md, 10px);
  padding: 1.25rem 1.5rem 1.5rem;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-sm, 0 1px 2px rgba(57,64,73,0.06));
}

.course-card:hover {
  border-color: #c0c7d0;
  box-shadow: var(--shadow-md, 0 2px 8px rgba(57,64,73,0.08));
  transform: translateY(-2px);
}

/* ── Course Card Header ── */
.course-card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.6rem;
}

.course-card-header-left {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.4rem;
  flex: 1;
  min-width: 0;
}

.course-code-badge {
  font-size: 0.82rem;
  font-weight: 700;
  font-family: "JetBrains Mono", "Fira Code", monospace;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  background: var(--color-primary-light, #e6f7f6);
  padding: 0.18rem 0.55rem;
  border-radius: 4px;
  border: 1px solid #b2e8e4;
  color: var(--color-primary-dark, #008a81);
  white-space: nowrap;
}

.course-sections-count {
  font-size: 0.7rem;
  color: var(--color-text-secondary, #64707d);
  background: var(--color-background, #f4f6f9);
  border: 1px solid var(--color-border-light, #eceff3);
  border-radius: 10px;
  padding: 0.1rem 0.5rem;
  white-space: nowrap;
}

.course-card-header-actions {
  display: flex;
  gap: 0.2rem;
  flex-shrink: 0;
}

.card-icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: 1px solid;
  border-radius: 5px;
  cursor: pointer;
  transition: all 0.18s;
  padding: 0;
  background: transparent;
}

.edit-icon-btn {
  border-color: var(--color-border, #dfe4ea);
  color: var(--color-text-secondary, #64707d);
}

.edit-icon-btn:hover {
  background: var(--color-primary-light, #e6f7f6);
  border-color: var(--color-primary, #00A499);
  color: var(--color-primary-dark, #008a81);
}

.delete-icon-btn {
  border-color: #f5c6cb;
  color: var(--color-danger, #d7373f);
  background: var(--color-danger-light, #fceced);
}

.delete-icon-btn:hover {
  background: #f8d7da;
  border-color: #f1aeb5;
}

.course-name {
  margin: 0 0 0.9rem 0;
  color: var(--color-text, #2b2f36);
  line-height: 1.45;
  font-size: 0.95rem;
  font-weight: 600;
}

/* ── Course Card Footer ── */
.course-card-footer {
  margin-top: auto;
  padding-top: 0.75rem;
  border-top: 1px solid var(--color-border-light, #eceff3);
}

.add-section-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  width: 100%;
  padding: 0.48rem 0.5rem;
  background: var(--color-primary-light, #e6f7f6);
  border: 1.5px dashed var(--color-primary, #00A499);
  border-radius: var(--radius-sm, 6px);
  color: var(--color-primary-dark, #008a81);
  font-size: 0.82rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.18s;
}

.add-section-btn:hover {
  background: #ccf0ed;
  border-style: solid;
  border-color: var(--color-primary-dark, #008a81);
}

.course-actions {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 0.4rem;
  margin-top: auto;
}

.edit-btn, .delete-btn, .section-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  padding: 0.4rem 0.5rem;
  border: 1px solid;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  font-size: 0.78rem;
  font-weight: 500;
  transition: all 0.18s;
  text-align: center;
  white-space: nowrap;
  min-width: 0;
  overflow: hidden;
}

.edit-btn {
  background: var(--color-background, #f4f6f9);
  border-color: var(--color-border, #dfe4ea);
  color: var(--color-text-secondary, #64707d);
}

.edit-btn:hover {
  background: var(--color-border-light, #eceff3);
  border-color: #c0c7d0;
  color: var(--color-dark, #394049);
}

.section-btn {
  background: var(--color-primary-light, #e6f7f6);
  border-color: #b2e8e4;
  color: var(--color-primary-dark, #008a81);
}

.section-btn:hover {
  background: #ccf0ed;
  border-color: var(--color-primary, #00A499);
}

/* ── Course Sections ── */
.course-sections {
  margin: 1rem 0;
  padding: 0.85rem;
  background: var(--color-background, #f4f6f9);
  border: 1px solid var(--color-border-light, #eceff3);
  border-radius: var(--radius-sm, 6px);
}

.sections-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.6rem;
}

.course-sections h5 {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  margin: 0;
  color: var(--color-dark, #394049);
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: 0.025em;
}

/* Semester filter */
.semester-filter { position: relative; }
.semester-dropdown { position: relative; }

.semester-dropdown-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.3rem 0.65rem;
  background: var(--color-surface, white);
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: 4px;
  color: var(--color-text-secondary, #64707d);
  font-size: 0.72rem;
  cursor: pointer;
  transition: all 0.18s;
  min-width: 110px;
  justify-content: space-between;
}

.semester-dropdown-btn:hover {
  border-color: var(--color-primary, #00A499);
}

.semester-dropdown.active .semester-dropdown-btn {
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 0 0 2px rgba(0, 164, 153, 0.1);
}

.semester-dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  background: var(--color-surface, white);
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-sm, 6px);
  box-shadow: var(--shadow-lg, 0 4px 16px rgba(57,64,73,0.10));
  z-index: 1000;
  min-width: 200px;
  max-height: 200px;
  overflow-y: auto;
}

.semester-options {
  max-height: 150px;
  overflow-y: auto;
}

.semester-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0.45rem 0.7rem;
  border: none;
  background: var(--color-surface, white);
  color: var(--color-text, #2b2f36);
  font-size: 0.82rem;
  cursor: pointer;
  transition: background 0.15s;
  text-align: left;
}

.semester-option:hover {
  background: var(--color-primary-light, #e6f7f6);
}

.semester-option.selected {
  background: var(--color-primary-light, #e6f7f6);
  color: var(--color-primary-dark, #008a81);
  font-weight: 600;
}

.current-badge {
  background: var(--color-primary, #006b64);
  color: white;
  font-size: 0.62rem;
  padding: 0.1rem 0.35rem;
  border-radius: 10px;
  font-weight: 600;
}

.sections-list {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.section-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.45rem 0.7rem;
  background: var(--color-surface, white);
  border: 1px solid var(--color-border-light, #eceff3);
  border-radius: 4px;
  transition: all 0.18s;
}

.section-item:hover {
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 1px 3px rgba(0, 164, 153, 0.1);
}

.section-name {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  color: var(--color-dark, #394049);
  font-weight: 500;
  font-size: 0.84rem;
  flex: 1;
  min-width: 0;
}

.section-arrow-icon {
  flex-shrink: 0;
  color: #c0c7d0;
  transition: color 0.18s;
}

.section-name.clickable {
  cursor: pointer;
  transition: color 0.18s;
}

.section-name.clickable:hover {
  color: var(--color-primary, #00A499);
  text-decoration: underline;
}

.section-name.clickable:hover .section-arrow-icon {
  color: var(--color-primary, #00A499);
}

.section-actions {
  display: flex;
  gap: 0.2rem;
  flex-shrink: 0;
}

.section-edit-btn, .section-delete-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border: 1px solid;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.18s;
  padding: 0;
}

.section-edit-btn {
  background: var(--color-background, #f4f6f9);
  border-color: var(--color-border, #dfe4ea);
  color: var(--color-text-secondary, #64707d);
}

.section-edit-btn:hover {
  background: var(--color-primary-light, #e6f7f6);
  border-color: var(--color-primary, #00A499);
  color: var(--color-primary-dark, #008a81);
}

.section-delete-btn {
  background: var(--color-danger-light, #fceced);
  border-color: #f5c6cb;
  color: var(--color-danger, #d7373f);
}

.section-delete-btn:hover {
  background: #f8d7da;
  border-color: #f1aeb5;
}

.no-sections {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0 0 0.85rem 0;
  padding: 0.6rem 0.8rem;
  background: var(--color-background, #f4f6f9);
  border: 1px dashed var(--color-border, #dfe4ea);
  border-radius: var(--radius-sm, 6px);
  color: var(--color-text-secondary, #64707d);
}

.no-sections-icon {
  flex-shrink: 0;
  color: #c0c7d0;
}

.no-sections-text {
  margin: 0;
  color: var(--color-text-secondary, #64707d);
  font-size: 0.82rem;
  font-style: italic;
  text-align: center;
}

.role-info {
  display: grid;
  gap: 1.25rem;
}

.role-info p {
  color: var(--color-text-secondary, #64707d);
  margin-bottom: 1.5rem;
  line-height: 1.6;
}

.info-card {
  background: var(--color-background, #f4f6f9);
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-md, 10px);
  padding: 1.5rem;
}

.info-card h3 {
  margin: 0 0 0.85rem 0;
  color: var(--color-dark, #394049);
  font-size: 1.1rem;
}

.info-card ul {
  margin: 0;
  padding-left: 1.25rem;
  color: var(--color-text-secondary, #64707d);
}

.info-card li {
  margin-bottom: 0.4rem;
  line-height: 1.5;
  font-size: 0.9rem;
}

/* ── Modal ── */
.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(57, 64, 73, 0.45);
  backdrop-filter: blur(3px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal {
  background: var(--color-surface, white);
  border-radius: var(--radius-lg, 14px);
  max-width: 500px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: var(--shadow-xl, 0 8px 28px rgba(57,64,73,0.12));
  border: 1px solid var(--color-border, #dfe4ea);
}

.confirm-modal { max-width: 420px; }

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.modal-header h3 {
  margin: 0;
  color: var(--color-dark, #394049);
  font-size: 1.1rem;
  font-weight: 700;
}

.close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border: none;
  background: var(--color-background, #f4f6f9);
  border-radius: var(--radius-sm, 6px);
  color: var(--color-text-secondary, #64707d);
  cursor: pointer;
  transition: all 0.18s;
}

.close-btn:hover {
  background: var(--color-border-light, #eceff3);
  color: var(--color-dark, #394049);
}

.modal-body { padding: 1.5rem; }

.confirm-modal .modal-body {
  padding: 1.5rem 1.5rem 1rem;
}

.confirm-modal .modal-body p {
  margin: 0 0 0.75rem 0;
  line-height: 1.6;
  color: var(--color-text, #2b2f36);
  font-size: 0.9rem;
}

.form-group { margin-bottom: 1.5rem; }

.form-group label {
  display: block;
  margin-bottom: 0.4rem;
  color: var(--color-dark, #394049);
  font-weight: 600;
  font-size: 0.88rem;
}

.form-group input {
  width: 100%;
  padding: 0.65rem 0.85rem;
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-sm, 6px);
  font-size: 0.9rem;
  color: var(--color-text, #2b2f36);
  transition: border-color 0.18s;
}

.form-group input:focus {
  outline: none;
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.1);
}

.form-group input:disabled {
  background: var(--color-background, #f4f6f9);
  color: var(--color-text-secondary, #64707d);
  cursor: not-allowed;
}

.form-group small {
  display: block;
  margin-top: 0.25rem;
  color: var(--color-text-secondary, #64707d);
  font-size: 0.8rem;
}

.form-group select {
  width: 100%;
  padding: 0.6rem 0.85rem;
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-sm, 6px);
  background: var(--color-surface, white);
  color: var(--color-text, #2b2f36);
  font-size: 0.9rem;
  transition: border-color 0.18s, box-shadow 0.18s;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%2364707d' stroke-width='2'%3E%3Cpolyline points='6,9 12,15 18,9'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 0.75rem center;
  padding-right: 2rem;
}

.form-group select:focus {
  outline: none;
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.1);
}

.form-group select:disabled {
  background-color: var(--color-background, #f4f6f9);
  color: var(--color-text-secondary, #64707d);
  cursor: not-allowed;
}

/* Two-column form row */
.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.85rem;
  margin-bottom: 1.5rem;
}

@media (max-width: 480px) {
  .form-row {
    grid-template-columns: 1fr;
  }
}

.form-row .form-group {
  margin-bottom: 0;
}

/* Optional field badge */
.optional-badge {
  display: inline-block;
  margin-left: 0.35rem;
  font-size: 0.68rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  background: var(--color-background, #f4f6f9);
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: 10px;
  padding: 0.05rem 0.45rem;
  color: var(--color-text-secondary, #64707d);
  vertical-align: middle;
}

/* Staff assignment hint */
.staff-assign-hint {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  margin-bottom: 1rem;
  padding: 0.6rem 0.8rem;
  background: var(--color-primary-light, #e6f7f6);
  border: 1px solid #b2e8e4;
  border-radius: var(--radius-sm, 6px);
  color: var(--color-primary-dark, #008a81);
  font-size: 0.82rem;
  line-height: 1.4;
}

.staff-assign-hint svg {
  flex-shrink: 0;
  margin-top: 0.1rem;
}

.modal-actions {
  display: flex;
  gap: 0.65rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
  padding: 1rem 1.5rem 1.5rem;
  border-top: 1px solid var(--color-border-light, #eceff3);
}

.cancel-btn, .submit-btn {
  padding: 0.6rem 1.25rem;
  border: 1px solid;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  font-weight: 600;
  transition: all 0.18s;
  font-size: 0.88rem;
  min-width: 95px;
}

.cancel-btn {
  background: var(--color-surface, white);
  border-color: var(--color-border, #dfe4ea);
  color: var(--color-text-secondary, #64707d);
}

.cancel-btn:hover {
  background: var(--color-background, #f4f6f9);
  border-color: #c0c7d0;
}

.submit-btn {
  background: var(--color-primary, #006b64);
  border-color: var(--color-primary, #006b64);
  color: white;
}

.submit-btn:hover:not(:disabled) {
  background: var(--color-primary-dark, #008a81);
  border-color: var(--color-primary-dark, #008a81);
}

.submit-btn:disabled {
  background: #b0b8c1;
  border-color: #b0b8c1;
  cursor: not-allowed;
}

.modal-delete-btn {
  background: var(--color-danger, #d7373f);
  border-color: var(--color-danger, #d7373f);
  color: white;
  padding: 0.6rem 1.25rem;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  font-weight: 600;
  transition: all 0.18s;
  font-size: 0.88rem;
  min-width: 95px;
  border-style: solid;
  border-width: 1px;
}

.modal-delete-btn:hover:not(:disabled) {
  background: #b82d34;
  border-color: #b82d34;
}

.modal-delete-btn:disabled {
  background: #e79195;
  border-color: #e79195;
  cursor: not-allowed;
  opacity: 0.7;
}

.warning {
  color: var(--color-accent, #EA7600);
  font-weight: 600;
  margin-top: 0.35rem;
  font-size: 0.88rem;
}

.info-section {
  background: var(--color-background, #f4f6f9);
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-sm, 6px);
  padding: 1rem;
  margin-bottom: 1.5rem;
}

.info-section h4 {
  margin: 0 0 0.6rem 0;
  color: var(--color-dark, #394049);
  font-size: 0.95rem;
}

.info-section p {
  margin: 0.2rem 0;
  color: var(--color-text-secondary, #64707d);
  font-size: 0.85rem;
}

.info-section strong {
  color: var(--color-dark, #394049);
}

/* ── Student Course Cards ── */
.student-course-card {
  background: var(--color-surface, white);
  border: 1px solid var(--color-border, #dfe4ea);
  border-top: 4px solid var(--color-primary, #00A499);
  border-radius: var(--radius-md, 10px);
  padding: 1.5rem;
  transition: all 0.22s ease;
  cursor: pointer;
  position: relative;
}

.student-course-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg, 0 4px 16px rgba(57,64,73,0.10));
  border-color: var(--color-primary, #00A499);
}

.student-course-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.student-course-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--color-dark, #394049);
  margin: 0;
  flex: 1;
}

.student-course-code {
  background: var(--color-primary, #006b64);
  color: white;
  padding: 0.2rem 0.65rem;
  border-radius: 12px;
  font-size: 0.72rem;
  font-weight: 600;
  margin-left: 0.75rem;
  letter-spacing: 0.3px;
}

.student-course-info {
  display: grid;
  gap: 0.6rem;
  margin-bottom: 1.25rem;
}

.student-course-detail {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.85rem;
  color: var(--color-text-secondary, #64707d);
}

.student-course-detail svg {
  width: 15px;
  height: 15px;
  color: var(--color-text-secondary, #64707d);
  opacity: 0.6;
}

.student-course-action {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 0.45rem;
  padding: 0.65rem 1.25rem;
  background: var(--color-primary, #006b64);
  color: white;
  border: none;
  border-radius: var(--radius-sm, 6px);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.18s ease;
  font-size: 0.85rem;
  width: 100%;
}

.student-course-action:hover {
  background: var(--color-primary-dark, #008a81);
  box-shadow: 0 2px 8px rgba(0, 164, 153, 0.3);
}

.student-course-action svg {
  width: 15px;
  height: 15px;
}

.student-no-courses {
  text-align: center;
  padding: 3rem 2rem;
  background: var(--color-background, #f4f6f9);
  border: 2px dashed var(--color-border, #dfe4ea);
  border-radius: var(--radius-md, 10px);
  color: var(--color-text-secondary, #64707d);
}

.student-no-courses-icon {
  width: 44px;
  height: 44px;
  color: var(--color-border, #dfe4ea);
  margin: 0 auto 1rem;
}

.student-no-courses h3 {
  margin: 0 0 0.4rem 0;
  color: var(--color-dark, #394049);
  font-size: 1.15rem;
}

.student-no-courses p {
  margin: 0;
  font-size: 0.85rem;
  line-height: 1.5;
}

.student-courses-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 1.25rem;
  margin-top: 1.25rem;
}

/* ── Responsive ── */
@media (max-width: 768px) {
  .header-section {
    flex-direction: column;
    gap: 0.75rem;
    align-items: flex-start;
  }

  .courses-grid { grid-template-columns: 1fr; }
  .student-courses-grid { grid-template-columns: 1fr; }
  .course-actions { grid-template-columns: 1fr; }
  .modal-actions { flex-direction: column; }

  .banner-title { font-size: 1.3rem; }
  .container { padding: 0 1rem; }
  .courses-container { padding: 1.5rem; }
}

@media (max-width: 480px) {
  .page-banner { padding: 1.5rem 0 1.25rem; }
  .banner-title { font-size: 1.15rem; }
  .container { padding: 0 0.75rem; }
  .courses-container { padding: 1rem; }
}
</style>
