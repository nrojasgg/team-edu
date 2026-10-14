<template>
    <div v-if="isOpen" class="modal-overlay" @click.self="showSociogramGraph ? closeSociogramGraph() : closeModal()">
        <div class="modal-content">
            <div class="modal-header">
                <div>
                    <h2>{{ questionnaire?.title || "Cuestionario" }}</h2>
                    <p class="modal-subtitle">Respuestas de estudiantes</p>
                </div>
                <button @click="showSociogramGraph ? closeSociogramGraph() : closeModal()" class="close-btn">
                    <svg
                        width="24"
                        height="24"
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

            <div v-if="!showSociogramGraph" class="modal-body">
                <div v-if="loading" class="loading-state">
                    <div class="spinner"></div>
                    <p>Cargando respuestas...</p>
                </div>

                <div v-else-if="error" class="error-state">
                    <svg
                        width="48"
                        height="48"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                    >
                        <circle cx="12" cy="12" r="10" />
                        <line x1="12" y1="8" x2="12" y2="12" />
                        <line x1="12" y1="16" x2="12.01" y2="16" />
                    </svg>
                    <p>{{ error }}</p>
                </div>

                <div v-else-if="!showSociogramGraph" class="students-list">
                    <!-- Summary Stats -->
                    <div class="stats-section">
                        <div class="stat-card">
                            <div class="stat-value">
                                {{ studentsData.length }}
                            </div>
                            <div class="stat-label">Total de estudiantes</div>
                        </div>
                        <div class="stat-card completed">
                            <div class="stat-value">{{ completedCount }}</div>
                            <div class="stat-label">Completados</div>
                        </div>
                        <div class="stat-card pending">
                            <div class="stat-value">{{ pendingCount }}</div>
                            <div class="stat-label">Pendientes</div>
                        </div>
                        <div class="stat-card percentage">
                            <div class="stat-value">
                                {{ completionPercentage }}%
                            </div>
                            <div class="stat-label">Completitud</div>
                        </div>
                    </div>

                    <!-- Filter Options -->
                    <div class="filter-section">
                        <label class="filter-label">
                            <input type="radio" v-model="filter" value="all" />
                            <span>Todos ({{ studentsData.length }})</span>
                        </label>
                        <label class="filter-label">
                            <input
                                type="radio"
                                v-model="filter"
                                value="completed"
                            />
                            <span>Completados ({{ completedCount }})</span>
                        </label>
                        <label class="filter-label">
                            <input
                                type="radio"
                                v-model="filter"
                                value="pending"
                            />
                            <span>Pendientes ({{ pendingCount }})</span>
                        </label>
                    </div>

                    <!-- Students Table -->
                    <div class="table-container">
                        <table class="students-table">
                            <thead>
                                <tr>
                                    <th>Estudiante</th>
                                    <th>Email</th>
                                    <th>Estado</th>
                                    <th>Fecha de respuesta</th>
                                    <th>Resultado</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr
                                    v-for="student in filteredStudents"
                                    :key="student.id"
                                    :class="{
                                        'completed-row': student.has_answered,
                                    }"
                                >
                                    <td class="student-name">
                                        {{ student.name }}
                                    </td>
                                    <td class="student-email">
                                        {{ student.email }}
                                    </td>
                                    <td class="status-cell">
                                        <span
                                            v-if="student.has_answered"
                                            class="status-badge completed"
                                        >
                                            <svg
                                                width="16"
                                                height="16"
                                                viewBox="0 0 24 24"
                                                fill="none"
                                                stroke="currentColor"
                                                stroke-width="2"
                                            >
                                                <polyline
                                                    points="20 6 9 17 4 12"
                                                />
                                            </svg>
                                            Completado
                                        </span>
                                        <span
                                            v-else
                                            class="status-badge pending"
                                        >
                                            <svg
                                                width="16"
                                                height="16"
                                                viewBox="0 0 24 24"
                                                fill="none"
                                                stroke="currentColor"
                                                stroke-width="2"
                                            >
                                                <circle
                                                    cx="12"
                                                    cy="12"
                                                    r="10"
                                                />
                                                <line
                                                    x1="12"
                                                    y1="8"
                                                    x2="12"
                                                    y2="12"
                                                />
                                            </svg>
                                            Pendiente
                                        </span>
                                    </td>
                                    <td class="date-cell">
                                        {{ formatDate(student.submitted_at) }}
                                    </td>
                                    <td class="result-cell">
                                        <!-- Resultado para tests psicométricos (Enneagram, BFI-44, MBTI) -->
                                        <div
                                            v-if="
                                                (student.result || student.ocean_vector) && !isSociogram
                                            "
                                            class="result-details"
                                        >
                                            <span v-if="student.result" class="result-badge"
                                                >Eneatipo
                                                {{ student.result }}</span
                                            >
                                            <div
                                                v-if="student.ocean_vector"
                                                class="ocean-mini-grid"
                                                title="Vector OCEAN (O, C, E, A, N)"
                                            >
                                                <span class="ocean-mini-val o"
                                                    >O:
                                                    {{
                                                        (
                                                            student.ocean_vector
                                                                .openness * 10
                                                        ).toFixed(1)
                                                    }}</span
                                                >
                                                <span class="ocean-mini-val c"
                                                    >C:
                                                    {{
                                                        (
                                                            student.ocean_vector
                                                                .conscientiousness *
                                                            10
                                                        ).toFixed(1)
                                                    }}</span
                                                >
                                                <span class="ocean-mini-val e"
                                                    >E:
                                                    {{
                                                        (
                                                            student.ocean_vector
                                                                .extraversion *
                                                            10
                                                        ).toFixed(1)
                                                    }}</span
                                                >
                                                <span class="ocean-mini-val a"
                                                    >A:
                                                    {{
                                                        (
                                                            student.ocean_vector
                                                                .agreeableness *
                                                            10
                                                        ).toFixed(1)
                                                    }}</span
                                                >
                                                <span class="ocean-mini-val n"
                                                    >N:
                                                    {{
                                                        (
                                                            student.ocean_vector
                                                                .neuroticism *
                                                            10
                                                        ).toFixed(1)
                                                    }}</span
                                                >
                                            </div>
                                        </div>
                                        <!-- Resultado para sociograma: botón para ver respuesta -->
                                        <button
                                            v-else-if="
                                                isSociogram &&
                                                student.has_answered
                                            "
                                            class="view-response-btn"
                                            @click="
                                                openSociogramResponse(student)
                                            "
                                        >
                                            <svg
                                                width="14"
                                                height="14"
                                                viewBox="0 0 24 24"
                                                fill="none"
                                                stroke="currentColor"
                                                stroke-width="2"
                                            >
                                                <path
                                                    d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"
                                                />
                                                <circle cx="12" cy="12" r="3" />
                                            </svg>
                                            Ver respuesta
                                        </button>
                                        <span v-else class="no-result">-</span>
                                    </td>
                                </tr>
                            </tbody>
                        </table>

                        <div
                            v-if="filteredStudents.length === 0"
                            class="no-data"
                        >
                            No hay estudiantes
                            {{
                                filter === "completed"
                                    ? "que hayan completado"
                                    : filter === "pending"
                                      ? "pendientes"
                                      : ""
                            }}
                            el cuestionario
                        </div>
                    </div>
                </div>
            </div>

            <div v-if="!showSociogramGraph" class="modal-footer">
                <button @click="exportToCSV" class="btn btn-export">
                    <svg
                        width="16"
                        height="16"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                    >
                        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                        <polyline points="7 10 12 15 17 10" />
                        <line x1="12" y1="15" x2="12" y2="3" />
                    </svg>
                    Exportar CSV
                </button>
                <button @click="closeModal" class="btn btn-close">
                    Cerrar
                </button>
            </div>

            <!-- Sociogram graph view (replaces table when a student is selected) -->
            <div v-if="showSociogramGraph" class="sociogram-graph-view">
                <div class="sociogram-graph-toolbar">
                    <div class="sociogram-toolbar-title">
                        <span class="sociogram-toolbar-label">Socio</span>
                        <span class="sociogram-toolbar-name">{{ sociogramDetailStudent?.name }}</span>
                    </div>
                    <div class="sociogram-graph-filters">
                        <label class="sociogram-filter">
                            <input type="radio" v-model="sociogramLayer" value="all" />
                            Todas
                        </label>
                        <label class="sociogram-filter">
                            <input type="radio" v-model="sociogramLayer" value="real" />
                            Real (Q1/Q3)
                        </label>
                        <label class="sociogram-filter">
                            <input type="radio" v-model="sociogramLayer" value="perceived" />
                            Percibido (Q2/Q4)
                        </label>
                    </div>
                    <div class="toolbar-actions">
                        <button class="btn-export-graph" @click="exportSociogramSvg">
                            Exportar SVG
                        </button>
                        <button class="btn-export-graph" @click="exportSociogramPng">
                            Exportar PNG
                        </button>
                        <button class="btn-export-graph" @click="resetGraphViewport">
                            Reset vista
                        </button>
                    </div>
                </div>

                <svg
                    ref="sociogramSvg"
                    class="sociogram-graph-svg"
                    viewBox="0 0 960 520"
                    aria-label="Grafo sociométrico"
                    @wheel.prevent="onGraphWheel"
                    @mousedown="onGraphMouseDown"
                    @mousemove="onGraphMouseMove"
                    @mouseup="onGraphMouseUp"
                    @mouseleave="onGraphMouseUp"
                >
                    <defs>
                        <marker id="arrow-real-choice" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
                            <path d="M0,0 L0,6 L9,3 z" fill="#1d4ed8" />
                        </marker>
                        <marker id="arrow-real-rejection" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
                            <path d="M0,0 L0,6 L9,3 z" fill="#dc2626" />
                        </marker>
                        <marker id="arrow-perceived-choice" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
                            <path d="M0,0 L0,6 L9,3 z" fill="#0ea5e9" />
                                    </marker>
                        <marker id="arrow-perceived-rejection" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
                            <path d="M0,0 L0,6 L9,3 z" fill="#f97316" />
                        </marker>
                    </defs>

                    <g :transform="graphTransform">
                        <g>
                            <path
                                v-for="edge in sociogramGraphEdges"
                                :key="edge.id"
                                :d="edge.path"
                                :stroke="edge.color"
                                :stroke-width="edge.strokeWidth"
                                :stroke-dasharray="edge.strokeDasharray"
                                fill="none"
                                :marker-end="edge.marker"
                                stroke-linecap="round"
                                :opacity="edge.opacity"
                            >
                                <title>{{ edge.tooltip }}</title>
                            </path>
                        </g>

                        <g>
                            <g
                                v-for="node in sociogramGraphNodes"
                                :key="node.id"
                                class="sociogram-node"
                            >
                                <circle
                                    :cx="node.x"
                                    :cy="node.y"
                                    :r="node.radius"
                                    :fill="node.isSelected ? '#00A499' : '#38bdf8'"
                                />
                                <text :x="node.x" :y="node.y + 4" text-anchor="middle">
                                    {{ node.label }}
                                </text>
                                <title>{{ node.fullName }}</title>
                            </g>
                        </g>
                    </g>
                </svg>

                <div class="sociogram-legend">
                    <span class="legend-item">
                        <i class="legend-line real-choice"></i>
                        Elección real (Q1)
                    </span>
                    <span class="legend-item">
                        <i class="legend-line real-rejection"></i>
                        Rechazo real (Q3)
                    </span>
                    <span class="legend-item">
                        <i class="legend-line perceived-choice"></i>
                        Elección percibida (Q2)
                    </span>
                    <span class="legend-item">
                        <i class="legend-line perceived-rejection"></i>
                        Rechazo percibido (Q4)
                    </span>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, watch, reactive } from "vue";
import { useDatetimeInput } from "~/composables/useDatetimeInput";

const props = defineProps({
    isOpen: Boolean,
    questionnaire: Object,
    sectionId: [String, Number],
    allStudents: Array,
});

const emit = defineEmits(["close"]);

const studentsData = ref([]);
const loading = ref(false);
const error = ref(null);
const filter = ref("all");

// Sociogram graph view state
const showSociogramGraph = ref(false);
const sociogramDetailStudent = ref(null);
const sociogramLayer = ref("real");
const sociogramSvg = ref(null);
const graphViewport = reactive({
    scale: 1,
    tx: 0,
    ty: 0,
    isPanning: false,
    lastX: 0,
    lastY: 0,
});

const sociogramQuestions = [
    {
        key: "question1",
        label: "¿A quién(es) elegirías para trabajar en equipo?",
    },
    {
        key: "question2",
        label: "¿Quién(es) crees que te elegirían para trabajar en equipo?",
    },
    {
        key: "question3",
        label: "¿A quién(es) NO elegirías para trabajar en equipo?",
    },
    {
        key: "question4",
        label: "¿Quién(es) crees que NO te elegirían para trabajar en equipo?",
    },
];

// Determina si el cuestionario activo es un sociograma
const isSociogram = computed(
    () =>
        props.questionnaire?.plugin === "group_compatibility" ||
        props.questionnaire?.survey_type != null,
);

const completedCount = computed(
    () => studentsData.value.filter((s) => s.has_answered).length,
);

const pendingCount = computed(
    () => studentsData.value.filter((s) => !s.has_answered).length,
);

const completionPercentage = computed(() => {
    if (studentsData.value.length === 0) return 0;
    return Math.round((completedCount.value / studentsData.value.length) * 100);
});

const filteredStudents = computed(() => {
    if (filter.value === "completed") {
        return studentsData.value.filter((s) => s.has_answered);
    } else if (filter.value === "pending") {
        return studentsData.value.filter((s) => !s.has_answered);
    }
    return studentsData.value;
});

const sociogramGraphNodes = computed(() => {
    const selected = sociogramDetailStudent.value;
    if (!selected) return [];

    const selectedId = String(selected.id);
    const response = selected.sociogram_response;
    if (!response) return [];

    // Collect IDs of students the selected student actually interacted with (response !== 0)
    const connectedIds = new Set();

    const addConnected = (answers) => {
        if (!Array.isArray(answers)) return;
        answers.forEach((a) => {
            const val = Number(a.response);
            if (val !== 0) {
                connectedIds.add(String(a.student_id));
            }
        });
    };

    addConnected(response.question1);
    addConnected(response.question2);
    addConnected(response.question3);
    addConnected(response.question4);

    console.log("[Sociogram Debug] Selected ID:", selectedId);
    console.log("[Sociogram Debug] Response:", response);
    console.log("[Sociogram Debug] Connected IDs:", Array.from(connectedIds));

    if (connectedIds.size === 0) return [];

    // Get the students that are in the connected set
    const all = studentsData.value || props.allStudents || [];
    const connectedStudents = all.filter((s) => connectedIds.has(String(s.id)));

    const total = connectedStudents.length;
    if (total === 0) return [];

    const cx = 480;
    const cy = 260;
    const nodeRadius = 22;
    const orbitRadius = Math.max(140, Math.min(220, 120 + total * 8));

    const nodes = [];

    // Selected student at the center
    const fullName = getStudentName(selectedId);
    const nameParts = (fullName || "").replace(/^ID \d+$/, "").split(" ").filter(Boolean);
    const initials = nameParts.length >= 2
        ? nameParts.slice(0, 2).map((n) => n.charAt(0)).join("").toUpperCase()
        : (nameParts[0] || selected.id || "?").substring(0, 2).toUpperCase();

    nodes.push({
        id: selectedId,
        label: initials,
        fullName: `${fullName} (Seleccionado)`,
        x: cx,
        y: cy,
        radius: nodeRadius + 6,
        isSelected: true,
    });

    // Connected students around in a circle
    connectedStudents.forEach((student, idx) => {
        const angle = (Math.PI * 2 * idx) / total - Math.PI / 2;
        const x = cx + orbitRadius * Math.cos(angle);
        const y = cy + orbitRadius * Math.sin(angle);
        const sFullName = getStudentName(student.id);
        const sNameParts = (sFullName || "").replace(/^ID \d+$/, "").split(" ").filter(Boolean);
        const sInitials = sNameParts.length >= 2
            ? sNameParts.slice(0, 2).map((n) => n.charAt(0)).join("").toUpperCase()
            : (sNameParts[0] || student.id || "?").substring(0, 2).toUpperCase();
        nodes.push({
            id: String(student.id),
            label: sInitials,
            fullName: sFullName,
            x,
            y,
            radius: nodeRadius,
            isSelected: false,
        });
    });

    return nodes;
});

const sociogramGraphNodeMap = computed(() => {
    const map = {};
    sociogramGraphNodes.value.forEach((n) => {
        map[n.id] = n;
    });
    return map;
});

const sociogramEdges = computed(() => {
    const selected = sociogramDetailStudent.value;
    if (!selected) return [];

    const selectedId = String(selected.id);
    const response = selected.sociogram_response;
    if (!response) return [];

    const edges = [];

    const pushEdges = (sourceID, answers, expectedResponse, type, question) => {
        if (!Array.isArray(answers)) return;
        answers.forEach((answer) => {
            const targetID = String(answer.student_id || "");
            if (!targetID || targetID === sourceID) return;
            if (answer.response !== expectedResponse) return;
            edges.push({
                source: sourceID,
                target: targetID,
                type,
                question,
                weight: expectedResponse,
            });
        });
    };

    // Only use the selected student's responses
    pushEdges(selectedId, response.question1, 1, "real_choice", "Q1");
    pushEdges(selectedId, response.question3, 1, "real_rejection", "Q3");
    pushEdges(selectedId, response.question2, 1, "perceived_choice", "Q2");
    pushEdges(selectedId, response.question4, 1, "perceived_rejection", "Q4");

    const index = new Map();
    edges.forEach((edge, idx) => {
        index.set(`${edge.source}|${edge.target}|${edge.type}`, idx);
    });

    return edges.map((edge) => {
        const isMutual = edge.isMutual || index.has(
            `${edge.target}|${edge.source}|${edge.type}`,
        );
        const isConflicted =
            (edge.type === "real_choice" &&
                index.has(`${edge.target}|${edge.source}|real_rejection`)) ||
            (edge.type === "real_rejection" &&
                index.has(`${edge.target}|${edge.source}|real_choice`));

        return {
            ...edge,
            isMutual,
            isConflicted,
        };
    });
});

const visibleSociogramEdges = computed(() => {
    if (sociogramLayer.value === "real") {
        return sociogramEdges.value.filter((edge) => edge.type.startsWith("real_"));
    }
    if (sociogramLayer.value === "perceived") {
        return sociogramEdges.value.filter((edge) =>
            edge.type.startsWith("perceived_"),
        );
    }
    return sociogramEdges.value;
});

const sociogramGraphEdges = computed(() => {
    const nodeMap = sociogramGraphNodeMap.value;
    const rendered = [];
    const reverseMap = new Set(
        visibleSociogramEdges.value.map((edge) => `${edge.target}|${edge.source}`),
    );

    const edgeStyle = (type) => {
        if (type === "real_choice") {
            return {
                color: "#1d4ed8",
                dash: "none",
                marker: "url(#arrow-real-choice)",
            };
        }
        if (type === "real_rejection") {
            return {
                color: "#dc2626",
                dash: "none",
                marker: "url(#arrow-real-rejection)",
            };
        }
        if (type === "perceived_choice") {
            return {
                color: "#0ea5e9",
                dash: "6 4",
                marker: "url(#arrow-perceived-choice)",
            };
        }
        return {
            color: "#f97316",
            dash: "6 4",
            marker: "url(#arrow-perceived-rejection)",
        };
    };

    visibleSociogramEdges.value.forEach((edge, idx) => {
        const source = nodeMap[edge.source];
        const target = nodeMap[edge.target];
        if (!source || !target) return;

        const dx = target.x - source.x;
        const dy = target.y - source.y;
        const dist = Math.sqrt(dx * dx + dy * dy) || 1;
        const ux = dx / dist;
        const uy = dy / dist;
        const startX = source.x + ux * source.radius;
        const startY = source.y + uy * source.radius;
        const endX = target.x - ux * target.radius;
        const endY = target.y - uy * target.radius;

        let path = `M ${startX} ${startY} L ${endX} ${endY}`;
        if (reverseMap.has(`${edge.source}|${edge.target}`)) {
            const mx = (startX + endX) / 2;
            const my = (startY + endY) / 2;
            const nx = -uy;
            const ny = ux;
            const curve = edge.source < edge.target ? 18 : -18;
            const cx = mx + nx * curve;
            const cy = my + ny * curve;
            path = `M ${startX} ${startY} Q ${cx} ${cy} ${endX} ${endY}`;
        }

        const style = edgeStyle(edge.type);
        const sourceName = getStudentName(edge.source);
        const targetName = getStudentName(edge.target);
        const mutualLabel = edge.isMutual ? " (reciproco)" : "";
        const conflictLabel = edge.isConflicted ? " (conflicto)" : "";

        rendered.push({
            id: `${edge.source}-${edge.target}-${edge.type}-${idx}`,
            path,
            color: style.color,
            strokeDasharray: style.dash,
            marker: style.marker,
            strokeWidth: edge.isMutual ? 3.4 : 2.2,
            opacity: edge.isConflicted ? 0.95 : 0.8,
            tooltip: `${edge.question}: ${sourceName} -> ${targetName}${mutualLabel}${conflictLabel}`,
        });
    });

    return rendered;
});

const graphTransform = computed(
    () =>
        `translate(${graphViewport.tx} ${graphViewport.ty}) scale(${graphViewport.scale})`,
);

const sociogramMetrics = computed(() => {
    const nodes = sociogramGraphNodes.value;
    const edges = visibleSociogramEdges.value;
    const nodeIds = nodes.map((n) => n.id);
    const inDegree = {};
    const outDegree = {};

    nodeIds.forEach((id) => {
        inDegree[id] = 0;
        outDegree[id] = 0;
    });

    edges.forEach((edge) => {
        if (!(edge.source in outDegree) || !(edge.target in inDegree)) return;
        outDegree[edge.source] += 1;
        inDegree[edge.target] += 1;
    });

    const isolates = nodeIds.filter(
        (id) => inDegree[id] === 0 && outDegree[id] === 0,
    );

    const maxDenominator = Math.max(1, nodes.length - 1);
    const centralityTop = nodeIds
        .map((id) => {
            const degreeTotal = inDegree[id] + outDegree[id];
            const score = degreeTotal / (2 * maxDenominator);
            return {
                id,
                name: getStudentName(id),
                inDegree: inDegree[id],
                outDegree: outDegree[id],
                score,
            };
        })
        .sort((a, b) => b.score - a.score)
        .slice(0, 3);

    const cliqueAdj = buildCliqueAdjacency(nodeIds, edges);
    const cliques = findMaximalCliques(cliqueAdj)
        .filter((clique) => clique.length >= 3)
        .sort((a, b) => b.length - a.length)
        .slice(0, 3)
        .map((clique) => clique.map((id) => getStudentName(id)));

    return {
        isolates,
        centralityTop,
        cliques,
    };
});

const mutualEdgesCount = computed(() => {
    const seen = new Set();
    let total = 0;
    visibleSociogramEdges.value.forEach((edge) => {
        if (!edge.isMutual) return;
        const keyA = `${edge.source}|${edge.target}|${edge.type}`;
        const keyB = `${edge.target}|${edge.source}|${edge.type}`;
        if (seen.has(keyA) || seen.has(keyB)) return;
        seen.add(keyA);
        seen.add(keyB);
        total += 1;
    });
    return total;
});

watch(
    () => props.isOpen,
    (newValue) => {
        if (newValue && props.questionnaire && props.allStudents) {
            loadStudentsResponses();
        }
    },
);

watch(
    () => props.allStudents,
    (newStudents) => {
        if (props.isOpen && props.questionnaire && newStudents) {
            loadStudentsResponses();
        }
    },
);

const loadStudentsResponses = async () => {
    loading.value = true;
    error.value = null;

    try {
        const token = localStorage.getItem("auth_token");
        const API_BASE_URL = "http://127.0.0.1:3000/api/v1";

        const questionnaireId =
            props.questionnaire.id || props.questionnaire._id;

        // Elegir el endpoint correcto según el tipo de cuestionario
        const plugin = props.questionnaire?.plugin || "";
        let endpoint;
        if (plugin === "group_compatibility") {
            endpoint = `${API_BASE_URL}/core/group-compatibility/surveys/${questionnaireId}/responses`;
        } else if (plugin === "bfi44") {
            endpoint = `${API_BASE_URL}/core/bfi44/questionnaires/${questionnaireId}/responses`;
        } else if (plugin === "mbti") {
            endpoint = `${API_BASE_URL}/core/mbti/questionnaires/${questionnaireId}/responses`;
        } else {
            // eneagram y cualquier otro plugin
            endpoint = `${API_BASE_URL}/core/enneagram/questionnaires/${questionnaireId}/responses`;
        }

        const response = await fetch(endpoint, {
            headers: {
                Authorization: `Bearer ${token}`,
                "Content-Type": "application/json",
            },
        });

        if (!response.ok) {
            throw new Error(
                "Error al cargar las respuestas de los estudiantes",
            );
        }

        const data = await response.json();

        if (isSociogram.value) {
            // ── Sociograma ─────────────────────────────────────────────────────
            // Las respuestas usan student_id numérico (string), no email
            const responsesMap = new Map();
            if (data.success && Array.isArray(data.data)) {
                data.data.forEach((resp) => {
                    if (resp.student_id != null) {
                        responsesMap.set(String(resp.student_id), resp);
                    }
                });
            }

            if (props.allStudents && props.allStudents.length > 0) {
                studentsData.value = props.allStudents.map((student) => {
                    const sid = String(student.id);
                    const resp = responsesMap.get(sid);
                    const studentName =
                        student.full_name ||
                        `${student.nombres || ""} ${student.paterno || ""} ${student.materno || ""}`.trim();

                    if (resp) {
                        return {
                            id: student.id,
                            name: studentName,
                            email: student.email || "",
                            is_complete: resp.is_complete ?? true,
                            has_answered: true,
                            estado: "Respondido",
                            submitted_at:
                                resp.submitted_at || resp.created_at || null,
                            result: null,
                            dominant_type: null,
                            ocean_vector: null,
                            sociogram_response: {
                                question1: resp.question1 || [],
                                question2: resp.question2 || [],
                                question3: resp.question3 || [],
                                question4: resp.question4 || [],
                            },
                        };
                    } else {
                        return {
                            id: student.id,
                            name: studentName,
                            email: student.email || "",
                            is_complete: false,
                            has_answered: false,
                            estado: "Pendiente",
                            submitted_at: null,
                            result: null,
                            dominant_type: null,
                            ocean_vector: null,
                            sociogram_response: null,
                        };
                    }
                });
            } else {
                studentsData.value = [];
            }
        } else {
            // ── Psicométrico (Enneagram, BFI-44, MBTI) ────────────────────────
            if (data.success && data.data && data.data.length > 0) {
                // Crear mapa por student_id o email según el plugin
                const responsesMapById = new Map();
                const responsesMapByEmail = new Map();
                
                data.data.forEach((resp) => {
                    // BFI-44 usa student_id (numérico o string)
                    if (resp.student_id) {
                        responsesMapById.set(String(resp.student_id), resp);
                    }
                    // Otros plugins usan email
                    if (resp.email) {
                        responsesMapByEmail.set(resp.email.toLowerCase(), resp);
                    }
                });

                if (props.allStudents && props.allStudents.length > 0) {
                    studentsData.value = props.allStudents.map((student) => {
                        // Intentar match por student_id primero (para BFI-44)
                        const studentId = String(student.id);
                        let resp = responsesMapById.get(studentId);
                        
                        // Si no encuentra, intentar por email (para Enneagram, MBTI)
                        if (!resp) {
                            const studentEmail = student.email?.toLowerCase();
                            resp = studentEmail ? responsesMapByEmail.get(studentEmail) : null;
                        }

                        if (resp) {
                            const isComplete = resp.is_complete || false;
                            const dominantType =
                                resp.results?.dominant_type ||
                                resp.dominant_type ||
                                null;
                            const oceanVector =
                                resp.results?.ocean_vector || null;

                            return {
                                id: student.id,
                                name:
                                    resp.name ||
                                    student.full_name ||
                                    `${student.nombres || ""} ${student.paterno || ""} ${student.materno || ""}`.trim(),
                                email: resp.email || student.email || "",
                                is_complete: isComplete,
                                has_answered: isComplete,
                                estado: isComplete ? "Respondido" : "Pendiente",
                                submitted_at: resp.submitted_at || null,
                                result: dominantType,
                                dominant_type: dominantType,
                                ocean_vector: oceanVector,
                            };
                        } else {
                            return {
                                id: student.id,
                                name:
                                    student.full_name ||
                                    `${student.nombres || ""} ${student.paterno || ""} ${student.materno || ""}`.trim(),
                                email: student.email || "",
                                is_complete: false,
                                has_answered: false,
                                estado: "Pendiente",
                                submitted_at: null,
                                result: null,
                                dominant_type: null,
                            };
                        }
                    });
                } else {
                    studentsData.value = data.data.map((resp) => {
                        const isComplete = resp.is_complete || false;
                        const dominantType =
                            resp.results?.dominant_type ||
                            resp.dominant_type ||
                            null;
                        const oceanVector = resp.results?.ocean_vector || null;

                        return {
                            name: resp.name || "",
                            email: resp.email || "",
                            is_complete: isComplete,
                            has_answered: isComplete,
                            estado: isComplete ? "Respondido" : "Pendiente",
                            submitted_at: resp.submitted_at || null,
                            result: dominantType,
                            dominant_type: dominantType,
                            ocean_vector: oceanVector,
                        };
                    });
                }
            } else if (props.allStudents && props.allStudents.length > 0) {
                studentsData.value = props.allStudents.map((student) => ({
                    id: student.id,
                    name:
                        student.full_name ||
                        `${student.nombres || ""} ${student.paterno || ""} ${student.materno || ""}`.trim(),
                    email: student.email || "",
                    is_complete: false,
                    has_answered: false,
                    estado: "Pendiente",
                    submitted_at: null,
                    result: null,
                    dominant_type: null,
                }));
            } else {
                studentsData.value = [];
            }
        }
    } catch (err) {
        console.error("Error loading students responses:", err);
        error.value = err.message || "Error al cargar las respuestas";
    } finally {
        loading.value = false;
    }
};

// Devuelve el nombre legible de un estudiante dado su id (string o número)
const getStudentName = (studentId) => {
    const sid = String(studentId);
    const found = props.allStudents?.find((s) => String(s.id) === sid);
    if (!found) return `ID ${sid}`;
    return (
        found.full_name ||
        `${found.nombres || ""} ${found.paterno || ""} ${found.materno || ""}`.trim() ||
        `ID ${sid}`
    );
};

const openSociogramResponse = (student) => {
    sociogramDetailStudent.value = student;
    sociogramLayer.value = "all";
    resetGraphViewport();
    showSociogramGraph.value = true;
};

const closeSociogramGraph = () => {
    showSociogramGraph.value = false;
    sociogramDetailStudent.value = null;
};

const resetGraphViewport = () => {
    graphViewport.scale = 1;
    graphViewport.tx = 0;
    graphViewport.ty = 0;
    graphViewport.isPanning = false;
};

const onGraphWheel = (event) => {
    if (!sociogramSvg.value) return;
    const rect = sociogramSvg.value.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;
    const zoomFactor = event.deltaY < 0 ? 1.12 : 0.9;
    const oldScale = graphViewport.scale;
    const newScale = Math.min(3.5, Math.max(0.45, oldScale * zoomFactor));
    const worldX = (x - graphViewport.tx) / oldScale;
    const worldY = (y - graphViewport.ty) / oldScale;

    graphViewport.scale = newScale;
    graphViewport.tx = x - worldX * newScale;
    graphViewport.ty = y - worldY * newScale;
};

const onGraphMouseDown = (event) => {
    if (event.button !== 0) return;
    graphViewport.isPanning = true;
    graphViewport.lastX = event.clientX;
    graphViewport.lastY = event.clientY;
};

const onGraphMouseMove = (event) => {
    if (!graphViewport.isPanning) return;
    const dx = event.clientX - graphViewport.lastX;
    const dy = event.clientY - graphViewport.lastY;
    graphViewport.tx += dx;
    graphViewport.ty += dy;
    graphViewport.lastX = event.clientX;
    graphViewport.lastY = event.clientY;
};

const onGraphMouseUp = () => {
    graphViewport.isPanning = false;
};

const exportSociogramSvg = () => {
    if (!sociogramSvg.value) return;
    const serializer = new XMLSerializer();
    let source = serializer.serializeToString(sociogramSvg.value);
    if (!source.includes("xmlns=\"http://www.w3.org/2000/svg\"")) {
        source = source.replace(
            "<svg",
            '<svg xmlns="http://www.w3.org/2000/svg"',
        );
    }
    const blob = new Blob([source], {
        type: "image/svg+xml;charset=utf-8",
    });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `sociograma_${props.questionnaire?.title || "encuesta"}.svg`;
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(url);
};

const exportSociogramPng = () => {
    if (!sociogramSvg.value) return;
    const serializer = new XMLSerializer();
    let source = serializer.serializeToString(sociogramSvg.value);
    if (!source.includes("xmlns=\"http://www.w3.org/2000/svg\"")) {
        source = source.replace(
            "<svg",
            '<svg xmlns="http://www.w3.org/2000/svg"',
        );
    }

    const svgBlob = new Blob([source], { type: "image/svg+xml;charset=utf-8" });
    const url = URL.createObjectURL(svgBlob);
    const image = new Image();

    image.onload = () => {
        const canvas = document.createElement("canvas");
        canvas.width = 960;
        canvas.height = 520;
        const ctx = canvas.getContext("2d");
        if (!ctx) {
            URL.revokeObjectURL(url);
            return;
        }
        ctx.fillStyle = "#ffffff";
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        ctx.drawImage(image, 0, 0, canvas.width, canvas.height);

        canvas.toBlob((pngBlob) => {
            if (!pngBlob) {
                URL.revokeObjectURL(url);
                return;
            }
            const pngUrl = URL.createObjectURL(pngBlob);
            const link = document.createElement("a");
            link.href = pngUrl;
            link.download = `sociograma_${props.questionnaire?.title || "encuesta"}.png`;
            document.body.appendChild(link);
            link.click();
            link.remove();
            URL.revokeObjectURL(pngUrl);
            URL.revokeObjectURL(url);
        }, "image/png");
    };

    image.onerror = () => {
        URL.revokeObjectURL(url);
    };

    image.src = url;
};

const buildCliqueAdjacency = (nodeIds, edges) => {
    const adj = {};
    nodeIds.forEach((id) => {
        adj[id] = new Set();
    });

    edges.forEach((edge) => {
        if (!edge.isMutual) return;
        if (!edge.type.endsWith("choice")) return;
        if (!adj[edge.source] || !adj[edge.target]) return;
        adj[edge.source].add(edge.target);
        adj[edge.target].add(edge.source);
    });
    return adj;
};

const findMaximalCliques = (adj) => {
    const result = [];
    const P = new Set(Object.keys(adj));

    const bronKerbosch = (R, Pset, Xset) => {
        if (Pset.size === 0 && Xset.size === 0) {
            result.push(Array.from(R));
            return;
        }

        const pValues = Array.from(Pset);
        for (const v of pValues) {
            const neighbors = adj[v] || new Set();
            const nextR = new Set(R);
            nextR.add(v);
            const nextP = new Set(Array.from(Pset).filter((n) => neighbors.has(n)));
            const nextX = new Set(Array.from(Xset).filter((n) => neighbors.has(n)));
            bronKerbosch(nextR, nextP, nextX);
            Pset.delete(v);
            Xset.add(v);
        }
    };

    bronKerbosch(new Set(), P, new Set());
    return result;
};

const { formatDatetimeDisplay } = useDatetimeInput();
const formatDate = (date) => formatDatetimeDisplay(date) || "-";

const exportToCSV = () => {
    const headers = [
        "Nombre",
        "Email",
        "Estado",
        "Fecha de respuesta",
        "Resultado",
    ];
    const rows = studentsData.value.map((student) => [
        student.name,
        student.email,
        student.is_complete ? "Completado" : "Pendiente",
        formatDate(student.submitted_at),
        isSociogram.value
            ? student.has_answered
                ? "Respondido"
                : "-"
            : student.result
              ? `Eneatipo ${student.dominant_type}`
              : "-",
    ]);

    const csvContent = [
        headers.join(","),
        ...rows.map((row) => row.map((cell) => `"${cell}"`).join(",")),
    ].join("\n");

    const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
    const link = document.createElement("a");
    const url = URL.createObjectURL(blob);

    link.setAttribute("href", url);
    link.setAttribute(
        "download",
        `respuestas_${props.questionnaire?.title || "cuestionario"}.csv`,
    );
    link.style.visibility = "hidden";

    document.body.appendChild(link);
    link.click();
    link.remove();
};

const closeModal = () => {
    emit("close");
};
</script>

<style scoped>
/* ── Botón ver respuesta sociograma ── */
.view-response-btn {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 4px 10px;
    font-size: 0.78rem;
    font-weight: 500;
    color: #008a81;
    background: rgba(0, 164, 153, 0.08);
    border: 1px solid rgba(0, 164, 153, 0.3);
    border-radius: 6px;
    cursor: pointer;
    transition:
        background 0.18s,
        border-color 0.18s;
}
.view-response-btn:hover {
    background: rgba(0, 164, 153, 0.16);
    border-color: rgba(0, 164, 153, 0.55);
}

/* ── Sociogram graph view — Analytical Dark Canvas ── */
.sociogram-graph-view {
    display: flex;
    flex-direction: column;
    gap: 0.85rem;
    margin-top: 0.25rem;
    animation: graphViewEnter 0.25s ease-out;
}

@keyframes graphViewEnter {
    from { opacity: 0; transform: translateY(6px); }
    to   { opacity: 1; transform: translateY(0); }
}

/* Toolbar with frosted glass effect */
.sociogram-graph-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
    padding: 0.6rem 0.85rem;
    background: rgba(15, 23, 42, 0.04);
    backdrop-filter: blur(8px);
    border-radius: 10px;
    border: 1px solid rgba(15, 23, 42, 0.08);
}

.sociogram-toolbar-title {
    display: flex;
    align-items: baseline;
    gap: 0.4rem;
    min-width: 0;
}

.sociogram-toolbar-label {
    font-size: 0.72rem;
    font-weight: 600;
    color: #94a3b8;
    text-transform: uppercase;
    letter-spacing: 0.8px;
}

.sociogram-toolbar-name {
    font-size: 0.95rem;
    font-weight: 700;
    color: #0f172a;
    font-family: 'Red Hat Display', sans-serif;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.sociogram-graph-filters {
    display: flex;
    gap: 0.35rem;
    flex-wrap: wrap;
}

.sociogram-filter {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    font-size: 0.74rem;
    font-weight: 600;
    color: #475569;
    background: rgba(255, 255, 255, 0.7);
    border: 1.5px solid rgba(15, 23, 42, 0.1);
    border-radius: 999px;
    padding: 0.25rem 0.65rem;
    cursor: pointer;
    transition: all 0.18s ease;
    letter-spacing: 0.1px;
}

.sociogram-filter:hover {
    background: rgba(0, 164, 153, 0.08);
    border-color: rgba(0, 164, 153, 0.4);
    color: #00A499;
}

.sociogram-filter input {
    display: none;
}

/* Active filter pill — teal accent */
.sociogram-filter:has(input:checked) {
    background: #00A499;
    border-color: #00A499;
    color: #fff;
    box-shadow: 0 2px 8px rgba(0, 164, 153, 0.35);
}

.toolbar-actions {
    display: flex;
    gap: 0.35rem;
}

.btn-export-graph {
    border: 1.5px solid rgba(15, 23, 42, 0.12);
    background: rgba(255, 255, 255, 0.8);
    color: #334155;
    font-size: 0.74rem;
    font-weight: 600;
    padding: 0.3rem 0.75rem;
    border-radius: 7px;
    cursor: pointer;
    transition: all 0.18s ease;
    letter-spacing: 0.1px;
}

.btn-export-graph:hover {
    background: rgba(15, 23, 42, 0.06);
    border-color: rgba(15, 23, 42, 0.22);
    color: #0f172a;
}

.btn-export-graph:active {
    transform: scale(0.96);
}

/* Light canvas SVG — the focal point */
.sociogram-graph-svg {
    width: 100%;
    min-height: 420px;
    border-radius: 14px;
    background: #ffffff;
    cursor: grab;
    border: 1px solid rgba(0, 0, 0, 0.08);
    box-shadow:
        inset 0 1px 0 rgba(255, 255, 255, 1),
        0 4px 24px rgba(0, 0, 0, 0.08);
}

.sociogram-graph-svg:active {
    cursor: grabbing;
}

/* Node styling — dark on light */
.sociogram-node circle {
    stroke: rgba(0, 0, 0, 0.8);
    stroke-width: 2.5;
    filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.15));
    transition: r 0.2s ease;
}

.sociogram-node:hover circle {
    stroke-width: 3.5;
}

.sociogram-node text {
    fill: #1e293b;
    font-size: 11px;
    font-weight: 700;
    font-family: 'Red Hat Display', 'Segoe UI', sans-serif;
    pointer-events: none;
    text-shadow: 0 1px 2px rgba(255, 255, 255, 0.8);
    letter-spacing: 0.5px;
}

/* Clean inline legend bar */
.sociogram-legend {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem 1rem;
    padding: 0.55rem 0.85rem;
    background: rgba(15, 23, 42, 0.03);
    border-radius: 8px;
    border: 1px solid rgba(15, 23, 42, 0.07);
}

.legend-item {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.72rem;
    color: #475569;
    font-weight: 600;
    letter-spacing: 0.2px;
}

.legend-line {
    width: 20px;
    height: 0;
    border-top-width: 2.5px;
    border-top-style: solid;
    flex-shrink: 0;
}

.legend-line.real-choice    { border-top-color: #3b82f6; }
.legend-line.real-rejection { border-top-color: #ef4444; }
.legend-line.perceived-choice {
    border-top-color: #38bdf8;
    border-top-style: dashed;
}
.legend-line.perceived-rejection {
    border-top-color: #fb923c;
    border-top-style: dashed;
}

.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 1rem;
}

.modal-content {
    background: white;
    border-radius: 12px;
    max-width: 1200px;
    width: 100%;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
    display: flex;
    flex-direction: column;
}

.modal-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 1.5rem;
    border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.modal-header h2 {
    margin: 0 0 0.25rem 0;
    color: var(--color-dark, #394049);
    font-size: 1.5rem;
}

.modal-subtitle {
    margin: 0;
    color: #94a3b8;
    font-size: 0.875rem;
}

.close-btn {
    background: none;
    border: none;
    cursor: pointer;
    color: #94a3b8;
    padding: 0.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    transition: all 0.2s;
}

.close-btn:hover {
    background: var(--color-background, #f4f6f9);
    color: var(--color-dark, #394049);
}

.modal-body {
    flex: 1;
    padding: 1.5rem;
    overflow-y: auto;
}

.loading-state,
.error-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    color: #94a3b8;
}

.spinner {
    width: 40px;
    height: 40px;
    border: 3px solid var(--color-border, #dfe4ea);
    border-top-color: var(--color-primary, #00a499);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 1rem;
}

@keyframes spin {
    to {
        transform: rotate(360deg);
    }
}

.error-state svg {
    color: var(--color-danger, #d7373f);
    margin-bottom: 1rem;
}

/* Stats Section */
.stats-section {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 1rem;
    margin-bottom: 1.5rem;
}

.stat-card {
    background: var(--color-background, #f4f6f9);
    border: 1px solid var(--color-border, #dfe4ea);
    border-radius: 8px;
    padding: 1rem;
    text-align: center;
}

.stat-card.completed {
    background: #f0fdf4;
    border-color: #86efac;
}

.stat-card.pending {
    background: #fef3c7;
    border-color: #fcd34d;
}

.stat-card.percentage {
    background: #ede9fe;
    border-color: #c4b5fd;
}

.stat-value {
    font-size: 2rem;
    font-weight: 700;
    color: var(--color-dark, #394049);
    margin-bottom: 0.25rem;
}

.stat-label {
    font-size: 0.75rem;
    color: #94a3b8;
    text-transform: uppercase;
    font-weight: 500;
}

/* Filter Section */
.filter-section {
    display: flex;
    gap: 1.5rem;
    margin-bottom: 1.5rem;
    padding: 1rem;
    background: var(--color-background, #f4f6f9);
    border-radius: 8px;
}

.filter-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    font-size: 0.875rem;
    color: var(--color-dark, #394049);
}

.filter-label input[type="radio"] {
    cursor: pointer;
}

/* Table */
.table-container {
    border: 1px solid var(--color-border, #dfe4ea);
    border-radius: 8px;
    overflow: hidden;
}

.students-table {
    width: 100%;
    border-collapse: collapse;
}

.students-table thead {
    background: var(--color-background, #f4f6f9);
}

.students-table th {
    padding: 0.75rem 1rem;
    text-align: left;
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--color-text-secondary, #64748b);
    text-transform: uppercase;
    border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.students-table td {
    padding: 0.75rem 1rem;
    font-size: 0.875rem;
    color: var(--color-dark, #394049);
    border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.students-table tbody tr:last-child td {
    border-bottom: none;
}

.students-table tbody tr:hover {
    background: var(--color-background, #f4f6f9);
}

.completed-row {
    background: #f0fdf4 !important;
}

.student-name {
    font-weight: 500;
}

.student-email {
    color: #94a3b8;
}

.status-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem 0.75rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 500;
}

.status-badge.completed {
    background: #d1fae5;
    color: #065f46;
}

.status-badge.pending {
    background: #fed7aa;
    color: #92400e;
}

.result-badge {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    background: #dbeafe;
    color: #1e40af;
    border-radius: 6px;
    font-size: 0.75rem;
    font-weight: 500;
    margin-bottom: 0.25rem;
}

.result-details {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.ocean-mini-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
}

.ocean-mini-val {
    font-size: 0.65rem;
    font-weight: 600;
    padding: 0.1rem 0.3rem;
    border-radius: 3px;
    background: #f1f5f9;
    color: #475569;
    border-left: 2px solid #94a3b8;
}

.ocean-mini-val.o {
    border-color: #ef4444;
}
.ocean-mini-val.c {
    border-color: #f59e0b;
}
.ocean-mini-val.e {
    border-color: #3b82f6;
}
.ocean-mini-val.a {
    border-color: #10b981;
}
.ocean-mini-val.n {
    border-color: #8b5cf6;
}

.no-result {
    color: #cbd5e0;
}

.no-data {
    padding: 3rem;
    text-align: center;
    color: #94a3b8;
    font-style: italic;
}

/* Modal Footer */
.modal-footer {
    display: flex;
    gap: 1rem;
    justify-content: space-between;
    padding: 1.5rem;
    border-top: 1px solid var(--color-border, #dfe4ea);
    background: var(--color-background, #f4f6f9);
}

.btn {
    padding: 0.75rem 1.5rem;
    border-radius: 6px;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    border: none;
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.btn-export {
    background: var(--color-primary, #006b64);
    color: white;
}

.btn-export:hover {
    background: var(--color-primary-dark, #008a7f);
}

.btn-close {
    background: var(--color-background, #f4f6f9);
    color: var(--color-dark, #394049);
}

.btn-close:hover {
    background: var(--color-border, #dfe4ea);
}

@media (max-width: 768px) {
    .graph-header-row {
        flex-direction: column;
        align-items: flex-start;
    }

    .sociogram-graph-toolbar {
        flex-direction: column;
        align-items: flex-start;
    }

    .sociogram-graph-svg {
        min-height: 220px;
    }

    .stats-section {
        grid-template-columns: repeat(2, 1fr);
    }

    .filter-section {
        flex-direction: column;
        gap: 0.75rem;
    }

    .table-container {
        overflow-x: auto;
    }

    .students-table {
        min-width: 600px;
    }

    .modal-footer {
        flex-direction: column;
    }

    .btn {
        width: 100%;
        justify-content: center;
    }
}
</style>
