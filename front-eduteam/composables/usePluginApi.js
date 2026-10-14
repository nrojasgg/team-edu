import { useAuth } from "~/composables/useAuth";

export function usePluginApi() {
  const API_BASE_URL = "http://127.0.0.1:3000/api/v1";
  const { authenticatedFetch } = useAuth();

  const endpointPatterns = {
    // Plugins psychometric (generic proxy uses /tests for questionnaires)
    bfi44: { testEndpoint: "tests", responseEndpoint: "tests", sectionResponsesEndpoint: "responses" },
    mbti: { testEndpoint: "tests", responseEndpoint: "tests", sectionResponsesEndpoint: "responses" },
    chronotype: { testEndpoint: "tests", responseEndpoint: "tests", sectionResponsesEndpoint: "responses" },
    enneagram: { testEndpoint: "tests", responseEndpoint: "tests", sectionResponsesEndpoint: "responses" },
    demographic: { testEndpoint: "tests", responseEndpoint: "tests", sectionResponsesEndpoint: "responses" },
    // Plugins with different endpoint patterns
    group_compatibility: { testEndpoint: "tests", responseEndpoint: "tests", sectionResponsesEndpoint: "responses/section" },
    peer_evaluation: { testEndpoint: "surveys", responseEndpoint: "responses", sectionResponsesEndpoint: "responses" },
  };

  function getEndpoint(pluginName, ...parts) {
    return `${API_BASE_URL}/core/plugin/${pluginName}/${parts.join("/")}`;
  }

  function getTestEndpoint(pluginName) {
    const pattern = endpointPatterns[pluginName] || { testEndpoint: "tests" };
    return pattern.testEndpoint;
  }

  function getSectionResponsesEndpoint(pluginName) {
    const pattern = endpointPatterns[pluginName] || {};
    return pattern.sectionResponsesEndpoint || "tests";
  }

  async function getPlugins() {
    try {
      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/plugins/metadata`
      );
      if (!response) return { success: false, error: "Authentication failed" };
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || `HTTP ${response.status}` };
      }
      const data = await response.json();
      return { success: true, data: data.data || data };
    } catch (error) {
      console.error("Error fetching plugins:", error);
      return { success: false, error: error.message };
    }
  }

  async function getPluginsExtended() {
    try {
      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/plugins/metadata/extended`
      );
      if (!response) return { success: false, error: "Authentication failed" };
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || `HTTP ${response.status}` };
      }
      const data = await response.json();
      return { success: true, data: data.data || data };
    } catch (error) {
      console.error("Error fetching extended plugins:", error);
      return { success: false, error: error.message };
    }
  }

  async function createTest(pluginName, testData) {
    try {
      const endpoint = getTestEndpoint(pluginName);
      const response = await authenticatedFetch(
        getEndpoint(pluginName, endpoint),
        {
          method: "POST",
          body: JSON.stringify(testData),
        }
      );
      if (!response) return { success: false, error: "Authentication failed" };
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || `HTTP ${response.status}` };
      }
      const data = await response.json();
      return { success: true, data: data.data || data };
    } catch (error) {
      console.error(`Error creating test for ${pluginName}:`, error);
      return { success: false, error: error.message };
    }
  }

  async function getTests(pluginName, sectionId) {
    try {
      const endpoint = getTestEndpoint(pluginName);
      const url = sectionId
        ? getEndpoint(pluginName, endpoint) + `?section_id=${sectionId}`
        : getEndpoint(pluginName, endpoint);
      const response = await authenticatedFetch(url);
      if (!response) return { success: false, error: "Authentication failed" };
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || `HTTP ${response.status}` };
      }
      const data = await response.json();
      return { success: true, data: data.data || data };
    } catch (error) {
      console.error(`Error fetching tests for ${pluginName}:`, error);
      return { success: false, error: error.message };
    }
  }

  async function getTest(pluginName, testId) {
    try {
      const endpoint = getTestEndpoint(pluginName);
      const response = await authenticatedFetch(
        getEndpoint(pluginName, endpoint, testId)
      );
      if (!response) return { success: false, error: "Authentication failed" };
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || `HTTP ${response.status}` };
      }
      const data = await response.json();
      return { success: true, data: data.data || data };
    } catch (error) {
      console.error(`Error fetching test ${testId} for ${pluginName}:`, error);
      return { success: false, error: error.message };
    }
  }

  async function submitResponse(pluginName, testId, payload, studentId) {
    try {
      const endpoint = getTestEndpoint(pluginName);
      
      // Format payload based on plugin type
      let body;
      if (pluginName === 'group_compatibility') {
        // group_compatibility expects: { question1: [...], question2: [...], student_id }
        body = { ...payload, student_id: studentId };
      } else if (pluginName === 'demographic') {
        // demographic expects: { answers: [...], student_id }
        body = { ...payload, student_id: studentId };
      } else {
        // Standard Likert format: { answers: [...], student_id }
        body = { answers: payload, student_id: studentId };
      }
      
      const response = await authenticatedFetch(
        getEndpoint(pluginName, endpoint, testId, "responses"),
        {
          method: "POST",
          body: JSON.stringify(body),
        }
      );
      if (!response) return { success: false, error: "Authentication failed" };
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || `HTTP ${response.status}` };
      }
      const data = await response.json();
      return { success: true, data: data.data || data };
    } catch (error) {
      console.error(`Error submitting response for ${pluginName}:`, error);
      return { success: false, error: error.message };
    }
  }

  async function getStudentResponse(pluginName, sectionId, studentId) {
    try {
      const url = getEndpoint(pluginName, "student", studentId) +
        (sectionId ? `?section_id=${sectionId}` : "");
      const response = await authenticatedFetch(url);
      if (!response) return { success: false, error: "Authentication failed" };
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || `HTTP ${response.status}` };
      }
      const data = await response.json();
      return { success: true, data: data.data || data };
    } catch (error) {
      console.error(`Error fetching student response for ${pluginName}:`, error);
      return { success: false, error: error.message };
    }
  }

  async function getStudentSurveyResponse(pluginName, surveyId, studentId) {
    try {
      const response = await authenticatedFetch(
        getEndpoint(pluginName, "surveys", surveyId, "responses", studentId)
      );
      if (!response) return { success: false, error: "Authentication failed" };
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || `HTTP ${response.status}` };
      }
      const data = await response.json();
      return { success: true, data: data.data || data };
    } catch (error) {
      console.error(`Error fetching student survey response for ${pluginName}:`, error);
      return { success: false, error: error.message };
    }
  }

  async function getSectionResponses(pluginName, sectionId) {
    try {
      const endpoint = getSectionResponsesEndpoint(pluginName);
      let url;
      if (sectionId) {
        if (pluginName === "group_compatibility") {
          url = getEndpoint(pluginName, endpoint, sectionId.toString());
        } else {
          url = getEndpoint(pluginName, endpoint) + `?section_id=${sectionId}`;
        }
      } else {
        url = getEndpoint(pluginName, endpoint);
      }
      const response = await authenticatedFetch(url);
      if (!response) return { success: false, error: "Authentication failed" };
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || `HTTP ${response.status}` };
      }
      const data = await response.json();
      return { success: true, data: data.data || data };
    } catch (error) {
      console.error(`Error fetching section responses for ${pluginName}:`, error);
      return { success: false, error: error.message };
    }
  }

  async function checkStudentTestResponse(pluginName, testId, studentId) {
    try {
      const url = `${API_BASE_URL}/core/plugin/${pluginName}/tests/${testId}/responses/student/${studentId}`;
      const response = await authenticatedFetch(url);
      if (!response) return { success: false, error: "Authentication failed" };
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || `HTTP ${response.status}` };
      }
      const data = await response.json();
      return { 
        success: true, 
        hasResponded: data.has_responded || false, 
        isComplete: data.is_complete || false,
        data: data.data 
      };
    } catch (error) {
      console.error(`Error checking student test response for ${pluginName}:`, error);
      return { success: false, error: error.message };
    }
  }

  async function deleteTest(pluginName, testId) {
    try {
      const endpoint = getTestEndpoint(pluginName);
      const response = await authenticatedFetch(
        getEndpoint(pluginName, endpoint, testId),
        { method: "DELETE" }
      );
      if (!response) return { success: false, error: "Authentication failed" };
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || `HTTP ${response.status}` };
      }
      return { success: true };
    } catch (error) {
      console.error(`Error deleting test ${testId} for ${pluginName}:`, error);
      return { success: false, error: error.message };
    }
  }

  async function updateTest(pluginName, testId, updateData) {
    try {
      const endpoint = getTestEndpoint(pluginName);
      const response = await authenticatedFetch(
        getEndpoint(pluginName, endpoint, testId),
        {
          method: "PUT",
          body: JSON.stringify(updateData),
        }
      );
      if (!response) return { success: false, error: "Authentication failed" };
      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        return { success: false, error: error.message || `HTTP ${response.status}` };
      }
      const data = await response.json();
      return { success: true, data: data.data || data };
    } catch (error) {
      console.error(`Error updating test ${testId} for ${pluginName}:`, error);
      return { success: false, error: error.message };
    }
  }

  return {
    getPlugins,
    getPluginsExtended,
    createTest,
    getTests,
    getTest,
    submitResponse,
    getStudentResponse,
    getStudentSurveyResponse,
    getSectionResponses,
    checkStudentTestResponse,
    deleteTest,
    updateTest,
  };
}
