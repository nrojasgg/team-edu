$baseUrl = "http://localhost:3000"
$surveyId = "69dfea738b038c43282c1fea"
$sectionId = "3"

$r = Invoke-RestMethod -Method POST -Uri "$baseUrl/api/v1/auth/login" -ContentType "application/json" -Body '{"email":"nicolas.rojas.g@usach.cl","password":"Hola123."}'
$token = $r.data.token
Write-Host "Token obtained"

$headers = @{ "Authorization" = "Bearer $token"; "Content-Type" = "application/json" }
$url = "$baseUrl/api/v1/core/plugin/group_compatibility/tests/$surveyId/responses"

function Post-Response($studentId, $q1, $q2, $q3, $q4) {
    $body = @{
        survey_id = $surveyId
        student_id = $studentId
        section_id = $sectionId
        survey_type = "sociometric"
        question1 = $q1
        question2 = $q2
        question3 = $q3
        question4 = $q4
    }
    try {
        $resp = Invoke-RestMethod -Uri $url -Method POST -Headers $headers -Body ($body | ConvertTo-Json -Compress)
        Write-Host "[OK] Student $studentId"
    } catch {
        Write-Host "[FAIL] Student $studentId : $($_.Exception.Message)"
    }
}

# Estudiantes: 11,113,114,115,116,117,118,119,120,121,122,123,124,125,126,127,128,129,130,131,132,133,134,135,136,137,138,139,140,141,142,143,144,145,146,147,148,149,150,151,152,153,154,155,156,157,158,159,160,161,162
# 51 estudiantes = 50 completos + 162

Write-Host "=== CREANDO DATOS CON MULTIPLES RELACIONES ==="

# Estudiante 11 - Afinidad con 113,114,115,126,127 y Rechazo con 152,153,154,128,129
Post-Response "11" `
    @(@{student_id="113";response=1},@{student_id="114";response=1},@{student_id="115";response=1},@{student_id="126";response=1},@{student_id="127";response=1},@{student_id="118";response=1}) `
    @(@{student_id="113";response=1},@{student_id="114";response=1},@{student_id="126";response=1}) `
    @(@{student_id="152";response=1},@{student_id="153";response=1},@{student_id="154";response=1},@{student_id="128";response=1},@{student_id="129";response=1}) `
    @(@{student_id="152";response=1},@{student_id="153";response=1},@{student_id="128";response=1})

# Estudiante 113 - Afinidad con 11,114,115,116 y Rechazo con 153,154,155,130,131
Post-Response "113" `
    @(@{student_id="11";response=1},@{student_id="114";response=1},@{student_id="115";response=1},@{student_id="116";response=1},@{student_id="119";response=1},@{student_id="120";response=1}) `
    @(@{student_id="11";response=1},@{student_id="114";response=1},@{student_id="115";response=1}) `
    @(@{student_id="153";response=1},@{student_id="154";response=1},@{student_id="155";response=1},@{student_id="130";response=1},@{student_id="131";response=1}) `
    @(@{student_id="154";response=1},@{student_id="155";response=1},@{student_id="130";response=1})

# Estudiante 114 - Afinidad con 11,113,115,117 y Rechazo con 152,156,157,132,133
Post-Response "114" `
    @(@{student_id="11";response=1},@{student_id="113";response=1},@{student_id="115";response=1},@{student_id="117";response=1},@{student_id="121";response=1},@{student_id="122";response=1}) `
    @(@{student_id="113";response=1},@{student_id="115";response=1},@{student_id="117";response=1}) `
    @(@{student_id="152";response=1},@{student_id="156";response=1},@{student_id="157";response=1},@{student_id="132";response=1},@{student_id="133";response=1}) `
    @(@{student_id="156";response=1},@{student_id="157";response=1},@{student_id="132";response=1})

# Estudiante 115 - Afinidad con 11,113,114,118 y Rechazo con 153,158,159,134,135
Post-Response "115" `
    @(@{student_id="11";response=1},@{student_id="113";response=1},@{student_id="114";response=1},@{student_id="118";response=1},@{student_id="123";response=1},@{student_id="124";response=1}) `
    @(@{student_id="114";response=1},@{student_id="118";response=1},@{student_id="11";response=1}) `
    @(@{student_id="153";response=1},@{student_id="158";response=1},@{student_id="159";response=1},@{student_id="134";response=1},@{student_id="135";response=1}) `
    @(@{student_id="158";response=1},@{student_id="159";response=1},@{student_id="134";response=1})

# Estudiante 116 - Afinidad con 113,117,119,120 y Rechazo con 152,154,160,161,136
Post-Response "116" `
    @(@{student_id="113";response=1},@{student_id="117";response=1},@{student_id="119";response=1},@{student_id="120";response=1},@{student_id="125";response=1},@{student_id="126";response=1}) `
    @(@{student_id="117";response=1},@{student_id="119";response=1},@{student_id="120";response=1}) `
    @(@{student_id="152";response=1},@{student_id="154";response=1},@{student_id="160";response=1},@{student_id="161";response=1},@{student_id="136";response=1}) `
    @(@{student_id="160";response=1},@{student_id="161";response=1},@{student_id="136";response=1})

# Estudiante 117 - Afinidad con 114,116,118,119 y Rechazo con 153,155,157,158,137
Post-Response "117" `
    @(@{student_id="114";response=1},@{student_id="116";response=1},@{student_id="118";response=1},@{student_id="119";response=1},@{student_id="127";response=1},@{student_id="128";response=1}) `
    @(@{student_id="116";response=1},@{student_id="118";response=1},@{student_id="119";response=1}) `
    @(@{student_id="153";response=1},@{student_id="155";response=1},@{student_id="157";response=1},@{student_id="158";response=1},@{student_id="137";response=1}) `
    @(@{student_id="157";response=1},@{student_id="158";response=1},@{student_id="137";response=1})

# Estudiante 118 - Afinidad con 115,117,119,121 y Rechazo con 152,156,159,160,138
Post-Response "118" `
    @(@{student_id="115";response=1},@{student_id="117";response=1},@{student_id="119";response=1},@{student_id="121";response=1},@{student_id="129";response=1},@{student_id="130";response=1}) `
    @(@{student_id="117";response=1},@{student_id="119";response=1},@{student_id="121";response=1}) `
    @(@{student_id="152";response=1},@{student_id="156";response=1},@{student_id="159";response=1},@{student_id="160";response=1},@{student_id="138";response=1}) `
    @(@{student_id="159";response=1},@{student_id="160";response=1},@{student_id="138";response=1})

# Estudiante 119 - Afinidad con 116,118,120,122 y Rechazo con 153,154,157,161,139
Post-Response "119" `
    @(@{student_id="116";response=1},@{student_id="118";response=1},@{student_id="120";response=1},@{student_id="122";response=1},@{student_id="131";response=1},@{student_id="132";response=1}) `
    @(@{student_id="118";response=1},@{student_id="120";response=1},@{student_id="122";response=1}) `
    @(@{student_id="153";response=1},@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="161";response=1},@{student_id="139";response=1}) `
    @(@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="161";response=1})

# Estudiante 120 - Afinidad con 116,119,121,123 y Rechazo con 152,155,158,159,140
Post-Response "120" `
    @(@{student_id="116";response=1},@{student_id="119";response=1},@{student_id="121";response=1},@{student_id="123";response=1},@{student_id="133";response=1},@{student_id="134";response=1}) `
    @(@{student_id="119";response=1},@{student_id="121";response=1},@{student_id="123";response=1}) `
    @(@{student_id="152";response=1},@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="159";response=1},@{student_id="140";response=1}) `
    @(@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="159";response=1})

# Estudiante 121 - Afinidad con 117,120,122,124 y Rechazo con 153,156,160,161,141
Post-Response "121" `
    @(@{student_id="117";response=1},@{student_id="120";response=1},@{student_id="122";response=1},@{student_id="124";response=1},@{student_id="135";response=1},@{student_id="136";response=1}) `
    @(@{student_id="120";response=1},@{student_id="122";response=1},@{student_id="124";response=1}) `
    @(@{student_id="153";response=1},@{student_id="156";response=1},@{student_id="160";response=1},@{student_id="161";response=1},@{student_id="141";response=1}) `
    @(@{student_id="156";response=1},@{student_id="160";response=1},@{student_id="161";response=1})

# Estudiante 122 - Afinidad con 118,121,123,125 y Rechazo con 152,154,157,158,142
Post-Response "122" `
    @(@{student_id="118";response=1},@{student_id="121";response=1},@{student_id="123";response=1},@{student_id="125";response=1},@{student_id="137";response=1},@{student_id="138";response=1}) `
    @(@{student_id="121";response=1},@{student_id="123";response=1},@{student_id="125";response=1}) `
    @(@{student_id="152";response=1},@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="158";response=1},@{student_id="142";response=1}) `
    @(@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="158";response=1})

# Estudiante 123 - Afinidad con 119,122,124,126 y Rechazo con 153,155,159,160,143
Post-Response "123" `
    @(@{student_id="119";response=1},@{student_id="122";response=1},@{student_id="124";response=1},@{student_id="126";response=1},@{student_id="139";response=1},@{student_id="140";response=1}) `
    @(@{student_id="122";response=1},@{student_id="124";response=1},@{student_id="126";response=1}) `
    @(@{student_id="153";response=1},@{student_id="155";response=1},@{student_id="159";response=1},@{student_id="160";response=1},@{student_id="143";response=1}) `
    @(@{student_id="155";response=1},@{student_id="159";response=1},@{student_id="160";response=1})

# Estudiante 124 - Afinidad con 120,123,125,127 y Rechazo con 152,156,161,162,144
Post-Response "124" `
    @(@{student_id="120";response=1},@{student_id="123";response=1},@{student_id="125";response=1},@{student_id="127";response=1},@{student_id="141";response=1},@{student_id="142";response=1}) `
    @(@{student_id="123";response=1},@{student_id="125";response=1},@{student_id="127";response=1}) `
    @(@{student_id="152";response=1},@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1},@{student_id="144";response=1}) `
    @(@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1})

# Estudiante 125 - Afinidad con 121,124,126,128 y Rechazo con 153,154,157,158,145
Post-Response "125" `
    @(@{student_id="121";response=1},@{student_id="124";response=1},@{student_id="126";response=1},@{student_id="128";response=1},@{student_id="143";response=1},@{student_id="144";response=1}) `
    @(@{student_id="124";response=1},@{student_id="126";response=1},@{student_id="128";response=1}) `
    @(@{student_id="153";response=1},@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="158";response=1},@{student_id="145";response=1}) `
    @(@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="158";response=1})

# Estudiante 126 - Afinidad con 122,125,127,129 y Rechazo con 152,155,159,160,146
Post-Response "126" `
    @(@{student_id="122";response=1},@{student_id="125";response=1},@{student_id="127";response=1},@{student_id="129";response=1},@{student_id="145";response=1},@{student_id="146";response=1}) `
    @(@{student_id="125";response=1},@{student_id="127";response=1},@{student_id="129";response=1}) `
    @(@{student_id="152";response=1},@{student_id="155";response=1},@{student_id="159";response=1},@{student_id="160";response=1},@{student_id="146";response=1}) `
    @(@{student_id="155";response=1},@{student_id="159";response=1},@{student_id="160";response=1})

# Estudiante 127 - Afinidad con 123,126,128,130 y Rechazo con 153,156,161,162,147
Post-Response "127" `
    @(@{student_id="123";response=1},@{student_id="126";response=1},@{student_id="128";response=1},@{student_id="130";response=1},@{student_id="147";response=1},@{student_id="148";response=1}) `
    @(@{student_id="126";response=1},@{student_id="128";response=1},@{student_id="130";response=1}) `
    @(@{student_id="153";response=1},@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1},@{student_id="147";response=1}) `
    @(@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1})

# Estudiante 128 - Afinidad con 124,127,129,131 y Rechazo con 152,154,157,159,148
Post-Response "128" `
    @(@{student_id="124";response=1},@{student_id="127";response=1},@{student_id="129";response=1},@{student_id="131";response=1},@{student_id="148";response=1},@{student_id="149";response=1}) `
    @(@{student_id="127";response=1},@{student_id="129";response=1},@{student_id="131";response=1}) `
    @(@{student_id="152";response=1},@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1},@{student_id="148";response=1}) `
    @(@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1})

# Estudiante 129 - Afinidad con 125,128,130,132 y Rechazo con 153,155,158,160,149
Post-Response "129" `
    @(@{student_id="125";response=1},@{student_id="128";response=1},@{student_id="130";response=1},@{student_id="132";response=1},@{student_id="149";response=1},@{student_id="150";response=1}) `
    @(@{student_id="128";response=1},@{student_id="130";response=1},@{student_id="132";response=1}) `
    @(@{student_id="153";response=1},@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1},@{student_id="149";response=1}) `
    @(@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1})

# Estudiante 130 - Afinidad con 126,129,131,133 y Rechazo con 152,156,161,162,150
Post-Response "130" `
    @(@{student_id="126";response=1},@{student_id="129";response=1},@{student_id="131";response=1},@{student_id="133";response=1},@{student_id="150";response=1},@{student_id="151";response=1}) `
    @(@{student_id="129";response=1},@{student_id="131";response=1},@{student_id="133";response=1}) `
    @(@{student_id="152";response=1},@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1},@{student_id="150";response=1}) `
    @(@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1})

# Estudiante 131 - Afinidad con 127,130,132,134 y Rechazo con 153,154,157,159,151
Post-Response "131" `
    @(@{student_id="127";response=1},@{student_id="130";response=1},@{student_id="132";response=1},@{student_id="134";response=1},@{student_id="151";response=1},@{student_id="11";response=1}) `
    @(@{student_id="130";response=1},@{student_id="132";response=1},@{student_id="134";response=1}) `
    @(@{student_id="153";response=1},@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1},@{student_id="151";response=1}) `
    @(@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1})

# Estudiante 132 - Afinidad con 128,131,133,135 y Rechazo con 152,155,158,160,162
Post-Response "132" `
    @(@{student_id="128";response=1},@{student_id="131";response=1},@{student_id="133";response=1},@{student_id="135";response=1},@{student_id="162";response=1},@{student_id="113";response=1}) `
    @(@{student_id="131";response=1},@{student_id="133";response=1},@{student_id="135";response=1}) `
    @(@{student_id="152";response=1},@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1},@{student_id="162";response=1}) `
    @(@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1})

# Estudiante 133 - Afinidad con 129,132,134,136 y Rechazo con 153,156,161,11,114
Post-Response "133" `
    @(@{student_id="129";response=1},@{student_id="132";response=1},@{student_id="134";response=1},@{student_id="136";response=1},@{student_id="11";response=1},@{student_id="114";response=1}) `
    @(@{student_id="132";response=1},@{student_id="134";response=1},@{student_id="136";response=1}) `
    @(@{student_id="153";response=1},@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="11";response=1},@{student_id="114";response=1}) `
    @(@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="11";response=1})

# Estudiante 134 - Afinidad con 130,133,135,137 y Rechazo con 152,154,157,159,115
Post-Response "134" `
    @(@{student_id="130";response=1},@{student_id="133";response=1},@{student_id="135";response=1},@{student_id="137";response=1},@{student_id="115";response=1},@{student_id="117";response=1}) `
    @(@{student_id="133";response=1},@{student_id="135";response=1},@{student_id="137";response=1}) `
    @(@{student_id="152";response=1},@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1},@{student_id="115";response=1}) `
    @(@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1})

# Estudiante 135 - Afinidad con 131,134,136,138 y Rechazo con 153,155,158,160,116
Post-Response "135" `
    @(@{student_id="131";response=1},@{student_id="134";response=1},@{student_id="136";response=1},@{student_id="138";response=1},@{student_id="116";response=1},@{student_id="118";response=1}) `
    @(@{student_id="134";response=1},@{student_id="136";response=1},@{student_id="138";response=1}) `
    @(@{student_id="153";response=1},@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1},@{student_id="116";response=1}) `
    @(@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1})

# Estudiante 136 - Afinidad con 132,135,137,139 y Rechazo con 152,156,161,162,119
Post-Response "136" `
    @(@{student_id="132";response=1},@{student_id="135";response=1},@{student_id="137";response=1},@{student_id="139";response=1},@{student_id="119";response=1},@{student_id="121";response=1}) `
    @(@{student_id="135";response=1},@{student_id="137";response=1},@{student_id="139";response=1}) `
    @(@{student_id="152";response=1},@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1},@{student_id="119";response=1}) `
    @(@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1})

# Estudiante 137 - Afinidad con 133,136,138,140 y Rechazo con 153,154,157,159,120
Post-Response "137" `
    @(@{student_id="133";response=1},@{student_id="136";response=1},@{student_id="138";response=1},@{student_id="140";response=1},@{student_id="120";response=1},@{student_id="122";response=1}) `
    @(@{student_id="136";response=1},@{student_id="138";response=1},@{student_id="140";response=1}) `
    @(@{student_id="153";response=1},@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1},@{student_id="120";response=1}) `
    @(@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1})

# Estudiante 138 - Afinidad con 134,137,139,141 y Rechazo con 152,155,158,160,123
Post-Response "138" `
    @(@{student_id="134";response=1},@{student_id="137";response=1},@{student_id="139";response=1},@{student_id="141";response=1},@{student_id="123";response=1},@{student_id="125";response=1}) `
    @(@{student_id="137";response=1},@{student_id="139";response=1},@{student_id="141";response=1}) `
    @(@{student_id="152";response=1},@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1},@{student_id="123";response=1}) `
    @(@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1})

# Estudiante 139 - Afinidad con 135,138,140,142 y Rechazo con 153,156,161,162,124
Post-Response "139" `
    @(@{student_id="135";response=1},@{student_id="138";response=1},@{student_id="140";response=1},@{student_id="142";response=1},@{student_id="124";response=1},@{student_id="126";response=1}) `
    @(@{student_id="138";response=1},@{student_id="140";response=1},@{student_id="142";response=1}) `
    @(@{student_id="153";response=1},@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1},@{student_id="124";response=1}) `
    @(@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1})

# Estudiante 140 - Afinidad con 136,139,141,143 y Rechazo con 152,154,157,159,127
Post-Response "140" `
    @(@{student_id="136";response=1},@{student_id="139";response=1},@{student_id="141";response=1},@{student_id="143";response=1},@{student_id="127";response=1},@{student_id="129";response=1}) `
    @(@{student_id="139";response=1},@{student_id="141";response=1},@{student_id="143";response=1}) `
    @(@{student_id="152";response=1},@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1},@{student_id="127";response=1}) `
    @(@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1})

# Estudiante 141 - Afinidad con 137,140,142,144 y Rechazo con 153,155,158,160,128
Post-Response "141" `
    @(@{student_id="137";response=1},@{student_id="140";response=1},@{student_id="142";response=1},@{student_id="144";response=1},@{student_id="128";response=1},@{student_id="130";response=1}) `
    @(@{student_id="140";response=1},@{student_id="142";response=1},@{student_id="144";response=1}) `
    @(@{student_id="153";response=1},@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1},@{student_id="128";response=1}) `
    @(@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1})

# Estudiante 142 - Afinidad con 138,141,143,145 y Rechazo con 152,156,161,162,131
Post-Response "142" `
    @(@{student_id="138";response=1},@{student_id="141";response=1},@{student_id="143";response=1},@{student_id="145";response=1},@{student_id="131";response=1},@{student_id="133";response=1}) `
    @(@{student_id="141";response=1},@{student_id="143";response=1},@{student_id="145";response=1}) `
    @(@{student_id="152";response=1},@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1},@{student_id="131";response=1}) `
    @(@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1})

# Estudiante 143 - Afinidad con 139,142,144,146 y Rechazo con 153,154,157,159,132
Post-Response "143" `
    @(@{student_id="139";response=1},@{student_id="142";response=1},@{student_id="144";response=1},@{student_id="146";response=1},@{student_id="132";response=1},@{student_id="134";response=1}) `
    @(@{student_id="142";response=1},@{student_id="144";response=1},@{student_id="146";response=1}) `
    @(@{student_id="153";response=1},@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1},@{student_id="132";response=1}) `
    @(@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1})

# Estudiante 144 - Afinidad con 140,143,145,147 y Rechazo con 152,155,158,160,135
Post-Response "144" `
    @(@{student_id="140";response=1},@{student_id="143";response=1},@{student_id="145";response=1},@{student_id="147";response=1},@{student_id="135";response=1},@{student_id="137";response=1}) `
    @(@{student_id="143";response=1},@{student_id="145";response=1},@{student_id="147";response=1}) `
    @(@{student_id="152";response=1},@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1},@{student_id="135";response=1}) `
    @(@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1})

# Estudiante 145 - Afinidad con 141,144,146,148 y Rechazo con 153,156,161,162,136
Post-Response "145" `
    @(@{student_id="141";response=1},@{student_id="144";response=1},@{student_id="146";response=1},@{student_id="148";response=1},@{student_id="136";response=1},@{student_id="138";response=1}) `
    @(@{student_id="144";response=1},@{student_id="146";response=1},@{student_id="148";response=1}) `
    @(@{student_id="153";response=1},@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1},@{student_id="136";response=1}) `
    @(@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1})

# Estudiante 146 - Afinidad con 142,145,147,149 y Rechazo con 152,154,157,159,139
Post-Response "146" `
    @(@{student_id="142";response=1},@{student_id="145";response=1},@{student_id="147";response=1},@{student_id="149";response=1},@{student_id="139";response=1},@{student_id="141";response=1}) `
    @(@{student_id="145";response=1},@{student_id="147";response=1},@{student_id="149";response=1}) `
    @(@{student_id="152";response=1},@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1},@{student_id="139";response=1}) `
    @(@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1})

# Estudiante 147 - Afinidad con 143,146,148,150 y Rechazo con 153,155,158,160,140
Post-Response "147" `
    @(@{student_id="143";response=1},@{student_id="146";response=1},@{student_id="148";response=1},@{student_id="150";response=1},@{student_id="140";response=1},@{student_id="142";response=1}) `
    @(@{student_id="146";response=1},@{student_id="148";response=1},@{student_id="150";response=1}) `
    @(@{student_id="153";response=1},@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1},@{student_id="140";response=1}) `
    @(@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1})

# Estudiante 148 - Afinidad con 144,147,149,151 y Rechazo con 152,156,161,162,143
Post-Response "148" `
    @(@{student_id="144";response=1},@{student_id="147";response=1},@{student_id="149";response=1},@{student_id="151";response=1},@{student_id="143";response=1},@{student_id="145";response=1}) `
    @(@{student_id="147";response=1},@{student_id="149";response=1},@{student_id="151";response=1}) `
    @(@{student_id="152";response=1},@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1},@{student_id="143";response=1}) `
    @(@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1})

# Estudiante 149 - Afinidad con 145,148,150,11 y Rechazo con 153,154,157,159,146
Post-Response "149" `
    @(@{student_id="145";response=1},@{student_id="148";response=1},@{student_id="150";response=1},@{student_id="11";response=1},@{student_id="146";response=1},@{student_id="147";response=1}) `
    @(@{student_id="148";response=1},@{student_id="150";response=1},@{student_id="11";response=1}) `
    @(@{student_id="153";response=1},@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1},@{student_id="146";response=1}) `
    @(@{student_id="154";response=1},@{student_id="157";response=1},@{student_id="159";response=1})

# Estudiante 150 - Afinidad con 146,149,151,113 y Rechazo con 152,155,158,160,147
Post-Response "150" `
    @(@{student_id="146";response=1},@{student_id="149";response=1},@{student_id="151";response=1},@{student_id="113";response=1},@{student_id="147";response=1},@{student_id="148";response=1}) `
    @(@{student_id="149";response=1},@{student_id="151";response=1},@{student_id="113";response=1}) `
    @(@{student_id="152";response=1},@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1},@{student_id="147";response=1}) `
    @(@{student_id="155";response=1},@{student_id="158";response=1},@{student_id="160";response=1})

# Estudiante 151 - Afinidad con 147,150,113,114 y Rechazo con 153,156,161,162,149
Post-Response "151" `
    @(@{student_id="147";response=1},@{student_id="150";response=1},@{student_id="113";response=1},@{student_id="114";response=1},@{student_id="149";response=1},@{student_id="11";response=1}) `
    @(@{student_id="150";response=1},@{student_id="113";response=1},@{student_id="114";response=1}) `
    @(@{student_id="153";response=1},@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1},@{student_id="149";response=1}) `
    @(@{student_id="156";response=1},@{student_id="161";response=1},@{student_id="162";response=1})

# Estudiante 152 - Afinidad con 154,155,156,157 y Rechazo con 11,12,13,14,15 (rechaza a los que lo rechazaron)
Post-Response "152" `
    @(@{student_id="154";response=1},@{student_id="155";response=1},@{student_id="156";response=1},@{student_id="157";response=1},@{student_id="158";response=1},@{student_id="159";response=1}) `
    @(@{student_id="155";response=1},@{student_id="156";response=1},@{student_id="157";response=1}) `
    @(@{student_id="11";response=1},@{student_id="12";response=1},@{student_id="13";response=1},@{student_id="14";response=1},@{student_id="15";response=1}) `
    @(@{student_id="12";response=1},@{student_id="13";response=1},@{student_id="14";response=1})

# Estudiante 153 - Afinidad con 158,159,160,161 y Rechazo con 11,113,114,115,116
Post-Response "153" `
    @(@{student_id="158";response=1},@{student_id="159";response=1},@{student_id="160";response=1},@{student_id="161";response=1},@{student_id="162";response=1},@{student_id="11";response=1}) `
    @(@{student_id="159";response=1},@{student_id="160";response=1},@{student_id="161";response=1}) `
    @(@{student_id="11";response=1},@{student_id="113";response=1},@{student_id="114";response=1},@{student_id="115";response=1},@{student_id="116";response=1}) `
    @(@{student_id="113";response=1},@{student_id="114";response=1},@{student_id="115";response=1})

# Estudiante 154 - Afinidad con 160,161,162,11 y Rechazo con 113,117,118,119,120
Post-Response "154" `
    @(@{student_id="160";response=1},@{student_id="161";response=1},@{student_id="162";response=1},@{student_id="11";response=1},@{student_id="113";response=1},@{student_id="114";response=1}) `
    @(@{student_id="161";response=1},@{student_id="162";response=1},@{student_id="11";response=1}) `
    @(@{student_id="113";response=1},@{student_id="117";response=1},@{student_id="118";response=1},@{student_id="119";response=1},@{student_id="120";response=1}) `
    @(@{student_id="117";response=1},@{student_id="118";response=1},@{student_id="119";response=1})

# Estudiante 155 - Afinidad con 11,113,162,114 y Rechazo con 116,117,121,122,123
Post-Response "155" `
    @(@{student_id="11";response=1},@{student_id="113";response=1},@{student_id="162";response=1},@{student_id="114";response=1},@{student_id="115";response=1},@{student_id="116";response=1}) `
    @(@{student_id="113";response=1},@{student_id="162";response=1},@{student_id="114";response=1}) `
    @(@{student_id="116";response=1},@{student_id="117";response=1},@{student_id="121";response=1},@{student_id="122";response=1},@{student_id="123";response=1}) `
    @(@{student_id="117";response=1},@{student_id="121";response=1},@{student_id="122";response=1})

# Estudiante 156 - Afinidad con 11,115,116,117 y Rechazo con 118,119,124,125,126
Post-Response "156" `
    @(@{student_id="11";response=1},@{student_id="115";response=1},@{student_id="116";response=1},@{student_id="117";response=1},@{student_id="118";response=1},@{student_id="119";response=1}) `
    @(@{student_id="115";response=1},@{student_id="116";response=1},@{student_id="117";response=1}) `
    @(@{student_id="118";response=1},@{student_id="119";response=1},@{student_id="124";response=1},@{student_id="125";response=1},@{student_id="126";response=1}) `
    @(@{student_id="124";response=1},@{student_id="125";response=1},@{student_id="126";response=1})

# Estudiante 157 - Afinidad con 118,119,120,121 y Rechazo con 11,113,127,128,129
Post-Response "157" `
    @(@{student_id="118";response=1},@{student_id="119";response=1},@{student_id="120";response=1},@{student_id="121";response=1},@{student_id="122";response=1},@{student_id="123";response=1}) `
    @(@{student_id="119";response=1},@{student_id="120";response=1},@{student_id="121";response=1}) `
    @(@{student_id="11";response=1},@{student_id="113";response=1},@{student_id="127";response=1},@{student_id="128";response=1},@{student_id="129";response=1}) `
    @(@{student_id="113";response=1},@{student_id="127";response=1},@{student_id="128";response=1})

# Estudiante 158 - Afinidad con 122,123,124,125 y Rechazo con 11,115,130,131,132
Post-Response "158" `
    @(@{student_id="122";response=1},@{student_id="123";response=1},@{student_id="124";response=1},@{student_id="125";response=1},@{student_id="126";response=1},@{student_id="127";response=1}) `
    @(@{student_id="123";response=1},@{student_id="124";response=1},@{student_id="125";response=1}) `
    @(@{student_id="11";response=1},@{student_id="115";response=1},@{student_id="130";response=1},@{student_id="131";response=1},@{student_id="132";response=1}) `
    @(@{student_id="115";response=1},@{student_id="130";response=1},@{student_id="131";response=1})

# Estudiante 159 - Afinidad con 126,127,128,129 y Rechazo con 11,116,133,134,135
Post-Response "159" `
    @(@{student_id="126";response=1},@{student_id="127";response=1},@{student_id="128";response=1},@{student_id="129";response=1},@{student_id="130";response=1},@{student_id="131";response=1}) `
    @(@{student_id="127";response=1},@{student_id="128";response=1},@{student_id="129";response=1}) `
    @(@{student_id="11";response=1},@{student_id="116";response=1},@{student_id="133";response=1},@{student_id="134";response=1},@{student_id="135";response=1}) `
    @(@{student_id="116";response=1},@{student_id="133";response=1},@{student_id="134";response=1})

# Estudiante 160 - Afinidad con 130,131,132,133 y Rechazo con 11,117,136,137,138
Post-Response "160" `
    @(@{student_id="130";response=1},@{student_id="131";response=1},@{student_id="132";response=1},@{student_id="133";response=1},@{student_id="134";response=1},@{student_id="135";response=1}) `
    @(@{student_id="131";response=1},@{student_id="132";response=1},@{student_id="133";response=1}) `
    @(@{student_id="11";response=1},@{student_id="117";response=1},@{student_id="136";response=1},@{student_id="137";response=1},@{student_id="138";response=1}) `
    @(@{student_id="117";response=1},@{student_id="136";response=1},@{student_id="137";response=1})

# Estudiante 161 - Afinidad con 134,135,136,137 y Rechazo con 11,118,139,140,141
Post-Response "161" `
    @(@{student_id="134";response=1},@{student_id="135";response=1},@{student_id="136";response=1},@{student_id="137";response=1},@{student_id="138";response=1},@{student_id="139";response=1}) `
    @(@{student_id="135";response=1},@{student_id="136";response=1},@{student_id="137";response=1}) `
    @(@{student_id="11";response=1},@{student_id="118";response=1},@{student_id="139";response=1},@{student_id="140";response=1},@{student_id="141";response=1}) `
    @(@{student_id="118";response=1},@{student_id="139";response=1},@{student_id="140";response=1})

# Estudiante 162 - Afinidad con 138,139,140,141 y Rechazo con 11,119,142,143,144
Post-Response "162" `
    @(@{student_id="138";response=1},@{student_id="139";response=1},@{student_id="140";response=1},@{student_id="141";response=1},@{student_id="142";response=1},@{student_id="143";response=1}) `
    @(@{student_id="139";response=1},@{student_id="140";response=1},@{student_id="141";response=1}) `
    @(@{student_id="11";response=1},@{student_id="119";response=1},@{student_id="142";response=1},@{student_id="143";response=1},@{student_id="144";response=1}) `
    @(@{student_id="119";response=1},@{student_id="142";response=1},@{student_id="143";response=1})

Write-Host ""
Write-Host "=== DATOS CREADOS ==="
Write-Host "51 estudiantes con ~6 elecciones y ~5 rechazos cada uno"
