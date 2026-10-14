<template>
  <div class="invitation-codes-panel">
    <div class="panel-header">
      <h2>Códigos de Invitación</h2>
      <div class="header-actions">
        <button v-if="isAdmin" class="create-btn" @click="$emit('create')">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          Crear
        </button>
        <button class="close-btn" @click="$emit('close')">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
    </div>

    <div class="panel-body">
      <div v-if="loading" class="loading-state">
        <div class="spinner"></div>
        <p>Cargando códigos...</p>
      </div>

      <div v-else-if="error" class="error-state">
        <p>{{ error }}</p>
        <button class="retry-btn" @click="loadCodes">Reintentar</button>
      </div>

      <div v-else-if="codesList.length === 0" class="empty-state">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M15 7h3a5 5 0 0 1 5 5 5 5 0 0 1-5 5h-3m-6 0H6a5 5 0 0 1-5-5 5 5 0 0 1 5-5h3"/>
          <line x1="8" y1="12" x2="16" y2="12"/>
        </svg>
        <p>No hay códigos de invitación activos</p>
        <span v-if="isAdmin">Usa el botón "Crear" para generar uno nuevo</span>
      </div>

      <div v-else class="codes-list">
        <div v-for="code in codesList" :key="code.id" class="code-card" :class="{ expanded: expandedCodeId === code.id }">
          <div class="code-header" @click="toggleCode(code.id)">
            <div class="code-info">
              <div class="code-badge">
                <span class="code-value">{{ code.code }}</span>
                <button class="copy-btn" @click.stop="copyCode(code.code)" title="Copiar código">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
                  </svg>
                </button>
              </div>
              <span class="status-badge" :class="getStatusClass(code)">{{ getStatusText(code) }}</span>
            </div>
            <svg class="expand-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
          </div>

          <div class="code-body">
            <div class="code-meta">
              <div class="meta-item">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <polyline points="12 6 12 12 16 14"/>
                </svg>
                <span>Expira: {{ formatDate(code.expires_at) }}</span>
                <span v-if="getTimeRemaining(code.expires_at)" class="countdown">
                  ({{ getTimeRemaining(code.expires_at) }})
                </span>
              </div>
              <div class="meta-item">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                  <circle cx="9" cy="7" r="4"/>
                  <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
                  <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                </svg>
                <span>Usado: {{ code.used_count || 0 }}{{ code.max_uses ? `/${code.max_uses}` : ' veces' }}</span>
              </div>
            </div>

            <div class="code-actions">
              <button v-if="isAdmin" class="action-btn revoke-btn" @click.stop="openRevokeModal(code)" :disabled="revokingCodeId === code.id">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <line x1="15" y1="9" x2="9" y2="15"/>
                  <line x1="9" y1="9" x2="15" y2="15"/>
                </svg>
                {{ revokingCodeId === code.id ? 'Revocando...' : 'Revocar' }}
              </button>
            </div>

            <div v-if="expandedCodeId === code.id" class="students-section">
              <h4>Estudiantes unidos:</h4>
              <div v-if="loadingStudents" class="loading-small">
                <div class="spinner-small"></div>
              </div>
              <div v-else-if="code.students && code.students.length > 0" class="students-list">
                <div v-for="student in code.students" :key="student.id" class="student-item">
                  <div class="student-avatar">
                    <span>{{ getInitials(student.full_name || student.nombres) }}</span>
                  </div>
                  <div class="student-info">
                    <span class="student-name">{{ student.full_name || `${student.nombres} ${student.paterno || ''} ${student.materno || ''}`.trim() }}</span>
                    <span class="student-email">{{ student.email }}</span>
                  </div>
                </div>
              </div>
              <p v-else class="no-students">Aún no hay estudiantes que hayan usado este código</p>
            </div>
            <button v-else class="expand-btn" @click.stop="loadAndExpandCode(code)">
              Ver estudiantes que usaron el código
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Modal de confirmación de revocación -->
  <div v-if="showRevokeModal" class="revoke-modal-overlay" @click.self="closeRevokeModal">
    <div class="revoke-modal">
      <div class="revoke-modal-header">
        <h3>Revocar código de invitación</h3>
        <button class="close-btn" @click="closeRevokeModal">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
      <div class="revoke-modal-body">
        <p>¿Estás seguro de que deseas revocar el código <strong>{{ codeToRevoke?.code }}</strong>?</p>
        <p class="warning-text">Los estudiantes que ya lo usaron no se verán afectados.</p>
        <p v-if="revokeError" class="error-text">{{ revokeError }}</p>
      </div>
      <div class="revoke-modal-footer">
        <button class="btn btn-cancel" @click="closeRevokeModal">Cancelar</button>
        <button class="btn btn-revoke" @click="revokeCode" :disabled="revokingCodeId !== null">
          {{ revokingCodeId !== null ? 'Revocando...' : 'Revocar' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'

const props = defineProps({
  sectionId: {
    type: [Number, String],
    required: true
  },
  isAdmin: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['close', 'create'])

const API_BASE_URL = 'http://127.0.0.1:3000/api/v1'

const codes = ref([])
const codesList = computed(() => codes.value || [])
const loading = ref(false)
const error = ref(null)
const expandedCodeId = ref(null)
const loadingStudents = ref(false)
const revokingCodeId = ref(null)
const showRevokeModal = ref(false)
const codeToRevoke = ref(null)
const revokeError = ref(null)

const loadCodes = async () => {
  loading.value = true
  error.value = null
  codes.value = []
  
  try {
    const token = localStorage.getItem('auth_token')
    const sectionId = Number(props.sectionId)
    
    const response = await fetch(`${API_BASE_URL}/invitation-codes/section/${sectionId}`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })

    if (!response.ok) {
      throw new Error('Error al cargar los códigos de invitación')
    }

    const result = await response.json()
    
    if (!result) {
      codes.value = []
    } else if (Array.isArray(result)) {
      codes.value = result
    } else {
      codes.value = result.data || result.codes || []
    }
  } catch (err) {
    error.value = err.message || 'Error al cargar los códigos'
    console.error('Error loading codes:', err)
  } finally {
    loading.value = false
  }
}

const loadAndExpandCode = async (code) => {
  if (expandedCodeId.value === code.id) {
    expandedCodeId.value = null
    return
  }
  expandedCodeId.value = code.id

  if (code.students) return

  loadingStudents.value = true
  try {
    const token = localStorage.getItem('auth_token')
    const response = await fetch(`${API_BASE_URL}/invitation-codes/${code.id}/students`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })

    if (response.ok) {
      const result = await response.json()
      code.students = result.data || result.students || []
    }
  } catch (err) {
    console.error('Error loading students:', err)
  } finally {
    loadingStudents.value = false
  }
}

const toggleCode = (codeId) => {
  if (expandedCodeId.value === codeId) {
    expandedCodeId.value = null
  } else {
    loadAndExpandCode({ id: codeId })
  }
}

const revokeCode = async () => {
  if (!codeToRevoke.value) return

  revokeError.value = null
  revokingCodeId.value = codeToRevoke.value.id
  
  try {
    const token = localStorage.getItem('auth_token')
    const response = await fetch(`${API_BASE_URL}/invitation-codes/${codeToRevoke.value.id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })

    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.message || 'Error al revocar el código')
    }

    showRevokeModal.value = false
    codeToRevoke.value = null
    await loadCodes()
  } catch (err) {
    revokeError.value = err.message || 'Error al revocar el código'
  } finally {
    revokingCodeId.value = null
  }
}

const openRevokeModal = (code) => {
  codeToRevoke.value = code
  revokeError.value = null
  showRevokeModal.value = true
}

const closeRevokeModal = () => {
  showRevokeModal.value = false
  codeToRevoke.value = null
  revokeError.value = null
}

const copyCode = async (codeValue) => {
  try {
    await navigator.clipboard.writeText(codeValue)
  } catch (err) {
    const textArea = document.createElement('textarea')
    textArea.value = codeValue
    document.body.appendChild(textArea)
    textArea.select()
    document.execCommand('copy')
    document.body.removeChild(textArea)
  }
}

const formatDate = (dateString) => {
  if (!dateString) return 'Sin fecha'
  const date = new Date(dateString)
  return date.toLocaleDateString('es-CL', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getStatusClass = (code) => {
  if (code.revoked || code.used_count >= code.max_uses) return 'status-expired'
  if (new Date(code.expires_at) < new Date()) return 'status-expired'
  return 'status-active'
}

const getStatusText = (code) => {
  if (code.revoked) return 'Revocado'
  if (code.used_count >= code.max_uses) return 'Agotado'
  if (new Date(code.expires_at) < new Date()) return 'Expirado'
  return 'Activo'
}

const getTimeRemaining = (expiresAt) => {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires - now

  if (diff <= 0) return null

  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))

  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h`
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
  return `${minutes}m`
}

const getInitials = (name) => {
  if (!name) return '?'
  return name.split(' ').map(w => w[0]).join('').toUpperCase().substring(0, 2)
}

onMounted(() => {
  loadCodes()
})

watch(() => props.sectionId, () => {
  loadCodes()
})

defineExpose({
  loadCodes
})
</script>

<style scoped>
.invitation-codes-panel {
  background: var(--color-surface, white);
  border-radius: var(--radius-lg, 14px);
  border: 1px solid var(--color-border, #dfe4ea);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  max-height: 90vh;
  width: 90vw;
  max-width: 600px;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  background: linear-gradient(135deg, #394049 0%, #2a3038 100%);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.panel-header h2 {
  margin: 0;
  color: white;
  font-size: 1.1rem;
  font-weight: 600;
}

.create-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 1rem;
  background: var(--color-primary, #00A499);
  color: white;
  border: none;
  border-radius: var(--radius-sm, 6px);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.18s;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.4rem;
  background: rgba(255, 255, 255, 0.1);
  border: none;
  border-radius: 6px;
  color: white;
  cursor: pointer;
  transition: all 0.15s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.create-btn:hover {
  background: #008a81;
  box-shadow: 0 2px 8px rgba(0, 164, 153, 0.3);
}

.panel-body {
  padding: 1rem 1.5rem 1.5rem;
  flex: 1;
  overflow-y: auto;
  max-height: calc(90vh - 70px);
}

.loading-state,
.error-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2.5rem 1rem;
  text-align: center;
}

.empty-state svg {
  color: #9ca3af;
  margin-bottom: 1rem;
}

.empty-state p {
  margin: 0 0 0.5rem;
  color: var(--color-text-secondary, #64707d);
  font-weight: 500;
}

.empty-state span {
  font-size: 0.85rem;
  color: #9ca3af;
}

.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid var(--color-border, #dfe4ea);
  border-top-color: var(--color-primary, #00A499);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 1rem;
}

.spinner-small {
  width: 20px;
  height: 20px;
  border: 2px solid var(--color-border, #dfe4ea);
  border-top-color: var(--color-primary, #00A499);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-state p {
  color: var(--color-danger, #d7373f);
  margin: 0 0 1rem;
}

.retry-btn {
  padding: 0.5rem 1rem;
  background: var(--color-primary, #006b64);
  color: white;
  border: none;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 500;
}

.codes-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.code-card {
  background: var(--color-background, #f4f6f9);
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-md, 10px);
  overflow: hidden;
  transition: all 0.2s;
}

.code-card:hover {
  border-color: var(--color-primary, #00A499);
}

.code-card.expanded {
  border-color: var(--color-primary, #00A499);
  box-shadow: 0 2px 12px rgba(0, 164, 153, 0.1);
}

.code-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  cursor: pointer;
  background: white;
}

.code-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.code-badge {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--color-background, #f4f6f9);
  padding: 0.4rem 0.75rem;
  border-radius: 6px;
  border: 1px solid var(--color-border, #dfe4ea);
}

.code-value {
  font-family: 'SF Mono', Monaco, monospace;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--color-dark, #394049);
  letter-spacing: 1px;
}

.copy-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem;
  background: none;
  border: none;
  color: #9ca3af;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.15s;
}

.copy-btn:hover {
  background: var(--color-border, #dfe4ea);
  color: var(--color-primary, #00A499);
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.65rem;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 600;
}

.status-badge.status-active {
  background: #dcfce7;
  color: #166534;
}

.status-badge.status-expired {
  background: #fee2e2;
  color: #991b1b;
}

.expand-icon {
  color: #9ca3af;
  transition: transform 0.2s;
}

.expanded .expand-icon {
  transform: rotate(180deg);
}

.code-body {
  padding: 0 1rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.code-meta {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  padding: 0.75rem;
  background: var(--color-background, #f4f6f9);
  border-radius: 6px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  color: var(--color-text-secondary, #64707d);
}

.meta-item svg {
  flex-shrink: 0;
  color: #9ca3af;
}

.countdown {
  color: var(--color-primary, #00A499);
  font-weight: 500;
}

.code-actions {
  display: flex;
  gap: 0.5rem;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.45rem 0.85rem;
  border: none;
  border-radius: var(--radius-sm, 6px);
  font-size: 0.82rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.revoke-btn {
  background: #fee2e2;
  color: #991b1b;
}

.revoke-btn:hover:not(:disabled) {
  background: #fecaca;
}

.students-section {
  margin-top: 0.5rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--color-border, #dfe4ea);
}

.students-section h4 {
  margin: 0 0 0.75rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-dark, #394049);
}

.students-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.student-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  background: white;
  border-radius: 6px;
  border: 1px solid var(--color-border, #dfe4ea);
}

.student-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--color-primary, #00A499);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.7rem;
  font-weight: 600;
  flex-shrink: 0;
}

.student-info {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  min-width: 0;
}

.student-name {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--color-dark, #394049);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.student-email {
  font-size: 0.75rem;
  color: #9ca3af;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.no-students {
  font-size: 0.85rem;
  color: #9ca3af;
  margin: 0;
  font-style: italic;
}

.expand-btn {
  background: none;
  border: none;
  color: var(--color-primary, #00A499);
  font-size: 0.82rem;
  font-weight: 500;
  cursor: pointer;
  padding: 0.25rem 0;
  text-align: left;
}

.expand-btn:hover {
  text-decoration: underline;
}

.loading-small {
  display: flex;
  justify-content: center;
  padding: 0.5rem;
}

/* Revoke confirmation modal */
.revoke-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1200;
}

.revoke-modal {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 25px rgba(0, 0, 0, 0.15);
  width: 90vw;
  max-width: 420px;
  overflow: hidden;
}

.revoke-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.revoke-modal-header h3 {
  margin: 0;
  color: var(--color-dark, #394049);
  font-size: 1.1rem;
  font-weight: 600;
}

.revoke-modal-header .close-btn {
  background: none;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 4px;
  display: flex;
  align-items: center;
  transition: all 0.15s;
}

.revoke-modal-header .close-btn:hover {
  background: var(--color-background, #f4f6f9);
  color: var(--color-dark, #394049);
}

.revoke-modal-body {
  padding: 1.5rem;
}

.revoke-modal-body p {
  margin: 0 0 0.75rem;
  color: var(--color-text-secondary, #64748b);
  line-height: 1.5;
}

.revoke-modal-body .warning-text {
  color: #9ca3af;
  font-size: 0.9rem;
}

.revoke-modal-body .error-text {
  color: var(--color-danger, #d7373f);
  background: #fef2f2;
  padding: 0.75rem;
  border-radius: 6px;
  margin-top: 1rem;
  margin-bottom: 0;
  font-size: 0.9rem;
}

.revoke-modal-footer {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--color-border, #dfe4ea);
  background: var(--color-background, #f4f6f9);
}

.btn {
  padding: 0.6rem 1.25rem;
  border: none;
  border-radius: 6px;
  font-weight: 500;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-cancel {
  background: white;
  color: var(--color-text-secondary, #64748b);
  border: 1px solid var(--color-border, #dfe4ea);
}

.btn-cancel:hover:not(:disabled) {
  background: var(--color-background, #f4f6f9);
}

.btn-revoke {
  background: var(--color-danger, #d7373f);
  color: white;
}

.btn-revoke:hover:not(:disabled) {
  background: #b82e35;
}

@media (max-width: 640px) {
  .panel-body {
    padding: 0.75rem 1rem 1rem;
  }

  .code-header {
    padding: 0.75rem;
  }

  .code-body {
    padding: 0 0.75rem 0.75rem;
  }
}
</style>
