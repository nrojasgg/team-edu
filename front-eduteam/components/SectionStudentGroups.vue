<template>
    <div class="groups-tab">
        <div class="groups-content">
            <!-- Coordinación  de Cuestionarios -->
            <div class="questionnaires-section">
                <div class="questionnaires-header">
                    <h3>Cuestionarios</h3>
                    <button
                        v-if="!readOnly"
                        @click="$emit('create-questionnaire')"
                        class="add-questionnaire-btn"
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
                            <path d="M12 8v8" />
                            <path d="M8 12h8" />
                        </svg>
                        Añadir Test
                    </button>
                </div>

                <div
                    v-if="loadingQuestionnaires"
                    class="loading-questionnaires"
                >
                    Cargando cuestionarios...
                </div>

                <div
                    v-else-if="questionnaires.length === 0"
                    class="no-questionnaires"
                >
                    No hay cuestionarios creados para esta sección
                </div>

                <div v-else class="questionnaires-container">
                    <div
                        v-for="questionnaire in questionnaires"
                        :key="questionnaire.id || questionnaire._id"
                        class="questionnaire-card"
                        :class="{
                            'questionnaire-active':
                                getQuestionnaireStatus(questionnaire) ===
                                'active',
                            'questionnaire-expired':
                                getQuestionnaireStatus(questionnaire) ===
                                'expired',
                            'questionnaire-draft':
                                getQuestionnaireStatus(questionnaire) ===
                                'draft',
                        }"
                    >
                        <div class="questionnaire-header">
                            <h4 class="questionnaire-title">
                                {{
                                    questionnaire.title ||
                                    questionnaire.name ||
                                    "Test sin nombre"
                                }}
                            </h4>
                            <span class="questionnaire-plugin">{{
                                getPluginLabel(questionnaire)
                            }}</span>
                        </div>

                        <div class="questionnaire-info">
                            <p class="questionnaire-description">
                                {{
                                    questionnaire.description ||
                                    "Sin descripción"
                                }}
                            </p>

                            <div class="questionnaire-dates">
                                <div class="date-info">
                                    <strong>Inicio:</strong>
                                    {{ formatDate(questionnaire.start_date) }}
                                </div>
                                <div class="date-info">
                                    <strong>Término:</strong>
                                    {{ formatDate(questionnaire.end_date) }}
                                </div>
                            </div>
                        </div>

                        <div class="questionnaire-actions">
                            <!-- Student view: Only answer button -->
                            <button
                                v-if="readOnly"
                                class="action-btn view-btn"
                                title="Responder test"
                                @click="openEnneagramTest(questionnaire)"
                            >
                                <svg
                                    width="16"
                                    height="16"
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
                                Responder
                            </button>

                            <!-- Admin view: Edit, View responses, Delete -->
                            <template v-else>
                                <button
                                    class="action-btn edit-btn"
                                    title="Editar cuestionario"
                                    @click="
                                        $emit(
                                            'edit-questionnaire',
                                            questionnaire,
                                        )
                                    "
                                >
                                    <svg
                                        width="16"
                                        height="16"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                    >
                                        <path
                                            d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"
                                        />
                                        <path
                                            d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"
                                        />
                                    </svg>
                                    <span class="btn-text">Editar</span>
                                </button>
                                <button
                                    class="action-btn view-btn"
                                    title="Respuestas"
                                    @click="
                                        $emit('view-responses', questionnaire)
                                    "
                                >
                                    <svg
                                        width="16"
                                        height="16"
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
                                    <span class="btn-text">Respuestas</span>
                                </button>
                                <button
                                    @click="
                                        $emit(
                                            'delete-questionnaire',
                                            questionnaire,
                                        )
                                    "
                                    class="action-btn delete-btn"
                                    title="Eliminar cuestionario"
                                >
                                    <svg
                                        width="16"
                                        height="16"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                    >
                                        <polyline points="3,6 5,6 21,6" />
                                        <path
                                            d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2 2h4a2 2 0 0 1 2 2v2"
                                        />
                                        <path d="M10 11V17" />
                                        <path d="M14 11V17" />
                                    </svg>
                                    <span class="btn-text">Eliminar</span>
                                </button>
                            </template>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Panel de Estudiantes disponibles -->
            <div v-if="!readOnly" class="available-students-panel">
                <h3>Estudiantes disponibles</h3>
                <div class="grouping-buttons">
                    <button
                        @click="openFormByPluginModal"
                        class="grouping-btn plugin-btn"
                        :disabled="availableStudents.length === 0"
                    >
                        <svg
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                        >
                            <path
                                d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"
                            />
                        </svg>
                        Formar grupos en base a plugins
                    </button>
                    <button
                        @click="formRandomGroups"
                        class="grouping-btn random-btn"
                        :disabled="availableStudents.length === 0"
                    >
                        <svg
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                        >
                            <path
                                d="M14.828 14.828a4 4 0 0 1-5.656 0M9 10h1.586a1 1 0 0 1 .707.293l.707.707A1 1 0 0 0 13.414 11H15m-3-3h.01M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0z"
                            />
                        </svg>
                        Formar grupos aleatorios
                    </button>
                </div>
                <div
                    class="students-pool"
                    @dragover="onDragOver"
                    @drop="onDropToAvailable"
                >
                    <div
                        v-for="student in availableStudents"
                        :key="`available-${student.id}`"
                        class="student-card"
                        draggable="true"
                        @dragstart="onDragStart($event, student, 'available')"
                    >
                        <div class="student-avatar">
                            <svg
                                width="24"
                                height="24"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                            >
                                <path
                                    d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"
                                />
                                <circle cx="12" cy="7" r="4" />
                            </svg>
                        </div>
                        <div class="student-info">
                            <div class="student-name">
                                {{ student.nombres || student.first_name }}
                                {{ student.paterno || student.last_name }}
                            </div>
                            <div class="student-details">
                                {{ student.email }}
                                <span
                                    v-if="student.eneatype"
                                    class="eneatype-badge"
                                    >Eneatipo {{ student.eneatype }}</span
                                >
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Panel de grupos -->
            <div v-if="!readOnly" class="groups-panel">
                <h3>Grupos de trabajo</h3>
                <div class="groups-header">
                    <div
                        v-if="globalScore != null"
                        class="global-score-badge"
                        title="Puntaje global promedio de todos los grupos formados"
                    >
                        <span class="global-score-label">Puntaje global</span>
                        <span class="global-score-value">
                            {{ (globalScore * 100).toFixed(1) }}%
                        </span>
                    </div>
                    <button @click="addNewGroup" class="add-group-btn">
                        <svg
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                        >
                            <circle cx="12" cy="12" r="10" />
                            <path d="M12 8v8" />
                            <path d="M8 12h8" />
                        </svg>
                        Nuevo Grupo
                    </button>
                    <button
                        @click="exportToExcel"
                        class="export-groups-btn"
                        :disabled="groups.length === 0"
                        title="Exportar grupos a Excel"
                    >
                        <svg
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                        >
                            <path
                                d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"
                            />
                            <polyline points="7 10 12 15 17 10" />
                            <line x1="12" y1="15" x2="12" y2="3" />
                        </svg>
                        Exportar a Excel
                    </button>
                    <button
                        @click="deleteAllGroups"
                        class="delete-all-groups-btn"
                        :disabled="groups.length === 0"
                    >
                        <svg
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                        >
                            <path d="M3 6h18" />
                            <path
                                d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
                            />
                            <path d="M10 11v6" />
                            <path d="M14 11v6" />
                        </svg>
                        Eliminar todos los grupos
                    </button>
                </div>

                <div class="groups-container">
                    <div
                        v-for="group in groups"
                        :key="group.id"
                        class="group-card"
                    >
                        <div class="group-header">
                            <input
                                v-model="group.name"
                                class="group-name-input"
                                @blur="saveGroupName(group)"
                            />
                            <button
                                @click="deleteGroup(group.id)"
                                class="delete-group-btn"
                                title="Eliminar grupo"
                            >
                                <svg
                                    width="16"
                                    height="16"
                                    viewBox="0 0 24 24"
                                    fill="none"
                                    stroke="currentColor"
                                    stroke-width="2"
                                >
                                    <polyline points="3,6 5,6 21,6" />
                                    <path
                                        d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2 2h4a2 2 0 0 1 2 2v2"
                                    />
                                </svg>
                            </button>
                        </div>

                        <div
                            class="group-students-area"
                            @dragover="onDragOver"
                            @drop="onDropToGroup($event, group.id)"
                        >
                            <div
                                v-for="student in group.students"
                                :key="`group-${group.id}-${student.id}`"
                                class="student-card"
                                draggable="true"
                                @dragstart="
                                    onDragStart(
                                        $event,
                                        student,
                                        'group',
                                        group.id,
                                    )
                                "
                            >
                                <div class="student-avatar">
                                    <svg
                                        width="24"
                                        height="24"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        stroke="currentColor"
                                        stroke-width="2"
                                    >
                                        <path
                                            d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"
                                        />
                                        <circle cx="12" cy="7" r="4" />
                                    </svg>
                                </div>
                                <div class="student-info">
                                    <div class="student-name">
                                        {{
                                            student.nombres ||
                                            student.first_name
                                        }}
                                        {{
                                            student.paterno || student.last_name
                                        }}
                                    </div>
                                    <div class="student-details">
                                        {{ student.email }}
                                        <span
                                            v-if="student.eneatype"
                                            class="eneatype-badge"
                                            >Eneatipo
                                            {{ student.eneatype }}</span
                                        >
                                    </div>
                                </div>
                            </div>

                            <div
                                v-if="group.students.length === 0"
                                class="empty-group"
                            >
                                Arrastra estudiantes aquí
                            </div>
                        </div>

                        <!-- Score inline + botón Ver Detalle -->
                        <div
                            v-if="group.students.length > 0"
                            class="group-score-summary"
                        >
                            <!-- Pills de criterios: solo cuando hay providerScores -->
                            <div
                                v-if="
                                    group.providerScores &&
                                    Object.keys(group.providerScores).length
                                "
                                class="group-score-bar-row"
                            >
                                <div
                                    class="group-score-circle"
                                    :class="
                                        getAffinityClass(
                                            group.affinityScore ?? 0,
                                        )
                                    "
                                >
                                    {{
                                        (
                                            (group.affinityScore ?? 0) * 100
                                        ).toFixed(0)
                                    }}%
                                </div>
                                <div class="group-criteria-pills">
                                    <span
                                        v-for="(
                                            score, key
                                        ) in group.providerScores"
                                        :key="key"
                                        class="criterion-pill"
                                        :class="getScoreClass(score)"
                                    >
                                        {{ criterionLabel(key) }}:
                                        {{ (score * 100).toFixed(0) }}%
                                    </span>
                                </div>
                            </div>
                            <button
                                @click="openGroupDetail(group)"
                                class="detail-btn"
                            >
                                Ver Detalle
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Modal de Detalle del Grupo -->
        <div
            v-if="showDetailModal"
            class="modal-overlay"
            @click.self="showDetailModal = false"
        >
            <div class="modal-content detail-modal" @click.stop>
                <div class="modal-header">
                    <h3>Detalle: {{ detailGroup?.name }}</h3>
                    <button
                        @click="showDetailModal = false"
                        class="close-modal-btn"
                    >
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
                <div class="detail-body" v-if="detailGroup">
                    <!-- Score global -->
                    <div class="detail-global-score">
                        <div
                            class="affinity-circle large"
                            :class="
                                getAffinityClass(detailGroup.affinityScore ?? 0)
                            "
                        >
                            <span class="affinity-percentage">
                                {{
                                    (
                                        (detailGroup.affinityScore ?? 0) * 100
                                    ).toFixed(1)
                                }}%
                            </span>
                        </div>
                        <div class="detail-global-info">
                            <h4>Puntaje global del grupo</h4>
                        </div>
                    </div>

                    <!-- Criterios usados -->
                    <div class="detail-criteria-section">
                        <h4>Criterios utilizados</h4>
                        <table class="criteria-table">
                            <thead>
                                <tr>
                                    <th>Criterio</th>
                                    <th>Score</th>
                                    <th>Objetivo</th>
                                    <th>Peso</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr
                                    v-for="(
                                        score, key
                                    ) in detailGroup.providerScores"
                                    :key="key"
                                >
                                    <td class="ct-name">
                                        {{ criterionLabel(key) }}
                                    </td>

                                    <td class="ct-plain">
                                        {{ (score * 100).toFixed(1) }}%
                                    </td>

                                    <td class="ct-plain">
                                        <span
                                            v-if="
                                                detailGroup.criterionBreakdown &&
                                                detailGroup.criterionBreakdown[
                                                    key
                                                ]
                                            "
                                        >
                                            {{
                                                (
                                                    detailGroup
                                                        .criterionBreakdown[key]
                                                        .target * 100
                                                ).toFixed(0)
                                            }}%
                                        </span>
                                        <span
                                            v-else-if="
                                                detailGroup.targets &&
                                                detailGroup.targets[key] != null
                                            "
                                        >
                                            {{
                                                (
                                                    detailGroup.targets[key] *
                                                    100
                                                ).toFixed(0)
                                            }}%
                                        </span>
                                        <span v-else class="ct-no-data">—</span>
                                    </td>

                                    <td class="ct-plain">
                                        <span
                                            v-if="
                                                detailGroup.criterionBreakdown &&
                                                detailGroup.criterionBreakdown[
                                                    key
                                                ]
                                            "
                                        >
                                            {{
                                                (
                                                    detailGroup
                                                        .criterionBreakdown[key]
                                                        .weight * 100
                                                ).toFixed(0)
                                            }}%
                                        </span>
                                        <span
                                            v-else-if="
                                                detailGroupNormalizedWeights[
                                                    key
                                                ] != null
                                            "
                                        >
                                            {{
                                                (
                                                    detailGroupNormalizedWeights[
                                                        key
                                                    ] * 100
                                                ).toFixed(0)
                                            }}%
                                        </span>
                                        <span v-else class="ct-no-data">—</span>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>

                    <!-- Tabla de estudiantes por criterio -->
                    <div class="detail-students-section">
                        <h4>Datos de los estudiantes</h4>

                        <div
                            v-if="hasSocialAffinityCriterion"
                            class="group-social-graph-panel"
                        >
                            <div class="group-social-graph-head">
                                <h5>Sociograma del grupo (afinidad social)</h5>
                                <span class="group-social-graph-note"
                                    >Solo se consideran relaciones entre
                                    integrantes del grupo (Q1/Q3).</span
                                >
                            </div>

                            <div
                                v-if="groupSociogramLoading"
                                class="group-social-graph-state"
                            >
                                Cargando sociograma de la seccion...
                            </div>
                            <div
                                v-else-if="groupSociogramError"
                                class="group-social-graph-state error"
                            >
                                {{ groupSociogramError }}
                            </div>
                            <div
                                v-else-if="groupSocialGraphNodes.length === 0"
                                class="group-social-graph-state"
                            >
                                No hay integrantes para graficar.
                            </div>
                            <div
                                v-else-if="groupSocialGraphEdges.length === 0"
                                class="group-social-graph-state"
                            >
                                No hay relaciones de afinidad social entre los
                                integrantes del grupo.
                            </div>
                            <template v-else>
                                <svg
                                    class="group-social-graph-svg"
                                    viewBox="0 0 760 340"
                                    aria-label="Sociograma del grupo"
                                >
                                    <defs>
                                        <marker
                                            id="group-arrow-choice"
                                            markerWidth="10"
                                            markerHeight="10"
                                            refX="9"
                                            refY="3"
                                            orient="auto"
                                        >
                                            <path
                                                d="M0,0 L0,6 L9,3 z"
                                                fill="#1d4ed8"
                                            />
                                        </marker>
                                        <marker
                                            id="group-arrow-rejection"
                                            markerWidth="10"
                                            markerHeight="10"
                                            refX="9"
                                            refY="3"
                                            orient="auto"
                                        >
                                            <path
                                                d="M0,0 L0,6 L9,3 z"
                                                fill="#dc2626"
                                            />
                                        </marker>
                                    </defs>

                                    <g>
                                        <path
                                            v-for="edge in groupSocialGraphEdges"
                                            :key="edge.id"
                                            :d="edge.path"
                                            :stroke="edge.color"
                                            :stroke-width="edge.strokeWidth"
                                            fill="none"
                                            :marker-end="edge.marker"
                                            :opacity="edge.opacity"
                                        >
                                            <title>{{ edge.tooltip }}</title>
                                        </path>
                                    </g>

                                    <g>
                                        <g
                                            v-for="node in groupSocialGraphNodes"
                                            :key="node.id"
                                            class="group-social-node"
                                        >
                                            <circle
                                                :cx="node.x"
                                                :cy="node.y"
                                                :r="node.radius"
                                            />
                                            <text
                                                :x="node.x"
                                                :y="node.y + 4"
                                                text-anchor="middle"
                                            >
                                                {{ node.shortName }}
                                            </text>
                                            <title>{{ node.fullName }}</title>
                                        </g>
                                    </g>
                                </svg>

                                <div class="group-social-legend">
                                    <span class="legend-item">
                                        <i class="legend-line choice"></i>
                                        Eleccion (Q1)
                                    </span>
                                    <span class="legend-item">
                                        <i class="legend-line rejection"></i>
                                        Rechazo (Q3)
                                    </span>
                                </div>
                            </template>
                        </div>

                        <div class="detail-students-table-wrap">
                            <table class="students-criteria-table">
                                <thead>
                                    <tr>
                                        <th class="sct-name-col">Nombre</th>
                                        <th
                                            v-for="key in detailGroupNonRelationalProviderKeys"
                                            :key="key"
                                        >
                                            {{ criterionLabel(key) }}
                                        </th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr
                                        v-for="student in detailGroup.students"
                                        :key="student.id"
                                    >
                                        <td class="sct-name-col">
                                            {{
                                                student.nombres ||
                                                student.first_name
                                            }}
                                            {{
                                                student.paterno ||
                                                student.last_name
                                            }}
                                        </td>
                                        <td
                                            v-for="attr in getStudentCriteriaAttrs(
                                                student,
                                                detailGroup,
                                            )"
                                            :key="attr.key"
                                            class="sct-val"
                                        >
                                            {{ attr.value || "—" }}
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Modal del Test de Likert (para bfi44, mbti, enneagram, chronotype) -->
        <LikertTest
            :is-open="showLikertModal"
            :questionnaire="selectedQuestionnaire"
            :metadata="selectedQuestionnaireMetadata"
            @close="closeAnyModal"
            @submitted="handleTestCompleted"
            @view-results="handleViewResults"
        />

        <!-- Modal del Test de Selección (para group_compatibility) -->
        <SelectionTest
            :is-open="showSelectionModal"
            :questionnaire="selectedQuestionnaire"
            :student-id="currentStudentId"
            :section-id="sectionId"
            @close="closeAnyModal"
            @submitted="handleTestCompleted"
        />

        <!-- Modal del Test de Formulario (para demographic) -->
        <FormTest
            :is-open="showFormModal"
            :questionnaire="selectedQuestionnaire"
            :student-id="currentStudentId"
            :section-id="sectionId"
            @close="closeAnyModal"
            @submitted="handleTestCompleted"
        />
    </div>

    <!-- Modal para configurar tamaño de grupos -->
    <GroupSizeModal
        :show="showGroupSizeModal"
        @confirm="handleGroupSizeConfirm"
        @close="showGroupSizeModal = false"
    />

    <!-- Modal para formar grupos por plugins -->
    <FormGroupsByPluginModal
        :show="showFormByPluginModal"
        :students="availableStudents"
        :student-test-status="studentTestStatus"
        :questionnaires="questionnaires"
        @confirm="handleFormByPluginConfirm"
        @close="showFormByPluginModal = false"
    />
</template>

<script setup>
import { computed, ref } from "vue";
import LikertTest from "~/components/generic/LikertTest.vue";
import SelectionTest from "~/components/generic/SelectionTest.vue";
import FormTest from "~/components/generic/FormTest.vue";
import GroupSizeModal from "~/components/GroupSizeModal.vue";
import FormGroupsByPluginModal from "~/components/FormGroupsByPluginModal.vue";
import { useDatetimeInput } from "~/composables/useDatetimeInput";
import { usePluginApi } from "~/composables/usePluginApi";
import { useAuth } from "~/composables/useAuth";

const props = defineProps({
    sectionId: [String, Number],
    questionnaires: Array,
    loadingQuestionnaires: Boolean,
    availableStudents: Array,
    groups: Array,
    globalScore: Object,
    studentTestStatus: Object,
    openPluginModalTrigger: Number,
    readOnly: Boolean,
    studentId: [String, Number],
});

const { user } = useAuth();

const currentStudentId = computed(() => {
    return props.studentId?.toString() || user.value?.id?.toString() || null;
});

const emit = defineEmits([
    "create-questionnaire",
    "edit-questionnaire",
    "view-responses",
    "delete-questionnaire",
    "add-group",
    "delete-all-groups",
    "delete-group",
    "update-group-name",
    "save-group-name",
    "form-groups-by-plugin",
    "form-random-groups",
    "drag-start",
    "drop-to-group",
    "drop-to-available",
    "drag-over",
    "open-plugin-modal",
    "export-groups",
]);

const { formatDatetimeDisplay } = useDatetimeInput();
const pluginApi = usePluginApi();

const formatDate = (date) => {
    if (!date || date === null || date === undefined) return "No definida";

    try {
        const formatted = formatDatetimeDisplay(date);
        if (!formatted) return "Fecha inválida";
        return formatted;
    } catch (error) {
        console.error("Error formatting date:", error, "Original date:", date);
        return "Fecha inválida";
    }
};

// Estado para los modales de test genéricos
const showLikertModal = ref(false);
const showSelectionModal = ref(false);
const showFormModal = ref(false);
const selectedQuestionnaire = ref(null);
const selectedQuestionnaireMetadata = ref(null);

// Estado para el modal de configuración de grupos aleatorios
const showGroupSizeModal = ref(false);

// Estado para el modal de formar grupos por plugins
const showFormByPluginModal = ref(false);

// Watch for external trigger to open plugin modal (used to coordinate data loading)
watch(() => props.openPluginModalTrigger, (newVal) => {
    if (newVal > 0) {
        showFormByPluginModal.value = true;
    }
});

const addNewGroup = () => {
    emit("add-group");
};

const deleteAllGroups = () => {
    emit("delete-all-groups");
};

// Get plugin label helper
const getPluginLabel = (questionnaire) => {
    if (questionnaire.plugin === "group_compatibility") {
        return "Afinidad Social";
    } else if (questionnaire.plugin === "demographic") {
        return "Dato Demográfico";
    }
    return "Afinidad Psicológica";
};

const openFormByPluginModal = () => {
    emit("open-plugin-modal");
};

const handleFormByPluginConfirm = (formConfig) => {
    // Transform new modal format to backend format
    const backendConfig = {
        teamSize: formConfig.teamSize,
        minTeamSize: formConfig.teamSize - 1,
        maxTeamSize: formConfig.teamSize + 1,
        eligibleStudentIds:
            formConfig.eligibleStudentIds && formConfig.eligibleStudentIds.length > 0
                ? formConfig.eligibleStudentIds
                : props.availableStudents.map((s) => s.id.toString()),
        additionalCriteria: {},
        enforceEquity: !!formConfig.enforceEquity,
    };

    // Add criteria from the modal
    if (formConfig.criteria && formConfig.criteria.length > 0) {
        formConfig.criteria.forEach((criterion) => {
            backendConfig.additionalCriteria[criterion.id] = {
                target: criterion.target,
                weight: criterion.weight,
                type: criterion.type || "",
                label: criterion.label || criterion.id,
                goal: criterion.goal || "",
            };
        });
    }

    emit("form-groups-by-plugin", backendConfig);
};

const formRandomGroups = () => {
    showGroupSizeModal.value = true;
};

const handleGroupSizeConfirm = (studentsPerGroup) => {
    emit("form-random-groups", studentsPerGroup);
};

const deleteGroup = (groupId) => {
    emit("delete-group", groupId);
};

const exportToExcel = () => {
    emit("export-groups");
};

const saveGroupName = (group) => {
    emit("save-group-name", group);
};

const getQuestionnaireStatus = (questionnaire) => {
    const now = new Date();
    const startDate = questionnaire.start_date
        ? new Date(questionnaire.start_date.replace("Z", ""))
        : null;
    const endDate = questionnaire.end_date
        ? new Date(questionnaire.end_date.replace("Z", ""))
        : null;

    if (!startDate || !endDate) return "draft";

    if (now >= startDate && now <= endDate) return "active";
    if (now > endDate) return "expired";
    return "draft";
};

// Drag and Drop functions
const onDragStart = (event, student, sourceType, sourceId = null) => {
    emit("drag-start", event, student, sourceType, sourceId);
};

const onDragOver = (event) => {
    event.preventDefault();
};

const onDropToGroup = (event, groupId) => {
    event.preventDefault();
    emit("drop-to-group", event, groupId);
};

const onDropToAvailable = (event) => {
    event.preventDefault();
    emit("drop-to-available", event);
};

// ── Modal de Detalle del Grupo ───────────────────────────────────────────────
const showDetailModal = ref(false);
const detailGroup = ref(null);
const sectionSociogramResponses = ref([]);
const loadedSociogramSectionId = ref(null);
const groupSociogramLoading = ref(false);
const groupSociogramError = ref("");

const openGroupDetail = async (group) => {
    detailGroup.value = group;
    showDetailModal.value = true;

    if (hasSocialAffinityCriterion.value) {
        await ensureSectionSociogramResponses();
    }
};

const hasSocialAffinityCriterion = computed(() => {
    const providers = detailGroup.value?.providerScores || {};
    return (
        providers.social_affinity != null ||
        providers.group_compatibility != null
    );
});

const ensureSectionSociogramResponses = async () => {
    const sectionId = String(props.sectionId || "");
    if (!sectionId) {
        groupSociogramError.value = "No se encontro sectionId para cargar sociograma.";
        return;
    }

    if (loadedSociogramSectionId.value === sectionId) {
        return;
    }

    groupSociogramLoading.value = true;
    groupSociogramError.value = "";

    try {
        const { success, data, error } = await pluginApi.getSectionResponses("group_compatibility", sectionId);

        if (!success) {
            throw new Error(error || "No se pudieron obtener respuestas sociometricas.");
        }

        const raw = data || [];
        sectionSociogramResponses.value = Array.isArray(raw) ? raw : [];
        loadedSociogramSectionId.value = sectionId;
    } catch (error) {
        console.error("Error loading section sociogram responses:", error);
        groupSociogramError.value =
            "No fue posible cargar las respuestas de afinidad social.";
        sectionSociogramResponses.value = [];
    } finally {
        groupSociogramLoading.value = false;
    }
};

const groupMemberIds = computed(() => {
    const members = detailGroup.value?.students || [];
    return members.map((s) => String(s.id));
});

const latestResponsesByStudent = computed(() => {
    const memberSet = new Set(groupMemberIds.value);
    const map = new Map();

    sectionSociogramResponses.value.forEach((resp) => {
        const sid = String(resp?.student_id || "");
        if (!sid || !memberSet.has(sid)) return;

        const current = map.get(sid);
        const currentTs = current
            ? new Date(
                  current.submitted_at ||
                      current.updated_at ||
                      current.created_at ||
                      0,
              ).getTime()
            : 0;
        const incomingTs = new Date(
            resp.submitted_at || resp.updated_at || resp.created_at || 0,
        ).getTime();

        if (!current || incomingTs >= currentTs) {
            map.set(sid, resp);
        }
    });

    return map;
});

const groupSocialGraphNodes = computed(() => {
    const members = detailGroup.value?.students || [];
    const total = members.length;
    if (!total) return [];

    const cx = 380;
    const cy = 170;
    const orbitRadius = Math.max(80, Math.min(130, 75 + total * 8));

    return members.map((student, idx) => {
        const angle = (Math.PI * 2 * idx) / total - Math.PI / 2;
        const x = cx + orbitRadius * Math.cos(angle);
        const y = cy + orbitRadius * Math.sin(angle);
        const first = student.nombres || student.first_name || "";
        const last = student.paterno || student.last_name || "";
        const fullName = `${first} ${last}`.trim() || `ID ${student.id}`;
        const nameParts = fullName.split(" ").filter(Boolean);
        const shortName = nameParts.length >= 2
            ? nameParts.slice(0, 2).map((n) => n.charAt(0)).join("").toUpperCase()
            : (nameParts[0] || `ID ${student.id}`).substring(0, 2).toUpperCase();
        return {
            id: String(student.id),
            x,
            y,
            radius: 17,
            shortName,
            fullName,
        };
    });
});

const groupSocialGraphNodeMap = computed(() => {
    const map = {};
    groupSocialGraphNodes.value.forEach((node) => {
        map[node.id] = node;
    });
    return map;
});

const groupSocialEdgesRaw = computed(() => {
    const memberSet = new Set(groupMemberIds.value);
    const edges = [];

    latestResponsesByStudent.value.forEach((resp, sourceId) => {
        const processAnswers = (answers, type, expected, question) => {
            if (!Array.isArray(answers)) return;
            answers.forEach((ans) => {
                const targetId = String(ans?.student_id || "");
                if (!targetId || targetId === sourceId) return;
                if (!memberSet.has(targetId)) return;
                if (ans.response !== expected) return;
                edges.push({
                    source: sourceId,
                    target: targetId,
                    type,
                    question,
                });
            });
        };

        processAnswers(resp.question1, "choice", 1, "Q1");
        processAnswers(resp.question3, "rejection", 1, "Q3");
    });

    return edges;
});

const groupSocialGraphEdges = computed(() => {
    const nodeMap = groupSocialGraphNodeMap.value;
    const edges = groupSocialEdgesRaw.value;
    const reverseIndex = new Set(edges.map((e) => `${e.target}|${e.source}|${e.type}`));

    return edges
        .map((edge, idx) => {
            const source = nodeMap[edge.source];
            const target = nodeMap[edge.target];
            if (!source || !target) return null;

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
            if (reverseIndex.has(`${edge.source}|${edge.target}|${edge.type}`)) {
                const mx = (startX + endX) / 2;
                const my = (startY + endY) / 2;
                const nx = -uy;
                const ny = ux;
                const curve = edge.source < edge.target ? 13 : -13;
                path = `M ${startX} ${startY} Q ${mx + nx * curve} ${my + ny * curve} ${endX} ${endY}`;
            }

            const isChoice = edge.type === "choice";
            const sourceName = nodeMap[edge.source]?.fullName || edge.source;
            const targetName = nodeMap[edge.target]?.fullName || edge.target;

            return {
                id: `${edge.source}-${edge.target}-${edge.type}-${idx}`,
                path,
                color: isChoice ? "#1d4ed8" : "#dc2626",
                strokeWidth: 2.2,
                marker: isChoice
                    ? "url(#group-arrow-choice)"
                    : "url(#group-arrow-rejection)",
                opacity: 0.84,
                tooltip: `${edge.question}: ${sourceName} -> ${targetName}`,
            };
        })
        .filter(Boolean);
});

// Etiquetas legibles para cada criterio (provider ID → texto UI)
const CRITERION_LABELS = {
    career_code: "Carrera",
    enneagram: "Eneagrama",
    bfi44: "BFI-44",
    mbti: "MBTI",
    social_affinity: "Afinidad Social",
    group_compatibility: "Afinidad Social",
    sexo: "Género",
    sex: "Género",
    gender: "Género",
    conscientiousness: "Responsabilidad",
    extraversion: "Sociabilidad",
    agreeableness: "Amabilidad",
    openness: "Apertura",
    neuroticism: "Est. Emocional",
};

const criterionLabel = (key) => {
    if (CRITERION_LABELS[key]) return CRITERION_LABELS[key];
    // Intenta humanizar claves demográficas tipo "demographic_<id>_<slug>"
    if (key.startsWith("demographic_")) {
        const parts = key.split("_");
        // Toma la parte después del id (índice 2 en adelante)
        return parts.slice(2).join(" ").replace(/_/g, " ");
    }
    return key.replace(/_/g, " ");
};

/**
 * Para cada criterio activo del grupo devuelve el valor del estudiante
 * y un status (ok / fail / neutral) para colorearlo.
 */
const getStudentCriteriaAttrs = (student, group) => {
    if (!group.providerScores) return [];
    const attrs = [];

    for (const key of Object.keys(group.providerScores)) {
        if (RELATIONAL_KEYS.has(key)) continue;
        const groupScore = group.providerScores[key];
        let value = null;
        let status = "neutral";
        let tooltip = "";

        // Con la semántica de Error Absoluto, groupScore = 1 − |target − real|.
        // "Preciso" cuando score ≥ 0.80 (diferencia ≤ 0.20 respecto al objetivo).
        const isPrecise = criterionMet(groupScore);
        const scorePct = Math.round(groupScore * 100);

        // Recuperar diversidad real desde breakdown si está disponible
        const breakdown = group.criterionBreakdown?.[key];
        const realPct = breakdown
            ? Math.round(breakdown.real_value * 100)
            : scorePct;
        const targetPct = breakdown ? Math.round(breakdown.target * 100) : null;

        // OCEAN keys: conscientiousness, extraversion, agreeableness, openness, neuroticism
        const OCEAN_KEYS = new Set([
            "openness",
            "conscientiousness",
            "extraversion",
            "agreeableness",
            "neuroticism",
        ]);

        if (key === "career_code") {
            value =
                student.career_code ||
                student.carrera ||
                student.codigo_carrera ||
                "—";
            const allCareers = group.students.map(
                (s) => s.career_code || s.carrera || s.codigo_carrera || "",
            );
            const unique = new Set(allCareers.filter(Boolean));
            status = isPrecise ? "attr-ok" : "attr-fail";
            const careerList =
                unique.size <= 1
                    ? "todos del mismo código"
                    : [...unique].join(", ");
            tooltip =
                targetPct != null
                    ? `Carreras: ${careerList} | Real: ${realPct}% · Objetivo: ${targetPct}% · Score: ${scorePct}%`
                    : `Carreras: ${careerList} (score: ${scorePct}%)`;
        } else if (key === "sexo" || key === "sex" || key === "gender") {
            value = student.sexo || student.sex || student.gender || "—";
            const allSex = group.students.map(
                (s) => s.sexo || s.sex || s.gender || "",
            );
            const unique = new Set(allSex.filter(Boolean));
            status = isPrecise ? "attr-ok" : "attr-fail";
            const genderList =
                unique.size <= 1 ? "género homogéneo" : [...unique].join(", ");
            tooltip =
                targetPct != null
                    ? `Géneros: ${genderList} | Real: ${realPct}% · Objetivo: ${targetPct}% · Score: ${scorePct}%`
                    : `Géneros: ${genderList} (score: ${scorePct}%)`;
        } else if (key === "enneagram" || key === "bfi44" || key === "mbti") {
            value = student.eneatype
                ? `E${student.eneatype}`
                : student.mbti_type || "Sin datos";
            status =
                value === "Sin datos"
                    ? "attr-missing"
                    : isPrecise
                      ? "attr-ok"
                      : "attr-fail";
            tooltip =
                value === "Sin datos"
                    ? "Este estudiante no completó el test"
                    : `Score de afinidad psicológica: ${scorePct}%${targetPct != null ? ` (objetivo: ${targetPct}%)` : ""}`;
        } else if (OCEAN_KEYS.has(key)) {
            // Read from ocean_vector populated during data loading
            const ocean = student.ocean_vector || {};
            const shortKey = key.charAt(0);
            const rawOceanVal =
                ocean[key] ??
                ocean[shortKey] ??
                ocean[shortKey.toUpperCase()] ??
                ocean[shortKey.toLowerCase()];
            const parsedOceanVal = Number(rawOceanVal);

            if (Number.isFinite(parsedOceanVal)) {
                value = parsedOceanVal.toFixed(3);
            } else {
                // Fallback: try direct field on student (legacy)
                const directVal = student[key];
                value =
                    directVal != null
                        ? Number(directVal).toFixed(3)
                        : "Sin datos";
            }
            status = value === "Sin datos" ? "attr-missing" : "neutral";
            tooltip =
                value === "Sin datos"
                    ? "Este estudiante no tiene datos OCEAN"
                    : targetPct != null
                    ? `${criterionLabel(key)}: ${value} (objetivo: ${targetPct}%)`
                    : `${criterionLabel(key)}: ${value}`;
        } else if (key.startsWith("demographic_")) {
            value = "Respondido";
            status = "attr-ok";
            tooltip = "Estudiante tiene datos demográficos";
        } else {
            value = "—";
            status = isPrecise ? "attr-ok" : "attr-fail";
            tooltip =
                targetPct != null
                    ? `${criterionLabel(key)} | Real: ${realPct}% · Objetivo: ${targetPct}% · Score: ${scorePct}%`
                    : `${criterionLabel(key)} (score: ${scorePct}%)`;
        }

        attrs.push({
            key,
            label: criterionLabel(key),
            value,
            status,
            tooltip,
        });
    }
    return attrs;
};

const getAffinityClass = (compatibility) => {
    if (compatibility >= 0.7) return "high-affinity";
    if (compatibility >= 0.5) return "medium-affinity";
    return "low-affinity";
};

/**
 * Clase CSS para el score de error absoluto (1 − |target − real|).
 * Verde ≥ 0.85: muy preciso | Amarillo ≥ 0.65: aceptable | Rojo < 0.65: impreciso.
 */
const getScoreClass = (score) => {
    if (score >= 0.85) return "score-precise";
    if (score >= 0.65) return "score-moderate";
    return "score-imprecise";
};

/**
 * Clase CSS para la fila completa: resalta suavemente filas imprecisas.
 */
const criterionRowClass = (score) => {
    if (!criterionMet(score)) return "ct-row-imprecise";
    return "";
};

/**
 * Devuelve los pesos del grupo activo normalizados (suma = 1.0) para mostrar
 * en la tabla de criterios el peso real que utilizó el algoritmo.
 */
const detailGroupNormalizedWeights = computed(() => {
    if (!detailGroup.value?.weights) return {};
    const weights = detailGroup.value.weights;
    const total = Object.values(weights).reduce(
        (sum, w) => sum + (w > 0 ? w : 0),
        0,
    );
    if (total === 0) return {};
    const normalized = {};
    for (const [key, w] of Object.entries(weights)) {
        normalized[key] = w > 0 ? w / total : 0;
    }
    return normalized;
});

const RELATIONAL_KEYS = new Set(["social", "social_affinity", "sociogram", "relational"]);

const detailGroupNonRelationalProviderKeys = computed(() => {
    const providers = detailGroup.value?.providerScores || {};
    return Object.keys(providers).filter((key) => !RELATIONAL_KEYS.has(key));
});

/**
 * Con la nueva semántica de Error Absoluto:
 *   score = 1 − |target − real|
 * Un score alto significa que la diversidad real está cerca del objetivo.
 * Consideramos "cumplido" cuando el score ≥ 0.80 (la diferencia |target−real| ≤ 0.20).
 */
const criterionMet = (score) => {
    return score >= 0.8;
};

// Funciones para los modales de test genéricos
const openEnneagramTest = async (questionnaire) => {
    console.log("Abriendo test:", questionnaire);
    selectedQuestionnaire.value = questionnaire;

    if (questionnaire.plugin === "group_compatibility") {
        showSelectionModal.value = true;
    } else if (questionnaire.plugin === "demographic") {
        showFormModal.value = true;
    } else {
        // Likert-type plugins: bfi44, mbti, enneagram, chronotype, bigfive
        const likertTypes = ["bfi44", "mbti", "enneagram", "chronotype", "bigfive"];
        if (likertTypes.includes(questionnaire.plugin)) {
            const pluginToLoad = questionnaire.plugin === "bigfive" ? "bfi44" : questionnaire.plugin;
            const pluginApi = usePluginApi();
            const result = await pluginApi.getPluginsExtended();
            if (result.success && result.data) {
                const pluginsList = Array.isArray(result.data) ? result.data : [result.data];
                const metadata = pluginsList.find(p => p.nombre_interno === pluginToLoad);
                selectedQuestionnaireMetadata.value = metadata;
            }
            showLikertModal.value = true;
        }
    }
};

const closeAnyModal = () => {
    console.log("Cerrando modal del test");
    showLikertModal.value = false;
    showSelectionModal.value = false;
    showFormModal.value = false;
    selectedQuestionnaire.value = null;
    selectedQuestionnaireMetadata.value = null;
};

const handleTestCompleted = (result) => {
    console.log("Test completado:", result);
    alert("¡Test completado exitosamente!");
};

const handleViewResults = (results) => {
    console.log("Ver resultados:", results);
};
</script>

<style scoped>
.groups-content {
    display: flex;
    flex-direction: column;
    gap: 2rem;
}

/* Coordinación de Cuestionarios */
/* ── Score inline en tarjeta de grupo ───────────────────────────── */
.group-score-summary {
    margin-top: 0.75rem;
    padding: 0.75rem;
    background: var(--color-background, #f4f6f9);
    border-radius: var(--radius-sm, 6px);
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.group-score-bar-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
}

.group-score-circle {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    font-weight: 700;
    color: white;
    flex-shrink: 0;
}

.group-score-circle.high-affinity {
    background: #38a169;
}
.group-score-circle.medium-affinity {
    background: #dd6b20;
}
.group-score-circle.low-affinity {
    background: #e53e3e;
}

.group-criteria-pills {
    display: flex;
    flex-wrap: wrap;
    gap: 0.375rem;
}

.criterion-pill {
    font-size: 0.7rem;
    font-weight: 600;
    padding: 0.2rem 0.5rem;
    border-radius: 999px;
    white-space: nowrap;
}

.criterion-pill.high-affinity,
.criterion-pill.score-precise {
    background: #dcfce7;
    color: #166534;
}
.criterion-pill.medium-affinity,
.criterion-pill.score-moderate {
    background: #fef3c7;
    color: #92400e;
}
.criterion-pill.low-affinity,
.criterion-pill.score-imprecise {
    background: #fee2e2;
    color: #991b1b;
}

.detail-btn {
    align-self: flex-end;
    padding: 0.375rem 0.875rem;
    background: var(--color-primary, #00a499);
    color: white;
    border: none;
    border-radius: var(--radius-sm, 6px);
    font-size: 0.8125rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
}
.detail-btn:hover {
    background: #008a7f;
}

/* ── Modal de Detalle ────────────────────────────────────────────── */
.detail-modal {
    max-width: 860px;
    width: 96vw;
    max-height: 92vh;
    overflow-y: auto;
    border-radius: var(--radius-lg, 14px);
    background: #fff;
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
}

.detail-body {
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
}

.detail-global-score {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem;
    background: var(--color-background, #f4f6f9);
    border-radius: var(--radius-md, 10px);
}

.affinity-circle {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.1rem;
    font-weight: 700;
    color: white;
}

.affinity-circle.high-affinity {
    background: #38a169;
}
.affinity-circle.medium-affinity {
    background: #dd6b20;
}
.affinity-circle.low-affinity {
    background: #e53e3e;
}

.affinity-circle.large {
    width: 72px;
    height: 72px;
    font-size: 1rem;
}

.detail-global-info h4 {
    margin: 0 0 0.25rem;
    font-size: 1rem;
    font-weight: 700;
    color: var(--color-text, #1a202c);
}

.detail-algorithm {
    margin: 0;
    font-size: 0.8125rem;
    color: var(--color-text-muted, #718096);
}

.detail-criteria-section h4,
.detail-students-section h4 {
    margin: 0 0 0.75rem;
    font-size: 0.9375rem;
    font-weight: 700;
    color: var(--color-text, #1a202c);
}

.group-social-graph-panel {
    border: 1px solid #dce2ea;
    border-radius: 10px;
    padding: 0.75rem;
    margin-bottom: 0.9rem;
    background: linear-gradient(180deg, #f8fbff 0%, #ffffff 100%);
}

.group-social-graph-head {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    margin-bottom: 0.55rem;
}

.group-social-graph-head h5 {
    margin: 0;
    font-size: 0.88rem;
    color: #1f2937;
}

.group-social-graph-note {
    font-size: 0.72rem;
    color: #64748b;
}

.group-social-graph-state {
    font-size: 0.8rem;
    color: #475569;
    padding: 0.55rem 0.4rem;
}

.group-social-graph-state.error {
    color: #b91c1c;
}

.group-social-graph-svg {
    width: 100%;
    min-height: 240px;
    border: 1px solid #dbe2eb;
    border-radius: 8px;
    background: #ffffff;
}

.group-social-node circle {
    fill: #111827;
    stroke: #ffffff;
    stroke-width: 2;
}

.group-social-node text {
    fill: #ffffff;
    font-size: 9px;
    font-weight: 700;
    pointer-events: none;
}

.group-social-legend {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem 0.9rem;
    margin-top: 0.5rem;
}

.group-social-legend .legend-item {
    display: inline-flex;
    align-items: center;
    gap: 0.28rem;
    font-size: 0.7rem;
    color: #475569;
}

.group-social-legend .legend-line {
    width: 16px;
    height: 0;
    border-top-width: 2px;
    border-top-style: solid;
}

.group-social-legend .legend-line.choice {
    border-top-color: #1d4ed8;
}

.group-social-legend .legend-line.rejection {
    border-top-color: #dc2626;
}

/* ── Tabla de criterios ─────────────────────────────────────────── */
.criteria-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.8125rem;
}

.criteria-table thead tr {
    background: var(--color-background, #f4f6f9);
}

.criteria-table th {
    padding: 0.5rem 0.75rem;
    text-align: left;
    font-weight: 600;
    font-size: 0.72rem;
    color: #6b7280;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    border-bottom: 2px solid #e5e7eb;
    white-space: nowrap;
}

.criteria-table td {
    padding: 0.6rem 0.75rem;
    border-bottom: 1px solid #f0f0f0;
    vertical-align: middle;
}

.criteria-table tbody tr:last-child td {
    border-bottom: none;
}

.criteria-table tbody tr:hover {
    background: #fafafa;
}

.ct-name {
    font-weight: 600;
    color: #1f2937;
    white-space: nowrap;
}

.ct-plain {
    color: #374151;
    font-size: 0.8125rem;
    white-space: nowrap;
}

.ct-no-data {
    color: #9ca3af;
}

/* ── Fin tabla de criterios ─────────────────────────────────────── */

/* ── Tabla de estudiantes por criterio ──────────────────────────── */
.detail-students-table-wrap {
    overflow-x: auto;
}

.students-criteria-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.8125rem;
}

.students-criteria-table thead tr {
    background: var(--color-background, #f4f6f9);
}

.students-criteria-table th {
    padding: 0.5rem 0.75rem;
    text-align: left;
    font-weight: 600;
    font-size: 0.72rem;
    color: #6b7280;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    border-bottom: 2px solid #e5e7eb;
    white-space: nowrap;
}

.students-criteria-table td {
    padding: 0.55rem 0.75rem;
    border-bottom: 1px solid #f0f0f0;
    color: #374151;
    vertical-align: middle;
}

.students-criteria-table tbody tr:last-child td {
    border-bottom: none;
}

.students-criteria-table tbody tr:hover {
    background: #fafafa;
}

.sct-name-col {
    font-weight: 600;
    color: #1f2937 !important;
    white-space: nowrap;
}

.sct-val {
    white-space: nowrap;
    text-align: center;
}

.questionnaires-section {
    background: var(--color-surface, #ffffff);
    border-radius: var(--radius-md, 10px);
    box-shadow: var(--shadow-sm, 0 1px 3px rgba(0, 0, 0, 0.08));
}

.questionnaires-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
}

.questionnaires-header h3 {
    margin: 0;
    color: var(--color-dark, #394049);
    font-size: 1.5rem;
}

.add-questionnaire-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 2rem;
    background-color: var(--color-primary, #00a499);
    color: white;
    border: none;
    border-radius: var(--radius-sm, 6px);
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    transition: all 0.2s;
}

.add-questionnaire-btn:hover {
    background-color: var(--color-primary-dark, #008a7f);
    box-shadow: 0 4px 12px rgba(0, 164, 153, 0.3);
}

.loading-questionnaires {
    text-align: center;
    padding: 2rem;
    color: #94a3b8;
}

.no-questionnaires {
    text-align: center;
    padding: 2rem;
    color: #94a3b8;
    border: 2px dashed var(--color-border, #dfe4ea);
    border-radius: var(--radius-md, 10px);
}

.questionnaires-container {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1rem;
}

.questionnaire-card {
    background: var(--color-surface, #ffffff);
    border: 1px solid var(--color-border, #dfe4ea);
    border-radius: var(--radius-md, 10px);
    padding: 1.5rem;
    transition: all 0.2s;
    position: relative;
    overflow: hidden;
}

.questionnaire-card:hover {
    box-shadow: var(--shadow-md, 0 4px 6px rgba(0, 0, 0, 0.08));
    transform: translateY(-1px);
}

.questionnaire-card::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 4px;
}

.questionnaire-active::before {
    background: var(--color-primary, #00a499);
}

.questionnaire-expired::before {
    background: var(--color-danger, #d7373f);
}

.questionnaire-draft::before {
    background: #94a3b8;
}

.questionnaire-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1rem;
}

.questionnaire-title {
    margin: 0;
    color: var(--color-dark, #394049);
    font-size: 1.1rem;
    font-weight: 600;
    flex: 1;
}

.questionnaire-plugin {
    background: var(--color-primary, #006b64);
    color: white;
    font-size: 0.75rem;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-weight: 500;
}

.questionnaire-info {
    margin-bottom: 1rem;
}

.questionnaire-description {
    margin: 0 0 0.75rem 0;
    color: var(--color-text-secondary, #64748b);
    font-size: 0.9rem;
    line-height: 1.4;
}

.questionnaire-dates {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.5rem;
}

.date-info {
    color: var(--color-text-secondary, #64748b);
    font-size: 0.75rem;
}

.questionnaire-actions {
    display: flex;
    gap: 0.5rem;
    justify-content: flex-end;
    flex-wrap: wrap;
}

.action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 36px;
    padding: 0 8px;
    border: none;
    border-radius: var(--radius-sm, 6px);
    cursor: pointer;
    transition: all 0.2s;
    gap: 4px;
    font-size: 0.8rem;
    font-weight: 500;
    white-space: nowrap;
    flex: 1;
    min-width: 0;
}

.action-btn svg {
    flex-shrink: 0;
}

.edit-btn {
    background-color: var(--color-primary, #00a499);
    color: white;
}

.edit-btn:hover {
    background-color: var(--color-primary-dark, #008a7f);
}

.view-btn {
    background-color: var(--color-accent, #ea7600);
    color: white;
}

.view-btn:hover {
    background-color: #d06a00;
}

.delete-btn {
    background-color: var(--color-danger, #d7373f);
    color: white;
}

.delete-btn:hover {
    background-color: #b82e35;
}

.btn-text {
    display: inline;
}

/* Responsive Design */
@media (max-width: 768px) {
    .questionnaire-actions {
        gap: 0.25rem;
        justify-content: center;
    }

    .action-btn {
        height: 32px;
        padding: 0 6px;
        font-size: 0.75rem;
        gap: 2px;
        min-width: auto;
        flex: none;
    }

    .btn-text {
        display: none;
    }

    .action-btn svg {
        width: 14px;
        height: 14px;
    }
}

@media (max-width: 480px) {
    .questionnaire-actions {
        flex-direction: column;
        align-items: stretch;
        gap: 0.5rem;
    }

    .action-btn {
        width: 100%;
        justify-content: center;
        padding: 8px 12px;
    }

    .btn-text {
        display: inline;
    }
}
.available-students-panel {
    background: var(--color-surface, #ffffff);
    border-radius: var(--radius-md, 10px);
    box-shadow: var(--shadow-sm, 0 1px 3px rgba(0, 0, 0, 0.08));
}

.available-students-panel h3 {
    margin: 0 0 1.5rem 0;
    color: var(--color-dark, #394049);
    font-size: 1.5rem;
}

.grouping-buttons {
    display: flex;
    gap: 1rem;
    margin-bottom: 1.5rem;
    flex-wrap: wrap;
}

.grouping-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    border: none;
    border-radius: var(--radius-sm, 6px);
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    transition: all 0.2s;
    min-width: 200px;
    justify-content: center;
}

.grouping-btn:disabled {
    background-color: var(--color-border, #dfe4ea);
    color: #a0aec0;
    cursor: not-allowed;
    opacity: 0.6;
}

.grouping-btn:disabled:hover {
    background-color: var(--color-border, #dfe4ea);
}

.enneagram-btn {
    background-color: var(--color-primary, #00a499);
    color: white;
}

.enneagram-btn:hover {
    background-color: var(--color-primary-dark, #008a7f);
}

.plugin-btn {
    background-color: var(--color-accent, #ea7600);
    color: white;
}

.plugin-btn:hover {
    background-color: #d06a00;
}

.random-btn {
    background-color: var(--color-dark, #394049);
    color: white;
}

.random-btn:hover {
    background-color: #2d3139;
}

.students-pool {
    min-height: 200px;
    border: 2px dashed var(--color-border, #dfe4ea);
    border-radius: var(--radius-md, 10px);
    padding: 1rem;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 1rem;
}

/* Panel de grupos */
.groups-panel {
    background: var(--color-surface, #ffffff);
    border-radius: var(--radius-md, 10px);
    box-shadow: var(--shadow-sm, 0 1px 3px rgba(0, 0, 0, 0.08));
}

.groups-panel h3 {
    margin: 0 0 1.5rem 0;
    color: var(--color-dark, #394049);
    font-size: 1.5rem;
}

.groups-header {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    margin-bottom: 1.5rem;
}

.global-score-badge {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.45rem 0.9rem;
    background: var(--color-background, #f4f6f9);
    border: 1px solid #e5e7eb;
    border-radius: var(--radius-sm, 6px);
    font-size: 0.8125rem;
    margin-right: auto;
}

.global-score-label {
    color: #6b7280;
    font-weight: 500;
}

.global-score-value {
    font-weight: 700;
    color: var(--color-text, #1a202c);
}

.add-group-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    background-color: var(--color-primary, #00a499);
    color: white;
    border: none;
    border-radius: var(--radius-sm, 6px);
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    transition: all 0.2s;
}

.add-group-btn:hover {
    background-color: var(--color-primary-dark, #008a7f);
}

.delete-all-groups-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    background-color: var(--color-danger, #d7373f);
    color: white;
    border: none;
    border-radius: var(--radius-sm, 6px);
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    transition: all 0.2s;
    margin-left: 0.5rem;
}

.delete-all-groups-btn:hover:not(:disabled) {
    background-color: #b82e35;
}

.delete-all-groups-btn:disabled {
    background-color: #fce8e9;
    color: #374151;
    cursor: not-allowed;
}

.export-groups-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    background-color: var(--color-accent, #b55600);
    color: white;
    border: none;
    border-radius: var(--radius-sm, 6px);
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    transition: all 0.2s;
    margin-left: 0.5rem;
}

.export-groups-btn:hover:not(:disabled) {
    background-color: #8f4400;
}

.export-groups-btn:disabled {
    background-color: #fdebd0;
    color: #374151;
    cursor: not-allowed;
}

.groups-container {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 1.5rem;
}

.group-card {
    border: 1px solid var(--color-border, #dfe4ea);
    border-radius: var(--radius-md, 10px);
    overflow: hidden;
}

.group-header {
    background: var(--color-background, #f4f6f9);
    padding: 1rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.group-name-input {
    background: transparent;
    border: none;
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--color-dark, #394049);
    flex: 1;
    padding: 0.25rem;
    border-radius: 4px;
}

.group-name-input:focus {
    outline: none;
    background: var(--color-surface, #ffffff);
    box-shadow: 0 0 0 2px var(--color-primary, #00a499);
}

.delete-group-btn {
    background: var(--color-danger, #d7373f);
    color: white;
    border: none;
    border-radius: 4px;
    padding: 0.5rem;
    cursor: pointer;
    transition: background 0.2s;
}

.delete-group-btn:hover {
    background: #b82e35;
}

.group-students-area {
    min-height: 150px;
    padding: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.empty-group {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #94a3b8;
    font-style: italic;
    border: 2px dashed var(--color-border, #dfe4ea);
    border-radius: var(--radius-sm, 6px);
    padding: 2rem;
}

/* Estudiante cards */
.student-card {
    background: linear-gradient(
        135deg,
        var(--color-background, #f4f6f9) 0%,
        var(--color-surface, #ffffff) 100%
    );
    border: 2px solid var(--color-border, #dfe4ea);
    border-radius: var(--radius-md, 10px);
    padding: 0.875rem;
    cursor: move;
    transition: all 0.3s ease;
    display: flex;
    gap: 0.75rem;
    align-items: center;
    position: relative;
    overflow: hidden;
}

.student-card::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 3px;
    background: var(--color-primary, #00a499);
    opacity: 0;
    transition: opacity 0.3s ease;
}

.student-card:hover {
    background: linear-gradient(
        135deg,
        var(--color-surface, #ffffff) 0%,
        var(--color-background, #f4f6f9) 100%
    );
    box-shadow: var(--shadow-md, 0 6px 12px rgba(0, 0, 0, 0.08));
    transform: translateY(-2px);
    border-color: #cbd5e0;
}

.student-card:hover::before {
    opacity: 1;
}

.student-card:active {
    transform: scale(0.98);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.student-avatar {
    flex-shrink: 0;
    width: 48px;
    height: 48px;
    background: var(--color-primary, #006b64);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    box-shadow: 0 4px 6px rgba(0, 164, 153, 0.3);
    transition: all 0.3s ease;
}

.student-card:hover .student-avatar {
    transform: scale(1.1);
    box-shadow: 0 6px 10px rgba(0, 164, 153, 0.4);
}

.student-info {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    flex: 1;
    min-width: 0;
}

.student-name {
    font-weight: 600;
    color: var(--color-dark, #394049);
    font-size: 0.9375rem;
    line-height: 1.3;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.student-details {
    color: #94a3b8;
    font-size: 0.75rem;
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.375rem;
    line-height: 1.4;
}

.eneatype-badge {
    display: inline-flex;
    align-items: center;
    padding: 0.1875rem 0.5rem;
    background: linear-gradient(
        135deg,
        rgba(0, 164, 153, 0.1) 0%,
        rgba(0, 164, 153, 0.2) 100%
    );
    color: var(--color-primary-dark, #008a7f);
    border-radius: 10px;
    font-size: 0.6875rem;
    font-weight: 600;
    box-shadow: 0 2px 4px rgba(0, 164, 153, 0.1);
    white-space: nowrap;
}

/* Affinity Modal */
.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.45);
    backdrop-filter: blur(4px);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
    padding: 1rem;
}

.modal-content {
    background: var(--color-surface, #ffffff);
    border-radius: var(--radius-lg, 14px);
    max-width: 600px;
    width: 100%;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: var(--shadow-xl, 0 20px 25px rgba(0, 0, 0, 0.12));
}

.modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem;
    border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.modal-header h3 {
    margin: 0;
    color: var(--color-dark, #394049);
    font-size: 1.25rem;
}

.close-modal-btn {
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

.close-modal-btn:hover {
    background: var(--color-background, #f4f6f9);
    color: var(--color-dark, #394049);
}

.spinner {
    width: 40px;
    height: 40px;
    border: 4px solid var(--color-border, #dfe4ea);
    border-top-color: var(--color-primary, #00a499);
    border-radius: 50%;
    animation: spin 1s linear infinite;
}

@keyframes spin {
    to {
        transform: rotate(360deg);
    }
}

/* OCEAN mini grid */
.eneatype-info {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    margin-top: 0.25rem;
}

.ocean-mini-grid {
    display: flex;
    gap: 0.15rem;
    flex-wrap: wrap;
}

.ocean-mini-val {
    font-size: 0.6rem;
    font-weight: 800;
    padding: 0.05rem 0.2rem;
    border-radius: 3px;
    background: var(--color-background, #f4f6f9);
    color: #475569;
    border-left: 2px solid #94a3b8;
    line-height: 1;
}

.ocean-mini-val.o {
    border-left-color: #ef4444;
}
.ocean-mini-val.c {
    border-left-color: #f59e0b;
}
.ocean-mini-val.e {
    border-left-color: #3b82f6;
}
.ocean-mini-val.a {
    border-left-color: #10b981;
}
.ocean-mini-val.n {
    border-left-color: #8b5cf6;
}

.no-eneatype {
    color: #a0aec0;
    font-size: 0.75rem;
    font-style: italic;
}

/* Responsive */
@media (max-width: 768px) {
    .questionnaires-header {
        flex-direction: column;
        gap: 1rem;
        align-items: stretch;
    }

    .questionnaires-container {
        grid-template-columns: 1fr;
    }

    .questionnaire-dates {
        grid-template-columns: 1fr;
    }

    .groups-container {
        grid-template-columns: 1fr;
    }

    .students-pool {
        grid-template-columns: 1fr;
    }
}
</style>
