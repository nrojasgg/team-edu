<template>
  <header class="app-header">
    <!-- Thin accent strip -->
    <div class="header-accent"></div>
    <div class="header-content">
      <div class="header-left">
        <div class="brand-mark">
          <span class="brand-icon">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M22 10v6M2 10l10-5 10 5-10 5z"/>
              <path d="M6 12v5c0 1.657 2.686 3 6 3s6-1.343 6-3v-5"/>
            </svg>
          </span>
          <span class="brand-name">TeamEdu</span>
        </div>
        <nav class="main-navigation">
          <NuxtLink to="/home" class="nav-link">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
              <polyline points="9 22 9 12 15 12 15 22"/>
            </svg>
            <span>Inicio</span>
          </NuxtLink>
          <NuxtLink to="/courses" class="nav-link">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/>
              <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>
            </svg>
            <span>Cursos</span>
          </NuxtLink>
        </nav>
      </div>
      
      <div class="header-right">
        <div class="user-info">
          <div class="user-avatar">
            <img 
              v-if="currentImageUrl && !imageError" 
              :src="currentImageUrl" 
              :alt="(props.user?.fullname || props.user?.nombres) || 'Usuario'"
              class="avatar-image"
              @error="handleImageError"
              @load="handleImageLoad"
              crossorigin="anonymous"
              referrerpolicy="no-referrer"
            />
            <span v-else class="avatar-text">
              <span v-if="userInitials">{{ userInitials }}</span>
              <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                <circle cx="12" cy="7" r="4"/>
              </svg>
            </span>
          </div>
          <div class="user-details">
            <span class="user-greeting">{{ userFirstName }}</span>
            <span :class="['user-role', `role-${props.user?.role_name || 'unknown'}`]">{{ roleDisplayName }}</span>
          </div>
        </div>
        <button @click="handleLogout" class="logout-btn" title="Cerrar Sesión">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
            <polyline points="16,17 21,12 16,7"/>
            <line x1="21" y1="12" x2="9" y2="12"/>
          </svg>
          <span class="logout-label">Salir</span>
        </button>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'

// Props
const props = defineProps({
  user: {
    type: Object,
    default: null
  }
})

// Emits
const emit = defineEmits(['logout'])

const { logout } = useAuth()

// Estado para manejo de errores de imagen
const imageError = ref(false)

// Computadas para roles
const roleDisplayName = computed(() => {
  if (!props.user?.role_name) return 'Usuario'
  
  const roleNames = {
    'estudiante': 'Estudiante',
    'ayudante': 'Ayudante',
    'profesor': 'Profesor',
    'coordinador': 'Coordinador',
    'admin': 'Administrador'
  }
  
  return roleNames[props.user.role_name] || props.user.role_name
})


// Computada para mostrar solo el primer nombre
const userFirstName = computed(() => {
  const name = props.user?.fullname || props.user?.nombres;
  if (!name) return 'Usuario';
  return name.split(' ')[0];
});

const userInitials = computed(() => {
  const name = props.user?.fullname || props.user?.nombres;
  if (!name) return '';
  return name
    .split(' ')
    .map(name => name.charAt(0))
    .slice(0, 2)
    .join('')
    .toUpperCase();
});

// Computada para verificar si es admin
const isAdmin = computed(() => props.user?.role_name === 'admin')

// Variables para URLs alternativas de imagen
const currentImageUrlIndex = ref(0)
const alternativeUrls = ref([])

// Función para procesar URL de imagen de Google
const processImageUrl = (url) => {
  if (!url) return null
  
  // Si es una URL de Google, asegurar que tenga el parámetro correcto para evitar CORS
  if (url.includes('googleusercontent.com')) {
    // Remover parámetros de tamaño y agregar uno que funcione mejor
    const baseUrl = url.split('=')[0]
    return `${baseUrl}=s96-c` // Tamaño 96px y formato circular
  }
  
  return url
}

// Función para obtener URLs alternativas en caso de error
const getAlternativeImageUrls = (originalUrl) => {
  if (!originalUrl || !originalUrl.includes('googleusercontent.com')) {
    return []
  }
  
  const baseUrl = originalUrl.split('=')[0]
  return [
    `${baseUrl}=s96-c`,
    `${baseUrl}=s96`,
    `${baseUrl}=w96-h96-c`,
    originalUrl
  ]
}

// URL de imagen procesada
const processedImageUrl = computed(() => {
  return processImageUrl(props.user?.picture_url)
})

const tryNextImageUrl = () => {
  if (alternativeUrls.value.length === 0) {
    alternativeUrls.value = getAlternativeImageUrls(props.user?.picture_url)
  }
  
  currentImageUrlIndex.value++
  if (currentImageUrlIndex.value < alternativeUrls.value.length) {
    imageError.value = false
    return alternativeUrls.value[currentImageUrlIndex.value]
  }
  
  return null
}

// URL actual de imagen con fallbacks
const currentImageUrl = computed(() => {
  if (alternativeUrls.value.length > 0 && currentImageUrlIndex.value < alternativeUrls.value.length) {
    return alternativeUrls.value[currentImageUrlIndex.value]
  }
  return processedImageUrl.value
})

// Funciones para manejo de errores de imagen
const handleImageError = (event) => {
  console.warn('Error al cargar la imagen del usuario:', event)
  console.warn('URL que falló:', event.target.src)
  console.warn('URL original:', props.user?.picture_url)
  console.warn('Intentando URL alternativa...')
  
  const nextUrl = tryNextImageUrl()
  if (nextUrl) {
    console.log('Probando nueva URL:', nextUrl)
    // El reactivo currentImageUrl se actualizará automáticamente
  } else {
    console.warn('Se agotaron todas las URLs alternativas')
    imageError.value = true
  }
}

const handleImageLoad = () => {
  console.log('Imagen del usuario cargada correctamente')
  console.log('URL exitosa:', currentImageUrl.value)
  imageError.value = false
}

// Métodos
const handleLogout = () => {
  logout()
  emit('logout')
}

// Reiniciar estado de imagen cuando cambie el usuario
watch(() => props.user, () => {
  imageError.value = false
  currentImageUrlIndex.value = 0
  alternativeUrls.value = []
})

// Cargar semestre actual al montar el componente
onMounted(async () => {
  console.log('TopBar - Usuario recibido:', props.user)
  console.log('TopBar - ¿Es admin?:', props.user?.role_name === 'admin')
})
</script>

<style scoped>
/* ════════════════════════════════════════════
   APP HEADER – Professional Academic TopBar
   ════════════════════════════════════════════ */
.app-header {
  position: sticky;
  top: 0;
  z-index: 900;
  background: var(--color-dark, #394049);
  box-shadow: var(--shadow-md, 0 2px 8px rgba(57,64,73,0.08));
}

.header-accent {
  height: 3px;
  background: linear-gradient(90deg, var(--color-primary, #00A499) 0%, var(--color-accent, #EA7600) 100%);
}

.header-content {
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 1.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
}

/* ── Left: Brand + Nav ── */
.header-left {
  display: flex;
  align-items: center;
  gap: 2rem;
}

.brand-mark {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  flex-shrink: 0;
}

.brand-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-primary, #00A499);
}

.brand-name {
  font-family: var(--font-family-titles, "Red Hat Display", sans-serif);
  font-weight: 800;
  font-size: 1.15rem;
  letter-spacing: -0.3px;
  color: white;
}

.main-navigation {
  display: flex;
  gap: 0.25rem;
  align-items: center;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.45rem 0.9rem;
  color: rgba(255, 255, 255, 0.7);
  text-decoration: none;
  border-radius: var(--radius-sm, 6px);
  transition: all 0.18s ease;
  font-weight: 500;
  font-size: 0.85rem;
  position: relative;
}

.nav-link:hover {
  color: white;
  background: rgba(0, 0, 0, 0.2);
}

.nav-link.router-link-active {
  color: white;
  background: rgba(0, 107, 100, 0.65);
}

.nav-link.router-link-active::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0.9rem;
  right: 0.9rem;
  height: 2px;
  background: var(--color-accent, #EA7600);
  border-radius: 2px;
}

.nav-link svg {
  flex-shrink: 0;
  opacity: 0.85;
}

/* ── Right: User + Logout ── */
.header-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.user-avatar {
  width: 34px;
  height: 34px;
  position: relative;
  flex-shrink: 0;
  border-radius: 50%;
  background: var(--color-primary, #00A499);
  overflow: hidden;
}

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}

.avatar-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-weight: 700;
  font-size: 0.75rem;
  color: white;
  letter-spacing: 0.5px;
}

.user-details {
  display: flex;
  flex-direction: column;
  line-height: 1.3;
}

.user-greeting {
  font-size: 0.82rem;
  font-weight: 600;
  color: white;
}

.user-role {
  font-size: 0.68rem;
  font-weight: 600;
  display: inline-block;
  padding: 0.1rem 0.55rem;
  border-radius: 10px;
  text-align: center;
  letter-spacing: 0.3px;
  text-transform: uppercase;
  width: fit-content;
}

/* Role colors using corporate palette */
.role-estudiante {
  background: #005951;
  color: #a7f3ed;
}

.role-ayudante {
  background: #1a4971;
  color: #bee3f8;
}

.role-profesor {
  background: #4c1d95;
  color: #ddd6fe;
}

.role-coordinador {
  background: #78340a;
  color: #fed7aa;
}

.role-admin {
  background: #7f1d1d;
  color: #fecaca;
}

.role-unknown {
  background: #475569;
  color: white;
}

.logout-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4rem 0.9rem;
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: var(--radius-sm, 6px);
  color: rgba(255, 255, 255, 0.75);
  cursor: pointer;
  transition: all 0.18s;
  font-weight: 500;
  font-size: 0.82rem;
}

.logout-btn:hover {
  background: rgba(215, 55, 63, 0.15);
  border-color: rgba(215, 55, 63, 0.4);
  color: white;
}

/* ── Dropdown (keep existing logic) ── */
.nav-dropdown { position: relative; }

.dropdown-arrow {
  margin-left: 0.2rem;
  transition: transform 0.2s ease;
}

.nav-dropdown:hover .dropdown-arrow { transform: rotate(180deg); }

.dropdown-menu {
  position: absolute;
  top: calc(100% + 0.5rem);
  left: 0;
  background: var(--color-surface, white);
  border-radius: var(--radius-md, 10px);
  box-shadow: var(--shadow-xl, 0 8px 28px rgba(57,64,73,0.12));
  border: 1px solid var(--color-border, #dfe4ea);
  min-width: 260px;
  opacity: 0;
  visibility: hidden;
  transform: translateY(-8px);
  transition: all 0.22s ease;
  z-index: 1000;
  max-height: 70vh;
  overflow-y: auto;
}

.nav-dropdown:hover .dropdown-menu {
  opacity: 1;
  visibility: visible;
  transform: translateY(0);
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.7rem 1rem;
  color: var(--color-text, #2b2f36);
  text-decoration: none;
  border-left: 3px solid transparent;
  transition: all 0.15s ease;
}

.dropdown-item:hover {
  background: var(--color-primary-light, #e6f7f6);
  border-left-color: var(--color-primary, #00A499);
}

.dropdown-item .type-number {
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--color-dark, #394049);
  min-width: 28px;
  text-align: center;
}

.dropdown-item .type-info {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.dropdown-item .type-name {
  font-weight: 600;
  color: var(--color-dark, #394049);
  font-size: 0.9rem;
}

.dropdown-item .type-subtitle {
  font-size: 0.78rem;
  color: var(--color-text-secondary, #64707d);
}

/* ── Responsive ── */
@media (max-width: 768px) {
  .header-content {
    height: 50px;
    padding: 0 1rem;
  }

  .header-left { gap: 1rem; }

  .brand-name { display: none; }

  .nav-link span { display: none; }

  .nav-link { padding: 0.4rem 0.6rem; }

  .user-details { display: none; }

  .logout-label { display: none; }

  .logout-btn { padding: 0.4rem 0.55rem; }
}

@media (max-width: 480px) {
  .header-content { height: 46px; padding: 0 0.75rem; }
  .header-left { gap: 0.5rem; }
  .brand-icon svg { width: 18px; height: 18px; }
}
</style>
