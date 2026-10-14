const fs = require('fs');
const path = require('path');

const BASE_URL = "http://127.0.0.1:3003/api/v1";
const TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzYyODUwMzMsImlhdCI6MTc3NjE5ODYzMywiaWQiOjksInBlcm1pc3Npb25zIjpbIkFjY2VkZXIgYSBjdWVzdGlvbmFyaW8gdmlzdGFzIiwiQWNjZWRlciBhIGN1cnNvcyB2aXN0b3MiLCJBY2NlZGVyIGEgZW5jdWVzdGFzIHZpc3RhcyIsIkFncmVnYXIgZXN0dWRpYW50ZSBhIGxhIHNlY2Npw7NuIiwiQ3JlYXIgYWRtaW5pc3RyYWRvciIsIkNyZWFyIGPDs2RpZ28gZGUgaW52aXRhY2nDs24iLCJDcmVhciBjb29yZGluYWRvciIsIkNyZWFyIGN1ZXN0aW9uYXJpbyIsIkNyZWFyIGN1ZXN0aW9uYXJpbyBwYXJhIGN1cnNvIiwiQ3JlYXIgY3Vlc3Rpb25hcmlvIHBhcmEgZ3J1cG8iLCJDcmVhciBjdWVzdGlvbmFyaW8gcGFyYSBzZWNjacOzbiIsIkNyZWFyIGN1cnNvIiwiQ3JlYXIgZW5jdWVzdGEgcGFyYSBjdXJzbyIsIkNyZWFyIGVuY3Vlc3RhIHBhcmEgZ3J1cG8iLCJDcmVhciBlbmN1ZXN0YSBwYXJhIHNlY2Npw7NuIiwiQ3JlYXIgZXN0dWRpYW50ZSIsIkNyZWFyIGdydXBvcyBhbGVhdG9yaW9zIiwiQ3JlYXIgZ3J1cG9zIGNvbiBwbHVnaW5zIiwiQ3JlYXIgcHJvZmVzb3IiLCJDcmVhciBzZWNjacOzbiIsImRlbGV0ZV9jb3Vyc2VzIiwiZGVsZXRlX3VzZXJzIiwiRWRpdGFyIGFkbWluaXN0cmFkb3IiLCJFZGl0YXIgY29vcmRpbmFkb3IiLCJFZGl0YXIgY3Vlc3Rpb25hcmlvIiwiRWRpdGFyIGN1cnNvIiwiRWRpdGFyIGVuY3Vlc3RhIiwiRWRpdGFyIGVzdHVkaWFudGUiLCJFZGl0YXIgZ3J1cG9zIiwiRWRpdGFyIHByb2Zlc29yIiwiRWRpdGFyIHNlY2Npw7NuIiwiRWRpdGFyIHNlbWVzdHJlIGFjdHVhbCIsIkVsaW1pbmFyIHVzdWFyaW8iLCJyZWFkX2NvdXJzZXMiLCJyZWFkX3VzZXJzIiwiUmV2b2NhciBjw7NkaWdvIGRlIGludml0YWNpw7NuIiwiVXNhciBjw7NkaWdvIGRlIGludml0YWNpw7NuIiwiVmVyIGdydXBvcyBwcm9waW9zIiwiVmVyIHNlY2Npb25lcyBwcm9waWFzIiwiVmVyIHRvZGFzIGxhcyBjdWVzdGlvbmFyaW8iLCJWZXIgdG9kYXMgbGFzIGVuY3Vlc3RhcyIsIlZlciB0b2RhcyBsYXMgcmVzcHVlc3RhcyBlbiBjdWVzdGlvbmFyaW8iLCJWZXIgdG9kYXMgbGFzIHJlc3B1ZXN0YXMgZW4gZW5jdWVzdGEiLCJWZXIgdG9kYXMgbGFzIHNlY2Npb25lcyIsIlZlciB0b2RvcyBsb3MgY3Vyc29zIiwid3JpdGVfY291cnNlcyIsIndyaXRlX3VzZXJzIl0sInJvbGUiOiJhZG1pbiJ9.8qb0vo1oN1YWCgB3CmmvGCujjK2CuDugGfOfVTMsYSA";
const TEST_ID = "69dea1742a057afd8f524908";
const SECTION_ID = 3;

const BFI44_QUESTIONS = [
  { id: 1, dimension: "E", is_reverse: false },
  { id: 2, dimension: "E", is_reverse: true },
  { id: 3, dimension: "E", is_reverse: false },
  { id: 4, dimension: "E", is_reverse: false },
  { id: 5, dimension: "E", is_reverse: false },
  { id: 6, dimension: "E", is_reverse: true },
  { id: 7, dimension: "E", is_reverse: false },
  { id: 8, dimension: "E", is_reverse: true },
  { id: 9, dimension: "A", is_reverse: false },
  { id: 10, dimension: "A", is_reverse: false },
  { id: 11, dimension: "A", is_reverse: true },
  { id: 12, dimension: "A", is_reverse: false },
  { id: 13, dimension: "A", is_reverse: false },
  { id: 14, dimension: "A", is_reverse: true },
  { id: 15, dimension: "A", is_reverse: false },
  { id: 16, dimension: "A", is_reverse: true },
  { id: 17, dimension: "C", is_reverse: false },
  { id: 18, dimension: "C", is_reverse: true },
  { id: 19, dimension: "C", is_reverse: true },
  { id: 20, dimension: "C", is_reverse: false },
  { id: 21, dimension: "C", is_reverse: false },
  { id: 22, dimension: "C", is_reverse: true },
  { id: 23, dimension: "C", is_reverse: false },
  { id: 24, dimension: "C", is_reverse: true },
  { id: 25, dimension: "C", is_reverse: false },
  { id: 26, dimension: "N", is_reverse: false },
  { id: 27, dimension: "N", is_reverse: true },
  { id: 28, dimension: "N", is_reverse: false },
  { id: 29, dimension: "N", is_reverse: false },
  { id: 30, dimension: "N", is_reverse: true },
  { id: 31, dimension: "N", is_reverse: false },
  { id: 32, dimension: "N", is_reverse: true },
  { id: 33, dimension: "N", is_reverse: false },
  { id: 34, dimension: "O", is_reverse: false },
  { id: 35, dimension: "O", is_reverse: false },
  { id: 36, dimension: "O", is_reverse: false },
  { id: 37, dimension: "O", is_reverse: false },
  { id: 38, dimension: "O", is_reverse: false },
  { id: 39, dimension: "O", is_reverse: true },
  { id: 40, dimension: "O", is_reverse: false },
  { id: 41, dimension: "O", is_reverse: true },
  { id: 42, dimension: "A", is_reverse: false },
  { id: 43, dimension: "A", is_reverse: true },
  { id: 44, dimension: "O", is_reverse: false },
];

function generateOceanScores() {
  return {
    O: Math.random() * 0.5 + 0.3,
    C: Math.random() * 0.5 + 0.3,
    E: Math.random() * 0.5 + 0.3,
    A: Math.random() * 0.5 + 0.3,
    N: Math.random() * 0.5 + 0.3
  };
}

function generateAnswers(ocean) {
  const answers = [];
  
  for (const q of BFI44_QUESTIONS) {
    let score;
    const dimScore = ocean[q.dimension];
    
    if (q.is_reverse) {
      score = Math.max(1, Math.min(5, Math.round(6 - dimScore * 5 + (Math.random() - 0.5))));
    } else {
      score = Math.max(1, Math.min(5, Math.round(dimScore * 5 + (Math.random() - 0.5))));
    }
    
    answers.push({ question_id: q.id, score });
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

  return apiRequest(`/plugin/bfi44/tests/${testId}/responses`, {
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
    
    const ocean = generateOceanScores();
    const answers = generateAnswers(ocean);

    process.stdout.write(`[${idx}/50] ${student.nombres} ${student.paterno}... `);

    const result = await submitResponse(TEST_ID, student.id.toString(), answers);

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
      ocean,
      success: result.ok,
      error: result.ok ? null : result.data?.message
    });

    await new Promise(resolve => setTimeout(resolve, 100));
  }

  const outputPath = path.join(__dirname, 'bfi44_upload_results.json');
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
