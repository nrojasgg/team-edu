<template>
  <div class="configuracion-page">
    <!-- Header con información del usuario -->
    <TopBar :user="user" @logout="handleLogout" />

    <!-- Contenido principal -->
    <main class="main-content">
      <div class="container">
        <div class="page-header">
          <h1>Configuración del Sistema</h1>
          <p>Gestiona roles y permisos de usuarios</p>
        </div>

        <!-- Panel de roles -->
        <section class="roles-section">
          <div class="section-header">
            <h2>Roles del Sistema</h2>
            <button class="add-role-btn" @click="showAddRoleModal = true">
              <span>+</span> Agregar Rol
            </button>
          </div>

          <div v-if="loadingRoles" class="loading-state">
            <div class="spinner"></div>
            <p>Cargando roles...</p>
          </div>

          <div v-else-if="roles.length === 0" class="empty-state">
            <p>No hay roles configurados en el sistema.</p>
            <button class="retry-btn" @click="loadRoles">Cargar Roles</button>
          </div>

          <div v-else class="roles-grid">
            <div
              v-for="role in roles"
              :key="role.id"
              class="role-card"
              :class="{ active: selectedRole?.id === role.id }"
              @click="selectRole(role)"
            >
              <div class="role-header">
                <h3>{{ role.name }}</h3>
                <div class="role-actions">
                  <button class="edit-btn" @click.stop="editRole(role)">✏️</button>
                  <button class="delete-btn" @click.stop="deleteRole(role)">🗑️</button>
                </div>
              </div>
              <p>{{ getRoleDescription(role.name) }}</p>
              <div class="role-stats">
                <span class="permission-count">{{ role.permission_count || 0 }} permisos</span>
              </div>
            </div>
          </div>
        </section>

        <!-- Panel de permisos del rol seleccionado -->
        <section v-if="selectedRole" class="permissions-section">
          <div class="section-header">
            <h2>Permisos de {{ selectedRole.name }}</h2>
            <div class="header-actions">
              <span v-if="saveSuccessMessage" class="success-message">{{ saveSuccessMessage }}</span>
              <button class="save-btn" @click="savePermissions" :disabled="!hasChanges || savingPermissions">
                {{ savingPermissions ? 'Guardando...' : 'Guardar Cambios' }}
              </button>
            </div>
          </div>

          <div v-if="loadingPermissions" class="loading-state">
            <div class="spinner"></div>
            <p>Cargando permisos...</p>
          </div>

          <div v-else class="permissions-grid">
            <div
              v-for="permission in allPermissions"
              :key="permission.id"
              class="permission-item"
            >
              <label class="permission-checkbox">
                <input
                  type="checkbox"
                  :checked="isPermissionAssigned(permission.id)"
                  @change="togglePermission(permission.id)"
                />
                <span class="checkmark"></span>
                <div class="permission-info">
                  <strong>{{ permission.name }}</strong>
                  <p>{{ getPermissionDescription(permission.name) }}</p>
                </div>
              </label>
            </div>
          </div>
        </section>
      </div>
    </main>

    <!-- Modal para agregar/editar rol -->
    <div v-if="showAddRoleModal || showEditRoleModal" class="modal-overlay" @click="closeRoleModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>{{ showEditRoleModal ? 'Editar Rol' : 'Agregar Rol' }}</h3>
          <button class="modal-close" @click="closeRoleModal">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="saveRole">
            <div class="form-group">
              <label for="roleName">Nombre del Rol</label>
              <input
                id="roleName"
                v-model="roleForm.name"
                type="text"
                required
                placeholder="Ej: profesor, estudiante, admin"
              />
            </div>
          </form>
        </div>
        <div class="modal-footer">
          <button class="cancel-btn" @click="closeRoleModal">Cancelar</button>
          <button class="confirm-btn" @click="saveRole" :disabled="!roleForm.name.trim()">
            {{ showEditRoleModal ? 'Actualizar' : 'Crear' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal de confirmación para eliminar rol -->
    <ConfirmModal
      :show="showConfirmModal"
      title="Eliminar Rol"
      :message="confirmMessage"
      confirm-text="Eliminar"
      confirm-class="btn-danger"
      @confirm="handleConfirmDeleteRole"
      @cancel="showConfirmModal = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import ConfirmModal from '~/components/ConfirmModal.vue'

// Aplicar middleware de autenticación
definePageMeta({
  middleware: 'auth'
})

const { getUser, logout, getToken } = useAuth()

// Estado del usuario
const user = ref(null)

// Estado de roles
const roles = ref([])
const selectedRole = ref(null)
const loadingRoles = ref(false)

// Estado de permisos
const allPermissions = ref([])
const rolePermissions = ref([])
const loadingPermissions = ref(false)

// Estado de cambios
const originalPermissions = ref([])
const hasChanges = computed(() => {
  if (!selectedRole.value) return false
  const current = rolePermissions.value.map(p => p.id).sort()
  const original = originalPermissions.value.sort()
  return JSON.stringify(current) !== JSON.stringify(original)
})

// Estado de guardado
const savingPermissions = ref(false)
const saveSuccessMessage = ref('')

// Estado de modales
const showAddRoleModal = ref(false)
const showEditRoleModal = ref(false)
const roleForm = ref({ name: '' })
const editingRole = ref(null)

const API_BASE_URL = 'http://127.0.0.1:3000/api/v1'

// Verificar si es admin
const isAdmin = computed(() => user.value?.role_name === 'admin')

// Cargar datos iniciales
onMounted(async () => {
  user.value = getUser()

  if (!user.value || !isAdmin.value) {
    navigateTo('/home')
    return
  }

  await loadRoles()
  await loadAllPermissions()
})

// Funciones de navegación
const handleLogout = () => {
  logout()
}

// Funciones de roles
const loadRoles = async () => {
  loadingRoles.value = true
  try {
    const token = getToken()
    const response = await fetch(`${API_BASE_URL}/roles`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })

    if (response.ok) {
      const result = await response.json()
      roles.value = result
    } else {
      console.error('Error loading roles:', response.statusText)
    }
  } catch (error) {
    console.error('Error loading roles:', error)
  } finally {
    loadingRoles.value = false
  }
}

const selectRole = async (role) => {
  selectedRole.value = role
  await loadRolePermissions(role.id)
}

const loadRolePermissions = async (roleId) => {
  loadingPermissions.value = true
  try {
    const token = getToken()
    const response = await fetch(`${API_BASE_URL}/roles/${roleId}/permissions`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })

    if (response.ok) {
      const result = await response.json()
      rolePermissions.value = result
      originalPermissions.value = result.map(p => p.id)
    } else {
      console.error('Error loading role permissions:', response.statusText)
      rolePermissions.value = []
      originalPermissions.value = []
    }
  } catch (error) {
    console.error('Error loading role permissions:', error)
    rolePermissions.value = []
    originalPermissions.value = []
  } finally {
    loadingPermissions.value = false
  }
}

const loadAllPermissions = async () => {
  try {
    const token = getToken()
    const response = await fetch(`${API_BASE_URL}/permissions`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })

    if (response.ok) {
      const result = await response.json()
      allPermissions.value = result
    } else {
      console.error('Error loading permissions:', response.statusText)
    }
  } catch (error) {
    console.error('Error loading permissions:', error)
  }
}

const isPermissionAssigned = (permissionId) => {
  return rolePermissions.value.some(p => p.id === permissionId)
}

const togglePermission = (permissionId) => {
  const index = rolePermissions.value.findIndex(p => p.id === permissionId)
  if (index > -1) {
    rolePermissions.value.splice(index, 1)
  } else {
    const permission = allPermissions.value.find(p => p.id === permissionId)
    if (permission) {
      rolePermissions.value.push(permission)
    }
  }
}

const savePermissions = async () => {
  if (!selectedRole.value || savingPermissions.value) return

  savingPermissions.value = true
  saveSuccessMessage.value = ''

  try {
    const token = getToken()
    const currentPermissionIds = rolePermissions.value.map(p => p.id)

    // Remove permissions that are no longer assigned
    const permissionsToRemove = originalPermissions.value.filter(id => !currentPermissionIds.includes(id))
    for (const permId of permissionsToRemove) {
      await fetch(`${API_BASE_URL}/roles/${selectedRole.value.id}/permissions/${permId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
    }

    // Add new permissions
    const permissionsToAdd = currentPermissionIds.filter(id => !originalPermissions.value.includes(id))
    for (const permId of permissionsToAdd) {
      await fetch(`${API_BASE_URL}/roles/${selectedRole.value.id}/permissions`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ permission_id: permId })
      })
    }

    // Update original permissions
    originalPermissions.value = [...currentPermissionIds]

    // Update the permission count in the selected role and roles array
    const newCount = currentPermissionIds.length
    selectedRole.value.permission_count = newCount

    // Update the role in the roles array as well
    const roleIndex = roles.value.findIndex(r => r.id === selectedRole.value.id)
    if (roleIndex > -1) {
      roles.value[roleIndex].permission_count = newCount
    }

    // Show success message temporarily
    saveSuccessMessage.value = 'Permisos actualizados correctamente'
    setTimeout(() => {
      saveSuccessMessage.value = ''
    }, 3000)

  } catch (error) {
    console.error('Error saving permissions:', error)
    // You could show an error message here
  } finally {
    savingPermissions.value = false
  }
}

// Funciones de gestión de roles
const editRole = (role) => {
  editingRole.value = role
  roleForm.value = { name: role.name }
  showEditRoleModal.value = true
}

// Confirm modal state
const showConfirmModal = ref(false)
const confirmMessage = ref('')
const roleToDelete = ref(null)

const deleteRole = (role) => {
  roleToDelete.value = role
  confirmMessage.value = `¿Estás seguro de que quieres eliminar el rol "${role.name}"?`
  showConfirmModal.value = true
}

const handleConfirmDeleteRole = async () => {
  if (!roleToDelete.value) return
  const role = roleToDelete.value

  try {
    const token = getToken()
    const response = await fetch(`${API_BASE_URL}/roles/${role.id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })

    if (response.ok) {
      roles.value = roles.value.filter(r => r.id !== role.id)
      if (selectedRole.value?.id === role.id) {
        selectedRole.value = null
        rolePermissions.value = []
      }
      console.log('Rol eliminado correctamente')
    } else {
      console.error('Error deleting role')
    }
  } catch (error) {
    console.error('Error deleting role:', error)
  } finally {
    roleToDelete.value = null
    showConfirmModal.value = false
  }
}

const isEditingRole = () => showEditRoleModal.value && editingRole.value

const fetchSaveRole = (token) => {
  const body = JSON.stringify({ name: roleForm.value.name })
  const headers = { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' }
  if (isEditingRole()) {
    return fetch(`${API_BASE_URL}/roles/${editingRole.value.id}`, { method: 'PUT', headers, body })
  }
  return fetch(`${API_BASE_URL}/roles`, { method: 'POST', headers, body })
}

const applyRoleSaveResult = (result) => {
  if (isEditingRole()) {
    const index = roles.value.findIndex(r => r.id === editingRole.value.id)
    if (index > -1) roles.value[index] = result
  } else {
    result.permission_count = 0
    roles.value.push(result)
  }
  closeRoleModal()
}

const saveRole = async () => {
  if (!roleForm.value.name.trim()) return

  try {
    const token = getToken()
    const response = await fetchSaveRole(token)

    if (response.ok) {
      applyRoleSaveResult(await response.json())
    } else {
      console.error('Error saving role')
    }
  } catch (error) {
    console.error('Error saving role:', error)
  }
}

const closeRoleModal = () => {
  showAddRoleModal.value = false
  showEditRoleModal.value = false
  roleForm.value = { name: '' }
  editingRole.value = null
}

// Funciones auxiliares
const getRoleDescription = (roleName) => {
  const descriptions = {
    admin: 'Acceso completo al sistema',
    profesor: 'Puede gestionar cursos y ver información de estudiantes',
    estudiante: 'Acceso limitado a cursos y evaluaciones propias'
  }
  return descriptions[roleName] || 'Rol personalizado'
}

const getPermissionDescription = (permissionName) => {
  const descriptions = {
    read_users: 'Ver información de usuarios',
    write_users: 'Crear y modificar usuarios',
    delete_users: 'Eliminar usuarios',
    read_courses: 'Ver información de cursos',
    write_courses: 'Crear y modificar cursos',
    delete_courses: 'Eliminar cursos'
  }
  return descriptions[permissionName] || 'Permiso personalizado'
}

// Meta información
useHead({
  title: 'TeamEdu - Configuración',
  meta: [
    { name: 'description', content: 'Configuración del sistema TeamEdu - Gestión de roles y permisos' }
  ]
})
</script>

<style scoped>
.configuracion-page {
  min-height: 100vh;
  background: var(--color-background, #f4f6f9);
}

.main-content {
  padding: 2rem 0;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1rem;
}

.page-header {
  margin-bottom: 3rem;
  text-align: center;
}

.page-header h1 {
  color: var(--color-dark, #394049);
  font-size: 2.5rem;
  margin-bottom: 0.5rem;
}

.page-header p {
  color: #94a3b8;
  font-size: 1.1rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.section-header h2 {
  color: var(--color-dark, #394049);
  font-size: 1.5rem;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.success-message {
  color: #48bb78;
  font-weight: 500;
  font-size: 0.9rem;
  opacity: 0;
  animation: fadeIn 0.3s ease-in-out forwards;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

.add-role-btn {
  background: #276749;
  color: white;
  border: none;
  border-radius: 6px;
  padding: 0.5rem 1rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9rem;
  transition: background 0.2s;
}

.add-role-btn:hover {
  background: var(--color-primary, #00A499);
}

.save-btn {
  background: var(--color-primary, #006b64);
  color: white;
  border: none;
  border-radius: 6px;
  padding: 0.5rem 1rem;
  cursor: pointer;
  font-size: 0.9rem;
  transition: background 0.2s;
}

.save-btn:hover:not(:disabled) {
  background: var(--color-primary-dark, #008a7f);
}

.save-btn:disabled {
  background: #cbd5e0;
  cursor: not-allowed;
}

.roles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
  margin-bottom: 3rem;
}

.role-card {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  border: 2px solid var(--color-border, #dfe4ea);
  cursor: pointer;
  transition: all 0.2s;
}

.role-card:hover {
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.role-card.active {
  border-color: var(--color-primary, #00A499);
  background: #ebf8ff;
}

.role-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.role-header h3 {
  color: var(--color-dark, #394049);
  font-size: 1.2rem;
  margin: 0;
}

.role-actions {
  display: flex;
  gap: 0.5rem;
}

.edit-btn, .delete-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 4px;
  transition: background 0.2s;
}

.edit-btn:hover {
  background: var(--color-background, #f4f6f9);
}

.delete-btn:hover {
  background: #fed7d7;
}

.role-card p {
  color: #94a3b8;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

.role-stats {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.permission-count {
  background: var(--color-border, #dfe4ea);
  color: var(--color-text-secondary, #64748b);
  padding: 0.25rem 0.5rem;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
}

.permissions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 1rem;
}

.permission-item {
  background: white;
  border-radius: 8px;
  border: 1px solid var(--color-border, #dfe4ea);
  transition: border-color 0.2s;
}

.permission-item:hover {
  border-color: var(--color-primary, #00A499);
}

.permission-checkbox {
  display: flex;
  align-items: flex-start;
  padding: 1rem;
  cursor: pointer;
  width: 100%;
}

.permission-checkbox input {
  margin-right: 1rem;
  margin-top: 0.25rem;
}

.permission-info strong {
  color: var(--color-dark, #394049);
  display: block;
  margin-bottom: 0.25rem;
}

.permission-info p {
  color: #94a3b8;
  font-size: 0.9rem;
  margin: 0;
}

.loading-state, .empty-state {
  text-align: center;
  padding: 3rem;
  color: #94a3b8;
}

.spinner {
  border: 3px solid var(--color-border, #dfe4ea);
  border-top: 3px solid var(--color-primary, #00A499);
  border-radius: 50%;
  width: 30px;
  height: 30px;
  animation: spin 1s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.retry-btn {
  background: var(--color-primary, #006b64);
  color: white;
  border: none;
  border-radius: 6px;
  padding: 0.5rem 1rem;
  cursor: pointer;
  margin-top: 1rem;
}

.retry-btn:hover {
  background: var(--color-primary-dark, #008a7f);
}

/* Modal styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  padding: 1.5rem;
  border-bottom: 1px solid var(--color-border, #dfe4ea);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 {
  margin: 0;
  color: var(--color-dark, #394049);
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #94a3b8;
}

.modal-body {
  padding: 1.5rem;
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
  border-radius: 6px;
  font-size: 1rem;
}

.form-group input:focus {
  outline: none;
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.1);
}

.modal-footer {
  padding: 1.5rem;
  border-top: 1px solid var(--color-border, #dfe4ea);
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
}

.cancel-btn {
  background: var(--color-border, #dfe4ea);
  color: var(--color-text-secondary, #64748b);
  border: none;
  border-radius: 6px;
  padding: 0.5rem 1rem;
  cursor: pointer;
  transition: background 0.2s;
}

.cancel-btn:hover {
  background: #cbd5e0;
}

.confirm-btn {
  background: var(--color-primary, #006b64);
  color: white;
  border: none;
  border-radius: 6px;
  padding: 0.5rem 1rem;
  cursor: pointer;
  transition: background 0.2s;
}

.confirm-btn:hover:not(:disabled) {
  background: var(--color-primary-dark, #008a7f);
}

.confirm-btn:disabled {
  background: #cbd5e0;
  cursor: not-allowed;
}
</style>