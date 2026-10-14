const fs = require('fs');
const path = require('path');

const BASE_URL = "http://127.0.0.1:3003/api/v1";
const TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzYyODI4MzAsImlhdCI6MTc3NjE5NjQzMCwiaWQiOjksInBlcm1pc3Npb25zIjpbIkFjY2VkZXIgYSBjdWVzdGlvbmFyaW8gdmlzdGFzIiwiQWNjZWRlciBhIGN1cnNvcyB2aXN0b3MiLCJBY2NlZGVyIGEgZW5jdWVzdGFzIHZpc3RhcyIsIkFncmVnYXIgZXN0dWRpYW50ZSBhIGxhIHNlY2Npw7NuIiwiQ3JlYXIgYWRtaW5pc3RyYWRvciIsIkNyZWFyIGPDs2RpZ28gZGUgaW52aXRhY2nDs24iLCJDcmVhciBjb29yZGluYWRvciIsIkNyZWFyIGN1ZXN0aW9uYXJpbyIsIkNyZWFyIGN1ZXN0aW9uYXJpbyBwYXJhIGN1cnNvIiwiQ3JlYXIgY3Vlc3Rpb25hcmlvIHBhcmEgZ3J1cG8iLCJDcmVhciBjdWVzdGlvbmFyaW8gcGFyYSBzZWNjacOzbiIsIkNyZWFyIGN1cnNvIiwiQ3JlYXIgZW5jdWVzdGEgcGFyYSBjdXJzbyIsIkNyZWFyIGVuY3Vlc3RhIHBhcmEgZ3J1cG8iLCJDcmVhciBlbmN1ZXN0YSBwYXJhIHNlY2Npw7NuIiwiQ3JlYXIgZXN0dWRpYW50ZSIsIkNyZWFyIGdydXBvcyBhbGVhdG9yaW9zIiwiQ3JlYXIgZ3J1cG9zIGNvbiBwbHVnaW5zIiwiQ3JlYXIgcHJvZmVzb3IiLCJDcmVhciBzZWNjacOzbiIsImRlbGV0ZV9jb3Vyc2VzIiwiZGVsZXRlX3VzZXJzIiwiRWRpdGFyIGFkbWluaXN0cmFkb3IiLCJFZGl0YXIgY29vcmRpbmFkb3IiLCJFZGl0YXIgY3Vlc3Rpb25hcmlvIiwiRWRpdGFyIGN1cnNvIiwiRWRpdGFyIGVuY3Vlc3RhIiwiRWRpdGFyIGVzdHVkaWFudGUiLCJFZGl0YXIgZ3J1cG9zIiwiRWRpdGFyIHByb2Zlc29yIiwiRWRpdGFyIHNlY2Npw7NuIiwiRWRpdGFyIHNlbWVzdHJlIGFjdHVhbCIsIkVsaW1pbmFyIHVzdWFyaW8iLCJyZWFkX2NvdXJzZXMiLCJyZWFkX3VzZXJzIiwiUmV2b2NhciBjw7NkaWdvIGRlIGludml0YWNpw7NuIiwiVXNhciBjw7NkaWdvIGRlIGludml0YWNpw7NuIiwiVmVyIGdydXBvcyBwcm9waW9zIiwiVmVyIHNlY2Npb25lcyBwcm9waWFzIiwiVmVyIHRvZGFzIGxhcyBjdWVzdGlvbmFyaW8iLCJWZXIgdG9kYXMgbGFzIGVuY3Vlc3RhcyIsIlZlciB0b2RhcyBsYXMgcmVzcHVlc3RhcyBlbiBjdWVzdGlvbmFyaW8iLCJWZXIgdG9kYXMgbGFzIHJlc3B1ZXN0YXMgZW4gZW5jdWVzdGEiLCJWZXIgdG9kYXMgbGFzIHNlY2Npb25lcyIsIlZlciB0b2RvcyBsb3MgY3Vyc29zIiwid3JpdGVfY291cnNlcyIsIndyaXRlX3VzZXJzIl0sInJvbGUiOiJhZG1pbiJ9.y-5g4nn7ajif_YI9VAk8A7WCeickluGoxZdm69eat5A";
const TEST_ID = "69de998f2a057afd8f524906";
const SECTION_ID = 3;

const ENNEAGRAM_TYPE_PATTERNS = {
  1: { name: "The Reformer", high: [5, 6, 7], low: [2, 3, 4] },
  2: { name: "The Helper", high: [1, 3, 6], low: [7, 8, 9] },
  3: { name: "The Achiever", high: [1, 2, 5], low: [4, 7, 9] },
  4: { name: "The Individualist", high: [3, 5, 9], low: [1, 6, 7] },
  5: { name: "The Investigator", high: [3, 6, 7], low: [1, 2, 4] },
  6: { name: "The Loyalist", high: [1, 4, 5], low: [3, 8, 9] },
  7: { name: "The Enthusiast", high: [2, 5, 8], low: [1, 3, 4] },
  8: { name: "The Challenger", high: [2, 5, 7], low: [3, 4, 6] },
  9: { name: "The Peacemaker", high: [3, 4, 6], low: [1, 2, 8] }
};

function generateAnswers(enneatype, totalQuestions = 45) {
  const pattern = ENNEAGRAM_TYPE_PATTERNS[enneatype] || ENNEAGRAM_TYPE_PATTERNS[1];
  const answers = [];

  for (let q = 1; q <= totalQuestions; q++) {
    const questionGroup = ((q - 1) % 9) + 1;
    let score;

    if (pattern.high.includes(questionGroup)) {
      score = Math.floor(Math.random() * 2) + 4;
    } else if (pattern.low.includes(questionGroup)) {
      score = Math.floor(Math.random() * 2) + 1;
    } else {
      score = Math.floor(Math.random() * 3) + 2;
    }

    answers.push({ question_id: q, score });
  }

  return answers;
}

async function apiRequest(endpoint, options = {}) {
  const url = `${BASE_URL}${endpoint}`;
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${TOKEN}`,
    ...options.headers
  };

  try {
    const response = await fetch(url, { ...options, headers });
    const data = await response.json().catch(() => null);
    return { ok: response.ok, status: response.status, data };
  } catch (error) {
    return { ok: false, status: 0, data: { message: error.message } };
  }
}

async function submitResponse(testId, studentId, answers) {
  const payload = {
    student_id: studentId,
    answers
  };

  return apiRequest(`/plugin/enneagram/tests/${testId}/responses`, {
    method: 'POST',
    body: JSON.stringify(payload)
  });
}

async function main() {
  if (TOKEN === "YOUR_TOKEN_HERE") {
    console.error("ERROR: Please set your API token in the TOKEN constant");
    process.exit(1);
  }

  const studentsPath = path.join(__dirname, 'students_50.json');
  const studentsData = JSON.parse(fs.readFileSync(studentsPath, 'utf8'));
  const students = Array.isArray(studentsData) ? studentsData : studentsData.students;

  console.log(`Loaded ${students.length} students from students_50.json`);
  console.log(`Using test ID: ${TEST_ID}\n`);

  const results = [];
  let successCount = 0;
  let failCount = 0;

  for (let i = 0; i < students.length; i++) {
    const student = students[i];
    const idx = i + 1;
    
    const enneatype = parseInt(student.eneatipo?.replace('e','')) || 
                      parseInt(student.enneagram_type?.replace('e','')) || 
                      parseInt(student.enneatipo) || 1;
    
    const answers = generateAnswers(enneatype, 45);

    process.stdout.write(`[${idx}/50] ${student.nombres} ${student.paterno} (type e${enneatype})... `);

    const result = await submitResponse(TEST_ID, student.rut, answers);

    if (result.ok) {
      console.log('SUCCESS');
      successCount++;
    } else {
      console.log(`FAILED: ${JSON.stringify(result.data)}`);
      failCount++;
    }

    results.push({
      student_id: student.id,
      rut: student.rut,
      enneatype: `e${enneatype}`,
      success: result.ok,
      error: result.ok ? null : result.data?.message
    });

    await new Promise(resolve => setTimeout(resolve, 100));
  }

  const outputPath = path.join(__dirname, 'enneagram_upload_results.json');
  fs.writeFileSync(outputPath, JSON.stringify({
    test_id: TEST_ID,
    timestamp: new Date().toISOString(),
    summary: { total: students.length, success: successCount, failed: failCount },
    results
  }, null, 2));

  console.log(`\n=== Completed: ${successCount} succeeded, ${failCount} failed ===`);
  console.log(`Results saved to: ${outputPath}`);
}

main().catch(error => {
  console.error('Fatal error:', error);
  process.exit(1);
});
