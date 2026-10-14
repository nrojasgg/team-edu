<template>
    <div class="view-content">
        <div v-if="groups.length === 0" class="no-groups">
            <p>No hay grupos definidos para esta sección.</p>
        </div>

        <div v-else class="averages-groups">
            <!-- Barra de configuración rápida de nota base -->
            <div class="grade-calculator-card">
                <div class="grade-calculator-info">
                    <svg
                        class="info-icon"
                        width="22"
                        height="22"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                    >
                        <circle cx="12" cy="12" r="10" />
                        <line x1="12" y1="16" x2="12" y2="12" />
                        <line x1="12" y1="8" x2="12.01" y2="8" />
                    </svg>
                    <div>
                        <h4 class="calculator-title">Cálculo de Notas con Evaluación entre Pares</h4>
                        <p class="calculator-desc">
                            Ingresa la <strong>nota del grupo</strong> para calcular la nota individual de cada estudiante según su desempeño relativo:
                            <code>Nota = Nota Grupal + Módulo</code> (escala 1.0 a 7.0).
                        </p>
                    </div>
                </div>
                <div class="grade-calculator-quick-apply">
                    <label class="quick-apply-label">Aplicar a todos los grupos:</label>
                    <div class="quick-apply-inputs">
                        <input
                            type="number"
                            step="0.1"
                            min="1.0"
                            max="7.0"
                            placeholder="Ej: 6.7"
                            v-model.number="globalGradeInput"
                            class="global-grade-input"
                            @keyup.enter="applyGlobalGrade"
                        />
                        <button
                            @click="applyGlobalGrade"
                            class="btn-apply-global"
                            :disabled="!isValidGlobalGrade"
                            title="Asignar esta nota base a todos los grupos"
                        >
                            Aplicar
                        </button>
                    </div>
                </div>
            </div>

            <div
                v-for="group in groups"
                :key="group.id"
                class="averages-group-section"
            >
                <div
                    class="averages-group-header"
                    @click="toggleGroup(group.id)"
                >
                    <div class="averages-header-left">
                        <h3>{{ group.name }}</h3>
                        <span class="group-students-count"
                            >{{ group.students.length }} estudiantes</span
                        >
                    </div>
                    <div class="averages-header-right">
                        <!-- Input de nota grupal específica para este grupo -->
                        <div class="group-grade-pill" @click.stop>
                            <label class="group-grade-label">Nota grupal:</label>
                            <input
                                type="number"
                                step="0.1"
                                min="1.0"
                                max="7.0"
                                placeholder="Ej: 6.7"
                                v-model.number="groupGrades[group.id]"
                                class="group-grade-input"
                                @click.stop
                                title="Ingresa la nota del proyecto o entrega para este grupo (escala 1.0 a 7.0)"
                            />
                        </div>

                        <span class="baseline-badge"
                            >Línea base:
                            {{ getGroupBaseline(group).toFixed(2) }}</span
                        >
                        <svg
                            :class="{ rotated: expandedGroup === group.id }"
                            class="chevron-icon"
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                        >
                            <polyline points="6 9 12 15 18 9" />
                        </svg>
                    </div>
                </div>

                <Transition name="accordion-slide">
                    <div
                        v-if="expandedGroup === group.id"
                        class="averages-table-container"
                    >
                        <table class="averages-table">
                            <thead>
                                <tr>
                                    <th>Estudiante</th>
                                    <th>Tareas</th>
                                    <th>Participación</th>
                                    <th>Organización</th>
                                    <th>Compromiso</th>
                                    <th>Prosocialidad</th>
                                    <th>Promedio</th>
                                    <th>Módulo</th>
                                    <th class="grade-header-th">Nota Calculada</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr
                                    v-for="student in getGroupStudentAverages(
                                        group,
                                    )"
                                    :key="student.id"
                                >
                                    <td class="student-name-cell">
                                        {{ student.name }}
                                    </td>
                                    <td class="score-cell">
                                        {{ student.tareas.toFixed(2) }}
                                    </td>
                                    <td class="score-cell">
                                        {{ student.participacion.toFixed(2) }}
                                    </td>
                                    <td class="score-cell">
                                        {{ student.organizacion.toFixed(2) }}
                                    </td>
                                    <td class="score-cell">
                                        {{ student.compromiso.toFixed(2) }}
                                    </td>
                                    <td class="score-cell">
                                        {{ student.prosocialidad.toFixed(2) }}
                                    </td>
                                    <td class="score-cell avg-cell">
                                        {{ student.promedio.toFixed(2) }}
                                    </td>
                                    <td
                                        class="score-cell"
                                        :class="
                                            student.modulo >= 0
                                                ? 'positive-module'
                                                : 'negative-module'
                                        "
                                    >
                                        {{ student.modulo >= 0 ? "+" : ""
                                        }}{{ student.modulo.toFixed(2) }}
                                    </td>
                                    <!-- Columna de Nota Calculada -->
                                    <td class="score-cell grade-cell">
                                        <div
                                            v-if="hasGroupGrade(group.id)"
                                            class="calculated-grade-badge"
                                            :class="getGradeBadgeClass(calculateGrade(group.id, student.modulo))"
                                            :title="`Nota base grupo: ${getGroupGrade(group.id).toFixed(1)} | Módulo: ${student.modulo >= 0 ? '+' : ''}${student.modulo.toFixed(2)} → Nota: ${calculateGrade(group.id, student.modulo).toFixed(1)}`"
                                        >
                                            {{ calculateGrade(group.id, student.modulo).toFixed(1) }}
                                        </div>
                                        <span
                                            v-else
                                            class="empty-grade-dash"
                                            title="Ingresa la nota grupal arriba para calcular la nota"
                                        >
                                            —
                                        </span>
                                    </td>
                                </tr>
                            </tbody>
                            <tfoot>
                                <tr class="averages-row">
                                    <td><strong>Promedios</strong></td>
                                    <td class="score-cell">
                                        {{
                                            getGroupCategoryAverages(
                                                group,
                                            ).tareas.toFixed(2)
                                        }}
                                    </td>
                                    <td class="score-cell">
                                        {{
                                            getGroupCategoryAverages(
                                                group,
                                            ).participacion.toFixed(2)
                                        }}
                                    </td>
                                    <td class="score-cell">
                                        {{
                                            getGroupCategoryAverages(
                                                group,
                                            ).organizacion.toFixed(2)
                                        }}
                                    </td>
                                    <td class="score-cell">
                                        {{
                                            getGroupCategoryAverages(
                                                group,
                                            ).compromiso.toFixed(2)
                                        }}
                                    </td>
                                    <td class="score-cell">
                                        {{
                                            getGroupCategoryAverages(
                                                group,
                                            ).prosocialidad.toFixed(2)
                                        }}
                                    </td>
                                    <td></td>
                                    <td></td>
                                    <td></td>
                                </tr>
                                <tr class="stddev-row">
                                    <td>
                                        <strong>Desviación estándar</strong>
                                    </td>
                                    <td class="score-cell">
                                        {{
                                            getGroupCategoryStdDev(
                                                group,
                                            ).tareas.toFixed(2)
                                        }}
                                    </td>
                                    <td class="score-cell">
                                        {{
                                            getGroupCategoryStdDev(
                                                group,
                                            ).participacion.toFixed(2)
                                        }}
                                    </td>
                                    <td class="score-cell">
                                        {{
                                            getGroupCategoryStdDev(
                                                group,
                                            ).organizacion.toFixed(2)
                                        }}
                                    </td>
                                    <td class="score-cell">
                                        {{
                                            getGroupCategoryStdDev(
                                                group,
                                            ).compromiso.toFixed(2)
                                        }}
                                    </td>
                                    <td class="score-cell">
                                        {{
                                            getGroupCategoryStdDev(
                                                group,
                                            ).prosocialidad.toFixed(2)
                                        }}
                                    </td>
                                    <td></td>
                                    <td></td>
                                    <td></td>
                                </tr>
                            </tfoot>
                        </table>
                    </div>
                </Transition>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from "vue";
import { useSurveyGroupStats } from "~/composables/useSurveyGroupStats";

const props = defineProps({
    groups: {
        type: Array,
        default: () => [],
    },
    allStudents: {
        type: Array,
        default: () => [],
    },
    allResponses: {
        type: Array,
        default: () => [],
    },
    surveyId: {
        type: String,
        default: "",
    },
});

const allStudentsRef = computed(() => props.allStudents);
const allResponsesRef = computed(() => props.allResponses);

const {
    expandedGroup,
    toggleGroup,
    getStudentName,
    getGroupStudentAverages,
    getGroupStudentAveragesRaw,
    getGroupBaseline,
    getGroupCategoryAverages,
    getGroupCategoryStdDev,
} = useSurveyGroupStats(allStudentsRef, allResponsesRef);

// Notas de grupos: diccionario { [groupId]: number }
const groupGrades = ref({});
const globalGradeInput = ref(null);

// Clave de almacenamiento en localStorage
const storageKey = computed(() => {
    return props.surveyId ? `peer_eval_grades_${props.surveyId}` : "peer_eval_grades_temp";
});

// Cargar notas previamente guardadas
onMounted(() => {
    if (typeof window !== "undefined") {
        try {
            const saved = localStorage.getItem(storageKey.value);
            if (saved) {
                groupGrades.value = JSON.parse(saved);
            }
        } catch (e) {
            console.error("Error loading saved grades from localStorage:", e);
        }
    }
});

// Guardar notas cuando cambien
watch(
    groupGrades,
    (newVal) => {
        if (typeof window !== "undefined") {
            try {
                localStorage.setItem(storageKey.value, JSON.stringify(newVal));
            } catch (e) {
                console.error("Error saving grades to localStorage:", e);
            }
        }
    },
    { deep: true },
);

// Métodos de cálculo de notas
const hasGroupGrade = (groupId) => {
    const val = groupGrades.value[groupId];
    return val !== undefined && val !== null && val !== "" && !isNaN(val) && Number(val) > 0;
};

const getGroupGrade = (groupId) => {
    return Number(groupGrades.value[groupId]) || 0;
};

const calculateGrade = (groupId, modulo) => {
    const base = getGroupGrade(groupId);
    const result = base + Number(modulo);
    // Limitar estrictamente en escala de notas chilena: 1.0 a 7.0
    return Math.min(7.0, Math.max(1.0, result));
};

const getGradeBadgeClass = (grade) => {
    if (grade >= 6.0) return "grade-badge-excellent";
    if (grade >= 5.0) return "grade-badge-good";
    if (grade >= 4.0) return "grade-badge-pass";
    return "grade-badge-fail";
};

const isValidGlobalGrade = computed(() => {
    const val = Number(globalGradeInput.value);
    return !isNaN(val) && val >= 1.0 && val <= 7.0;
});

const applyGlobalGrade = () => {
    if (!isValidGlobalGrade.value) return;
    const grade = Number(globalGradeInput.value);
    props.groups.forEach((g) => {
        groupGrades.value[g.id] = grade;
    });
};
</script>

<style scoped>
.view-content {
    min-height: 200px;
}

.no-groups {
    padding: 3rem;
    text-align: center;
    color: #94a3b8;
    font-style: italic;
}

/* Averages View Styles */
.averages-groups {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.averages-group-section {
    background: white;
    border: 1px solid var(--color-border, #dfe4ea);
    border-radius: 8px;
    overflow: hidden;
}

.averages-group-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.25rem;
    background: var(--color-background, #f4f6f9);
    border-bottom: 1px solid var(--color-border, #dfe4ea);
    cursor: pointer;
    transition: background-color 0.2s ease;
}

.averages-group-header:hover {
    background: #f1f5f9;
}

.averages-header-left {
    display: flex;
    align-items: center;
    gap: 1rem;
}

.averages-header-left h3 {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: #1e293b;
}

.group-students-count {
    background: var(--color-border, #dfe4ea);
    color: #374151;
    padding: 0.25rem 0.625rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 500;
}

.averages-header-right {
    display: flex;
    align-items: center;
    gap: 1rem;
}

.baseline-badge {
    background: var(--color-primary, #006b64);
    color: white;
    padding: 0.375rem 0.75rem;
    border-radius: 9999px;
    font-size: 0.875rem;
    font-weight: 600;
}

.chevron-icon {
    transition: transform 0.3s ease;
    color: #64748b;
}

.chevron-icon.rotated {
    transform: rotate(180deg);
}

.averages-table-container {
    overflow-x: auto;
}

.averages-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
}

.averages-table th,
.averages-table td {
    padding: 0.75rem 1rem;
    text-align: left;
    border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.averages-table th {
    background: #f1f5f9;
    font-weight: 600;
    color: #475569;
    white-space: nowrap;
}

.averages-table tbody tr:hover {
    background: var(--color-background, #f4f6f9);
}

.student-name-cell {
    font-weight: 500;
    color: #1e293b;
    min-width: 200px;
}

.score-cell {
    text-align: center;
    font-family: "SF Mono", Monaco, monospace;
    color: #475569;
}

.avg-cell {
    font-weight: 600;
    color: #1e293b;
    background: #f1f5f9;
}

.positive-module {
    color: #059669;
    font-weight: 600;
}

.negative-module {
    color: #dc2626;
    font-weight: 600;
}

.averages-table tfoot tr {
    background: var(--color-background, #f4f6f9);
}

.averages-row td {
    font-weight: 500;
    color: #334155;
    border-top: 2px solid #cbd5e1;
}

.stddev-row td {
    color: #64748b;
    font-style: italic;
}

/* Grade Calculator Card */
.grade-calculator-card {
    background: #f8fafc;
    border: 1px solid #cbd5e1;
    border-left: 4px solid var(--color-primary, #006b64);
    border-radius: 8px;
    padding: 1rem 1.25rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1.5rem;
    flex-wrap: wrap;
    margin-bottom: 0.5rem;
}

.grade-calculator-info {
    display: flex;
    align-items: flex-start;
    gap: 0.875rem;
    flex: 1;
    min-width: 280px;
}

.info-icon {
    color: var(--color-primary, #006b64);
    margin-top: 0.15rem;
    flex-shrink: 0;
}

.calculator-title {
    margin: 0 0 0.25rem 0;
    font-size: 0.95rem;
    font-weight: 600;
    color: #1e293b;
}

.calculator-desc {
    margin: 0;
    font-size: 0.825rem;
    color: #475569;
    line-height: 1.4;
}

.calculator-desc code {
    background: #e2e8f0;
    padding: 0.15rem 0.35rem;
    border-radius: 4px;
    font-size: 0.8rem;
    font-weight: 600;
    color: #0f172a;
}

.grade-calculator-quick-apply {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    align-items: flex-end;
}

.quick-apply-label {
    font-size: 0.75rem;
    font-weight: 500;
    color: #475569;
}

.quick-apply-inputs {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.global-grade-input {
    width: 75px;
    padding: 0.35rem 0.5rem;
    border: 1px solid #cbd5e1;
    border-radius: 6px;
    font-size: 0.875rem;
    font-weight: 600;
    text-align: center;
    background: white;
    color: #0f172a;
    outline: none;
    transition: border-color 0.2s;
}

.global-grade-input:focus {
    border-color: var(--color-primary, #006b64);
    box-shadow: 0 0 0 2px rgba(0, 107, 100, 0.15);
}

.btn-apply-global {
    background: var(--color-primary, #006b64);
    color: white;
    border: none;
    border-radius: 6px;
    padding: 0.375rem 0.75rem;
    font-size: 0.8rem;
    font-weight: 500;
    cursor: pointer;
    transition: opacity 0.2s, background-color 0.2s;
}

.btn-apply-global:hover:not(:disabled) {
    background: #00554f;
}

.btn-apply-global:disabled {
    opacity: 0.45;
    cursor: not-allowed;
}

/* Group Grade Pill in Header */
.group-grade-pill {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    background: white;
    padding: 0.25rem 0.6rem;
    border-radius: 6px;
    border: 1px solid #cbd5e1;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.group-grade-label {
    font-size: 0.775rem;
    font-weight: 600;
    color: #334155;
    white-space: nowrap;
}

.group-grade-input {
    width: 65px;
    padding: 0.2rem 0.4rem;
    border: 1px solid #cbd5e1;
    border-radius: 4px;
    font-size: 0.85rem;
    font-weight: 700;
    text-align: center;
    background: #f8fafc;
    color: #0f172a;
    outline: none;
    transition: all 0.2s;
}

.group-grade-input:focus {
    background: white;
    border-color: var(--color-primary, #006b64);
    box-shadow: 0 0 0 2px rgba(0, 107, 100, 0.15);
}

/* Table Grade Column */
.grade-header-th {
    background: #e6f4f1 !important;
    color: var(--color-primary, #006b64) !important;
    text-align: center !important;
    font-weight: 700 !important;
}

.grade-cell {
    background: #fafcfb;
    font-family: inherit;
}

.calculated-grade-badge {
    display: inline-block;
    padding: 0.25rem 0.65rem;
    border-radius: 6px;
    font-weight: 700;
    font-size: 0.925rem;
    font-family: "SF Mono", Monaco, monospace;
    letter-spacing: 0.02em;
    cursor: default;
    transition: transform 0.15s ease;
}

.calculated-grade-badge:hover {
    transform: scale(1.05);
}

.grade-badge-excellent {
    background: #dcfce7;
    color: #15803d;
    border: 1px solid #86efac;
}

.grade-badge-good {
    background: #e0f2fe;
    color: #0369a1;
    border: 1px solid #7dd3fc;
}

.grade-badge-pass {
    background: #fef9c3;
    color: #a16207;
    border: 1px solid #fde047;
}

.grade-badge-fail {
    background: #fee2e2;
    color: #b91c1c;
    border: 1px solid #fca5a5;
}

.empty-grade-dash {
    color: #94a3b8;
    font-weight: 500;
    font-size: 1.1rem;
}

/* Accordion Slide Animation */
.accordion-slide-enter-active {
    animation: accordion-open 0.3s ease-out;
}

.accordion-slide-leave-active {
    animation: accordion-close 0.25s ease-in;
}

@keyframes accordion-open {
    0% {
        opacity: 0;
        max-height: 0;
    }
    100% {
        opacity: 1;
        max-height: 800px;
    }
}

@keyframes accordion-close {
    0% {
        opacity: 1;
        max-height: 800px;
    }
    100% {
        opacity: 0;
        max-height: 0;
    }
}

/* Responsive Styles */
@media (max-width: 640px) {
    .averages-group-header {
        flex-direction: column;
        align-items: flex-start;
        gap: 0.75rem;
        padding: 0.875rem 1rem;
    }

    .averages-header-left {
        flex-direction: column;
        align-items: flex-start;
        gap: 0.5rem;
        width: 100%;
    }

    .averages-header-left h3 {
        font-size: 0.9375rem;
    }

    .averages-header-right {
        width: 100%;
        justify-content: space-between;
        gap: 0.75rem;
        flex-wrap: wrap;
    }

    .baseline-badge {
        font-size: 0.75rem;
        padding: 0.25rem 0.625rem;
    }

    .group-students-count {
        font-size: 0.7rem;
        padding: 0.2rem 0.5rem;
    }

    .averages-table th,
    .averages-table td {
        padding: 0.5rem 0.625rem;
        font-size: 0.75rem;
    }

    .student-name-cell {
        min-width: 120px;
    }
}

@media (max-width: 480px) {
    .averages-group-header {
        padding: 0.75rem;
    }

    .averages-header-left h3 {
        font-size: 0.875rem;
    }

    .baseline-badge {
        font-size: 0.7rem;
        padding: 0.2rem 0.5rem;
    }

    .chevron-icon {
        width: 18px;
        height: 18px;
    }
}
</style>
