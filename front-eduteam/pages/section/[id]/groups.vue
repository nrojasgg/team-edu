<template>
    <div>
        <!-- Vista para estudiantes: Cuestionarios y Tu grupo -->
        <div v-if="user?.role_name === 'estudiante'" class="student-view">
            <!-- Coordinación  de Cuestionarios -->
            <div class="questionnaires-section">
                <div class="questionnaires-header">
                    <h3>Cuestionarios</h3>
                </div>

                <div
                    v-if="loadingQuestionnaires"
                    class="loading-questionnaires"
                >
                    Cargando cuestionarios...
                </div>

                <div
                    v-else-if="questionnaires.length === 0"
                    class="no-questionnaires"
                >
                    No hay cuestionarios disponibles para esta coordinación
                </div>

                <div v-else class="questionnaires-container">
                    <div
                        v-for="questionnaire in questionnaires"
                        :key="questionnaire.id || questionnaire._id"
                        class="questionnaire-card"
                    >
                        <div class="questionnaire-header">
                            <h4 class="questionnaire-title">
                                {{
                                    questionnaire.title ||
                                    questionnaire.name ||
                                    "Test sin nombre"
                                }}
                            </h4>
                            <span class="questionnaire-plugin">{{
                                getPluginLabel(questionnaire)
                            }}</span>
                        </div>

                        <div class="questionnaire-info">
                            <p class="questionnaire-description">
                                {{
                                    questionnaire.description ||
                                    "Sin descripción"
                                }}
                            </p>

                            <div class="questionnaire-dates">
                                <div class="date-info">
                                    <strong>Inicio:</strong>
                                    {{ formatDate(questionnaire.start_date) }}
                                </div>
                                <div class="date-info">
                                    <strong>Término:</strong>
                                    {{ formatDate(questionnaire.end_date) }}
                                </div>
                            </div>
                        </div>

                        <div class="questionnaire-actions">
                            <button
                                class="action-btn respond-btn"
                                :class="{
                                    'completed-btn':
                                        completedQuestionnaireIds.has(
                                            questionnaire.id ||
                                                questionnaire._id,
                                        ),
                                }"
                                :disabled="
                                    completedQuestionnaireIds.has(
                                        questionnaire.id || questionnaire._id,
                                    )
                                "
                                @click="openLikertTest(questionnaire)"
                            >
                                <svg
                                    v-if="
                                        !completedQuestionnaireIds.has(
                                            questionnaire.id ||
                                                questionnaire._id,
                                        )
                                    "
                                    width="16"
                                    height="16"
                                    viewBox="0 0 24 24"
                                    fill="none"
                                    stroke="currentColor"
                                    stroke-width="2"
                                >
                                    <path
                                        d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"
                                    />
                                    <circle cx="12" cy="12" r="3" />
                                </svg>
                                <svg
                                    v-else
                                    width="16"
                                    height="16"
                                    viewBox="0 0 24 24"
                                    fill="none"
                                    stroke="currentColor"
                                    stroke-width="2"
                                >
                                    <polyline points="20 6 9 17 4 12" />
                                </svg>
                                {{
                                    completedQuestionnaireIds.has(
                                        questionnaire.id || questionnaire._id,
                                    )
                                        ? "Ya respondido"
                                        : "Responder"
                                }}
                            </button>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Coordinación  de Tu grupo -->
            <div class="my-group-section">
                <div class="my-group-header">
                    <h3>Tu Grupo</h3>
                </div>

                <div v-if="loadingMyGroup" class="loading-group">
                    Cargando información de tu grupo...
                </div>

                <div
                    v-else-if="
                        !myGroupInfo || myGroupInfo.students.length === 0
                    "
                    class="no-group"
                >
                    <svg
                        width="48"
                        height="48"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                    >
                        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
                        <circle cx="9" cy="7" r="4" />
                        <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
                        <path d="M16 3.13a4 4 0 0 1 0 7.75" />
                    </svg>
                    <p>Aún no has sido asignado a un grupo</p>
                </div>

                <div v-else class="group-info">
                    <div class="group-name-badge">
                        <svg
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                        >
                            <path
                                d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"
                            />
                            <circle cx="9" cy="7" r="4" />
                            <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
                            <path d="M16 3.13a4 4 0 0 1 0 7.75" />
                        </svg>
                        {{ myGroupInfo.group_name }}
                    </div>

                    <div class="group-members">
                        <div
                            v-for="member in myGroupMembers"
                            :key="member.id"
                            class="member-card"
                        >
                            <div class="member-avatar">
                                <svg
                                    width="24"
                                    height="24"
                                    viewBox="0 0 24 24"
                                    fill="none"
                                    stroke="currentColor"
                                    stroke-width="2"
                                >
                                    <path
                                        d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"
                                    />
                                    <circle cx="12" cy="7" r="4" />
                                </svg>
                            </div>
                            <div class="member-info">
                                <div class="member-name">
                                    {{ member.nombres }} {{ member.paterno }}
                                </div>
                                <div class="member-email">
                                    {{ member.email }}
                                </div>
                                <div
                                    v-if="member.eneatype"
                                    class="eneatype-info"
                                >
                                    <div class="eneatype-badge">
                                        Eneatipo {{ member.eneatype }}
                                    </div>
                                    <div
                                        v-if="member.ocean_vector"
                                        class="ocean-mini-grid"
                                        title="Vector OCEAN (O, C, E, A, N)"
                                    >
                                        <span class="ocean-mini-val o">{{
                                            (
                                                member.ocean_vector.openness *
                                                10
                                            ).toFixed(0)
                                        }}</span>
                                        <span class="ocean-mini-val c">{{
                                            (
                                                member.ocean_vector
                                                    .conscientiousness * 10
                                            ).toFixed(0)
                                        }}</span>
                                        <span class="ocean-mini-val e">{{
                                            (
                                                member.ocean_vector
                                                    .extraversion * 10
                                            ).toFixed(0)
                                        }}</span>
                                        <span class="ocean-mini-val a">{{
                                            (
                                                member.ocean_vector
                                                    .agreeableness * 10
                                            ).toFixed(0)
                                        }}</span>
                                        <span class="ocean-mini-val n">{{
                                            (
                                                member.ocean_vector
                                                    .neuroticism * 10
                                            ).toFixed(0)
                                        }}</span>
                                    </div>
                                </div>
                                <div v-else class="no-eneatype">
                                    Sin eneatipo
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Vista para administradores/profesores -->
        <SectionStudentGroups
            v-else
            :section-id="route.params.id"
            :questionnaires="questionnaires"
            :loading-questionnaires="loadingQuestionnaires"
            :available-students="availableStudents"
            :groups="groups"
            :global-score="globalScore"
            :student-test-status="studentTestStatus"
            :open-plugin-modal-trigger="openPluginModalTrigger"
            :read-only="false"
            :student-id="user?.id"
            @create-questionnaire="openCreateQuestionnaireModal"
            @edit-questionnaire="openEditQuestionnaireModal"
            @view-responses="openViewResponsesModal"
            @delete-questionnaire="handleDeleteQuestionnaire"
            @add-group="addNewGroup"
            @delete-all-groups="deleteAllGroups"
            @delete-group="deleteGroup"
            @update-group-name="updateGroupName"
            @save-group-name="saveGroupName"
            @form-groups-by-plugin="formGroupsByPlugin"
            @form-random-groups="formRandomGroups"
            @drag-start="onDragStart"
            @drop-to-group="onDropToGroup"
            @drop-to-available="onDropToAvailable"
            @drag-over="onDragOver"
            @export-groups="exportGroupsToExcel"
            @open-plugin-modal="handleOpenPluginModal"
        />

        <!-- Modal de Crear Cuestionario -->
        <CreateQuestionnaireModal
            :is-open="showCreateQuestionnaireModal"
            :user="user"
            @close="closeCreateQuestionnaireModal"
            @create="handleCreateQuestionnaire"
        />

        <!-- Edit Questionnaire Modal -->
        <EditQuestionnaireModal
            :is-open="showEditQuestionnaireModal"
            :questionnaire="selectedQuestionnaire"
            @close="closeEditQuestionnaireModal"
            @save="handleSaveQuestionnaire"
        />

        <!-- View Students Response Modal -->
        <ViewStudentsResponseModal
            :is-open="showViewResponsesModal"
            :questionnaire="selectedQuestionnaire"
            :section-id="route.params.id"
            :all-students="students"
            @close="closeViewResponsesModal"
        />

        <!-- Selection Test Modal (for group_compatibility) -->
        <SelectionTest
            :is-open="showSelectionTest"
            :questionnaire="selectedQuestionnaireForTest"
            :student-id="user?.id"
            :section-id="route.params.id"
            @close="closeGenericTest"
            @submitted="onResponseSubmitted"
        />

        <!-- Form Test Modal (for demographic) -->
        <FormTest
            :is-open="showFormTest"
            :questionnaire="selectedQuestionnaireForTest"
            :student-id="user?.id"
            :section-id="route.params.id"
            @close="closeGenericTest"
            @submitted="onResponseSubmitted"
        />

        <!-- Generic Likert Test Modal (for Likert-type plugins) -->
        <LikertTest
            v-if="showLikertTest"
            :is-open="showLikertTest"
            :questionnaire="selectedGenericTest"
            :metadata="genericTestMetadata"
            @close="closeGenericTest"
            @submitted="onResponseSubmitted"
            @view-results="onViewTestResults"
        />

        <!-- Generic Test Results Modal -->
        <TestResults
            :is-open="showTestResults"
            :results="testResultsData"
            :plugin-name="testResultsPlugin"
            @close="closeTestResults"
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

        <!-- Conflict Resolution Modal (shown before Pareto) -->
        <ConflictResolutionModal
            v-if="showConflictModal && groupsModule.paretoFront.value"
            :criteria-parameters="conflictCriteriaParameters"
            :pareto-front="groupsModule.paretoFront.value"
            @weighted-solution="handleWeightedSolution"
            @show-pareto="handleShowPareto"
            @dismiss="handleConflictDismiss"
        />

        <!-- Pareto Front Chart Modal -->
        <ParetoFrontChart
            v-if="showParetoChart && groupsModule.paretoFront.value"
            :solutions="groupsModule.paretoFront.value"
            :knee-index="groupsModule.paretoResult.value?.data?.metadata?.knee_index ?? -1"
            @select="handleParetoSelect"
            @dismiss="handleParetoDismiss"
            @back="handleParetoBack"
        />
    </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import SectionStudentGroups from "~/components/SectionStudentGroups.vue";
import EditQuestionnaireModal from "~/components/EditQuestionnaireModal.vue";
import ViewStudentsResponseModal from "~/components/ViewStudentsResponseModal.vue";
import CreateQuestionnaireModal from "~/components/CreateQuestionnaireModal.vue";
import ConfirmModal from "~/components/ConfirmModal.vue";
import InfoModal from "~/components/InfoModal.vue";
import ConflictResolutionModal from "~/components/ConflictResolutionModal.vue";
import ParetoFrontChart from "~/components/ParetoFrontChart.vue";
import LikertTest from "~/components/generic/LikertTest.vue";
import SelectionTest from "~/components/generic/SelectionTest.vue";
import FormTest from "~/components/generic/FormTest.vue";
import TestResults from "~/components/generic/TestResults.vue";
import { useSectionGroups } from "~/composables/useSectionGroups";
import { useSectionStudents } from "~/composables/useSectionStudents";
import { usePluginApi } from "~/composables/usePluginApi";
import { useAuth } from "~/composables/useAuth";

const props = defineProps({
    sectionData: Object,
    courseData: Object,
    semesterInfo: String,
    user: Object,
});

const route = useRoute();
const { getToken } = useAuth();

// Initialize composables
const groupsModule = useSectionGroups();
const studentsModule = useSectionStudents();
const pluginApi = usePluginApi();

// Access composable state directly (refs returned from composable)
const groups = groupsModule.groups;
const availableStudents = groupsModule.availableStudents;
const questionnaires = groupsModule.questionnaires;
const loadingQuestionnaires = groupsModule.loadingQuestionnaires;
const globalScore = groupsModule.globalScore;
const students = studentsModule.students;

// Student test status for plugin modal
const studentTestStatus = ref({});

// Trigger to open plugin modal (used to coordinate data loading)
const openPluginModalTrigger = ref(0);

// Use refs from composable
const showCreateQuestionnaireModal = computed(
    () => groupsModule.showCreateQuestionnaireModal.value,
);

// Confirm modal state
const showConfirmModal = ref(false);
const confirmMessage = ref("");
const pendingAction = ref(null);

// Info modal state
const showInfoModal = ref(false);
const infoTitle = ref("");
const infoMessage = ref("");

// Helper function to show info modal
const showInfo = (title, message) => {
    infoTitle.value = title;
    infoMessage.value = message;
    showInfoModal.value = true;
};

// New state for edit and view modals
const showEditQuestionnaireModal = ref(false);
const showViewResponsesModal = ref(false);
const showParetoChart = ref(false);
const showConflictModal = ref(false);
const conflictCriteriaParameters = ref({});
const conflictFormConfig = ref(null);
const selectedQuestionnaire = ref(null);

// State for student's group
const myGroupInfo = ref(null);
const myGroupMembers = ref([]);
const loadingMyGroup = ref(false);
const selectedQuestionnaireForTest = ref(null);
const completedQuestionnaireIds = ref(new Set());

// Generic test modal states
const showSelectionTest = ref(false);  // group_compatibility
const showFormTest = ref(false);       // demographic
const showLikertTest = ref(false);     // bfi44, mbti, enneagram, chronotype
const selectedGenericTest = ref(null);
const genericTestMetadata = ref(null);
const pluginMetadataCache = ref({});

// Results modal state
const showTestResults = ref(false);
const testResultsData = ref(null);
const testResultsPlugin = ref('');

// Handlers
const openCreateQuestionnaireModal = () => {
    groupsModule.openCreateQuestionnaireModal();
};

const closeCreateQuestionnaireModal = () => {
    groupsModule.closeCreateQuestionnaireModal();
};

const openEditQuestionnaireModal = (questionnaire) => {
    selectedQuestionnaire.value = questionnaire;
    showEditQuestionnaireModal.value = true;
};

const closeEditQuestionnaireModal = () => {
    showEditQuestionnaireModal.value = false;
    selectedQuestionnaire.value = null;
};

const openViewResponsesModal = (questionnaire) => {
    selectedQuestionnaire.value = questionnaire;
    showViewResponsesModal.value = true;
};

const closeViewResponsesModal = () => {
    showViewResponsesModal.value = false;
    selectedQuestionnaire.value = null;
};

const handleSaveQuestionnaire = async (updatedData) => {
    try {
        const token = getToken();
        const result = await groupsModule.updateQuestionnaire(
            updatedData,
            token,
        );
        showInfo("Éxito", result.message);
        closeEditQuestionnaireModal();

        // Reload questionnaires
        const sectionId = route.params.id;
        await groupsModule.loadQuestionnaires(sectionId, token);
    } catch (error) {
        showInfo("Error", "Error al actualizar cuestionario: " + error.message);
    }
};

const handleCreateQuestionnaire = async (questionnaireData) => {
    try {
        const sectionId = route.params.id;
        const userId = props.user.id;
        const token = getToken();

        // Transform data to match composable expectations
        const transformedData = {
            name: questionnaireData.name,
            plugin: questionnaireData.plugin,
            startDate: questionnaireData.startDate,
            endDate: questionnaireData.endDate,
            description: questionnaireData.description,
            fields: questionnaireData.fields || [],
        };

        const result = await groupsModule.createQuestionnaire(
            transformedData,
            sectionId,
            userId,
            token,
        );
        showInfo("Éxito", result.message);
        closeCreateQuestionnaireModal();
    } catch (error) {
        showInfo("Error", "Error al crear cuestionario: " + error.message);
    }
};

const handleDeleteQuestionnaire = (questionnaire) => {
    const questionnaireTitle =
        questionnaire.title || questionnaire.name || "este cuestionario";
    confirmMessage.value = `¿Estás seguro de que deseas eliminar el cuestionario "${questionnaireTitle}"?`;
    pendingAction.value = () => performDeleteQuestionnaire(questionnaire);
    showConfirmModal.value = true;
};

const performDeleteQuestionnaire = async (questionnaire) => {
    try {
        const token = getToken();
        const result = await groupsModule.deleteQuestionnaire(
            questionnaire,
            token,
        );
        showInfo("Éxito", result.message);
    } catch (error) {
        showInfo("Error", "Error al eliminar cuestionario: " + error.message);
    }
};

const addNewGroup = () => {
    groupsModule.addNewGroup();
};

const handleOpenPluginModal = async () => {
    const sectionId = route.params.id;
    const token = getToken();

    // Load responses first before opening the modal
    const testStatus = await groupsModule.loadSectionResponses(sectionId, token);
    if (testStatus) {
        studentTestStatus.value = testStatus;
    }

    // Trigger the modal to open
    openPluginModalTrigger.value = Date.now();
};

const formGroupsByPlugin = async (formConfig) => {
    try {
        const token = getToken();

        // First, delete all existing groups to start fresh
        await groupsModule.deleteAllGroups(token);

        // Now form new groups with the plugin configuration
        const result = await groupsModule.formGroupsByPlugins(token, formConfig);

        // If a Pareto front was returned, show the conflict resolution modal first
        if (groupsModule.paretoFront.value && groupsModule.paretoFront.value.length > 0) {
            conflictCriteriaParameters.value = formConfig.criteriaParameters || {};
            conflictFormConfig.value = formConfig;
            showConflictModal.value = true;
            return;
        }

        // Reload groups with student enrichment
        const sectionId = route.params.id;
        await groupsModule.initializeGroupsData(
            students.value,
            sectionId,
            token,
        );

        const pluginNames = [];
        if (formConfig.psychologicalPlugin) {
            pluginNames.push(
                formConfig.psychologicalPlugin === "enneagram"
                    ? "Afinidad Psicológica"
                    : "MBTI",
            );
        }
        if (formConfig.sociometricPlugin) {
            pluginNames.push("Afinidad Social");
        }

        showInfo(
            "Éxito",
            `Grupos formados exitosamente usando: ${pluginNames.join(" + ")}`,
        );
    } catch (error) {
        console.error("Error forming groups by plugins:", error);
        showInfo("Error", "Error al formar grupos: " + error.message);
    }
};

const handleParetoSelect = async (solution) => {
    try {
        const token = getToken();
        await groupsModule.selectParetoSolution(solution, token);

        const sectionId = route.params.id;
        await groupsModule.initializeGroupsData(
            students.value,
            sectionId,
            token,
        );

        showParetoChart.value = false;
        showInfo("Éxito", "Grupos formados exitosamente desde el Frente de Pareto");
    } catch (error) {
        console.error("Error selecting Pareto solution:", error);
        showInfo("Error", "Error al guardar solución: " + error.message);
    }
};

const handleParetoDismiss = () => {
    groupsModule.dismissParetoFront();
    showParetoChart.value = false;
};

const handleParetoBack = () => {
    showParetoChart.value = false;
    showConflictModal.value = true;
};

const handleShowPareto = () => {
    showConflictModal.value = false;
    showParetoChart.value = true;
};

const handleConflictDismiss = () => {
    showConflictModal.value = false;
    groupsModule.dismissParetoFront();
};

const handleWeightedSolution = async ({ psyWeight, socWeight, solution }) => {
    try {
        showConflictModal.value = false;
        const token = getToken();

        if (solution?.teams) {
            // Use the selected solution from Pareto front
            await groupsModule.selectParetoSolution(solution, token);
        } else {
            // Fallback: select knee point solution
            const pareto = groupsModule.paretoFront.value;
            if (pareto && pareto.length > 0) {
                const wPsy = psyWeight / 100;
                const wSoc = socWeight / 100;
                let bestSolution = null;
                let bestScore = -Infinity;
                for (const s of pareto) {
                    const score = (s.psy_score * wPsy) + (s.soc_score * wSoc);
                    if (score > bestScore) {
                        bestScore = score;
                        bestSolution = s;
                    }
                }
                if (bestSolution) {
                    await groupsModule.selectParetoSolution(bestSolution, token);
                }
            }
        }

        const sectionId = route.params.id;
        await groupsModule.initializeGroupsData(
            students.value,
            sectionId,
            token,
        );

        showInfo("Éxito", `Grupos formados con prioridad Estructural ${psyWeight}% / Relacional ${socWeight}%`);
    } catch (error) {
        console.error("Error applying weighted solution:", error);
        showInfo("Error", "Error al formar grupos: " + error.message);
    }
};

const formRandomGroups = async (studentsPerGroup = 4) => {
    try {
        const token = getToken();
        await groupsModule.formRandomGroups(token, studentsPerGroup);

        // Reload groups with student enrichment
        const sectionId = route.params.id;
        await groupsModule.initializeGroupsData(
            students.value,
            sectionId,
            token,
        );

        showInfo("Éxito", "Grupos formados exitosamente de manera aleatoria");
    } catch (error) {
        console.error("Error forming random groups:", error);
        showInfo(
            "Error",
            "Error al formar grupos aleatorios: " + error.message,
        );
    }
};

const deleteGroup = (groupId) => {
    confirmMessage.value =
        "¿Estás seguro de que deseas eliminar este grupo? Todos los estudiantes serán devueltos al pool disponible.";
    pendingAction.value = () => performDeleteGroup(groupId);
    showConfirmModal.value = true;
};

const performDeleteGroup = async (groupId) => {
    try {
        const token = getToken();
        await groupsModule.deleteGroup(groupId, token);
    } catch (error) {
        console.error("Error deleting group:", error);
        showInfo("Error", "Error al eliminar el grupo: " + error.message);
    }
};

const deleteAllGroups = () => {
    confirmMessage.value =
        "¿Estás seguro de que deseas eliminar TODOS los grupos? Todos los estudiantes serán devueltos al pool disponible.";
    pendingAction.value = () => performDeleteAllGroups();
    showConfirmModal.value = true;
};

const performDeleteAllGroups = async () => {
    try {
        const token = getToken();
        await groupsModule.deleteAllGroups(token);
        showInfo("Éxito", "Todos los grupos han sido eliminados exitosamente");
    } catch (error) {
        console.error("Error deleting all groups:", error);
        showInfo(
            "Error",
            "Error al eliminar todos los grupos: " + error.message,
        );
    }
};

const handleConfirm = () => {
    if (pendingAction.value) {
        pendingAction.value();
        pendingAction.value = null;
    }
    showConfirmModal.value = false;
};

const updateGroupName = (groupId, newName) => {
    groupsModule.updateGroupName(groupId, newName);
};

const exportGroupsToExcel = async () => {
    try {
        const token = getToken();
        const sectionId = route.params.id;
        await groupsModule.exportGroupsToExcel(sectionId, token);
        showInfo("Éxito", "Archivo Excel descargado exitosamente");
    } catch (error) {
        console.error("Error exporting groups to Excel:", error);
        showInfo("Error", "Error al exportar grupos a Excel: " + error.message);
    }
};

const saveGroupName = (group) => {
    const token = getToken();
    groupsModule.saveGroupName(group, token);
};

const onDragStart = (event, student, sourceType, sourceId) => {
    groupsModule.onDragStart(event, student, sourceType, sourceId);
};

const onDropToGroup = (event, targetGroupId) => {
    const token = getToken();
    groupsModule.onDropToGroup(event, targetGroupId, token);
};

const onDropToAvailable = (event) => {
    const token = getToken();
    groupsModule.onDropToAvailable(event, token);
};

const onDragOver = (event) => {
    groupsModule.onDragOver(event);
};

const loadPluginMetadata = async (pluginType) => {
    if (pluginMetadataCache.value[pluginType]) {
        console.log(`[loadPluginMetadata] Cache hit for ${pluginType}`);
        return pluginMetadataCache.value[pluginType];
    }

    console.log(`[loadPluginMetadata] Cache miss for ${pluginType}, fetching from API`);
    const result = await pluginApi.getPluginsExtended();
    
    if (!result.success) {
        console.error(`[loadPluginMetadata] API error for ${pluginType}:`, result.error);
        return null;
    }
    
    if (!result.data) {
        console.error(`[loadPluginMetadata] No data returned for ${pluginType}`);
        return null;
    }
    
    // Handle both array and object response formats
    let pluginsList;
    if (Array.isArray(result.data)) {
        pluginsList = result.data;
    } else if (result.data.data && Array.isArray(result.data.data)) {
        // Nested format: { data: [...] }
        pluginsList = result.data.data;
    } else if (typeof result.data === 'object') {
        // Single object format: wrap in array
        pluginsList = [result.data];
    } else {
        console.error(`[loadPluginMetadata] Unexpected data format for ${pluginType}:`, typeof result.data);
        return null;
    }
    
    console.log(`[loadPluginMetadata] Found ${pluginsList.length} plugins in API response`);
    
    const pluginMeta = pluginsList.find(p => p.nombre_interno === pluginType);
    if (pluginMeta) {
        console.log(`[loadPluginMetadata] Found metadata for ${pluginType}:`, {
            hasFullQuestions: !!pluginMeta.full_questions,
            fullQuestionsLength: pluginMeta.full_questions?.length,
            hasPreguntas: !!pluginMeta.preguntas
        });
        pluginMetadataCache.value[pluginType] = pluginMeta;
        return pluginMeta;
    }
    
    // Fallback: try to find by checking if any plugin has full_questions
    console.warn(`[loadPluginMetadata] Plugin ${pluginType} not found by nombre_interno, checking all plugins`);
    for (const p of pluginsList) {
        console.log(`[loadPluginMetadata] Available plugin:`, p.nombre_interno, {
            hasFullQuestions: !!p.full_questions,
            fullQuestionsLength: p.full_questions?.length
        });
        if (p.full_questions && p.full_questions.length > 0) {
            console.log(`[loadPluginMetadata] Found plugin with full_questions:`, p.nombre_interno);
        }
    }
    
    return null;
};

const openLikertTest = async (questionnaire) => {
    selectedQuestionnaireForTest.value = questionnaire;
    console.log('[openLikertTest] Called with questionnaire:', {
        id: questionnaire.id,
        plugin: questionnaire.plugin,
        title: questionnaire.title
    });

    // group_compatibility uses SelectionTest
    if (questionnaire.plugin === "group_compatibility") {
        selectedQuestionnaireForTest.value = questionnaire;
        showSelectionTest.value = true;
        return;
    }

    // demographic uses FormTest
    if (questionnaire.plugin === "demographic") {
        selectedQuestionnaireForTest.value = questionnaire;
        showFormTest.value = true;
        return;
    }

    // All other plugins (bfi44, mbti, enneagram, chronotype) use LikertTest
    const likertTypes = ["bfi44", "mbti", "enneagram", "chronotype", "bigfive"];
    
    if (likertTypes.includes(questionnaire.plugin)) {
        const pluginToLoad = questionnaire.plugin === "bigfive" ? "bfi44" : questionnaire.plugin;
        
        const metadata = await loadPluginMetadata(pluginToLoad);
        
        // Check if student already responded before opening modal
        const studentId = getUserIdFromToken();
        
        if (studentId && questionnaire.id) {
            const result = await pluginApi.checkStudentTestResponse(questionnaire.plugin, questionnaire.id, studentId);
            
            if (result.success && result.hasResponded) {
                // Extract results if nested
                let resultsData = result.data;
                if (Array.isArray(result.data) && result.data.length > 0) {
                    resultsData = result.data[0];
                }
                testResultsData.value = resultsData?.results || resultsData;
                testResultsPlugin.value = questionnaire.plugin;
                showTestResults.value = true;
                return;
            }
        }
        
        selectedGenericTest.value = questionnaire;
        genericTestMetadata.value = metadata;
        showLikertTest.value = true;
        return;
    }

    // Fallback - shouldn't reach here with current plugins
    console.warn('[openLikertTest] Unknown plugin type:', questionnaire.plugin);
};

// Helper to get user ID from token (same logic as LikertTest.vue)
const getUserIdFromToken = () => {
    try {
        const token = localStorage.getItem('auth_token');
        if (!token) return null;
        const payload = token.split('.')[1];
        const decodedPayload = JSON.parse(atob(payload));
        const userId = decodedPayload.user_id || decodedPayload.id || decodedPayload.sub;
        return userId ? userId.toString() : null;
    } catch (err) {
        console.error('[getUserIdFromToken] Error:', err);
        return null;
    }
};

const closeGenericTest = () => {
    showSelectionTest.value = false;
    showFormTest.value = false;
    showLikertTest.value = false;
    selectedQuestionnaireForTest.value = null;
    selectedGenericTest.value = null;
    genericTestMetadata.value = null;
};

const onResponseSubmitted = () => {
    if (selectedQuestionnaireForTest.value) {
        const qId =
            selectedQuestionnaireForTest.value.id ||
            selectedQuestionnaireForTest.value._id;
        completedQuestionnaireIds.value = new Set([
            ...completedQuestionnaireIds.value,
            qId,
        ]);
    }
    closeGenericTest();
};

const onViewTestResults = (results) => {
    console.log('[onViewTestResults] Showing results:', results);
    testResultsData.value = results?.results || results;
    testResultsPlugin.value = selectedGenericTest.value?.plugin || '';
    showTestResults.value = true;
    closeGenericTest();
};

const closeTestResults = () => {
    showTestResults.value = false;
    testResultsData.value = null;
    testResultsPlugin.value = '';
};

// Format date helper
const formatDate = (dateString) => {
    if (!dateString) return "N/A";
    const date = new Date(dateString);
    return date.toLocaleDateString("es-ES", {
        year: "numeric",
        month: "long",
        day: "numeric",
    });
};

// Get plugin label helper
const getPluginLabel = (questionnaire) => {
    const isNotAdmin = props.user?.role_name !== "administrador";

    if (questionnaire.plugin === "group_compatibility") {
        return "Afinidad Social";
    } else if (questionnaire.plugin === "bfi44") {
        return isNotAdmin
            ? "Afinidad Psicológica 1"
            : "Afinidad Psicológica 1 (BFI-44)";
    } else if (questionnaire.plugin === "mbti") {
        return isNotAdmin
            ? "Afinidad Psicológica 3"
            : "Afinidad Psicológica 3 (MBTI)";
    } else if (questionnaire.plugin === "demographic") {
        return "Dato Demográfico";
    }

    return isNotAdmin
        ? "Afinidad Psicológica 2"
        : "Afinidad Psicológica 2 (Eneagrama)";
};

// Load student's group information
const loadMyGroupInfo = async () => {
    const sectionId = route.params.id;
    const token = getToken();

    try {
        loadingMyGroup.value = true;

        // Cargar información del grupo del estudiante con datos completos
        myGroupInfo.value = await groupsModule.loadMyGroupMates(
            sectionId,
            token,
        );

        // Los datos ya vienen completos desde el backend, no necesitamos filtrar
        if (myGroupInfo.value && myGroupInfo.value.students.length > 0) {
            myGroupMembers.value = myGroupInfo.value.students;
        }
    } catch (error) {
        console.error("Error loading group info:", error);
        myGroupInfo.value = null;
        myGroupMembers.value = [];
    } finally {
        loadingMyGroup.value = false;
    }
};

// Load data on mount
onMounted(async () => {
    const sectionId = route.params.id;
    const token = getToken();

    // Solo cargar datos si no es estudiante
    if (props.user?.role_name === "estudiante") {
        // Para estudiantes, cargar questionnaires y su grupo
        await groupsModule.loadQuestionnaires(sectionId, token);
        await loadMyGroupInfo();
    } else {
        // Load students and groups first (lightweight)
        const groupData = await groupsModule.loadSectionGroupData(sectionId, token);

        if (groupData) {
            students.value = groupData.students || [];
        }

        // Load questionnaires separately (lightweight)
        await groupsModule.loadSectionSurveyData(sectionId);

        // studentTestStatus is now empty - responses are loaded on-demand
        // when the user clicks "Ver Respuestas" in the modal
        studentTestStatus.value = {};
        console.log("Section data loaded (lazy mode): students:", students.value.length);
    }
});
</script>

<style scoped>
/* Student View */
.student-view {
    display: flex;
    flex-direction: column;
    gap: 2rem;
}

/* Questionnaires Section */
.questionnaires-section {
    background: var(--color-surface, #ffffff);
    border-radius: var(--radius-md, 10px);
    box-shadow: var(--shadow-sm, 0 1px 3px rgba(0, 0, 0, 0.08));
}

.questionnaires-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
    padding: 1.5rem;
    border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.questionnaires-header h3 {
    margin: 0;
    color: var(--color-dark, #394049);
    font-size: 1.5rem;
    font-weight: 600;
}

.loading-questionnaires,
.no-questionnaires {
    text-align: center;
    padding: 2rem;
    color: #94a3b8;
}

.no-questionnaires {
    border: 2px dashed var(--color-border, #dfe4ea);
    border-radius: var(--radius-md, 10px);
    margin: 1rem;
}

.questionnaires-container {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1rem;
    padding: 1rem;
}

.questionnaire-card {
    background: var(--color-surface, #ffffff);
    border: 1px solid var(--color-border, #dfe4ea);
    border-radius: var(--radius-md, 10px);
    padding: 1.5rem;
    transition: all 0.2s;
    position: relative;
    overflow: hidden;
}

.questionnaire-card::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 4px;
    background: var(--color-primary, #00a499);
}

.questionnaire-card:hover {
    box-shadow: var(--shadow-md, 0 4px 6px rgba(0, 0, 0, 0.08));
    transform: translateY(-1px);
}

.questionnaire-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1rem;
}

.questionnaire-title {
    margin: 0;
    color: var(--color-dark, #394049);
    font-size: 1.125rem;
    font-weight: 600;
    flex: 1;
}

.questionnaire-plugin {
    background: rgba(0, 164, 153, 0.1);
    color: var(--color-primary-dark, #008a7f);
    padding: 0.25rem 0.75rem;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: 500;
}

.questionnaire-info {
    margin-bottom: 1rem;
}

.questionnaire-description {
    color: var(--color-text-secondary, #64748b);
    font-size: 0.875rem;
    margin: 0 0 1rem 0;
}

.questionnaire-dates {
    display: flex;
    gap: 1.5rem;
    font-size: 0.8125rem;
}

.date-info {
    color: #94a3b8;
}

.date-info strong {
    color: var(--color-text-secondary, #64748b);
    font-weight: 600;
}

.questionnaire-actions {
    display: flex;
    gap: 0.5rem;
    justify-content: flex-end;
}

.respond-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.625rem 1.25rem;
    background-color: var(--color-primary, #00a499);
    color: white;
    border: none;
    border-radius: var(--radius-sm, 6px);
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    transition: all 0.2s;
}

.respond-btn:hover {
    background-color: var(--color-primary-dark, #008a7f);
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

/* My Group Section */
.my-group-section {
    background: var(--color-surface, #ffffff);
    border-radius: var(--radius-md, 10px);
    box-shadow: var(--shadow-sm, 0 1px 3px rgba(0, 0, 0, 0.08));
}

.my-group-header {
    margin-bottom: 1.5rem;
    padding: 1.5rem;
    border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.my-group-header h3 {
    margin: 0;
    color: var(--color-dark, #394049);
    font-size: 1.5rem;
    font-weight: 600;
}

.loading-group {
    text-align: center;
    padding: 2rem;
    color: #94a3b8;
}

.no-group {
    text-align: center;
    padding: 3rem 2rem;
    color: #94a3b8;
}

.no-group svg {
    margin: 0 auto 1rem;
    color: #cbd5e0;
}

.no-group p {
    margin: 0;
    font-size: 1rem;
}

.group-info {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    padding: 1.5rem;
}

.group-name-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.25rem;
    background: linear-gradient(
        135deg,
        var(--color-dark, #394049) 0%,
        #2d3139 100%
    );
    color: white;
    border-radius: 12px;
    font-size: 1.125rem;
    font-weight: 600;
    align-self: flex-start;
    box-shadow: 0 4px 6px rgba(57, 64, 73, 0.25);
}

.group-members {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 1rem;
}

.member-card {
    background: linear-gradient(
        135deg,
        var(--color-background, #f4f6f9) 0%,
        var(--color-surface, #ffffff) 100%
    );
    border: 2px solid var(--color-border, #dfe4ea);
    border-radius: var(--radius-lg, 14px);
    padding: 1.25rem;
    display: flex;
    gap: 1rem;
    transition: all 0.3s ease;
    position: relative;
    overflow: hidden;
}

.member-card::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 4px;
    background: var(--color-primary, #00a499);
}

.member-card:hover {
    transform: translateY(-4px);
    box-shadow: var(--shadow-lg, 0 8px 16px rgba(0, 0, 0, 0.1));
    border-color: #cbd5e0;
}

.member-avatar {
    flex-shrink: 0;
    width: 56px;
    height: 56px;
    background: var(--color-primary, #006b64);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    box-shadow: 0 4px 6px rgba(0, 164, 153, 0.25);
}

.member-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
}

.member-name {
    font-weight: 600;
    color: var(--color-dark, #394049);
    font-size: 1rem;
    line-height: 1.3;
}

.member-email {
    color: #94a3b8;
    font-size: 0.8125rem;
}

.eneatype-badge {
    display: inline-block;
    padding: 0.2rem 0.6rem;
    background: linear-gradient(
        135deg,
        rgba(0, 164, 153, 0.1) 0%,
        rgba(0, 164, 153, 0.2) 100%
    );
    color: var(--color-primary-dark, #008a7f);
    border-radius: 12px;
    font-size: 0.7rem;
    font-weight: 700;
    align-self: flex-start;
    box-shadow: 0 1px 2px rgba(0, 164, 153, 0.1);
}

.eneatype-info {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    margin-top: 0.25rem;
}

.ocean-mini-grid {
    display: flex;
    gap: 0.15rem;
    flex-wrap: wrap;
}

.ocean-mini-val {
    font-size: 0.6rem;
    font-weight: 800;
    padding: 0.05rem 0.2rem;
    border-radius: 3px;
    background: var(--color-background, #f4f6f9);
    color: #475569;
    border-left: 2px solid #94a3b8;
    line-height: 1;
}

.ocean-mini-val.o {
    border-left-color: #ef4444;
}
.ocean-mini-val.c {
    border-left-color: #f59e0b;
}
.ocean-mini-val.e {
    border-left-color: #3b82f6;
}
.ocean-mini-val.a {
    border-left-color: #10b981;
}
.ocean-mini-val.n {
    border-left-color: #8b5cf6;
}

.no-eneatype {
    color: #a0aec0;
    font-size: 0.75rem;
    font-style: italic;
}

/* Responsive Design */
@media (max-width: 768px) {
    .group-members {
        grid-template-columns: 1fr;
    }

    .questionnaire-dates {
        flex-direction: column;
        gap: 0.5rem;
    }

    .evaluation-dates {
        flex-direction: column;
        gap: 0.5rem;
    }

    .evaluation-actions {
        flex-direction: column;
    }

    .action-btn {
        justify-content: center;
    }
}

/* Override action-btn styling for respond-btn */
.action-btn.respond-btn {
    width: auto;
    height: auto;
    padding: 0.625rem 1.25rem;
    background: var(--color-primary, #006b64);
    color: white;
    font-weight: 500;
    font-size: 0.875rem;
    gap: 0.5rem;
    border: none;
    border-radius: var(--radius-sm, 6px);
    cursor: pointer;
    transition: all 0.2s;
}

.action-btn.respond-btn:hover {
    background: var(--color-primary-dark, #008a7f);
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
</style>
