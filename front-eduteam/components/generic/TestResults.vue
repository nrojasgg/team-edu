<template>
  <div v-if="isOpen" class="results-modal-overlay" @click.self="$emit('close')">
    <div class="results-modal">
      <div class="results-modal-header">
        <div class="header-content">
          <h2>Resultados del Test</h2>
          <p class="subtitle">{{ pluginName }}</p>
        </div>
        <button @click="$emit('close')" class="close-btn">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <div class="results-modal-body">
        <div v-if="loading" class="loading-state">
          <div class="spinner"></div>
          <p>Cargando resultados...</p>
        </div>

        <div v-else-if="error" class="error-state">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="8" x2="12" y2="12"/>
            <line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
          <p>{{ error }}</p>
          <button @click="$emit('close')" class="btn-secondary">Cerrar</button>
        </div>

        <div v-else-if="results" class="results-content">
          <!-- SECCIÓN 1: Resultado Categórico Principal -->
          <div v-if="mainMessage || mbtiType || categoricalResult" class="section categorical-section">
            <h3 class="section-title">Resultado Principal</h3>
            
            <!-- MBTI Type Badge -->
            <div v-if="mbtiType" class="mbti-type-badge">
              {{ mbtiType }}
            </div>
            
            <!-- Enneagram or Chronotype main result -->
            <div v-if="mainMessage" class="main-result">
              <div class="result-icon">
                <svg v-if="resultCategory === 'positive'" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
                  <polyline points="22 4 12 14.01 9 11.01"/>
                </svg>
                <svg v-else width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <polyline points="12 6 12 12 16 14"/>
                </svg>
              </div>
              <h3>{{ mainMessage }}</h3>
              <p v-if="resultDescription" class="result-description">{{ resultDescription }}</p>
            </div>

            <!-- Chronotype category badge -->
            <div v-if="categoricalResult && !mbtiType" class="category-result">
              <div class="category-badge" :class="categoryClass">
                {{ categoricalResult.value }}
              </div>
              <p v-if="categoricalResult.label" class="category-label">{{ categoricalResult.label }}</p>
            </div>
          </div>

          <!-- SECCIÓN 2: Datos Dimensionales -->
          <div v-if="mbtiDimensions || enneagramScores" class="section dimensions-section">
            <h3 class="section-title">Puntuaciones Dimensionales</h3>
            
            <!-- MBTI Dimensions -->
            <div v-if="mbtiDimensions" class="dimensions-grid">
              <div 
                v-for="dim in mbtiDimensions" 
                :key="dim.name"
                class="dimension-item"
              >
                <div class="dimension-header">
                  <span class="dimension-name">{{ getDimensionLabel(dim.name) }}</span>
                  <span class="dimension-preference">{{ dim.preference }}</span>
                </div>
                <div class="dimension-scores">
                  <span class="dim-a">A: {{ dim.score_a }}</span>
                  <span class="dim-b">B: {{ dim.score_b }}</span>
                </div>
                <div class="dimension-bar">
                  <div 
                    class="dim-bar-fill" 
                    :style="{ width: `${(dim.score_a / (dim.score_a + dim.score_b)) * 100}%` }"
                  ></div>
                </div>
              </div>
            </div>

            <!-- Enneagram Scores E1-E9 -->
            <div v-if="enneagramScores" class="scores-grid">
              <div 
                v-for="(score, type_) in enneagramScores" 
                :key="type_"
                class="eneagram-score-item"
                :class="{ 'dominant': isDominantType(type_) }"
              >
                <span class="eneagram-type">E{{ type_ }}</span>
                <span class="eneagram-value">{{ score.toFixed(1) }}</span>
              </div>
            </div>
          </div>

          <!-- SECCIÓN 3: Datos Sociales (OCEAN) -->
          <div v-if="oceanVector" class="section social-section">
            <h3 class="section-title">Perfil Social OCEAN</h3>
            <div class="ocean-bars">
              <div v-for="(value, key) in oceanVector" :key="key" class="ocean-bar-item">
                <div class="ocean-bar-label">
                  <span class="ocean-key">{{ getOceanLabel(key) }}</span>
                  <span class="ocean-value">{{ (value * 100).toFixed(0) }}%</span>
                </div>
                <div class="ocean-bar">
                  <div 
                    class="ocean-bar-fill" 
                    :style="{ width: `${value * 100}%`, backgroundColor: getOceanColor(key) }"
                  ></div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="no-results">
          <p>No se encontraron resultados.</p>
        </div>

        <div class="results-actions">
          <button @click="$emit('close')" class="btn-primary">
            Cerrar
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  isOpen: {
    type: Boolean,
    default: false
  },
  results: {
    type: Object,
    default: null
  },
  pluginName: {
    type: String,
    default: ''
  },
  loading: {
    type: Boolean,
    default: false
  },
  error: {
    type: String,
    default: ''
  }
})

defineEmits(['close'])

const oceanLabels = {
  openness: 'Apertura',
  conscientiousness: 'Conciencia',
  extraversion: 'Extraversión',
  agreeableness: 'Amabilidad',
  neuroticism: 'Neuroticismo'
}

const oceanColors = {
  openness: '#ef4444',
  conscientiousness: '#f59e0b',
  extraversion: '#3b82f6',
  agreeableness: '#10b981',
  neuroticism: '#8b5cf6'
}

const getOceanLabel = (key) => oceanLabels[key] || key
const getOceanColor = (key) => oceanColors[key] || '#6b7280'

// Normalize results - handle both flat and nested formats
const normalizedResults = computed(() => {
  if (!props.results) return {}
  // If results has a nested 'results' property, use that (some API responses nest results)
  if (props.results.results && typeof props.results.results === 'object') {
    return props.results.results
  }
  return props.results
})

const resultCategory = computed(() => {
  if (!normalizedResults.value) return 'neutral'
  // Positive categories: matutino types, good scores
  if (normalizedResults.value.category?.includes('matutino') || normalizedResults.value.category?.includes('extremo')) {
    return 'positive'
  }
  return 'neutral'
})

const mainMessage = computed(() => {
  if (!normalizedResults.value) return ''
  
  // Chronotype
  if (normalizedResults.value.category) {
    return `Tu cronotipo es: ${normalizedResults.value.category}`
  }
  
  // Enneagram
  if (normalizedResults.value.dominant_type) {
    return `Tu Eneatipo dominante es: ${normalizedResults.value.dominant_type}`
  }
  
  return ''
})

const resultDescription = computed(() => {
  if (!normalizedResults.value) return ''
  
  if (normalizedResults.value.category === 'Matutino extremo') {
    return 'Eres una persona claramente matutina. Tu energía es más alta en las horas de la mañana.'
  }
  if (normalizedResults.value.category === 'Matutino moderado') {
    return 'Tienes tendencias matutinas. rindes mejor durante la mañana.'
  }
  if (normalizedResults.value.category === 'Intermedio') {
    return 'Tienes un patrón de sueño equilibrado. Puedes adaptarte a diferentes horarios.'
  }
  if (normalizedResults.value.category === 'Vespertino moderado') {
    return 'Tienes tendencias vespertinas. Tu energía es más alta hacia la tarde-noche.'
  }
  if (normalizedResults.value.category === 'Vespertino extremo') {
    return 'Eres claramente vespertino. Tu mejor momento es durante la tarde-noche.'
  }
  
  return ''
})

const categoryClass = computed(() => {
  if (!normalizedResults.value) return ''
  if (normalizedResults.value.dominant_type) return `type-${normalizedResults.value.dominant_type}`
  if (normalizedResults.value.category?.includes('matutino')) return 'category-morning'
  if (normalizedResults.value.category?.includes('vespertino')) return 'category-evening'
  return ''
})

const categoricalResult = computed(() => {
  if (!normalizedResults.value) return null
  
  if (normalizedResults.value.dominant_type) {
    return {
      value: normalizedResults.value.dominant_type,
      label: 'Eneatipo Dominante'
    }
  }
  
  if (normalizedResults.value.category) {
    return {
      value: normalizedResults.value.category,
      label: 'Categoría de Cronotipo'
    }
  }
  
  return null
})

const numericScore = computed(() => {
  if (!normalizedResults.value) return null
  
  if (normalizedResults.value.total_score !== undefined) {
    return {
      value: normalizedResults.value.total_score,
      max: 86,
      label: 'Puntuación Total (rango 16-86)'
    }
  }
  
  return null
})

const oceanVector = computed(() => {
  if (!normalizedResults.value?.ocean_vector) return null
  return normalizedResults.value.ocean_vector
})

const enneagramScores = computed(() => {
  if (!normalizedResults.value) return null
  
  // Check for E1-E9 scores (enneagram)
  const scores = {}
  for (let i = 1; i <= 9; i++) {
    const key = `e${i}`
    if (normalizedResults.value[key] !== undefined) {
      scores[i] = normalizedResults.value[key]
    }
  }
  
  return Object.keys(scores).length > 0 ? scores : null
})

const mbtiDimensions = computed(() => {
  if (!normalizedResults.value?.dimensions) return null
  return normalizedResults.value.dimensions
})

const mbtiType = computed(() => {
  if (!normalizedResults.value) return null
  return normalizedResults.value.mbti_type || null
})

const isDominantType = (type_) => {
  return normalizedResults.value?.dominant_type === parseInt(type_)
}

const dimensionLabels = {
  'EI': 'Extraversión vs Introversión',
  'SN': 'Sensorial vs Intuición',
  'TF': 'Pensamiento vs Sentimiento',
  'JP': 'Juicio vs Percepción'
}

const getDimensionLabel = (name) => dimensionLabels[name] || name
</script>

<style scoped>
.results-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.results-modal {
  background: white;
  border-radius: 16px;
  max-width: 600px;
  width: 100%;
  max-height: 90vh;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
}

.results-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem 2rem;
  border-bottom: 1px solid var(--color-border, #dfe4ea);
  background: linear-gradient(135deg, #059669 0%, #10b981 100%);
  color: white;
}

.header-content h2 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
}

.subtitle {
  margin: 0;
  font-size: 0.875rem;
  opacity: 0.9;
}

.close-btn {
  background: none;
  border: none;
  color: white;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 8px;
  transition: background 0.2s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.results-modal-body {
  padding: 2rem;
  overflow-y: auto;
  flex: 1;
}

.loading-state, .error-state, .no-results {
  text-align: center;
  padding: 3rem;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #10b981;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.results-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.section {
  background: #f9fafb;
  border-radius: 12px;
  padding: 1.25rem;
}

.section-title {
  margin: 0 0 1rem 0;
  color: #374151;
  font-size: 0.875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.categorical-section {
  background: linear-gradient(135deg, #f0fdf4 0%, #ecfdf5 100%);
  border: 1px solid #10b981;
}

.main-result {
  text-align: center;
  padding: 1rem;
}

.result-icon {
  margin-bottom: 0.75rem;
  color: #10b981;
}

.main-result h3 {
  margin: 0;
  color: #065f46;
  font-size: 1.25rem;
}

.result-description {
  margin: 0.5rem 0 0 0;
  color: #047857;
  font-size: 0.875rem;
}

.mbti-type-badge {
  display: block;
  width: 120px;
  height: 120px;
  margin: 0 auto 1rem;
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: white;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2.5rem;
  font-weight: 700;
  box-shadow: 0 8px 24px rgba(59, 130, 246, 0.4);
}

.category-result {
  text-align: center;
}

.category-badge {
  display: inline-block;
  padding: 1rem 2rem;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border-radius: 50%;
  font-size: 1.5rem;
  font-weight: 700;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
}

.category-badge.type-1 { background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%); box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4); }
.category-badge.type-2 { background: linear-gradient(135deg, #f97316 0%, #ea580c 100%); box-shadow: 0 4px 12px rgba(249, 115, 22, 0.4); }
.category-badge.type-3 { background: linear-gradient(135deg, #eab308 0%, #ca8a04 100%); box-shadow: 0 4px 12px rgba(234, 179, 8, 0.4); }
.category-badge.type-4 { background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%); box-shadow: 0 4px 12px rgba(34, 197, 94, 0.4); }
.category-badge.type-5 { background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%); box-shadow: 0 4px 12px rgba(6, 182, 212, 0.4); }
.category-badge.type-6 { background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%); box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4); }
.category-badge.type-7 { background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%); box-shadow: 0 4px 12px rgba(139, 92, 246, 0.4); }
.category-badge.type-8 { background: linear-gradient(135deg, #f43f5e 0%, #e11d48 100%); box-shadow: 0 4px 12px rgba(244, 63, 94, 0.4); }
.category-badge.type-9 { background: linear-gradient(135deg, #64748b 0%, #475569 100%); box-shadow: 0 4px 12px rgba(100, 116, 139, 0.4); }

.category-label {
  margin: 0.75rem 0 0 0;
  color: #6b7280;
  font-weight: 500;
  font-size: 0.875rem;
}

/* Dimensional Section */
.dimensions-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}

.dimension-item {
  background: white;
  border-radius: 8px;
  padding: 0.75rem;
  border: 1px solid #e5e7eb;
}

.dimension-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.dimension-name {
  font-weight: 600;
  color: #374151;
  font-size: 0.875rem;
}

.dimension-preference {
  font-weight: 700;
  color: #3b82f6;
  font-size: 1rem;
}

.dimension-scores {
  display: flex;
  gap: 1rem;
  font-size: 0.75rem;
  color: #6b7280;
  margin-bottom: 0.25rem;
}

.dimension-bar {
  height: 6px;
  background: #e5e7eb;
  border-radius: 3px;
  overflow: hidden;
}

.dim-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #8b5cf6);
  border-radius: 3px;
  transition: width 0.5s ease;
}

/* Scores Grid (for Enneagram) */
.scores-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.5rem;
}

.eneagram-score-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0.75rem;
  background: white;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
}

.eneagram-score-item.dominant {
  border-color: #10b981;
  background: #f0fdf4;
}

.eneagram-type {
  font-weight: 600;
  color: #374151;
}

.eneagram-value {
  font-weight: 500;
  color: #6b7280;
}

/* Social Section (OCEAN) */
.ocean-bar-item {
  margin-bottom: 0.75rem;
}

.ocean-bar-item:last-child {
  margin-bottom: 0;
}

.ocean-bar-label {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.25rem;
}

.ocean-key {
  font-weight: 500;
  color: #4b5563;
  font-size: 0.875rem;
}

.ocean-value {
  font-weight: 600;
  color: #374151;
}

.ocean-bar {
  height: 8px;
  background: #e5e7eb;
  border-radius: 4px;
  overflow: hidden;
}

.ocean-bar-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.5s ease;
}

.results-actions {
  margin-top: 1.5rem;
  display: flex;
  justify-content: center;
  gap: 1rem;
}

.btn-primary {
  padding: 0.75rem 2rem;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
}

.btn-secondary {
  padding: 0.75rem 1.5rem;
  background: #e5e7eb;
  color: #374151;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
}

@media (max-width: 480px) {
  .dimensions-grid {
    grid-template-columns: 1fr;
  }
  .scores-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}
</style>
