<template>
  <div v-if="isOpen" class="modal-overlay" @click="closeModal">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>Agregar Estudiante</h3>
        <button @click="closeModal" class="close-btn">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="m18 6-12 12"/>
            <path d="m6 6 12 12"/>
          </svg>
        </button>
      </div>
      
      <div class="modal-body">
        <p class="modal-description">¿Cómo deseas agregar al estudiante a esta coordinación?</p>
        
        <div class="selection-grid">
          <button class="selection-card" @click="handleSelectExisting">
            <div class="card-icon existing">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/>
                <circle cx="9" cy="7" r="4"/>
                <path d="m16 11 2 2 4-4"/>
              </svg>
            </div>
            <div class="card-info">
              <h4>Estudiante Existente</h4>
              <p>Buscar un estudiante que ya está registrado en el sistema.</p>
            </div>
          </button>

          <button class="selection-card" @click="handleSelectNew">
            <div class="card-icon new">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/>
                <circle cx="9" cy="7" r="4"/>
                <line x1="19" y1="8" x2="19" y2="14"/>
                <line x1="16" y1="11" x2="22" y2="11"/>
              </svg>
            </div>
            <div class="card-info">
              <h4>Nuevo Estudiante</h4>
              <p>Crear un nuevo registro de estudiante desde cero.</p>
            </div>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  isOpen: Boolean
})

const emit = defineEmits(['close', 'select-existing', 'select-new'])

const closeModal = () => {
  emit('close')
}

const handleSelectExisting = () => {
  emit('select-existing')
  closeModal()
}

const handleSelectNew = () => {
  emit('select-new')
  closeModal()
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
  max-width: 550px;
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

.modal-description {
  color: var(--color-text-secondary, #64748b);
  margin-bottom: 1.5rem;
  text-align: center;
}

.selection-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.selection-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1.5rem;
  background: white;
  border: 2px solid var(--color-border, #dfe4ea);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
}

.selection-card:hover {
  border-color: var(--color-primary, #00A499);
  background-color: var(--color-background, #f4f6f9);
  transform: translateY(-2px);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
}

.card-icon {
  margin-bottom: 1rem;
  padding: 1rem;
  border-radius: 50%;
}

.card-icon.existing {
  background-color: #ebf8ff;
  color: var(--color-primary, #00A499);
}

.card-icon.new {
  background-color: #f0fff4;
  color: var(--color-primary, #00A499);
}

.card-info h4 {
  margin: 0 0 0.5rem 0;
  color: var(--color-dark, #394049);
  font-size: 1.1rem;
}

.card-info p {
  margin: 0;
  color: #94a3b8;
  font-size: 0.875rem;
}

@media (max-width: 480px) {
  .selection-grid {
    grid-template-columns: 1fr;
  }
}
</style>
