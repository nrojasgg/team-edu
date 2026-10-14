<template>
  <div v-if="isOpen" class="modal-overlay" @click="closeModal">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>Cargar Estudiantes desde CSV</h3>
        <button @click="closeModal" class="close-btn">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="m18 6-12 12"/>
            <path d="m6 6 12 12"/>
          </svg>
        </button>
      </div>
      
      <div class="modal-body">
        <div class="upload-section">
          <div class="file-drop-zone" 
               :class="{ 'dragover': isDragOver, 'has-file': selectedFile }"
               @dragover.prevent="isDragOver = true"
               @dragleave.prevent="isDragOver = false"
               @drop.prevent="handleFileDrop">
            
            <div v-if="!selectedFile" class="drop-content">
              <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                <polyline points="14,2 14,8 20,8"/>
                <line x1="16" y1="13" x2="8" y2="13"/>
                <line x1="16" y1="17" x2="8" y2="17"/>
                <polyline points="10,9 9,9 8,9"/>
              </svg>
              <p>Arrastra tu archivo CSV aquí o haz clic para seleccionar</p>
              <p class="file-info">El archivo debe tener el formato con datos desde la línea 7</p>
            </div>
            
            <div v-else class="file-selected">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                <polyline points="14,2 14,8 20,8"/>
                <line x1="16" y1="13" x2="8" y2="13"/>
                <line x1="16" y1="17" x2="8" y2="17"/>
                <polyline points="10,9 9,9 8,9"/>
              </svg>
              <div class="file-details">
                <p class="file-name">{{ selectedFile.name }}</p>
                <p class="file-size">{{ formatFileSize(selectedFile.size) }}</p>
              </div>
              <button @click="removeFile" class="remove-file-btn">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="m18 6-12 12"/>
                  <path d="m6 6 12 12"/>
                </svg>
              </button>
            </div>
            
            <input 
              ref="fileInput" 
              type="file" 
              accept=".csv" 
              @change="handleFileSelect" 
              style="display: none"
            />
          </div>
          
          <button @click="$refs.fileInput.click()" class="select-file-btn">
            Seleccionar Archivo CSV
          </button>
        </div>
        
        <!-- Preview de estudiantes -->
        <div v-if="previewStudents.length > 0" class="preview-section">
          <h4>Vista Previa ({{ previewStudents.length }} estudiantes encontrados)</h4>
          <div class="preview-table-container">
            <table class="preview-table">
              <thead>
                <tr>
                  <th>RUT</th>
                  <th>Nombres</th>
                  <th>Paterno</th>
                  <th>Materno</th>
                  <th>Carrera</th>
                  <th>Correo Electrónico</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(student, index) in previewStudents.slice(0, 2)" :key="index">
                  <td>{{ student.rut }}</td>
                  <td>{{ student.nombres }}</td>
                  <td>{{ student.paterno }}</td>
                  <td>{{ student.materno }}</td>
                  <td>{{ student.carrera }}</td>
                  <td>{{ student.email }}</td>
                </tr>
              </tbody>
            </table>
            <p v-if="previewStudents.length > 2" class="more-students">
              ... y {{ previewStudents.length - 2 }} estudiantes más
            </p>
          </div>
        </div>
        
        <!-- Errores de parsing -->
        <div v-if="parseErrors.length > 0" class="errors-section">
          <h4>Errores encontrados:</h4>
          <ul class="error-list">
            <li v-for="(error, index) in parseErrors" :key="index" class="error-item">
              Línea {{ error.line }}: {{ error.message }}
            </li>
          </ul>
        </div>
        
        <div class="modal-actions">
          <button type="button" @click="closeModal" class="cancel-btn">
            Cancelar
          </button>
          <button 
            type="button" 
            @click="uploadStudents" 
            :disabled="loading || previewStudents.length === 0"
            class="submit-btn"
          >
            <span v-if="loading">Cargando...</span>
            <span v-else>Cargar {{ previewStudents.length }} Estudiantes</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { isValidEmail, formatRUT } from '~/composables/useValidation'

const props = defineProps({
  isOpen: Boolean
})

const emit = defineEmits(['close', 'studentsUploaded'])

const loading = ref(false)
const isDragOver = ref(false)
const selectedFile = ref(null)
const previewStudents = ref([])
const parseErrors = ref([])

const fileInput = ref(null)

// Función para formatear el tamaño del archivo
const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Number.parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// Función para manejar la selección de archivo
const handleFileSelect = async (event) => {
  const file = event.target.files[0]
  if (file && file.type === 'text/csv') {
    selectedFile.value = file
    await parseCSVFile(file)
  }
}

// Función para manejar el drop del archivo
const handleFileDrop = async (event) => {
  isDragOver.value = false
  const file = event.dataTransfer.files[0]
  if (file && file.type === 'text/csv') {
    selectedFile.value = file
    await parseCSVFile(file)
  }
}

// Función para remover el archivo seleccionado
const removeFile = () => {
  selectedFile.value = null
  previewStudents.value = []
  parseErrors.value = []
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

// Función para parsear el archivo CSV
const parseCSVFile = async (file) => {
  const csvContent = await file.text()
  parseCSVContent(csvContent)
}

// Función para parsear el contenido del CSV
const parseCSVContent = (csvContent) => {
  const lines = csvContent.split('\n')
  const students = []
  const errors = []
  
  // Comenzar desde la línea 7 (índice 6)
  for (let i = 7; i < lines.length; i++) {
    const line = lines[i].trim()
    if (!line) continue // Saltar líneas vacías
    
    try {
      const columns = parseCSVLine(line)
      
      // Verificar que tenemos al menos las columnas necesarias
      if (columns.length < 8) {
        errors.push({
          line: i + 1,
          message: 'Línea incompleta - faltan columnas'
        })
        continue
      }
      
      // Extraer datos según el formato:
      // Nº, R.U.N, PATERNO, MATERNO, NOMBRES, INGRESO, CARRERA, CORREO ELECTRÓNICO
      const rut = columns[1]?.trim()
      const paterno = columns[2]?.trim()
      const materno = columns[3]?.trim()
      const nombres = columns[4]?.trim()
      const carrera = columns[6]?.trim()
      const email = columns[7]?.trim()
      
      // Validar campos obligatorios
      if (!rut || !nombres || !email) {
        errors.push({
          line: i + 1,
          message: 'Faltan datos obligatorios (RUT, nombres o correo)'
        })
        continue
      }
      
      // Construir nombre completo
      const fullName = `${nombres} ${paterno} ${materno}`.trim()
      
      // Validar formato de email básico (sin regex vulnerable a ReDoS)
      if (!isValidEmail(email)) {
        errors.push({
          line: i + 1,
          message: 'Formato de correo electrónico inválido'
        })
        continue
      }
      
      students.push({
        rut: formatRUT(rut),
        nombres: nombres,
        paterno: paterno,
        materno: materno,
        carrera: carrera,
        fullName: fullName, // Mantener para compatibilidad con la vista
        email: email.toLowerCase()
      })
      
    } catch (error) {
      errors.push({
        line: i + 1,
        message: 'Error al procesar la línea: ' + error.message
      })
    }
  }
  
  previewStudents.value = students
  parseErrors.value = errors
}

// Email validation provided by composable `useValidation`

// addDotsToRUT provided by composable `useValidation`

// Función para parsear una línea CSV (maneja comillas y comas)
const parseCSVLine = (line) => {
  const result = []
  let current = ''
  let inQuotes = false
  let i = 0
  
  while (i < line.length) {
    const char = line[i]
    const nextChar = line[i + 1]
    
    if (char === '"') {
      if (inQuotes && nextChar === '"') {
        current += '"'
        i += 2 // Skip this quote and the next one
      } else {
        inQuotes = !inQuotes
        i++
      }
    } else if (char === ',' && !inQuotes) {
      result.push(current)
      current = ''
      i++
    } else {
      current += char
      i++
    }
  }
  
  result.push(current)
  return result
}

// formatRUT provided by composable `useValidation`

// Función para cargar los estudiantes
const uploadStudents = async () => {
  if (previewStudents.value.length === 0) return
  
  loading.value = true
  
  try {
    // Emitir evento con la Lista de estudiantes
    emit('studentsUploaded', previewStudents.value)
    
    // Cerrar modal
    closeModal()
  } catch (error) {
    console.error('Error al cargar estudiantes:', error)
  } finally {
    loading.value = false
  }
}

// Función para cerrar el modal
const closeModal = () => {
  removeFile()
  emit('close')
}
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
  padding: 1rem;
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 100%;
  max-width: 700px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
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
  font-size: 1.5rem;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: var(--color-background, #f4f6f9);
  color: var(--color-text-secondary, #64748b);
}

.modal-body {
  padding: 1.5rem;
}

.upload-section {
  margin-bottom: 2rem;
}

.file-drop-zone {
  border: 2px dashed #cbd5e0;
  border-radius: 12px;
  padding: 2rem;
  text-align: center;
  transition: all 0.3s;
  cursor: pointer;
  margin-bottom: 1rem;
}

.file-drop-zone.dragover {
  border-color: var(--color-primary, #00A499);
  background: var(--color-background, #f4f6f9);
}

.file-drop-zone.has-file {
  border-color: var(--color-primary, #00A499);
  background: #f0fff4;
}

.drop-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.drop-content svg {
  color: #94a3b8;
}

.drop-content p {
  margin: 0;
  color: var(--color-text-secondary, #64748b);
}

.file-info {
  font-size: 0.875rem;
  color: #94a3b8 !important;
}

.file-selected {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.file-selected svg {
  color: var(--color-primary, #00A499);
}

.file-details {
  flex: 1;
  text-align: left;
}

.file-name {
  margin: 0;
  font-weight: 500;
  color: var(--color-dark, #394049);
}

.file-size {
  margin: 0;
  font-size: 0.875rem;
  color: #94a3b8;
}

.remove-file-btn {
  background: #fed7d7;
  border: none;
  color: #b82e35;
  padding: 0.5rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.remove-file-btn:hover {
  background: #feb2b2;
}

.select-file-btn {
  width: 100%;
  padding: 0.75rem 1.5rem;
  background: var(--color-primary, #006b64);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
}

.select-file-btn:hover {
  background: var(--color-primary-dark, #008a7f);
}

.preview-section {
  margin-bottom: 2rem;
}

.preview-section h4 {
  margin: 0 0 1rem 0;
  color: var(--color-dark, #394049);
  font-size: 1.125rem;
}

.preview-table-container {
  border: 1px solid var(--color-border, #dfe4ea);
  border-radius: 8px;
  overflow: hidden;
}

.preview-table {
  width: 100%;
  border-collapse: collapse;
}

.preview-table th,
.preview-table td {
  padding: 0.75rem;
  text-align: left;
  border-bottom: 1px solid var(--color-border, #dfe4ea);
}

.preview-table th {
  background: var(--color-background, #f4f6f9);
  font-weight: 600;
  color: var(--color-dark, #394049);
}

.preview-table td {
  color: var(--color-text-secondary, #64748b);
}

.more-students {
  padding: 1rem;
  background: var(--color-background, #f4f6f9);
  margin: 0;
  text-align: center;
  color: #475569;
  font-style: italic;
}

.errors-section {
  margin-bottom: 2rem;
  padding: 1rem;
  background: #fed7d7;
  border-radius: 8px;
}

.errors-section h4 {
  margin: 0 0 1rem 0;
  color: #b82e35;
}

.error-list {
  margin: 0;
  padding-left: 1.5rem;
}

.error-item {
  color: #b82e35;
  margin-bottom: 0.5rem;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 2rem;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border, #dfe4ea);
}

.cancel-btn, .submit-btn {
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.cancel-btn {
  background: var(--color-background, #f4f6f9);
  color: var(--color-text-secondary, #64748b);
  border: 1px solid var(--color-border, #dfe4ea);
}

.cancel-btn:hover {
  background: var(--color-background, #f4f6f9);
  border-color: #cbd5e0;
}

.submit-btn {
  background: var(--color-primary, #006b64);
  color: white;
}

.submit-btn:hover:not(:disabled) {
  background: var(--color-primary-dark, #008a7f);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
