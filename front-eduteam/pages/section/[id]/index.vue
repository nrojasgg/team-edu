<template>
  <div>
    <SectionGeneralInfo 
      :course-data="courseData"
      :section-data="sectionData"
      :semester-info="semesterInfo"
      :professors="professors"
      :assistants="assistants"
      :assignments="assignments"
      :read-only="user?.role_name === 'estudiante'"
      :user-role="user?.role_name"
      @update-assignment="handleUpdateAssignment"
    />

    <!-- Info Modal -->
    <InfoModal
      :show="showInfoModal"
      :title="infoTitle"
      :message="infoMessage"
      @close="showInfoModal = false"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import SectionGeneralInfo from '~/components/SectionGeneralInfo.vue'
import InfoModal from '~/components/InfoModal.vue'
import { useSectionGeneral } from '~/composables/useSectionGeneral'

const props = defineProps({
  sectionData: Object,
  courseData: Object,
  semesterInfo: String,
  user: Object
})

const route = useRoute()
const { getToken } = useAuth()

// Initialize composable
const generalModule = useSectionGeneral()

const {
  professors,
  assistants,
  assignments
} = generalModule

// Info modal state
const showInfoModal = ref(false)
const infoTitle = ref('')
const infoMessage = ref('')

const showInfo = (title, message) => {
  infoTitle.value = title
  infoMessage.value = message
  showInfoModal.value = true
}

// Handlers
const handleUpdateAssignment = async (payload) => {
  try {
    const result = await generalModule.updateAssignment(route.params.id, payload)
    showInfo('Éxito', result.message || 'Equipo docente actualizado correctamente')
  } catch (err) {
    showInfo('Error', err.message || 'Error al actualizar el equipo docente')
  }
}

// Load professors and assistants when component mounts
onMounted(async () => {
  const token = getToken()
  await generalModule.loadProfessorsAndAssistants(token)
})
</script>
