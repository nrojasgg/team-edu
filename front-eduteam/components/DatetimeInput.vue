<template>
  <div class="datetime-input-wrapper">
    <input
      :value="formattedValue"
      type="datetime-local"
      class="datetime-input"
      @input="handleInput"
      :required="required"
      :disabled="disabled"
      v-bind="$attrs"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: [String, Date],
    default: ''
  },
  required: {
    type: Boolean,
    default: false
  },
  disabled: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue'])

// Formatea la fecha al formato YYYY-MM-DDTHH:mm
const formattedValue = computed(() => {
  if (!props.modelValue) return ''
  
  try {
    const date = new Date(props.modelValue)
    if (Number.isNaN(date.getTime())) return ''
    
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    
    return `${year}-${month}-${day}T${hours}:${minutes}`
  } catch (error) {
    console.error('Error formateando fecha:', error)
    return ''
  }
})

const handleInput = (event) => {
  const value = event.target.value
  if (value) {
    // El value de datetime-local está en formato YYYY-MM-DDTHH:mm
    const date = new Date(value)
    if (!Number.isNaN(date.getTime())) {
      emit('update:modelValue', date.toISOString())
    }
  } else {
    emit('update:modelValue', '')
  }
}
</script>

<style scoped>
.datetime-input-wrapper {
  width: 100%;
}

.datetime-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 14px;
  font-family: inherit;
  background: white;
  color: #333;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
  font-variant-numeric: tabular-nums;
}

.datetime-input:focus {
  outline: none;
  border-color: #0066cc;
  box-shadow: 0 0 0 2px rgba(0, 102, 204, 0.1);
}

.datetime-input:disabled {
  background: #f5f5f5;
  color: #595959;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .datetime-input {
    font-size: 16px;
    padding: 12px;
  }
}
</style>
