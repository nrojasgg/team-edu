package core

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	hcPlugin "github.com/hashicorp/go-plugin"
)

// Plugin interface que todos los plugins deben implementar
type Plugin interface {
	Name() string
	Version() string
	Initialize() error
	Execute(ctx context.Context, data interface{}) (interface{}, error)
	Shutdown() error
}

// pluginEntry almacena el plugin junto a su estado operativo.
// El estado activo/inactivo se gestiona de forma independiente al ciclo de
// vida del plugin (Initialize/Shutdown), permitiendo deshabilitar un criterio
// sin descargarlo de memoria.
type pluginEntry struct {
	plugin  Plugin
	enabled bool
}

// errPluginNotFoundFmt is the format string used when a plugin is not found.
const errPluginNotFoundFmt = "plugin %q not found"

// PluginManager maneja todos los plugins registrados y sus subprocesos.
// Cada plugin corre como un subproceso independiente comunicado mediante
// net/rpc a través de hashicorp/go-plugin.
//
// Seguridad de concurrencia:
//   - mu (sync.RWMutex) protege entries y clients en todas las operaciones.
//   - RLock para lecturas (GetPlugin, ExecutePlugin, GetLoadedPlugins, IsEnabled).
//   - Lock para escrituras (RegisterPlugin, LoadPlugin, Enable/DisablePlugin).
type PluginManager struct {
	mu      sync.RWMutex
	entries map[string]*pluginEntry
	clients []*hcPlugin.Client // subprocesos; se terminan con Kill() en Shutdown
}

// NewPluginManager crea una nueva instancia del gestor de plugins.
func NewPluginManager() *PluginManager {
	return &PluginManager{
		entries: make(map[string]*pluginEntry),
		clients: make([]*hcPlugin.Client, 0),
	}
}

// ---------------------------------------------------------------------------
// Registro
// ---------------------------------------------------------------------------

// RegisterPlugin registra y activa un nuevo plugin. Es seguro para uso concurrente.
//
// Initialize() se llama FUERA del lock para no bloquear lecturas durante una
// operación de I/O potencialmente larga (ej. creación de índices en MongoDB).
func (pm *PluginManager) RegisterPlugin(plugin Plugin) error {
	if plugin == nil {
		return errors.New("plugin_manager: plugin cannot be nil")
	}

	name := plugin.Name()
	if name == "" {
		return errors.New("plugin_manager: plugin name cannot be empty")
	}

	if err := plugin.Initialize(); err != nil {
		return fmt.Errorf("plugin_manager: failed to initialize plugin %q: %w", name, err)
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	return pm.registerLocked(name, plugin)
}

// registerLocked inserta el plugin con estado enabled=true.
// PRECONDICIÓN: pm.mu.Lock() debe estar adquirido por el llamador.
func (pm *PluginManager) registerLocked(name string, plugin Plugin) error {
	if _, exists := pm.entries[name]; exists {
		return fmt.Errorf("plugin_manager: plugin already registered: %q", name)
	}

	pm.entries[name] = &pluginEntry{
		plugin:  plugin,
		enabled: true,
	}
	log.Printf("plugin_manager: registered %q v%s", name, plugin.Version())
	return nil
}

// ---------------------------------------------------------------------------
// Carga de plugins como subprocesos (hashicorp/go-plugin)
// ---------------------------------------------------------------------------

// LoadPlugin lanza binaryPath como subproceso y registra el plugin que expone
// a través de net/rpc usando hashicorp/go-plugin.
//
// El subproceso hereda las variables de entorno del proceso padre (MONGO_URI, etc.)
// y debe haber sido compilado desde plugins_src/<nombre>/.
// El cliente gestionado se almacena internamente y se termina en Shutdown().
func (pm *PluginManager) LoadPlugin(binaryPath string) error {
	client := hcPlugin.NewClient(&hcPlugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig,
		Plugins:          PluginMap,
		Cmd:              exec.Command(binaryPath),
		AllowedProtocols: []hcPlugin.Protocol{hcPlugin.ProtocolNetRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("plugin_manager: connect to subprocess %q: %w", binaryPath, err)
	}

	raw, err := rpcClient.Dispense("plugin")
	if err != nil {
		client.Kill()
		return fmt.Errorf("plugin_manager: dispense from %q: %w", binaryPath, err)
	}

	p, ok := raw.(Plugin)
	if !ok {
		client.Kill()
		return fmt.Errorf("plugin_manager: subprocess %q did not return a Plugin", binaryPath)
	}

	if err := pm.RegisterPlugin(p); err != nil {
		client.Kill()
		return fmt.Errorf("plugin_manager: register from %q: %w", binaryPath, err)
	}

	pm.mu.Lock()
	pm.clients = append(pm.clients, client)
	pm.mu.Unlock()

	return nil
}

// ---------------------------------------------------------------------------
// Activación / Desactivación de plugins
// ---------------------------------------------------------------------------

// EnablePlugin activa un plugin previamente desactivado.
// Retorna error si el plugin no está registrado.
// Un plugin recién registrado comienza siempre en estado activo; este método
// es solo relevante después de haber llamado a DisablePlugin.
func (pm *PluginManager) EnablePlugin(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	entry, exists := pm.entries[name]
	if !exists {
		return fmt.Errorf(errPluginNotFoundFmt, name)
	}

	if entry.enabled {
		log.Printf("plugin_manager: EnablePlugin — %q already enabled", name)
		return nil
	}

	entry.enabled = true
	log.Printf("plugin_manager: plugin %q enabled", name)
	return nil
}

// DisablePlugin desactiva un plugin en tiempo de ejecución sin descargarlo.
// Después de esta llamada, ExecutePlugin y GetPlugin retornarán
// ErrPluginDisabled para este plugin hasta que se llame a EnablePlugin.
func (pm *PluginManager) DisablePlugin(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	entry, exists := pm.entries[name]
	if !exists {
		return fmt.Errorf(errPluginNotFoundFmt, name)
	}

	if !entry.enabled {
		log.Printf("plugin_manager: DisablePlugin — %q already disabled", name)
		return nil
	}

	entry.enabled = false
	log.Printf("plugin_manager: plugin %q disabled", name)
	return nil
}

// IsEnabled reporta si un plugin está activo.
// Retorna (false, error) si el plugin no existe.
func (pm *PluginManager) IsEnabled(name string) (bool, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	entry, exists := pm.entries[name]
	if !exists {
		return false, fmt.Errorf(errPluginNotFoundFmt, name)
	}
	return entry.enabled, nil
}

// ErrPluginDisabled se retorna cuando se intenta ejecutar o acceder a un plugin
// que ha sido desactivado via DisablePlugin.
var ErrPluginDisabled = errors.New("plugin is disabled")

// ---------------------------------------------------------------------------
// Consulta y ejecución
// ---------------------------------------------------------------------------

// GetPlugin obtiene un plugin por nombre.
// Retorna ErrPluginDisabled si el plugin existe pero está inactivo.
func (pm *PluginManager) GetPlugin(name string) (Plugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	entry, exists := pm.entries[name]
	if !exists {
		return nil, fmt.Errorf(errPluginNotFoundFmt, name)
	}
	if !entry.enabled {
		return nil, fmt.Errorf("plugin_manager: %w: %q", ErrPluginDisabled, name)
	}
	return entry.plugin, nil
}

// ExecutePlugin ejecuta un plugin específico con un timeout de 30 segundos.
// Retorna ErrPluginDisabled si el plugin está inactivo.
func (pm *PluginManager) ExecutePlugin(name string, data interface{}) (interface{}, error) {
	plugin, err := pm.GetPlugin(name)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return plugin.Execute(ctx, data)
}

// GetLoadedPlugins retorna información de todos los plugins registrados,
// incluyendo su estado enabled/disabled.
func (pm *PluginManager) GetLoadedPlugins() []PluginInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	info := make([]PluginInfo, 0, len(pm.entries))
	for name, entry := range pm.entries {
		info = append(info, PluginInfo{
			Name:    name,
			Version: entry.plugin.Version(),
			Enabled: entry.enabled,
		})
	}
	return info
}

// GetLoadedPluginNames retorna solo los nombres de los plugins registrados
// (tanto activos como inactivos). Útil para compatibilidad con el endpoint /health.
func (pm *PluginManager) GetLoadedPluginNames() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	names := make([]string, 0, len(pm.entries))
	for name := range pm.entries {
		names = append(names, name)
	}
	return names
}

// PluginInfo contiene el estado observable de un plugin registrado.
type PluginInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Enabled bool   `json:"enabled"`
}

// ---------------------------------------------------------------------------
// Ciclo de vida
// ---------------------------------------------------------------------------

// Shutdown cierra todos los plugins de forma ordenada y termina sus subprocesos.
// Debe llamarse cuando el servidor Fiber se apaga.
func (pm *PluginManager) Shutdown() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for name, entry := range pm.entries {
		if err := entry.plugin.Shutdown(); err != nil {
			log.Printf("plugin_manager: Shutdown error for %q: %v", name, err)
		}
	}

	// Dar tiempo a los plugins para procesar el Shutdown RPC antes de matar procesos
	time.Sleep(500 * time.Millisecond)

	for _, client := range pm.clients {
		client.Kill()
	}
	return nil
}
