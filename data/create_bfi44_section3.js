const fs = require('fs');
const path = require('path');
const http = require('http');

const TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzYyODUwMzMsImlhdCI6MTc3NjE5ODYzMywiaWQiOjksInBlcm1pc3Npb25zIjpbIkFjY2VkZXIgYSBjdWVzdGlvbmFyaW8gdmlzdGFzIiwiQWNjZWRlciBhIGN1cnNvcyB2aXN0b3MiLCJBY2NlZGVyIGEgZW5jdWVzdGFzIHZpc3RhcyIsIkFncmVnYXIgZXN0dWRpYW50ZSBhIGxhIHNlY2Npw7NuIiwiQ3JlYXIgYWRtaW5pc3RyYWRvciIsIkNyZWFyIGPDs2RpZ28gZGUgaW52aXRhY2nDs24iLCJDcmVhciBjb29yZGluYWRvciIsIkNyZWFyIGN1ZXN0aW9uYXJpbyIsIkNyZWFyIGN1ZXN0aW9uYXJpbyBwYXJhIGN1cnNvIiwiQ3JlYXIgY3Vlc3Rpb25hcmlvIHBhcmEgZ3J1cG8iLCJDcmVhciBjdWVzdGlvbmFyaW8gcGFyYSBzZWNjacOzbiIsIkNyZWFyIGN1cnNvIiwiQ3JlYXIgZW5jdWVzdGEgcGFyYSBjdXJzbyIsIkNyZWFyIGVuY3Vlc3RhIHBhcmEgZ3J1cG8iLCJDcmVhciBlbmN1ZXN0YSBwYXJhIHNlY2Npw7NuIiwiQ3JlYXIgZXN0dWRpYW50ZSIsIkNyZWFyIGdydXBvcyBhbGVhdG9yaW9zIiwiQ3JlYXIgZ3J1cG9zIGNvbiBwbHVnaW5zIiwiQ3JlYXIgcHJvZmVzb3IiLCJDcmVhciBzZWNjacOzbiIsImRlbGV0ZV9jb3Vyc2VzIiwiZGVsZXRlX3VzZXJzIiwiRWRpdGFyIGFkbWluaXN0cmFkb3IiLCJFZGl0YXIgY29vcmRpbmFkb3IiLCJFZGl0YXIgY3Vlc3Rpb25hcmlvIiwiRWRpdGFyIGN1cnNvIiwiRWRpdGFyIGVuY3Vlc3RhIiwiRWRpdGFyIGVzdHVkaWFudGUiLCJFZGl0YXIgZ3J1cG9zIiwiRWRpdGFyIHByb2Zlc29yIiwiRWRpdGFyIHNlY2Npw7NuIiwiRWRpdGFyIHNlbWVzdHJlIGFjdHVhbCIsIkVsaW1pbmFyIHVzdWFyaW8iLCJyZWFkX2NvdXJzZXMiLCJyZWFkX3VzZXJzIiwiUmV2b2NhciBjw7NkaWdvIGRlIGludml0YWNpw7NuIiwiVXNhciBjw7NkaWdvIGRlIGludml0YWNpw7NuIiwiVmVyIGdydXBvcyBwcm9waW9zIiwiVmVyIHNlY2Npb25lcyBwcm9waWFzIiwiVmVyIHRvZGFzIGxhcyBjdWVzdGlvbmFyaW8iLCJWZXIgdG9kYXMgbGFzIGVuY3Vlc3RhcyIsIlZlciB0b2RhcyBsYXMgcmVzcHVlc3RhcyBlbiBjdWVzdGlvbmFyaW8iLCJWZXIgdG9kYXMgbGFzIHJlc3B1ZXN0YXMgZW4gZW5jdWVzdGEiLCJWZXIgdG9kYXMgbGFzIHNlY2Npb25lcyIsIlZlciB0b2RvcyBsb3MgY3Vyc29zIiwid3JpdGVfY291cnNlcyIsIndyaXRlX3VzZXJzIl0sInJvbGUiOiJhZG1pbiJ9.8qb0vo1oN1YWCgB3CmmvGCujjK2CuDugGfOfVTMsYSA";
const SECTION_ID = "3";

function makeRequest(path, method, body = null) {
  return new Promise((resolve) => {
    const options = {
      hostname: 'localhost',
      port: 3003,
      path: `/api/v1${path}`,
      method,
      headers: { 'Authorization': `Bearer ${TOKEN}`, 'Content-Type': 'application/json' }
    };
    const req = http.request(options, (res) => {
      let data = '';
      res.on('data', c => data += c);
      res.on('end', () => { try { resolve(JSON.parse(data)); } catch(e) { resolve({error:true,raw:data}); } });
    });
    req.on('error', resolve);
    if (body) req.write(JSON.stringify(body));
    req.end();
  });
}

async function main() {
  console.log("Creating BFI-44 test for section 3...");
  const testResult = await makeRequest('/plugin/bfi44/tests', 'POST', {
    title: "BFI-44 - Sección 3",
    description: "BFI-44 test para sección 3",
    section_id: SECTION_ID,
    created_by: "admin"
  });

  if (!testResult.success) {
    console.log("Failed:", testResult);
    return;
  }
  const testId = testResult.data.id;
  console.log("Test created:", testId);

  console.log("Importing responses...");
  const data = JSON.parse(fs.readFileSync(path.join(__dirname, "students_50.json"), "utf8"));
  const students = data.students || [];
  let ok = 0, fail = 0;
  
  for (let i = 0; i < students.length; i++) {
    const s = students[i];
    if (!s.ocean) { fail++; continue; }
    const answers = [];
    for (let j = 1; j <= 44; j++) answers.push({ question_id: j, score: s.ocean.questions?.find(q => q.id === j)?.score || 3 });
    const r = await makeRequest(`/plugin/bfi44/tests/${testId}/responses`, 'POST', {
      test_id: testId, student_id: String(113 + i), answers  // system user IDs start at 113
    });
    if (r.success) ok++; else fail++;
    await new Promise(r => setTimeout(r, 20));
  }
  console.log(`Done: ${ok} ok, ${fail} failed`);
  fs.writeFileSync(path.join(__dirname, "bfi44_section3.json"), JSON.stringify({test_id:testId, section_id:SECTION_ID}, null, 2));
}
main().catch(console.error);