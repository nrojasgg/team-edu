import { ref, reactive } from "vue";
import { useAuth } from "~/composables/useAuth";
import { usePluginApi } from "~/composables/usePluginApi";

export const useSectionGroups = () => {
  const API_BASE_URL = "http://127.0.0.1:3000/api/v1";
  const { authenticatedFetch } = useAuth();
  const pluginApi = usePluginApi();

  // State
  const groups = ref([]);
  const availableStudents = ref([]);
  const draggedStudent = ref(null);
  const globalScore = ref(null);
  const paretoFront = ref(null);
  const paretoResult = ref(null);

  // Holds the last known criteria config and demographic data for re-evaluation after drag & drop.
  // These are populated from last_group_history when the section data is loaded.
  const lastCriteriaParameters = ref(null); // { [criterionName]: { target, weight } }
  const lastDemographicData = ref({}); // { [studentId]: { numeric_attrs, categorical_attrs, social_connections } }
  const lastEnforceEquity = ref(false); // mirrors enforce_equity used in the last form-flexible call
  const lastSectionCandidates = ref([]); // all candidate IDs of the section — used as kGlobal population in evaluate-groups
  // Raw compatibility responses from the sociogram survey, used to build social_connections.
  // Shape: Array<{ student_id: string, question1: [{student_id, response}], question3: [{student_id, response}], ... }>
  const compatibilityResponses = ref([]);

  // Debounce timer reference for evaluateGroups calls triggered by drag & drop.
  let _evaluateDebounceTimer = null;
  const questionnaires = ref([]);
  const loadingQuestionnaires = ref(false);
  const showCreateQuestionnaireModal = ref(false);
  const currentSectionId = ref(null);

  const newQuestionnaire = reactive({
    name: "",
    plugin: "",
    startDate: "",
    endDate: "",
    description: "",
  });

  /**
   * Load groups from API
   * @param {string} sectionId - Section ID
   * @param {string} token - Authentication token
   * @param {Array} allStudents - Optional array of all students to enrich group data and filter available students
   */

  // Helper: enrich a group student with full profile data
  const enrichGroupStudent = (groupStudent, allStudents) => {
    const fullStudent = allStudents.find((s) => s.id === groupStudent.id);
    return fullStudent
      ? { ...fullStudent, dominant_type: groupStudent.dominant_type }
      : groupStudent;
  };

  // Helper: enrich a group student with full profile and track in set
  const enrichGroupStudentWithTracking = (
    groupStudent,
    allStudents,
    trackingSet,
  ) => {
    trackingSet.add(groupStudent.id);
    const fullStudent = allStudents.find((s) => s.id === groupStudent.id);
    return fullStudent
      ? {
          ...fullStudent,
          dominant_type: groupStudent.dominant_type || fullStudent.eneatype,
        }
      : groupStudent;
  };

  // Helper: process questionnaire API response and normalize to array with plugin field
  const processQuestionnairesResponse = async (response, plugin) => {
    if (!response?.ok) return [];
    const responseData = await response.json();
    let data = [];
    if (responseData.success && responseData.data) {
      data = responseData.data;
    } else if (Array.isArray(responseData)) {
      data = responseData;
    }
    return data.map((q) => ({ ...q, plugin }));
  };

  const loadGroups = async (sectionId, token, allStudents = null) => {
    // NOTE: loadGroups is the lightweight loader (groups panel only).
    // It does not receive targets from the API record directly.
    // Full data including targets is loaded by loadCompleteSectionGroupsData.
    try {
      currentSectionId.value = sectionId;
      console.log(
        "Loading groups for section:",
        sectionId,
        "with students:",
        allStudents?.length || 0,
      );

      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/groups/section/${sectionId}`,
      );

      if (!response) return; // Token expirado, ya fue manejado por authenticatedFetch

      console.log("Groups API response status:", response.status);

      if (response.ok) {
        const responseData = await response.json();
        console.log("Groups API response data:", responseData);

        if (responseData.success && responseData.data) {
          // Organize data into groups
          const groupsMap = new Map();

          responseData.data.forEach((item) => {
            if (!groupsMap.has(item.group_number)) {
              groupsMap.set(item.group_number, {
                id: item.group_number,
                name: item.group_name,
                number: item.group_number,
                affinityScore: item.affinity_score ?? null,
                providerScores: item.provider_scores ?? null,
                algorithm: item.algorithm ?? null,
                weights: item.weights ?? null,
                targets: item.targets ?? null,
                criterionBreakdown: item.criterion_breakdown ?? null,
                students: [],
              });
            }

            // Add student to group (only ID and dominant_type initially)
            groupsMap.get(item.group_number).students.push({
              id: item.student_id,
              dominant_type: item.dominant_type,
            });
          });

          groups.value = Array.from(groupsMap.values());
          console.log("Loaded groups:", groups.value.length);

          // If allStudents provided, enrich group data and filter available students
          if (allStudents && allStudents.length > 0) {
            // Collect all student IDs that are in groups
            const studentsInGroups = new Set();
            groups.value.forEach((group) => {
              group.students.forEach((student) => {
                studentsInGroups.add(student.id);
              });
            });

            console.log("Students in groups:", studentsInGroups.size);

            // Build a map of existing ocean_vector data from availableStudents
            const oceanVectorMap = new Map();
            availableStudents.value.forEach((s) => {
              if (s.ocean_vector) {
                oceanVectorMap.set(s.id, s.ocean_vector);
              }
            });

            // Filter available students (those not in any group)
            availableStudents.value = allStudents.filter(
              (s) => !studentsInGroups.has(s.id),
            );

            // Restore ocean_vector to availableStudents after filter
            availableStudents.value.forEach((s) => {
              if (oceanVectorMap.has(s.id)) {
                s.ocean_vector = oceanVectorMap.get(s.id);
              }
            });

            console.log("Available students:", availableStudents.value.length);

            // Enrich group students with full student data (merging ocean_vector)
            groups.value.forEach((group) => {
              group.students = group.students.map((gs) => {
                const fullStudent = allStudents.find((s) => s.id === gs.id);
                if (fullStudent) {
                  // Preserve ocean_vector if available in availableStudents but not in allStudents
                  const ov = oceanVectorMap.get(gs.id);
                  return {
                    ...fullStudent,
                    dominant_type: gs.dominant_type || fullStudent.eneatype,
                    ...(ov && !fullStudent.ocean_vector ? { ocean_vector: ov } : {}),
                  };
                }
                return gs;
              });
            });
          }
        } else {
          groups.value = [];
          console.log("No groups data");
          // If no groups, all students are available
          if (allStudents && allStudents.length > 0) {
            availableStudents.value = [...allStudents];
            console.log(
              "No groups, all students available:",
              availableStudents.value.length,
            );
          }
        }
      }
    } catch (error) {
      console.error("Error loading groups:", error);
      groups.value = [];
    }
  };

  /**
   * Save group assignment to API
   * @param {number} sectionId - Section ID
   * @param {number} groupNumber - Group number
   * @param {string} groupName - Group name
   * @param {number} studentId - Student ID
   * @param {string} dominantType - Dominant enneagram type
   * @param {string} token - Authentication token
   * @param {Object} [affinityMeta] - Optional: { affinityScore, providerScores, algorithm, weights }
   */
  const saveGroupAssignment = async (
    sectionId,
    groupNumber,
    groupName,
    studentId,
    dominantType,
    token,
    affinityMeta = null,
  ) => {
    try {
      // Validate studentId
      if (!studentId || studentId === 0) {
        console.error("Invalid student ID:", studentId);
        throw new Error("ID de estudiante inválido");
      }

      const payload = {
        section_id: Number.parseInt(sectionId),
        group_number: Number.parseInt(groupNumber),
        group_name: groupName || `Grupo ${groupNumber}`,
        student_id: Number.parseInt(studentId),
        dominant_type: dominantType || "",
      };
      if (affinityMeta) {
        if (affinityMeta.affinityScore != null)
          payload.affinity_score = affinityMeta.affinityScore;
        if (affinityMeta.providerScores)
          payload.provider_scores = affinityMeta.providerScores;
        if (affinityMeta.algorithm) payload.algorithm = affinityMeta.algorithm;
        if (affinityMeta.weights) payload.weights = affinityMeta.weights;
        if (affinityMeta.targets) payload.targets = affinityMeta.targets;
        if (affinityMeta.criterionBreakdown)
          payload.criterion_breakdown = affinityMeta.criterionBreakdown;
      }

      console.log("Sending payload to API:", JSON.stringify(payload, null, 2));

      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/groups/`,
        {
          method: "POST",
          body: JSON.stringify(payload),
        },
      );

      if (!response) return; // Token expirado, ya fue manejado por authenticatedFetch

      if (!response.ok) {
        throw new Error("Failed to save group assignment");
      }

      return await response.json();
    } catch (error) {
      console.error("Error saving group assignment:", error);
      throw error;
    }
  };

  /**
   * Remove student from group in API
   * @param {number} sectionId - Section ID
   * @param {number} studentId - Student ID
   * @param {string} token - Authentication token
   */
  const removeStudentFromGroup = async (sectionId, studentId, token) => {
    try {
      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/groups/section/${sectionId}/student/${studentId}`,
        {
          method: "DELETE",
        },
      );

      if (!response) return; // Token expirado, ya fue manejado por authenticatedFetch

      if (!response.ok) {
        throw new Error("Failed to remove student from group");
      }

      return await response.json();
    } catch (error) {
      console.error("Error removing student from group:", error);
      throw error;
    }
  };

  /**
   * Initialize groups data from students list
   * @param {Array} students - Array of students
   * @param {string} sectionId - Section ID
   * @param {string} token - Authentication token
   */
  const initializeGroupsData = async (students, sectionId, token) => {
    console.log(
      "Initializing groups data with students:",
      students?.length || 0,
      "for section:",
      sectionId,
    );
    // Load existing groups from API and automatically filter/enrich with students data
    if (sectionId && token) {
      await loadGroups(sectionId, token, students);
    } else {
      // If no API call, just set all students as available
      availableStudents.value = [...students];
      console.log(
        "No API call, setting all students as available:",
        availableStudents.value.length,
      );
    }
  };

  /**
   * Add a new empty group
   */
  const addNewGroup = () => {
    const newGroupNumber = groups.value.length + 1;
    groups.value.push({
      id: newGroupNumber,
      number: newGroupNumber,
      name: `Grupo ${newGroupNumber}`,
      students: [],
    });
  };

  /**
   * Delete a group and return its students to available pool
   * @param {number} groupId - Group ID to delete
   * @param {string} token - Authentication token
   */
  const deleteGroup = async (groupId, token) => {
    const groupIndex = groups.value.findIndex((g) => g.id === groupId);
    if (groupIndex !== -1) {
      const group = groups.value[groupIndex];
      const studentsToReturn = [...group.students];

      try {
        if (!currentSectionId.value) {
          throw new Error("No se ha podido identificar la sección actual");
        }

        // Remove all students from this group in the API
        const promises = studentsToReturn.map((student) =>
          removeStudentFromGroup(currentSectionId.value, student.id, token),
        );
        await Promise.all(promises);

        // Update local state
        availableStudents.value.push(...studentsToReturn);
        groups.value.splice(groupIndex, 1);
      } catch (error) {
        console.error("Error deleting group:", error);
        throw error;
      }
    }
  };

  /**
   * Delete all groups and return all students to available pool
   * @param {string} token - Authentication token
   */
  const deleteAllGroups = async (token) => {
    try {
      if (!currentSectionId.value) {
        throw new Error("No se ha podido identificar la sección actual");
      }

      console.log("Deleting all groups for section:", currentSectionId.value);

      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/groups/section/${currentSectionId.value}`,
        {
          method: "DELETE",
        },
      );

      if (!response) return; // Token expirado

      if (!response.ok) {
        throw new Error("Error al eliminar los grupos en el servidor");
      }

      // Collect all students from all groups to return them to available pool locally
      const allStudentsToReturn = [];
      groups.value.forEach((group) => {
        allStudentsToReturn.push(...group.students);
      });

      // Update local state
      availableStudents.value.push(...allStudentsToReturn);
      groups.value = [];
    } catch (error) {
      console.error("Error deleting all groups:", error);
      throw error;
    }
  };

  /**
   * Update group name and save to API for all students in the group
   * @param {Object} group - Group object
   * @param {string} token - Authentication token
   */
  const saveGroupName = async (group, token) => {
    try {
      // Update all students in this group with the new name
      const promises = group.students.map((student) =>
        saveGroupAssignment(
          currentSectionId.value,
          group.number || group.id,
          group.name,
          student.id,
          student.eneatype || student.dominant_type || "",
          token,
        ),
      );

      await Promise.all(promises);
    } catch (error) {
      console.error("Error saving group name:", error);
      alert("Error al guardar el nombre del grupo");
    }
  };

  /**
   * Update group name locally
   * @param {number} groupId - Group ID
   * @param {string} newName - New group name
   */
  const updateGroupName = (groupId, newName) => {
    const group = groups.value.find((g) => g.id === groupId);
    if (group) {
      group.name = newName;
    }
  };

  /**
   * Handle drag start event
   * @param {Event} event - Drag event
   * @param {Object} student - Student being dragged
   * @param {string} sourceType - Source type ('available' or 'group')
   * @param {number|null} sourceId - Source group ID (if applicable)
   */
  const onDragStart = (event, student, sourceType, sourceId = null) => {
    draggedStudent.value = {
      student,
      sourceType,
      sourceId,
    };
    event.dataTransfer.effectAllowed = "move";
  };

  /**
   * Handle drag over event
   * @param {Event} event - Drag event
   */
  const onDragOver = (event) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
  };

  /**
   * Handle drop to group event
   * @param {Event} event - Drop event
   * @param {number} targetGroupId - Target group ID
   * @param {string} token - Authentication token
   */
  const onDropToGroup = async (event, targetGroupId, token) => {
    event.preventDefault();
    if (!draggedStudent.value) return;

    const { student, sourceType, sourceId } = draggedStudent.value;
    const targetGroup = groups.value.find((g) => g.id === targetGroupId);

    if (!targetGroup) return;

    // Don't do anything if dropping on the same group
    if (sourceType === "group" && sourceId === targetGroupId) {
      draggedStudent.value = null;
      return;
    }

    // Validate student has a valid ID
    if (!student || !student.id || student.id === 0) {
      console.error("Invalid student object or ID:", student);
      alert("Error: El estudiante no tiene un ID válido");
      draggedStudent.value = null;
      return;
    }

    try {
      // Debug: Log student object to see its structure
      console.log("Student being dropped:", student);
      console.log("Student ID:", student.id);

      // Save to API
      await saveGroupAssignment(
        currentSectionId.value,
        Number.parseInt(targetGroup.number || targetGroup.id),
        targetGroup.name || `Grupo ${targetGroup.number || targetGroup.id}`,
        student.id,
        String(student.eneatype || student.dominant_type || ""),
        token,
      );

      // Remove student from source
      if (sourceType === "available") {
        const index = availableStudents.value.findIndex(
          (s) => s.id === student.id,
        );
        if (index !== -1) {
          availableStudents.value.splice(index, 1);
        }
      } else if (sourceType === "group") {
        const sourceGroup = groups.value.find((g) => g.id === sourceId);
        if (sourceGroup) {
          const index = sourceGroup.students.findIndex(
            (s) => s.id === student.id,
          );
          if (index !== -1) {
            sourceGroup.students.splice(index, 1);
          }
        }
      }

      // Add student to target group
      targetGroup.students.push(student);
    } catch (error) {
      console.error("Error moving student to group:", error);
      alert("Error al mover el estudiante al grupo");
    }

    draggedStudent.value = null;
    scheduleEvaluation(token);
  };

  /**
   * Handle drag over event
   * Handle drop to available pool event
   * @param {Event} event - Drop event
   * @param {string} token - Authentication token
   */
  /**
   * Recalculate affinity scores for all current groups using the evaluate-groups endpoint.
   * This is a pure-compute call — nothing is persisted. It uses the criteria from the
   * last algorithm run (stored in lastCriteriaParameters) so the scores stay comparable.
   * Safe to call after every drag & drop (uses debounce internally).
   * @param {string} token - Authentication token
   */
  const evaluateGroups = async (token) => {
    if (
      !lastCriteriaParameters.value ||
      Object.keys(lastCriteriaParameters.value).length === 0
    ) {
      // No criteria available — scores cannot be recalculated.
      return;
    }

    if (!currentSectionId.value) return;

    const groupsPayload = groups.value
      .filter((g) => g.students && g.students.length > 0)
      .map((g) => ({
        group_id: String(g.id),
        member_ids: g.students.map((s) => String(s.id)),
      }));

    if (groupsPayload.length === 0) return;

    // Merge cached demographic data with any students that may have been added
    // after initial load (e.g. moved from available list).
    const demographicData = { ...lastDemographicData.value };
    const allCurrentStudents = [
      ...availableStudents.value,
      ...groups.value.flatMap((g) => g.students),
    ];
    const socialMapForEval = buildSocialConnectionsMap(
      compatibilityResponses.value,
    );
    allCurrentStudents.forEach((s) => {
      const sid = String(s.id);
      if (!demographicData[sid]) {
        demographicData[sid] = buildStudentDemographicEntry(
          s,
          socialMapForEval,
        );
      }
    });

    // section_candidates: full population used to compute kGlobal (Blau denominator).
    // When available, use the cached list from the last formation call, which includes
    // every student in the section — not just those currently assigned to groups.
    // Fallback: union of grouped + available students visible right now.
    const sectionCandidates =
      lastSectionCandidates.value && lastSectionCandidates.value.length > 0
        ? lastSectionCandidates.value
        : [
            ...new Set([
              ...Object.keys(demographicData),
              ...groupsPayload.flatMap((g) => g.member_ids),
            ]),
          ];

    const payload = {
      section_id: String(currentSectionId.value),
      groups: groupsPayload,
      criteria_parameters: lastCriteriaParameters.value,
      demographic_data: demographicData,
      section_candidates: sectionCandidates,
      enforce_equity: !!lastEnforceEquity.value,
      equity_strategy: lastEnforceEquity.value ? "minimax" : "",
    };

    try {
      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/team-formation/evaluate-groups`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        },
      );

      if (!response || !response.ok) return;

      const responseData = await response.json();
      if (!responseData.success || !responseData.data) return;

      const { groups: evaluatedGroups, global_score } = responseData.data;
      if (global_score != null) globalScore.value = global_score;

      // Patch the reactive groups with the freshly computed scores immediately.
      evaluatedGroups.forEach((evaluated) => {
        const group = groups.value.find(
          (g) => String(g.id) === evaluated.group_id,
        );
        if (group) {
          group.affinityScore = evaluated.affinity_score;
          group.providerScores = evaluated.provider_scores;
          if (evaluated.criterion_breakdown) {
            group.criterionBreakdown = evaluated.criterion_breakdown;
          }
          // Preserve targets so the detail modal can show "Cumplido / No cumplido"
          // after a drag-and-drop re-evaluation. Targets come from the cached
          // criteria parameters and don't change between evaluations.
          if (
            !group.targets &&
            lastCriteriaParameters.value &&
            Object.keys(lastCriteriaParameters.value).length > 0
          ) {
            const rebuiltTargets = {};
            for (const [name, cfg] of Object.entries(
              lastCriteriaParameters.value,
            )) {
              rebuiltTargets[name] = cfg.target ?? 0.5;
            }
            group.targets = rebuiltTargets;
          }
        }
      });

      // Persist scores to MongoDB in the background — fire and forget.
      // UI is already updated above; this call must not block the user.
      persistGroupScores(evaluatedGroups).catch((err) =>
        console.warn("persistGroupScores background error:", err),
      );
    } catch (err) {
      // Non-fatal: scores simply won't update until the next successful call.
      console.warn("evaluateGroups failed:", err);
    }
  };

  /**
   * Persists recalculated group scores to MongoDB via PATCH so they survive a page reload.
   * Only updates affinity_score and provider_scores — does not touch group composition.
   * @param {Array} evaluatedGroups - Array of { group_id, affinity_score, provider_scores }
   */
  const persistGroupScores = async (evaluatedGroups) => {
    if (
      !currentSectionId.value ||
      !evaluatedGroups ||
      evaluatedGroups.length === 0
    )
      return;

    try {
      await authenticatedFetch(
        `${API_BASE_URL}/core/groups/section/${currentSectionId.value}/scores`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            groups: evaluatedGroups.map((g) => ({
              group_id: g.group_id,
              affinity_score: g.affinity_score,
              provider_scores: g.provider_scores,
              ...(g.criterion_breakdown
                ? { criterion_breakdown: g.criterion_breakdown }
                : {}),
            })),
          }),
        },
      );
    } catch (err) {
      console.warn("persistGroupScores failed:", err);
    }
  };

  /**
   * Schedules a debounced call to evaluateGroups (fires 800ms after the last drag & drop).
   * @param {string} token - Authentication token
   */
  const scheduleEvaluation = (token) => {
    if (_evaluateDebounceTimer) clearTimeout(_evaluateDebounceTimer);
    _evaluateDebounceTimer = setTimeout(() => evaluateGroups(token), 300);
  };

  const onDropToAvailable = async (event, token) => {
    event.preventDefault();
    if (!draggedStudent.value) return;

    const { student, sourceType, sourceId } = draggedStudent.value;

    // Only allow dropping from groups
    if (sourceType !== "group") {
      draggedStudent.value = null;
      return;
    }

    try {
      // Remove from API
      await removeStudentFromGroup(currentSectionId.value, student.id, token);

      const sourceGroup = groups.value.find((g) => g.id === sourceId);
      if (sourceGroup) {
        const index = sourceGroup.students.findIndex(
          (s) => s.id === student.id,
        );
        if (index !== -1) {
          sourceGroup.students.splice(index, 1);
        }
      }

      availableStudents.value.push(student);
    } catch (error) {
      console.error("Error removing student from group:", error);
      alert("Error al remover el estudiante del grupo");
    }

    draggedStudent.value = null;
    scheduleEvaluation(token);
  };

  /**
   * Load questionnaires for a section
   * @param {string} sectionId - Section ID
   * @param {string} token - Authentication token
   */
  const loadQuestionnaires = async (sectionId, token) => {
    try {
      loadingQuestionnaires.value = true;

      const pluginsResult = await pluginApi.getPlugins();
      if (!pluginsResult.success) {
        console.warn("Failed to load plugins:", pluginsResult.error);
        questionnaires.value = [];
        return;
      }

      const plugins = pluginsResult.data || [];
      const allQuestionnaires = [];

      for (const plugin of plugins) {
        if (!plugin.activo) continue;

        const testsResult = await pluginApi.getTests(plugin.nombre_interno, sectionId);
        if (testsResult.success && Array.isArray(testsResult.data)) {
          const mapped = testsResult.data.map((q) => ({ ...q, plugin: plugin.nombre_interno }));
          allQuestionnaires.push(...mapped);
        }
      }

      questionnaires.value = allQuestionnaires;
    } catch (error) {
      console.error("Error loading questionnaires:", error);
      questionnaires.value = [];
    } finally {
      loadingQuestionnaires.value = false;
    }
  };

  /**
   * Fetches heavy response arrays lazily when the group-formation
   * algorithm runs (not on initial page load).
   * Uses individual endpoints to load only responses.
   * Returns studentTestStatus object for use by UI components.
   */
  let _lastResponsesCache = { sectionId: null, timestamp: 0, status: null };

  const loadSectionResponses = async (sectionId, token, forceRefresh = false) => {
    try {
      const now = Date.now();
      if (!forceRefresh && _lastResponsesCache.sectionId === sectionId && (now - _lastResponsesCache.timestamp) < 60000) {
        console.log(`[Cache HIT] Using cached section responses (age: ${now - _lastResponsesCache.timestamp}ms)`);
        return _lastResponsesCache.status;
      }

      console.log("Loading section responses for algorithm...");

      const pluginsResult = await pluginApi.getPlugins();
      if (!pluginsResult.success) {
        console.warn("Failed to load plugins:", pluginsResult.error);
        return null;
      }

      const plugins = pluginsResult.data || [];
      const allResponses = {};
      const responsePromises = [];

      for (const plugin of plugins) {
        if (!plugin.activo) continue;

        const promise = pluginApi.getSectionResponses(plugin.nombre_interno, sectionId)
          .then((result) => {
            if (result.success && Array.isArray(result.data)) {
              allResponses[plugin.nombre_interno] = result.data;
            } else {
              allResponses[plugin.nombre_interno] = [];
            }
          })
          .catch(() => {
            allResponses[plugin.nombre_interno] = [];
          });
        responsePromises.push(promise);
      }

      await Promise.all(responsePromises);

      const enneagramResponses = allResponses.enneagram || [];
      const bfi44Responses = allResponses.bfi44 || [];
      const mbtiResponses = allResponses.mbti || [];
      const compatibilityResponsesData = allResponses.group_compatibility || [];
      const demographicResponses = allResponses.demographic || [];

      // Include both pools: available students and students already assigned to groups.
      // In lazy mode, OCEAN responses were previously applied only to available students,
      // leaving grouped students without visible OCEAN data in detail tables.
      const studentRefs = [];
      availableStudents.value.forEach((s) => {
        if (s?.id != null) studentRefs.push(s);
      });
      groups.value.forEach((g) => {
        (g.students || []).forEach((s) => {
          if (s?.id != null) studentRefs.push(s);
        });
      });

      const studentsById = new Map();
      studentRefs.forEach((s) => {
        const sid = String(s.id);
        if (!studentsById.has(sid)) studentsById.set(sid, s);
      });

      const students = Array.from(studentsById.values());

      // Build studentTestStatus like the original loadCompleteSectionGroupsData did
      const testStatus = {};

      // Initialize with existing eneatype data from students
      students.forEach((s) => {
        if (s.id) {
          testStatus[s.id] = {
            hasEnneagram: !!(s.eneatype && s.eneatype > 0),
            hasBFI44: false,
            hasMBTI: false,
            hasSociogram: false,
            demographicByTestId: {},
          };
        }
      });

      // Enrich with enneagram responses
      enneagramResponses.forEach((resp) => {
        const sid = String(resp.student_id);
        if (testStatus[sid]) {
          testStatus[sid].hasEnneagram = true;
        } else if (testStatus[Number(resp.student_id)]) {
          testStatus[Number(resp.student_id)].hasEnneagram = true;
        } else {
          for (const s of students) {
            if (String(s.id) === sid || s.id === resp.student_id) {
              testStatus[s.id].hasEnneagram = true;
              break;
            }
          }
        }
      });

      // Enrich with BFI-44 responses (also counts as personality test)
      bfi44Responses.forEach((resp) => {
        const sid = String(resp.student_id);
        // Try multiple matching formats
        if (testStatus[sid]) {
          testStatus[sid].hasEnneagram = true;
          testStatus[sid].hasBFI44 = true;
        } else if (testStatus[Number(resp.student_id)]) {
          testStatus[Number(resp.student_id)].hasEnneagram = true;
          testStatus[Number(resp.student_id)].hasBFI44 = true;
        } else {
          // Try matching against student id in availableStudents
          for (const s of students) {
            if (String(s.id) === sid || s.id === resp.student_id) {
              testStatus[s.id].hasEnneagram = true;
              testStatus[s.id].hasBFI44 = true;
              break;
            }
          }
        }
      });

      // Enrich with MBTI responses (also counts as personality test)
      mbtiResponses.forEach((resp) => {
        const sid = String(resp.student_id);
        if (testStatus[sid]) {
          testStatus[sid].hasEnneagram = true;
          testStatus[sid].hasMBTI = true;
        } else if (testStatus[Number(resp.student_id)]) {
          testStatus[Number(resp.student_id)].hasEnneagram = true;
          testStatus[Number(resp.student_id)].hasMBTI = true;
        } else {
          for (const s of students) {
            if (String(s.id) === sid || s.id === resp.student_id) {
              testStatus[s.id].hasEnneagram = true;
              testStatus[s.id].hasMBTI = true;
              break;
            }
          }
        }
      });

      // Enrich with sociometric/compatibility responses
      compatibilityResponsesData.forEach((resp) => {
        const sid = resp.student_id != null ? String(resp.student_id) : null;
        if (!sid) return;
        const entry = testStatus[sid] || testStatus[Number(sid)];
        if (entry) {
          entry.hasSociogram = true;
        }
      });

      // Enrich with demographic responses
      demographicResponses.forEach((resp) => {
        const sid = resp.student_id != null ? String(resp.student_id) : null;
        const rawTestId = resp.test_id;
        const testId = rawTestId && typeof rawTestId === "object" && rawTestId.$oid
            ? rawTestId.$oid
            : rawTestId != null ? String(rawTestId) : null;

        if (!sid || !testId) return;

        const entry = testStatus[sid] || testStatus[Number(sid)];
        if (entry) {
          entry.demographicByTestId[testId] = true;
        }
      });

      // Build ocean_vector persona map from responses
      const personaMap = new Map();

      bfi44Responses.forEach((r) => {
        if (r.results?.ocean_vector) {
          personaMap.set(String(r.student_id), r.results.ocean_vector);
        }
      });

      enneagramResponses.forEach((r) => {
        const sid = String(r.student_id);
        if (r.results?.ocean_vector && !personaMap.has(sid)) {
          personaMap.set(sid, r.results.ocean_vector);
        }
      });

      mbtiResponses.forEach((r) => {
        const sid = String(r.student_id);
        if (r.results?.ocean_vector && !personaMap.has(sid)) {
          personaMap.set(sid, r.results.ocean_vector);
        }
      });

      // Apply ocean vectors to every in-memory reference (available + grouped)
      studentRefs.forEach((s) => {
        const sid = String(s.id);
        if (personaMap.has(sid)) {
          s.ocean_vector = personaMap.get(sid);
        }
      });

      compatibilityResponses.value = compatibilityResponsesData || [];
      const socialMap = buildSocialConnectionsMap(compatibilityResponses.value);
      const demographicCache = {};
      students.forEach((s) => {
        demographicCache[String(s.id)] = buildStudentDemographicEntry(s, socialMap);
      });
      lastDemographicData.value = demographicCache;

      console.log("Lazy responses loaded:", {
        bfi44: bfi44Responses.length,
        enneagram: enneagramResponses.length,
        mbti: mbtiResponses.length,
        compatibility: compatibilityResponses.value.length,
        demographic: demographicResponses.length,
      });

      _lastResponsesCache = { sectionId, timestamp: Date.now(), status: testStatus };
      return testStatus;
    } catch (error) {
      console.error("Error loading section responses:", error);
      return null;
    }
  };

  /**
   * Load section group data (students + groups) - lightweight initial load
   * @param {string} sectionId - Section ID
   * @param {string} token - Authentication token
   */
  const loadSectionGroupData = async (sectionId, token) => {
    try {
      currentSectionId.value = sectionId;
      console.log("Loading section group data for section:", sectionId);

      const response = await authenticatedFetch(
        `${API_BASE_URL}/sections/${sectionId}/group-data`,
      );

      if (!response) return null;

      if (!response.ok) {
        throw new Error(`Failed to load group data: ${response.status}`);
      }

      const data = await response.json();
      const allStudents = data.students || [];

      // Process groups
      if (data.groups && data.groups.length > 0) {
        const groupsMap = new Map();

        data.groups.forEach((item) => {
          if (!groupsMap.has(item.group_number)) {
            groupsMap.set(item.group_number, {
              id: item.group_number,
              name: item.group_name,
              number: item.group_number,
              affinityScore: item.affinity_score ?? null,
              providerScores: item.provider_scores ?? null,
              algorithm: item.algorithm ?? null,
              weights: item.weights ?? null,
              targets: item.targets ?? null,
              criterionBreakdown: item.criterion_breakdown ?? null,
              students: [],
            });
          }

          groupsMap.get(item.group_number).students.push({
            id: item.student_id,
            dominant_type: item.dominant_type,
          });
        });

        const processedGroups = Array.from(groupsMap.values());

        // Enrich groups with full student data and collect IDs in groups
        const studentsInGroups = new Set();
        processedGroups.forEach((group) => {
          group.students = group.students.map((gs) =>
            enrichGroupStudentWithTracking(gs, allStudents, studentsInGroups),
          );
        });

        groups.value = processedGroups;
        availableStudents.value = allStudents.filter(
          (s) => !studentsInGroups.has(s.id),
        );
      } else {
        groups.value = [];
        availableStudents.value = [...allStudents];
      }

      // Update lastSectionCandidates if not already set
      if (lastSectionCandidates.value.length === 0 && allStudents.length > 0) {
        lastSectionCandidates.value = allStudents.map((s) => String(s.id));
      }

      console.log("Loaded section group data:", {
        students: allStudents.length,
        groups: groups.value.length,
        available: availableStudents.value.length,
      });

      return {
        students: allStudents,
        groups: groups.value,
      };
    } catch (error) {
      console.error("Error loading section group data:", error);
      return null;
    }
  };

  /**
   * Load section survey data (questionnaires only) - lightweight
   * @param {string} sectionId - Section ID
   */
  const loadSectionSurveyData = async (sectionId) => {
    try {
      console.log("Loading section survey data for section:", sectionId);

      const response = await authenticatedFetch(
        `${API_BASE_URL}/sections/${sectionId}/survey-data`,
      );

      if (!response || !response.ok) {
        throw new Error(`Failed to load survey data: ${response.status}`);
      }

      const data = await response.json();

      // Build questionnaires array from all sources
      const allQuestionnaires = [];

      if (data.enneagram_questionnaires?.length > 0) {
        allQuestionnaires.push(
          ...data.enneagram_questionnaires.map((q) => ({
            ...q,
            plugin: "enneagram",
          })),
        );
      }

      if (data.bfi44_questionnaires?.length > 0) {
        allQuestionnaires.push(
          ...data.bfi44_questionnaires.map((q) => ({
            ...q,
            plugin: "bfi44",
          })),
        );
      }

      if (data.mbti_questionnaires?.length > 0) {
        allQuestionnaires.push(
          ...data.mbti_questionnaires.map((q) => ({
            ...q,
            plugin: "mbti",
          })),
        );
      }

      if (data.compatibility_surveys?.length > 0) {
        allQuestionnaires.push(
          ...data.compatibility_surveys.map((s) => ({
            ...s,
            plugin: "group_compatibility",
          })),
        );
      }

      if (data.demographic_tests?.length > 0) {
        allQuestionnaires.push(
          ...data.demographic_tests.map((t) => ({
            ...t,
            plugin: "demographic",
          })),
        );
      }

      questionnaires.value = allQuestionnaires;
      console.log("Loaded section survey data:", {
        questionnaires: questionnaires.value.length,
      });

      return { questionnaires: allQuestionnaires };
    } catch (error) {
      console.error("Error loading section survey data:", error);
      return null;
    }
  };

  /**
   * Load complete section groups data in a single consolidated request
   * This replaces multiple separate API calls with a single optimized endpoint
   * @param {string} sectionId - Section ID
   * @param {string} token - Authentication token
   * @returns {Promise<Object>} Complete section groups data
   */
  const loadCompleteSectionGroupsData = async (sectionId, token) => {
    try {
      currentSectionId.value = sectionId;
      console.log(
        "Loading complete section groups data for section:",
        sectionId,
      );

      const response = await authenticatedFetch(
        `${API_BASE_URL}/sections/${sectionId}/complete-groups-data?exclude_responses=true`,
      );

      if (!response) return null; // Token expirado, ya fue manejado por authenticatedFetch

      if (!response.ok) {
        throw new Error(
          `Failed to load complete section data: ${response.status}`,
        );
      }

      const responseData = await response.json();

      if (!responseData.success) {
        console.warn("No data in complete section response");
        return null;
      }

      const data = responseData;

      // 1. Process all students first to have them ready for enrichment
      const allStudents = data.students || [];

      // Enrich students with their OCEAN vectors from the responses
      const personaMap = new Map();
      const bfiResponses = data.bfi44_responses || [];
      const enneagramResponses = data.enneagram_responses || [];
      const mbtiResponses = data.mbti_responses || [];

      // BFI-44 is the most reliable source for OCEAN
      bfiResponses.forEach((r) => {
        if (r.results?.ocean_vector)
          personaMap.set(r.student_id.toString(), r.results.ocean_vector);
      });

      // Fallback to Enneagram if not in BFI
      enneagramResponses.forEach((r) => {
        const sid = r.student_id.toString();
        if (r.results?.ocean_vector && !personaMap.has(sid)) {
          personaMap.set(sid, r.results.ocean_vector);
        }
      });

      // Fallback to MBTI if not in BFI/Enneagram
      mbtiResponses.forEach((r) => {
        const sid = r.student_id.toString();
        if (r.results?.ocean_vector && !personaMap.has(sid)) {
          personaMap.set(sid, r.results.ocean_vector);
        }
      });

      // Apply vectors to student objects
      allStudents.forEach((s) => {
        const sid = s.id.toString();
        if (personaMap.has(sid)) {
          s.ocean_vector = personaMap.get(sid);
        }
      });

      // 2. Organize data into groups (grouping the flat list)
      if (data.groups && data.groups.length > 0) {
        const groupsMap = new Map();

        data.groups.forEach((item) => {
          if (!groupsMap.has(item.group_number)) {
            groupsMap.set(item.group_number, {
              id: item.group_number,
              name: item.group_name,
              number: item.group_number,
              affinityScore: item.affinity_score ?? null,
              providerScores: item.provider_scores ?? null,
              algorithm: item.algorithm ?? null,
              weights: item.weights ?? null,
              targets: item.targets ?? null,
              criterionBreakdown: item.criterion_breakdown ?? null,
              students: [],
            });
          }

          // Add student to group (initially just what we have from group record)
          // We will enrich it later if we have the full students list
          groupsMap.get(item.group_number).students.push({
            id: item.student_id,
            dominant_type: item.dominant_type,
          });
        });

        const processedGroups = Array.from(groupsMap.values());

        // 3. Enrich groups with full student data and collect IDs in groups
        const studentsInGroups = new Set();
        processedGroups.forEach((group) => {
          group.students = group.students.map((gs) =>
            enrichGroupStudentWithTracking(gs, allStudents, studentsInGroups),
          );
        });

        // Enrich groups with last_group_history data when stored scores are missing.
        // The history record cross-references by member_ids (strings) so we match
        // each group's student IDs against each team's member list.
        const history = data.last_group_history ?? null;

        // Cache criteria parameters and demographic data from history for later re-evaluation.
        // IMPORTANT: only update lastCriteriaParameters from history if it is not already
        // populated from the direct algorithm response. saveGroupHistory runs in a background
        // goroutine — when loadCompleteSectionGroupsData is called right after form-flexible,
        // the history record may not yet be in MongoDB (or it's the previous one). Overwriting
        // here would discard the correct criteria that were just set from the
        // live response, replacing them with an incomplete or stale history record.
        if (history) {
          const alreadyHasCriteria =
            lastCriteriaParameters.value &&
            Object.keys(lastCriteriaParameters.value).length > 0;

          // Restore enforce_equity from history when not already set from a live response.
          if (!alreadyHasCriteria && history.enforce_equity != null) {
            lastEnforceEquity.value = !!history.enforce_equity;
          }

          if (!alreadyHasCriteria) {
            if (
              history.criteria_parameters &&
              Object.keys(history.criteria_parameters).length > 0
            ) {
              lastCriteriaParameters.value = { ...history.criteria_parameters };
            } else if (
              history.weights &&
              Object.keys(history.weights).length > 0
            ) {
              const rebuildCriteria = {};
              for (const [name, weight] of Object.entries(history.weights)) {
                rebuildCriteria[name] = {
                  weight,
                  target: history.targets?.[name] ?? 0.5,
                };
              }
              lastCriteriaParameters.value = rebuildCriteria;
            }
          }

          // Restore global score from history record
          if (history.global_score != null && history.global_score > 0) {
            globalScore.value = history.global_score;
          }
        }

        if (history && Array.isArray(history.teams)) {
          processedGroups.forEach((group) => {
            // Only enrich if the group doesn't already have provider scores saved
            if (
              group.providerScores &&
              Object.keys(group.providerScores).length > 0
            )
              return;

            const groupStudentIds = new Set(
              group.students.map((s) => String(s.id)),
            );

            // Find the history team whose members overlap the most with this group
            let bestTeam = null;
            let bestOverlap = 0;
            for (const team of history.teams) {
              const memberIds = team.member_ids ?? [];
              const overlap = memberIds.filter((id) =>
                groupStudentIds.has(String(id)),
              ).length;
              if (overlap > bestOverlap) {
                bestOverlap = overlap;
                bestTeam = team;
              }
            }

            if (bestTeam && bestOverlap > 0) {
              group.affinityScore =
                bestTeam.affinity_score ?? group.affinityScore;

              // Guard against corrupt legacy data: if every provider_score value
              // exactly matches its corresponding weight, the history was saved
              // before the fix and contains raw weights instead of real scores.
              if (bestTeam.criterion_breakdown) {
                group.criterionBreakdown = bestTeam.criterion_breakdown;
              }
              const historyProviderScores = bestTeam.provider_scores ?? null;
              const historyWeights = history.weights ?? {};
              const isCorruptScores =
                historyProviderScores &&
                Object.keys(historyProviderScores).length > 0 &&
                Object.entries(historyProviderScores).every(
                  ([key, val]) => historyWeights[key] === val,
                );
              if (!isCorruptScores) {
                group.providerScores =
                  historyProviderScores ?? group.providerScores;
              }

              group.algorithm = history.algorithm ?? group.algorithm;
              group.weights = history.weights ?? group.weights;
              // Targets: prefer per-group saved value, then history-level targets
              group.targets =
                group.targets ?? bestTeam.targets ?? history.targets ?? null;
            }
          });
        }

        groups.value = processedGroups;
        console.log("Processed groups:", groups.value.length);

        // Fallback: if history had no global_score, compute it as average of group affinity scores
        if (globalScore.value == null) {
          const scored = processedGroups.filter((g) => g.affinityScore != null);
          if (scored.length > 0) {
            globalScore.value =
              scored.reduce((sum, g) => sum + g.affinityScore, 0) /
              scored.length;
          }
        }

        // 4. Update available students
        availableStudents.value = allStudents.filter(
          (s) => !studentsInGroups.has(s.id),
        );
        console.log(
          "Available students updated:",
          availableStudents.value.length,
        );
      } else {
        groups.value = [];
        availableStudents.value = [...allStudents];
        console.log(
          "No groups, all students available:",
          availableStudents.value.length,
        );
      }

      // Store raw compatibility responses so social_connections can be built
      // on demand when constructing form-flexible or evaluate-groups payloads.
      // This must run outside the groups if/else so it executes even when there
      // are no groups yet (e.g. first time the teacher opens the section).
      compatibilityResponses.value = data.compatibility_responses || [];

      // Build social connections map from the loaded sociogram responses.
      const socialMap = buildSocialConnectionsMap(compatibilityResponses.value);

      // Build demographic data cache from all known students.
      // Done here (outside if/else) so the cache is always populated regardless
      // of whether groups exist.
      const demographicCache = {};
      allStudents.forEach((s) => {
        demographicCache[String(s.id)] = buildStudentDemographicEntry(
          s,
          socialMap,
        );
      });
      lastDemographicData.value = demographicCache;

      // Restore the full section candidate list from the student roster so that
      // evaluate-groups can use the correct kGlobal population for Blau scoring.
      // Only overwrite when not already populated from a live form-flexible response.
      if (lastSectionCandidates.value.length === 0 && allStudents.length > 0) {
        lastSectionCandidates.value = allStudents.map((s) => String(s.id));
      }

      // Update questionnaires (combine enneagram, BFI-44, and sociometric)
      const allQuestionnaires = [];

      if (
        data.enneagram_questionnaires &&
        data.enneagram_questionnaires.length > 0
      ) {
        const enneagramData = data.enneagram_questionnaires.map((q) => ({
          ...q,
          plugin: "enneagram",
        }));
        allQuestionnaires.push(...enneagramData);
      }

      if (data.bfi44_questionnaires && data.bfi44_questionnaires.length > 0) {
        const bfi44Data = data.bfi44_questionnaires.map((q) => ({
          ...q,
          plugin: "bfi44",
        }));
        allQuestionnaires.push(...bfi44Data);
      }

      if (data.mbti_questionnaires && data.mbti_questionnaires.length > 0) {
        const mbtiData = data.mbti_questionnaires.map((q) => ({
          ...q,
          plugin: "mbti",
        }));
        allQuestionnaires.push(...mbtiData);
      }

      if (data.compatibility_surveys && data.compatibility_surveys.length > 0) {
        const compatibilityData = data.compatibility_surveys.map((s) => ({
          ...s,
          plugin: "group_compatibility",
        }));
        allQuestionnaires.push(...compatibilityData);
      }

      if (data.demographic_tests && data.demographic_tests.length > 0) {
        const demographicData = data.demographic_tests.map((t) => ({
          ...t,
          plugin: "demographic",
        }));
        allQuestionnaires.push(...demographicData);
      }

      questionnaires.value = allQuestionnaires;
      console.log("Loaded questionnaires:", questionnaires.value.length);

      return {
        groups: groups.value,
        availableStudents: availableStudents.value,
        questionnaires: questionnaires.value,
        students: allStudents,
        enneagram_responses: data.enneagram_responses,
        bfi44_responses: data.bfi44_responses,
        mbti_responses: data.mbti_responses,
        compatibility_responses: data.compatibility_responses,
        demographic_responses: data.demographic_responses,
        last_group_history: data.last_group_history ?? null,
      };
    } catch (error) {
      console.error("Error loading complete section data:", error);
      return null;
    }
  };

  /**
   * Open create questionnaire modal
   */
  const openCreateQuestionnaireModal = () => {
    newQuestionnaire.name = "";
    newQuestionnaire.startDate = "";
    newQuestionnaire.endDate = "";
    newQuestionnaire.description = "";
    showCreateQuestionnaireModal.value = true;
  };

  /**
   * Close create questionnaire modal
   */
  const closeCreateQuestionnaireModal = () => {
    showCreateQuestionnaireModal.value = false;
  };

  /**
   * Create a new questionnaire
   * @param {string} sectionId - Section ID
   * @param {number} userId - User ID creating the questionnaire
   * @param {string} token - Authentication token
   */
  const createQuestionnaire = async (
    questionnaireData,
    sectionId,
    userId,
    token,
  ) => {
    try {
      if (!questionnaireData.name.trim()) {
        throw new Error("El nombre del cuestionario es obligatorio");
      }

      if (!questionnaireData.startDate || !questionnaireData.endDate) {
        throw new Error("Las fechas de inicio y término son obligatorias");
      }

      const startDate = new Date(questionnaireData.startDate);
      const endDate = new Date(questionnaireData.endDate);

      if (startDate >= endDate) {
        throw new Error(
          "La fecha de inicio debe ser anterior a la fecha de término",
        );
      }

      const requestBody = {
        title: questionnaireData.name.trim(),
        plugin: questionnaireData.plugin,
        start_date: questionnaireData.startDate,
        end_date: questionnaireData.endDate,
        description: questionnaireData.description.trim(),
        section_id: sectionId,
        created_by: userId.toString(),
      };

      if (questionnaireData.plugin === "group_compatibility") {
        requestBody.survey_type = "full_section";
      } else if (questionnaireData.plugin === "demographic") {
        requestBody.fields = questionnaireData.fields || [];
      }

      const result = await pluginApi.createTest(questionnaireData.plugin, requestBody);

      if (!result.success) {
        throw new Error(result.error || "Error creating questionnaire");
      }

      await loadQuestionnaires(sectionId, token);
      return { success: true, message: "Cuestionario creado exitosamente" };
    } catch (error) {
      console.error("Error creating questionnaire:", error);
      throw error;
    }
  };

  /**
   * Delete a questionnaire
   * @param {Object} questionnaire - Questionnaire to delete
   * @param {string} token - Authentication token
   */
  const deleteQuestionnaire = async (questionnaire, token) => {
    try {
      const result = await pluginApi.deleteTest(questionnaire.plugin, questionnaire.id);

      if (!result.success) {
        throw new Error(result.error || "Error deleting questionnaire");
      }

      questionnaires.value = questionnaires.value.filter(
        (q) => q.id !== questionnaire.id,
      );
      return {
        success: true,
        message: "Cuestionario eliminado exitosamente",
      };
    } catch (error) {
      console.error("Error deleting questionnaire:", error);
      throw error;
    }
  };

  /**
   * Update a questionnaire
   * @param {Object} updatedData - Updated questionnaire data
   * @param {string} token - Authentication token
   */
  const updateQuestionnaire = async (updatedData, token) => {
    try {
      if (!updatedData.title?.trim()) {
        throw new Error("El título del cuestionario es obligatorio");
      }

      if (!updatedData.start_date || !updatedData.end_date) {
        throw new Error("Las fechas de inicio y término son obligatorias");
      }

      const requestBody = {
        title: updatedData.title.trim(),
        description: updatedData.description?.trim() || "",
        start_date: updatedData.start_date,
        end_date: updatedData.end_date,
      };

      const result = await pluginApi.updateTest(updatedData.plugin, updatedData.id, requestBody);

      if (!result.success) {
        throw new Error(result.error || "Error updating questionnaire");
      }

      return {
        success: true,
        message: "Cuestionario actualizado exitosamente",
      };
    } catch (error) {
      console.error("Error updating questionnaire:", error);
      throw error;
    }
  };

  /**
   * Load student's group mates for a specific section with complete information
   * Uses the authenticated student's ID from the JWT token
   * Returns full student details (name, email, eneatype, etc.)
   * @param {string} sectionId - Section ID
   * @param {string} token - Authentication token
   * @returns {Promise<Object>} Object with group_name and students array (with full details)
   */
  const loadMyGroupMates = async (sectionId, token) => {
    try {
      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/groups/section/${sectionId}/my-group-complete`,
      );

      if (!response) return; // Token expirado, ya fue manejado por authenticatedFetch

      if (!response.ok) {
        if (response.status === 404) {
          console.warn("Student not assigned to any group in this section");
          return { group_name: null, students: [] };
        }
        throw new Error("Failed to load group mates");
      }

      const responseData = await response.json();

      if (responseData.success && responseData.data) {
        return responseData.data;
      } else {
        return { group_name: null, students: [] };
      }
    } catch (error) {
      console.error("Error loading group mates:", error);
      return { group_name: null, students: [] };
    }
  };

  /**
   * Load all students in a section
   * @param {string} sectionId - Section ID
   * @returns {Promise<Array>} Array of students with id and name
   */
  const loadSectionStudents = async (sectionId) => {
    try {
      // Use the optimized students endpoint from the API Gateway
      const response = await authenticatedFetch(
        `${API_BASE_URL}/sections/${sectionId}/students-complete`,
      );

      if (!response) return []; // Token expirado

      if (!response.ok) {
        console.warn(
          `Failed to load students (${response.status}). Using empty list.`,
        );
        // Return empty list - the modal will still work, just without the names
        return [];
      }

      const responseData = await response.json();

      // Extract students from the response
      let studentsList = [];
      if (responseData.success && responseData.students) {
        studentsList = responseData.students;
      } else if (responseData.students) {
        studentsList = responseData.students;
      } else if (Array.isArray(responseData)) {
        studentsList = responseData;
      }

      // Map to expected format with id, name, and picture_url
      return studentsList.map((student) => ({
        id: student.id?.toString() || "",
        name: formatStudentName(student),
        picture_url: student.picture_url || null,
      }));
    } catch (error) {
      console.error("Error loading section students:", error);
      return [];
    }
  };

  // Helper to format student name
  const formatStudentName = (student) => {
    if (!student) return "Estudiante";

    const nombres = student.nombres || "";
    const paterno = student.paterno || "";
    const materno = student.materno || "";

    if (nombres && (paterno || materno)) {
      return `${nombres} ${paterno}${materno ? " " + materno : ""}`.trim();
    } else if (student.name) {
      return student.name;
    } else if (student.student_name) {
      return student.student_name;
    }

    return "Estudiante";
  };

  /**
   * Submit a sociometric (group compatibility) response
   * @param {Object} responseData - Response data with answers
   * @returns {Promise<Object>} Result of submission
   */
  const submitSociometricResponse = async (responseData) => {
    try {
      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/group-compatibility/surveys/${responseData.survey_id}/responses`,
        {
          method: "POST",
          body: JSON.stringify(responseData),
        },
      );

      if (!response) return null; // Token expirado

      if (response.ok) {
        const result = await response.json();
        return result;
      } else {
        const errorData = await response.json();
        throw new Error(
          errorData.message ||
            errorData.error_message ||
            String(errorData.error) ||
            response.statusText,
        );
      }
    } catch (error) {
      console.error("Error submitting sociometric response:", error);
      throw error;
    }
  };

  /**
   * Form groups based on enneagram types
   * @param {string} token - Authentication token
   */
  const formGroupsByEnneagram = async (token) => {
    try {
      if (!currentSectionId.value) {
        throw new Error("No section ID available");
      }

      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/groups/section/${currentSectionId.value}/form-by-enneagram`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
          },
        },
      );

      if (!response) return; // Token expirado, ya fue manejado por authenticatedFetch

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(
          errorData.message || "Error al formar grupos por eneagrama",
        );
      }

      const result = await response.json();

      // Reload groups to reflect changes
      await loadGroups(currentSectionId.value, token);

      return result;
    } catch (error) {
      console.error("Error in formGroupsByEnneagram:", error);
      throw error;
    }
  };

  /**
   * Form random groups
   * @param {string} token - Authentication token
   * @param {number} studentsPerGroup - Number of students per group
   */
  const formRandomGroups = async (token, studentsPerGroup = 4) => {
    try {
      if (!currentSectionId.value) {
        throw new Error("No section ID available");
      }

      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/groups/section/${currentSectionId.value}/form-random`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({
            students_per_group: studentsPerGroup,
          }),
        },
      );

      if (!response) return; // Token expirado, ya fue manejado por authenticatedFetch

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(
          errorData.message || "Error al formar grupos aleatorios",
        );
      }

      const result = await response.json();

      // Reload groups to reflect changes
      await loadGroups(currentSectionId.value, token);

      return result;
    } catch (error) {
      console.error("Error in formRandomGroups:", error);
      throw error;
    }
  };

  /**
   * Form groups using the flexible team formation endpoint
   * @param {string} token - Authentication token
   * @param {Object} formConfig - Configuration for team formation
   * @param {string} formConfig.psychologicalPlugin - 'enneagram' | 'mbti' | ''
   * @param {string} formConfig.sociometricPlugin - 'sociogram' | ''
   * @param {number} formConfig.teamSize - Number of students per team
   * @param {number} formConfig.psychologicalWeight - Weight for psychological plugin (0-1)
   * @param {number} formConfig.sociometricWeight - Weight for sociometric plugin (0-1)
   * @param {string[]} formConfig.eligibleStudentIds - Array of student IDs eligible for grouping
   */

  // Helper: extract criteria parameters from form config
  const buildCriteriaParameters = (formConfig) => {
    const criteriaParameters = {};

    if (formConfig.psychologicalPlugin) {
      criteriaParameters[formConfig.psychologicalPlugin] = {
        target: 0.5,
        weight: formConfig.psychologicalWeight,
        type: formConfig.psychologicalType || "categorical",
        label: formConfig.psychologicalLabel || formConfig.psychologicalPlugin,
      };
    }
    if (formConfig.sociometricPlugin) {
      criteriaParameters["social"] = {
        target: 0.5,
        weight: formConfig.sociometricWeight,
        type: "social",
        label: "Afinidad Social",
      };
    }

    if (formConfig.additionalCriteria) {
      for (const [criterionId, criterionConfig] of Object.entries(
        formConfig.additionalCriteria,
      )) {
        criteriaParameters[criterionId] = {
          target: criterionConfig.target,
          weight: criterionConfig.weight,
          goal: criterionConfig.goal || "",
          type: criterionConfig.type || "",
          label: criterionConfig.label || criterionConfig.name || criterionId,
        };
      }
    }

    return criteriaParameters;
  };

  // Helper: build a social connections map from raw compatibility responses.
  // Returns { [voterId]: { [peerId]: score } } where score is:
  //   +1  for question1 answers (¿A quién elegirías para trabajar en equipo?)
  //   -1  for question3 answers (¿A quién NO elegirías?)
  //   0   is omitted (neutral / no data)
  // Only question1 and question3 are used because they represent the voter's
  // explicit outgoing preferences, which map directly to social_connections.
  const buildSocialConnectionsMap = (responses) => {
    const socialMap = {}; // { voterId: { peerId: score } }

    if (!Array.isArray(responses)) return socialMap;

    for (const resp of responses) {
      const voterId = String(resp.student_id);
      if (!socialMap[voterId]) socialMap[voterId] = {};

      // question1: positive choices (+1)
      if (Array.isArray(resp.question1)) {
        for (const answer of resp.question1) {
          if (answer.student_id != null && answer.response !== 0) {
            const peerId = String(answer.student_id);
            // response is already -1/0/1; for question1 positive choices use +1
            socialMap[voterId][peerId] =
              answer.response > 0 ? 1 : answer.response;
          }
        }
      }

      // question3: negative choices (-1)
      if (Array.isArray(resp.question3)) {
        for (const answer of resp.question3) {
          if (answer.student_id != null && answer.response !== 0) {
            const peerId = String(answer.student_id);
            // response is already -1/0/1; for question3 rejections use -1
            socialMap[voterId][peerId] =
              answer.response < 0 ? -1 : -answer.response;
          }
        }
      }
    }

    return socialMap;
  };

  // Helper: extract demographic attributes for a single student.
  // socialConnectionsMap is an optional pre-built map { voterId: { peerId: score } }
  // produced by buildSocialConnectionsMap(); when provided the student's outgoing
  // social preferences are included in social_connections so the social_affinity
  // criterion can evaluate team cohesion during the genetic algorithm.
  const buildStudentDemographicEntry = (
    student,
    socialConnectionsMap = null,
  ) => {
    const numericAttrs = {};
    const categoricalAttrs = {};
    // OCEAN values live in student.ocean_vector (populated by loadCompleteSectionGroupsData).
    // Direct fields like student.openness are never set, so we must read from ocean_vector.
    const ocean = student.ocean_vector || {};
    const readOcean = (longKey, shortKey) => {
      const raw =
        ocean[longKey] ??
        ocean[shortKey] ??
        ocean[shortKey.toUpperCase()] ??
        ocean[shortKey.toLowerCase()];
      const parsed = Number(raw);
      return Number.isFinite(parsed) ? parsed : null;
    };

    const openness = readOcean("openness", "o");
    const conscientiousness = readOcean("conscientiousness", "c");
    const extraversion = readOcean("extraversion", "e");
    const agreeableness = readOcean("agreeableness", "a");
    const neuroticism = readOcean("neuroticism", "n");

    if (openness != null) numericAttrs.openness = openness;
    if (conscientiousness != null)
      numericAttrs.conscientiousness = conscientiousness;
    if (extraversion != null) numericAttrs.extraversion = extraversion;
    if (agreeableness != null) numericAttrs.agreeableness = agreeableness;
    if (neuroticism != null) numericAttrs.neuroticism = neuroticism;
    if (student.sexo) categoricalAttrs.sex = student.sexo;
    const careerVal = student.career_code || student.carrera;
    if (careerVal) categoricalAttrs.career_code = careerVal;

    // Populate social_connections from the pre-built sociogram map when available.
    const sid = String(student.id);
    const socialConnections =
      socialConnectionsMap && socialConnectionsMap[sid]
        ? { ...socialConnectionsMap[sid] }
        : {};

    return {
      numeric_attrs: numericAttrs,
      categorical_attrs: categoricalAttrs,
      social_connections: socialConnections,
    };
  };

  // Helper: find student by ID among available students and groups
  const findStudentById = (studentId) =>
    availableStudents.value.find((s) => s.id.toString() === studentId) ||
    groups.value
      .flatMap((g) => g.students)
      .find((s) => s.id.toString() === studentId);

  // Helper: build NEW FORMAT request body with criteria_parameters and demographic_data
  const buildNewFormatRequestBody = (formConfig, sectionId) => {
    const criteriaParameters = buildCriteriaParameters(formConfig);
    const demographicData = {};

    // Prefer the pre-built demographic cache (populated by loadCompleteSectionGroupsData
    // with social_connections already injected) so we don't rebuild the social map on
    // every call. Fall back to a fresh build for any student not yet in the cache.
    const cachedDemo = lastDemographicData.value || {};
    const socialMap = buildSocialConnectionsMap(compatibilityResponses.value);

    for (const studentId of formConfig.eligibleStudentIds) {
      const sid = String(studentId);
      if (cachedDemo[sid]) {
        demographicData[sid] = cachedDemo[sid];
      } else {
        const student = findStudentById(sid);
        if (student) {
          demographicData[sid] = buildStudentDemographicEntry(
            student,
            socialMap,
          );
        }
      }
    }

    return {
      section_id: sectionId.toString(),
      candidates: formConfig.eligibleStudentIds,
      team_size: formConfig.teamSize,
      min_team_size: formConfig.minTeamSize || formConfig.teamSize - 1,
      max_team_size: formConfig.maxTeamSize || formConfig.teamSize + 1,
      algorithm: "genetic",
      enforce_equity: !!formConfig.enforceEquity,
      equity_strategy: formConfig.enforceEquity ? "minimax" : "",
      criteria_parameters: criteriaParameters,
      demographic_data: demographicData,
    };
  };

  // Helper: build LEGACY FORMAT request body with provider_config
  const buildLegacyFormatRequestBody = (formConfig, sectionId) => {
    const providerConfig = {
      personality_test: formConfig.psychologicalPlugin || "",
      use_sociogram: !!formConfig.sociometricPlugin,
    };
    if (formConfig.psychologicalPlugin) {
      providerConfig.personality_weight = formConfig.psychologicalWeight;
    }
    if (formConfig.sociometricPlugin) {
      providerConfig.sociogram_weight = formConfig.sociometricWeight;
    }
    return {
      section_id: sectionId.toString(),
      candidates: formConfig.eligibleStudentIds,
      team_size: formConfig.teamSize,
      min_team_size: formConfig.minTeamSize || formConfig.teamSize - 1,
      max_team_size: formConfig.maxTeamSize || formConfig.teamSize + 1,
      algorithm: "genetic",
      provider_config: providerConfig,
    };
  };

  const formGroupsByPlugins = async (token, formConfig) => {
    try {
      if (!currentSectionId.value) {
        throw new Error("No section ID available");
      }

      // Ensure heavy response data is loaded before running the algorithm
      await loadSectionResponses(currentSectionId.value, token);

      const useNewFormat =
        formConfig.additionalCriteria &&
        Object.keys(formConfig.additionalCriteria).length > 0;

      const requestBody = useNewFormat
        ? buildNewFormatRequestBody(formConfig, currentSectionId.value)
        : buildLegacyFormatRequestBody(formConfig, currentSectionId.value);

      console.log(
        "Forming groups with config:",
        JSON.stringify(requestBody, null, 2),
      );

      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/team-formation/form-flexible`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify(requestBody),
        },
      );

      if (!response) return; // Token expirado, ya fue manejado por authenticatedFetch

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(
          errorData.message ||
            errorData.error ||
            "Error al formar grupos con plugins",
        );
      }

      const result = await response.json();

      // Debug: Log raw API response
      console.log("Raw API response:", JSON.stringify(result, null, 2));
      console.log(
        "Teams from API:",
        JSON.stringify(result.data?.teams, null, 2),
      );

      // Check if this is a multi-objective result with a Pareto front
      const paretoFrontData = result.data?.metadata?.pareto_front;
      if (result.success && paretoFrontData && paretoFrontData.length > 0) {
        paretoFront.value = paretoFrontData;
        paretoResult.value = result;
        // Don't save teams yet — wait for user to select a solution
        return result;
      }

      // If the response has teams, we need to save them as groups
      if (result.success && result.data?.teams) {
        // Attach global metadata so saveTeamsAsGroups can pass them per-team
        const teams = result.data.teams;
        teams._globalAlgorithm = result.data.algorithm ?? null;
        teams._globalWeights = result.data.weights ?? null;
        teams._globalTargets = result.data.targets ?? null;
        if (result.data.global_score != null)
          globalScore.value = result.data.global_score;
        await saveTeamsAsGroups(teams, token);

        const oceanVectorMap = new Map();
        availableStudents.value.forEach(s => {
          if (s.ocean_vector) oceanVectorMap.set(s.id, s.ocean_vector);
        });

        await loadCompleteSectionGroupsData(currentSectionId.value, token);

        availableStudents.value.forEach(s => {
          if (oceanVectorMap.has(s.id)) s.ocean_vector = oceanVectorMap.get(s.id);
        });

        groups.value.forEach(group => {
          group.students.forEach(s => {
            if (oceanVectorMap.has(s.id)) s.ocean_vector = oceanVectorMap.get(s.id);
          });
        });

        // Refresh demographic cache with all students currently in memory.
        const demographicCache = {};
        const allKnown = [
          ...availableStudents.value,
          ...groups.value.flatMap((g) => g.students),
        ];
        const socialMap = buildSocialConnectionsMap(
          compatibilityResponses.value,
        );
        allKnown.forEach((s) => {
          demographicCache[String(s.id)] = buildStudentDemographicEntry(
            s,
            socialMap,
          );
        });
        lastDemographicData.value = demographicCache;
      }

      // Reload using the consolidated endpoint so scores, weights and targets
      // are immediately visible without requiring a page reload.
      await loadCompleteSectionGroupsData(currentSectionId.value, token);

      return result;
    } catch (error) {
      console.error("Error in formGroupsByPlugins:", error);
      throw error;
    }
  };

  /**
   * Select a solution from the Pareto front and save its teams as groups.
   * @param {Object} solution - The selected Pareto solution (with .teams array)
   * @param {string} token - Authentication token
   */
  const selectParetoSolution = async (solution, token) => {
    try {
      if (!solution?.teams || !paretoResult.value) return;

      const result = paretoResult.value;
      const teams = solution.teams;
      teams._globalAlgorithm = result.data.algorithm ?? null;
      teams._globalWeights = result.data.weights ?? null;
      teams._globalTargets = result.data.targets ?? null;
      if (solution.global_score != null)
        globalScore.value = solution.global_score;

      await saveTeamsAsGroups(teams, token);

      const oceanVectorMap = new Map();
      availableStudents.value.forEach(s => {
        if (s.ocean_vector) oceanVectorMap.set(s.id, s.ocean_vector);
      });

      await loadCompleteSectionGroupsData(currentSectionId.value, token);

      availableStudents.value.forEach(s => {
        if (oceanVectorMap.has(s.id)) s.ocean_vector = oceanVectorMap.get(s.id);
      });

      groups.value.forEach(group => {
        group.students.forEach(s => {
          if (oceanVectorMap.has(s.id)) s.ocean_vector = oceanVectorMap.get(s.id);
        });
      });

      const demographicCache = {};
      const allKnown = [
        ...availableStudents.value,
        ...groups.value.flatMap((g) => g.students),
      ];
      const socialMap = buildSocialConnectionsMap(
        compatibilityResponses.value,
      );
      allKnown.forEach((s) => {
        demographicCache[String(s.id)] = buildStudentDemographicEntry(
          s,
          socialMap,
        );
      });
      lastDemographicData.value = demographicCache;

      // Clear pareto state
      paretoFront.value = null;
      paretoResult.value = null;
    } catch (error) {
      console.error("Error selecting Pareto solution:", error);
      throw error;
    }
  };

  /**
   * Dismiss the Pareto front without selecting a solution.
   */
  const dismissParetoFront = () => {
    paretoFront.value = null;
    paretoResult.value = null;
  };

  /**
   * Save team formation results as groups in the database
   * @param {Array} teams - Array of team objects from team-formation endpoint
   * @param {string} token - Authentication token
   */
  const saveTeamsAsGroups = async (teams, token) => {
    try {
      console.log("Saving teams as groups:", JSON.stringify(teams, null, 2));

      // Build assignments list for bulk save
      const assignments = [];
      for (let i = 0; i < teams.length; i++) {
        const team = teams[i];
        const groupNumber = i + 1;
        const groupName = `Grupo ${groupNumber}`;
        const affinityMeta = {
          affinityScore: team.affinity_score ?? null,
          providerScores:
            team.metadata?.provider_scores ??
            team.metadata?.ProviderScores ??
            null,
          algorithm: teams._globalAlgorithm ?? null,
          weights: teams._globalWeights ?? null,
          targets: teams._globalTargets ?? null,
          criterionBreakdown:
            team.metadata?.criterion_breakdown ??
            team.metadata?.CriterionBreakdown ??
            null,
        };

        for (const member of team.members) {
          const studentId = Number.parseInt(member.id || member);
          if (!studentId) continue;
          assignments.push({
            group_number: groupNumber,
            group_name: groupName,
            student_id: studentId,
            dominant_type: member.dominant_type || "",
            affinity_score: affinityMeta.affinityScore,
            provider_scores: affinityMeta.providerScores,
            algorithm: affinityMeta.algorithm,
            weights: affinityMeta.weights,
            targets: affinityMeta.targets,
            criterion_breakdown: affinityMeta.criterionBreakdown,
          });
        }
      }

      // 1. Try single Bulk Save
      let bulkSucceeded = false;
      try {
        const bulkRes = await authenticatedFetch(
          `${API_BASE_URL}/core/groups/bulk`,
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
              section_id: Number.parseInt(currentSectionId.value),
              assignments,
            }),
          },
        );
        if (bulkRes && bulkRes.ok) {
          bulkSucceeded = true;
          console.log(`[Bulk Save] Successfully saved ${assignments.length} assignments in 1 request.`);
        }
      } catch (bulkErr) {
        console.warn("[Bulk Save] Failed, falling back to sequential:", bulkErr);
      }

      // 2. Fallback to sequential save if bulk failed
      if (!bulkSucceeded) {
        console.log("[Sequential Save] Saving students one by one...");
        for (let i = 0; i < teams.length; i++) {
          const team = teams[i];
          const groupNumber = i + 1;
          const groupName = `Grupo ${groupNumber}`;

          for (const member of team.members) {
            const studentId = Number.parseInt(member.id || member);
            await saveGroupAssignment(
              currentSectionId.value,
              groupNumber,
              groupName,
              studentId,
              member.dominant_type || "",
              token,
              {
                affinityScore: team.affinity_score ?? null,
                providerScores:
                  team.metadata?.provider_scores ??
                  team.metadata?.ProviderScores ??
                  null,
                algorithm: teams._globalAlgorithm ?? null,
                weights: teams._globalWeights ?? null,
                targets: teams._globalTargets ?? null,
                criterionBreakdown:
                  team.metadata?.criterion_breakdown ??
                  team.metadata?.CriterionBreakdown ??
                  null,
              },
            );
          }
        }
      }
    } catch (error) {
      console.error("Error saving teams as groups:", error);
      throw error;
    }
  };

  /**
   * Get student test status (which tests they have completed)
   * @param {string} sectionId - Section ID
   * @param {string} token - Authentication token
   * @returns {Promise<Object>} Map of student IDs to test completion status
   */
  const getStudentTestStatus = async (sectionId, token) => {
    try {
      // Get enneagram responses
      const enneagramResponse = await authenticatedFetch(
        `${API_BASE_URL}/core/enneagram/responses?section_id=${sectionId}`,
      );

      // Get sociometric responses
      const sociogramResponse = await authenticatedFetch(
        `${API_BASE_URL}/core/group-compatibility/responses/section/${sectionId}`,
      );

      const testStatus = {};

      // Process enneagram responses
      if (enneagramResponse?.ok) {
        const data = await enneagramResponse.json();
        const responses = data.data || data || [];
        responses.forEach((response) => {
          const studentId = response.student_id?.toString();
          if (studentId) {
            if (!testStatus[studentId]) {
              testStatus[studentId] = {
                hasEnneagram: false,
                hasMbti: false,
                hasSociogram: false,
              };
            }
            testStatus[studentId].hasEnneagram = true;
          }
        });
      }

      // Process sociometric responses
      if (sociogramResponse?.ok) {
        const data = await sociogramResponse.json();
        const responses = data.data || data || [];
        responses.forEach((response) => {
          const studentId = response.student_id?.toString();
          if (studentId) {
            if (!testStatus[studentId]) {
              testStatus[studentId] = {
                hasEnneagram: false,
                hasMbti: false,
                hasSociogram: false,
              };
            }
            testStatus[studentId].hasSociogram = true;
          }
        });
      }

      return testStatus;
    } catch (error) {
      console.error("Error getting student test status:", error);
      return {};
    }
  };

  /**
   * Export groups and students to Excel file
   * @param {string} sectionId - Section ID
   * @param {string} token - Authentication token
   */
  const exportGroupsToExcel = async (sectionId, token) => {
    try {
      const response = await authenticatedFetch(
        `${API_BASE_URL}/core/groups/section/${sectionId}/export`,
        {
          method: "GET",
        },
      );

      if (!response) return; // Token expirado

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || "Error al exportar grupos a Excel");
      }

      // Get the blob from response
      const blob = await response.blob();

      // Create a temporary URL for the blob
      const url = globalThis.URL.createObjectURL(blob);

      // Create a link and click it to download
      const link = document.createElement("a");
      link.href = url;
      link.download = `grupos-estudiantes-seccion-${sectionId}.xlsx`;
      document.body.appendChild(link);
      link.click();

      // Cleanup
      link.remove();
      globalThis.URL.revokeObjectURL(url);

      return {
        success: true,
        message: "Archivo Excel descargado exitosamente",
      };
    } catch (error) {
      console.error("Error exporting groups to Excel:", error);
      throw error;
    }
  };

  return {
    // State
    groups,
    availableStudents,
    draggedStudent,
    questionnaires,
    loadingQuestionnaires,
    showCreateQuestionnaireModal,
    newQuestionnaire,
    globalScore,
    lastEnforceEquity,
    paretoFront,
    paretoResult,

    // Methods
    initializeGroupsData,
    addNewGroup,
    deleteGroup,
    deleteAllGroups,
    updateGroupName,
    saveGroupName,
    loadGroups,
    saveGroupAssignment,
    removeStudentFromGroup,
    onDragStart,
    onDragOver,
    onDropToGroup,
    onDropToAvailable,
    loadQuestionnaires,
    loadSectionGroupData,
    loadSectionSurveyData,
    loadCompleteSectionGroupsData,
    loadSectionResponses,
    openCreateQuestionnaireModal,
    closeCreateQuestionnaireModal,
    createQuestionnaire,
    deleteQuestionnaire,
    updateQuestionnaire,
    loadMyGroupMates,
    loadSectionStudents,
    submitSociometricResponse,
    formGroupsByEnneagram,
    formRandomGroups,
    formGroupsByPlugins,
    selectParetoSolution,
    dismissParetoFront,
    getStudentTestStatus,
    exportGroupsToExcel,
    evaluateGroups,
    lastCriteriaParameters,
    lastEnforceEquity,
  };
};
