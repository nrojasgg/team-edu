<template>
  <div class="surveys-tab">
    <div class="surveys-header">
      <h3>Evaluaciones</h3>
      <div v-if="userRole !== 'estudiante'" class="header-actions">
        <button @click="$emit('create-peer-evaluation')" class="create-peer-evaluation-btn">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
            <circle cx="9" cy="7" r="4"/>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
          </svg>
          Nueva evaluación entre pares
        </button>
      </div>
    </div>

    <div v-if="loadingSurveys" class="loading-surveys">
      Cargando evaluaciones...
    </div>

    <div v-else-if="surveys.length === 0" class="no-surveys">
      <p>No hay evaluaciones creadas para esta coordinación </p>
    </div>

    <div v-else class="surveys-container">
      <div 
        v-for="survey in surveys" 
        :key="survey.id"
        class="survey-card"
        :class="{ 
          'survey-active': survey.status === 'active',
          'survey-expired': survey.status === 'expired',
          'survey-draft': survey.status === 'draft'
        }"
      >
        <div class="survey-header">
          <h4 class="survey-title">{{ survey.name }}</h4>
          <span class="survey-plugin">{{ survey.type }}</span>
        </div>

        <div class="survey-info">
          <p class="survey-description">{{ survey.description || 'Sin descripción' }}</p>
          
          <div class="survey-dates">
            <div class="date-info">
              <strong>Creada:</strong> {{ formatDate(survey.createdAt) }}
            </div>
            <div class="date-info">
              <strong>Expira:</strong> {{ formatDate(survey.expiresAt) }}
            </div>
          </div>
        </div>

        <div v-if="userRole !== 'estudiante'" class="survey-actions">
          <button 
            @click="$emit('edit-survey', survey)" 
            class="action-btn edit-btn" 
            title="Editar encuesta"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
            </svg>
            <span class="btn-text">Editar</span>
          </button>
          <button 
            @click="$emit('view-responses', survey)" 
            class="action-btn view-btn" 
            title="Ver respuestas"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
              <circle cx="12" cy="12" r="3"/>
            </svg>
            <span class="btn-text">Respuestas</span>
          </button>
          <button 
            @click="$emit('delete-survey', survey)" 
            class="action-btn delete-btn" 
            title="Eliminar encuesta"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="3,6 5,6 21,6"/>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2 2h4a2 2 0 0 1 2 2v2"/>
              <path d="M10 11V17"/>
              <path d="M14 11V17"/>
            </svg>
            <span class="btn-text">Eliminar</span>
          </button>
        </div>
        
        <div v-else-if="userRole === 'estudiante'" class="survey-actions">
          <button 
            v-if="survey.status === 'active'"
            @click="$emit('respond-survey', survey)" 
            class="action-btn view-btn" 
            title="Responder evaluación"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
              <circle cx="12" cy="12" r="3"/>
            </svg>
            <span class="btn-text">Responder</span>
          </button>
          <button 
            v-else
            @click="$emit('view-results', survey)" 
            class="action-btn results-btn" 
            title="Ver resultados"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
            </svg>
            <span class="btn-text">Ver Resultados</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useDatetimeInput } from '~/composables/useDatetimeInput'

const props = defineProps({
  surveys: {
    type: Array,
    default: () => []
  },
  loadingSurveys: {
    type: Boolean,
    default: false
  },
  userRole: {
    type: String,
    default: ''
  }
})

const emit = defineEmits([
  'create-peer-evaluation',
  'edit-survey',
  'view-responses',
  'delete-survey',
  'respond-survey'
])

const { formatDatetimeDisplay } = useDatetimeInput()
const formatDate = (date) => formatDatetimeDisplay(date) || 'No definida'

const getStatusText = (status) => {
  const statusMap = {
    'active': 'Activa',
    'expired': 'Expirada',
    'draft': 'Borrador'
  }
  return statusMap[status] || status
}

const getResponsePercentage = (survey) => {
  if (!survey.totalStudents || survey.totalStudents === 0) return 0
  return Math.round((survey.responsesCount / survey.totalStudents) * 100)
}
</script>

<style scoped>
.surveys-tab {
  padding-top: 0;
  padding-bottom: 2rem;
}

.surveys-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.surveys-header h3 {
  margin: 0;
  color: var(--color-dark, #394049);
  font-size: 1.5rem;
}

.header-actions {
  display: flex;
  gap: 1rem;
}

.create-survey-btn,
.create-peer-evaluation-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  border: none;
  border-radius: var(--radius-sm, 6px);
  font-weight: 500;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
}

.create-survey-btn {
  background-color: var(--color-primary, #00A499);
  color: white;
}

.create-survey-btn:hover {
  background-color: var(--color-primary-dark, #008a7f);
  box-shadow: 0 4px 12px rgba(0, 164, 153, 0.3);
}

.create-peer-evaluation-btn {
  background-color: var(--color-accent, #EA7600);
  color: white;
}

.create-peer-evaluation-btn:hover {
  background-color: #d06a00;
  box-shadow: 0 4px 12px rgba(234, 118, 0, 0.3);
}

.loading-surveys {
  text-align: center;
  padding: 3rem;
  color: #94a3b8;
}

.no-surveys {
  text-align: center;
  padding: 3rem;
  color: #4b5563;
  border: 2px dashed var(--color-border, #dfe4ea);
  border-radius: var(--radius-md, 10px);
  background: var(--color-background, #f4f6f9);
}

.surveys-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1rem;
}

.survey-card {
  background: var(--color-surface, #ffffff);
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-md, 10px);
  padding: 1.5rem;
  transition: all 0.2s;
  position: relative;
  overflow: hidden;
}

.survey-card:hover {
  box-shadow: var(--shadow-md, 0 4px 6px rgba(0,0,0,0.08));
  transform: translateY(-1px);
}

.survey-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
}

.survey-active::before {
  background: var(--color-primary, #00A499);
}

.survey-expired::before {
  background: var(--color-danger, #d7373f);
}

.survey-draft::before {
  background: #94a3b8;
}

.survey-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.survey-title {
  margin: 0;
  color: var(--color-dark, #394049);
  font-size: 1.1rem;
  font-weight: 600;
  flex: 1;
}

.survey-plugin {
  background: var(--color-primary, #006b64);
  color: white;
  font-size: 0.75rem;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-weight: 500;
}

.survey-info {
  margin-bottom: 1rem;
}

.survey-description {
  margin: 0 0 0.75rem 0;
  color: var(--color-text-secondary, #64748b);
  font-size: 0.9rem;
  line-height: 1.4;
}

.survey-dates {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
}

.date-info {
  color: var(--color-text-secondary, #64748b);
  font-size: 0.75rem;
}

.survey-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  flex-wrap: wrap;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 36px;
  padding: 0 8px;
  border: none;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  transition: all 0.2s;
  gap: 4px;
  font-size: 0.8rem;
  font-weight: 500;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

.action-btn svg {
  flex-shrink: 0;
}

.btn-text {
  display: inline;
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

.view-btn {
  background-color: var(--color-accent, #EA7600);
  color: white;
}

.view-btn:hover {
  background-color: #d06a00;
}

.delete-btn {
  background-color: var(--color-danger, #d7373f);
  color: white;
}

.delete-btn:hover {
  background-color: #b82e35;
}

/* Responsive */
@media (max-width: 768px) {
  .surveys-header {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }
  
  .surveys-container {
    grid-template-columns: 1fr;
  }
  
  .survey-dates {
    grid-template-columns: 1fr;
  }
  
  .survey-header {
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .survey-status {
    margin-left: 0;
    align-self: flex-start;
  }
  
  .survey-actions {
    justify-content: center;
    gap: 0.25rem;
  }
  
  .action-btn {
    height: 32px;
    padding: 0 6px;
    font-size: 0.75rem;
    gap: 2px;
    min-width: auto;
    flex: none;
  }
  
  .btn-text {
    display: none;
  }
  
  .action-btn svg {
    width: 14px;
    height: 14px;
  }
}

@media (max-width: 480px) {
  .survey-actions {
    flex-direction: column;
    align-items: stretch;
    gap: 0.5rem;
  }
  
  .action-btn {
    width: 100%;
    justify-content: center;
    padding: 8px 12px;
  }
  
  .btn-text {
    display: inline;
  }
}
</style>
