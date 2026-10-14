<template>
  <div class="pareto-modal-overlay" @click.self="$emit('dismiss')">
    <div class="pareto-modal">
      <div class="pareto-header">
        <button class="pareto-back" @click="$emit('back')" title="Volver a prioridad">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="19" y1="12" x2="5" y2="12" />
            <polyline points="12 19 5 12 12 5" />
          </svg>
          Volver
        </button>
        <h2>Frente de Pareto — Soluciones Multi-Objetivo</h2>
        <button class="pareto-close" @click="$emit('dismiss')">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      </div>

      <p class="pareto-description">
        Cada punto representa una solución posible de formación de grupos.
        El eje X muestra el score psicométrico y el eje Y el score social.
        Selecciona la solución que prefieras.
      </p>

      <div class="pareto-chart-container">
        <Scatter :data="chartData" :options="chartOptions" />
      </div>

      <div v-if="selectedSolution" class="pareto-selection">
        <div class="pareto-selection-info">
          <span class="pareto-badge psy">
            Psicométrico: {{ (selectedSolution.psy_score * 100).toFixed(1) }}%
          </span>
          <span class="pareto-badge soc">
            Social: {{ (selectedSolution.soc_score * 100).toFixed(1) }}%
          </span>
          <span class="pareto-badge global">
            Global: {{ (selectedSolution.global_score * 100).toFixed(1) }}%
          </span>
        </div>
        <div class="pareto-actions">
          <button class="pareto-btn pareto-btn-confirm" @click="$emit('select', selectedSolution)">
            Confirmar esta solución
          </button>
          <button class="pareto-btn pareto-btn-cancel" @click="selectedSolution = null">
            Deseleccionar
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from "vue";
import { Scatter } from "vue-chartjs";
import {
  Chart as ChartJS,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
} from "chart.js";

ChartJS.register(LinearScale, PointElement, LineElement, Tooltip, Legend);

const props = defineProps({
  solutions: {
    type: Array,
    required: true,
  },
  kneeIndex: {
    type: Number,
    default: -1,
  },
});

const emit = defineEmits(["select", "dismiss", "back"]);

const selectedSolution = ref(null);

const deduplicatedSolutions = computed(() => {
  const seen = new Set();
  return props.solutions.filter((s) => {
    const key = `${s.psy_score.toFixed(6)}_${s.soc_score.toFixed(6)}`;
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
});

const chartData = computed(() => {
  const regular = [];
  const knee = [];
  const selected = [];

  deduplicatedSolutions.value.forEach((s, i) => {
    const point = { x: s.psy_score, y: s.soc_score };
    if (selectedSolution.value && selectedSolution.value.id === s.id) {
      selected.push(point);
    } else if (props.kneeIndex >= 0 && s.id === props.kneeIndex) {
      knee.push(point);
    } else {
      regular.push(point);
    }
  });

  const datasets = [
    {
      label: "Soluciones",
      data: regular,
      backgroundColor: "rgba(59, 130, 246, 0.6)",
      borderColor: "rgba(59, 130, 246, 1)",
      pointRadius: 6,
      pointHoverRadius: 9,
    },
  ];

  if (knee.length > 0) {
    datasets.push({
      label: "Knee Point (recomendado)",
      data: knee,
      backgroundColor: "rgba(16, 185, 129, 0.8)",
      borderColor: "rgba(16, 185, 129, 1)",
      pointRadius: 10,
      pointHoverRadius: 13,
      pointStyle: "star",
    });
  }

  if (selected.length > 0) {
    datasets.push({
      label: "Seleccionada",
      data: selected,
      backgroundColor: "rgba(239, 68, 68, 0.9)",
      borderColor: "rgba(239, 68, 68, 1)",
      pointRadius: 12,
      pointHoverRadius: 15,
      pointStyle: "rectRot",
    });
  }

  return { datasets };
});

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  animation: { duration: 200 },
  scales: {
    x: {
      title: {
        display: true,
        text: "Score Psicométrico",
        font: { size: 14, weight: "bold" },
      },
      min: 0,
      max: 1,
      grid: { color: "rgba(0,0,0,0.06)" },
    },
    y: {
      title: {
        display: true,
        text: "Score Social",
        font: { size: 14, weight: "bold" },
      },
      min: 0,
      max: 1,
      grid: { color: "rgba(0,0,0,0.06)" },
    },
  },
  plugins: {
    tooltip: {
      callbacks: {
        label: (ctx) => {
          const s = deduplicatedSolutions.value.find(
            (sol) =>
              Math.abs(sol.psy_score - ctx.parsed.x) < 0.0001 &&
              Math.abs(sol.soc_score - ctx.parsed.y) < 0.0001,
          );
          if (!s) return `(${ctx.parsed.x.toFixed(3)}, ${ctx.parsed.y.toFixed(3)})`;
          return [
            `Psicométrico: ${(s.psy_score * 100).toFixed(1)}%`,
            `Social: ${(s.soc_score * 100).toFixed(1)}%`,
            `Global: ${(s.global_score * 100).toFixed(1)}%`,
            `Grupos: ${s.teams?.length ?? "?"}`,
          ];
        },
      },
    },
    legend: {
      position: "top",
    },
  },
  onClick: (_event, elements) => {
    if (elements.length === 0) return;
    const el = elements[0];
    const dataset = chartData.value.datasets[el.datasetIndex];
    const point = dataset.data[el.index];
    const found = deduplicatedSolutions.value.find(
      (s) =>
        Math.abs(s.psy_score - point.x) < 0.0001 &&
        Math.abs(s.soc_score - point.y) < 0.0001,
    );
    if (found) {
      selectedSolution.value = found;
    }
  },
}));
</script>

<style scoped>
.pareto-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(2px);
}

.pareto-modal {
  background: var(--color-surface, #ffffff);
  border-radius: var(--radius-lg, 14px);
  padding: 24px;
  width: 90vw;
  max-width: 800px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: var(--shadow-xl, 0 8px 28px rgba(57,64,73,0.12));
}

.pareto-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  gap: 12px;
}

.pareto-header h2 {
  margin: 0;
  font-size: 1.2rem;
  color: var(--color-dark, #394049);
  font-family: var(--font-family-titles, "Red Hat Display", sans-serif);
  flex: 1;
  text-align: center;
}

.pareto-back {
  background: none;
  border: 2px solid var(--color-border, #dfe4ea);
  cursor: pointer;
  color: var(--color-dark-soft, #4f5660);
  padding: 6px 14px;
  border-radius: var(--radius-sm, 6px);
  font-size: 0.85rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.15s;
  flex-shrink: 0;
}

.pareto-back:hover {
  background: var(--color-primary-light, #e6f7f6);
  border-color: var(--color-primary, #00A499);
  color: var(--color-primary, #00A499);
}

.pareto-close {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-secondary, #64707d);
  padding: 6px;
  border-radius: var(--radius-sm, 6px);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.15s;
}

.pareto-close:hover {
  background: var(--color-background, #f4f6f9);
  color: var(--color-dark, #394049);
}

.pareto-description {
  color: var(--color-text-secondary, #64707d);
  font-size: 0.9rem;
  margin-bottom: 16px;
}

.pareto-chart-container {
  height: 400px;
  margin-bottom: 16px;
}

.pareto-selection {
  background: var(--color-background, #f4f6f9);
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: var(--radius-md, 10px);
  padding: 16px;
}

.pareto-selection-info {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.pareto-badge {
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 0.85rem;
  font-weight: 600;
}

.pareto-badge.psy {
  background: var(--color-primary-light, #e6f7f6);
  color: var(--color-primary-dark, #008a81);
}

.pareto-badge.soc {
  background: var(--color-accent-light, #fff3e6);
  color: var(--color-accent-dark, #c96500);
}

.pareto-badge.global {
  background: var(--color-background, #f4f6f9);
  color: var(--color-dark, #394049);
  border: 1px solid var(--color-border, #dfe4ea);
}

.pareto-actions {
  display: flex;
  gap: 10px;
}

.pareto-btn {
  padding: 10px 20px;
  border-radius: var(--radius-md, 10px);
  border: none;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
}

.pareto-btn-confirm {
  background: var(--color-primary, #00A499);
  color: white;
  box-shadow: var(--shadow-md, 0 2px 8px rgba(0,164,153,0.25));
}

.pareto-btn-confirm:hover {
  background: var(--color-primary-dark, #008a81);
}

.pareto-btn-cancel {
  background: var(--color-surface, #ffffff);
  color: var(--color-dark-soft, #4f5660);
  border: 2px solid var(--color-border, #dfe4ea);
}

.pareto-btn-cancel:hover {
  border-color: var(--color-primary, #00A499);
  color: var(--color-primary, #00A499);
}
</style>
