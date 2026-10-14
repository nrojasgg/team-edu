<template>
    <div v-if="isOpen" class="modal-overlay" @click.self="closeModal">
        <div class="modal-content">
            <div class="modal-header">
                <div>
                    <h2>{{ survey?.name || "Evaluación entre Pares" }}</h2>
                    <p class="modal-subtitle">Resultados de la evaluación</p>
                </div>
                <button @click="closeModal" class="close-btn">
                    <svg
                        width="24"
                        height="24"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                    >
                        <line x1="18" y1="6" x2="6" y2="18" />
                        <line x1="6" y1="6" x2="18" y2="18" />
                    </svg>
                </button>
            </div>

            <div class="modal-body">
                <div v-if="loading" class="loading-state">
                    <div class="spinner"></div>
                    <p>Cargando resultados...</p>
                </div>

                <div v-else-if="error" class="error-state">
                    <svg
                        width="48"
                        height="48"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                    >
                        <circle cx="12" cy="12" r="10" />
                        <line x1="12" y1="8" x2="12" y2="12" />
                        <line x1="12" y1="16" x2="12.01" y2="16" />
                    </svg>
                    <p>{{ error }}</p>
                </div>

                <div
                    v-else-if="!results || results.length === 0"
                    class="no-results"
                >
                    <svg
                        width="48"
                        height="48"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                    >
                        <path
                            d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                        />
                    </svg>
                    <p>No hay resultados disponibles para esta evaluación</p>
                </div>

                <div v-else class="results-content">
                    <!-- View Mode Tabs -->
                    <div class="view-mode-tabs">
                        <button
                            class="view-mode-tab tab-details"
                            :class="{ active: viewMode === 'details' }"
                            @click="viewMode = 'details'"
                        >
                            <svg
                                width="18"
                                height="18"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                            >
                                <path
                                    d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"
                                />
                                <polyline points="14 2 14 8 20 8" />
                                <line x1="16" y1="13" x2="8" y2="13" />
                                <line x1="16" y1="17" x2="8" y2="17" />
                                <polyline points="10 9 9 9 8 9" />
                            </svg>
                            Detalles
                        </button>
                        <button
                            v-if="props.survey?.plugin !== 'group_compatibility'"
                            class="view-mode-tab tab-averages"
                            :class="{ active: viewMode === 'averages' }"
                            @click="viewMode = 'averages'"
                        >
                            <svg
                                width="18"
                                height="18"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                            >
                                <line x1="18" y1="20" x2="18" y2="10" />
                                <line x1="12" y1="20" x2="12" y2="4" />
                                <line x1="6" y1="20" x2="6" y2="14" />
                            </svg>
                            Promedios
                        </button>
                        <button
                            v-if="props.survey?.plugin !== 'group_compatibility'"
                            class="view-mode-tab"
                            :class="{ active: viewMode === 'visualization' }"
                            @click="viewMode = 'visualization'"
                        >
                            <svg
                                width="18"
                                height="18"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                            >
                                <circle cx="12" cy="12" r="10" />
                                <path d="M12 2a10 10 0 0 1 10 10" />
                                <path d="M12 12l5-5" />
                            </svg>
                            Visualización
                        </button>
                    </div>

                    <!-- Details View -->
                    <SurveyDetailsView
                        v-if="viewMode === 'details'"
                        :groups="groups"
                        :all-students="allStudentsArray"
                        :all-responses="allResponses"
                        :results="results"
                        :students-data="studentsData"
                        :filter="filter"
                        :survey-plugin="props.survey?.plugin"
                        @open-responses="openResponsesModal"
                    />

                    <!-- Averages View -->
                    <SurveyAveragesView
                        v-if="viewMode === 'averages'"
                        :groups="groups"
                        :all-students="allStudentsArray"
                        :all-responses="allResponses"
                        :survey-id="props.survey?.id"
                    />

                    <!-- Visualization View -->
                    <SurveyVisualizationView
                        v-if="viewMode === 'visualization'"
                        :groups="groups"
                        :all-students="allStudentsArray"
                        :all-responses="allResponses"
                    />
                </div>
            </div>

            <!-- Modal Footer -->
            <div class="modal-footer">
                <button
                    v-if="props.survey?.plugin !== 'group_compatibility'"
                    @click="exportToExcel"
                    class="export-btn"
                    :disabled="loading || exportingExcel"
                >
                    <svg
                        v-if="!exportingExcel"
                        width="18"
                        height="18"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                    >
                        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                        <polyline points="7 10 12 15 17 10" />
                        <line x1="12" y1="15" x2="12" y2="3" />
                    </svg>
                    <div v-else class="spinner-small"></div>
                    {{
                        exportingExcel
                            ? "Exportando..."
                            : "Exportar Detalle Excel"
                    }}
                </button>
            </div>
        </div>
    </div>

    <!-- Responses Detail Modal -->
    <div
        v-if="showResponsesModal"
        class="modal-overlay"
        @click.self="closeResponsesModal"
    >
        <div class="modal-content responses-modal">
            <div class="modal-header">
                <div>
                    <h2>Respuestas de {{ selectedStudentName }}</h2>
                    <p class="modal-subtitle">Evaluaciones del grupo</p>
                </div>
                <button @click="closeResponsesModal" class="close-btn">
                    <svg
                        width="24"
                        height="24"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                    >
                        <line x1="18" y1="6" x2="6" y2="18" />
                        <line x1="6" y1="6" x2="18" y2="18" />
                    </svg>
                </button>
            </div>

            <div class="modal-body">
                <div v-if="studentResponses.length === 0" class="no-results">
                    <p>No hay respuestas disponibles</p>
                </div>

                <div v-else class="responses-accordion">
                    <div
                        v-for="response in studentResponses"
                        :key="response._id || response.id"
                        class="response-accordion-item"
                    >
                        <div
                            class="response-accordion-header"
                            @click="
                                toggleResponseAccordion(
                                    response._id || response.id,
                                )
                            "
                        >
                            <div class="accordion-student-info">
                                <h3 class="evaluated-student-name">
                                    {{ response.is_sociogram ? response.question_text : getStudentName(response.evaluated_id) }}
                                </h3>
                                <span class="response-date">{{
                                    formatDate(
                                        response.submitted_at ||
                                            response.created_at,
                                    )
                                }}</span>
                            </div>
                            <div class="accordion-toggle">
                                <span class="response-summary">{{
                                    response.is_sociogram ? (response.selected_students && response.selected_students.length > 0 ? `${response.selected_students.length} seleccionados` : 'Ninguno seleccionado') : getResponseSummary(response)
                                }}</span>
                                <svg
                                    :class="{
                                        rotated:
                                            expandedResponseId ===
                                            (response._id || response.id),
                                    }"
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
                                v-if="
                                    expandedResponseId ===
                                    (response._id || response.id)
                                "
                                class="response-accordion-content"
                            >
                                <div
                                    v-if="response.is_sociogram"
                                    class="questions-container"
                                >
                                    <div class="question-response">
                                        <div class="question-header">
                                            <span class="question-category">Estudiantes seleccionados:</span>
                                        </div>
                                        <div class="question-text" style="font-weight: 500; font-size: 1.05rem; padding: 0.5rem 0;">
                                            <div v-if="response.selected_students && response.selected_students.length > 0" class="selected-students-list">
                                                <ul style="padding-left: 1.2rem; margin: 0.5rem 0;">
                                                    <li v-for="name in response.selected_students" :key="name" style="margin-bottom: 0.25rem;">
                                                        {{ name }}
                                                    </li>
                                                </ul>
                                            </div>
                                            <span v-else class="text-muted" style="font-style: italic;">Ningún estudiante seleccionado</span>
                                        </div>
                                    </div>
                                </div>
                                <div
                                    v-else-if="
                                        response.responses &&
                                        response.responses.length > 0
                                    "
                                    class="questions-container"
                                >
                                    <div
                                        v-for="(
                                            question, index
                                        ) in response.responses"
                                        :key="index"
                                        class="question-response"
                                    >
                                        <div class="question-header">
                                            <span class="question-category">{{
                                                getQuestionCategory(
                                                    question.question_id,
                                                )
                                            }}</span>
                                            <span
                                                class="answer-value score-badge"
                                                :class="
                                                    getScoreClass(
                                                        question.score,
                                                    )
                                                "
                                                >{{ question.score }}/4</span
                                            >
                                        </div>
                                        <div class="question-text">
                                            {{
                                                getQuestionText(
                                                    question.question_id,
                                                )
                                            }}
                                        </div>
                                    </div>
                                </div>
                                <div v-else class="no-questions">
                                    <p>Sin detalles de respuestas</p>
                                </div>
                            </div>
                        </Transition>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from "vue";
import { useDatetimeInput } from "~/composables/useDatetimeInput";

import { useSectionGroups } from "~/composables/useSectionGroups";
import SurveyDetailsView from "~/components/survey/SurveyDetailsView.vue";
import SurveyAveragesView from "~/components/survey/SurveyAveragesView.vue";
import SurveyVisualizationView from "~/components/survey/SurveyVisualizationView.vue";

const props = defineProps({
    isOpen: {
        type: Boolean,
        default: false,
    },
    survey: {
        type: Object,
        default: null,
    },
    sectionId: {
        type: String,
        required: true,
    },
    allStudents: {
        type: Object,
        default: () => ref([]),
    },
});

const emit = defineEmits(["close"]);

const loading = ref(false);
const error = ref("");
const results = ref([]);
const allResponses = ref([]);
const filter = ref("all");
const groups = ref([]);
const showResponsesModal = ref(false);
const selectedStudentId = ref(null);
const selectedStudentName = ref("");
const studentResponses = ref([]);
const expandedResponseId = ref(null);
const viewMode = ref("details");
const exportingExcel = ref(false);

// Predefined questions for peer evaluation
const peerEvaluationQuestions = [
    {
        id: 1,
        text: "Realiza las tareas que le son asignadas dentro del grupo en los plazos requeridos.",
        category: "Tareas",
    },
    {
        id: 2,
        text: "Participa de forma activa en los espacios de encuentro del equipo, compartiendo la información, conocimientos y las experiencias.",
        category: "Participación",
    },
    {
        id: 3,
        text: "Colabora en la definición, organización y distribución de las tareas del equipo.",
        category: "Organización",
    },
    {
        id: 4,
        text: "Se orienta a la consecución de acuerdos y objetivos comunes y se compromete con ellos.",
        category: "Compromiso",
    },
    {
        id: 5,
        text: "Toma en cuenta los puntos de vista de los demás y retroalimenta constructivamente.",
        category: "Prosocialidad",
    },
];

// Initialize composables
const groupsModule = useSectionGroups();

// Computed properties
const allStudentsArray = computed(() => {
    return Array.isArray(props.allStudents)
        ? props.allStudents
        : props.allStudents?.value || [];
});

const studentsData = computed(() => {
    const evaluatorIds = new Set(
        allResponses.value.map((r) => String(r.evaluator_id)),
    );
    const evaluatedIds = new Set(
        allResponses.value.map((r) => String(r.evaluated_id)),
    );

    return allStudentsArray.value.map((student) => {
        const sid = String(student.id);
        const hasEvaluated = evaluatorIds.has(sid);
        const hasBeenEvaluated = evaluatedIds.has(sid);

        const studentEvaluations = allResponses.value.filter(
            (r) => String(r.evaluator_id) === sid,
        );
        const lastEvaluation =
            studentEvaluations.length > 0
                ? studentEvaluations.sort(
                      (a, b) => new Date(b.created_at) - new Date(a.created_at),
                  )[0].created_at
                : null;

        return {
            id: student.id,
            name:
                `${student.nombres || ""} ${student.paterno || ""} ${student.materno || ""}`.trim() ||
                `ID: ${student.id}`,
            email: student.email || "",
            has_evaluated: hasEvaluated,
            has_been_evaluated: hasBeenEvaluated,
            last_evaluation: lastEvaluation,
        };
    });
});

// Resolución reactiva de nombres de estudiantes
const getStudentName = (studentId) => {
    const student = allStudentsArray.value.find(
        (s) => String(s.id) === String(studentId),
    );
    if (student) {
        return `${student.nombres || ""} ${student.paterno || ""} ${student.materno || ""}`.trim();
    }
    return `ID: ${studentId}`;
};

const getScoreClass = (score) => {
    if (score >= 4) return "score-excellent";
    if (score >= 3) return "score-good";
    if (score >= 2) return "score-average";
    return "score-poor";
};

const getQuestionText = (questionId) => {
    const question = peerEvaluationQuestions.find((q) => q.id === questionId);
    return question ? question.text : `Pregunta ${questionId}`;
};

const getQuestionCategory = (questionId) => {
    const question = peerEvaluationQuestions.find((q) => q.id === questionId);
    return question ? question.category : "";
};

const { formatDatetimeDisplay } = useDatetimeInput();
const formatDate = (dateString) => formatDatetimeDisplay(dateString);

const closeModal = () => {
    emit("close");
};

// Toggle response accordion
const toggleResponseAccordion = (responseId) => {
    if (expandedResponseId.value === responseId) {
        expandedResponseId.value = null;
    } else {
        expandedResponseId.value = responseId;
    }
};

// Get response summary
const getResponseSummary = (response) => {
    if (!response.responses || response.responses.length === 0) return "";
    const total = response.responses.reduce(
        (sum, q) => sum + (q.score || 0),
        0,
    );
    const avg = (total / response.responses.length).toFixed(1);
    return `Promedio: ${avg}/4`;
};

// Open responses modal for a student
const openResponsesModal = (studentId, group) => {
    selectedStudentId.value = studentId;
    selectedStudentName.value = getStudentName(studentId) || `ID: ${studentId}`;

    const responses = allResponses.value.filter(
        (r) => String(r.evaluator_id) === String(studentId),
    );
    const groupStudentIds = new Set(group.students.map((s) => String(s.id)));
    const filtered = responses.filter((r) =>
        r.is_sociogram || groupStudentIds.has(String(r.evaluated_id)),
    );

    studentResponses.value = filtered;
    showResponsesModal.value = true;
};

// Close responses modal
const closeResponsesModal = () => {
    showResponsesModal.value = false;
    selectedStudentId.value = null;
    selectedStudentName.value = "";
    studentResponses.value = [];
    expandedResponseId.value = null;
};

// Load results
onMounted(async () => {
    if (props.isOpen && props.survey) {
        await loadResults();
    }
});

watch(
    () => props.isOpen,
    async (newVal) => {
        if (newVal && props.survey) {
            await loadResults();
        }
    },
);

const loadResults = async () => {
    if (!props.survey?.id) {
        error.value = "No se pudo cargar los resultados: evaluación no válida";
        loading.value = false;
        return;
    }

    try {
        loading.value = true;
        error.value = "";

        const { useAuth } = await import("~/composables/useAuth");
        const { authenticatedFetch } = useAuth();

        const API_BASE_URL = "http://127.0.0.1:3000/api/v1";

        let summaryUrl = "";
        let responsesUrl = "";

        if (props.survey?.plugin === "peer_evaluation") {
            summaryUrl = `${API_BASE_URL}/core/peer-evaluation/surveys/${props.survey.id}/results`;
            responsesUrl = `${API_BASE_URL}/core/peer-evaluation/surveys/${props.survey.id}/responses`;
        } else if (props.survey?.plugin === "group_compatibility") {
            responsesUrl = `${API_BASE_URL}/core/group-compatibility/surveys/${props.survey.id}/responses`;
        } else {
            // Fallback for backward compatibility/other plugins
            summaryUrl = `${API_BASE_URL}/sections/${props.sectionId}/responses/summary`;
            responsesUrl = `${API_BASE_URL}/sections/${props.sectionId}/responses`;
        }

        const promises = [];
        if (summaryUrl) {
            promises.push(authenticatedFetch(summaryUrl));
        } else {
            promises.push(Promise.resolve(null));
        }
        promises.push(authenticatedFetch(responsesUrl));

        const [summaryResponse, responsesResponse] = await Promise.all(promises);

        if ((summaryUrl && !summaryResponse) || !responsesResponse) {
            throw new Error("No autorizado para acceder a los resultados");
        }

        if ((summaryUrl && !summaryResponse.ok) || !responsesResponse.ok) {
            throw new Error("Error al obtener los resultados");
        }

        const summaryData = summaryUrl ? await summaryResponse.json() : null;
        const responsesData = await responsesResponse.json();

        if (props.survey?.plugin === "group_compatibility") {
            const rawResponses = responsesData && responsesData.data ? (Array.isArray(responsesData.data) ? responsesData.data : [responsesData.data]) : [];
            const mapped = [];

            for (const resp of rawResponses) {
                const sId = resp.student_id;
                const submittedAt = resp.submitted_at || resp.created_at;

                // Get question texts from props.survey.questions or default
                const getQText = (qId, defaultText) => {
                    const q = (props.survey?.questions || []).find(x => x.id === qId);
                    return q ? q.text : defaultText;
                };

                const q1Text = getQText(1, "¿A quién(es) elegirías para trabajar en equipo?");
                const q2Text = getQText(2, "¿Quién(es) crees que te elegirían para trabajar en equipo?");
                const q3Text = getQText(3, "¿A quién(es) NO elegirías para trabajar en equipo?");
                const q4Text = getQText(4, "¿Quién(es) crees que NO te elegirían para trabajar en equipo?");

                // Map chosen student names
                const getNames = (answers) => {
                    return (answers || [])
                        .map(ans => getStudentName(ans.student_id))
                        .filter(Boolean);
                };

                mapped.push({
                    id: `${resp.id || resp._id}_q1`,
                    _id: `${resp.id || resp._id}_q1`,
                    evaluator_id: sId,
                    evaluated_id: sId,
                    is_sociogram: true,
                    question_text: q1Text,
                    selected_students: getNames(resp.question1),
                    submitted_at: submittedAt,
                    created_at: resp.created_at
                });

                mapped.push({
                    id: `${resp.id || resp._id}_q2`,
                    _id: `${resp.id || resp._id}_q2`,
                    evaluator_id: sId,
                    evaluated_id: sId,
                    is_sociogram: true,
                    question_text: q2Text,
                    selected_students: getNames(resp.question2),
                    submitted_at: submittedAt,
                    created_at: resp.created_at
                });

                mapped.push({
                    id: `${resp.id || resp._id}_q3`,
                    _id: `${resp.id || resp._id}_q3`,
                    evaluator_id: sId,
                    evaluated_id: sId,
                    is_sociogram: true,
                    question_text: q3Text,
                    selected_students: getNames(resp.question3),
                    submitted_at: submittedAt,
                    created_at: resp.created_at
                });

                mapped.push({
                    id: `${resp.id || resp._id}_q4`,
                    _id: `${resp.id || resp._id}_q4`,
                    evaluator_id: sId,
                    evaluated_id: sId,
                    is_sociogram: true,
                    question_text: q4Text,
                    selected_students: getNames(resp.question4),
                    submitted_at: submittedAt,
                    created_at: resp.created_at
                });
            }

            allResponses.value = mapped;
            results.value = rawResponses;
        } else {
            if (summaryData && summaryData.data) {
                results.value = Array.isArray(summaryData.data)
                    ? summaryData.data
                    : [summaryData.data];
            } else {
                results.value = [];
            }

            if (responsesData && responsesData.data) {
                allResponses.value = Array.isArray(responsesData.data)
                    ? responsesData.data
                    : [responsesData.data];
            } else {
                allResponses.value = [];
            }
        }

        // Construir grupos desde group_snapshots del survey (tienen los group_id reales de MongoDB)
        if (
            props.survey?.group_snapshots &&
            props.survey.group_snapshots.length > 0
        ) {
            groups.value = props.survey.group_snapshots.map((snap) => ({
                id: snap.group_id,
                name: snap.group_name,
                number: snap.group_number,
                students: (snap.student_ids || []).map((sid) => ({ id: sid })),
            }));
        } else {
            try {
                const groupPromise = groupsModule.loadGroups(
                    props.sectionId,
                    localStorage.getItem("auth_token"),
                    allStudentsArray.value,
                );
                const timeoutPromise = new Promise((_, reject) =>
                    setTimeout(
                        () => reject(new Error("Groups loading timeout")),
                        5000,
                    ),
                );
                await Promise.race([groupPromise, timeoutPromise]);
                groups.value = groupsModule.groups.value;
            } catch (groupError) {
                console.warn("Failed to load groups:", groupError.message);
                groups.value = [];
            }
        }
    } catch (err) {
        console.error("Error loading survey results:", err);
        error.value = err.message || "Error al cargar los resultados.";
        results.value = [];
        allResponses.value = [];
        groups.value = [];
    } finally {
        loading.value = false;
    }
};

// Export to Excel function
const exportToExcel = async () => {
    if (!props.survey?.id) {
        console.error("No survey ID available");
        return;
    }

    try {
        exportingExcel.value = true;

        const { useAuth } = await import("~/composables/useAuth");
        const { authenticatedFetch } = useAuth();

        const API_BASE_URL = "http://127.0.0.1:3000/api/v1";
        const exportUrl = `${API_BASE_URL}/surveys/${props.survey.id}/export/excel`;

        const response = await authenticatedFetch(exportUrl, {
            method: "GET",
            headers: {
                Accept: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
            },
        });

        if (!response || !response.ok) {
            throw new Error("Error al exportar a Excel");
        }

        // Get the blob from response
        const blob = await response.blob();

        // Extract filename from Content-Disposition header or use default
        // Using string split instead of regex to avoid ReDoS vulnerability
        let filename = "evaluacion_pares.xlsx";
        const contentDisposition = response.headers.get("Content-Disposition");
        if (contentDisposition) {
            const filenameToken = 'filename="';
            const startIdx = contentDisposition.indexOf(filenameToken);
            if (startIdx !== -1) {
                const valueStart = startIdx + filenameToken.length;
                const endIdx = contentDisposition.indexOf('"', valueStart);
                if (endIdx !== -1) {
                    filename = contentDisposition.slice(valueStart, endIdx);
                }
            }
        }

        // Create object URL from blob
        const url = globalThis.URL.createObjectURL(blob);

        // Create temporary link element
        const a = document.createElement("a");
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);

        // Trigger download
        a.click();

        // Cleanup
        globalThis.URL.revokeObjectURL(url);
        a.remove();

        console.log("✓ Archivo Excel descargado exitosamente:", filename);
    } catch (err) {
        console.error("Error al exportar a Excel:", err);
        error.value = err.message || "Error al exportar a Excel";
    } finally {
        exportingExcel.value = false;
    }
};
</script>

<style scoped>
.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 1rem;
}

.modal-content {
    background: white;
    border-radius: 8px;
    width: 100%;
    max-width: 1200px;
    max-height: 90vh;
    overflow: hidden;
    box-shadow:
        0 20px 25px -5px rgba(0, 0, 0, 0.1),
        0 10px 10px -5px rgba(0, 0, 0, 0.04);
}

.modal-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 1.5rem;
    border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.modal-header h2 {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--color-dark, #394049);
}

.modal-subtitle {
    margin: 0.25rem 0 0 0;
    font-size: 0.875rem;
    color: #94a3b8;
}

.close-btn {
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.25rem;
    border-radius: 4px;
    color: #94a3b8;
    transition: color 0.2s;
}

.close-btn:hover {
    color: var(--color-text-secondary, #64748b);
}

.modal-body {
    padding: 1.5rem;
    max-height: calc(90vh - 120px);
    overflow-y: auto;
}

.loading-state,
.error-state,
.no-results {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem 1rem;
    text-align: center;
}

/* View Mode Tabs */
.view-mode-tabs {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1.5rem;
    background: #f1f5f9;
    padding: 0.375rem;
    border-radius: 8px;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    scrollbar-width: none;
    -ms-overflow-style: none;
}

.view-mode-tabs::-webkit-scrollbar {
    display: none;
}

.view-mode-tab {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.625rem 1.25rem;
    border: none;
    background: transparent;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    color: #64748b;
    transition: all 0.2s;
    white-space: nowrap;
    flex: 1;
    min-width: 0;
}

.view-mode-tab:hover {
    color: #334155;
    background: var(--color-border, #dfe4ea);
}

.view-mode-tab.active {
    background: white;
    color: var(--color-primary, #006b64);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.view-mode-tab svg {
    flex-shrink: 0;
    width: 18px;
    height: 18px;
}

@media (max-width: 640px) {
    .view-mode-tabs {
        gap: 0.25rem;
        padding: 0.25rem;
        border-radius: 6px;
    }

    .view-mode-tab {
        padding: 0.5rem 0.625rem;
        font-size: 0.75rem;
        gap: 0.3rem;
        border-radius: 5px;
        flex: 1 1 0;
    }

    .view-mode-tab.tab-details {
        flex: 0 0 auto;
        padding: 0.5rem 0.5rem;
    }

    .view-mode-tab.tab-averages {
        flex: 0 0 auto;
        padding: 0.5rem 0.5rem;
    }

    .view-mode-tab svg {
        width: 14px;
        height: 14px;
    }
}

.spinner {
    width: 40px;
    height: 40px;
    border: 3px solid var(--color-border, #dfe4ea);
    border-top: 3px solid var(--color-primary, #00a499);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 1rem;
}

@keyframes spin {
    0% {
        transform: rotate(0deg);
    }
    100% {
        transform: rotate(360deg);
    }
}

.error-state svg,
.no-results svg {
    color: var(--color-danger, #d7373f);
    margin-bottom: 1rem;
}

/* Responses Modal */
.responses-modal {
    max-width: 900px;
}

/* Responses Accordion */
.responses-accordion {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.response-accordion-item {
    border: 1px solid var(--color-border, #dfe4ea);
    border-radius: 8px;
    overflow: hidden;
    background: white;
}

.response-accordion-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.25rem;
    cursor: pointer;
    background: var(--color-background, #f4f6f9);
    transition: background 0.2s;
}

.response-accordion-header:hover {
    background: var(--color-background, #f4f6f9);
}

.accordion-student-info {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.accordion-student-info .evaluated-student-name {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-dark, #394049);
}

.accordion-student-info .response-date {
    font-size: 0.75rem;
    color: #94a3b8;
}

.accordion-toggle {
    display: flex;
    align-items: center;
    gap: 1rem;
}

.response-summary {
    font-size: 0.875rem;
    color: var(--color-text-secondary, #64748b);
    font-weight: 500;
}

.accordion-toggle svg {
    color: #94a3b8;
    transition: transform 0.2s;
}

.accordion-toggle svg.rotated {
    transform: rotate(180deg);
}

.response-accordion-content {
    padding: 1.25rem;
    border-top: 1px solid var(--color-border, #dfe4ea);
    background: white;
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
        padding-top: 1.25rem;
        padding-bottom: 1.25rem;
    }
}

@keyframes accordion-close {
    0% {
        opacity: 1;
        max-height: 500px;
        padding-top: 1.25rem;
        padding-bottom: 1.25rem;
    }
    100% {
        opacity: 0;
        max-height: 0;
        padding-top: 0;
        padding-bottom: 0;
    }
}

.questions-container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.question-response {
    background: white;
    padding: 1rem;
    border-radius: 6px;
    border-left: 3px solid var(--color-primary, #00a499);
}

.question-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
}

.question-category {
    font-size: 0.75rem;
    font-weight: 700;
    color: var(--color-primary, #00a499);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    background: #ebf8ff;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
}

.question-text {
    font-size: 0.875rem;
    color: var(--color-dark, #394049);
    line-height: 1.5;
}

.score-badge {
    padding: 0.25rem 0.75rem;
    border-radius: 9999px;
    font-weight: 600;
}

.score-excellent {
    background: #c6f6d5;
    color: #22543d;
}

.score-good {
    background: #bee3f8;
    color: #2a4365;
}

.score-average {
    background: #fef3c7;
    color: #744210;
}

.score-poor {
    background: #fed7d7;
    color: #742a2a;
}

.no-questions {
    text-align: center;
    padding: 2rem 1rem;
    color: #94a3b8;
    font-style: italic;
}

/* Modal Footer */
.modal-footer {
    display: flex;
    justify-content: flex-start;
    align-items: center;
    padding: 1rem 1.5rem;
    border-top: 1px solid var(--color-border, #dfe4ea);
    background: var(--color-background, #f4f6f9);
    border-radius: 0 0 8px 8px;
}

.export-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.625rem 1.25rem;
    background: var(--color-primary, #006b64);
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.export-btn:hover:not(:disabled) {
    background: var(--color-primary-dark, #008a7f);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.15);
    transform: translateY(-1px);
}

.export-btn:active:not(:disabled) {
    transform: translateY(0);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.export-btn:disabled {
    background: #cbd5e1;
    cursor: not-allowed;
    opacity: 0.6;
}

.export-btn svg {
    flex-shrink: 0;
}

.spinner-small {
    width: 16px;
    height: 16px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top: 2px solid white;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
}

/* Responsive */
@media (max-width: 768px) {
    .modal-content {
        max-width: 95vw;
        max-height: 95vh;
    }

    .response-header {
        flex-direction: column;
        align-items: flex-start;
        gap: 0.5rem;
    }

    .modal-footer {
        padding: 0.75rem 1rem;
    }

    .export-btn {
        width: 100%;
        justify-content: center;
        padding: 0.75rem 1rem;
    }
}
</style>
