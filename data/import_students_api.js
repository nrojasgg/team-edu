const fs = require('fs');
const path = require('path');

const BASE_URL = 'http://127.0.0.1:3000/api/v1';
const TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzYyODEzMTEsImlhdCI6MTc3NjE5NDkxMSwiaWQiOjksInBlcm1pc3Npb25zIjpbIkFjY2VkZXIgYSBjdWVzdGlvbmFyaW8gdmlzdGFzIiwiQWNjZWRlciBhIGN1cnNvcyB2aXN0b3MiLCJBY2NlZGVyIGEgZW5jdWVzdGFzIHZpc3RhcyIsIkFncmVnYXIgZXN0dWRpYW50ZSBhIGxhIHNlY2Npw7NuIiwiQ3JlYXIgYWRtaW5pc3RyYWRvciIsIkNyZWFyIGPDs2RpZ28gZGUgaW52aXRhY2nDs24iLCJDcmVhciBjb29yZGluYWRvciIsIkNyZWFyIGN1ZXN0aW9uYXJpbyIsIkNyZWFyIGN1ZXN0aW9uYXJpbyBwYXJhIGN1cnNvIiwiQ3JlYXIgY3Vlc3Rpb25hcmlvIHBhcmEgZ3J1cG8iLCJDcmVhciBjdWVzdGlvbmFyaW8gcGFyYSBzZWNjacOzbiIsIkNyZWFyIGN1cnNvIiwiQ3JlYXIgZW5jdWVzdGEgcGFyYSBjdXJzbyIsIkNyZWFyIGVuY3Vlc3RhIHBhcmEgZ3J1cG8iLCJDcmVhciBlbmN1ZXN0YSBwYXJhIHNlY2Npw7NuIiwiQ3JlYXIgZXN0dWRpYW50ZSIsIkNyZWFyIGdydXBvcyBhbGVhdG9yaW9zIiwiQ3JlYXIgZ3J1cG9zIGNvbiBwbHVnaW5zIiwiQ3JlYXIgcHJvZmVzb3IiLCJDcmVhciBzZWNjacOzbiIsImRlbGV0ZV9jb3Vyc2VzIiwiZGVsZXRlX3VzZXJzIiwiRWRpdGFyIGFkbWluaXN0cmFkb3IiLCJFZGl0YXIgY29vcmRpbmFkb3IiLCJFZGl0YXIgY3Vlc3Rpb25hcmlvIiwiRWRpdGFyIGN1cnNvIiwiRWRpdGFyIGVuY3Vlc3RhIiwiRWRpdGFyIGVzdHVkaWFudGUiLCJFZGl0YXIgZ3J1cG9zIiwiRWRpdGFyIHByb2Zlc29yIiwiRWRpdGFyIHNlY2Npw7NuIiwiRWRpdGFyIHNlbWVzdHJlIGFjdHVhbCIsIkVsaW1pbmFyIHVzdWFyaW8iLCJyZWFkX2NvdXJzZXMiLCJyZWFkX3VzZXJzIiwiUmV2b2NhciBjw7NkaWdvIGRlIGludml0YWNpw7NuIiwiVXNhciBjw7NkaWdvIGRlIGludml0YWNpw7NuIiwiVmVyIGdydXBvcyBwcm9waW9zIiwiVmVyIHNlY2Npb25lcyBwcm9waWFzIiwiVmVyIHRvZGFzIGxhcyBjdWVzdGlvbmFyaW8iLCJWZXIgdG9kYXMgbGFzIGVuY3Vlc3RhcyIsIlZlciB0b2RhcyBsYXMgcmVzcHVlc3RhcyBlbiBjdWVzdGlvbmFyaW8iLCJWZXIgdG9kYXMgbGFzIHJlc3B1ZXN0YXMgZW4gZW5jdWVzdGEiLCJWZXIgdG9kYXMgbGFzIHNlY2Npb25lcyIsIlZlciB0b2RvcyBsb3MgY3Vyc29zIiwid3JpdGVfY291cnNlcyIsIndyaXRlX3VzZXJzIl0sInJvbGUiOiJhZG1pbiJ9.hCeRmJlAi6tW_xEHdmN8u6M-W2Bjq9IfvPAhNoeY4Ts';
const STUDENTS_FILE = path.join(__dirname, 'students_50.json');
const RESULTS_FILE = path.join(__dirname, 'created_users.json');

function generatePassword() {
  return Array.from({ length: 16 }, () => Math.floor(Math.random() * 16).toString(16)).join('');
}

async function apiRequest(endpoint, options = {}) {
  const url = `${BASE_URL}${endpoint}`;
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

async function checkUserExists(rut) {
  const result = await apiRequest(`/users/by-rut?rut=${encodeURIComponent(rut)}`);
  // Response structure: { success: true, data: { id: 10, rut: "...", ... } }
  if (result.ok && result.data && result.data.id) {
    return result.data.id;
  }
  return null;
}

async function checkUserExistsByEmail(email) {
  const result = await apiRequest(`/users/by-email?email=${encodeURIComponent(email)}`);
  if (result.ok && result.data && result.data.id) {
    return result.data.id;
  }
  return null;
}

async function createUser(student) {
  const password = generatePassword();
  const body = {
    nombres: student.nombres,
    paterno: student.paterno,
    materno: student.materno,
    rut: student.rut,
    email: student.email,
    password,
    role_id: 5,
    career_code: student.career_code
  };

  const result = await apiRequest('/users', {
    method: 'POST',
    body: JSON.stringify(body)
  });
  
  // Debug: log actual response if failed
  if (!result.ok) {
    console.log(`    DEBUG: status=${result.status}, data=${JSON.stringify(result.data)}`);
  }
  
  return result;
}

async function enrollStudent(studentId, sectionId = 3) {
  const body = {
    student_id: studentId,
    section_id: sectionId
  };

  return apiRequest('/student-sections', {
    method: 'POST',
    body: JSON.stringify(body)
  });
}

async function step1_CreateUsers(students, results) {
  console.log('\n=== STEP 1: Creating Users + Enrolling in Section 3 ===\n');
  let created = 0;
  let skipped = 0;

  for (let i = 0; i < students.length; i++) {
    const student = students[i];
    const idx = i + 1;
    process.stdout.write(`[${idx}/50] ${student.nombres} ${student.paterno} (${student.rut})... `);

    // Check if user already exists by RUT
    let userId = await checkUserExists(student.rut);

    // If not found by RUT, check by email (user might exist with different RUT)
    if (!userId) {
      userId = await checkUserExistsByEmail(student.email);
    }

    if (userId) {
      console.log(`EXISTS (ID: ${userId})`);
      skipped++;
    } else {
      // User doesn't exist, create new one
      const result = await createUser(student);

      if (result.ok) {
        userId = result.data?.id || result.data?.user?.id;
        console.log(`CREATED (ID: ${userId})`);
        created++;
      } else {
        console.log(`ERROR: ${result.data?.message || result.status}`);
        results[student.rut] = { rut: student.rut, internal_user_id: null, action: 'error', error: result.data?.message };
        continue;
      }
    }

    // Store user ID
    results[student.rut] = { rut: student.rut, internal_user_id: userId, action: userId && !skipped ? 'created' : 'exists' };

    // Now enroll in section 3
    if (userId) {
      process.stdout.write(`Enrolling in section 3... `);
      const enrollResult = await enrollStudent(userId, 3);
      if (enrollResult.ok) {
        results[student.rut].enrolled = true;
        console.log(`ENROLLED`);
      } else {
        results[student.rut].enrolled = false;
        console.log(`ENROLL ERROR: ${enrollResult.data?.message || enrollResult.status}`);
      }
    }
  }

  console.log(`\n--- Step 1 Complete: ${created} created, ${skipped} already existed ---`);
}

async function step2_EnrollStudents(students, results) {
  console.log('\n=== STEP 2: Enrolling Students in Section 3 ===\n');
  let enrolled = 0;
  let errors = 0;

  for (let i = 0; i < students.length; i++) {
    const student = students[i];
    const idx = i + 1;
    const resultEntry = results[student.rut];

    if (!resultEntry || !resultEntry.internal_user_id) {
      console.log(`[${idx}/50] ${student.rut}... SKIP (no user ID)`);
      errors++;
      continue;
    }

    process.stdout.write(`[${idx}/50] ${student.rut} (User ID: ${resultEntry.internal_user_id})... `);

    const enrollResult = await enrollStudent(resultEntry.internal_user_id, 3);

    if (enrollResult.ok) {
      console.log(`ENROLLED`);
      resultEntry.enrolled = true;
      enrolled++;
    } else {
      console.log(`ERROR: ${enrollResult.data?.message || enrollResult.status}`);
      resultEntry.enrolled = false;
      resultEntry.enroll_error = enrollResult.data?.message;
      errors++;
    }
  }

  console.log(`\n--- Step 2 Complete: ${enrolled} enrolled, ${errors} errors ---`);
}

async function main() {
  const args = process.argv.slice(2);
  let step = 'all';

  for (const arg of args) {
    if (arg.startsWith('--step=')) {
      step = arg.split('=')[1];
    }
  }

  if (!['all', 'users', 'enroll'].includes(step)) {
    console.error('Invalid --step value. Use: all, users, or enroll');
    process.exit(1);
  }

  console.log(`Starting import students script (step: ${step})`);

  if (!fs.existsSync(STUDENTS_FILE)) {
    console.error(`Students file not found: ${STUDENTS_FILE}`);
    process.exit(1);
  }

  const rawData = JSON.parse(fs.readFileSync(STUDENTS_FILE, 'utf8'));

  // Handle both nested {section_id, students} and direct array formats
  const students = Array.isArray(rawData) ? rawData : (rawData.students || []);

  if (!Array.isArray(students) || students.length === 0) {
    console.error('Invalid or empty students data');
    process.exit(1);
  }

  console.log(`Loaded ${students.length} students from ${STUDENTS_FILE}`);

  let results = {};

  if (fs.existsSync(RESULTS_FILE)) {
    try {
      results = JSON.parse(fs.readFileSync(RESULTS_FILE, 'utf8'));
      console.log(`Loaded existing results from ${RESULTS_FILE}`);
    } catch (e) {
      results = {};
    }
  }

  if (step === 'all' || step === 'users') {
    await step1_CreateUsers(students, results);
    fs.writeFileSync(RESULTS_FILE, JSON.stringify(results, null, 2));
    console.log(`Results saved to ${RESULTS_FILE}`);
  }

  // Only run step2 separately if explicitly requested (for re-running enrollment only)
  if (step === 'enroll') {
    await step2_EnrollStudents(students, results);
    fs.writeFileSync(RESULTS_FILE, JSON.stringify(results, null, 2));
    console.log(`Results saved to ${RESULTS_FILE}`);
  }

  console.log('\n=== Import Complete ===');
  console.log(`Results saved to: ${RESULTS_FILE}`);
}

main().catch(console.error);
