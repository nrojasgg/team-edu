<template>
    <div v-if="show" class="modal-overlay" @click="closeModal">
        <div class="modal-content" @click.stop>
            <div class="modal-header">
                <h3>Formar Grupos Personalizados</h3>
                <button @click="closeModal" class="close-btn">
                    <svg
                        width="20"
                        height="20"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                    >
                        <line x1="18" y1="6" x2="6" y2="18" />
                        <line x1="6" y1="6" x2="18" y2="18" />
                    </svg>
                </button>
            </div>

            <div class="modal-body">
                <p class="subtitle">
                    Selecciona los atributos que deseas optimizar para la
                    formación de grupos
                </p>

                <!-- Tamaño de Grupo -->
                <div class="group-size-section">
                    <div class="size-icon">
                        <svg
                            width="24"
                            height="24"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                        >
                            <path
                                d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"
                            />
                            <circle cx="9" cy="7" r="4" />
                            <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
                            <path d="M16 3.13a4 4 0 0 1 0 7.75" />
                        </svg>
                    </div>
                    <div class="size-content">
                        <h4>Tamaño de Grupo</h4>
                        <div class="size-input-group">
                            <input
                                type="number"
                                v-model.number="config.teamSize"
                                min="2"
                                max="10"
                                class="size-input"
                            />
                            <span class="size-label">estudiantes</span>
                        </div>
                    </div>
                </div>

                <!-- Forzar Equidad entre Grupos -->
                <div class="equity-toggle-row">
                    <div class="equity-toggle-info">
                        <span class="equity-toggle-label">
                            Forzar equidad entre grupos
                        </span>
                        <div class="equity-tooltip-wrapper">
                            <svg
                                class="equity-help-icon"
                                width="15"
                                height="15"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                            >
                                <circle cx="12" cy="12" r="10" />
                                <path
                                    d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"
                                />
                                <line x1="12" y1="17" x2="12.01" y2="17" />
                            </svg>
                            <div class="equity-tooltip">
                                Penaliza la desigualdad para evitar que se
                                generen grupos con puntuaciones muy bajas,
                                aunque esto sacrifique a los grupos perfectos.
                            </div>
                        </div>
                    </div>

                    <button
                        type="button"
                        class="equity-toggle-btn"
                        :class="{ active: config.enforceEquity }"
                        @click="config.enforceEquity = !config.enforceEquity"
                        :aria-pressed="config.enforceEquity"
                        aria-label="Forzar equidad entre grupos"
                    >
                        <span class="equity-toggle-track">
                            <span class="equity-toggle-thumb" />
                        </span>
                    </button>
                </div>

                <!-- Atributos a Optimizar -->
                <div class="attributes-section">
                    <div class="section-header">
                        <h4>Atributos a Optimizar</h4>
                        <span
                            class="selection-hint"
                            :class="{
                                'text-error': selectedCriteriaCount === 0,
                            }"
                        >
                            <svg
                                width="16"
                                height="16"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                            >
                                <circle cx="12" cy="12" r="10" />
                                <line x1="12" y1="16" x2="12" y2="12" />
                                <line x1="12" y1="8" x2="12.01" y2="8" />
                            </svg>
                            Selecciona al menos uno
                        </span>
                    </div>

                    <div class="criteria-grid">
                        <CriterionCard
                            v-model="criteria.career.enabled"
                            v-model:target="criteria.career.target"
                            v-model:weight="criteria.career.weight"
                            :criterion="{
                                id: 'career_code',
                                label: 'Código de Carrera',
                                description:
                                    'Agrupa estudiantes por código de carrera para favorecer interdisciplinariedad o cohesión',
                                type: 'categorical',
                            }"
                            :eligible-count="eligibleForCareer"
                            :total-count="students.length"
                        />

                        <!-- Campos de encuestas demográficas con cobertura suficiente -->
                        <template v-if="availableDemographicFields.length > 0">
                            <CriterionCard
                                v-for="field in availableDemographicFields"
                                :key="field.id"
                                v-model="demographicCriteria[field.id].enabled"
                                v-model:target="
                                    demographicCriteria[field.id].target
                                "
                                v-model:weight="
                                    demographicCriteria[field.id].weight
                                "
                                :criterion="{
                                    id: field.id,
                                    label: field.label,
                                    description: field.description,
                                    type: field.type,
                                }"
                                :eligible-count="
                                    demographicCoverageByTestId[
                                        field.questionnaireId
                                    ] ?? 0
                                "
                                :total-count="students.length"
                            />
                        </template>

                        <!-- Afinidad Social — visible cuando hay cobertura suficiente de sociograma -->
                        <template v-if="hasEnoughSociogramData">
                            <CriterionCard
                                v-model="criteria.social_affinity.enabled"
                                v-model:target="criteria.social_affinity.target"
                                v-model:weight="criteria.social_affinity.weight"
                                :criterion="{
                                    id: 'social_affinity',
                                    label: 'Afinidad Social',
                                    description:
                                        'Compatibilidad social y relaciones entre estudiantes según el sociograma',
                                    type: 'social',
                                }"
                                :eligible-count="sociogramCount"
                                :total-count="students.length"
                            />
                        </template>

                        <!-- Atributos OCEAN — visibles cuando hay cobertura suficiente de test de personalidad -->
                        <template v-if="hasEnoughOceanData">
                            <CriterionCard
                                v-model="criteria.conscientiousness.enabled"
                                v-model:target="
                                    criteria.conscientiousness.target
                                "
                                v-model:weight="
                                    criteria.conscientiousness.weight
                                "
                                :criterion="{
                                    id: 'conscientiousness',
                                    label: 'Responsabilidad',
                                    description:
                                        'Qué tan organizado y confiable es el estudiante',
                                    type: 'numeric',
                                }"
                                :eligible-count="eligibleForOcean"
                                :total-count="students.length"
                            />
                            <CriterionCard
                                v-model="criteria.extraversion.enabled"
                                v-model:target="criteria.extraversion.target"
                                v-model:weight="criteria.extraversion.weight"
                                :criterion="{
                                    id: 'extraversion',
                                    label: 'Sociabilidad',
                                    description:
                                        'Qué tan extrovertido o introvertido es el estudiante',
                                    type: 'numeric',
                                }"
                                :eligible-count="eligibleForOcean"
                                :total-count="students.length"
                            />
                            <CriterionCard
                                v-model="criteria.agreeableness.enabled"
                                v-model:target="criteria.agreeableness.target"
                                v-model:weight="criteria.agreeableness.weight"
                                :criterion="{
                                    id: 'agreeableness',
                                    label: 'Amabilidad',
                                    description:
                                        'Qué tan cooperativo y amable es el estudiante',
                                    type: 'numeric',
                                }"
                                :eligible-count="eligibleForOcean"
                                :total-count="students.length"
                            />
                            <CriterionCard
                                v-model="criteria.openness.enabled"
                                v-model:target="criteria.openness.target"
                                v-model:weight="criteria.openness.weight"
                                :criterion="{
                                    id: 'openness',
                                    label: 'Apertura',
                                    description:
                                        'Qué tan creativo e innovador es el estudiante',
                                    type: 'numeric',
                                }"
                                :eligible-count="eligibleForOcean"
                                :total-count="students.length"
                            />
                            <CriterionCard
                                v-model="criteria.neuroticism.enabled"
                                v-model:target="criteria.neuroticism.target"
                                v-model:weight="criteria.neuroticism.weight"
                                :criterion="{
                                    id: 'neuroticism',
                                    label: 'Estabilidad Emocional',
                                    description:
                                        'Qué tan emocionalmente estable es el estudiante',
                                    type: 'numeric',
                                }"
                                :eligible-count="eligibleForOcean"
                                :total-count="students.length"
                            />
                            <CriterionCard
                                v-model="criteria.openness_music.enabled"
                                v-model:target="
                                    criteria.openness_music.target
                                "
                                v-model:weight="
                                    criteria.openness_music.weight
                                "
                                :criterion="{
                                    id: 'openness_music',
                                    label: 'Apertura Musical',
                                    description:
                                        'Apreciación por la música y las artes',
                                    type: 'numeric',
                                }"
                                :eligible-count="eligibleForOcean"
                                :total-count="students.length"
                            />
                            <CriterionCard
                                v-model="criteria.emotionality.enabled"
                                v-model:target="
                                    criteria.emotionality.target
                                "
                                v-model:weight="
                                    criteria.emotionality.weight
                                "
                                :criterion="{
                                    id: 'emotionality',
                                    label: 'Emocionalidad',
                                    description:
                                        'Qué tan consciente es de sus emociones',
                                    type: 'numeric',
                                }"
                                :eligible-count="eligibleForOcean"
                                :total-count="students.length"
                            />
                        </template>
                    </div>
                </div>

                <!-- Atributos sin datos disponibles -->
                <div
                    class="unavailable-section"
                    :class="{ expanded: unavailableExpanded }"
                >
                    <button
                        type="button"
                        class="unavailable-header"
                        @click="unavailableExpanded = !unavailableExpanded"
                    >
                        <svg
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                        >
                            <circle cx="12" cy="12" r="10" />
                            <line x1="12" y1="8" x2="12" y2="12" />
                            <line x1="12" y1="16" x2="12.01" y2="16" />
                        </svg>
                        <span
                            >{{ unavailableAttributes.length }} atributo(s) sin
                            datos disponibles</span
                        >
                        <svg
                            class="chevron"
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                        >
                            <polyline points="6 9 12 15 18 9" />
                        </svg>
                    </button>

                    <div v-if="unavailableExpanded" class="unavailable-content">
                        <p class="unavailable-description">
                            Los siguientes atributos no tienen datos completos
                            de la mayoría de estudiantes. Para activarlos,
                            necesitarían completar cuestionarios adicionales.
                        </p>

                        <div class="unavailable-list">
                            <div
                                v-for="attr in unavailableAttributes"
                                :key="attr.id"
                                class="unavailable-item"
                            >
                                <div class="unavailable-icon">
                                    <svg
                                        width="16"
                                        height="16"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                    >
                                        <circle cx="12" cy="12" r="10" />
                                        <line x1="15" y1="9" x2="9" y2="15" />
                                        <line x1="9" y1="9" x2="15" y2="15" />
                                    </svg>
                                </div>
                                <div class="unavailable-info">
                                    <strong>{{ attr.label }}</strong>
                                    <p>{{ attr.description }}</p>
                                    <p class="data-status">
                                        {{ attr.availableCount }} de
                                        {{ students.length }} estudiantes tienen
                                        datos
                                    </p>
                                    <p class="requirement-note">
                                        {{ attr.requirement }}
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Validación final y resumen de elegibilidad -->
                <div
                    v-if="selectedCriteriaCount === 0"
                    class="validation-message"
                >
                    <svg
                        width="20"
                        height="20"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                    >
                        <circle cx="12" cy="12" r="10" />
                        <line x1="12" y1="8" x2="12" y2="12" />
                        <line x1="12" y1="16" x2="12.01" y2="16" />
                    </svg>
                    <p>Selecciona al menos un atributo para optimizar</p>
                </div>

                <div
                    v-else
                    class="eligibility-banner"
                    :class="{ 'has-exclusions': excludedStudentsCount > 0 }"
                >
                    <div class="eligibility-content">
                        <span class="eligibility-title">
                            Estudiantes participantes: <strong>{{ eligibleStudents.length }}</strong> de {{ props.students.length }}
                        </span>
                        <span v-if="excludedStudentsCount > 0" class="eligibility-excluded">
                            ({{ excludedStudentsCount }} excluidos por cuestionarios pendientes para los criterios seleccionados)
                        </span>
                    </div>
                    <p v-if="eligibleStudents.length < config.teamSize" class="eligibility-warning">
                        Se requieren al menos {{ config.teamSize }} estudiantes con datos completos para formar grupos.
                    </p>
                </div>
            </div>

            <div class="modal-footer">
                <button @click="closeModal" class="btn btn-secondary">
                    Cancelar
                </button>
                <button
                    @click="confirm"
                    class="btn btn-primary"
                    :disabled="!canFormGroups"
                >
                    Formar Grupos
                </button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, watch } from "vue";
import CriterionCard from "./CriterionCard.vue";

const props = defineProps({
    show: {
        type: Boolean,
        default: false,
    },
    students: {
        type: Array,
        default: () => [],
    },
    studentTestStatus: {
        type: Object,
        default: () => ({}),
    },
    questionnaires: {
        type: Array,
        default: () => [],
    },
});

const emit = defineEmits(["confirm", "close"]);

const config = ref({
    teamSize: 4,
    enforceEquity: false,
});

const criteria = ref({
    career: {
        id: "career_code",
        label: "Código de Carrera",
        enabled: false,
        target: 0.5,
        weight: 0.5,
        type: "categorical",
    },
    conscientiousness: {
        id: "conscientiousness",
        label: "Responsabilidad",
        enabled: false,
        target: 0.5,
        weight: 0.5,
        type: "numeric",
    },
    extraversion: {
        id: "extraversion",
        label: "Sociabilidad",
        enabled: false,
        target: 0.5,
        weight: 0.5,
        type: "numeric",
    },
    agreeableness: {
        id: "agreeableness",
        label: "Amabilidad",
        enabled: false,
        target: 0.5,
        weight: 0.5,
        type: "numeric",
    },
    openness: {
        id: "openness",
        label: "Apertura",
        enabled: false,
        target: 0.5,
        weight: 0.5,
        type: "numeric",
    },
    neuroticism: {
        id: "neuroticism",
        label: "Estabilidad Emocional",
        enabled: false,
        target: 0.5,
        weight: 0.5,
        type: "numeric",
    },
    openness_music: {
        id: "openness_music",
        label: "Apertura Musical",
        enabled: false,
        target: 0.5,
        weight: 0.5,
        type: "numeric",
    },
    emotionality: {
        id: "emotionality",
        label: "Emocionalidad",
        enabled: false,
        target: 0.5,
        weight: 0.5,
        type: "numeric",
    },
    social_affinity: {
        id: "social_affinity",
        label: "Afinidad Social",
        enabled: false,
        target: 0.0,
        weight: 0.5,
        type: "social",
    },
});

const unavailableExpanded = ref(false);

// Cobertura de datos OCEAN (eneagrama u otro test de personalidad)
// Count students with ANY personality test data (enneagram, BFI-44, or MBTI)
const personalityCount = computed(
    () =>
        props.students.filter((s) => {
            const status = props.studentTestStatus[s.id];
            return (
                status?.hasEnneagram ||
                status?.hasBFI44 ||
                status?.hasMBTI ||
                (s.eneatype && s.eneatype > 0)
            );
        }).length,
);

const enneagramCount = computed(
    () =>
        props.students.filter((s) => {
            const status = props.studentTestStatus[s.id];
            return status?.hasEnneagram || (s.eneatype && s.eneatype > 0);
        }).length,
);

const hasEnoughOceanData = computed(
    () =>
        props.students.length > 0 &&
        personalityCount.value >= props.students.length * 0.5,
);

const sociogramCount = computed(
    () =>
        props.students.filter((s) => {
            const status = props.studentTestStatus[s.id];
            return status?.hasSociogram || s.has_sociogram_response === true;
        }).length,
);

const hasEnoughSociogramData = computed(
    () =>
        props.students.length > 0 &&
        sociogramCount.value >= props.students.length * 0.5,
);

// Estudiantes elegibles para criterios OCEAN (any personality test counts)
const eligibleForOcean = computed(() => personalityCount.value);

// Umbral mínimo de cobertura para que un campo demográfico sea seleccionable
const DEMOGRAPHIC_MIN_COVERAGE = 4;

// Atributos no disponibles (requieren tests adicionales)
const unavailableAttributes = computed(() => {
    const attrs = [];

    const hasEnneagram = enneagramCount.value;

    const hasSociogram = sociogramCount.value;

    // Campos demográficos con cobertura insuficiente (< DEMOGRAPHIC_MIN_COVERAGE).
    // Se calcula inline para evitar dependencia circular con demographicCoverageByTestId
    // que está declarado más adelante.
    for (const q of props.questionnaires) {
        if (
            q.plugin !== "demographic" ||
            !Array.isArray(q.fields) ||
            q.fields.length === 0
        )
            continue;
        const rawId = q.id || q._id || "q";
        const qId =
            rawId && typeof rawId === "object" && rawId.$oid
                ? rawId.$oid
                : String(rawId);
        // Calcular cobertura para este cuestionario
        let count = 0;
        for (const student of props.students) {
            const sid = student.id != null ? String(student.id) : null;
            if (!sid) continue;
            const status =
                props.studentTestStatus[sid] ||
                props.studentTestStatus[Number(sid)];
            if (status?.demographicByTestId?.[qId]) count++;
        }
        if (count >= DEMOGRAPHIC_MIN_COVERAGE) continue;
        // Agregar cada campo del cuestionario a no disponibles
        q.fields.forEach((f, idx) => {
            const slug = f.title
                ? f.title
                      .toLowerCase()
                      .replace(/[^a-z0-9]+/g, "_")
                      .replace(/^_|_$/g, "")
                : `field_${idx}`;
            attrs.push({
                id: `demographic_${qId}_${slug}`,
                label: f.title || `Campo ${idx + 1}`,
                description: `${q.title || q.name}: ${f.title}`,
                availableCount: count,
                requirement: `(Requiere al menos ${DEMOGRAPHIC_MIN_COVERAGE} respuestas en la encuesta demográfica)`,
            });
        });
    }

    // Solo mostrar atributos OCEAN como «sin datos» cuando la cobertura sea insuficiente
    if (!hasEnoughOceanData.value) {
        attrs.push(
            {
                id: "conscientiousness",
                label: "Responsabilidad",
                description: "Qué tan organizado y confiable es el estudiante",
                availableCount: hasEnneagram,
                requirement:
                    "(Requiere Test de personalidad (Eneagrama, BFI-44 o MBTI))",
            },
            {
                id: "extraversion",
                label: "Sociabilidad",
                description:
                    "Qué tan extrovertido o introvertido es el estudiante",
                availableCount: hasEnneagram,
                requirement:
                    "(Requiere Test de personalidad (Eneagrama, BFI-44 o MBTI))",
            },
            {
                id: "agreeableness",
                label: "Amabilidad",
                description: "Qué tan cooperativo y amable es el estudiante",
                availableCount: hasEnneagram,
                requirement:
                    "(Requiere Test de personalidad (Eneagrama, BFI-44 o MBTI))",
            },
            {
                id: "openness",
                label: "Apertura",
                description: "Qué tan creativo e innovador es el estudiante",
                availableCount: hasEnneagram,
                requirement:
                    "(Requiere Test de personalidad (Eneagrama, BFI-44 o MBTI))",
            },
            {
                id: "neuroticism",
                label: "Estabilidad Emocional",
                description: "Qué tan emocionalmente estable es el estudiante",
                availableCount: hasEnneagram,
                requirement:
                    "(Requiere Test de personalidad (Eneagrama, BFI-44 o MBTI))",
            },
        );
    }

    attrs.push({
        id: "chronotype",
        label: "Cronotipo",
        description: "Preferencia de trabajo en mañana o tarde",
        availableCount: 0,
        requirement: "(Requiere Test de cronotipo)",
    });

    if (!hasEnoughSociogramData.value) {
        attrs.push({
            id: "social_affinity",
            label: "Afinidad Social",
            description: "Compatibilidad social y relaciones entre estudiantes",
            availableCount: hasSociogram,
            requirement: "(Requiere Test de afinidad social)",
        });
    }

    return attrs;
});

// Estudiantes elegibles para cada criterio
const eligibleForCareer = computed(() => {
    return props.students.filter((s) => s.career_code || s.codigo_carrera)
        .length;
});

// Campos demográficos dinámicos desde las encuestas demográficas de la sección.
// Incluye el questionnaireId para poder calcular cobertura real.
const demographicFields = computed(() => {
    const fields = [];
    for (const q of props.questionnaires) {
        if (
            q.plugin === "demographic" &&
            Array.isArray(q.fields) &&
            q.fields.length > 0
        ) {
            // El id puede venir como string plano o como objeto { $oid: "..." }
            const rawId = q.id || q._id || "q";
            const qId =
                rawId && typeof rawId === "object" && rawId.$oid
                    ? rawId.$oid
                    : String(rawId);
            q.fields.forEach((f, idx) => {
                const slug = f.title
                    ? f.title
                          .toLowerCase()
                          .replace(/[^a-z0-9]+/g, "_")
                          .replace(/^_|_$/g, "")
                    : `field_${idx}`;
                fields.push({
                    id: `demographic_${qId}_${slug}`,
                    label: f.title || `Campo ${idx + 1}`,
                    description: `${q.title || q.name}: ${f.title}`,
                    type: f.type === "numeric" ? "numeric" : "categorical",
                    // Guardamos el qId normalizado para cruzar con studentTestStatus
                    questionnaireId: qId,
                });
            });
        }
    }
    return fields;
});

// Cobertura real por test demográfico: cuántos estudiantes de props.students
// tienen una entrada en studentTestStatus[sid].demographicByTestId[testId].
const demographicCoverageByTestId = computed(() => {
    const coverage = {};
    for (const field of demographicFields.value) {
        const testId = field.questionnaireId;
        if (testId in coverage) continue; // ya calculado para este test
        let count = 0;
        for (const student of props.students) {
            const sid = student.id != null ? String(student.id) : null;
            if (!sid) continue;
            const status =
                props.studentTestStatus[sid] ||
                props.studentTestStatus[Number(sid)];
            if (status?.demographicByTestId?.[testId]) {
                count++;
            }
        }
        coverage[testId] = count;
    }
    return coverage;
});

// Campos demográficos con cobertura suficiente (>= DEMOGRAPHIC_MIN_COVERAGE) — aparecen en la grilla seleccionable
const availableDemographicFields = computed(() =>
    demographicFields.value.filter(
        (f) =>
            (demographicCoverageByTestId.value[f.questionnaireId] ?? 0) >=
            DEMOGRAPHIC_MIN_COVERAGE,
    ),
);

const demographicCriteria = ref({});

watch(
    demographicFields,
    (fields) => {
        const next = {};
        for (const f of fields) {
            next[f.id] = demographicCriteria.value[f.id] || {
                id: f.id,
                label: f.label,
                enabled: false,
                target: 0.5,
                weight: 0.5,
                type: f.type,
                questionnaireId: f.questionnaireId,
            };
        }
        demographicCriteria.value = next;
    },
    { immediate: true },
);

const selectedCriteriaCount = computed(() => {
    return (
        Object.values(criteria.value).filter((c) => c.enabled).length +
        Object.values(demographicCriteria.value).filter((c) => c.enabled).length
    );
});

// Estudiantes elegibles: solo aquellos que han respondido TODOS los cuestionarios seleccionados.
// Aquellos que no han respondido son excluidos de la formación de grupos.
const eligibleStudents = computed(() => {
    const enabled = [
        ...Object.values(criteria.value).filter((c) => c.enabled),
        ...Object.values(demographicCriteria.value).filter((c) => c.enabled),
    ];
    if (enabled.length === 0) return props.students;

    return props.students.filter((student) => {
        const sid = student.id != null ? String(student.id) : null;
        if (!sid) return false;
        const status =
            props.studentTestStatus?.[sid] ||
            props.studentTestStatus?.[Number(sid)] ||
            {};

        for (const crit of enabled) {
            // 1. Criterio social (sociograma)
            if (crit.type === "social" || crit.id === "social_affinity") {
                const hasSoc =
                    status.hasSociogram ||
                    student.has_sociogram_response === true;
                if (!hasSoc) return false;
            }
            // 2. Criterios de personalidad (OCEAN / Big Five / Eneagrama / MBTI / BFI-44)
            else if (
                crit.id === "conscientiousness" ||
                crit.id === "extraversion" ||
                crit.id === "agreeableness" ||
                crit.id === "openness" ||
                crit.id === "neuroticism" ||
                crit.id === "openness_music" ||
                crit.id === "emotionality"
            ) {
                const hasPersonality =
                    status.hasEnneagram ||
                    status.hasBFI44 ||
                    status.hasMBTI ||
                    (student.eneatype && student.eneatype > 0) ||
                    (student.ocean_vector &&
                        Object.keys(student.ocean_vector).length > 0);
                if (!hasPersonality) return false;
            }
            // 3. Criterio de carrera universitaria
            else if (crit.id === "career") {
                const hasCareer = Boolean(
                    student.career_code ||
                        student.codigo_carrera ||
                        student.carrera
                );
                if (!hasCareer) return false;
            }
            // 4. Criterios demográficos dinámicos
            else if (crit.id.startsWith("demographic_")) {
                const qId = crit.questionnaireId;
                const hasDemo = qId && status.demographicByTestId?.[qId];
                if (!hasDemo) return false;
            }
        }
        return true;
    });
});

const excludedStudentsCount = computed(() => {
    return Math.max(0, props.students.length - eligibleStudents.value.length);
});

const canFormGroups = computed(() => {
    return (
        selectedCriteriaCount.value > 0 &&
        eligibleStudents.value.length >= config.value.teamSize
    );
});

const closeModal = () => {
    emit("close");
};

const normalizeTargetForCriterion = (criterion) => {
    if (criterion?.type === "social") {
        return criterion.target * 2 - 1;
    }
    return criterion?.target ?? 0.5;
};

const confirm = () => {
    if (!canFormGroups.value) return;

    const enabledCriteria = [
        ...Object.values(criteria.value).filter((c) => c.enabled),
        ...Object.values(demographicCriteria.value).filter((c) => c.enabled),
    ];

    const formConfig = {
        teamSize: config.value.teamSize,
        enforceEquity: config.value.enforceEquity,
        eligibleStudentIds: eligibleStudents.value.map((s) => String(s.id)),
        criteria: enabledCriteria.map((c) => ({
            id: c.id,
            target: normalizeTargetForCriterion(c),
            weight: c.weight,
            type: c.type,
            label: c.label,
            goal:
                c.target <= 0.3
                    ? "homogeneity"
                    : c.target >= 0.7
                    ? "heterogeneity"
                    : "target",
        })),
    };

    emit("confirm", formConfig);
    closeModal();
};

// Reset cuando el modal se abre
watch(
    () => props.show,
    (newVal) => {
        if (newVal) {
            config.value.teamSize = 4;
            config.value.enforceEquity = false;
            criteria.value.career.enabled = false;
            criteria.value.career.target = 0.5;
            criteria.value.career.weight = 0.5;
            for (const key of [
                "conscientiousness",
                "extraversion",
                "agreeableness",
                "openness",
                "neuroticism",
                "openness_music",
                "emotionality",
                "social_affinity",
            ]) {
                if (criteria.value[key]) {
                    criteria.value[key].enabled = false;
                    criteria.value[key].target = 0.5;
                    criteria.value[key].weight = 0.5;
                }
            }
            criteria.value.social_affinity.target = 0.0;
            criteria.value.social_affinity.type = "social";
            for (const key of Object.keys(demographicCriteria.value)) {
                demographicCriteria.value[key].enabled = false;
                demographicCriteria.value[key].target = 0.5;
                demographicCriteria.value[key].weight = 0.5;
            }
            unavailableExpanded.value = false;
        }
    },
);
</script>

<style scoped>
/* ── Equity toggle row ───────────────────────────────────────────────────── */
.equity-toggle-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 10px 14px;
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-radius: 10px;
    margin-bottom: 16px;
}

.equity-toggle-info {
    display: flex;
    align-items: center;
    gap: 6px;
    flex: 1;
    min-width: 0;
}

.equity-toggle-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: #334155;
    white-space: nowrap;
}

/* Tooltip */
.equity-tooltip-wrapper {
    position: relative;
    display: flex;
    align-items: center;
}

.equity-help-icon {
    color: #94a3b8;
    cursor: help;
    flex-shrink: 0;
}

.equity-tooltip {
    display: none;
    position: absolute;
    bottom: calc(100% + 8px);
    left: 50%;
    transform: translateX(-50%);
    background: #1e293b;
    color: #f1f5f9;
    font-size: 0.78rem;
    line-height: 1.5;
    padding: 8px 12px;
    border-radius: 8px;
    width: 240px;
    text-align: center;
    z-index: 100;
    pointer-events: none;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.25);
}

.equity-tooltip::after {
    content: "";
    position: absolute;
    top: 100%;
    left: 50%;
    transform: translateX(-50%);
    border: 5px solid transparent;
    border-top-color: #1e293b;
}

.equity-tooltip-wrapper:hover .equity-tooltip {
    display: block;
}

/* Toggle button */
.equity-toggle-btn {
    flex-shrink: 0;
    background: none;
    border: none;
    padding: 0;
    cursor: pointer;
    outline: none;
    border-radius: 999px;
}

.equity-toggle-btn:focus-visible .equity-toggle-track {
    box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.18);
}

.equity-toggle-track {
    display: flex;
    align-items: center;
    width: 44px;
    height: 24px;
    border-radius: 999px;
    background: #cbd5e1;
    padding: 2px;
    transition: background 0.2s ease;
}

.equity-toggle-btn.active .equity-toggle-track {
    background: var(--color-primary);
}

.equity-toggle-thumb {
    display: block;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: #ffffff;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
    transform: translateX(0);
    transition: transform 0.2s ease;
}

.equity-toggle-btn.active .equity-toggle-thumb {
    transform: translateX(20px);
}

/* ── End equity toggle ───────────────────────────────────────────────────── */

.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
    padding: 20px;
}

.modal-content {
    background: white;
    border-radius: 12px;
    max-width: 1000px;
    width: 100%;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 24px 32px;
    border-bottom: 1px solid #e5e7eb;
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-dark) 100%);
    border-radius: 12px 12px 0 0;
}

.modal-header h3 {
    margin: 0;
    font-size: 24px;
    font-weight: 600;
    color: white;
}

.close-btn {
    background: transparent;
    border: none;
    border-radius: 6px;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s;
    color: white;
}

.close-btn:hover {
    background: rgba(255, 255, 255, 0.2);
}

.modal-body {
    padding: 32px;
    overflow-y: auto;
    flex: 1;
}

.subtitle {
    color: #6b7280;
    font-size: 15px;
    margin: 0 0 28px 0;
}

/* Sección de Tamaño de Grupo */
.group-size-section {
    display: flex;
    gap: 16px;
    padding: 20px;
    background: linear-gradient(135deg, var(--color-primary-light) 0%, #f7fffe 100%);
    border-radius: 12px;
    border-left: 4px solid var(--color-primary);
    margin-bottom: 32px;
}

.size-icon {
    display: flex;
    align-items: flex-start;
    color: var(--color-primary);
}

.size-content {
    flex: 1;
}

.size-content h4 {
    margin: 0 0 12px 0;
    font-size: 16px;
    font-weight: 600;
    color: #1f2937;
    display: flex;
    align-items: center;
    gap: 8px;
}

.size-input-group {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 8px;
}

.size-input {
    width: 80px;
    padding: 8px 12px;
    border: 2px solid #d1d5db;
    border-radius: 8px;
    font-size: 16px;
    font-weight: 600;
    text-align: center;
    transition: all 0.2s;
}

.size-input:focus {
    outline: none;
    border-color: var(--color-primary-dark);
    box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.08);
}

.size-label {
    color: #4b5563;
    font-size: 15px;
}

.help-text {
    color: #6b7280;
    font-size: 13px;
    margin: 0;
}

/* Atributos a Optimizar */
.attributes-section {
    margin-bottom: 32px;
}

.section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
}

.section-header h4 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: #1f2937;
}

.selection-hint {
    display: flex;
    align-items: center;
    gap: 6px;
    color: #6b7280;
    font-size: 14px;
}

.selection-hint svg {
    color: var(--color-primary);
}

.selection-hint.text-error {
    color: #ef4444;
}

.selection-hint.text-error svg {
    color: #ef4444;
}

.criteria-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(450px, 1fr));
    gap: 20px;
}

/* Atributos No Disponibles */
.unavailable-section {
    margin-bottom: 24px;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    overflow: hidden;
}

.unavailable-header {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px 20px;
    background: #fef3f2;
    border: none;
    cursor: pointer;
    transition: all 0.2s;
    color: #991b1b;
    font-size: 14px;
    font-weight: 500;
}

.unavailable-header:hover {
    background: #fee2e2;
}

.unavailable-header svg:first-child {
    color: #dc2626;
    flex-shrink: 0;
}

.unavailable-header .chevron {
    margin-left: auto;
    transition: transform 0.2s;
    color: #991b1b;
}

.unavailable-section.expanded .chevron {
    transform: rotate(180deg);
}

.unavailable-content {
    padding: 20px;
    background: white;
    border-top: 1px solid #fee2e2;
}

.unavailable-description {
    color: #6b7280;
    font-size: 14px;
    margin: 0 0 16px 0;
    line-height: 1.5;
}

.unavailable-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.unavailable-item {
    display: flex;
    gap: 12px;
    padding: 12px;
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
}

.unavailable-icon {
    color: #dc2626;
    flex-shrink: 0;
}

.unavailable-info {
    flex: 1;
}

.unavailable-info strong {
    display: block;
    color: #1f2937;
    font-size: 14px;
    margin-bottom: 4px;
}

.unavailable-info p {
    color: #6b7280;
    font-size: 13px;
    margin: 4px 0;
    line-height: 1.4;
}

.data-status {
    color: #9ca3af !important;
    font-size: 12px !important;
}

.requirement-note {
    color: #dc2626 !important;
    font-style: italic;
    font-size: 12px !important;
}

/* Validación */
.validation-message {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px;
    background: var(--color-primary-light);
    border: 1px solid var(--color-primary);
    border-radius: 8px;
    color: var(--color-primary-dark);
    font-size: 14px;
}

.validation-message svg {
    color: var(--color-primary);
    flex-shrink: 0;
}

.validation-message p {
    margin: 0;
}

.eligibility-banner {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 14px 18px;
    background: #f0fdf4;
    border: 1px solid #86efac;
    border-radius: 8px;
    color: #166534;
    font-size: 14px;
    margin-top: 16px;
}

.eligibility-banner.has-exclusions {
    background: #fffbeb;
    border-color: #fde68a;
    color: #92400e;
}

.eligibility-content {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
}

.eligibility-excluded {
    font-size: 13px;
    color: #b45309;
}

.eligibility-warning {
    margin: 4px 0 0 0;
    font-size: 13px;
    font-weight: 500;
    color: #dc2626;
}

/* Footer */
.modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding: 20px 32px;
    border-top: 1px solid #e5e7eb;
    background: #f9fafb;
    border-radius: 0 0 12px 12px;
}

.btn {
    padding: 10px 20px;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    border: none;
}

.btn-secondary {
    background: white;
    color: #6b7280;
    border: 1px solid #d1d5db;
}

.btn-secondary:hover {
    background: #f9fafb;
    border-color: #9ca3af;
}

.btn-primary {
    background: linear-gradient(135deg, var(--color-primary-dark) 0%, var(--color-primary) 100%);
    color: white;
}

.btn-primary:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(0, 164, 153, 0.25);
}

.btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

@media (max-width: 768px) {
    .modal-content {
        max-width: 100%;
        max-height: 100vh;
        border-radius: 0;
    }

    .modal-header {
        border-radius: 0;
    }

    .criteria-grid {
        grid-template-columns: 1fr;
    }

    .config-row {
        flex-direction: column;
        align-items: stretch;
    }

    .config-label {
        min-width: auto;
    }
}
</style>
