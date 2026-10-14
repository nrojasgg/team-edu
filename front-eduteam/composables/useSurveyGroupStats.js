import { ref } from "vue";

const _unref = (val) =>
  val && typeof val === "object" && "value" in val ? val.value : val;

export const useSurveyGroupStats = (allStudents = [], allResponses = []) => {
  const _students = () => _unref(allStudents) || [];
  const _responses = () => _unref(allResponses) || [];
  const expandedGroup = ref(null);

  const toggleGroup = (groupId) => {
    if (expandedGroup.value === groupId) {
      expandedGroup.value = null;
    } else {
      expandedGroup.value = groupId;
    }
  };

  const getStudentName = (studentId) => {
    const student = _students().find((s) => String(s.id) === String(studentId));
    if (student) {
      return `${student.nombres || ""} ${student.paterno || ""} ${student.materno || ""}`.trim();
    }
    return `ID: ${studentId}`;
  };

  const _computeCategoryScores = (evaluationsReceived) => {
    const categoryScores = { 1: [], 2: [], 3: [], 4: [], 5: [] };
    evaluationsReceived.forEach((evaluation) => {
      if (evaluation.responses) {
        evaluation.responses.forEach((response) => {
          if (categoryScores[response.question_id]) {
            categoryScores[response.question_id].push(response.score);
          }
        });
      }
    });
    return categoryScores;
  };

  const _avg = (arr) =>
    arr.length > 0 ? arr.reduce((a, b) => a + b, 0) / arr.length : 0;

  const getGroupStudentAveragesRaw = (group) => {
    const groupStudentIds = new Set(group.students.map((s) => String(s.id)));
    const groupResponses = _responses().filter((r) =>
      groupStudentIds.has(String(r.evaluated_id)),
    );

    const studentAverages = [];

    group.students.forEach((student) => {
      const studentId = String(student.id);
      const evaluationsReceived = groupResponses.filter(
        (r) => String(r.evaluated_id) === studentId,
      );

      if (evaluationsReceived.length === 0) {
        studentAverages.push({ promedio: 0 });
        return;
      }

      const categoryScores = _computeCategoryScores(evaluationsReceived);

      const tareas = _avg(categoryScores[1]);
      const participacion = _avg(categoryScores[2]);
      const organizacion = _avg(categoryScores[3]);
      const compromiso = _avg(categoryScores[4]);
      const prosocialidad = _avg(categoryScores[5]);
      const promedio =
        (tareas + participacion + organizacion + compromiso + prosocialidad) /
        5;

      studentAverages.push({ promedio });
    });

    return studentAverages;
  };

  const getGroupStudentAverages = (group) => {
    const groupStudentIds = new Set(group.students.map((s) => String(s.id)));
    const groupResponses = _responses().filter((r) =>
      groupStudentIds.has(String(r.evaluated_id)),
    );

    const studentAverages = [];

    group.students.forEach((student) => {
      const studentId = String(student.id);
      const evaluationsReceived = groupResponses.filter(
        (r) => String(r.evaluated_id) === studentId,
      );

      if (evaluationsReceived.length === 0) {
        studentAverages.push({
          id: student.id,
          name: getStudentName(student.id) || `ID: ${student.id}`,
          tareas: 0,
          participacion: 0,
          organizacion: 0,
          compromiso: 0,
          prosocialidad: 0,
          promedio: 0,
          modulo: 0,
        });
        return;
      }

      const categoryScores = _computeCategoryScores(evaluationsReceived);

      const tareas = _avg(categoryScores[1]);
      const participacion = _avg(categoryScores[2]);
      const organizacion = _avg(categoryScores[3]);
      const compromiso = _avg(categoryScores[4]);
      const prosocialidad = _avg(categoryScores[5]);
      const promedio =
        (tareas + participacion + organizacion + compromiso + prosocialidad) /
        5;

      studentAverages.push({
        id: student.id,
        name: getStudentName(student.id),
        tareas,
        participacion,
        organizacion,
        compromiso,
        prosocialidad,
        promedio,
        modulo: 0,
      });
    });

    const baseline = getGroupBaseline(group);
    studentAverages.forEach((s) => {
      s.modulo = s.promedio - baseline;
    });

    return studentAverages;
  };

  const getGroupBaseline = (group) => {
    const studentAverages = getGroupStudentAveragesRaw(group);
    if (studentAverages.length === 0) return 0;

    const validAverages = studentAverages.filter((s) => s.promedio > 0);
    if (validAverages.length === 0) return 0;

    return (
      validAverages.reduce((sum, s) => sum + s.promedio, 0) /
      validAverages.length
    );
  };

  const getGroupCategoryAverages = (group) => {
    const studentAverages = getGroupStudentAverages(group);
    const validStudents = studentAverages.filter((s) => s.promedio > 0);

    if (validStudents.length === 0) {
      return {
        tareas: 0,
        participacion: 0,
        organizacion: 0,
        compromiso: 0,
        prosocialidad: 0,
      };
    }

    const avgKey = (key) =>
      validStudents.reduce((sum, s) => sum + s[key], 0) / validStudents.length;

    return {
      tareas: avgKey("tareas"),
      participacion: avgKey("participacion"),
      organizacion: avgKey("organizacion"),
      compromiso: avgKey("compromiso"),
      prosocialidad: avgKey("prosocialidad"),
    };
  };

  const getGroupCategoryStdDev = (group) => {
    const studentAverages = getGroupStudentAverages(group);
    const validStudents = studentAverages.filter((s) => s.promedio > 0);

    if (validStudents.length === 0) {
      return {
        tareas: 0,
        participacion: 0,
        organizacion: 0,
        compromiso: 0,
        prosocialidad: 0,
      };
    }

    const categoryAverages = getGroupCategoryAverages(group);

    const stdDev = (key) => {
      const mean = categoryAverages[key];
      const squaredDiffs = validStudents.map((s) => Math.pow(s[key] - mean, 2));
      const avgSquaredDiff =
        squaredDiffs.reduce((a, b) => a + b, 0) / validStudents.length;
      return Math.sqrt(avgSquaredDiff);
    };

    return {
      tareas: stdDev("tareas"),
      participacion: stdDev("participacion"),
      organizacion: stdDev("organizacion"),
      compromiso: stdDev("compromiso"),
      prosocialidad: stdDev("prosocialidad"),
    };
  };

  return {
    expandedGroup,
    toggleGroup,
    getStudentName,
    getGroupStudentAverages,
    getGroupStudentAveragesRaw,
    getGroupBaseline,
    getGroupCategoryAverages,
    getGroupCategoryStdDev,
  };
};
