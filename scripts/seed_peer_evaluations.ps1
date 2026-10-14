$token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODAwNTkwODksImlhdCI6MTc3OTk3MjY4OSwiaWQiOjM3LCJwZXJtaXNzaW9ucyI6WyJBY2NlZGVyIGEgY3Vlc3Rpb25hcmlvIHZpc3RhcyIsIkFjY2VkZXIgYSBjdXJzb3MgdmlzdG9zIiwiQWNjZWRlciBhIGVuY3Vlc3RhcyB2aXN0YXMiLCJBZ3JlZ2FyIGVzdHVkaWFudGUgYSBsYSBzZWNjacOzbiIsIkNyZWFyIGFkbWluaXN0cmFkb3IiLCJDcmVhciBjw7NkaWdvIGRlIGludml0YWNpw7NuIiwiQ3JlYXIgY29vcmRpbmFkb3IiLCJDcmVhciBjdWVzdGlvbmFyaW8iLCJDcmVhciBjdWVzdGlvbmFyaW8gcGFyYSBjdXJzbyIsIkNyZWFyIGN1ZXN0aW9uYXJpbyBwYXJhIGdydXBvIiwiQ3JlYXIgY3Vlc3Rpb25hcmlvIHBhcmEgc2VjY2nDs24iLCJDcmVhciBjdXJzbyIsIkNyZWFyIGVuY3Vlc3RhIHBhcmEgY3Vyc28iLCJDcmVhciBlbmN1ZXN0YSBwYXJhIGdydXBvIiwiQ3JlYXIgZW5jdWVzdGEgcGFyYSBzZWNjacOzbiIsIkNyZWFyIGVzdHVkaWFudGUiLCJDcmVhciBncnVwb3MgYWxlYXRvcmlvcyIsIkNyZWFyIGdydXBvcyBjb24gcGx1Z2lucyIsIkNyZWFyIHByb2Zlc29yIiwiQ3JlYXIgc2VjY2nDs24iLCJkZWxldGVfY291cnNlcyIsImRlbGV0ZV91c2VycyIsIkVkaXRhciBhZG1pbmlzdHJhZG9yIiwiRWRpdGFyIGNvb3JkaW5hZG9yIiwiRWRpdGFyIGN1ZXN0aW9uYXJpbyIsIkVkaXRhciBjdXJzbyIsIkVkaXRhciBlbmN1ZXN0YSIsIkVkaXRhciBlc3R1ZGlhbnRlIiwiRWRpdGFyIGdydXBvcyIsIkVkaXRhciBwcm9mZXNvciIsIkVkaXRhciBzZWNjacOzbiIsIkVkaXRhciBzZW1lc3RyZSBhY3R1YWwiLCJFbGltaW5hciB1c3VhcmlvIiwicmVhZF9jb3Vyc2VzIiwicmVhZF91c2VycyIsIlZlciBncnVwb3MgcHJvcGlvcyIsIlZlciBzZWNjaW9uZXMgcHJvcGlhcyIsIlZlciB0b2RhcyBsYXMgY3Vlc3Rpb25hcmlvIiwiVmVyIHRvZGFzIGxhcyBlbmN1ZXN0YXMiLCJWZXIgdG9kYXMgbGFzIHJlc3B1ZXN0YXMgZW4gY3Vlc3Rpb25hcmlvIiwiVmVyIHRvZGFzIGxhcyByZXNwdWVzdGFzIGVuIGVuY3Vlc3RhIiwiVmVyIHRvZGFzIGxhcyBzZWNjaW9uZXMiLCJWZXIgdG9kb3MgbG9zIGN1cnNvcyIsIndyaXRlX2NvdXJzZXMiLCJ3cml0ZV91c2VycyJdLCJyb2xlIjoiYWRtaW4ifQ.GpseFDZHoFK0st7HtI1UCJG5NIGjrMfFoCOEKibIvDA"
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}
$baseUrl = "http://127.0.0.1:3000/api/v1/core"
$surveyId = "6a1835798d6e093c1c9fb4f3"

# Define 6 groups
$groups = @(
    @{ group_id = "g1"; group_name = "Grupo 1"; group_number = 1; student_ids = @("38","39","40","41","42") },
    @{ group_id = "g2"; group_name = "Grupo 2"; group_number = 2; student_ids = @("43","44","45","46") },
    @{ group_id = "g3"; group_name = "Grupo 3"; group_number = 3; student_ids = @("47","48","49","50","51") },
    @{ group_id = "g4"; group_name = "Grupo 4"; group_number = 4; student_ids = @("52","53","54","55") },
    @{ group_id = "g5"; group_name = "Grupo 5"; group_number = 5; student_ids = @("56","57","58","59") },
    @{ group_id = "g6"; group_name = "Grupo 6"; group_number = 6; student_ids = @("60","61","62","63") }
)

# Step 1: Update survey with group snapshots
Write-Host "=== Step 1: Adding group snapshots to survey ===" -ForegroundColor Cyan
$snapshots = $groups | ForEach-Object {
    @{
        group_id = $_.group_id
        group_name = $_.group_name
        group_number = $_.group_number
        student_ids = $_.student_ids
    }
}

$updateBody = @{
    group_snapshots = $snapshots
} | ConvertTo-Json -Depth 5

$updateResult = Invoke-RestMethod -Uri "$baseUrl/peer-evaluation/surveys/$surveyId" -Method PUT -Headers $headers -Body $updateBody
Write-Host "Survey updated: $($updateResult.success)" -ForegroundColor Green

# Step 2: Submit peer evaluation responses
Write-Host "`n=== Step 2: Submitting peer evaluation responses ===" -ForegroundColor Cyan

$rng = New-Object System.Random(42)
$totalResponses = 0

foreach ($group in $groups) {
    $members = $group.student_ids
    $groupId = $group.group_id
    Write-Host "  Processing $($group.group_name) ($($members.Count) members)..." -ForegroundColor Yellow

    foreach ($evaluator in $members) {
        foreach ($evaluated in $members) {
            if ($evaluator -eq $evaluated) { continue }

            # Generate random scores 1-4 for each of the 5 questions
            $responses = @(
                @{ question_id = 1; score = $rng.Next(1, 5) },
                @{ question_id = 2; score = $rng.Next(1, 5) },
                @{ question_id = 3; score = $rng.Next(1, 5) },
                @{ question_id = 4; score = $rng.Next(1, 5) },
                @{ question_id = 5; score = $rng.Next(1, 5) }
            )

            $body = @{
                survey_id = $surveyId
                evaluator_id = $evaluator
                evaluated_id = $evaluated
                group_id = $groupId
                responses = $responses
            } | ConvertTo-Json -Depth 5

            try {
                $result = Invoke-RestMethod -Uri "$baseUrl/peer-evaluation/surveys/$surveyId/responses" -Method POST -Headers $headers -Body $body
                if ($result.success) {
                    $totalResponses++
                } else {
                    Write-Host "    FAILED: evaluator=$evaluator evaluated=$evaluated - $($result | ConvertTo-Json -Compress)" -ForegroundColor Red
                }
            } catch {
                Write-Host "    ERROR: evaluator=$evaluator evaluated=$evaluated - $($_.Exception.Message)" -ForegroundColor Red
            }
        }
    }
}

Write-Host "`n=== Done! Submitted $totalResponses peer evaluation responses ===" -ForegroundColor Green

# Step 3: Verify results
Write-Host "`n=== Step 3: Verifying results ===" -ForegroundColor Cyan
$results = Invoke-RestMethod -Uri "$baseUrl/peer-evaluation/surveys/$surveyId/results" -Headers $headers
Write-Host "Results:" -ForegroundColor Green
$results | ConvertTo-Json -Depth 5
