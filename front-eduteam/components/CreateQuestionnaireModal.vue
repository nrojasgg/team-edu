<template>
    <div v-if="isOpen" class="modal-overlay" @click.self="closeModal">
        <div class="modal-content">
            <div class="modal-header">
                <h2>Crear Nuevo Cuestionario</h2>
                <button @click="closeModal" class="close-btn">
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

            <div class="modal-body">
                <form @submit.prevent="handleSubmit">
                    <div class="form-group">
                        <label for="questionnaire-name"
                            >Nombre del Cuestionario *</label
                        >
                        <input
                            id="questionnaire-name"
                            v-model="formData.name"
                            type="text"
                            required
                            class="form-input"
                            placeholder="Ej: Test de Personalidad Eneagrama"
                        />
                    </div>

                    <div class="form-group">
                        <label for="questionnaire-plugin"
                            >Tipo de Plugin *</label
                        >
                        <select
                            id="questionnaire-plugin"
                            v-model="formData.plugin"
                            required
                            class="form-input"
                            :disabled="loadingPlugins"
                        >
                            <option
                                v-for="type in testTypes"
                                :key="type.value"
                                :value="type.value"
                            >
                                {{ type.label }}
                            </option>
                        </select>
                    </div>

                    <!-- Field builder (only for demographic) -->
                    <div
                        v-if="formData.plugin === 'demographic'"
                        class="form-group demographic-builder"
                    >
                        <label>Campos del cuestionario demográfico *</label>
                        <p class="builder-hint">
                            Escriba el tipo de dato que quiere preguntar, por
                            ejemplo: género, edad, comuna
                        </p>

                        <div
                            v-for="(field, fIndex) in formData.fields"
                            :key="field.id"
                            class="field-card"
                        >
                            <div class="field-row">
                                <input
                                    v-model="field.title"
                                    type="text"
                                    class="form-input"
                                    :placeholder="'Ej: Género, Edad, Comuna'"
                                />
                                <select
                                    v-model="field.type"
                                    class="form-input type-select"
                                >
                                    <option value="choice">
                                        Múltiples opciones
                                    </option>
                                    <option value="numeric">Numérico</option>
                                </select>
                                <button
                                    type="button"
                                    @click="removeField(fIndex)"
                                    class="remove-btn"
                                >
                                    ✕
                                </button>
                            </div>

                            <!-- Options list for 'choice' type -->
                            <div
                                v-if="field.type === 'choice'"
                                class="options-area"
                            >
                                <div
                                    v-for="(opt, oIndex) in field.options"
                                    :key="oIndex"
                                    class="option-row"
                                >
                                    <input
                                        v-model="field.options[oIndex]"
                                        type="text"
                                        class="form-input option-input"
                                        :placeholder="'Opción ' + (oIndex + 1)"
                                    />
                                    <button
                                        type="button"
                                        @click="removeOption(fIndex, oIndex)"
                                        class="remove-btn small"
                                    >
                                        ✕
                                    </button>
                                </div>
                                <button
                                    type="button"
                                    @click="addOption(fIndex)"
                                    class="add-option-btn"
                                >
                                    + Añadir opción
                                </button>
                            </div>
                        </div>

                        <button
                            type="button"
                            @click="addField"
                            class="add-field-btn"
                        >
                            + Añadir campo
                        </button>
                    </div>

                    <div class="form-row">
                        <div class="form-group">
                            <label for="questionnaire-start-date"
                                >Fecha de Inicio *</label
                            >
                            <DatetimeInput
                                id="questionnaire-start-date"
                                v-model="formData.startDate"
                                required
                            />
                        </div>

                        <div class="form-group">
                            <label for="questionnaire-end-date"
                                >Fecha de Término *</label
                            >
                            <DatetimeInput
                                id="questionnaire-end-date"
                                v-model="formData.endDate"
                                required
                            />
                        </div>
                    </div>

                    <div class="form-group">
                        <label for="questionnaire-description"
                            >Descripción</label
                        >
                        <textarea
                            id="questionnaire-description"
                            v-model="formData.description"
                            class="form-textarea"
                            rows="3"
                            placeholder="Descripción opcional del cuestionario"
                        ></textarea>
                    </div>

                    <div class="modal-footer">
                        <button
                            type="button"
                            @click="closeModal"
                            class="btn btn-cancel"
                        >
                            Cancelar
                        </button>
                        <button
                            type="submit"
                            class="btn btn-primary"
                            :disabled="submitting"
                        >
                            {{
                                submitting ? "Creando..." : "Crear Cuestionario"
                            }}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, watch, computed, onMounted } from "vue";
import { usePluginApi } from "~/composables/usePluginApi";

const props = defineProps({
    isOpen: Boolean,
    user: Object,
});

const emit = defineEmits(["close", "create"]);

const pluginApi = usePluginApi()
const plugins = ref([])
const loadingPlugins = ref(false)

const testTypes = computed(() => 
    plugins.value
        .filter(p => p.activo)
        .map(p => ({
            value: p.nombre_interno,
            label: p.nombre
        }))
)

const loadPlugins = async () => {
    loadingPlugins.value = true
    const result = await pluginApi.getPlugins()
    if (result.success) {
        plugins.value = result.data || []
    } else {
        plugins.value = []
    }
    loadingPlugins.value = false
}

onMounted(() => {
    loadPlugins()
})

const formData = ref({
    name: "",
    plugin: "",
    startDate: "",
    endDate: "",
    description: "",
    fields: [],
});

const getDefaultPlugin = () => {
    return testTypes.value.length > 0 ? testTypes.value[0].value : ""
}

const submitting = ref(false);

// Reset form when modal opens
watch(
    () => props.isOpen,
    (isOpen) => {
        if (isOpen) {
            // Set default dates: start today at 08:00, end in 7 days at 23:59
            const today = new Date();
            today.setHours(8, 0, 0, 0); // 08:00:00.000
            const endDate = new Date(today.getTime() + 7 * 24 * 60 * 60 * 1000);
            endDate.setHours(23, 59, 0, 0); // 23:59:00.000

            formData.value = {
                name: "",
                plugin: getDefaultPlugin(),
                startDate: today.toISOString(),
                endDate: endDate.toISOString(),
                description: "",
                fields: [],
            };
            submitting.value = false;
        }
    },
);

const closeModal = () => {
    if (!submitting.value) {
        emit("close");
    }
};

// ---- Demographic field builder helpers ----
const addField = () => {
    formData.value.fields.push({
        id: "field_" + Date.now(),
        title: "",
        type: "choice",
        options: [""],
        required: false,
    });
};

const removeField = (index) => {
    formData.value.fields.splice(index, 1);
};

const addOption = (fieldIndex) => {
    formData.value.fields[fieldIndex].options.push("");
};

const removeOption = (fieldIndex, optIndex) => {
    formData.value.fields[fieldIndex].options.splice(optIndex, 1);
};

const handleSubmit = async () => {
    if (submitting.value) return;

    // Validate dates
    if (
        new Date(formData.value.endDate) <= new Date(formData.value.startDate)
    ) {
        alert("La fecha de término debe ser posterior a la fecha de inicio");
        return;
    }

    // Validate demographic fields
    if (formData.value.plugin === "demographic") {
        if (formData.value.fields.length === 0) {
            alert("Debe añadir al menos un campo al cuestionario demográfico");
            return;
        }
        for (const f of formData.value.fields) {
            if (!f.title.trim()) {
                alert("Todos los campos deben tener un título");
                return;
            }
            if (
                f.type === "choice" &&
                f.options.filter((o) => o.trim()).length === 0
            ) {
                alert(
                    `El campo "${f.title}" necesita al menos una opción válida`,
                );
                return;
            }
        }
    }

    submitting.value = true;

    const data = {
        name: formData.value.name,
        plugin: formData.value.plugin,
        startDate: formData.value.startDate,
        endDate: formData.value.endDate,
        description: formData.value.description,
        fields: formData.value.fields || [],
    };

    emit("create", data);
    closeModal();
    submitting.value = false;
};
</script>

<style scoped>
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
    backdrop-filter: blur(4px);
}

.modal-content {
    background: white;
    border-radius: 16px;
    box-shadow: 0 25px 50px rgba(0, 0, 0, 0.15);
    width: 90vw;
    max-width: 600px;
    max-height: 90vh;
    overflow-y: auto;
    animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
    from {
        opacity: 0;
        transform: translateY(20px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}

.modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem 2rem;
    border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.modal-header h2 {
    margin: 0;
    color: var(--color-dark, #394049);
    font-size: 1.25rem;
    font-weight: 600;
}

.close-btn {
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.5rem;
    border-radius: 6px;
    color: #94a3b8;
    transition: all 0.2s;
}

.close-btn:hover {
    background: var(--color-background, #f4f6f9);
    color: var(--color-text-secondary, #64748b);
}

.modal-body {
    padding: 2rem;
}

.form-group {
    margin-bottom: 1.5rem;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    color: var(--color-text-secondary, #64748b);
    font-weight: 500;
}

.form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
}

@media (max-width: 640px) {
    .form-row {
        grid-template-columns: 1fr;
    }
}

.form-input,
.form-textarea {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid var(--color-border, #dfe4ea);
    border-radius: 6px;
    font-size: 0.875rem;
    background-color: white;
    color: var(--color-text, #2b2f36);
    transition: border-color 0.2s;
}

.form-input:focus,
.form-textarea:focus {
    outline: none;
    border-color: var(--color-primary, #00a499);
    box-shadow: 0 0 0 3px rgba(0, 164, 153, 0.1);
}

.form-textarea {
    resize: vertical;
    min-height: 80px;
}

.modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
    padding: 1.5rem 2rem;
    border-top: 1px solid var(--color-border, #dfe4ea);
    background: var(--color-background, #f4f6f9);
}

.btn {
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-weight: 500;
    transition: all 0.2s;
}

.btn-cancel {
    background: #475569;
    color: white;
}

.btn-cancel:hover {
    background: var(--color-text-secondary, #475569);
}

.btn-primary {
    background: var(--color-primary, #006b64);
    color: white;
}

.btn-primary:hover:not(:disabled) {
    background: var(--color-primary-dark, #008a7f);
}

.btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

/* --- Demographic field builder --- */
.demographic-builder .builder-hint {
    margin: 0 0 1rem;
    font-size: 0.8125rem;
    color: #64748b;
    font-style: italic;
}

.field-card {
    border: 1px solid #dfe4ea;
    border-radius: 8px;
    padding: 0.875rem;
    margin-bottom: 0.75rem;
    background: #f8fafc;
}

.field-row {
    display: flex;
    gap: 0.5rem;
    align-items: center;
    margin-bottom: 0.5rem;
}

.type-select {
    flex: 0 0 auto;
    width: auto;
}

.options-area {
    padding-left: 0.5rem;
    border-left: 2px solid #e2e8f0;
    margin-top: 0.5rem;
}

.option-row {
    display: flex;
    gap: 0.5rem;
    align-items: center;
    margin-bottom: 0.375rem;
}

.option-input {
    flex: 1;
}

.remove-btn {
    background: none;
    border: 1px solid #e53e3e;
    color: #e53e3e;
    border-radius: 4px;
    padding: 0.25rem 0.5rem;
    cursor: pointer;
    font-size: 0.8rem;
    line-height: 1;
    flex-shrink: 0;
    transition: all 0.15s;
}

.remove-btn:hover {
    background: #fff5f5;
}

.remove-btn.small {
    padding: 0.15rem 0.4rem;
}

.add-option-btn {
    background: none;
    border: 1px dashed #00a499;
    color: #00a499;
    border-radius: 4px;
    padding: 0.25rem 0.75rem;
    cursor: pointer;
    font-size: 0.8125rem;
    margin-top: 0.25rem;
    transition: all 0.15s;
}

.add-option-btn:hover {
    background: rgba(0, 164, 153, 0.06);
}

.add-field-btn {
    display: block;
    width: 100%;
    padding: 0.5rem;
    background: none;
    border: 1px dashed #94a3b8;
    color: #64748b;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.875rem;
    margin-top: 0.25rem;
    transition: all 0.15s;
}

.add-field-btn:hover {
    border-color: #00a499;
    color: #00a499;
}
</style>
