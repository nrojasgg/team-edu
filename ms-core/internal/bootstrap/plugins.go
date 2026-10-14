package bootstrap

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"ms-core/internal/core"
)

const defaultPluginDir = "./plugins_dist"

// LoadPlugins escanea pluginDir buscando binarios de plugins y lanza cada uno
// como subproceso independiente mediante hashicorp/go-plugin (net/rpc).
//
// Cada binario debe haber sido compilado desde plugins_src/<nombre>/ e implementar
// core.Plugin a través de core.ServePlugin().
// El subproceso hereda todas las variables de entorno del padre (MONGO_URI, etc.).
//
// Los errores de carga individual son no-fatales: se registran y la ejecución
// continúa. Solo retorna error si el directorio no puede ser leído.
func LoadPlugins(pm *core.PluginManager) error {
	pluginDir := os.Getenv("PLUGIN_DIR")
	if pluginDir == "" {
		pluginDir = defaultPluginDir
	}

	log.Printf("bootstrap: loading plugins from directory: %s", pluginDir)

	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		log.Printf("bootstrap: warning — could not read plugin dir %q: %v", pluginDir, err)
		return nil
	}

	loaded := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := filepath.Ext(name)

		// Ignorar archivos ocultos como .gitkeep.
		if len(name) > 0 && name[0] == '.' {
			continue
		}

		// En Windows: solo aceptar ejecutables .exe e ignorar binarios ELF de Linux (sin extensión)
		if runtime.GOOS == "windows" && ext != ".exe" {
			continue
		}

		// En Linux: aceptar solo binarios sin extensión e ignorar binarios de Windows (.exe u otros)
		if runtime.GOOS == "linux" && ext != "" {
			if ext == ".exe" {
				log.Printf("bootstrap: skipping Windows binary %q on Linux", name)
			}
			continue
		}

		binaryPath := filepath.Join(pluginDir, name)
		if err := pm.LoadPlugin(binaryPath); err != nil {
			log.Printf("bootstrap: failed to load plugin %q: %v", name, err)
			continue
		}
		loaded++
	}

	if loaded == 0 {
		log.Printf("bootstrap: warning — no plugins loaded from %q", pluginDir)
	} else {
		log.Printf("bootstrap: loaded %d plugin(s)", loaded)
	}

	return nil
}
