<template>
    <div class="criterion-card" :class="{ selected: modelValue }">
        <div class="card-header">
            <input
                type="checkbox"
                :id="`criterion-${criterion.id}`"
                :checked="modelValue"
                @change="$emit('update:modelValue', $event.target.checked)"
                class="criterion-checkbox"
            />
            <label :for="`criterion-${criterion.id}`" class="card-title">
                <span>{{ criterion.label }}</span>
            </label>
        </div>

        <p class="card-description">
            {{ criterion.description }}
        </p>

        <div class="progress-bar">
            <div
                class="progress-fill"
                :style="{
                    width: `${progressPercentage}%`,
                }"
            ></div>
        </div>
        <p class="progress-text">
            {{ eligibleCount }} de {{ totalCount }} estudiantes
        </p>

        <div v-if="modelValue" class="criterion-config">
            <div class="config-section">
                <div class="config-header">
                    <div class="config-label">
                        <span>{{ diversityLabel }}:</span>
                        <span
                            class="help-icon"
                            :title="helpTitle"
                            >ⓘ</span
                        >
                    </div>
                    <div class="objective-badge">
                        {{ objectiveLabel }}
                    </div>
                </div>
                <div class="slider-container">
                    <input
                        type="range"
                        v-model.number="localTarget"
                        @change="onDiversityChange"
                        min="0"
                        max="1"
                        step="0.1"
                        class="slider"
                    />
                    <div class="slider-labels">
                        <span>{{ similarLabel }}</span>
                        <span class="weight-value">{{ objectiveLabel }}</span>
                        <span>{{ diverseLabel }}</span>
                    </div>
                </div>
            </div>

            <div class="config-section">
                <div class="config-header">
                    <div class="config-label">
                        <span>Importancia:</span>
                        <span
                            class="help-icon"
                            title="Determina qué tan importante es este criterio en la optimización general"
                            >ⓘ</span
                        >
                    </div>
                    <div class="weight-badge">
                        {{ weightLabel }}
                    </div>
                </div>
                <div class="slider-container">
                    <input
                        type="range"
                        v-model.number="localWeight"
                        @change="onWeightChange"
                        min="0"
                        max="1"
                        step="0.1"
                        class="slider"
                    />
                    <div class="slider-labels">
                        <span>0%</span>
                        <span class="weight-value"
                            >{{ Math.round(localWeight * 100) }}%</span
                        >
                        <span>100%</span>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { computed, ref, watch } from "vue";

const props = defineProps({
    criterion: {
        type: Object,
        required: true,
    },
    modelValue: {
        type: Boolean,
        default: false,
    },
    target: {
        type: Number,
        default: 0.5,
    },
    weight: {
        type: Number,
        default: 0.5,
    },
    eligibleCount: {
        type: Number,
        required: true,
    },
    totalCount: {
        type: Number,
        required: true,
    },
});

const emit = defineEmits([
    "update:modelValue",
    "update:target",
    "update:weight",
]);

const localWeight = ref(props.weight);
const localTarget = ref(props.target ?? 0.5);

watch(() => props.weight, (val) => { localWeight.value = val; });
watch(() => props.target, (val) => { localTarget.value = val ?? 0.5; });

const progressPercentage = computed(() => {
    if (props.totalCount === 0) return 0;
    return (props.eligibleCount / props.totalCount) * 100;
});

function onDiversityChange() {
    emit("update:target", localTarget.value);
}

function onWeightChange() {
    const v = localWeight.value;
    emit("update:weight", v);
}

const isSocial = computed(() => props.criterion?.type === "social");

const diversityLabel = computed(() =>
    isSocial.value ? "Tipo de relación" : "Diversidad",
);

const similarLabel = computed(() =>
    isSocial.value ? "Mala relación" : "Similares",
);

const diverseLabel = computed(() =>
    isSocial.value ? "Buena relación" : "Diversos",
);

const helpTitle = computed(() =>
    isSocial.value
        ? "Define si prefieres grupos con buenas o malas relaciones entre sí"
        : "Define qué tan diversos o similares deben ser los grupos en este atributo",
);

const objectiveLabel = computed(() => {
    const t = localTarget.value;
    if (isSocial.value) {
        if (t <= 0.3) return "Relaciones separadas";
        if (t >= 0.7) return "Buenas relaciones";
        return "Relaciones mixtas";
    }
    if (t <= 0.3) return "Grupos similares";
    if (t >= 0.7) return "Grupos diversos";
    return "Grupos mixtos";
});

const weightLabel = computed(() => {
    const percent = Math.round(localWeight.value * 100);

    if (percent === 0) return "Ignorar";
    if (percent <= 20) return "Muy bajo";
    if (percent <= 40) return "Bajo";
    if (percent <= 60) return "Moderado";
    if (percent <= 80) return "Alto";
    return "Crítico";
});
</script>

<style scoped>
.criterion-card {
    border: 2px solid var(--color-border);
    border-radius: 12px;
    padding: 20px;
    transition: all 0.3s;
    background: var(--color-surface);
}

.criterion-card:hover {
    border-color: var(--color-primary);
    box-shadow: 0 4px 12px rgba(0,164,153,0.04);
}

.criterion-card.selected {
    border-color: var(--color-primary-dark);
    /* Selection uses primary green at 80% opacity as requested */
    background: #00a49911;
    box-shadow: 0 4px 16px rgba(59, 163, 115, 0.12);
}

.card-header {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    margin-bottom: 12px;
}

.criterion-checkbox {
    width: 20px;
    height: 20px;
    cursor: pointer;
    accent-color: var(--color-primary);
    margin-top: 2px;
}

.card-title {
    flex: 1;
    display: flex;
    justify-content: space-between;
    align-items: center;
    cursor: pointer;
    font-size: 16px;
    font-weight: 600;
    color: var(--color-dark);
}

.card-description {
    color: var(--color-text-secondary);
    font-size: 14px;
    margin: 0 0 16px 32px;
    line-height: 1.5;
}

.progress-bar {
    height: 6px;
    background: var(--color-border);
    border-radius: 3px;
    overflow: hidden;
    margin: 0 0 8px 32px;
}

.progress-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--color-primary) 0%, var(--color-primary-dark) 100%);
    transition: width 0.3s;
}

.progress-text {
    color: var(--color-text-secondary);
    font-size: 13px;
    margin: 0 0 16px 32px;
}

.criterion-config {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid var(--color-border);
    display: flex;
    flex-direction: column;
    gap: 20px;
}

.config-section {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.config-header {
    margin-bottom: 8px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
}

.config-label {
    font-size: 14px;
    font-weight: 600;
    color: #1f2937;
    display: flex;
    align-items: center;
    gap: 6px;
}

.help-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    background: var(--color-border);
    color: var(--color-dark);
    border-radius: 50%;
    font-size: 11px;
    font-weight: 600;
    cursor: help;
    transition: all 0.2s;
}

.help-icon:hover {
    background: var(--color-primary);
    color: white;
}

.slider-container {
    flex: 1;
}

.slider {
    width: 100%;
    height: 6px;
    border-radius: 3px;
    background: var(--color-border);
    outline: none;
    -webkit-appearance: none;
    appearance: none;
    cursor: pointer;
}

.slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: var(--color-primary);
    cursor: pointer;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    transition: transform 0.1s ease, box-shadow 0.1s ease;
}

.slider::-moz-range-thumb {
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: var(--color-primary);
    cursor: pointer;
    border: none;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    transition: transform 0.1s ease, box-shadow 0.1s ease;
}

.slider-labels {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 8px;
    font-size: 12px;
    color: var(--color-text-secondary);
}

.weight-value {
    font-weight: 600;
    color: var(--color-primary-dark);
    font-size: 14px;
}

.objective-badge,
.weight-badge {
    padding: 6px 12px;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 600;
    white-space: nowrap;
}

.objective-badge {
    /* softer badge background to reduce strong green appearance */
    background: rgba(0,164,153,0.06);
    color: var(--color-primary-dark);
    border: 1px solid rgba(0,164,153,0.12);
}

.weight-badge {
    background: var(--color-accent-light);
    color: var(--color-accent-dark);
    border: 1px solid var(--color-accent);
}
</style>
