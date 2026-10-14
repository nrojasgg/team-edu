// Composable para manejo de datetime con formato 24 horas
export const useDatetimeInput = () => {
  // Convierte un valor de datetime-local al formato YYYY-MM-DDTHH:mm
  const formatDatetimeForInput = (date) => {
    if (!date) return ''
    
    try {
      const d = new Date(date)
      if (Number.isNaN(d.getTime())) return ''
      
      const year = d.getFullYear()
      const month = String(d.getMonth() + 1).padStart(2, '0')
      const day = String(d.getDate()).padStart(2, '0')
      const hours = String(d.getHours()).padStart(2, '0')
      const minutes = String(d.getMinutes()).padStart(2, '0')
      
      return `${year}-${month}-${day}T${hours}:${minutes}`
    } catch (error) {
      console.error('Error formatting datetime for input:', error)
      return ''
    }
  }

  // Convierte el valor del input de text (YYYY-MM-DD HH:mm) a ISO string
  const parseDatetimeFromInput = (inputValue) => {
    if (!inputValue) return ''
    
    try {
      // El input tendrá el formato YYYY-MM-DDTHH:mm (de datetime-local)
      // o podría ser YYYY-MM-DD HH:mm si lo procesamos
      const d = new Date(inputValue)
      if (Number.isNaN(d.getTime())) return ''
      
      return d.toISOString()
    } catch (error) {
      console.error('Error parsing datetime from input:', error)
      return ''
    }
  }

  // Formatea datetime para mostrar en texto (DD/MM/YYYY HH:mm)
  const formatDatetimeDisplay = (date) => {
    if (!date) return ''
    
    try {
      const d = new Date(date)
      if (Number.isNaN(d.getTime())) return ''
      
      const day = String(d.getDate()).padStart(2, '0')
      const month = String(d.getMonth() + 1).padStart(2, '0')
      const year = d.getFullYear()
      const hours = String(d.getHours()).padStart(2, '0')
      const minutes = String(d.getMinutes()).padStart(2, '0')
      
      return `${day}/${month}/${year} ${hours}:${minutes}`
    } catch (error) {
      console.error('Error formatting datetime for display:', error)
      return ''
    }
  }

  return {
    formatDatetimeForInput,
    parseDatetimeFromInput,
    formatDatetimeDisplay
  }
}
