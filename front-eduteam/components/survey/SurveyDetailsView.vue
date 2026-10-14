<template>
    <div class="view-content">
        <div class="groups-accordion">
            <div v-if="groups.length === 0" class="no-groups">
                <p>
                    No hay grupos definidos para esta sección. Mostrando todos
                    los estudiantes.
                </p>

                <!-- Fallback: Show all students if no groups -->
                <table class="students-table">
                    <thead>
                        <tr>
                            <th>Estudiante</th>
                            <th>Email</th>
                            <th>Estado</th>
                            <th>Última evaluación</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr
                            v-for="student in filteredStudents"
                            :key="student.id"
                            :class="{ 'completed-row': student.has_evaluated }"
                        >
                            <td class="student-name">{{ student.name }}</td>
                            <td class="student-email">{{ student.email }}</td>
                            <td class="status-cell">
                                <div class="status-badges">
                                    <span
                                        v-if="student.has_evaluated"
                                        class="status-badge completed"
                                    >
                                        <svg
                                            width="16"
                                            height="16"
                                            viewBox="0 0 24 24"
                                            fill="none"
                                            stroke="currentColor"
                                            stroke-width="2"
                                        >
                                            <polyline points="20 6 9 17 4 12" />
                                        </svg>
                                        {{ surveyPlugin === 'group_compatibility' ? 'Respondido' : 'Ha evaluado' }}
                                    </span>
                                    <span
                                        v-if="student.has_been_evaluated && surveyPlugin !== 'group_compatibility'"
                                        class="status-badge evaluated"
                                    >
                                        <svg
                                            width="16"
                                            height="16"
                                            viewBox="0 0 24 24"
                                            fill="none"
                                            stroke="currentColor"
                                            stroke-width="2"
                                        >
                                            <circle cx="12" cy="12" r="10" />
                                            <line
                                                x1="12"
                                                y1="8"
                                                x2="12"
                                                y2="12"
                                            />
                                        </svg>
                                        Ha sido evaluado
                                    </span>
                                    <span
                                        v-if="
                                            !student.has_evaluated &&
                                            (surveyPlugin === 'group_compatibility' || !student.has_been_evaluated)
                                        "
                                        class="status-badge pending"
                                    >
                                        <svg
                                            width="16"
                                            height="16"
                                            viewBox="0 0 24 24"
                                            fill="none"
                                            stroke="currentColor"
                                            stroke-width="2"
                                        >
                                            <circle cx="12" cy="12" r="10" />
                                            <line
                                                x1="12"
                                                y1="8"
                                                x2="12"
                                                y2="12"
                                            />
                                        </svg>
                                        {{ surveyPlugin === 'group_compatibility' ? 'Pendiente' : 'Sin participación' }}
                                    </span>
                                </div>
                            </td>
                            <td class="date-cell">
                                {{
                                    student.last_evaluation
                                        ? formatDate(student.last_evaluation)
                                        : "-"
                                }}
                            </td>
                        </tr>
                    </tbody>
                </table>

                <div v-if="filteredStudents.length === 0" class="no-data">
                    No hay estudiantes
                    {{
                        filter === "evaluators"
                            ? "que hayan evaluado"
                            : filter === "evaluated"
                              ? "que hayan sido evaluados"
                              : ""
                    }}
                </div>
            </div>

            <div v-for="group in groups" :key="group.id" class="group-section">
                <div class="group-header" @click="toggleGroup(group.id)">
                    <div class="group-info">
                        <h3 class="group-name">{{ group.name }}</h3>
                        <span v-if="surveyPlugin === 'group_compatibility'" class="group-stats">
                            {{ group.students.length }} estudiantes •
                            {{ getEvaluatorsCount(group) }} han respondido
                        </span>
                        <span v-else class="group-stats">
                            {{ group.students.length }} estudiantes •
                            {{ getEvaluatorsCount(group) }} han evaluado •
                            {{ getEvaluatedCount(group) }} han sido evaluados
                        </span>
                    </div>
                    <div class="group-toggle">
                        <svg
                            :class="{ rotated: expandedGroups.has(group.id) }"
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
                        v-if="expandedGroups.has(group.id)"
                        class="group-content"
                    >
                        <!-- Filter Options for this group -->
                        <div class="filter-section">
                            <label class="filter-label">
                                <input
                                    type="radio"
                                    :name="`filter-${group.id}`"
                                    :value="'all'"
                                    v-model="localFilter"
                                />
                                <span>Todos ({{ group.students.length }})</span>
                            </label>
                            <label class="filter-label">
                                <input
                                    type="radio"
                                    :name="`filter-${group.id}`"
                                    :value="'evaluators'"
                                    v-model="localFilter"
                                />
                                <span v-if="surveyPlugin === 'group_compatibility'"
                                    >Han respondido ({{
                                        getEvaluatorsCount(group)
                                    }})</span
                                >
                                <span v-else
                                    >Han evaluado ({{
                                        getEvaluatorsCount(group)
                                    }})</span
                                >
                            </label>
                            <label v-if="surveyPlugin !== 'group_compatibility'" class="filter-label">
                                <input
                                    type="radio"
                                    :name="`filter-${group.id}`"
                                    :value="'evaluated'"
                                    v-model="localFilter"
                                />
                                <span
                                    >Han sido evaluados ({{
                                        getEvaluatedCount(group)
                                    }})</span
                                >
                            </label>
                        </div>

                        <!-- Students Table for this group -->
                        <div class="table-container">
                            <table class="students-table">
                                <thead>
                                    <tr>
                                        <th>Estudiante</th>
                                        <th>Email</th>
                                        <th>Preguntas respondidas</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr
                                        v-for="student in getFilteredStudentsForGroup(
                                            group,
                                        )"
                                        :key="student.id"
                                        :class="{
                                            'completed-row':
                                                student.has_evaluated,
                                        }"
                                    >
                                        <td class="student-name">
                                            {{ student.name }}
                                        </td>
                                        <td class="student-email">
                                            {{ student.email }}
                                        </td>
                                        <td class="responses-cell">
                                            <button
                                                class="view-responses-btn"
                                                @click="
                                                    $emit(
                                                        'openResponses',
                                                        student.id,
                                                        group,
                                                    )
                                                "
                                                :title="`Ver respuestas de ${student.name}`"
                                            >
                                                {{
                                                    getStudentResponseCount(
                                                        student.id,
                                                        group,
                                                    )
                                                }}
                                            </button>
                                        </td>
                                    </tr>
                                </tbody>
                            </table>

                            <div
                                v-if="
                                    getFilteredStudentsForGroup(group)
                                        .length === 0
                                "
                                class="no-data"
                            >
                                No hay estudiantes
                                {{
                                    localFilter === "evaluators"
                                        ? "que hayan evaluado"
                                        : localFilter === "evaluated"
                                          ? "que hayan sido evaluados"
                                          : ""
                                }}
                            </div>
                        </div>
                    </div>
                </Transition>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, computed } from "vue";
import { useDatetimeInput } from "~/composables/useDatetimeInput";

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
    results: {
        type: Array,
        default: () => [],
    },
    studentsData: {
        type: Array,
        default: () => [],
    },
    filter: {
        type: String,
        default: "all",
    },
    surveyPlugin: {
        type: String,
        default: "",
    },
});

const emit = defineEmits(["openResponses"]);

const expandedGroups = ref(new Set());
const localFilter = ref(props.filter);

// Computed properties
const filteredStudents = computed(() => {
    const evaluatorIds = new Set(
        props.allResponses.map((r) => String(r.evaluator_id)),
    );
    const evaluatedIds = new Set(
        props.allResponses.map((r) => String(r.evaluated_id)),
    );

    let filtered = props.allStudents.map((student) => ({
        id: student.id,
        name:
            `${student.nombres || ""} ${student.paterno || ""} ${student.materno || ""}`.trim() ||
            `ID: ${student.id}`,
        email: student.email || "",
        has_evaluated: evaluatorIds.has(String(student.id)),
        has_been_evaluated: evaluatedIds.has(String(student.id)),
    }));

    if (localFilter.value === "evaluators") {
        return filtered.filter((s) => s.has_evaluated);
    } else if (localFilter.value === "evaluated" && props.surveyPlugin !== "group_compatibility") {
        return filtered.filter((s) => s.has_been_evaluated);
    }
    return filtered;
});

// Methods
const toggleGroup = (groupId) => {
    if (expandedGroups.value.has(groupId)) {
        expandedGroups.value.delete(groupId);
    } else {
        expandedGroups.value.clear();
        expandedGroups.value.add(groupId);
    }
};

const getEvaluatorsCount = (group) => {
    return group.students.filter(
        (s) =>
            props.studentsData.find((sd) => String(sd.id) === String(s.id))
                ?.has_evaluated,
    ).length;
};

const getEvaluatedCount = (group) => {
    return group.students.filter(
        (s) =>
            props.studentsData.find((sd) => String(sd.id) === String(s.id))
                ?.has_been_evaluated,
    ).length;
};

const getFilteredStudentsForGroup = (group) => {
    const evaluatorIds = new Set(
        props.allResponses.map((r) => String(r.evaluator_id)),
    );
    const evaluatedIds = new Set(
        props.allResponses.map((r) => String(r.evaluated_id)),
    );

    const groupStudentIds = new Set(group.students.map((s) => String(s.id)));
    let filtered = props.allStudents
        .filter((s) => groupStudentIds.has(String(s.id)))
        .map((student) => ({
            id: student.id,
            name:
                `${student.nombres || ""} ${student.paterno || ""} ${student.materno || ""}`.trim() ||
                `ID: ${student.id}`,
            email: student.email || "",
            has_evaluated: evaluatorIds.has(String(student.id)),
            has_been_evaluated: evaluatedIds.has(String(student.id)),
        }));

    if (localFilter.value === "evaluators") {
        filtered = filtered.filter((s) => s.has_evaluated);
    } else if (localFilter.value === "evaluated" && props.surveyPlugin !== "group_compatibility") {
        filtered = filtered.filter((s) => s.has_been_evaluated);
    }

    return filtered;
};

const getStudentResponseCount = (studentId, group) => {
    if (props.surveyPlugin === "group_compatibility") {
        const hasSubmitted = props.allResponses.some(
            (r) => String(r.evaluator_id) === String(studentId)
        );
        return hasSubmitted ? "Respondido" : "Pendiente";
    }

    const evaluatorResponses = props.allResponses.filter(
        (r) =>
            String(r.evaluator_id) === String(studentId) &&
            group.students.some((s) => String(s.id) === String(r.evaluated_id)),
    );

    const answeredCount = evaluatorResponses.length * 5;
    const totalStudentsInGroup = group.students.length;
    const totalPossible = totalStudentsInGroup * 5;

    return `${answeredCount}/${totalPossible}`;
};

const { formatDatetimeDisplay } = useDatetimeInput();
const formatDate = (dateString) => formatDatetimeDisplay(dateString);
</script>

<style scoped>
.view-content {
    min-height: 200px;
}

/* Filter Section */
.filter-section {
    display: flex;
    gap: 1.5rem;
    margin-bottom: 1.5rem;
    padding: 1rem;
    background: var(--color-background, #f4f6f9);
    border-radius: 8px;
}

.filter-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    font-size: 0.875rem;
    color: var(--color-dark, #394049);
}

.filter-label input[type="radio"] {
    cursor: pointer;
}

/* Table */
.table-container {
    border: 1px solid var(--color-border, #dfe4ea);
    border-radius: 8px;
    overflow: hidden;
}

.students-table {
    width: 100%;
    border-collapse: collapse;
}

.students-table thead {
    background: var(--color-background, #f4f6f9);
}

.students-table th {
    padding: 0.75rem 1rem;
    text-align: left;
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--color-text-secondary, #64748b);
    text-transform: uppercase;
    border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.students-table td {
    padding: 0.75rem 1rem;
    font-size: 0.875rem;
    color: var(--color-dark, #394049);
    border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.students-table tbody tr:last-child td {
    border-bottom: none;
}

.students-table tbody tr:hover {
    background: var(--color-background, #f4f6f9);
}

.completed-row {
    background: #f0fdf4 !important;
}

.student-name {
    font-weight: 500;
}

.student-email {
    color: #94a3b8;
}

.responses-cell {
    font-weight: 600;
    color: #00a499;
    text-align: center;
}

.status-badges {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
}

.status-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem 0.75rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 500;
}

.status-badge.completed {
    background: #d1fae5;
    color: #065f46;
}

.status-badge.evaluated {
    background: #dbeafe;
    color: #1e40af;
}

.status-badge.pending {
    background: #fed7aa;
    color: #92400e;
}

.date-cell {
    color: #94a3b8;
}

.no-data {
    padding: 3rem;
    text-align: center;
    color: #94a3b8;
    font-style: italic;
}

/* Groups Accordion */
.groups-accordion {
    border: 1px solid var(--color-border, #dfe4ea);
    border-radius: 8px;
    overflow: hidden;
}

.group-section {
    border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.group-section:last-child {
    border-bottom: none;
}

.group-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.5rem;
    background: var(--color-background, #f4f6f9);
    cursor: pointer;
    transition: background-color 0.2s;
}

.group-header:hover {
    background: #f1f5f9;
}

.group-info {
    flex: 1;
}

.group-name {
    margin: 0 0 0.25rem 0;
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--color-dark, #394049);
}

.group-stats {
    font-size: 0.875rem;
    color: #94a3b8;
}

.group-toggle {
    margin-left: 1rem;
    transition: transform 0.2s;
}

.group-toggle svg.rotated {
    transform: rotate(180deg);
}

.group-content {
    padding: 1.5rem;
    background: white;
    border-top: 1px solid var(--color-border, #dfe4ea);
    overflow: hidden;
}

.no-groups {
    padding: 3rem;
    text-align: center;
    color: #94a3b8;
    font-style: italic;
}

/* View Responses Button */
.view-responses-btn {
    background: none;
    border: none;
    color: var(--color-primary, #00a499);
    cursor: pointer;
    font-weight: 600;
    text-decoration: underline;
    padding: 0;
    font-size: 0.875rem;
    transition: color 0.2s;
}

.view-responses-btn:hover {
    color: var(--color-primary-dark, #008a7f);
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
        padding-top: 0;
        padding-bottom: 0;
    }
    100% {
        opacity: 1;
        max-height: 500px;
        padding-top: 1.5rem;
        padding-bottom: 1.5rem;
    }
}

@keyframes accordion-close {
    0% {
        opacity: 1;
        max-height: 500px;
        padding-top: 1.5rem;
        padding-bottom: 1.5rem;
    }
    100% {
        opacity: 0;
        max-height: 0;
        padding-top: 0;
        padding-bottom: 0;
    }
}

/* Responsive */
@media (max-width: 768px) {
    .filter-section {
        flex-direction: column;
        gap: 0.75rem;
    }

    .table-container {
        overflow-x: auto;
    }

    .students-table {
        min-width: 600px;
    }
}
</style>
