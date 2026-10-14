<template>
  <div class="section-page">
    <!-- Header con información del usuario -->
    <TopBar :user="user" @logout="handleLogout" />

    <!-- Banner de sección institucional -->
    <div v-if="!loading && !error && sectionData" class="page-banner">
      <div class="container">
        <div class="banner-inner">
          <div class="banner-left">
            <nav class="breadcrumb" aria-label="Navegación">
              <NuxtLink to="/home" class="breadcrumb-link">Inicio</NuxtLink>
              <span class="breadcrumb-sep">›</span>
              <NuxtLink to="/courses" class="breadcrumb-link">Cursos</NuxtLink>
              <span class="breadcrumb-sep">›</span>
              <span class="breadcrumb-item">{{ courseData?.code }}</span>
              <span class="breadcrumb-sep">›</span>
              <span class="breadcrumb-item active">{{ sectionData?.name }}</span>
            </nav>
            <h1 class="banner-title">{{ courseData?.name }}</h1>
            <div class="banner-meta">
              <span class="meta-badge">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
                  <polyline points="9 22 9 12 15 12 15 22"/>
                </svg>
                Coordinación: {{ sectionData?.name }}
              </span>
              <span class="meta-badge">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="3" y="4" width="18" height="18" rx="2"/>
                  <line x1="16" y1="2" x2="16" y2="6"/>
                  <line x1="8" y1="2" x2="8" y2="6"/>
                  <line x1="3" y1="10" x2="21" y2="10"/>
                </svg>
                Período: {{ semesterInfo }}
              </span>
              <span class="meta-badge" v-if="courseData?.code">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/>
                  <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>
                </svg>
                Código: {{ courseData?.code }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <main class="main-content">
      <div class="container">
        <div v-if="loading" class="loading">
          Cargando información de la coordinación ...
        </div>
        
        <div v-else-if="error" class="error">
          {{ error }}
        </div>
        
        <div v-else class="section-container">
          <!-- Pestañas de navegación -->
          <div class="tabs-container">
            <div class="tabs">
              <NuxtLink 
                :to="`/section/${route.params.id}`"
                :class="['tab', { 'active': isActiveRoute('') }]"
              >
                Información general
              </NuxtLink>
              <NuxtLink 
                v-if="user?.role_name !== 'estudiante'"
                :to="`/section/${route.params.id}/list`"
                :class="['tab', { 'active': isActiveRoute('list') }]"
              >
                Lista de estudiantes
              </NuxtLink>
              <NuxtLink 
                :to="`/section/${route.params.id}/groups`"
                :class="['tab', { 'active': isActiveRoute('groups') }]"
              >
                Grupo de estudiantes
              </NuxtLink>
              <NuxtLink 
                :to="`/section/${route.params.id}/surveys`"
                :class="['tab', { 'active': isActiveRoute('surveys') }]"
              >
                Encuestas
              </NuxtLink>
            </div>
          </div>

          <!-- Contenido de las rutas -->
          <div class="tab-content">
            <NuxtPage 
              :section-data="sectionData"
              :course-data="courseData"
              :semester-info="semesterInfo"
              :user="user"
            />
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useSectionGeneral } from '~/composables/useSectionGeneral'

definePageMeta({
  middleware: 'auth'
})

const route = useRoute()
const { getUser, getToken, logout } = useAuth()

// Initialize composable
const generalModule = useSectionGeneral(route.params.id)

// Local state
const user = ref(null)

// Destructure composable states
const {
  sectionData,
  courseData,
  semesterInfo,
  loading,
  error
} = generalModule

// Helper to check active route
const isActiveRoute = (path) => {
  if (path === '') {
    return route.path === `/section/${route.params.id}` || route.path === `/section/${route.params.id}/`
  }
  return route.path.includes(`/section/${route.params.id}/${path}`)
}

// Load section data on mount
onMounted(async () => {
  const currentUser = getUser()
  if (!currentUser) {
    navigateTo('/')
    return
  }
  
  user.value = currentUser
  const token = getToken()
  const sectionId = route.params.id

  await generalModule.loadSectionData(sectionId, token)
})

// Función para manejar logout
const handleLogout = () => {
  logout()
}

useHead({
  title: `TeamEdu - ${sectionData.value?.name || 'Coordinación '}`,
  meta: [
    { name: 'description', content: 'Información de la coordinación ' }
  ]
})
</script>

<style scoped>
/* ════════════════════════════════════════════
   SECTION PAGE – Professional Academic Style
   ════════════════════════════════════════════ */
.section-page {
  min-height: 100vh;
  background: var(--color-background, #f4f6f9);
}

/* ── PAGE BANNER ── */
.page-banner {
  background: linear-gradient(135deg, var(--color-dark, #394049) 0%, #2a3038 100%);
  padding: 1.6rem 0 1.4rem;
  position: relative;
  overflow: hidden;
}

.page-banner::before {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 35%;
  height: 100%;
  background: linear-gradient(135deg, transparent 0%, rgba(0, 164, 153, 0.06) 100%);
  pointer-events: none;
}

.page-banner::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, var(--color-primary, #00A499), var(--color-accent, #EA7600), var(--color-primary, #00A499));
}

.banner-inner {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.banner-left { min-width: 0; }

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin-bottom: 0.45rem;
  flex-wrap: wrap;
}

.breadcrumb-link {
  font-size: 0.7rem;
  color: rgba(255,255,255,0.5);
  text-transform: uppercase;
  letter-spacing: 1px;
  font-weight: 500;
  text-decoration: none;
  transition: color 0.18s;
}

.breadcrumb-link:hover {
  color: rgba(255,255,255,0.85);
  text-decoration: none;
}

.breadcrumb-sep {
  color: rgba(255,255,255,0.3);
  font-size: 0.7rem;
}

.breadcrumb-item {
  font-size: 0.7rem;
  color: rgba(255,255,255,0.5);
  text-transform: uppercase;
  letter-spacing: 1px;
  font-weight: 500;
}

.breadcrumb-item.active {
  color: rgba(255,255,255,0.85);
  font-weight: 600;
}

.banner-title {
  margin: 0 0 0.5rem 0;
  font-size: 1.5rem;
  font-weight: 700;
  color: white;
  letter-spacing: -0.3px;
  line-height: 1.25;
}

.banner-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.meta-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.75rem;
  color: #ffffff;
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255,255,255,0.15);
  padding: 0.2rem 0.6rem;
  border-radius: 14px;
  font-weight: 500;
}

/* ── MAIN CONTENT ── */
.main-content {
  padding: 1.75rem 0 3rem;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1.5rem;
}

.loading, .error {
  text-align: center;
  padding: 3rem;
  background: var(--color-surface, white);
  border-radius: var(--radius-md, 10px);
  box-shadow: var(--shadow-sm, 0 1px 2px rgba(57,64,73,0.06));
  border: 1px solid var(--color-border, #dfe4ea);
  font-size: 0.9rem;
}

.loading {
  color: var(--color-text-secondary, #64707d);
}

.error {
  color: var(--color-danger, #d7373f);
}

.section-container {
  background: var(--color-surface, white);
  border-radius: var(--radius-md, 10px);
  box-shadow: var(--shadow-md, 0 2px 8px rgba(57,64,73,0.08));
  border: 1px solid var(--color-border, #dfe4ea);
  overflow: visible;
}

/* ── Tabs ── */
.tabs-container {
  border-bottom: 2px solid var(--color-border-light, #eceff3);
  background: var(--color-background, #f4f6f9);
  border-radius: var(--radius-md, 10px) var(--radius-md, 10px) 0 0;
  overflow: hidden;
}

.tabs {
  display: flex;
}

.tab {
  flex: 1;
  padding: 0.8rem 1.25rem;
  border: none;
  background: transparent;
  color: var(--color-text-secondary, #64707d);
  cursor: pointer;
  transition: all 0.18s;
  font-weight: 500;
  font-size: 0.85rem;
  border-bottom: 3px solid transparent;
  text-decoration: none;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 2.85rem;
  letter-spacing: 0.1px;
}

.tab:hover {
  background: rgba(0, 164, 153, 0.04);
  text-decoration: none;
  color: var(--color-dark, #394049);
}

.tab.active {
  background: var(--color-surface, white);
  color: var(--color-primary-dark, #008a81);
  border-bottom-color: var(--color-primary, #00A499);
  font-weight: 700;
  text-decoration: none;
}

.tab-content {
  padding: 2rem;
}

/* ── Responsive ── */
@media (max-width: 768px) {
  .banner-title { font-size: 1.2rem; }
  .banner-meta { gap: 0.35rem; }

  .tabs { flex-wrap: wrap; }

  .tab {
    flex: none;
    width: 50%;
    font-size: 0.78rem;
    padding: 0.65rem 0.5rem;
  }

  .tab-content { padding: 1rem; }
  .container { padding: 0 1rem; }
}

@media (max-width: 480px) {
  .page-banner { padding: 1.2rem 0 1rem; }
  .tab { width: 100%; }
  .container { padding: 0 0.75rem; }
  .tab-content { padding: 0.75rem; }
}
</style>
