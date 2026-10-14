<template>
    <div>
        <SectionSurveys
            :surveys="allSurveys"
            :loading-surveys="loadingPeerEvaluationSurveys"
            :user-role="props.user?.role_name"
            @create-peer-evaluation="openCreatePeerEvaluationModal"
            @view-responses="handleViewSurveyResponses"
            @edit-survey="handleEditSurvey"
            @delete-survey="handleDeleteSurvey"
            @respond-survey="openStudentResponseModal"
        />

        <!-- Modal de Crear Evaluación entre Pares -->
        <CreatePeerEvaluationModal
            :is-open="showCreatePeerEvaluationModal"
            :submitting="submittingCreatePeerEvaluation"
            @close="closeCreatePeerEvaluationModal"
            @submit="handleCreatePeerEvaluation"
        />

        <!-- Modal de Respuesta de Estudiante -->
        <StudentSurveyResponse
            :is-open="showStudentResponseModal"
            :survey="selectedSurveyForResponse"
            :section-id="route.params.id"
            :group-members="studentGroupMembers"
            :loading-group-members="loadingGroupMembers"
            :current-user-id="props.user?.id"
            :responses="surveyResponses"
            @close="closeStudentResponseModal"
            @submit-response="handleSubmitResponse"
            @view-response="handleViewMyResponse"
        />

        <!-- Confirm Modal -->
        <ConfirmModal
            :show="showConfirmModal"
            :message="confirmMessage"
            @confirm="handleConfirm"
            @cancel="showConfirmModal = false"
        />

        <!-- Info Modal -->
        <InfoModal
            :show="showInfoModal"
            :title="infoTitle"
            :message="infoMessage"
            @close="showInfoModal = false"
        />
        <ViewPeerEvaluationResponseModal
            :is-open="showViewMyResponseModal"
            :response="selectedMyResponse"
            :all-students="studentsModule.students"
            @close="closeViewMyResponseModal"
        />

        <!-- View Survey Responses Modal -->
        <ViewSurveyResponsesModal
            :is-open="showViewResponsesModal"
            :survey="selectedSurveyForView"
            :section-id="route.params.id"
            :all-students="studentsModule.students"
            @close="closeViewResponsesModal"
        />

        <!-- Selection Test (Group Compatibility) Modal for Students -->
        <SelectionTest
            :is-open="showSociometricTestModal"
            :questionnaire="selectedSociometricSurvey"
            :student-id="user?.id"
            :section-id="route.params.id"
            @close="closeSociometricModal"
            @submitted="onSociometricSubmitted"
        />
    </div>
</template>

<script setup>
import { onMounted, computed, toRefs } from "vue";
import ConfirmModal from "~/components/ConfirmModal.vue";
import InfoModal from "~/components/InfoModal.vue";
import SectionSurveys from "~/components/SectionSurveys.vue";
import CreatePeerEvaluationModal from "~/components/CreatePeerEvaluationModal.vue";
import StudentSurveyResponse from "~/components/StudentSurveyResponse.vue";
import ViewSurveyResponsesModal from "~/components/ViewSurveyResponsesModal.vue";
import { useSectionSurveys } from "~/composables/useSectionSurveys";
import { useSectionStudents } from "~/composables/useSectionStudents";
import { useSectionGroups } from "~/composables/useSectionGroups";
import { usePluginApi } from "~/composables/usePluginApi";
import ViewPeerEvaluationResponseModal from "~/components/ViewPeerEvaluationResponseModal.vue";
import SelectionTest from "~/components/generic/SelectionTest.vue";
import { useAuth } from "~/composables/useAuth";

const pluginApi = usePluginApi();

const props = defineProps({
    sectionData: Object,
    courseData: Object,
    semesterInfo: String,
    user: Object,
});

const route = useRoute();
const { getToken } = useAuth();

// Initialize composables
const surveysModule = useSectionSurveys();
const studentsModule = useSectionStudents();
const groupsModule = useSectionGroups();

// Extract reactive references from composables
const {
    peerEvaluationSurveys,
    compatibilitySurveys,
    loadingPeerEvaluationSurveys,
    showCreatePeerEvaluationModal,
} = toRefs(surveysModule);

// Use studentsModule.students directly to avoid naming conflicts
// const { students } = toRefs(studentsModule)
const showConfirmModal = ref(false);
const confirmMessage = ref("");
const pendingAction = ref(null);

// Info modal state
const showInfoModal = ref(false);
const infoTitle = ref("");
const infoMessage = ref("");

// Function to show info modal
const showInfo = (title, message) => {
    infoTitle.value = title;
    infoMessage.value = message;
    showInfoModal.value = true;
};

const showViewMyResponseModal = ref(false);
const selectedMyResponse = ref(null);
// Student response modal state
const showStudentResponseModal = ref(false);
const selectedSurveyForResponse = ref(null);
const studentGroupMembers = ref([]);
const loadingGroupMembers = ref(false);
const surveyResponses = ref([]);
const currentGroupId = ref(null);

// Sociometric modal state
const showSociometricTestModal = ref(false);
const selectedSociometricSurvey = ref(null);

// Submitting state for create modal
const submittingCreatePeerEvaluation = ref(false);

// View survey responses modal state
const showViewResponsesModal = ref(false);
const selectedSurveyForView = ref(null);

// Open student response modal
const openStudentResponseModal = async (survey) => {
    // Group compatibility surveys use the dedicated SelectionTest modal
    if (survey.plugin === "group_compatibility") {
        // Check if student already responded
        const studentId = user.value?.id?.toString();
        if (studentId && survey.id) {
            const result = await pluginApi.checkStudentTestResponse(survey.plugin, survey.id, studentId);
            if (result.success && result.hasResponded) {
                showInfo("Cuestionario completado", "Ya has respondido este cuestionario anteriormente.");
                return;
            }
        }
        selectedSociometricSurvey.value = survey;
        showSociometricTestModal.value = true;
        return;
    }
    // Peer evaluation surveys use the existing StudentSurveyResponse modal
    selectedSurveyForResponse.value = survey;
    showStudentResponseModal.value = true;
    await loadStudentGroupData(survey);
};

// Close sociometric modal
const closeSociometricModal = () => {
    showSociometricTestModal.value = false;
    selectedSociometricSurvey.value = null;
};

// Handle sociometric response submitted
const onSociometricSubmitted = () => {
    showInfo("Éxito", "Respuesta enviada exitosamente");
    closeSociometricModal();
};

// Close student response modal
const closeStudentResponseModal = () => {
    showStudentResponseModal.value = false;
    selectedSurveyForResponse.value = null;
    studentGroupMembers.value = [];
    surveyResponses.value = [];
};

// Close view responses modal
const closeViewResponsesModal = () => {
    showViewResponsesModal.value = false;
    selectedSurveyForView.value = null;
};

// Close view my response modal
const closeViewMyResponseModal = () => {
    showViewMyResponseModal.value = false;
    selectedMyResponse.value = null;
};

// Load student group data for survey response
const loadStudentGroupData = async (survey) => {
    try {
        loadingGroupMembers.value = true;
        const sectionId = route.params.id;
        const token = getToken();
        const userId = props.user?.id;
        console.log(
            "LoadGroupData: token present:",
            !!token,
            "userId:",
            userId,
        );

        // Ensure the full student list is loaded once
        await studentsModule.loadStudents(sectionId, token);
        const allStudents = studentsModule.students.value;

        // 1. Try to get peers from the survey snapshot (recorded at creation time)
        const snapshotResult = await surveysModule.getSurveyPeers(
            survey.id,
            String(userId),
            token,
        );

        let peerIds = [];
        let resolvedGroupId = null;

        if (
            snapshotResult.success &&
            snapshotResult.peers &&
            snapshotResult.peers.length > 0
        ) {
            peerIds = snapshotResult.peers;
            resolvedGroupId = snapshotResult.group_id;
            console.log(
                "Peers loaded from survey snapshot:",
                peerIds.length,
                "group_id:",
                resolvedGroupId,
            );
        } else {
            // 2. Fallback: load the student's current group
            console.warn(
                "No peers found in survey snapshot for user:",
                userId,
                "— falling back to current group",
            );
            const currentGroupResult =
                await surveysModule.getCurrentGroupPeers(sectionId);

            if (
                currentGroupResult.success &&
                currentGroupResult.peers &&
                currentGroupResult.peers.length > 0
            ) {
                peerIds = currentGroupResult.peers;

                // El endpoint my-group-complete ahora devuelve group_number.
                // Intentamos resolver el group_id real (ObjectId) buscando en los
                // snapshots de la encuesta por group_number.
                const fallbackGroupNumber =
                    currentGroupResult.group_number ?? null;
                if (
                    fallbackGroupNumber !== null &&
                    survey?.group_snapshots?.length > 0
                ) {
                    const matchingSnapshot = survey.group_snapshots.find(
                        (snap) => snap.group_number === fallbackGroupNumber,
                    );
                    resolvedGroupId = matchingSnapshot?.group_id ?? null;
                    console.log(
                        "Resolved group_id from snapshot by group_number:",
                        fallbackGroupNumber,
                        "→",
                        resolvedGroupId,
                    );
                } else {
                    resolvedGroupId = currentGroupResult.group_id ?? null;
                }

                // If the gateway returned full student objects, use them directly
                if (
                    currentGroupResult.students &&
                    currentGroupResult.students.length > 0 &&
                    typeof currentGroupResult.students[0] === "object"
                ) {
                    studentGroupMembers.value =
                        currentGroupResult.students.filter(
                            (s) => s && (s.id ?? s.student_id),
                        );
                    currentGroupId.value = resolvedGroupId;
                    console.log(
                        "Group members resolved directly from current group objects:",
                        studentGroupMembers.value.length,
                        "group_id:",
                        resolvedGroupId,
                    );
                    // Cargar respuestas existentes para marcar miembros ya evaluados
                    if (resolvedGroupId) {
                        const responsesResult =
                            await surveysModule.getSectionGroupResponses(
                                sectionId,
                                resolvedGroupId,
                                token,
                            );
                        surveyResponses.value = responsesResult.success
                            ? responsesResult.data
                            : [];
                        console.log(
                            "Loaded survey responses (fallback early-return):",
                            surveyResponses.value.length,
                        );
                    } else {
                        surveyResponses.value = [];
                    }
                    return;
                }

                console.log(
                    "Peers loaded from current group (fallback):",
                    peerIds.length,
                    "group_id:",
                    resolvedGroupId,
                );
            } else {
                console.warn(
                    "Student is not assigned to any group in this section",
                );
                studentGroupMembers.value = [];
                currentGroupId.value = null;
                surveyResponses.value = [];
                return;
            }
        }

        // Map peer IDs to full student objects from the already-loaded list
        const peerStudents = peerIds
            .map((peerId) =>
                allStudents.find((s) => String(s.id) === String(peerId)),
            )
            .filter(Boolean);

        studentGroupMembers.value = peerStudents;
        currentGroupId.value = resolvedGroupId;
        console.log(
            "Final group members:",
            peerStudents.length,
            "group_id:",
            resolvedGroupId,
        );

        // Load responses for the section/group to mark already-answered members
        if (resolvedGroupId) {
            const responsesResult =
                await surveysModule.getSectionGroupResponses(
                    sectionId,
                    resolvedGroupId,
                    token,
                );
            surveyResponses.value = responsesResult.success
                ? responsesResult.data
                : [];
            console.log(
                "Loaded survey responses:",
                surveyResponses.value.length,
            );
        } else {
            surveyResponses.value = [];
        }
    } catch (error) {
        console.error("Error loading student group data:", error);
        studentGroupMembers.value = [];
        surveyResponses.value = [];
    } finally {
        loadingGroupMembers.value = false;
    }
};

// Computed property to combine all surveys
const allSurveys = computed(() => {
    return [...peerEvaluationSurveys.value];
});

// Handlers
const openCreatePeerEvaluationModal = () => {
    submittingCreatePeerEvaluation.value = false;
    surveysModule.openCreatePeerEvaluationModal();
};

const closeCreatePeerEvaluationModal = () => {
    submittingCreatePeerEvaluation.value = false;
    surveysModule.closeCreatePeerEvaluationModal();
};

const handleViewSurveyResponses = (survey) => {
    selectedSurveyForView.value = survey;
    showViewResponsesModal.value = true;
};

const handleCreatePeerEvaluation = async (surveyData) => {
    submittingCreatePeerEvaluation.value = true;
    try {
        const sectionId = route.params.id;
        const token = getToken();

        const result = await surveysModule.createPeerEvaluationSurvey(
            surveyData,
            sectionId,
            token,
        );

        if (result.success) {
            showInfo("Éxito", "Evaluación entre pares creada exitosamente");
            closeCreatePeerEvaluationModal();
        }
    } catch (error) {
        console.error("Error creating peer evaluation survey:", error);
        showInfo(
            "Error",
            "Error al crear la evaluación entre pares: " + error.message,
        );
    } finally {
        submittingCreatePeerEvaluation.value = false;
    }
};

const handleViewMyResponse = async ({ surveyId, evaluatedId }) => {
    try {
        const evaluatorId = props.user?.id;
        if (!evaluatorId) {
            showInfo(
                "Error",
                "No se pudo identificar al evaluador. Vuelve a iniciar sesión.",
            );
            return;
        }
        const response = await surveysModule.getPeerEvaluationResponse(
            surveyId,
            evaluatorId,
            evaluatedId,
        );
        if (!response) {
            showInfo(
                "Info",
                "No se encontró una respuesta tuya para este miembro",
            );
            return;
        }
        // If the API returns an array pick the first
        selectedMyResponse.value = Array.isArray(response)
            ? response[0]
            : response;
        showViewMyResponseModal.value = true;
    } catch (error) {
        console.error("Error fetching my response:", error);
        const message = error?.message ?? String(error);
        showInfo("Error", "Error al obtener la respuesta: " + message);
    }
};

const handleEditSurvey = async (survey) => {
    try {
        // Para editar, necesitaríamos abrir un modal con los datos
        // Por ahora, mostrar mensaje de que se debe usar el modal de crear para editar
        showInfo(
            "Info",
            "Editar encuestas aún no está disponible. Elimina y crea una nueva.",
        );
    } catch (error) {
        showInfo("Error", "Error: " + error.message);
    }
};

const handleDeleteSurvey = (survey) => {
    confirmMessage.value = `¿Estás seguro de que deseas eliminar la encuesta "${survey.name}"?`;
    pendingAction.value = () => performDeleteSurvey(survey);
    showConfirmModal.value = true;
};

const performDeleteSurvey = async (survey) => {
    const token = getToken();
    try {
        const result = await surveysModule.deleteSurvey(survey.id, token);
        if (result.success) {
            showInfo("Éxito", result.message);
        } else {
            showInfo("Error", result.error || "Error al eliminar encuesta");
        }
    } catch (error) {
        showInfo("Error", "Error al eliminar encuesta: " + error.message);
    }
};

const handleConfirm = () => {
    if (pendingAction.value) {
        pendingAction.value();
        pendingAction.value = null;
    }
    showConfirmModal.value = false;
};

const handleSubmitResponse = async (responseData) => {
    try {
        const evaluatorId = props.user?.id;
        if (!evaluatorId) {
            console.error("handleSubmitResponse: missing evaluatorId");
            showInfo(
                "Error",
                "No se encontró el id del evaluador. Vuelve a iniciar sesión o contacta a tu profesor.",
            );
            return;
        }

        console.log("Submitting peer evaluation:", {
            surveyId: responseData.surveyId,
            evaluatorId: evaluatorId,
            evaluatedId: responseData.evaluatedId,
            groupId: currentGroupId.value,
            responses: responseData.responses,
        });
        const result = await surveysModule.submitPeerEvaluationResponse(
            responseData.surveyId,
            evaluatorId,
            responseData.evaluatedId,
            currentGroupId.value,
            responseData.responses,
        );

        if (result.success) {
            showInfo("Éxito", "Evaluación enviada exitosamente");
            // Reload the survey responses to ensure consistency
            await loadStudentGroupData(selectedSurveyForResponse.value);
        }
    } catch (error) {
        console.error("Error submitting response:", error);
        // Show full error value for debugging if message is missing
        const message = error?.message ?? String(error);
        showInfo("Error", "Error al enviar la evaluación: " + message);
    }
};

// Load surveys on mount
onMounted(async () => {
    const sectionId = route.params.id;
    const token = getToken();

    // Load students for the survey responses modal
    await studentsModule.loadStudents(sectionId, token);

    // Load peer evaluation surveys
    await surveysModule.loadPeerEvaluationSurveys(sectionId, token);
});
</script>

<style scoped>
/* No additional styles needed - modal has its own styles */
</style>
