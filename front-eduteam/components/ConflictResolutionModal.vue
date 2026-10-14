<template>
  <div class="conflict-overlay" @click.self="$emit('dismiss')">
    <div class="conflict-modal">
      <!-- Header -->
      <div class="conflict-header">
        <div class="conflict-header-icon">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
            <line x1="12" y1="9" x2="12" y2="13" />
            <line x1="12" y1="17" x2="12.01" y2="17" />
          </svg>
        </div>
        <div>
          <h2>Has seleccionado criterios con objetivos conflictivos</h2>
          <p class="conflict-subtitle">
            Los criterios elegidos pertenecen a funciones distintas y pueden generar soluciones diferentes.
            Elige cómo quieres priorizarlos.
          </p>
        </div>
        <button class="conflict-close" @click="$emit('dismiss')">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      </div>

      <!-- Criteria Categories -->
      <div class="conflict-categories">
        <!-- Structural / Psychometric -->
        <div class="category-card category-psy">
          <div class="category-header">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
            </svg>
            <h3>Funcion Estructural</h3>
          </div>
          <p class="category-description">
            Criterios basados en personalidad y atributos individuales de los estudiantes.
          </p>
          <div class="criteria-list">
            <div
              v-for="criterion in psychometricCriteria"
              :key="criterion.name"
              class="criterion-chip"
            >
              <span class="criterion-name">{{ criterion.label }}</span>
              <span class="criterion-target">Objetivo: {{ (criterion.target * 100).toFixed(0) }}%</span>
            </div>
          </div>
        </div>

        <!-- Relational / Social -->
        <div class="category-card category-soc">
          <div class="category-header">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2" />
              <circle cx="9" cy="7" r="4" />
              <path d="M23 21v-2a4 4 0 00-3-3.87" />
              <path d="M16 3.13a4 4 0 010 7.75" />
            </svg>
            <h3>Funcion Relacional</h3>
          </div>
          <p class="category-description">
            Criterios basados en preferencias sociales y compatibilidad entre estudiantes.
          </p>
          <div class="criteria-list">
            <div
              v-for="criterion in socialCriteria"
              :key="criterion.name"
              class="criterion-chip"
            >
              <span class="criterion-name">{{ criterion.label }}</span>
              <span class="criterion-target">Objetivo: {{ (criterion.target * 100).toFixed(0) }}%</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Priority Slider -->
      <div class="slider-section">
        <h3>Prioridad de objetivos</h3>
        <p class="slider-description">
          Ajusta el peso relativo entre los dos tipos de funcion.
        </p>

        <div class="slider-container">
          <div class="slider-labels">
            <span class="slider-label slider-label-psy" :class="{ 'active': psyWeight >= 60 }">
              Estructural
              <strong>{{ sliderValue }}%</strong>
            </span>
            <span class="slider-label slider-label-soc" :class="{ 'active': socWeight >= 60 }">
              <strong>{{ 100 - sliderValue }}%</strong>
              Relacional
            </span>
          </div>

          <div class="slider-track-wrapper">
            <input
              type="range"
              :value="sliderValue"
              min="0"
              max="100"
              step="5"
              class="priority-slider"
              @input="sliderValue = Number($event.target.value)"
            />
          </div>

          <div class="slider-center-label">
            <span v-if="Math.abs(sliderValue - 50) <= 10">Equilibrado</span>
            <span v-else-if="sliderValue > 50">Prioridad Estructural</span>
            <span v-else>Prioridad Relacional</span>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="conflict-actions">
        <button class="action-btn action-weighted" @click="handleWeightedSum">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2v20M2 12h20" />
          </svg>
          Formar equipos con esta prioridad
        </button>
        <button class="action-btn action-pareto" @click="$emit('show-pareto')">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="6" cy="18" r="3" />
            <circle cx="12" cy="12" r="3" />
            <circle cx="18" cy="6" r="3" />
          </svg>
          Ver todas las opciones
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  criteriaParameters: {
    type: Object,
    required: true,
  },
  paretoFront: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['weighted-solution', 'show-pareto', 'dismiss'])

const sliderValue = ref(50)

// Helper to check if a criterion belongs to the relational/social objective
const isSocialCriterion = (name, config) => {
  const type = (config?.type || '').toLowerCase();
  if (type === 'social' || type === 'relational' || type === 'relacional') {
    return true;
  }
  if (type === 'numeric' || type === 'statistical' || type === 'estadistico' || type === 'categorical' || type === 'categorico') {
    return false;
  }
  // Fallback si no viene type especificado
  const lowerName = name.toLowerCase();
  return lowerName.includes('soc') || lowerName.includes('affinit') || lowerName.includes('peer');
};

const categorizeCriteria = () => {
  const psychometric = []
  const social = []

  for (const [name, config] of Object.entries(props.criteriaParameters || {})) {
    const entry = {
      name,
      label: config.label || name,
      target: config.target ?? 0.5,
      weight: config.weight ?? 0.5,
      goal: config.goal || '',
      type: config.type || '',
    }

    if (isSocialCriterion(name, config)) {
      social.push(entry)
    } else {
      psychometric.push(entry)
    }
  }

  return { psychometric, social }
}

const psychometricCriteria = computed(() => categorizeCriteria().psychometric)
const socialCriteria = computed(() => categorizeCriteria().social)

const psyWeight = computed(() => sliderValue.value)
const socWeight = computed(() => 100 - sliderValue.value)

const handleWeightedSum = () => {
  // Pick the best solution from the Pareto front using weighted sum
  if (!props.paretoFront || props.paretoFront.length === 0) {
    // No Pareto front available, just emit with weights
    emit('weighted-solution', { psyWeight: psyWeight.value, socWeight: socWeight.value })
    return
  }

  const wPsy = psyWeight.value / 100
  const wSoc = socWeight.value / 100

  let bestSolution = null
  let bestScore = -Infinity

  for (const solution of props.paretoFront) {
    const score = (solution.psy_score * wPsy) + (solution.soc_score * wSoc)
    if (score > bestScore) {
      bestScore = score
      bestSolution = solution
    }
  }

  if (bestSolution) {
    emit('weighted-solution', {
      psyWeight: psyWeight.value,
      socWeight: socWeight.value,
      solution: bestSolution,
    })
  }
}
</script>

<style scoped>
.conflict-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(2px);
}

.conflict-modal {
  background: var(--color-surface, #ffffff);
  border-radius: var(--radius-lg, 14px);
  padding: 28px;
  width: 90vw;
  max-width: 720px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: var(--shadow-xl, 0 8px 28px rgba(57,64,73,0.12));
}

/* Header */
.conflict-header {
  display: flex;
  gap: 14px;
  align-items: flex-start;
  margin-bottom: 24px;
}

.conflict-header-icon {
  flex-shrink: 0;
  color: var(--color-accent, #EA7600);
  background: var(--color-accent-light, #fff3e6);
  border-radius: 50%;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.conflict-header h2 {
  margin: 0 0 6px 0;
  font-size: 1.2rem;
  color: var(--color-dark, #394049);
  font-weight: 700;
  font-family: var(--font-family-titles, "Red Hat Display", sans-serif);
}

.conflict-subtitle {
  margin: 0;
  color: var(--color-text-secondary, #64707d);
  font-size: 0.875rem;
  line-height: 1.5;
}

.conflict-close {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-secondary, #64707d);
  padding: 6px;
  border-radius: var(--radius-sm, 6px);
  margin-left: auto;
  flex-shrink: 0;
  transition: all 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.conflict-close:hover {
  background: var(--color-background, #f4f6f9);
  color: var(--color-dark, #394049);
}

/* Categories */
.conflict-categories {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 24px;
}

@media (max-width: 600px) {
  .conflict-categories {
    grid-template-columns: 1fr;
  }
}

.category-card {
  border-radius: var(--radius-md, 10px);
  padding: 18px;
  border: 2px solid transparent;
  transition: all 0.2s;
}

.category-psy {
  background: var(--color-primary-light, #e6f7f6);
  border-color: var(--color-primary, #00A499);
}

.category-psy svg {
  color: var(--color-primary, #00A499);
}

.category-soc {
  background: var(--color-accent-light, #fff3e6);
  border-color: var(--color-accent, #EA7600);
}

.category-soc svg {
  color: var(--color-accent, #EA7600);
}

.category-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.category-header h3 {
  margin: 0;
  font-size: 1rem;
  color: var(--color-dark, #394049);
  font-weight: 700;
}

.category-description {
  color: var(--color-text-secondary, #64707d);
  font-size: 0.8rem;
  margin: 0 0 12px 0;
  line-height: 1.4;
}

.criteria-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.criterion-chip {
  display: flex;
  flex-direction: column;
  background: var(--color-surface, #ffffff);
  border-radius: var(--radius-sm, 6px);
  padding: 6px 10px;
  font-size: 0.78rem;
  box-shadow: var(--shadow-sm, 0 1px 2px rgba(57,64,73,0.06));
}

.criterion-name {
  font-weight: 600;
  color: var(--color-text, #2b2f36);
}

.criterion-target {
  color: var(--color-text-secondary, #64707d);
  font-size: 0.72rem;
}

/* Slider */
.slider-section {
  background: var(--color-background, #f4f6f9);
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-md, 10px);
  padding: 20px;
  margin-bottom: 24px;
}

.slider-section h3 {
  margin: 0 0 4px 0;
  font-size: 1rem;
  color: var(--color-dark, #394049);
  font-weight: 700;
}

.slider-description {
  color: var(--color-text-secondary, #64707d);
  font-size: 0.82rem;
  margin: 0 0 16px 0;
}

.slider-container {
  position: relative;
}

.slider-labels {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
}

.slider-label {
  font-size: 0.82rem;
  color: var(--color-text-secondary, #64707d);
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s;
}

.slider-label.active {
  color: var(--color-dark, #394049);
  font-weight: 600;
}

.slider-track-wrapper {
  position: relative;
  height: 40px;
  display: flex;
  align-items: center;
}

.priority-slider {
  -webkit-appearance: none;
  appearance: none;
  width: 100%;
  height: 8px;
  border-radius: 4px;
  background: linear-gradient(to right,
    var(--color-primary, #00A499) 0%,
    var(--color-border, #dfe4ea) 50%,
    var(--color-accent, #EA7600) 100%);
  outline: none;
  cursor: pointer;
}

.priority-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--color-surface, #ffffff);
  border: 3px solid var(--color-primary, #00A499);
  box-shadow: var(--shadow-md, 0 2px 8px rgba(57,64,73,0.08));
  cursor: grab;
  transition: transform 0.1s;
}

.priority-slider::-webkit-slider-thumb:hover {
  transform: scale(1.15);
}

.priority-slider::-webkit-slider-thumb:active {
  cursor: grabbing;
  transform: scale(1.1);
}

.priority-slider::-moz-range-thumb {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--color-surface, #ffffff);
  border: 3px solid var(--color-primary, #00A499);
  box-shadow: var(--shadow-md, 0 2px 8px rgba(57,64,73,0.08));
  cursor: grab;
}

.slider-center-label {
  text-align: center;
  margin-top: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-dark, #394049);
}

/* Actions */
.conflict-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

@media (max-width: 500px) {
  .conflict-actions {
    flex-direction: column;
  }
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 22px;
  border-radius: var(--radius-md, 10px);
  border: none;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.action-weighted {
  background: var(--color-primary, #00A499);
  color: white;
  box-shadow: var(--shadow-md, 0 2px 8px rgba(0,164,153,0.25));
}

.action-weighted:hover {
  background: var(--color-primary-dark, #008a81);
  transform: translateY(-1px);
  box-shadow: var(--shadow-lg, 0 4px 16px rgba(0,164,153,0.3));
}

.action-pareto {
  background: var(--color-surface, #ffffff);
  color: var(--color-dark-soft, #4f5660);
  border: 2px solid var(--color-border, #dfe4ea);
}

.action-pareto:hover {
  border-color: var(--color-primary, #00A499);
  color: var(--color-primary, #00A499);
  background: var(--color-primary-light, #e6f7f6);
}
</style>