<template>
    <div v-if="isOpen" class="modal-overlay" @click="handleOverlayClick">
        <div class="modal-content" @click.stop>
            <div class="modal-header">
                <h3>{{ survey?.name || "Responder Encuesta" }}</h3>
                <button @click="$emit('close')" class="close-btn">
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
                <div v-if="loadingGroupMembers" class="loading-surveys">
                    Cargando encuesta...
                </div>

                <div v-else-if="!survey" class="no-survey">
                    <p>No se pudo cargar la encuesta</p>
                </div>

                <div v-else class="survey-content">
                    <p class="survey-description">{{ survey.description }}</p>

                    <div v-if="loadingGroupMembers" class="loading-members">
                        Cargando miembros del grupo...
                    </div>

                    <div
                        v-else-if="!groupMembers || groupMembers.length === 0"
                        class="no-group-warning"
                    >
                        <svg
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                        >
                            <path
                                d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"
                            />
                            <line x1="12" y1="9" x2="12" y2="13" />
                            <line x1="12" y1="17" x2="12.01" y2="17" />
                        </svg>
                        No estás asignado a ningún grupo. Contacta a tu
                        profesor.
                    </div>

                    <div v-else class="members-to-evaluate">
                        <h5>Miembros a evaluar ({{ groupMembers.length }}):</h5>
                        <div class="members-list">
                            <div
                                v-for="member in groupMembers"
                                :key="member.id"
                                class="member-item"
                                :class="{
                                    responded: hasResponded(
                                        survey.id,
                                        member.id,
                                    ),
                                }"
                            >
                                <div class="member-info">
                                    <span class="member-name">{{
                                        member.full_name ||
                                        `${member.first_name || ""} ${member.last_name || ""}`.trim() ||
                                        "Nombre no disponible"
                                    }}</span>
                                    <span
                                        v-if="member.id === currentUserId"
                                        class="self-indicator"
                                        >(Tú)</span
                                    >
                                </div>
                                <div class="member-status">
                                    <button
                                        v-if="
                                            hasResponded(survey.id, member.id)
                                        "
                                        @click="
                                            $emit('view-response', {
                                                surveyId: survey.id,
                                                evaluatedId: member.id,
                                            })
                                        "
                                        class="btn-view-response"
                                    >
                                        Ver Respuesta
                                    </button>
                                    <button
                                        v-else
                                        @click="respondToMember(member)"
                                        class="respond-btn"
                                    >
                                        Responder
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- Response Form Modal -->
    <div
        v-if="showResponseForm"
        class="modal-overlay"
        @click="closeResponseForm"
    >
        <div class="modal-content response-form-modal" @click.stop>
            <div class="modal-header">
                <h3>
                    Evaluar a
                    {{ selectedMember?.full_name || "Miembro del grupo" }}
                </h3>
                <button @click="closeResponseForm" class="close-btn">
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
                <div class="response-form">
                    <p class="form-instructions">
                        Evalúa el desempeño de tu compañero en cada una de las
                        siguientes categorías utilizando la escala del 1 al 4:
                    </p>

                    <div class="questions-list">
                        <div
                            v-for="question in peerEvaluationQuestions"
                            :key="question.id"
                            class="question-item"
                        >
                            <div class="question-header">
                                <span class="question-category">{{
                                    question.category
                                }}</span>
                                <span class="question-number">{{
                                    question.id
                                }}</span>
                            </div>
                            <p class="question-text">{{ question.question }}</p>

                            <div class="score-options">
                                <label
                                    v-for="option in scoreOptions"
                                    :key="option.value"
                                    class="score-option"
                                >
                                    <input
                                        type="radio"
                                        :name="`question-${question.id}`"
                                        :value="option.value"
                                        v-model="responseForm[question.id]"
                                        class="score-radio"
                                    />
                                    <span class="score-label">{{
                                        option.label
                                    }}</span>
                                </label>
                            </div>
                        </div>
                    </div>

                    <div class="form-actions">
                        <button
                            @click="closeResponseForm"
                            class="btn-cancel"
                            :disabled="submittingResponse"
                        >
                            Cancelar
                        </button>
                        <button
                            @click="submitResponse"
                            class="btn-submit"
                            :disabled="!isFormValid || submittingResponse"
                        >
                            {{
                                submittingResponse
                                    ? "Enviando..."
                                    : "Enviar Evaluación"
                            }}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, watch } from "vue";

const props = defineProps({
    isOpen: Boolean,
    survey: Object,
    sectionId: String,
    groupMembers: Array,
    loadingGroupMembers: Boolean,
    currentUserId: Number,
    responses: Array, // Array of completed responses { survey_id, evaluated_id }
});

const emit = defineEmits([
    "close",
    "respond-survey",
    "submit-response",
    "view-response",
]);

// Response form state
const showResponseForm = ref(false);
const selectedMember = ref(null);
const submittingResponse = ref(false);

// Hardcoded fallback questions used when the survey has no questions stored in the backend
const FALLBACK_PEER_EVALUATION_QUESTIONS = [
    {
        id: 1,
        category: "Tareas",
        question:
            "Realiza las tareas que le son asignadas dentro del grupo en los plazos requeridos.",
    },
    {
        id: 2,
        category: "Participación",
        question:
            "Participa de forma activa en los espacios de encuentro del equipo, compartiendo la información, conocimientos y las experiencias.",
    },
    {
        id: 3,
        category: "Organización",
        question:
            "Colabora en la definición, organización y distribución de las tareas del equipo.",
    },
    {
        id: 4,
        category: "Compromiso",
        question:
            "Se orienta a la consecución de acuerdos y objetivos comunes y se compromete con ellos.",
    },
    {
        id: 5,
        category: "Prosocialidad",
        question:
            "Toma en cuenta los puntos de vista de los demás y retroalimenta constructivamente.",
    },
];

// Use questions from the survey object when available, otherwise fall back to the hardcoded list.
// Backend questions have the shape { id, text, category } while hardcoded ones use { id, question, category }.
// We normalise to a single shape: { id, category, question }.
const peerEvaluationQuestions = computed(() => {
    const backendQuestions = props.survey?.questions;
    if (Array.isArray(backendQuestions) && backendQuestions.length > 0) {
        return backendQuestions.map((q, idx) => ({
            id: q.id ?? idx + 1,
            category: q.category || "",
            question: q.text || q.question || "",
        }));
    }
    return FALLBACK_PEER_EVALUATION_QUESTIONS;
});

// Form data for responses — keys are question IDs, built reactively from the active question list
const responseForm = ref(
    Object.fromEntries(
        FALLBACK_PEER_EVALUATION_QUESTIONS.map((q) => [q.id, null]),
    ),
);

// Reset form whenever the question list changes (e.g. when a different survey is opened)
watch(
    peerEvaluationQuestions,
    (questions) => {
        responseForm.value = Object.fromEntries(
            questions.map((q) => [q.id, null]),
        );
    },
    { immediate: false },
);

// Score options
const scoreOptions = [
    { value: 1, label: "1 - Muy bajo" },
    { value: 2, label: "2 - Bajo" },
    { value: 3, label: "3 - Alto" },
    { value: 4, label: "4 - Muy alto" },
];

const hasResponded = (surveyId, memberId) => {
    if (!props.responses || props.responses.length === 0) return false;

    // Check if current user has already evaluated this member in this specific survey
    return props.responses.some((r) => {
        const surveyMatch =
            String(r.survey_id || r.surveyId) === String(surveyId);
        const evaluatorMatch =
            String(r.evaluator_id || r.evaluatorId) ===
            String(props.currentUserId);
        const evaluatedMatch =
            String(r.evaluated_id || r.evaluatedId) === String(memberId);
        return surveyMatch && evaluatorMatch && evaluatedMatch;
    });
};

const handleOverlayClick = () => {
    emit("close");
};

const respondToMember = (member) => {
    selectedMember.value = member;
    showResponseForm.value = true;
    // Reset form
    responseForm.value = { 1: null, 2: null, 3: null, 4: null, 5: null };
};

const closeResponseForm = () => {
    showResponseForm.value = false;
    selectedMember.value = null;
};

const isFormValid = computed(() => {
    const questionIds = peerEvaluationQuestions.value.map((q) => String(q.id));
    return questionIds.every(
        (id) =>
            responseForm.value[id] !== null &&
            responseForm.value[id] !== undefined,
    );
});

const submitResponse = async () => {
    if (!isFormValid.value || !selectedMember.value || !props.survey) return;

    try {
        submittingResponse.value = true;

        // Convert form data to API format
        const responses = Object.entries(responseForm.value).map(
            ([questionId, score]) => ({
                question_id: Number.parseInt(questionId),
                score: score,
            }),
        );

        // Emit the response data to parent component
        emit("submit-response", {
            surveyId: props.survey.id,
            evaluatedId: selectedMember.value.id,
            responses: responses,
        });

        // Close the form
        closeResponseForm();
    } catch (error) {
        console.error("Error submitting response:", error);
        alert("Error al enviar la respuesta. Por favor intenta de nuevo.");
    } finally {
        submittingResponse.value = false;
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
    border-radius: 12px;
    box-shadow:
        0 20px 25px -5px rgba(0, 0, 0, 0.1),
        0 10px 10px -5px rgba(0, 0, 0, 0.04);
    max-width: 600px;
    width: 100%;
    max-height: 90vh;
    overflow-y: auto;
}

.modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem;
    border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.modal-header h3 {
    margin: 0;
    color: var(--color-dark, #394049);
    font-size: 1.25rem;
    font-weight: 600;
}

.close-btn {
    background: none;
    border: none;
    color: #94a3b8;
    cursor: pointer;
    padding: 0.25rem;
    border-radius: 4px;
    transition: all 0.2s;
}

.close-btn:hover {
    background: var(--color-background, #f4f6f9);
    color: var(--color-dark, #394049);
}

.modal-body {
    padding: 1.5rem;
}

.loading-surveys {
    text-align: center;
    padding: 2rem;
    color: #94a3b8;
}

.no-survey {
    text-align: center;
    padding: 2rem;
    color: #a0aec0;
}

.survey-description {
    color: var(--color-text-secondary, #64748b);
    margin-bottom: 1.5rem;
    line-height: 1.5;
}

.loading-members {
    padding: 1rem;
    text-align: center;
    color: #94a3b8;
    font-style: italic;
}

.no-group-warning {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1rem;
    background: #fff5f5;
    border: 1px solid #feb2b2;
    border-radius: 8px;
    color: #b82e35;
    margin-bottom: 1rem;
}

.members-to-evaluate h5 {
    margin: 0 0 1rem 0;
    color: var(--color-dark, #394049);
    font-size: 1rem;
    font-weight: 600;
}

.members-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-bottom: 1.5rem;
}

.member-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    background: var(--color-background, #f4f6f9);
    border: 1px solid var(--color-border, #dfe4ea);
    border-radius: 8px;
    transition: all 0.2s;
}

.member-item:hover {
    background: var(--color-background, #f4f6f9);
    border-color: #cbd5e0;
}

.member-item.responded {
    background: #f0fff4;
    border-color: #9ae6b4;
}

.member-info {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex: 1;
}

.member-name {
    color: var(--color-dark, #394049);
    font-weight: 500;
}

.self-indicator {
    color: var(--color-primary, #00a499);
    font-size: 0.875rem;
    font-weight: 600;
}

.member-status {
    display: flex;
    align-items: center;
}

.respond-btn {
    padding: 0.625rem 1.25rem;
    background: #006b64;
    color: white;
    border: none;
    border-radius: 6px;
    font-weight: 500;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s;
}

.respond-btn:hover {
    background: #008f85;
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.btn-view-response {
    padding: 0.625rem 1.25rem;
    background: var(--color-primary, #006b64);
    color: white;
    border: none;
    border-radius: 6px;
    font-weight: 500;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s;
}

.btn-view-response:hover {
    background: var(--color-primary-dark, #008a7f);
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

/* Response Form Modal Styles */
.response-form-modal .modal-content {
    max-width: 700px;
    max-height: 90vh;
    overflow-y: auto;
}

.response-form {
    padding: 0;
}

.form-instructions {
    color: var(--color-text-secondary, #64748b);
    font-size: 0.875rem;
    margin-bottom: 1.5rem;
    padding: 1rem;
    background: var(--color-background, #f4f6f9);
    border-radius: 6px;
    border-left: 4px solid #00a499;
}

.questions-list {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
}

.question-item {
    padding: 1.25rem;
    background: #f8f9fa;
    border: 1px solid var(--color-border, #dfe4ea);
    border-radius: 8px;
    transition: all 0.2s;
}

.question-item:hover {
    border-color: #cbd5e0;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.question-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.75rem;
}

.question-category {
    font-weight: 600;
    color: #005f59;
    background: #e6f7f6;
    padding: 0.25rem 0.75rem;
    border-radius: 20px;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.question-number {
    background: #006b64;
    color: white;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    font-weight: 600;
}

.question-text {
    color: var(--color-dark, #394049);
    font-size: 0.95rem;
    line-height: 1.5;
    margin-bottom: 1rem;
}

.score-options {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.score-option {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.5rem;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s;
}

.score-option:hover {
    background: var(--color-background, #f4f6f9);
}

.score-radio {
    width: 16px;
    height: 16px;
    accent-color: #00a499;
    cursor: pointer;
}

.score-label {
    color: var(--color-text-secondary, #64748b);
    font-size: 0.875rem;
    cursor: pointer;
}

.form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
    margin-top: 2rem;
    padding-top: 1.5rem;
    border-top: 1px solid var(--color-border, #dfe4ea);
}

.btn-cancel {
    padding: 0.75rem 1.5rem;
    background: var(--color-background, #f4f6f9);
    color: var(--color-text-secondary, #64748b);
    border: 1px solid var(--color-border, #dfe4ea);
    border-radius: 6px;
    cursor: pointer;
    font-weight: 500;
    transition: all 0.2s;
}

.btn-cancel:hover {
    background: var(--color-background, #f4f6f9);
    border-color: #cbd5e0;
}

.btn-submit {
    padding: 0.75rem 1.5rem;
    background: #006b64;
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-weight: 500;
    transition: all 0.2s;
}

.btn-submit:hover:not(:disabled) {
    background: #008f85;
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.btn-submit:disabled {
    background: #a0aec0;
    cursor: not-allowed;
    transform: none;
}
</style>
