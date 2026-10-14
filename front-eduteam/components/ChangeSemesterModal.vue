<template>
    <div v-if="isOpen" class="modal-overlay" @click="closeModal">
        <div class="modal-content" @click.stop>
            <div class="modal-header">
                <h2>Cambiar Semestre Actual</h2>
                <button @click="closeModal" class="close-btn">
                    <svg
                        width="24"
                        height="24"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                    >
                        <line x1="18" y1="6" x2="6" y2="18"></line>
                        <line x1="6" y1="6" x2="18" y2="18"></line>
                    </svg>
                </button>
            </div>

            <div class="modal-body">
                <!-- Semestres existentes -->
                <div class="section">
                    <h3>Activar Semestre Existente</h3>
                    <div v-if="loadingSemesters" class="loading">
                        Cargando semestres...
                    </div>
                    <div
                        v-else-if="existingSemesters.length > 0"
                        class="semesters-grid"
                    >
                        <div
                            v-for="semester in existingSemesters"
                            :key="semester.id"
                            class="semester-card"
                            :class="{ active: semester.is_active }"
                            @click="activateSemester(semester)"
                        >
                            <div class="semester-name">
                                {{ semester.year }}-{{ semester.term }}
                            </div>
                            <div class="semester-status">
                                <span
                                    v-if="semester.is_active"
                                    class="status-active"
                                    >Activo</span
                                >
                                <span v-else class="status-inactive"
                                    >Inactivo</span
                                >
                            </div>
                        </div>
                    </div>
                    <div v-else class="no-semesters">
                        No hay semestres existentes
                    </div>
                </div>

                <!-- Crear nuevo semestre -->
                <div class="section">
                    <h3>Crear Nuevo Semestre</h3>
                    <form
                        @submit.prevent="createNewSemester"
                        class="create-form"
                    >
                        <div class="form-group">
                            <label for="year">Año:</label>
                            <input
                                id="year"
                                v-model.number="newSemester.year"
                                type="number"
                                min="2020"
                                max="2050"
                                required
                                class="form-input"
                            />
                        </div>
                        <div class="form-group">
                            <label for="term">Período:</label>
                            <select
                                id="term"
                                v-model.number="newSemester.term"
                                required
                                class="form-select"
                            >
                                <option value="1">1er Semestre</option>
                                <option value="2">2do Semestre</option>
                            </select>
                        </div>
                        <button
                            type="submit"
                            :disabled="creating"
                            class="create-btn"
                        >
                            <span v-if="creating">Creando...</span>
                            <span v-else>Crear y Activar</span>
                        </button>
                    </form>
                </div>

                <!-- Mensajes de error -->
                <div v-if="error" class="error-message">
                    {{ error }}
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, watch, onMounted } from "vue";

const props = defineProps({
    isOpen: Boolean,
});

const emit = defineEmits(["close", "semesterChanged"]);

// Importar composables explícitamente
const { createOrActivateSemester, getAllSemesters, getCurrentSemester } =
    useSemester();

const existingSemesters = ref([]);
const loadingSemesters = ref(false);
const creating = ref(false);
const error = ref("");

const newSemester = ref({
    year: new Date().getFullYear(),
    term: 1,
});

const loadSemesters = async () => {
    loadingSemesters.value = true;
    try {
        console.log("Cargando semestres..."); // Debug
        const semesters = await getAllSemesters();
        console.log("Semestres obtenidos:", semesters); // Debug
        existingSemesters.value = semesters;
    } catch (err) {
        error.value = "Error al cargar semestres existentes";
        console.error("Error loading semesters:", err);
    } finally {
        loadingSemesters.value = false;
    }
};

const activateSemester = async (semester) => {
    if (semester.is_active) return;

    try {
        creating.value = true;
        error.value = "";

        await createOrActivateSemester(semester.year, semester.term);

        emit("semesterChanged");
        emit("close");
    } catch (err) {
        error.value = err.message || "Error al activar semestre";
    } finally {
        creating.value = false;
    }
};

const createNewSemester = async () => {
    try {
        creating.value = true;
        error.value = "";

        await createOrActivateSemester(
            newSemester.value.year,
            newSemester.value.term,
        );

        emit("semesterChanged");
        emit("close");
    } catch (err) {
        error.value = err.message || "Error al crear semestre";
    } finally {
        creating.value = false;
    }
};

const closeModal = () => {
    emit("close");
};

// Cargar semestres cuando se abre el modal
watch(
    () => props.isOpen,
    (isOpen) => {
        if (isOpen) {
            loadSemesters();
            error.value = "";
        }
    },
);

onMounted(() => {
    if (props.isOpen) {
        loadSemesters();
    }
});
</script>

<style scoped>
.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
}

.modal-content {
    background: white;
    border-radius: 12px;
    max-width: 600px;
    width: 90%;
    max-height: 80vh;
    overflow-y: auto;
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
}

.modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 24px 24px 0;
    border-bottom: 1px solid #e5e7eb;
    margin-bottom: 24px;
}

.modal-header h2 {
    margin: 0;
    color: #1f2937;
    font-size: 1.5rem;
    font-weight: 600;
}

.close-btn {
    background: none;
    border: none;
    cursor: pointer;
    color: #6b7280;
    padding: 8px;
    border-radius: 6px;
    transition: background-color 0.2s;
}

.close-btn:hover {
    background-color: #f3f4f6;
}

.modal-body {
    padding: 0 24px 24px;
}

.section {
    margin-bottom: 32px;
}

.section h3 {
    margin: 0 0 16px 0;
    color: #374151;
    font-size: 1.125rem;
    font-weight: 500;
}

.loading {
    text-align: center;
    color: #6b7280;
    padding: 20px;
}

.semesters-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 12px;
}

.semester-card {
    border: 2px solid #e5e7eb;
    border-radius: 8px;
    padding: 16px;
    text-align: center;
    cursor: pointer;
    transition: all 0.2s;
}

.semester-card:hover {
    border-color: #3b82f6;
    background-color: #f8fafc;
}

.semester-card.active {
    border-color: #10b981;
    background-color: #ecfdf5;
}

.semester-name {
    font-weight: 600;
    font-size: 1.125rem;
    color: #1f2937;
    margin-bottom: 8px;
}

.status-active {
    color: #10b981;
    font-size: 0.875rem;
    font-weight: 500;
}

.status-inactive {
    color: #6b7280;
    font-size: 0.875rem;
}

.no-semesters {
    text-align: center;
    color: #6b7280;
    padding: 32px;
    background-color: #f9fafb;
    border-radius: 8px;
    border: 2px dashed #d1d5db;
}

.create-form {
    display: grid;
    gap: 16px;
}

.form-group {
    display: grid;
    gap: 8px;
}

.form-group label {
    font-weight: 500;
    color: #374151;
    font-size: 0.875rem;
}

.form-input,
.form-select {
    padding: 12px;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-size: 1rem;
    background-color: white;
    color: #2b2f36;
    transition: border-color 0.2s;
}

.form-input:focus,
.form-select:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.create-btn {
    background-color: #1d4ed8;
    color: white;
    border: none;
    padding: 12px 24px;
    border-radius: 6px;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
}

.create-btn:hover:not(:disabled) {
    background-color: #2563eb;
}

.create-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.error-message {
    background-color: #fef2f2;
    border: 1px solid #fecaca;
    color: #991b1b;
    padding: 12px;
    border-radius: 6px;
    margin-top: 16px;
    font-size: 0.875rem;
}
</style>
