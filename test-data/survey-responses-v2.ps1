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

Write-Host "=== 20 AFINIDADES MUTUAS (20 pairs) ==="
# Pares de afinidad: A y B se eligen mutuamente
# A tiene Q1->B(+), Q2->B(+)
# B tiene Q1->A(+), Q2->A(+)
# Adicionalmente, cada estudiante tiene otras 1-2 elecciones adicionales

# Par 1: 11 <-> 113
Post-Response "11"  @(@{student_id="113";response=1},@{student_id="114";response=1}) @(@{student_id="113";response=1}) @() @()
Post-Response "113" @(@{student_id="11";response=1},@{student_id="118";response=1}) @(@{student_id="11";response=1}) @() @()

# Par 2: 114 <-> 115
Post-Response "114" @(@{student_id="115";response=1},@{student_id="11";response=1}) @(@{student_id="115";response=1}) @() @()
Post-Response "115" @(@{student_id="114";response=1},@{student_id="119";response=1}) @(@{student_id="114";response=1}) @() @()

# Par 3: 116 <-> 117
Post-Response "116" @(@{student_id="117";response=1},@{student_id="120";response=1}) @(@{student_id="117";response=1}) @() @()
Post-Response "117" @(@{student_id="116";response=1},@{student_id="121";response=1}) @(@{student_id="116";response=1}) @() @()

# Par 4: 118 <-> 119
Post-Response "118" @(@{student_id="119";response=1},@{student_id="113";response=1}) @(@{student_id="119";response=1}) @() @()
Post-Response "119" @(@{student_id="118";response=1},@{student_id="122";response=1}) @(@{student_id="118";response=1}) @() @()

# Par 5: 120 <-> 121
Post-Response "120" @(@{student_id="121";response=1},@{student_id="116";response=1}) @(@{student_id="121";response=1}) @() @()
Post-Response "121" @(@{student_id="120";response=1},@{student_id="123";response=1}) @(@{student_id="120";response=1}) @() @()

# Par 6: 122 <-> 123
Post-Response "122" @(@{student_id="123";response=1},@{student_id="119";response=1}) @(@{student_id="123";response=1}) @() @()
Post-Response "123" @(@{student_id="122";response=1},@{student_id="124";response=1}) @(@{student_id="122";response=1}) @() @()

# Par 7: 124 <-> 125
Post-Response "124" @(@{student_id="125";response=1},@{student_id="123";response=1}) @(@{student_id="125";response=1}) @() @()
Post-Response "125" @(@{student_id="124";response=1},@{student_id="126";response=1}) @(@{student_id="124";response=1}) @() @()

# Par 8: 126 <-> 127
Post-Response "126" @(@{student_id="127";response=1},@{student_id="125";response=1}) @(@{student_id="127";response=1}) @() @()
Post-Response "127" @(@{student_id="126";response=1},@{student_id="128";response=1}) @(@{student_id="126";response=1}) @() @()

# Par 9: 128 <-> 129
Post-Response "128" @(@{student_id="129";response=1},@{student_id="127";response=1}) @(@{student_id="129";response=1}) @() @()
Post-Response "129" @(@{student_id="128";response=1},@{student_id="130";response=1}) @(@{student_id="128";response=1}) @() @()

# Par 10: 130 <-> 131
Post-Response "130" @(@{student_id="131";response=1},@{student_id="129";response=1}) @(@{student_id="131";response=1}) @() @()
Post-Response "131" @(@{student_id="130";response=1},@{student_id="132";response=1}) @(@{student_id="130";response=1}) @() @()

# Par 11: 132 <-> 133
Post-Response "132" @(@{student_id="133";response=1},@{student_id="131";response=1}) @(@{student_id="133";response=1}) @() @()
Post-Response "133" @(@{student_id="132";response=1},@{student_id="134";response=1}) @(@{student_id="132";response=1}) @() @()

# Par 12: 134 <-> 135
Post-Response "134" @(@{student_id="135";response=1},@{student_id="133";response=1}) @(@{student_id="135";response=1}) @() @()
Post-Response "135" @(@{student_id="134";response=1},@{student_id="136";response=1}) @(@{student_id="134";response=1}) @() @()

# Par 13: 136 <-> 137
Post-Response "136" @(@{student_id="137";response=1},@{student_id="135";response=1}) @(@{student_id="137";response=1}) @() @()
Post-Response "137" @(@{student_id="136";response=1},@{student_id="138";response=1}) @(@{student_id="136";response=1}) @() @()

# Par 14: 138 <-> 139
Post-Response "138" @(@{student_id="139";response=1},@{student_id="137";response=1}) @(@{student_id="139";response=1}) @() @()
Post-Response "139" @(@{student_id="138";response=1},@{student_id="140";response=1}) @(@{student_id="138";response=1}) @() @()

# Par 15: 140 <-> 141
Post-Response "140" @(@{student_id="141";response=1},@{student_id="139";response=1}) @(@{student_id="141";response=1}) @() @()
Post-Response "141" @(@{student_id="140";response=1},@{student_id="142";response=1}) @(@{student_id="140";response=1}) @() @()

# Par 16: 142 <-> 143
Post-Response "142" @(@{student_id="143";response=1},@{student_id="141";response=1}) @(@{student_id="143";response=1}) @() @()
Post-Response "143" @(@{student_id="142";response=1},@{student_id="144";response=1}) @(@{student_id="142";response=1}) @() @()

# Par 17: 144 <-> 145
Post-Response "144" @(@{student_id="145";response=1},@{student_id="143";response=1}) @(@{student_id="145";response=1}) @() @()
Post-Response "145" @(@{student_id="144";response=1},@{student_id="146";response=1}) @(@{student_id="144";response=1}) @() @()

# Par 18: 146 <-> 147
Post-Response "146" @(@{student_id="147";response=1},@{student_id="145";response=1}) @(@{student_id="147";response=1}) @() @()
Post-Response "147" @(@{student_id="146";response=1},@{student_id="148";response=1}) @(@{student_id="146";response=1}) @() @()

# Par 19: 148 <-> 149
Post-Response "148" @(@{student_id="149";response=1},@{student_id="147";response=1}) @(@{student_id="149";response=1}) @() @()
Post-Response "149" @(@{student_id="148";response=1},@{student_id="150";response=1}) @(@{student_id="148";response=1}) @() @()

# Par 20: 150 <-> 151
Post-Response "150" @(@{student_id="151";response=1},@{student_id="149";response=1}) @(@{student_id="151";response=1}) @() @()
Post-Response "151" @(@{student_id="150";response=1},@{student_id="11";response=1}) @(@{student_id="150";response=1}) @() @()

Write-Host ""
Write-Host "=== 20 RECHAZOS MUTUOS ==="
# Pares de rechazo: A y B se rechazan mutuamente
# A tiene Q3->B(+), Q4->B(+)
# B tiene Q3->A(+), Q4->A(+)

# Par 1: 152 <-> 153
Post-Response "152" @() @() @(@{student_id="153";response=1},@{student_id="154";response=1}) @(@{student_id="153";response=1})
Post-Response "153" @() @() @(@{student_id="152";response=1},@{student_id="155";response=1}) @(@{student_id="152";response=1})

# Par 2: 154 <-> 155
Post-Response "154" @() @() @(@{student_id="155";response=1},@{student_id="152";response=1}) @(@{student_id="155";response=1})
Post-Response "155" @() @() @(@{student_id="154";response=1},@{student_id="156";response=1}) @(@{student_id="154";response=1})

# Par 3: 156 <-> 157
Post-Response "156" @() @() @(@{student_id="157";response=1},@{student_id="158";response=1}) @(@{student_id="157";response=1})
Post-Response "157" @() @() @(@{student_id="156";response=1},@{student_id="159";response=1}) @(@{student_id="156";response=1})

# Par 4: 158 <-> 159
Post-Response "158" @() @() @(@{student_id="159";response=1},@{student_id="156";response=1}) @(@{student_id="159";response=1})
Post-Response "159" @() @() @(@{student_id="158";response=1},@{student_id="160";response=1}) @(@{student_id="158";response=1})

# Par 5: 160 <-> 161
Post-Response "160" @() @() @(@{student_id="161";response=1},@{student_id="162";response=1}) @(@{student_id="161";response=1})
Post-Response "161" @() @() @(@{student_id="160";response=1},@{student_id="11";response=1}) @(@{student_id="160";response=1})

# Par 6: 162 <-> 11 (nuevo par de rechazo)
Post-Response "162" @() @() @(@{student_id="11";response=1},@{student_id="114";response=1}) @(@{student_id="11";response=1})

# Par 7: 113 <-> 114 (cambio - ya no son afinidad, ahora rechazo)
Post-Response "113" @() @() @(@{student_id="114";response=1},@{student_id="116";response=1}) @(@{student_id="114";response=1})
Post-Response "114" @() @() @(@{student_id="113";response=1},@{student_id="117";response=1}) @(@{student_id="113";response=1})

# Par 8: 115 <-> 116
Post-Response "115" @() @() @(@{student_id="116";response=1},@{student_id="118";response=1}) @(@{student_id="116";response=1})
Post-Response "116" @() @() @(@{student_id="115";response=1},@{student_id="119";response=1}) @(@{student_id="115";response=1})

# Par 9: 117 <-> 118
Post-Response "117" @() @() @(@{student_id="118";response=1},@{student_id="120";response=1}) @(@{student_id="118";response=1})
Post-Response "118" @() @() @(@{student_id="117";response=1},@{student_id="121";response=1}) @(@{student_id="117";response=1})

# Par 10: 119 <-> 120
Post-Response "119" @() @() @(@{student_id="120";response=1},@{student_id="122";response=1}) @(@{student_id="120";response=1})
Post-Response "120" @() @() @(@{student_id="119";response=1},@{student_id="123";response=1}) @(@{student_id="119";response=1})

# Par 11: 121 <-> 122
Post-Response "121" @() @() @(@{student_id="122";response=1},@{student_id="124";response=1}) @(@{student_id="122";response=1})
Post-Response "122" @() @() @(@{student_id="121";response=1},@{student_id="125";response=1}) @(@{student_id="121";response=1})

# Par 12: 123 <-> 124
Post-Response "123" @() @() @(@{student_id="124";response=1},@{student_id="126";response=1}) @(@{student_id="124";response=1})
Post-Response "124" @() @() @(@{student_id="123";response=1},@{student_id="127";response=1}) @(@{student_id="123";response=1})

# Par 13: 125 <-> 126
Post-Response "125" @() @() @(@{student_id="126";response=1},@{student_id="128";response=1}) @(@{student_id="126";response=1})
Post-Response "126" @() @() @(@{student_id="125";response=1},@{student_id="129";response=1}) @(@{student_id="125";response=1})

# Par 14: 127 <-> 128
Post-Response "127" @() @() @(@{student_id="128";response=1},@{student_id="130";response=1}) @(@{student_id="128";response=1})
Post-Response "128" @() @() @(@{student_id="127";response=1},@{student_id="131";response=1}) @(@{student_id="127";response=1})

# Par 15: 129 <-> 130
Post-Response "129" @() @() @(@{student_id="130";response=1},@{student_id="132";response=1}) @(@{student_id="130";response=1})
Post-Response "130" @() @() @(@{student_id="129";response=1},@{student_id="133";response=1}) @(@{student_id="129";response=1})

# Par 16: 131 <-> 132
Post-Response "131" @() @() @(@{student_id="132";response=1},@{student_id="134";response=1}) @(@{student_id="132";response=1})
Post-Response "132" @() @() @(@{student_id="131";response=1},@{student_id="135";response=1}) @(@{student_id="131";response=1})

# Par 17: 133 <-> 134
Post-Response "133" @() @() @(@{student_id="134";response=1},@{student_id="136";response=1}) @(@{student_id="134";response=1})
Post-Response "134" @() @() @(@{student_id="133";response=1},@{student_id="137";response=1}) @(@{student_id="133";response=1})

# Par 18: 135 <-> 136
Post-Response "135" @() @() @(@{student_id="136";response=1},@{student_id="138";response=1}) @(@{student_id="136";response=1})
Post-Response "136" @() @() @(@{student_id="135";response=1},@{student_id="139";response=1}) @(@{student_id="135";response=1})

# Par 19: 137 <-> 138
Post-Response "137" @() @() @(@{student_id="138";response=1},@{student_id="140";response=1}) @(@{student_id="138";response=1})
Post-Response "138" @() @() @(@{student_id="137";response=1},@{student_id="141";response=1}) @(@{student_id="137";response=1})

# Par 20: 139 <-> 140
Post-Response "139" @() @() @(@{student_id="140";response=1},@{student_id="142";response=1}) @(@{student_id="140";response=1})
Post-Response "140" @() @() @(@{student_id="139";response=1},@{student_id="143";response=1}) @(@{student_id="139";response=1})

Write-Host ""
Write-Host "=== 10 ASIMETRICAS (A elige a B, B rechaza a A) ==="
# A tiene Q1->B(+), Q2->B(+)
# B tiene Q3->A(+), Q4->A(+)

# Par 1: 141 -> 142 (141 likes 142, but 142 rejects 141)
Post-Response "141" @(@{student_id="142";response=1},@{student_id="143";response=1}) @(@{student_id="142";response=1}) @() @()
Post-Response "142" @() @() @(@{student_id="141";response=1},@{student_id="144";response=1}) @(@{student_id="141";response=1})

# Par 2: 143 -> 144
Post-Response "143" @(@{student_id="144";response=1},@{student_id="145";response=1}) @(@{student_id="144";response=1}) @() @()
Post-Response "144" @() @() @(@{student_id="143";response=1},@{student_id="146";response=1}) @(@{student_id="143";response=1})

# Par 3: 145 -> 146
Post-Response "145" @(@{student_id="146";response=1},@{student_id="147";response=1}) @(@{student_id="146";response=1}) @() @()
Post-Response "146" @() @() @(@{student_id="145";response=1},@{student_id="148";response=1}) @(@{student_id="145";response=1})

# Par 4: 147 -> 148
Post-Response "147" @(@{student_id="148";response=1},@{student_id="149";response=1}) @(@{student_id="148";response=1}) @() @()
Post-Response "148" @() @() @(@{student_id="147";response=1},@{student_id="150";response=1}) @(@{student_id="147";response=1})

# Par 5: 149 -> 150
Post-Response "149" @(@{student_id="150";response=1},@{student_id="151";response=1}) @(@{student_id="150";response=1}) @() @()
Post-Response "150" @() @() @(@{student_id="149";response=1},@{student_id="11";response=1}) @(@{student_id="149";response=1})

# Par 6: 151 -> 152
Post-Response "151" @(@{student_id="152";response=1},@{student_id="153";response=1}) @(@{student_id="152";response=1}) @() @()
Post-Response "152" @() @() @(@{student_id="151";response=1},@{student_id="154";response=1}) @(@{student_id="151";response=1})

# Par 7: 153 -> 154
Post-Response "153" @(@{student_id="154";response=1},@{student_id="155";response=1}) @(@{student_id="154";response=1}) @() @()
Post-Response "154" @() @() @(@{student_id="153";response=1},@{student_id="156";response=1}) @(@{student_id="153";response=1})

# Par 8: 155 -> 156
Post-Response "155" @(@{student_id="156";response=1},@{student_id="157";response=1}) @(@{student_id="156";response=1}) @() @()
Post-Response "156" @() @() @(@{student_id="155";response=1},@{student_id="158";response=1}) @(@{student_id="155";response=1})

# Par 9: 157 -> 158
Post-Response "157" @(@{student_id="158";response=1},@{student_id="159";response=1}) @(@{student_id="158";response=1}) @() @()
Post-Response "158" @() @() @(@{student_id="157";response=1},@{student_id="160";response=1}) @(@{student_id="157";response=1})

# Par 10: 159 -> 160
Post-Response "159" @(@{student_id="160";response=1},@{student_id="161";response=1}) @(@{student_id="160";response=1}) @() @()
Post-Response "160" @() @() @(@{student_id="159";response=1},@{student_id="162";response=1}) @(@{student_id="159";response=1})

Write-Host ""
Write-Host "=== DATOS COMPLETOS ==="
Write-Host "50 estudiantes con ~2 interacciones cada uno"
