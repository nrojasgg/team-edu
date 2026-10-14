<template>
    <div class="view-content">
        <div v-if="groups.length === 0" class="no-groups">
            <p>No hay grupos definidos para esta sección.</p>
        </div>

        <div v-else class="visualization-groups">
            <div
                v-for="group in groups"
                :key="group.id"
                class="visualization-group-section"
            >
                <div
                    class="visualization-group-header"
                    @click="toggleGroup(group.id)"
                >
                    <div class="visualization-header-left">
                        <h3>{{ group.name }}</h3>
                        <span class="group-students-count"
                            >{{ group.students.length }} estudiantes</span
                        >
                    </div>
                    <div class="visualization-header-right">
                        <svg
                            :class="{ rotated: expandedGroup === group.id }"
                            class="chevron-icon"
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                        >
                            <polyline points="6 9 12 15 18 9" />
                        </svg>
                    </div>
                </div>

                <Transition name="accordion-slide">
                    <div
                        v-if="expandedGroup === group.id"
                        class="visualization-chart-container"
                    >
                        <div class="radar-chart-wrapper">
                            <Radar
                                :data="getRadarChartData(group)"
                                :options="radarChartOptions"
                            />
                        </div>
                    </div>
                </Transition>
            </div>
        </div>
    </div>
</template>

<script setup>
import { computed } from "vue";
import { Radar } from "vue-chartjs";
import { useSurveyGroupStats } from "~/composables/useSurveyGroupStats";
import {
    Chart as ChartJS,
    RadialLinearScale,
    PointElement,
    LineElement,
    Filler,
    Tooltip,
    Legend,
} from "chart.js";

// Register Chart.js components
ChartJS.register(
    RadialLinearScale,
    PointElement,
    LineElement,
    Filler,
    Tooltip,
    Legend,
);

const props = defineProps({
    groups: {
        type: Array,
        default: () => [],
    },
    allStudents: {
        type: Array,
        default: () => [],
    },
    allResponses: {
        type: Array,
        default: () => [],
    },
});

const allStudentsRef = computed(() => props.allStudents);
const allResponsesRef = computed(() => props.allResponses);

const { expandedGroup, toggleGroup, getStudentName, getGroupStudentAverages } =
    useSurveyGroupStats(allStudentsRef, allResponsesRef);

// Colors for radar chart datasets
const chartColors = [
    { bg: "rgba(255, 99, 132, 0.2)", border: "rgba(255, 99, 132, 1)" },
    { bg: "rgba(54, 162, 235, 0.2)", border: "rgba(54, 162, 235, 1)" },
    { bg: "rgba(255, 206, 86, 0.2)", border: "rgba(255, 206, 86, 1)" },
    { bg: "rgba(75, 192, 192, 0.2)", border: "rgba(75, 192, 192, 1)" },
    { bg: "rgba(153, 102, 255, 0.2)", border: "rgba(153, 102, 255, 1)" },
    { bg: "rgba(255, 159, 64, 0.2)", border: "rgba(255, 159, 64, 1)" },
    { bg: "rgba(199, 199, 199, 0.2)", border: "rgba(199, 199, 199, 1)" },
    { bg: "rgba(83, 102, 255, 0.2)", border: "rgba(83, 102, 255, 1)" },
    { bg: "rgba(255, 99, 255, 0.2)", border: "rgba(255, 99, 255, 1)" },
    { bg: "rgba(99, 255, 132, 0.2)", border: "rgba(99, 255, 132, 1)" },
];

// Radar chart options
const radarChartOptions = {
    responsive: true,
    maintainAspectRatio: true,
    plugins: {
        legend: {
            position: "top",
            labels: {
                color: "#334155",
                font: {
                    size: 11,
                },
                padding: 10,
                usePointStyle: true,
                pointStyle: "circle",
            },
        },
        tooltip: {
            backgroundColor: "rgba(255, 255, 255, 0.95)",
            titleColor: "#1e293b",
            bodyColor: "#334155",
            borderColor: "#e2e8f0",
            borderWidth: 1,
            padding: 10,
            displayColors: true,
        },
    },
    scales: {
        r: {
            min: 0,
            max: 4,
            ticks: {
                stepSize: 1,
                color: "#64748b",
                backdropColor: "transparent",
                font: {
                    size: 9,
                },
            },
            grid: {
                color: "rgba(148, 163, 184, 0.3)",
                circular: true,
            },
            angleLines: {
                color: "rgba(148, 163, 184, 0.3)",
            },
            pointLabels: {
                color: "#334155",
                font: {
                    size: 11,
                    weight: "500",
                },
            },
        },
    },
};

// Toggle group accordion provided by useSurveyGroupStats

// Get radar chart data for a group
const getRadarChartData = (group) => {
    const studentAverages = getGroupStudentAverages(group);

    const datasets = studentAverages.map((student, index) => {
        const colorIndex = index % chartColors.length;
        return {
            label: student.name,
            data: [
                student.prosocialidad,
                student.tareas,
                student.participacion,
                student.organizacion,
                student.compromiso,
            ],
            backgroundColor: chartColors[colorIndex].bg,
            borderColor: chartColors[colorIndex].border,
            borderWidth: 2,
            pointBackgroundColor: chartColors[colorIndex].border,
            pointBorderColor: "#fff",
            pointBorderWidth: 1,
            pointRadius: 4,
            pointHoverRadius: 6,
        };
    });

    return {
        labels: [
            "Prosocialidad",
            "Tareas",
            "Participación",
            "Organización",
            "Compromiso",
        ],
        datasets,
    };
};
</script>

<style scoped>
.view-content {
    min-height: 200px;
}

.no-groups {
    padding: 3rem;
    text-align: center;
    color: #94a3b8;
    font-style: italic;
}

/* Visualization View Styles */
.visualization-groups {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.visualization-group-section {
    background: white;
    border: 1px solid var(--color-border, #dfe4ea);
    border-radius: 8px;
    overflow: hidden;
}

.visualization-group-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.25rem;
    background: var(--color-background, #f4f6f9);
    border-bottom: 1px solid var(--color-border, #dfe4ea);
    cursor: pointer;
    transition: background-color 0.2s ease;
}

.visualization-group-header:hover {
    background: #f1f5f9;
}

.visualization-header-left {
    display: flex;
    align-items: center;
    gap: 1rem;
}

.visualization-header-left h3 {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: #1e293b;
}

.group-students-count {
    background: var(--color-border, #dfe4ea);
    color: #374151;
    padding: 0.25rem 0.625rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 500;
}

.visualization-header-right {
    display: flex;
    align-items: center;
    gap: 1rem;
}

.chevron-icon {
    transition: transform 0.3s ease;
    color: #64748b;
}

.chevron-icon.rotated {
    transform: rotate(180deg);
}

.visualization-chart-container {
    background: #ffffff;
    padding: 1rem 1.5rem;
    border-top: 1px solid var(--color-border, #dfe4ea);
}

.radar-chart-wrapper {
    max-width: 380px;
    margin: 0 auto;
}

/* Accordion Slide Animation */
.accordion-slide-enter-active {
    animation: accordion-open 0.3s ease-out;
}

.accordion-slide-leave-active {
    animation: accordion-close 0.25s ease-in;
}

@keyframes accordion-open {
    0% {
        opacity: 0;
        max-height: 0;
    }
    100% {
        opacity: 1;
        max-height: 500px;
    }
}

@keyframes accordion-close {
    0% {
        opacity: 1;
        max-height: 500px;
    }
    100% {
        opacity: 0;
        max-height: 0;
    }
}

@media (max-width: 768px) {
    .radar-chart-wrapper {
        max-width: 280px;
    }

    .visualization-chart-container {
        padding: 0.75rem;
    }
}
</style>
