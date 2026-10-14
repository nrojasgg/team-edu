<template>
  <div class="home-page">
    <!-- Header con información del usuario -->
    <TopBar :user="user" @logout="handleLogout" />

    <!-- Hero/Welcome Banner -->
    <div class="welcome-banner">
      <div class="container">
        <div class="welcome-inner">
          <div class="welcome-text">
            <h1 class="welcome-title">Bienvenido, {{ userFirstName }}</h1>
            <p class="welcome-subtitle">Portal de Gestión Académica — TeamEdu</p>
          </div>
          <div class="welcome-meta" v-if="formatSemesterName">
            <span class="semester-badge">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="4" width="18" height="18" rx="2"/>
                <line x1="16" y1="2" x2="16" y2="6"/>
                <line x1="8" y1="2" x2="8" y2="6"/>
                <line x1="3" y1="10" x2="21" y2="10"/>
              </svg>
              {{ formatSemesterName }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Contenido principal -->
    <main class="main-content">
      <div class="container">
        <!-- Panel de navegación rápida -->
        <section class="quick-actions">
          <div class="section-header">
            <h2>
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="3" width="7" height="7"/>
                <rect x="14" y="3" width="7" height="7"/>
                <rect x="3" y="14" width="7" height="7"/>
                <rect x="14" y="14" width="7" height="7"/>
              </svg>
              Acciones Rápidas
            </h2>
          </div>
          <div class="actions-grid">
            <!-- Acción principal: Cursos -->
            <div class="action-card featured">
              <div class="card-accent"></div>
              <div class="card-body">
                <div class="card-icon-wrap primary">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/>
                    <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>
                  </svg>
                </div>
                <h3>Cursos</h3>
                <p v-if="isStudent">Revisa los cursos en los que estás inscrito</p>
                <p v-else-if="isTeacher">Gestiona tus cursos asignados</p>
                <p v-else>Administra los cursos del sistema</p>
                <button class="action-btn primary" @click="navigateToCourses">
                  {{ isStudent ? 'Ver Mis Cursos' : 'Gestionar Cursos' }}
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M5 12h14"/><path d="m12 5 7 7-7 7"/>
                  </svg>
                </button>
              </div>
            </div>

            <!-- Configuración (solo para admin) -->
            <div v-if="isAdmin" class="action-card">
              <div class="card-accent accent"></div>
              <div class="card-body">
                <div class="card-icon-wrap accent">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="3"/>
                    <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
                  </svg>
                </div>
                <h3>Configuración</h3>
                <p>Gestiona semestre, usuarios y permisos</p>
                <button class="action-btn accent" @click="showConfigurationModal = true">
                  Abrir Configuración
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M5 12h14"/><path d="m12 5 7 7-7 7"/>
                  </svg>
                </button>
              </div>
            </div>

            <!-- Plugins (solo para admin) -->
            <div v-if="isAdmin" class="action-card">
              <div class="card-accent plugins"></div>
              <div class="card-body">
                <div class="card-icon-wrap plugins">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M12 2L2 7l10 5 10-5-10-5z"/>
                    <path d="M2 17l10 5 10-5"/>
                    <path d="M2 12l10 5 10-5"/>
                  </svg>
                </div>
                <h3>Plugins</h3>
                <p>Gestiona los plugins cargados en el sistema</p>
                <button class="action-btn plugins-btn" @click="openPluginsModal">
                  Ver Plugins
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M5 12h14"/><path d="m12 5 7 7-7 7"/>
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <!-- Botón para dejar de simular (visible cuando se está simulando) -->
          <div v-if="isImpersonating" class="impersonate-panel">
            <div class="impersonate-inner">
              <div class="impersonate-icon">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                  <circle cx="9" cy="7" r="4"/>
                  <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
                  <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                </svg>
              </div>
              <div class="impersonate-text">
                <strong>Simulación Activa</strong>
                <span>Estás visualizando la aplicación como otro usuario.</span>
              </div>
              <button class="impersonate-stop-btn" @click="restoreOriginalToken">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
                Dejar de simular
              </button>
            </div>
          </div>
        </section>
      </div>
    </main>

    <!-- Modal de cambio de semestre -->
    <ChangeSemesterModal
      :is-open="showChangeSemesterModal"
      @close="closeChangeSemesterModal"
      @semester-changed="onSemesterChanged"
    />

    <!-- Modal de configuración -->
    <div v-if="showConfigurationModal" class="modal-overlay" @click="closeConfigurationModal">
      <div class="modal-content configuration-modal" @click.stop>
        <div class="modal-header">
          <h3>Configuración</h3>
          <button class="modal-close" @click="closeConfigurationModal">&times;</button>
        </div>
        <div class="modal-body configuration-body">
          <div class="config-options">
            <!-- Opción: Cambiar Semestre -->
            <button class="config-option" @click="openChangeSemesterFromConfig">
              <div class="option-icon">📅</div>
              <div class="option-content">
                <h4>Cambiar Semestre</h4>
                <p>Semestre actual: {{ formatSemesterName }}</p>
              </div>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="9 18 15 12 9 6"/>
              </svg>
            </button>

            <!-- Opción: Simular Usuario -->
            <button class="config-option" @click="openImpersonateFromConfig">
              <div class="option-icon">🎭</div>
              <div class="option-content">
                <h4>Simular Usuario</h4>
                <p>Prueba la aplicación como otro usuario</p>
              </div>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="9 18 15 12 9 6"/>
              </svg>
            </button>

            <!-- Opción: Gestionar Permisos -->
            <button class="config-option" @click="navigateToConfiguracion">
              <div class="option-icon">🔐</div>
              <div class="option-content">
                <h4>Gestionar Permisos</h4>
                <p>Configura roles y permisos del sistema</p>
              </div>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="9 18 15 12 9 6"/>
              </svg>
            </button>

            <!-- Opción: Gestionar Usuarios -->
            <button class="config-option" @click="openManageUsersModal">
              <div class="option-icon">👥</div>
              <div class="option-content">
                <h4>Gestionar Usuarios</h4>
                <p>Edita y administra todos los usuarios del sistema</p>
              </div>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="9 18 15 12 9 6"/>
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal de simulación de usuario -->
    <div v-if="showImpersonateModal" class="modal-overlay" @click="closeImpersonateModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>Simular Usuario</h3>
          <button class="modal-close" @click="closeImpersonateModal">&times;</button>
        </div>
        <div class="modal-body">
          <p>Selecciona un usuario para simular su experiencia en la aplicación:</p>
          
          <div v-if="modalLoadingUsers" class="loading-state">
            <div class="spinner"></div>
            <p>Cargando usuarios...</p>
          </div>
          
          <div v-else-if="modalUsers.length === 0" class="empty-state">
            <p>No se pudieron cargar los usuarios.</p>
            <button class="retry-btn" @click="loadModalUsers">Reintentar</button>
          </div>
          
          <div v-else class="user-selection">
            <select v-model="modalSelectedUserId" class="user-select">
              <option value="">Selecciona un usuario</option>
              <option v-for="u in modalUsers" :key="u.id" :value="u.id">
                {{ u.full_name }} ({{ u.email }})
              </option>
            </select>
          </div>
        </div>
        <div class="modal-footer">
          <button class="cancel-btn" @click="closeImpersonateModal">Cancelar</button>
          <button class="confirm-btn" :disabled="!modalSelectedUserId || modalImpersonating" @click="performImpersonation">
            {{ modalImpersonating ? 'Simulando...' : 'Simular' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal de gestionar usuarios -->
    <ManageUsersModal
      :is-open="showManageUsersModal"
      @close="closeManageUsersModal"
    />

    <!-- Modal de Plugins -->
    <div v-if="showPluginsModal" class="modal-overlay" @click="closePluginsModal">
      <div class="modal-content plugins-modal" @click.stop>
        <div class="modal-header">
          <h3>Plugins del Sistema</h3>
          <button class="modal-close" @click="closePluginsModal">&times;</button>
        </div>
        <div class="modal-body">
          <button class="add-plugin-btn" @click="handleAddPlugin">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
            </svg>
            Agregar plugin
          </button>
          <div v-if="loadingPlugins" class="loading-state">
            <div class="spinner"></div>
            <p>Cargando plugins...</p>
          </div>
          <div v-else-if="pluginsError" class="empty-state">
            <p>{{ pluginsError }}</p>
            <button class="retry-btn" @click="loadPlugins">Reintentar</button>
          </div>
          <div v-else-if="plugins.length === 0" class="empty-state">
            <p>No hay plugins cargados en el sistema.</p>
          </div>
          <div v-else class="plugins-list">
            <div v-for="plugin in plugins" :key="plugin.nombre" class="plugin-item" @click="openPluginDetail(plugin)">
              <div class="plugin-icon">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M12 2L2 7l10 5 10-5-10-5z"/>
                  <path d="M2 17l10 5 10-5"/>
                  <path d="M2 12l10 5 10-5"/>
                </svg>
              </div>
              <div class="plugin-info">
                <h4>{{ plugin.nombre }}</h4>
                <p>v{{ plugin.version }} · {{ plugin.tipo_plugin === 'conformacion_equipos' ? 'Formación de equipos' : 'Evaluación entre pares' }}</p>
                <p class="plugin-description">{{ plugin.descripcion }}</p>
              </div>
              <div class="plugin-status">
                <span :class="['status-badge', plugin.activo ? 'enabled' : 'disabled']">
                  {{ plugin.activo ? 'Activo' : 'Inactivo' }}
                </span>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="cancel-btn" @click="closePluginsModal">Cerrar</button>
        </div>
      </div>
    </div>

    <!-- Modal de Detalle del Plugin -->
    <div v-if="showPluginDetailModal && selectedPlugin" class="modal-overlay" @click="closePluginDetail">
      <div class="modal-content plugin-detail-modal" @click.stop>
        <div class="modal-header">
          <h3>{{ selectedPlugin.nombre }}</h3>
          <button class="modal-close" @click="closePluginDetail">&times;</button>
        </div>
        <div class="modal-body">
          <div class="detail-section">
            <h4>Información General</h4>
            <div class="detail-grid">
              <div class="detail-item">
                <span class="detail-label">Versión</span>
                <span class="detail-value">{{ selectedPlugin.version }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">Autor</span>
                <span class="detail-value">{{ selectedPlugin.autor }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">Tipo de Plugin</span>
                <span class="detail-value">{{ selectedPlugin.tipo_plugin === 'conformacion_equipos' ? 'Formación de Equipos' : 'Evaluación entre Pares' }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">Tipo de Respuestas</span>
                <span class="detail-value">{{ selectedPlugin.tipo_respuestas }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">Estado</span>
                <button 
                  class="toggle-btn"
                  :class="selectedPlugin.activo ? 'active' : 'inactive'"
                  @click="togglePluginStatus"
                  :disabled="togglingPlugin"
                >
                  <span class="toggle-slider"></span>
                  <span class="toggle-text">{{ selectedPlugin.activo ? 'Activado' : 'Desactivado' }}</span>
                </button>
              </div>
            </div>
          </div>

          <div class="detail-section">
            <h4>Descripción</h4>
            <p class="description-text">{{ selectedPlugin.descripcion }}</p>
          </div>

          <div class="detail-section" v-if="selectedPlugin.preguntas && Object.keys(selectedPlugin.preguntas).length > 0">
            <h4>Preguntas ({{ Object.keys(selectedPlugin.preguntas).length }})</h4>
            <div class="questions-list">
              <div v-for="(text, id) in selectedPlugin.preguntas" :key="id" class="question-item">
                <span class="question-id">{{ id }}</span>
                <span class="question-text">{{ text }}</span>
              </div>
            </div>
          </div>

          <div class="detail-section" v-if="selectedPlugin.resultados && selectedPlugin.resultados.length > 0">
            <h4>Resultados que Entrega</h4>
            <div class="results-list">
              <div v-for="result in selectedPlugin.resultados" :key="result.nombre" class="result-item">
                <div class="result-header">
                  <span class="result-name">{{ result.nombre }}</span>
                  <span :class="['result-type-badge', result.tipo]">{{ result.tipo }}</span>
                </div>
                <p class="result-description">{{ result.descripcion }}</p>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="cancel-btn" @click="closePluginDetail">Cerrar</button>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

// Aplicar middleware de autenticación
definePageMeta({
  middleware: 'auth'
})

const { getUser, logout } = useAuth()

// Estado del usuario
const user = ref(null)

// Estado del modal de cambio de semestre
const showChangeSemesterModal = ref(false)

// Estado del modal de configuración
const showConfigurationModal = ref(false)

// Estado del modal de simulación de usuario
const showImpersonateModal = ref(false)

// Estado del modal de gestionar usuarios
const showManageUsersModal = ref(false)

// Estado del modal de plugins
const showPluginsModal = ref(false)
const plugins = ref([])
const loadingPlugins = ref(false)
const pluginsError = ref(null)
const selectedPlugin = ref(null)
const showPluginDetailModal = ref(false)
const togglingPlugin = ref(false)

// Estado del modal de simulación
const modalUsers = ref([])
const modalLoadingUsers = ref(false)
const modalSelectedUserId = ref('')
const modalImpersonating = ref(false)

// Computadas para verificar roles
const isStudent = computed(() => user.value?.role_name === 'estudiante')
const isTeacher = computed(() => user.value?.role_name === 'profesor')
const isCoordinator = computed(() => user.value?.role_name === 'coordinador')
const isAdmin = computed(() => user.value?.role_name === 'admin')

// Computada para primer nombre
const userFirstName = computed(() => {
  const name = user.value?.fullname || user.value?.nombres;
  if (!name) return 'Usuario';
  return name.split(' ')[0];
})

// Semester composable
const { formatSemesterName, getCurrentSemester } = useSemester()

// Impersonate logic
const isImpersonating = ref(false)

const API_BASE_URL = 'http://127.0.0.1:3000/api/v1'

// Verificar si se está simulando
const checkIfImpersonating = () => {
  isImpersonating.value = !!localStorage.getItem('original_auth_token')
}

// Guardar el token original antes de simular
const saveOriginalToken = () => {
  const token = localStorage.getItem('auth_token')
  const userInfo = localStorage.getItem('user_info')
  if (token) {
    localStorage.setItem('original_auth_token', token)
  }
  if (userInfo) {
    localStorage.setItem('original_user_info', userInfo)
  }
}

const restoreOriginalToken = async () => {
  const originalToken = localStorage.getItem('original_auth_token')
  const originalUser = localStorage.getItem('original_user_info')
  if (originalToken) {
    localStorage.setItem('auth_token', originalToken)
    // Restaurar datos del usuario original si están guardados
    if (originalUser) {
      localStorage.setItem('user_info', originalUser)
    }
    localStorage.removeItem('original_auth_token')
    localStorage.removeItem('original_user_info')
    alert('Has dejado de simular. Ahora eres tú nuevamente.')
    globalThis.location.reload()
  } else {
    alert('No se encontró el token original.')
  }
}

const impersonateUser = async (userId) => {
  saveOriginalToken()
  try {
    const token = localStorage.getItem('auth_token')
    const response = await fetch(`${API_BASE_URL}/users/${userId}/impersonate`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })
    if (!response.ok) {
      const errorData = await response.json()
      alert(errorData.message || 'No se pudo simular el usuario')
      return
    }
    const result = await response.json()
    const newToken = result.data?.token
    const simulatedUser = result.data?.user
    if (newToken && simulatedUser) {
      localStorage.setItem('auth_token', newToken)
      // Guardar directamente los datos del usuario simulado devueltos por el backend
      localStorage.setItem('user_info', JSON.stringify(simulatedUser))
      alert('Ahora estás simulando al usuario seleccionado.')
      globalThis.location.reload()
    } else {
      alert('No se pudo obtener el token de simulación.')
    }
  } catch (err) {
    alert('Error al simular usuario: ' + err.message)
  }
}

onMounted(() => {
  user.value = getUser()
  checkIfImpersonating()
})

// Métodos
const handleLogout = () => {
  logout()
}

const navigateToCourses = () => {
  navigateTo('/courses')
}

const navigateToConfiguracion = () => {
  closeConfigurationModal()
  navigateTo('/configuracion')
}

// Métodos para el modal de configuración
const closeConfigurationModal = () => {
  showConfigurationModal.value = false
}

const openChangeSemesterFromConfig = () => {
  closeConfigurationModal()
  showChangeSemesterModal.value = true
}

const openImpersonateFromConfig = () => {
  closeConfigurationModal()
  openImpersonateModal()
}

// Métodos para el modal de gestionar usuarios
const openManageUsersModal = () => {
  closeConfigurationModal()
  showManageUsersModal.value = true
}

const closeManageUsersModal = () => {
  showManageUsersModal.value = false
}

// Funciones para el modal de plugins
const openPluginsModal = async () => {
  showPluginsModal.value = true
  await loadPlugins()
}

const closePluginsModal = () => {
  showPluginsModal.value = false
  pluginsError.value = null
}

const loadPlugins = async () => {
  loadingPlugins.value = true
  pluginsError.value = null
  try {
    const token = localStorage.getItem('auth_token')
    const response = await fetch(`${API_BASE_URL}/core/plugins/metadata`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })
    if (response.ok) {
      const result = await response.json()
      if (result.success && Array.isArray(result.data)) {
        plugins.value = result.data
      } else if (Array.isArray(result.data)) {
        plugins.value = result.data
      } else {
        plugins.value = []
      }
    } else if (response.status === 401) {
      pluginsError.value = 'No tienes permiso para ver los plugins.'
      plugins.value = []
    } else {
      pluginsError.value = 'Error al cargar los plugins.'
      plugins.value = []
    }
  } catch (err) {
    console.error('Error loading plugins:', err)
    pluginsError.value = 'Error de conexión al cargar los plugins.'
    plugins.value = []
  }
  loadingPlugins.value = false
}

const openPluginDetail = (plugin) => {
  selectedPlugin.value = plugin
  showPluginDetailModal.value = true
}

const closePluginDetail = () => {
  showPluginDetailModal.value = false
  selectedPlugin.value = null
}

const handleAddPlugin = () => {
  closePluginsModal()
  navigateTo('/admin/plugins')
}

const togglePluginStatus = async () => {
  if (!selectedPlugin.value || togglingPlugin.value) return
  
  togglingPlugin.value = true
  const newStatus = !selectedPlugin.value.activo
  const action = newStatus ? 'enable' : 'disable'
  const pluginName = selectedPlugin.value.nombre_interno || selectedPlugin.value.nombre
  
  try {
    const token = localStorage.getItem('auth_token')
    const response = await fetch(`${API_BASE_URL}/core/plugins/${pluginName}/${action}`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })
    
    if (response.ok) {
      selectedPlugin.value.activo = newStatus
      const pluginIndex = plugins.value.findIndex(p => p.nombre === selectedPlugin.value.nombre)
      if (pluginIndex !== -1) {
        plugins.value[pluginIndex].activo = newStatus
      }
    } else {
      const errorData = await response.json()
      alert(errorData.message || 'No se pudo cambiar el estado del plugin')
    }
  } catch (err) {
    console.error('Error toggling plugin status:', err)
    alert('Error de conexión al cambiar el estado del plugin')
  }
  
  togglingPlugin.value = false
}

// Cargar datos del usuario al montar el componente
onMounted(async () => {
  user.value = getUser()
  checkIfImpersonating()
  
  console.log('Usuario cargado:', user.value) // Debug
  console.log('Picture URL:', user.value?.picture_url) // Debug adicional
  console.log('¿Es admin?:', user.value?.role_name === 'admin') // Debug
  
  // Debug del token
  const { getToken } = useAuth()
  const token = getToken()
  console.log('Token disponible en home:', token) // Debug adicional
  
  // Si no hay usuario, redirigir al login
  if (!user.value) {
    navigateTo('/')
    return
  }
  
  // Cargar datos específicos del usuario
  loadUserData()
  
  // Cargar semestre actual
  await getCurrentSemester()
})

const loadUserData = async () => {
  // Simular carga de datos
  // En una implementación real, harías llamadas a la API aquí
  console.log('Datos del usuario cargados')
}

// Funciones para el modal de simulación
const openImpersonateModal = async () => {
  showImpersonateModal.value = true
  modalSelectedUserId.value = ''
  await loadModalUsers()
}

const closeImpersonateModal = () => {
  showImpersonateModal.value = false
  modalUsers.value = []
  modalSelectedUserId.value = ''
}

// Funciones para el modal de cambio de semestre
const closeChangeSemesterModal = () => {
  showChangeSemesterModal.value = false
}

const onSemesterChanged = async () => {
  showChangeSemesterModal.value = false
  await getCurrentSemester()
}

const loadModalUsers = async () => {
  modalLoadingUsers.value = true
  try {
    const token = localStorage.getItem('auth_token')
    const response = await fetch(`${API_BASE_URL}/users?limit=1000`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })
    if (response.ok) {
      const result = await response.json()
      if (Array.isArray(result.data)) {
        modalUsers.value = result.data
      } else if (Array.isArray(result.users)) {
        modalUsers.value = result.users
      } else {
        modalUsers.value = []
      }
    } else {
      modalUsers.value = []
    }
  } catch (err) {
    console.error('Error loading users:', err)
    modalUsers.value = []
  }
  modalLoadingUsers.value = false
}

const performImpersonation = async () => {
  if (!modalSelectedUserId.value) return
  
  modalImpersonating.value = true
  try {
    await impersonateUser(modalSelectedUserId.value)
    closeImpersonateModal()
  } catch (err) {
    console.error('Error during impersonation:', err)
  } finally {
    modalImpersonating.value = false
  }
}

// Meta información
useHead({
  title: `TeamEdu - Portal Principal`,
  meta: [
    { name: 'description', content: 'Portal principal del sistema de registro curricular TeamEdu' }
  ]
})
</script>

<style scoped>
/* ════════════════════════════════════════════
   HOME PAGE – Professional Academic Portal
   ════════════════════════════════════════════ */
.home-page {
  min-height: 100vh;
  background: var(--color-background, #f4f6f9);
}

/* ── Welcome Banner ── */
.welcome-banner {
  background: linear-gradient(135deg, var(--color-dark, #394049) 0%, #2a3038 100%);
  padding: 2rem 0;
  position: relative;
  overflow: hidden;
}

.welcome-banner::before {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 40%;
  height: 100%;
  background: linear-gradient(135deg, transparent 0%, rgba(0, 164, 153, 0.08) 100%);
  pointer-events: none;
}

.welcome-banner::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, var(--color-primary, #00A499), var(--color-accent, #EA7600), var(--color-primary, #00A499));
}

.welcome-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.welcome-title {
  margin: 0;
  font-size: 1.65rem;
  font-weight: 700;
  color: white;
  letter-spacing: -0.3px;
}

.welcome-subtitle {
  margin: 0.3rem 0 0;
  font-size: 0.88rem;
  color: rgba(255, 255, 255, 0.55);
  font-weight: 400;
}

.semester-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.8rem;
  color: white;
  background: rgba(0, 95, 89, 0.85);
  border: 1px solid rgba(0, 107, 100, 0.6);
  padding: 0.4rem 0.9rem;
  border-radius: 20px;
  font-weight: 500;
}

/* ── Main Content ── */
.main-content {
  padding: 2rem 0 3rem;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1.5rem;
}

/* ── Section Headers ── */
.section-header {
  margin-bottom: 1.5rem;
}

.section-header h2 {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0;
  color: var(--color-dark, #394049);
  font-size: 1.1rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.section-header h2 svg {
  color: var(--color-primary, #00A499);
}

/* ── Quick Actions Grid ── */
.quick-actions {
  margin-bottom: 2rem;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 1.5rem;
}

.action-card {
  background: var(--color-surface, white);
  border-radius: var(--radius-lg, 14px);
  border: 1px solid var(--color-border, #dfe4ea);
  overflow: hidden;
  transition: all 0.22s ease;
  box-shadow: var(--shadow-sm, 0 1px 2px rgba(57,64,73,0.06));
}

.action-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-lg, 0 4px 16px rgba(57,64,73,0.10));
}

.card-accent {
  height: 4px;
  background: var(--color-primary, #00A499);
}

.card-accent.accent {
  background: var(--color-accent, #EA7600);
}

.card-accent.plugins {
  background: #006b64;
}

.card-icon-wrap.plugins {
  background: #e6f7f6;
  color: #006b64;
}

.action-btn.plugins-btn {
  background: #006b64;
  color: white;
}

.action-btn.plugins-btn:hover {
  background: #008a81;
  box-shadow: 0 2px 8px rgba(0, 107, 100, 0.3);
}

.card-body {
  padding: 1.75rem 1.75rem 1.5rem;
  display: flex;
  flex-direction: column;
}

.card-icon-wrap {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md, 10px);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
}

.card-icon-wrap.primary {
  background: var(--color-primary-light, #e6f7f6);
  color: var(--color-primary, #00A499);
}

.card-icon-wrap.accent {
  background: var(--color-accent-light, #fff3e6);
  color: var(--color-accent, #EA7600);
}

.action-card h3 {
  margin: 0 0 0.4rem 0;
  color: var(--color-dark, #394049);
  font-size: 1.1rem;
  font-weight: 700;
}

.action-card p {
  margin: 0 0 1.25rem 0;
  color: var(--color-text-secondary, #64707d);
  font-size: 0.88rem;
  line-height: 1.55;
  flex-grow: 1;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.7rem 1rem;
  border: none;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  transition: all 0.18s;
  font-weight: 600;
  font-size: 0.88rem;
}

.action-btn.primary {
  background: var(--color-primary, #006b64);
  color: white;
}

.action-btn.primary:hover {
  background: var(--color-primary-dark, #008a81);
  box-shadow: 0 2px 8px rgba(0, 164, 153, 0.3);
}

.action-btn.accent {
  background: var(--color-accent, #b85000);
  color: white;
}

.action-btn.accent:hover {
  background: var(--color-accent-dark, #c96500);
  box-shadow: 0 2px 8px rgba(234, 118, 0, 0.3);
}

/* ── Impersonate Panel ── */
.impersonate-panel {
  margin-top: 1.5rem;
}

.impersonate-inner {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.25rem;
  background: var(--color-danger-light, #fceced);
  border: 1px solid #f5c6cb;
  border-radius: var(--radius-md, 10px);
  border-left: 4px solid var(--color-danger, #d7373f);
}

.impersonate-icon {
  flex-shrink: 0;
  color: var(--color-danger, #d7373f);
}

.impersonate-text {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.impersonate-text strong {
  font-size: 0.88rem;
  color: var(--color-danger, #d7373f);
}

.impersonate-text span {
  font-size: 0.8rem;
  color: var(--color-text-secondary, #64707d);
}

.impersonate-stop-btn {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.45rem 0.9rem;
  background: var(--color-danger, #d7373f);
  color: white;
  border: none;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  font-size: 0.82rem;
  font-weight: 600;
  transition: background 0.18s;
  white-space: nowrap;
}

.impersonate-stop-btn:hover {
  background: #b82d34;
}

/* ── Modal Styles ── */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(57, 64, 73, 0.45);
  backdrop-filter: blur(3px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--color-surface, white);
  border-radius: var(--radius-lg, 14px);
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: var(--shadow-xl, 0 8px 28px rgba(57,64,73,0.12));
  border: 1px solid var(--color-border, #dfe4ea);
}

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
  font-size: 1.15rem;
  font-weight: 700;
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.4rem;
  cursor: pointer;
  color: var(--color-text-secondary, #64707d);
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm, 6px);
  transition: all 0.18s;
}

.modal-close:hover {
  background: var(--color-background, #f4f6f9);
  color: var(--color-dark, #394049);
}

.modal-body {
  padding: 1.5rem;
}

.modal-body p {
  margin: 0 0 1rem 0;
  color: var(--color-text-secondary, #64707d);
  line-height: 1.55;
  font-size: 0.9rem;
}

.loading-state, .empty-state {
  text-align: center;
  padding: 2rem 0;
}

.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid var(--color-border, #dfe4ea);
  border-top: 3px solid var(--color-primary, #00A499);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.user-selection {
  margin-top: 1rem;
}

.user-select {
  width: 100%;
  padding: 0.7rem 0.85rem;
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-sm, 6px);
  font-size: 0.9rem;
  background: white;
  color: var(--color-text, #2b2f36);
  transition: border-color 0.18s;
}

.user-select:focus {
  outline: none;
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.1);
}

.retry-btn {
  background: var(--color-primary, #006b64);
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  font-size: 0.88rem;
  font-weight: 600;
  transition: background 0.18s;
}

.retry-btn:hover {
  background: var(--color-primary-dark, #008a81);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1.25rem 1.5rem;
  border-top: 1px solid var(--color-border, #dfe4ea);
}

.cancel-btn {
  background: var(--color-background, #f4f6f9);
  color: var(--color-text-secondary, #64707d);
  border: 1px solid var(--color-border, #dfe4ea);
  padding: 0.6rem 1.25rem;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  font-size: 0.88rem;
  font-weight: 500;
  transition: all 0.18s;
}

.cancel-btn:hover {
  background: var(--color-border-light, #eceff3);
  border-color: #c8cdd5;
}

.confirm-btn {
  background: var(--color-primary, #006b64);
  color: white;
  border: none;
  padding: 0.6rem 1.25rem;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  font-size: 0.88rem;
  font-weight: 600;
  transition: all 0.18s;
}

.confirm-btn:hover:not(:disabled) {
  background: var(--color-primary-dark, #008a81);
}

.confirm-btn:disabled {
  background: #b0b8c1;
  cursor: not-allowed;
}

/* ── Configuration Modal ── */
.configuration-modal {
  max-width: 480px;
}

.configuration-body {
  padding: 0 !important;
}

.config-options {
  display: flex;
  flex-direction: column;
}

.config-option {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.5rem;
  border: none;
  background: transparent;
  cursor: pointer;
  text-align: left;
  transition: all 0.15s;
  border-bottom: 1px solid var(--color-border-light, #eceff3);
}

.config-option:last-child {
  border-bottom: none;
}

.config-option:hover {
  background: var(--color-primary-light, #e6f7f6);
}

.option-icon {
  font-size: 1.3rem;
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-background, #f4f6f9);
  border-radius: var(--radius-sm, 6px);
}

.option-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.option-content h4 {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--color-dark, #394049);
}

.option-content p {
  margin: 0;
  font-size: 0.8rem;
  color: var(--color-text-secondary, #64707d);
}

.config-option svg {
  flex-shrink: 0;
  color: var(--color-text-secondary, #64707d);
  opacity: 0.5;
}

/* ── Responsive ── */
@media (max-width: 768px) {
  .actions-grid {
    grid-template-columns: 1fr;
  }

  .main-content {
    padding: 1.5rem 0;
  }

  .container {
    padding: 0 1rem;
  }

  .welcome-title {
    font-size: 1.3rem;
  }

  .welcome-inner {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }
}

@media (max-width: 480px) {
  .card-body {
    padding: 1.25rem;
  }

  .welcome-banner {
    padding: 1.5rem 0;
  }

  .welcome-title {
    font-size: 1.15rem;
  }

  .impersonate-inner {
    flex-direction: column;
    text-align: center;
  }
}

/* ── Plugins Modal ── */
.plugins-modal {
  max-width: 520px;
}

.plugins-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-top: 1rem;
}

.plugin-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.85rem 1rem;
  background: var(--color-background, #f4f6f9);
  border-radius: var(--radius-md, 10px);
  border: 1px solid var(--color-border, #dfe4ea);
  transition: all 0.15s;
}

.plugin-item:hover {
  border-color: #006b64;
  background: #e6f7f6;
  cursor: pointer;
}

.plugin-icon {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  background: white;
  border-radius: var(--radius-sm, 6px);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #006b64;
}

.plugin-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
}

.plugin-info h4 {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--color-dark, #394049);
}

.plugin-info p {
  margin: 0;
  font-size: 0.78rem;
  color: var(--color-text-secondary, #64707d);
}

.plugin-status {
  flex-shrink: 0;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.65rem;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 600;
}

.status-badge.enabled {
  background: #dcfce7;
  color: #166534;
}

.plugin-description {
  font-size: 0.75rem;
  color: #64707d;
  margin-top: 0.25rem;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.plugin-item:hover .plugin-description {
  color: #006b64;
}

.plugin-detail-modal {
  max-width: 700px;
  max-height: 85vh;
}

.plugin-detail-modal .modal-body {
  overflow-y: auto;
  padding: 1.5rem;
}

.detail-section {
  margin-bottom: 1.5rem;
}

.detail-section:last-child {
  margin-bottom: 0;
}

.detail-section h4 {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-primary, #00a499);
  margin: 0 0 0.75rem 0;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.75rem;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.detail-label {
  font-size: 0.75rem;
  color: #64707d;
  font-weight: 500;
}

.detail-value {
  font-size: 0.85rem;
  color: var(--color-dark, #394049);
  font-weight: 500;
}

.description-text {
  font-size: 0.85rem;
  color: var(--color-dark, #394049);
  line-height: 1.5;
  margin: 0;
}

.questions-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-height: 250px;
  overflow-y: auto;
  padding-right: 0.5rem;
}

.question-item {
  display: flex;
  gap: 0.75rem;
  padding: 0.5rem;
  background: var(--color-background, #f4f6f9);
  border-radius: 6px;
  font-size: 0.82rem;
}

.question-id {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-primary, #00a499);
  color: white;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
}

.question-text {
  color: var(--color-dark, #394049);
  line-height: 1.4;
}

.results-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.result-item {
  padding: 0.75rem;
  background: var(--color-background, #f4f6f9);
  border-radius: 8px;
  border-left: 3px solid var(--color-primary, #00a499);
}

.result-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.3rem;
}

.result-name {
  font-weight: 600;
  font-size: 0.88rem;
  color: var(--color-dark, #394049);
}

.result-type-badge {
  display: inline-flex;
  padding: 0.15rem 0.5rem;
  border-radius: 12px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
}

.result-type-badge.estadistico {
  background: #dbeafe;
  color: #1e40af;
}

.result-type-badge.relacional {
  background: #fce7f3;
  color: #9d174d;
}

.result-type-badge.categorico {
  background: #d1fae5;
  color: #065f46;
}

.result-description {
  font-size: 0.8rem;
  color: #64707d;
  margin: 0;
  line-height: 1.4;
}

.status-badge.disabled {
  background: #fee2e2;
  color: #991b1b;
}

.toggle-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.75rem;
  border: none;
  border-radius: 20px;
  cursor: pointer;
  font-size: 0.82rem;
  font-weight: 600;
  transition: all 0.2s;
  position: relative;
}

.toggle-btn.active {
  background: #dcfce7;
  color: #166534;
}

.toggle-btn.inactive {
  background: #fee2e2;
  color: #991b1b;
}

.toggle-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.toggle-btn .toggle-slider {
  width: 32px;
  height: 18px;
  background: currentColor;
  border-radius: 9px;
  position: relative;
  opacity: 0.4;
  transition: opacity 0.2s;
}

.toggle-btn.active .toggle-slider {
  opacity: 1;
}

.toggle-btn .toggle-slider::after {
  content: '';
  position: absolute;
  width: 14px;
  height: 14px;
  background: white;
  border-radius: 50%;
  top: 2px;
  left: 2px;
  transition: transform 0.2s;
}

.toggle-btn.active .toggle-slider::after {
  transform: translateX(14px);
}

.add-plugin-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.7rem 1rem;
  background: #006b64;
  color: white;
  border: none;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  font-weight: 600;
  font-size: 0.88rem;
  transition: all 0.18s;
  margin-bottom: 1rem;
}

.add-plugin-btn:hover {
  background: #008a81;
  box-shadow: 0 2px 8px rgba(0, 107, 100, 0.3);
}
</style>
