<template>
    <div v-if="isOpen" class="modal-overlay" @click="closeModal">
        <div class="modal-content manage-users-modal" @click.stop>
            <div class="modal-header">
                <h3>Gestionar Usuarios de Cursos</h3>
                <button class="modal-close" @click="closeModal">&times;</button>
            </div>

            <div class="modal-body">
                <!-- Barra de búsqueda y botón crear -->
                <div class="search-and-create">
                    <div class="search-bar">
                        <input
                            v-model="searchQuery"
                            type="text"
                            placeholder="Buscar por nombre, RUT o correo..."
                            class="search-input"
                        />
                        <svg
                            class="search-icon"
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                        >
                            <circle cx="11" cy="11" r="8" />
                            <path d="m21 21-4.35-4.35" />
                        </svg>
                    </div>
                    <button class="create-user-btn" @click="openCreateModal">
                        <svg
                            width="18"
                            height="18"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                        >
                            <path d="M12 5v14M5 12h14" />
                        </svg>
                        Crear Usuario
                    </button>
                </div>

                <!-- Estado de carga -->
                <div v-if="loading" class="loading-state">
                    <div class="spinner"></div>
                    <p>Cargando usuarios...</p>
                </div>

                <!-- Sin usuarios -->
                <div v-else-if="filteredUsers.length === 0" class="empty-state">
                    <p>
                        {{
                            searchQuery
                                ? "No se encontraron usuarios."
                                : "No hay usuarios disponibles."
                        }}
                    </p>
                </div>

                <!-- Tabla de usuarios -->
                <div v-else class="users-table-wrapper">
                    <table class="users-table">
                        <thead>
                            <tr>
                                <th class="sortable" @click="toggleSort('rut')">
                                    RUT<span class="sort-indicator">{{
                                        getSortIndicator("rut")
                                    }}</span>
                                </th>
                                <th
                                    class="sortable"
                                    @click="toggleSort('full_name')"
                                >
                                    Nombre Completo<span
                                        class="sort-indicator"
                                        >{{
                                            getSortIndicator("full_name")
                                        }}</span
                                    >
                                </th>
                                <th
                                    class="sortable"
                                    @click="toggleSort('role_name')"
                                >
                                    Rol<span class="sort-indicator">{{
                                        getSortIndicator("role_name")
                                    }}</span>
                                </th>
                                <th
                                    class="sortable"
                                    @click="toggleSort('email')"
                                >
                                    Correo<span class="sort-indicator">{{
                                        getSortIndicator("email")
                                    }}</span>
                                </th>
                                <th>Acciones</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr
                                v-for="user in filteredUsers"
                                :key="user.id"
                                class="user-row"
                            >
                                <td class="rut-cell">
                                    {{ user.rut || user.id }}
                                </td>
                                <td class="name-cell">
                                    <div class="user-avatar">
                                        <img
                                            v-if="user.picture_url"
                                            :src="user.picture_url"
                                            :alt="user.full_name"
                                            class="avatar-img"
                                        />
                                        <span v-else class="avatar-placeholder">
                                            {{ getInitials(user.full_name) }}
                                        </span>
                                    </div>
                                    <span class="user-name">{{
                                        user.full_name || user.nombres
                                    }}</span>
                                </td>
                                <td class="role-cell">
                                    <span
                                        class="role-badge"
                                        :class="`role-${user.role_name}`"
                                    >
                                        {{ getRoleDisplayName(user.role_name) }}
                                    </span>
                                </td>
                                <td class="email-cell">{{ user.email }}</td>
                                <td class="actions-cell">
                                    <button
                                        class="action-btn view-btn"
                                        :title="`Ver detalles de ${user.full_name}`"
                                        @click="viewUserDetails(user)"
                                    >
                                        <svg
                                            width="18"
                                            height="18"
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
                                    </button>
                                    <button
                                        class="action-btn edit-btn"
                                        :title="`Editar ${user.full_name}`"
                                        @click="editUser(user)"
                                    >
                                        <svg
                                            width="18"
                                            height="18"
                                            viewBox="0 0 24 24"
                                            fill="none"
                                            stroke="currentColor"
                                            stroke-width="2"
                                        >
                                            <path d="M12 20h9" />
                                            <path
                                                d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"
                                            />
                                        </svg>
                                    </button>
                                    <button
                                        class="action-btn delete-btn"
                                        :title="`Eliminar ${user.full_name}`"
                                        @click="deleteUser(user)"
                                    >
                                        <svg
                                            width="18"
                                            height="18"
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
                                    </button>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>

            <div class="modal-footer">
                <button class="close-btn" @click="closeModal">Cerrar</button>
            </div>
        </div>
    </div>

    <!-- Modal de detalles de usuario -->
    <div
        v-if="showDetailsModal && selectedUser"
        class="modal-overlay"
        @click="closeDetailsModal"
    >
        <div class="modal-content user-details-modal" @click.stop>
            <div class="modal-header">
                <h3>Detalles del Usuario</h3>
                <button class="modal-close" @click="closeDetailsModal">
                    &times;
                </button>
            </div>
            <div class="modal-body">
                <div class="user-details">
                    <div class="detail-group">
                        <span>Nombre Completo:</span>
                        <p>{{ selectedUser.full_name }}</p>
                    </div>
                    <div class="detail-group">
                        <span>Correo Electrónico:</span>
                        <p>{{ selectedUser.email }}</p>
                    </div>
                    <div class="detail-group">
                        <span>RUT:</span>
                        <p>{{ selectedUser.rut || "N/A" }}</p>
                    </div>
                    <div class="detail-group">
                        <span>Rol:</span>
                        <p>{{ getRoleDisplayName(selectedUser.role_name) }}</p>
                    </div>
                    <div class="detail-group">
                        <span>Estado:</span>
                        <p>{{ selectedUser.active ? "Activo" : "Inactivo" }}</p>
                    </div>
                </div>
            </div>
            <div class="modal-footer">
                <button class="close-btn" @click="closeDetailsModal">
                    Cerrar
                </button>
            </div>
        </div>
    </div>

    <!-- Modal de edición de usuario -->
    <div
        v-if="showEditModal && selectedUser"
        class="modal-overlay"
        @click="closeEditModal"
    >
        <div class="modal-content user-edit-modal" @click.stop>
            <div class="modal-header">
                <h3>Editar Usuario</h3>
                <button class="modal-close" @click="closeEditModal">
                    &times;
                </button>
            </div>
            <div class="modal-body">
                <form @submit.prevent="saveUserChanges">
                    <div class="form-group">
                        <label for="edit-nombres"
                            >Nombres: <span class="required">*</span></label
                        >
                        <input
                            id="edit-nombres"
                            v-model="selectedUser.nombres"
                            type="text"
                            required
                        />
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="edit-paterno">Apellido Paterno:</label>
                            <input
                                id="edit-paterno"
                                v-model="selectedUser.paterno"
                                type="text"
                            />
                        </div>
                        <div class="form-group">
                            <label for="edit-materno">Apellido Materno:</label>
                            <input
                                id="edit-materno"
                                v-model="selectedUser.materno"
                                type="text"
                            />
                        </div>
                    </div>
                    <div class="form-group">
                        <label for="edit-email">Correo Electrónico:</label>
                        <input
                            id="edit-email"
                            v-model="selectedUser.email"
                            type="email"
                            required
                        />
                    </div>
                    <div class="form-group">
                        <label for="edit-rut">RUT:</label>
                        <input
                            id="edit-rut"
                            v-model="selectedUser.rut"
                            type="text"
                        />
                    </div>
                    <div class="form-group">
                        <label for="edit-role">Rol:</label>
                        <select id="edit-role" v-model="selectedUser.role_id">
                            <option :value="1">Administrador</option>
                            <option :value="2">Profesor</option>
                            <option :value="3">Estudiante</option>
                        </select>
                    </div>
                    <div class="form-actions">
                        <button type="submit" class="save-btn">
                            Guardar Cambios
                        </button>
                        <button
                            type="button"
                            class="cancel-btn"
                            @click="closeEditModal"
                        >
                            Cancelar
                        </button>
                    </div>
                </form>
            </div>
        </div>
    </div>

    <!-- Modal de confirmación de eliminación -->
    <div
        v-if="showDeleteModal && selectedUser"
        class="modal-overlay"
        @click="closeDeleteModal"
    >
        <div class="modal-content delete-confirm-modal" @click.stop>
            <div class="modal-header">
                <h3>Confirmar Eliminación</h3>
                <button class="modal-close" @click="closeDeleteModal">
                    &times;
                </button>
            </div>
            <div class="modal-body">
                <p>
                    ¿Estás seguro de que deseas eliminar a
                    <strong>{{ selectedUser.full_name }}</strong
                    >?
                </p>
                <p class="warning-text">Esta acción no se puede deshacer.</p>
            </div>
            <div class="modal-footer">
                <button class="confirm-delete-btn" @click="confirmDelete">
                    Sí, Eliminar
                </button>
                <button class="cancel-btn" @click="closeDeleteModal">
                    Cancelar
                </button>
            </div>
        </div>
    </div>

    <!-- Modal de creación de usuario -->
    <div v-if="showCreateModal" class="modal-overlay" @click="closeCreateModal">
        <div class="modal-content user-edit-modal" @click.stop>
            <div class="modal-header">
                <h3>Crear Nuevo Usuario</h3>
                <button class="modal-close" @click="closeCreateModal">
                    &times;
                </button>
            </div>
            <div class="modal-body">
                <form @submit.prevent="createUser">
                    <div class="form-group">
                        <label for="create-nombres"
                            >Nombres: <span class="required">*</span></label
                        >
                        <input
                            id="create-nombres"
                            v-model="newUser.nombres"
                            type="text"
                            required
                            placeholder="Ingrese nombres"
                        />
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="create-paterno"
                                >Apellido Paterno:</label
                            >
                            <input
                                id="create-paterno"
                                v-model="newUser.paterno"
                                type="text"
                                placeholder="Ingrese apellido paterno"
                            />
                        </div>
                        <div class="form-group">
                            <label for="create-materno"
                                >Apellido Materno:</label
                            >
                            <input
                                id="create-materno"
                                v-model="newUser.materno"
                                type="text"
                                placeholder="Ingrese apellido materno"
                            />
                        </div>
                    </div>
                    <div class="form-group">
                        <label for="create-rut">RUT:</label>
                        <input
                            id="create-rut"
                            v-model="newUser.rut"
                            type="text"
                            placeholder="12345678-9"
                        />
                    </div>
                    <div class="form-group">
                        <label for="create-email"
                            >Correo Electrónico:
                            <span class="required">*</span></label
                        >
                        <input
                            id="create-email"
                            v-model="newUser.email"
                            type="email"
                            required
                            placeholder="correo@ejemplo.com"
                        />
                    </div>
                    <div class="form-group">
                        <label for="create-password"
                            >Contraseña: <span class="required">*</span></label
                        >
                        <input
                            id="create-password"
                            v-model="newUser.password"
                            type="password"
                            required
                            minlength="6"
                            placeholder="Mínimo 6 caracteres"
                        />
                    </div>
                    <div class="form-group">
                        <label for="create-role"
                            >Rol: <span class="required">*</span></label
                        >
                        <select
                            id="create-role"
                            v-model="newUser.role_id"
                            required
                        >
                            <option value="" disabled>Seleccione un rol</option>
                            <option :value="1">Administrador</option>
                            <option :value="2">Profesor</option>
                            <option :value="3">Estudiante</option>
                        </select>
                    </div>
                    <div v-if="createError" class="error-message">
                        {{ createError }}
                    </div>
                    <div class="form-actions">
                        <button
                            type="submit"
                            class="save-btn"
                            :disabled="creatingUser"
                        >
                            {{ creatingUser ? "Creando..." : "Crear Usuario" }}
                        </button>
                        <button
                            type="button"
                            class="cancel-btn"
                            @click="closeCreateModal"
                        >
                            Cancelar
                        </button>
                    </div>
                </form>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from "vue";

const props = defineProps({
    isOpen: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["close"]);

// State
const loading = ref(false);
const users = ref([]);
const searchQuery = ref("");
const selectedUser = ref(null);
const showDetailsModal = ref(false);
const showEditModal = ref(false);
const showDeleteModal = ref(false);
const showCreateModal = ref(false);
const sortColumn = ref(null);
const sortDirection = ref("asc"); // 'asc' o 'desc'
const creatingUser = ref(false);
const createError = ref("");
const newUser = ref({
    nombres: "",
    paterno: "",
    materno: "",
    rut: "",
    email: "",
    password: "",
    role_id: "",
});

const API_BASE_URL = "http://127.0.0.1:3000/api/v1";

// Computed
const filteredUsers = computed(() => {
    let result = users.value;

    // Filtrado por búsqueda
    if (searchQuery.value.trim()) {
        const query = searchQuery.value.toLowerCase();
        result = result.filter(
            (user) =>
                user.full_name?.toLowerCase().includes(query) ||
                user.email?.toLowerCase().includes(query) ||
                user.rut?.toLowerCase().includes(query),
        );
    }

    // Ordenamiento
    if (sortColumn.value) {
        result = result.slice().sort((a, b) => {
            let aVal, bVal;

            switch (sortColumn.value) {
                case "rut":
                    aVal = a.rut || "";
                    bVal = b.rut || "";
                    break;
                case "full_name":
                    aVal = a.full_name || "";
                    bVal = b.full_name || "";
                    break;
                case "role_name":
                    aVal = a.role_name || "";
                    bVal = b.role_name || "";
                    break;
                case "email":
                    aVal = a.email || "";
                    bVal = b.email || "";
                    break;
                default:
                    return 0;
            }

            // Comparar strings
            aVal = String(aVal).toLowerCase();
            bVal = String(bVal).toLowerCase();

            if (aVal < bVal) return sortDirection.value === "asc" ? -1 : 1;
            if (aVal > bVal) return sortDirection.value === "asc" ? 1 : -1;
            return 0;
        });
    }

    return result;
});

// Methods
const loadUsers = async () => {
    loading.value = true;
    try {
        const token = localStorage.getItem("auth_token");
        const response = await fetch(`${API_BASE_URL}/users?limit=1000`, {
            headers: {
                Authorization: `Bearer ${token}`,
                "Content-Type": "application/json",
            },
        });

        if (response.ok) {
            const result = await response.json();
            const rawUsers = result.data || result.users || [];

            // Mapear campos de la API a los esperados por el componente
            users.value = rawUsers.map((user) => ({
                id: user.ID || user.id,
                full_name:
                    user.FullName ||
                    user.full_name ||
                    `${user.Nombres || ""} ${user.Paterno || ""} ${user.Materno || ""}`.trim(),
                email: user.Email || user.email,
                rut: user.RUT || user.rut,
                role_id: user.RoleID || user.role_id,
                role_name: user.RoleName || user.role_name,
                picture_url: user.PictureURL || user.picture_url,
                active: user.Active !== false, // Asumir activo si no viene especificado
                nombres: user.Nombres || user.nombres,
                paterno: user.Paterno || user.paterno,
                materno: user.Materno || user.materno,
            }));
        } else {
            console.error("Error loading users:", response.status);
            users.value = [];
        }
    } catch (error) {
        console.error("Error loading users:", error);
        users.value = [];
    } finally {
        loading.value = false;
    }
};

const getRoleDisplayName = (roleName) => {
    const roleNames = {
        estudiante: "Estudiante",
        ayudante: "Ayudante",
        profesor: "Profesor",
        coordinador: "Coordinador",
        admin: "Administrador",
    };
    return roleNames[roleName] || roleName;
};

const getInitials = (name) => {
    if (!name) return "?";
    return name
        .split(" ")
        .map((word) => word[0])
        .join("")
        .toUpperCase()
        .substring(0, 2);
};

const toggleSort = (column) => {
    if (sortColumn.value === column) {
        // Si ya está ordenado por esta columna, cambiar dirección
        sortDirection.value = sortDirection.value === "asc" ? "desc" : "asc";
    } else {
        // Si es una columna nueva, ordenar ascendente
        sortColumn.value = column;
        sortDirection.value = "asc";
    }
};

const getSortIndicator = (column) => {
    if (sortColumn.value !== column) return "";
    return sortDirection.value === "asc" ? " ▲" : " ▼";
};

const viewUserDetails = (user) => {
    selectedUser.value = { ...user };
    showDetailsModal.value = true;
};

const closeDetailsModal = () => {
    showDetailsModal.value = false;
    selectedUser.value = null;
};

const editUser = (user) => {
    selectedUser.value = { ...user };
    showEditModal.value = true;
};

const closeEditModal = () => {
    showEditModal.value = false;
    selectedUser.value = null;
};

const saveUserChanges = async () => {
    try {
        const token = localStorage.getItem("auth_token");

        const roleId = Number.parseInt(selectedUser.value.role_id) || 3;

        const response = await fetch(
            `${API_BASE_URL}/users/${selectedUser.value.id}`,
            {
                method: "PUT",
                headers: {
                    Authorization: `Bearer ${token}`,
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    nombres: selectedUser.value.nombres || "",
                    paterno: selectedUser.value.paterno || "",
                    materno: selectedUser.value.materno || "",
                    email: selectedUser.value.email,
                    rut: selectedUser.value.rut || "",
                    role_id: roleId,
                }),
            },
        );

        if (response.ok) {
            alert("Usuario actualizado correctamente");
            closeEditModal();
            await loadUsers();
        } else {
            const errorData = await response.json();
            console.error("Error response:", errorData);
            alert(
                "Error al actualizar el usuario: " +
                    (errorData.message || response.statusText),
            );
        }
    } catch (error) {
        console.error("Error saving user:", error);
        alert("Error al guardar los cambios");
    }
};

const deleteUser = (user) => {
    selectedUser.value = { ...user };
    showDeleteModal.value = true;
};

const closeDeleteModal = () => {
    showDeleteModal.value = false;
    selectedUser.value = null;
};

const confirmDelete = async () => {
    try {
        const token = localStorage.getItem("auth_token");
        const response = await fetch(
            `${API_BASE_URL}/users/${selectedUser.value.id}`,
            {
                method: "DELETE",
                headers: {
                    Authorization: `Bearer ${token}`,
                    "Content-Type": "application/json",
                },
            },
        );

        if (response.ok) {
            alert("Usuario eliminado correctamente");
            closeDeleteModal();
            await loadUsers();
        } else {
            alert("Error al eliminar el usuario");
        }
    } catch (error) {
        console.error("Error deleting user:", error);
        alert("Error al eliminar el usuario");
    }
};

// Métodos para crear usuario
const openCreateModal = () => {
    newUser.value = {
        nombres: "",
        paterno: "",
        materno: "",
        rut: "",
        email: "",
        password: "",
        role_id: "",
    };
    createError.value = "";
    showCreateModal.value = true;
};

const closeCreateModal = () => {
    showCreateModal.value = false;
    createError.value = "";
};

const createUser = async () => {
    creatingUser.value = true;
    createError.value = "";

    try {
        const token = localStorage.getItem("auth_token");

        const payload = {
            nombres: newUser.value.nombres,
            paterno: newUser.value.paterno || undefined,
            materno: newUser.value.materno || undefined,
            rut: newUser.value.rut || undefined,
            email: newUser.value.email,
            password: newUser.value.password,
            role_id: Number.parseInt(newUser.value.role_id),
        };

        const response = await fetch(`${API_BASE_URL}/auth/register`, {
            method: "POST",
            headers: {
                Authorization: `Bearer ${token}`,
                "Content-Type": "application/json",
            },
            body: JSON.stringify(payload),
        });

        const result = await response.json();

        if (response.ok) {
            alert("Usuario creado correctamente");
            closeCreateModal();
            await loadUsers();
        } else {
            createError.value = result.message || "Error al crear el usuario";
        }
    } catch (error) {
        console.error("Error creating user:", error);
        createError.value = "Error de conexión al crear el usuario";
    } finally {
        creatingUser.value = false;
    }
};

const closeModal = () => {
    emit("close");
};

// Lifecycle
watch(
    () => props.isOpen,
    (newValue) => {
        if (newValue) {
            loadUsers();
        }
    },
);

onMounted(() => {
    if (props.isOpen) {
        loadUsers();
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
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
    display: flex;
    flex-direction: column;
    max-height: 90vh;
    animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
    from {
        transform: translateY(30px);
        opacity: 0;
    }
    to {
        transform: translateY(0);
        opacity: 1;
    }
}

.manage-users-modal {
    width: 90%;
    max-width: 1000px;
}

.user-details-modal,
.user-edit-modal,
.delete-confirm-modal {
    width: 90%;
    max-width: 500px;
}

.modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 24px;
    border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: #1f2937;
}

.modal-close {
    background: none;
    border: none;
    font-size: 28px;
    color: #6b7280;
    cursor: pointer;
    padding: 0;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 6px;
    transition: all 0.2s;
}

.modal-close:hover {
    background: #f3f4f6;
    color: #1f2937;
}

.modal-body {
    padding: 24px;
    overflow-y: auto;
    flex: 1;
}

.modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding: 16px 24px;
    border-top: 1px solid #e5e7eb;
}

/* Search Bar and Create Button */
.search-and-create {
    display: flex;
    gap: 12px;
    margin-bottom: 20px;
    align-items: center;
}

.search-bar {
    position: relative;
    flex: 1;
}

.search-input {
    width: 100%;
    padding: 12px 16px 12px 40px;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    font-size: 14px;
    transition: all 0.2s;
}

.search-input:focus {
    outline: none;
    border-color: #ec7803;
    box-shadow: 0 0 0 3px rgba(236, 120, 3, 0.1);
}

.search-icon {
    position: absolute;
    left: 12px;
    top: 50%;
    transform: translateY(-50%);
    color: #9ca3af;
    pointer-events: none;
}

.create-user-btn {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 20px;
    background: linear-gradient(135deg, #a85500 0%, #8f4800 100%);
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    white-space: nowrap;
    box-shadow: 0 2px 4px rgba(236, 120, 3, 0.2);
}

.create-user-btn:hover {
    background: linear-gradient(135deg, #cc6703 0%, #a85602 100%);
    transform: translateY(-1px);
    box-shadow: 0 4px 8px rgba(236, 120, 3, 0.3);
}

.create-user-btn:active {
    transform: translateY(0);
}

.create-user-btn svg {
    flex-shrink: 0;
}

/* Loading State */
.loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 60px 20px;
}

.spinner {
    width: 40px;
    height: 40px;
    border: 4px solid #e5e7eb;
    border-top-color: #ec7803;
    border-radius: 50%;
    animation: spin 1s linear infinite;
}

@keyframes spin {
    to {
        transform: rotate(360deg);
    }
}

.loading-state p {
    margin-top: 12px;
    color: #6b7280;
    font-size: 14px;
}

/* Empty State */
.empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 60px 20px;
    color: #6b7280;
    font-size: 14px;
}

/* Users Table */
.users-table-wrapper {
    overflow-x: auto;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
}

.users-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 14px;
}

.users-table thead {
    background: #f9fafb;
    border-bottom: 1px solid #e5e7eb;
}

.users-table th {
    padding: 12px 16px;
    text-align: left;
    font-weight: 600;
    color: #6b7280;
    white-space: nowrap;
}

.users-table th.sortable {
    cursor: pointer;
    user-select: none;
    transition:
        background 0.2s,
        color 0.2s;
}

.users-table th.sortable:hover {
    background: #f0f1f3;
    color: #374151;
}

.sort-indicator {
    color: #9ca3af;
    font-size: 12px;
    margin-left: 4px;
    display: inline-block;
    width: 12px;
    text-align: center;
}

.users-table tbody tr {
    border-bottom: 1px solid #e5e7eb;
    transition: background 0.2s;
}

.users-table tbody tr:hover {
    background: #f9fafb;
}

.users-table td {
    padding: 12px 16px;
}

.rut-cell {
    color: #6b7280;
    font-family: monospace;
}

.name-cell {
    display: flex;
    align-items: center;
    gap: 10px;
}

.user-avatar {
    flex-shrink: 0;
}

.avatar-img {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    object-fit: cover;
    background: #e5e7eb;
}

.avatar-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: #4b5563;
    color: white;
    font-weight: 600;
    font-size: 12px;
}

.user-name {
    color: #1f2937;
    font-weight: 500;
}

.role-cell {
    text-align: center;
}

.role-badge {
    display: inline-block;
    padding: 4px 12px;
    border-radius: 20px;
    font-size: 12px;
    font-weight: 500;
}

.role-badge.role-estudiante {
    background: #dbeafe;
    color: #1e40af;
}

.role-badge.role-ayudante {
    background: #dcfce7;
    color: #166534;
}

.role-badge.role-profesor {
    background: #fef3c7;
    color: #92400e;
}

.role-badge.role-coordinador {
    background: #fce7f3;
    color: #831843;
}

.role-badge.role-admin {
    background: #fee2e2;
    color: #991b1b;
}

.email-cell {
    color: #6b7280;
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.actions-cell {
    display: flex;
    gap: 8px;
    justify-content: center;
}

.action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border: none;
    background: transparent;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s;
    color: #6b7280;
}

.action-btn:hover {
    background: #f3f4f6;
}

.view-btn:hover {
    color: #3b82f6;
}

.edit-btn:hover {
    color: #f59e0b;
}

.delete-btn:hover {
    color: #ef4444;
}

/* User Details */
.user-details {
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.detail-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.detail-group span {
    font-weight: 600;
    color: #6b7280;
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.detail-group p {
    margin: 0;
    color: #1f2937;
    font-size: 14px;
}

/* Form Styles */
.form-group {
    margin-bottom: 16px;
}

.form-group label {
    display: block;
    margin-bottom: 6px;
    font-weight: 500;
    color: #374151;
    font-size: 14px;
}

.form-group input,
.form-group select {
    width: 100%;
    padding: 10px 12px;
    border: 1px solid #e5e7eb;
    border-radius: 6px;
    font-size: 14px;
    background-color: white;
    color: #1f2937;
    transition: all 0.2s;
}

.form-group input:focus,
.form-group select:focus {
    outline: none;
    border-color: #ec7803;
    box-shadow: 0 0 0 3px rgba(236, 120, 3, 0.1);
}

/* Form Row for side-by-side fields */
.form-row {
    display: flex;
    gap: 16px;
}

.form-row .form-group {
    flex: 1;
}

/* Required field indicator */
.required {
    color: #ef4444;
}

/* Error message */
.error-message {
    background: #fef2f2;
    border: 1px solid #fecaca;
    color: #991b1b;
    padding: 10px 14px;
    border-radius: 6px;
    margin-bottom: 16px;
    font-size: 14px;
}

/* Buttons */
.form-actions,
.modal-footer {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
}

.close-btn,
.cancel-btn {
    padding: 10px 20px;
    border: 1px solid #e5e7eb;
    background: #f7fafc;
    color: #4a5568;
    border-radius: 6px;
    cursor: pointer;
    font-size: 14px;
    font-weight: 500;
    transition: all 0.2s;
}

.close-btn:hover,
.cancel-btn:hover {
    background: #edf2f7;
}

.save-btn,
.confirm-delete-btn {
    padding: 10px 20px;
    background: #a85500;
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 14px;
    font-weight: 500;
    transition: all 0.2s;
}

.save-btn:hover {
    background: #8a4600;
}

.confirm-delete-btn {
    background: #dc2626;
}

.confirm-delete-btn:hover {
    background: #dc2626;
}

/* Responsive */
@media (max-width: 768px) {
    .manage-users-modal {
        width: 95%;
        max-height: 100vh;
    }

    .users-table {
        font-size: 12px;
    }

    .users-table th,
    .users-table td {
        padding: 8px 12px;
    }

    .modal-header {
        padding: 16px;
    }

    .modal-body {
        padding: 16px;
    }
}
</style>
