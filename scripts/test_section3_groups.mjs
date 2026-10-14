// Test script to diagnose http://127.0.0.1:2999/section/3/groups API performance and data payloads
const BASE_URL = 'http://127.0.0.1:3000'; // API Gateway

async function timedFetch(label, url, options = {}) {
  const start = performance.now();
  try {
    const res = await fetch(url, options);
    const duration = (performance.now() - start).toFixed(2);
    const text = await res.text();
    let data;
    try {
      data = JSON.parse(text);
    } catch {
      data = text;
    }
    console.log(`[${duration}ms] ${label} -> HTTP ${res.status}`);
    return { ok: res.ok, status: res.status, duration: Number(duration), data };
  } catch (err) {
    const duration = (performance.now() - start).toFixed(2);
    console.log(`[${duration}ms] ${label} -> ERROR: ${err.message}`);
    return { ok: false, error: err.message, duration: Number(duration) };
  }
}

async function run() {
  console.log('=== 1. AUTENTICACIÓN ===');
  let loginRes = await timedFetch('POST /api/v1/auth/login', `${BASE_URL}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: 'nicolas.rojas.g@usach.cl', password: 'Hola123.' })
  });

  const token = loginRes.data?.token || loginRes.data?.data?.token;
  if (!token) {
    console.error('No se pudo obtener el token. Respuesta:', loginRes.data);
    return;
  }
  console.log('Token obtenido exitosamente.\n');

  const authHeaders = {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  };

  console.log('=== 2. CARGA INICIAL DE LA PÁGINA /section/3/groups ===');
  
  // 2.1 Información de la sección
  const sectionRes = await timedFetch('GET /api/v1/sections/3', `${BASE_URL}/api/v1/sections/3`, { headers: authHeaders });

  // 2.2 Estudiantes completos (courses)
  const studentsRes = await timedFetch('GET /api/v1/sections/3/students-complete', `${BASE_URL}/api/v1/sections/3/students-complete`, { headers: authHeaders });
  const studentsCount = studentsRes.data?.students?.length ?? (Array.isArray(studentsRes.data) ? studentsRes.data.length : 0);
  console.log(`   -> Cantidad de estudiantes: ${studentsCount}`);

  // 2.3 Datos de grupos (ms-core)
  const groupDataRes = await timedFetch('GET /api/v1/sections/3/group-data', `${BASE_URL}/api/v1/sections/3/group-data`, { headers: authHeaders });
  console.log(`   -> Estudiantes devueltos por group-data: ${groupDataRes.data?.students?.length ?? 0}`);
  console.log(`   -> Grupos existentes: ${groupDataRes.data?.groups?.length ?? 0}`);

  // 2.4 Datos de encuestas / cuestionarios (ms-core)
  const surveyDataRes = await timedFetch('GET /api/v1/sections/3/survey-data', `${BASE_URL}/api/v1/sections/3/survey-data`, { headers: authHeaders });
  if (surveyDataRes.ok) {
    const surveys = surveyDataRes.data;
    console.log('   -> Cuestionarios detectados:');
    console.log('      - Eneagrama:', surveys.enneagram_questionnaires?.length ?? 0);
    console.log('      - BFI-44:', surveys.bfi44_questionnaires?.length ?? 0);
    console.log('      - MBTI:', surveys.mbti_questionnaires?.length ?? 0);
    console.log('      - Sociograma:', surveys.compatibility_surveys?.length ?? 0);
    console.log('      - Demográficos:', surveys.demographic_tests?.length ?? 0);
  }

  // 2.5 Plugins activos
  const pluginsRes = await timedFetch('GET /api/v1/core/plugins', `${BASE_URL}/api/v1/core/plugins`, { headers: authHeaders });
  const pluginsList = pluginsRes.data?.data || pluginsRes.data || [];
  console.log(`   -> Plugins activos: ${Array.isArray(pluginsList) ? pluginsList.map(p => p.nombre_interno || p.name).join(', ') : 'N/A'}`);

  console.log('\n=== 3. CARGA DE RESPUESTAS AL ABRIR EL MODAL DE GENERACIÓN ===');
  // loadSectionResponses usa usePluginApi con getEndpoint(pluginName, ...)
  const responseEndpoints = [
    { name: 'enneagram', url: `${BASE_URL}/api/v1/core/plugin/enneagram/responses?section_id=3` },
    { name: 'bfi44', url: `${BASE_URL}/api/v1/core/plugin/bfi44/responses?section_id=3` },
    { name: 'mbti', url: `${BASE_URL}/api/v1/core/plugin/mbti/responses?section_id=3` },
    { name: 'group_compatibility', url: `${BASE_URL}/api/v1/core/plugin/group_compatibility/responses/section/3` },
    { name: 'demographic', url: `${BASE_URL}/api/v1/core/plugin/demographic/responses?section_id=3` }
  ];

  const tStartResponses = performance.now();
  const responsesResults = await Promise.all(
    responseEndpoints.map(ep => timedFetch(`GET respuestas ${ep.name}`, ep.url, { headers: authHeaders }))
  );
  const totalResponsesTime = (performance.now() - tStartResponses).toFixed(2);
  console.log(`   -> Carga en paralelo de todas las respuestas: ${totalResponsesTime}ms`);
  responsesResults.forEach((r, idx) => {
    const ep = responseEndpoints[idx];
    const count = Array.isArray(r.data?.data) ? r.data.data.length : (Array.isArray(r.data) ? r.data.length : (r.data ? Object.keys(r.data).length : 0));
    console.log(`      - ${ep.name}: ${count} respuestas`);
  });

  console.log('\n=== 4. SIMULACIÓN DE CREACIÓN DE GRUPOS (ALGORITMO) ===');
  const candidates = (studentsRes.data?.students || []).map(s => String(s.id || s.user_id));
  console.log(`   -> Candidatos para la formación (${candidates.length}):`, candidates.slice(0, 5), '...');

  if (candidates.length === 0) {
    console.warn('   ! No hay candidatos en la sección 3 para ejecutar el algoritmo.');
    return;
  }

  // Construimos un payload idéntico a buildNewFormatRequestBody en useSectionGroups.js:
  const formationPayload = {
    section_id: "3",
    candidates: candidates,
    team_size: 4,
    algorithm: 'genetic',
    criteria_parameters: {
      openness: { target: 0.5, weight: 0.5 },
      social: { target: 0.5, weight: 0.5 }
    },
    demographic_data: {}
  };

  for (const cId of candidates) {
    formationPayload.demographic_data[cId] = {
      numeric_attrs: { openness: 0.7 },
      categorical_attrs: { sex: 'M' }
    };
  }

  const formationRes = await timedFetch(
    'POST /api/v1/core/team-formation/form-flexible',
    `${BASE_URL}/api/v1/core/team-formation/form-flexible`,
    {
      method: 'POST',
      headers: authHeaders,
      body: JSON.stringify(formationPayload)
    }
  );

  if (formationRes.ok) {
    const fData = formationRes.data?.data || formationRes.data;
    console.log('\n--- DATOS ENTREGADOS POR EL ALGORITMO ---');
    console.log('Algoritmo ejecutado:', fData?.algorithm);
    console.log('Global Score:', fData?.global_score);
    console.log('Equipos generados:', fData?.teams?.length);
    if (fData?.teams && fData.teams.length > 0) {
      console.log('Estructura del primer equipo:', JSON.stringify(fData.teams[0], null, 2).slice(0, 500) + '...');
    }
    if (fData?.metadata?.pareto_front) {
      console.log('Frente de Pareto generado! Soluciones:', fData.metadata.pareto_front.length);
      console.log('Knee Index:', fData.metadata.knee_index);
      console.log('Primera solución de Pareto:', {
        psy_score: fData.metadata.pareto_front[0].psy_score,
        soc_score: fData.metadata.pareto_front[0].soc_score,
        global_score: fData.metadata.pareto_front[0].global_score
      });
    }

    console.log('\n=== 5. SIMULACIÓN DE GUARDADO (saveTeamsAsGroups) ===');
    const teams = fData?.teams || [];
    const allAssignments = [];
    teams.forEach((t, idx) => {
      const gNum = idx + 1;
      (t.members || []).forEach(m => {
        allAssignments.push({
          group_number: gNum,
          group_name: `Grupo ${gNum}`,
          student_id: Number(m.id || m),
          dominant_type: m.dominant_type || "",
          affinity_score: t.affinity_score ?? null
        });
      });
    });
    console.log(`Total de asignaciones a guardar: ${allAssignments.length} estudiantes.`);

    // 5.1 Test Bulk Save
    const bulkSavePayload = {
      section_id: 3,
      assignments: allAssignments
    };
    const bulkSaveRes = await timedFetch('POST /api/v1/core/groups/bulk (TODOS los estudiantes a la vez)', `${BASE_URL}/api/v1/core/groups/bulk`, {
      method: 'POST',
      headers: authHeaders,
      body: JSON.stringify(bulkSavePayload)
    });

    if (bulkSaveRes.ok) {
      console.log(`   ⚡ [BULK SAVE EXITOSO] Guardadas ${allAssignments.length} asignaciones en solo ${bulkSaveRes.duration}ms!`);
      console.log('   Respuesta:', bulkSaveRes.data);
    } else {
      console.log('   Bulk endpoint response:', bulkSaveRes.status, bulkSaveRes.data);
    }

    // Limpiar eliminando los grupos de prueba
    await timedFetch('DELETE /api/v1/core/groups/section/3', `${BASE_URL}/api/v1/core/groups/section/3`, {
      method: 'DELETE',
      headers: authHeaders
    });
  } else {
    console.error('Error en formación de grupos:', formationRes.data);
  }
}

run().catch(console.error);
