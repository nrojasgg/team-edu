import { ref } from "vue";
import { useAuth } from "~/composables/useAuth";

// Default sociometric questions attached to every group_compatibility survey
const DEFAULT_SOCIOMETRIC_QUESTIONS = [
  {
    id: 1,
    text: "¿Con quién preferirías trabajar en el próximo proyecto de equipo?",
  },
  {
    id: 2,
    text: "¿A quién elegirías como representante o líder de tu equipo?",
  },
  {
    id: 3,
    text: "¿Con quién te sientes más cómodo/a colaborando en actividades de clase?",
  },
  {
    id: 4,
    text: "¿Quién crees que aportaría más positivamente a tu equipo de trabajo?",
  },
];

export const useSectionSurveys = () => {
  const API_BASE_URL = "http://127.0.0.1:3000/api/v1";
  const { authenticatedFetch } = useAuth();

  // Small helper to centralize fetch + error handling
  const callApi = async (url, options = {}) => {
    const response = await authenticatedFetch(url, options);
    if (!response) return null;
    if (!response.ok) {
      const errBody = await response.json().catch(() => ({}));
      const message = errBody.error || errBody.message || "API error";
      const e = new Error(message);
      e.status = response.status;
      throw e;
    }
    return await response.json();
  };

  // State
  const peerEvaluationSurveys = ref([]);
  const compatibilitySurveys = ref([]);
  const loadingPeerEvaluationSurveys = ref(false);
  const showCreatePeerEvaluationModal = ref(false);

  /**
   * Load peer evaluation surveys for a section
   * @param {string} sectionId - Section ID
   * @param {string} token - Authentication token
   */
  const loadPeerEvaluationSurveys = async (sectionId, token) => {
    if (!sectionId || !token) {
      console.warn(
        "Section ID and token are required to load peer evaluation surveys",
      );
      peerEvaluationSurveys.value = [];
      return;
    }

    try {
      loadingPeerEvaluationSurveys.value = true;

      // Call ms-core peer evaluation API through API Gateway
      const result = await callApi(
        `${API_BASE_URL}/core/peer-evaluation/surveys?section_id=${sectionId}`,
      );

      // Transform API data to component format
      const now = new Date();
      peerEvaluationSurveys.value = (result?.data || []).map((survey) => {
        const endDate = survey.end_date ? new Date(survey.end_date) : null;
        const startDate = survey.start_date
          ? new Date(survey.start_date)
          : null;
        const isExpired = endDate && endDate < now;
        const notStarted = startDate && startDate > now;
        let status = "active";
        if (!survey.is_active || isExpired) status = "expired";
        else if (notStarted) status = "draft";

        return {
          id: survey.id || survey._id,
          name: survey.title,
          type: "Evaluación entre Pares",
          plugin: "peer_evaluation",
          description: survey.description,
          questions: Array.isArray(survey.questions) ? survey.questions : [],
          group_snapshots: Array.isArray(survey.group_snapshots)
            ? survey.group_snapshots
            : [],
          createdAt: survey.created_at,
          expiresAt: survey.end_date,
          startDate: survey.start_date,
          endDate: survey.end_date,
          status,
          responsesCount: 0,
          totalStudents: 0,
        };
      });
    } catch (error) {
      console.error("Error loading peer evaluation surveys:", error);
      peerEvaluationSurveys.value = [];
    } finally {
      loadingPeerEvaluationSurveys.value = false;
    }
  };

  /**
   * Create a peer evaluation survey
   * @param {Object} surveyData - Survey data
   * @param {string} sectionId - Section ID
   * @param {string} token - Authentication token
   */
  const createPeerEvaluationSurvey = async (surveyData, sectionId, token) => {
    try {
      // Convert dates to ISO 8601 / RFC3339 format
      const startDate = surveyData.start_date
        ? new Date(surveyData.start_date).toISOString()
        : null;
      const endDate = surveyData.end_date
        ? new Date(surveyData.end_date).toISOString()
        : null;

      // Call ms-core peer evaluation API through API Gateway
      const result = await callApi(
        `${API_BASE_URL}/core/peer-evaluation/surveys`,
        {
          method: "POST",
          body: JSON.stringify({
            title: surveyData.title,
            description: surveyData.description,
            section_id: sectionId,
            created_by: token, // User ID derived from token
            start_date: startDate,
            end_date: endDate,
          }),
        },
      );

      // Add the new survey to the local list
      const newSurvey = {
        id: result.data.id || result.data._id,
        name: result.data.title,
        type: "Evaluación entre Pares",
        plugin: "peer_evaluation",
        description: result.data.description,
        createdAt: result.data.created_at,
        expiresAt: result.data.end_date,
        startDate: result.data.start_date,
        endDate: result.data.end_date,
        status: "active",
        responsesCount: 0,
        totalStudents: 0,
      };

      peerEvaluationSurveys.value.unshift(newSurvey);
      closeCreatePeerEvaluationModal();

      return {
        success: true,
        message: "Evaluación entre pares creada exitosamente",
        data: result.data,
      };
    } catch (error) {
      console.error("Error creating peer evaluation survey:", error);
      throw error;
    }
  };

  /**
   * Open create peer evaluation modal
   */
  const openCreatePeerEvaluationModal = () => {
    showCreatePeerEvaluationModal.value = true;
  };

  /**
   * Close create peer evaluation modal
   */
  const closeCreatePeerEvaluationModal = () => {
    showCreatePeerEvaluationModal.value = false;
  };

  /**
   * Get a specific peer evaluation survey
   * @param {string} surveyId - Survey ID
   */
  const getPeerEvaluationSurvey = async (surveyId) => {
    try {
      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/peer-evaluation/surveys/${surveyId}`,
      );

      if (!response) return null;

      if (!response.ok) {
        throw new Error("Error al obtener la evaluación");
      }

      const result = await response.json();
      return result.data;
    } catch (error) {
      console.error("Error getting peer evaluation survey:", error);
      throw error;
    }
  };

  /**
   * Submit a peer evaluation response
   * @param {string} surveyId - Survey ID
   * @param {string} evaluatorId - Evaluator student ID
   * @param {string} evaluatedId - Evaluated student ID
   * @param {Array} responses - Array of responses
   */
  const submitPeerEvaluationResponse = async (
    surveyId,
    evaluatorId,
    evaluatedId,
    groupId,
    responses,
  ) => {
    try {
      const bodyObj = {
        // include both snake_case and camelCase keys to support different backend parsers
        evaluator_id: String(evaluatorId),
        evaluatorId: evaluatorId,
        evaluated_id: String(evaluatedId),
        evaluatedId: evaluatedId,
        group_id: String(groupId),
        groupId: groupId,
        responses: responses,
      };
      const bodyStr = JSON.stringify(bodyObj);
      console.debug(
        "submitPeerEvaluationResponse => POST",
        `${API_BASE_URL}/core/peer-evaluation/surveys/${surveyId}/responses`,
        bodyStr,
      );
      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/peer-evaluation/surveys/${surveyId}/responses`,
        {
          method: "POST",
          body: bodyStr,
        },
      );

      if (!response) return; // Token expirado, ya fue manejado por authenticatedFetch

      if (!response.ok) {
        const errorData = await response.json();
        console.error(
          "submitPeerEvaluationResponse: server returned error",
          response.status,
          errorData,
        );
        // Normalize error message
        let message =
          (errorData && (errorData.error || errorData.message)) ||
          "Error al enviar la respuesta";
        if (typeof message !== "string") {
          try {
            message = JSON.stringify(message);
          } catch (e) {
            console.error("Error stringifying error message:", e);
            message = String(message);
          }
        }
        throw new Error(message);
      }

      const result = await response.json();
      console.log("submitPeerEvaluationResponse: success", result);
      return {
        success: true,
        message: "Respuesta enviada exitosamente",
        data: result.data,
      };
    } catch (error) {
      console.error("Error submitting peer evaluation response:", error);
      throw error;
    }
  };

  /**
   * Get peer evaluation results
   * @param {string} surveyId - Survey ID
   */
  const getPeerEvaluationResults = async (surveyId) => {
    try {
      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/peer-evaluation/surveys/${surveyId}/results`,
      );

      if (!response) return null;

      if (!response.ok) {
        throw new Error("Error al obtener los resultados");
      }

      const result = await response.json();
      return result.data;
    } catch (error) {
      console.error("Error getting peer evaluation results:", error);
      throw error;
    }
  };

  /**
   * Get peer evaluation responses for a survey, optionally filtered by evaluator and evaluated ids
   * @param {string} surveyId
   * @param {string|number} evaluatorId (optional)
   * @param {Array} evaluatedIds (optional) - array of evaluated student ids
   */
  const getPeerEvaluationResponses = async (
    surveyId,
    evaluatorId = null,
    evaluatedIds = null,
  ) => {
    try {
      let url = `${API_BASE_URL}/core/peer-evaluation/surveys/${surveyId}/responses`;
      const params = new URLSearchParams();
      if (evaluatorId) params.append("evaluator_id", evaluatorId);
      if (evaluatedIds && evaluatedIds.length > 0)
        params.append("evaluated_ids", evaluatedIds.join(","));
      if (params.toString()) url += "?" + params.toString();
      console.debug("getPeerEvaluationResponses => GET", url);
      const response = await authenticatedFetch(url);
      if (!response) return null;
      if (!response.ok) {
        const errorData = await response.json();
        console.error(
          "getPeerEvaluationResponses: server error",
          response.status,
          errorData,
        );
        let message =
          (errorData && (errorData.error || errorData.message)) ||
          "Error al obtener respuestas";
        if (typeof message !== "string") {
          try {
            message = JSON.stringify(message);
          } catch (e) {
            console.error("Error stringifying message:", e);
            message = String(message);
          }
        }
        throw new Error(message);
      }
      const result = await response.json();
      return result.data;
    } catch (error) {
      console.error("Error getting peer evaluation responses:", error);
      throw error;
    }
  };

  /**
   * Get a single peer evaluation response submitted by evaluator for evaluated student
   * @param {string} surveyId
   * @param {string|number} evaluatorId
   * @param {string|number} evaluatedId
   */
  const getPeerEvaluationResponse = async (
    surveyId,
    evaluatorId,
    evaluatedId,
  ) => {
    try {
      const url = `${API_BASE_URL}/core/peer-evaluation/surveys/${surveyId}/responses?evaluator_id=${evaluatorId}&evaluated_ids=${evaluatedId}`;
      console.debug("getPeerEvaluationResponse => GET", url);
      const response = await authenticatedFetch(url);

      if (!response) return null;

      if (!response.ok) {
        const errorData = await response.json();
        console.error(
          "getPeerEvaluationResponse: server error",
          response.status,
          errorData,
        );
        let message =
          (errorData && (errorData.error || errorData.message)) ||
          "Error al obtener la respuesta";
        if (typeof message !== "string") {
          try {
            message = JSON.stringify(message);
          } catch (e) {
            console.error("Error stringifying message:", e);
            message = String(message);
          }
        }
        throw new Error(message);
      }

      const result = await response.json();
      return result.data;
    } catch (error) {
      console.error("Error getting peer evaluation response:", error);
      throw error;
    }
  };

  /**
   * View survey responses
   * @param {Object} survey - Survey object
   */
  const viewSurveyResponses = (survey) => {
    console.log("Viewing responses for survey:", survey);
    // Navigate to responses page or open modal (functionality pending)
    return { survey, message: "Respuestas - Funcionalidad pendiente" };
  };

  /**
   * Edit a survey
   * @param {Object} surveyData - Updated survey data
   * @param {string} token - Authentication token
   * @returns {Promise<Object>} Updated survey data
   */
  const editSurvey = async (surveyData, token) => {
    if (!surveyData.id || !token) {
      console.warn("Survey ID and token are required to edit survey");
      return { success: false, error: "Survey ID and token are required" };
    }

    try {
      // Convert dates to ISO 8601 / RFC3339 format
      const startDate = surveyData.start_date
        ? new Date(surveyData.start_date).toISOString()
        : null;
      const endDate = surveyData.end_date
        ? new Date(surveyData.end_date).toISOString()
        : null;

      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/peer-evaluation/surveys/${surveyData.id}`,
        {
          method: "PUT",
          body: JSON.stringify({
            title: surveyData.title,
            description: surveyData.description,
            start_date: startDate,
            end_date: endDate,
            is_active: surveyData.is_active !== false,
          }),
        },
      );

      if (!response) return { success: false, error: "Token expirado" };

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(
          errorData.error ||
            errorData.message ||
            "Error al actualizar la evaluación",
        );
      }

      const result = await response.json();

      // Update survey in local list
      const index = peerEvaluationSurveys.value.findIndex(
        (s) => s.id === surveyData.id,
      );
      if (index >= 0) {
        peerEvaluationSurveys.value[index] = {
          ...peerEvaluationSurveys.value[index],
          name: result.data.title,
          description: result.data.description,
          startDate: result.data.start_date
            ? new Date(result.data.start_date.replace("Z", ""))
            : null,
          endDate: result.data.end_date
            ? new Date(result.data.end_date.replace("Z", ""))
            : null,
          status: result.data.is_active ? "active" : "expired",
        };
      }

      return {
        success: true,
        data: result.data,
        message: "Evaluación actualizada correctamente",
      };
    } catch (error) {
      console.error("Error updating survey:", error);
      return { success: false, error: error.message };
    }
  };

  /**
   * Delete a survey
   * @param {string} surveyId - Survey ID to delete
   * @param {string} token - Authentication token
   * @returns {Promise<Object>} Deletion result
   */
  const deleteSurvey = async (surveyId, token) => {
    if (!surveyId || !token) {
      console.warn("Survey ID and token are required to delete survey");
      return { success: false, error: "Survey ID and token are required" };
    }

    try {
      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/peer-evaluation/surveys/${surveyId}`,
        {
          method: "DELETE",
        },
      );

      if (!response) return { success: false, error: "Token expirado" };

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(
          errorData.error ||
            errorData.message ||
            "Error al eliminar la evaluación",
        );
      }

      await response.json();

      // Remove survey from list
      peerEvaluationSurveys.value = peerEvaluationSurveys.value.filter(
        (s) => s.id !== surveyId,
      );

      return { success: true, message: "Evaluación eliminada correctamente" };
    } catch (error) {
      console.error("Error deleting survey:", error);
      return { success: false, error: error.message };
    }
  };

  /**
   * Get survey status label
   * @param {string} status - Survey status
   * @returns {string} Status label in Spanish
   */
  const getSurveyStatusLabel = (status) => {
    const labels = {
      active: "Activa",
      expired: "Expirada",
      draft: "Borrador",
      closed: "Cerrada",
    };
    return labels[status] || status;
  };

  /**
   * Check if survey is active
   * @param {Object} survey - Survey object
   * @returns {boolean} True if survey is active
   */
  const isSurveyActive = (survey) => {
    if (survey.status === "active") {
      const now = new Date();
      const endDate = new Date(survey.endDate);
      return now < endDate;
    }
    return false;
  };

  /**
   * Get survey completion percentage
   * @param {Object} survey - Survey object
   * @returns {number} Completion percentage
   */
  const getSurveyCompletionPercentage = (survey) => {
    if (!survey.totalStudents || survey.totalStudents === 0) return 0;
    return Math.round((survey.responsesCount / survey.totalStudents) * 100);
  };

  /**
   * Get peer evaluation responses for a specific section and group
   * @param {string} sectionId - Section ID
   * @param {string} groupId - Group ID
   * @param {string} token - Authentication token
   * @returns {Promise<Object>} Response data with success status and responses array
   */
  const getSectionGroupResponses = async (sectionId, groupId, token) => {
    if (!sectionId || !groupId || !token) {
      console.warn(
        "Section ID, Group ID and token are required to get responses",
      );
      return { success: false, data: [] };
    }

    try {
      const response = await authenticatedFetch(
        `${API_BASE_URL}/sections/${sectionId}/responses/${groupId}`,
      );

      if (!response) {
        return { success: false, data: [] }; // Token expirado, ya fue manejado por authenticatedFetch
      }

      if (!response.ok) {
        console.error("Error getting section group responses from API");
        return { success: false, data: [] };
      }

      const result = await response.json();
      return {
        success: result.success || true,
        data: result.data || [],
      };
    } catch (error) {
      console.error("Error getting section group responses:", error);
      return { success: false, data: [] };
    }
  };

  /**
   * Get survey peers from the snapshot saved at survey creation time
   * @param {string} surveyId - Survey ID
   * @param {string} userId - User ID
   * @param {string} token - Authentication token
   * @returns {Promise<Object>} Response with peers array and group information
   */
  const getSurveyPeers = async (surveyId, userId, token) => {
    if (!surveyId || !userId || !token) {
      console.warn(
        "Survey ID, User ID and token are required to get survey peers",
      );
      return { success: false, peers: [], group_id: null, group_name: null };
    }

    try {
      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/peer-evaluation/survey-peers`,
        {
          method: "POST",
          body: JSON.stringify({
            action: "get_survey_peers",
            survey_id: surveyId,
            user_id: userId,
          }),
        },
      );

      if (!response) {
        return { success: false, peers: [], group_id: null, group_name: null }; // Token expirado
      }

      if (!response.ok) {
        console.warn(
          "Survey snapshot peers not found, will fall back to current group",
        );
        return { success: false, peers: [], group_id: null, group_name: null };
      }

      const result = await response.json();
      const peers = result.peers || [];
      if (peers.length === 0) {
        console.warn(
          "getSurveyPeers: empty peers list in snapshot, will fall back to current group",
        );
        return {
          success: false,
          peers: [],
          group_id: result.group_id || null,
          group_name: result.group_name || null,
        };
      }
      return {
        success: true,
        peers,
        group_id: result.group_id || null,
        group_name: result.group_name || null,
      };
    } catch (error) {
      console.error("Error getting survey peers:", error);
      return { success: false, peers: [], group_id: null, group_name: null };
    }
  };

  /**
   * Get current group mates for a student in a section (fallback when snapshot is missing).
   * Uses GET /core/groups/section/:sectionId/my-group-complete via the API Gateway.
   * @param {string} sectionId - Section ID
   * @returns {Promise<{success: boolean, peers: string[], group_id: string|null, group_name: string|null}>}
   */
  const getCurrentGroupPeers = async (sectionId) => {
    if (!sectionId) {
      console.warn("sectionId is required to get current group peers");
      return { success: false, peers: [], group_id: null, group_name: null };
    }

    try {
      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/groups/section/${sectionId}/my-group-complete`,
      );

      if (!response) {
        return { success: false, peers: [], group_id: null, group_name: null };
      }

      if (!response.ok) {
        console.warn("getCurrentGroupPeers: could not fetch current group");
        return { success: false, peers: [], group_id: null, group_name: null };
      }

      const result = await response.json();
      const students = result?.data?.students || [];
      const groupName = result?.data?.group_name || null;

      // students may be full objects or plain IDs depending on gateway enrichment
      const peerIds = students.map((s) =>
        String(typeof s === "object" ? (s.id ?? s.student_id) : s),
      );

      const groupNumber = result?.data?.group_number ?? null;

      return {
        success: peerIds.length > 0,
        peers: peerIds,
        group_id: null, // el endpoint no devuelve un ObjectId directamente
        group_number: groupNumber, // usado por surveys.vue para resolver el group_id del snapshot
        group_name: groupName,
        students, // keep full objects for enriched data
      };
    } catch (error) {
      console.error("Error getting current group peers:", error);
      return { success: false, peers: [], group_id: null, group_name: null };
    }
  };

  /**
   * Load group-compatibility (sociometric) surveys for a section
   */
  const loadCompatibilitySurveys = async (sectionId) => {
    if (!sectionId) return;
    try {
      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/group-compatibility/surveys/section/${sectionId}`,
      );
      if (!response?.ok) {
        compatibilitySurveys.value = [];
        return;
      }
      const result = await response.json();
      const raw = result.data || [];
      compatibilitySurveys.value = raw.map((survey) => ({
        ...survey,
        id: survey.id || survey._id,
        name: survey.title,
        type: "Afinidad Social",
        plugin: "group_compatibility",
        description: survey.description || "",
        createdAt: survey.created_at,
        expiresAt: survey.end_date,
        startDate: survey.start_date,
        endDate: survey.end_date,
        status: survey.is_active ? "active" : "expired",
        // attach default questions so SociometricTest can render them
        questions: DEFAULT_SOCIOMETRIC_QUESTIONS,
      }));
    } catch (e) {
      console.error("Error loading compatibility surveys:", e);
      compatibilitySurveys.value = [];
    }
  };

  return {
    // State
    peerEvaluationSurveys,
    compatibilitySurveys,
    loadingPeerEvaluationSurveys,
    showCreatePeerEvaluationModal,

    // Methods
    loadPeerEvaluationSurveys,
    loadCompatibilitySurveys,
    createPeerEvaluationSurvey,
    openCreatePeerEvaluationModal,
    closeCreatePeerEvaluationModal,
    getPeerEvaluationSurvey,
    submitPeerEvaluationResponse,
    getPeerEvaluationResults,
    getPeerEvaluationResponses,
    viewSurveyResponses,
    editSurvey,
    deleteSurvey,
    getSectionGroupResponses,
    getSurveyPeers,
    getCurrentGroupPeers,
    getPeerEvaluationResponse,

    // Utility methods
    getSurveyStatusLabel,
    isSurveyActive,
    getSurveyCompletionPercentage,
  };
};
