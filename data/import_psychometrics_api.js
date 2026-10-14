const fs = require('fs');
const path = require('path');

const BASE_URL = 'http://127.0.0.1:3000/api/v1';
const TOKEN = 'YOUR_TOKEN_HERE';
const DATA_DIR = path.join(__dirname);

const STUDENTS_FILE = path.join(DATA_DIR, 'students_50.json');
const USERS_FILE = path.join(DATA_DIR, 'created_users.json');
const RESULTS_FILE = path.join(DATA_DIR, 'psychometrics_upload_results.json');

const PSYCHOMETRIC_ENDPOINTS = [
  '/core/student-psychometrics',
  '/core/student-psychometric',
  '/psychometrics',
  '/plugin/bfi44/tests/{test_id}/responses'
];

const DEMOGRAPHIC_ENDPOINTS = [
  '/core/student-demographics',
  '/core/student-demographic',
  '/demographic/tests'
];

function parseArgs() {
  const args = process.argv.slice(2);
  const params = { token: TOKEN };

  for (const arg of args) {
    if (arg.startsWith('--token=')) {
      params.token = arg.split('=')[1];
    }
  }

  return params;
}

async function apiRequest(url, options = {}) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${TOKEN}`,
    ...options.headers
  };

  try {
    const response = await fetch(url, {
      ...options,
      headers
    });

    const data = await response.json().catch(() => null);
    return { ok: response.ok, status: response.status, data };
  } catch (error) {
    return { ok: false, status: 0, data: { message: error.message } };
  }
}

async function discoverEndpoint(baseUrl, endpoints, testPayload) {
  for (const endpoint of endpoints) {
    const url = `${baseUrl}${endpoint.replace('{test_id}', 'test-123')}`;
    console.log(`  Trying: ${endpoint}`);

    const result = await apiRequest(url, {
      method: 'POST',
      body: JSON.stringify(testPayload)
    });

    if (result.ok || (result.status >= 200 && result.status < 300)) {
      console.log(`  ✓ Endpoint responded: ${endpoint}`);
      return endpoint;
    }
  }
  return null;
}

async function discoverDemographicEndpoint(baseUrl, endpoints, testPayload) {
  for (const endpoint of endpoints) {
    const url = `${baseUrl}${endpoint}`;
    console.log(`  Trying: ${endpoint}`);

    const result = await apiRequest(url, {
      method: 'POST',
      body: JSON.stringify(testPayload)
    });

    if (result.ok || (result.status >= 200 && result.status < 300)) {
      console.log(`  ✓ Endpoint responded: ${endpoint}`);
      return endpoint;
    }
  }
  return null;
}

function generateOCEAN(enneatype) {
  const base = {
    o: 0.5 + (parseInt(enneatype) || 1) * 0.05,
    c: 0.5 + Math.random() * 0.3,
    e: 0.5 + Math.random() * 0.3,
    a: 0.5 + Math.random() * 0.3,
    n: 0.3 + Math.random() * 0.4
  };

  base.o = Math.min(1, Math.max(0, base.o));
  base.c = Math.min(1, Math.max(0, base.c));
  base.e = Math.min(1, Math.max(0, base.e));
  base.a = Math.min(1, Math.max(0, base.a));
  base.n = Math.min(1, Math.max(0, base.n));

  return base;
}

function generateSociometry(studentId, allStudentIds) {
  const sociometry = {};

  for (const otherId of allStudentIds) {
    if (otherId === studentId) continue;

    const random = Math.random();
    if (random > 0.6) {
      sociometry[otherId] = random > 0.8 ? 1.0 : 0.5;
    } else if (random < 0.2) {
      sociometry[otherId] = -0.5;
    }
  }

  return sociometry;
}

async function loadData() {
  if (!fs.existsSync(STUDENTS_FILE)) {
    throw new Error(`Students file not found: ${STUDENTS_FILE}`);
  }

  const students = JSON.parse(fs.readFileSync(STUDENTS_FILE, 'utf8'));

  if (!fs.existsSync(USERS_FILE)) {
    throw new Error(`Users file not found: ${USERS_FILE}`);
  }

  const userMapping = JSON.parse(fs.readFileSync(USERS_FILE, 'utf8'));

  let userMap = {};
  if (Array.isArray(userMapping)) {
    for (const user of userMapping) {
      if (user.student_id && user.user_id) {
        userMap[user.student_id] = user.user_id;
      } else if (user.student_id && user.internal_user_id) {
        userMap[user.student_id] = user.internal_user_id;
      }
    }
  } else {
    for (const [rut, entry] of Object.entries(userMapping)) {
      if (entry.student_id && entry.internal_user_id) {
        userMap[entry.student_id] = entry.internal_user_id;
      }
    }
  }

  return { students, userMap };
}

async function uploadPsychometrics(student, internalUserId, sectionId, allStudentIds, endpoint) {
  const ocean = generateOCEAN(student.enneatipo || student.enneagram_type || 1);
  const sociometry = generateSociometry(student.student_id, allStudentIds);

  const payload = {
    student_id: internalUserId,
    section_id: sectionId,
    ocean: {
      o: ocean.o,
      c: ocean.c,
      e: ocean.e,
      a: ocean.a,
      n: ocean.n
    },
    sociometry
  };

  const url = `${BASE_URL}${endpoint}`;
  return apiRequest(url, {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

async function uploadDemographics(student, internalUserId, sectionId, endpoint) {
  const gender = student.gender || student.sexo || null;
  const eneatipo = student.enneatipo || student.enneagram_type || null;
  const careerCode = student.career_code || null;

  if (!gender && !eneatipo && !careerCode) {
    return { ok: true, status: 204, data: { message: 'No demographic data to upload' }, skipped: true };
  }

  const payload = {
    student_id: internalUserId,
    section_id: sectionId,
    gender,
    eneatipo,
    career_code: careerCode
  };

  const url = `${BASE_URL}${endpoint}`;
  return apiRequest(url, {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

async function main() {
  const params = parseArgs();

  if (params.token !== TOKEN) {
    console.log(`Using token from --token argument`);
  }

  console.log('=== Psychometrics Import Script ===\n');
  console.log('Loading data files...');

  let students, userMap;
  try {
    ({ students, userMap } = await loadData());
  } catch (error) {
    console.error('Failed to load data:', error.message);
    process.exit(1);
  }

  console.log(`Loaded ${students.length} students`);
  console.log(`Loaded ${Object.keys(userMap).length} user mappings\n`);

  const allStudentIds = students.map(s => s.student_id);

  console.log('=== Discovering Psychometric Endpoint ===');
  const testPayload = {
    student_id: 'test-student',
    section_id: 3,
    ocean: { o: 0.5, c: 0.5, e: 0.5, a: 0.5, n: 0.5 },
    sociometry: {}
  };

  let psychometricEndpoint = await discoverEndpoint(BASE_URL, PSYCHOMETRIC_ENDPOINTS, testPayload);

  if (!psychometricEndpoint) {
    console.log('\n⚠ Could not discover working psychometric endpoint');
    console.log('  Will attempt direct plugin endpoints...');

    psychometricEndpoint = '/plugin/bfi44/tests/test-123/responses';
  }

  console.log(`\n=== Discovering Demographic Endpoint ===`);
  const demoPayload = {
    student_id: 'test-student',
    section_id: 3,
    gender: 'M',
    eneatipo: 3,
    career_code: '21.01'
  };

  let demographicEndpoint = await discoverDemographicEndpoint(BASE_URL, DEMOGRAPHIC_ENDPOINTS, demoPayload);

  if (!demographicEndpoint) {
    console.log('\n⚠ Could not discover working demographic endpoint');
  }

  const results = {
    timestamp: new Date().toISOString(),
    psychometric_endpoint: psychometricEndpoint,
    demographic_endpoint: demographicEndpoint || null,
    students: []
  };

  let psychometricSuccess = 0;
  let psychometricFailed = 0;
  let demographicSuccess = 0;
  let demographicSkipped = 0;
  let demographicFailed = 0;

  console.log('\n=== Uploading Psychometrics ===\n');

  for (let i = 0; i < students.length; i++) {
    const student = students[i];
    const idx = i + 1;

    const internalUserId = userMap[student.student_id];

    if (!internalUserId) {
      console.log(`[${idx}/${students.length}] ${student.student_id}... SKIP (no user mapping)`);
      results.students.push({
        student_id: student.student_id,
        psychometric_success: false,
        psychometric_error: 'No user mapping',
        demographic_success: false,
        demographic_error: 'No user mapping'
      });
      continue;
    }

    process.stdout.write(`[${idx}/${students.length}] ${student.student_id} (User ID: ${internalUserId})... `);

    const sectionId = student.section_id || 3;

    const psychResult = await uploadPsychometrics(student, internalUserId, sectionId, allStudentIds, psychometricEndpoint);

    if (psychResult.ok) {
      console.log('PSYCHOMETRIC OK');
      psychometricSuccess++;
    } else {
      console.log(`PSYCHOMETRIC FAIL: ${psychResult.data?.message || psychResult.status}`);
      psychometricFailed++;
    }

    if (demographicEndpoint) {
      const demoResult = await uploadDemographics(student, internalUserId, sectionId, demographicEndpoint);

      if (demoResult.skipped) {
        demographicSkipped++;
      } else if (demoResult.ok) {
        process.stdout.write(' DEMOGRAPHIC OK');
        demographicSuccess++;
      } else {
        process.stdout.write(` DEMOGRAPHIC FAIL: ${demoResult.data?.message || demoResult.status}`);
        demographicFailed++;
      }
    }

    console.log('');

    results.students.push({
      student_id: student.student_id,
      internal_user_id: internalUserId,
      section_id: sectionId,
      psychometric_success: psychResult.ok,
      psychometric_error: psychResult.ok ? null : (psychResult.data?.message || `HTTP ${psychResult.status}`),
      demographic_success: demographicEndpoint ? demoResult.ok || demoResult.skipped : null,
      demographic_error: null
    });

    await new Promise(resolve => setTimeout(resolve, 50));
  }

  console.log('\n=== Upload Summary ===');
  console.log(`Psychometrics: ${psychometricSuccess} succeeded, ${psychometricFailed} failed`);
  if (demographicEndpoint) {
    console.log(`Demographics: ${demographicSuccess} succeeded, ${demographicSkipped} skipped, ${demographicFailed} failed`);
  }

  fs.writeFileSync(RESULTS_FILE, JSON.stringify(results, null, 2));
  console.log(`\nResults saved to: ${RESULTS_FILE}`);
}

if (require.main === module) {
  main().catch(error => {
    console.error('Fatal error:', error);
    process.exit(1);
  });
}

module.exports = { main };