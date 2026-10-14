import urllib.request
import json

# 1. Login to API Gateway
login_req = urllib.request.Request(
    'http://127.0.0.1:3000/api/v1/auth/login',
    data=json.dumps({'email': 'nicolas.rojas.g@usach.cl', 'password': 'Hola123.'}).encode(),
    headers={'Content-Type': 'application/json'}
)
with urllib.request.urlopen(login_req) as resp:
    token = json.loads(resp.read().decode())['data']['token']

headers = {
    'Content-Type': 'application/json',
    'Authorization': f'Bearer {token}'
}

def call_api(endpoint, payload):
    req = urllib.request.Request(
        f'http://127.0.0.1:3000/api/v1/core/team-formation/{endpoint}',
        data=json.dumps(payload).encode(),
        headers=headers
    )
    with urllib.request.urlopen(req) as resp:
        return json.loads(resp.read().decode())

# Pool de 12 estudiantes que respondieron sus cuestionarios (excluyendo pendientes)
candidates = [str(i) for i in range(38, 50)]

print("=== TEST 1: 1 Criterio Psicológico (Conscientiousness) ===")
p1 = {
    'section_id': '3',
    'candidates': candidates,
    'team_size': 3,
    'criteria_parameters': {'conscientiousness': {'target': 0.8, 'weight': 1.0, 'goal': 'heterogeneity'}},
    'demographic_data': {c: {'numeric_attrs': {'conscientiousness': 0.2 + (int(c) % 5) * 0.18}} for c in candidates}
}
r1 = call_api('form-flexible', p1)
print("Success:", r1.get('success'), "| Equipos:", len(r1.get('data', {}).get('teams', [])), "| Global:", round(r1.get('data', {}).get('global_score', 0), 4))

print("\n=== TEST 2: 2 Criterios Psicológicos (Conscientiousness + Extraversion) ===")
p2 = {
    'section_id': '3',
    'candidates': candidates,
    'team_size': 3,
    'criteria_parameters': {
        'conscientiousness': {'target': 0.8, 'weight': 0.5, 'goal': 'heterogeneity'},
        'extraversion': {'target': 0.2, 'weight': 0.5, 'goal': 'homogeneity'}
    },
    'demographic_data': {c: {'numeric_attrs': {'conscientiousness': 0.2 + (int(c) % 5) * 0.18, 'extraversion': 0.5 + (int(c) % 3) * 0.1}} for c in candidates}
}
r2 = call_api('form-flexible', p2)
print("Success:", r2.get('success'), "| Equipos:", len(r2.get('data', {}).get('teams', [])), "| Global:", round(r2.get('data', {}).get('global_score', 0), 4))

print("\n=== TEST 3: 3 Criterios Psicológicos (Consc + Extra + Openness) ===")
p3 = {
    'section_id': '3',
    'candidates': candidates,
    'team_size': 3,
    'criteria_parameters': {
        'conscientiousness': {'target': 0.8, 'weight': 0.4, 'goal': 'heterogeneity'},
        'extraversion': {'target': 0.2, 'weight': 0.3, 'goal': 'homogeneity'},
        'openness': {'target': 0.5, 'weight': 0.3, 'goal': 'target'}
    },
    'demographic_data': {c: {'numeric_attrs': {'conscientiousness': 0.2 + (int(c) % 5) * 0.18, 'extraversion': 0.5 + (int(c) % 3) * 0.1, 'openness': 0.6}} for c in candidates}
}
r3 = call_api('form-flexible', p3)
print("Success:", r3.get('success'), "| Equipos:", len(r3.get('data', {}).get('teams', [])), "| Global:", round(r3.get('data', {}).get('global_score', 0), 4))

print("\n=== TEST 4: Solo Criterio Social ===")
p4 = {
    'section_id': '3',
    'candidates': candidates,
    'team_size': 3,
    'criteria_parameters': {'social': {'target': 0.5, 'weight': 1.0, 'type': 'social'}},
    'demographic_data': {c: {'social_connections': {str(38 + (int(c) + 1) % 12): 1.0, str(38 + (int(c) + 2) % 12): -1.0}} for c in candidates}
}
r4 = call_api('form-flexible', p4)
print("Success:", r4.get('success'), "| Equipos:", len(r4.get('data', {}).get('teams', [])), "| Global:", round(r4.get('data', {}).get('global_score', 0), 4))

print("\n=== TEST 5: Criterio Categórico (Índice de Blau) ===")
p5 = {
    'section_id': '3',
    'candidates': candidates,
    'team_size': 3,
    'criteria_parameters': {'career_code': {'target': 0.5, 'weight': 1.0, 'type': 'categorical'}},
    'demographic_data': {c: {'categorical_attrs': {'career_code': 'ICOM' if int(c) % 2 == 0 else 'INF'}} for c in candidates}
}
r5 = call_api('form-flexible', p5)
print("Success:", r5.get('success'), "| Equipos:", len(r5.get('data', {}).get('teams', [])), "| Global:", round(r5.get('data', {}).get('global_score', 0), 4))

print("\n=== TEST 6: Híbrido Psicológico + Social (NSGA-II Multi-Objetivo) ===")
p6 = {
    'section_id': '3',
    'candidates': candidates,
    'team_size': 3,
    'criteria_parameters': {
        'conscientiousness': {'target': 0.8, 'weight': 0.5, 'goal': 'heterogeneity'},
        'social': {'target': 0.5, 'weight': 0.5, 'type': 'social'}
    },
    'demographic_data': {c: {
        'numeric_attrs': {'conscientiousness': 0.2 + (int(c) % 5) * 0.18},
        'social_connections': {str(38 + (int(c) + 1) % 12): 1.0, str(38 + (int(c) + 2) % 12): -1.0}
    } for c in candidates}
}
r6 = call_api('form-flexible', p6)
data6 = r6.get('data', {})
meta6 = data6.get('metadata', {})
print("Success:", r6.get('success'), "| Algoritmo:", data6.get('algorithm'), "| Soluciones Pareto:", len(meta6.get('pareto_front', [])), "| Global:", round(data6.get('global_score', 0), 4))

print("\n=== TEST 7: Estrategia de Equidad (Minimax) ===")
p7 = {
    'section_id': '3',
    'candidates': candidates,
    'team_size': 3,
    'equity_strategy': 'minimax',
    'criteria_parameters': {'conscientiousness': {'target': 0.8, 'weight': 1.0, 'goal': 'heterogeneity'}},
    'demographic_data': {c: {'numeric_attrs': {'conscientiousness': 0.2 + (int(c) % 5) * 0.18}} for c in candidates}
}
r7 = call_api('form-flexible', p7)
t_scores = [t.get('affinity_score', 0) for t in r7.get('data', {}).get('teams', [])]
print("Success:", r7.get('success'), "| Global Minimax:", round(r7.get('data', {}).get('global_score', 0), 4), "| Peor equipo:", round(min(t_scores), 4))

print("\n=== TEST 8: evaluate-groups con recalculo minimax ===")
teams_to_eval = [{'group_id': t['id'], 'member_ids': [m['id'] for m in t['members']]} for t in r1.get('data', {}).get('teams', [])]
p8 = {
    'section_id': '3',
    'groups': teams_to_eval,
    'equity_strategy': 'minimax',
    'criteria_parameters': p1['criteria_parameters'],
    'demographic_data': p1['demographic_data']
}
r8 = call_api('evaluate-groups', p8)
print("Success:", r8.get('success'), "| Grupos evaluados:", len(r8.get('data', {}).get('groups', [])), "| Global minimax:", round(r8.get('data', {}).get('global_score', 0), 4))
