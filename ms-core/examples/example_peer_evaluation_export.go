package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Ejemplo de cómo exportar resultados de evaluación entre pares a Excel
func main() {
	// ID de la encuesta que quieres exportar
	surveyID := "679f1234567890abcdef1234" // Reemplazar con un ID real

	// URL del endpoint de exportación (puedes usar cualquiera de las dos rutas)
	// Opción 1: Ruta original
	url := fmt.Sprintf("http://localhost:8083/api/v1/peer-evaluation/surveys/%s/export", surveyID)
	// Opción 2: Ruta alternativa según especificación
	// url := fmt.Sprintf("http://localhost:8083/api/v1/surveys/%s/export/excel", surveyID)

	// Realizar la petición GET
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error al realizar la petición: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Verificar el código de estado
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error del servidor (status %d): %s\n", resp.StatusCode, string(body))
		return
	}

	// Leer el contenido del archivo Excel
	excelData, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error al leer la respuesta: %v\n", err)
		return
	}

	// Obtener el nombre del archivo desde los headers
	contentDisposition := resp.Header.Get("Content-Disposition")
	filename := "evaluacion_pares.xlsx" // Nombre por defecto
	if contentDisposition != "" {
		// Extraer el nombre del archivo del header Content-Disposition
		// Formato: attachment; filename="nombre_archivo.xlsx"
		if len(contentDisposition) > 22 {
			filename = contentDisposition[22 : len(contentDisposition)-1]
		}
	}

	// Guardar el archivo
	err = os.WriteFile(filename, excelData, 0644)
	if err != nil {
		fmt.Printf("Error al guardar el archivo: %v\n", err)
		return
	}

	fmt.Printf("✓ Archivo Excel exportado exitosamente: %s\n", filename)
	fmt.Printf("  Tamaño: %d bytes\n", len(excelData))
}

// Ejemplo alternativo usando curl desde la línea de comandos:
/*
curl -X GET "http://localhost:8083/api/v1/peer-evaluation/surveys/679f1234567890abcdef1234/export" \
  -H "Accept: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" \
  --output evaluacion_pares.xlsx
*/

// Estructura del archivo Excel generado:
/*
El archivo Excel contiene dos hojas según especificación del usuario:

1. "Detalle de Respuestas" - Detalle de todas las evaluaciones individuales:
   - Grupo ID
   - Nombre Evaluador (Nombre completo de quien evalúa)
   - Nombre Evaluado (Nombre completo de quien recibe la evaluación)
   - Criterio 1 (Tareas)
   - Criterio 2 (Participación)
   - Criterio 3 (Organización)
   - Criterio 4 (Compromiso)
   - Criterio 5 (Prosocialidad)
   - Feedback (Texto adicional)

2. "Resumen y Promedios" - Promedios agregados por estudiante evaluado:
   - Grupo ID
   - RUT
   - Nombre Completo
   - Promedio Final (Calculado en base a todos los criterios recibidos)
   - Cantidad de Evaluaciones Recibidas

RUTAS DISPONIBLES:
- GET /api/v1/peer-evaluation/surveys/:survey_id/export
- GET /api/v1/surveys/:surveyId/export/excel (Ruta alternativa)
*/

// Ejemplo de respuesta de error (JSON):
type ErrorResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

// Función auxiliar para manejar errores JSON
func handleJSONError(resp *http.Response) {
	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		fmt.Printf("Error: %s\n", errResp.Message)
	}
}

// Ejemplo de integración con frontend (JavaScript/TypeScript):
/*
async function exportPeerEvaluationToExcel(surveyId: string) {
  try {
    const response = await fetch(
      `http://localhost:8083/api/v1/peer-evaluation/surveys/${surveyId}/export`,
      {
        method: 'GET',
        headers: {
          'Accept': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
        }
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message);
    }

    // Obtener el nombre del archivo desde los headers
    const contentDisposition = response.headers.get('Content-Disposition');
    let filename = 'evaluacion_pares.xlsx';
    if (contentDisposition) {
      const match = contentDisposition.match(/filename="(.+)"/);
      if (match) filename = match[1];
    }

    // Descargar el archivo
    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);

    console.log('✓ Archivo Excel descargado exitosamente');
  } catch (error) {
    console.error('Error al exportar a Excel:', error);
    throw error;
  }
}
*/
