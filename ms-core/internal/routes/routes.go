package routes

import (
	"ms-core/internal/config"
	"ms-core/internal/core"
	"ms-core/internal/database"
	"ms-core/internal/handlers"
	"ms-core/internal/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Route path constants to avoid duplicate string literals.
const (
	routeResponses                     = "/responses"
	routeQuestionnaires                = "/questionnaires"
	routeQuestionnairesID              = "/questionnaires/:id"
	routeQuestionnairesTestIDResponses = "/questionnaires/:test_id/responses"
	routeTestsID                       = "/tests/:id"
	routeTestsTestIDResponses          = "/tests/:test_id/responses"
	routeSurveys                       = "/surveys"
	routeSurveysID                     = "/surveys/:id"
)

func SetupRoutes(app *fiber.App, pluginManager *core.PluginManager) {
	cfg := config.LoadConfig()

	authMW := middleware.AuthMiddleware(cfg.JWTSecret)
	adminMW := middleware.RoleMiddleware("admin")

	// Permisos agrupados para reutilización
	canCreateQuestionnaire := middleware.PermissionMiddleware(
		"Crear cuestionario",
		"Crear cuestionario para curso",
		"Crear cuestionario para sección",
		"Crear cuestionario para grupo",
	)
	canEditQuestionnaire := middleware.PermissionMiddleware("Editar cuestionario")
	canViewQuestionnaires := middleware.PermissionMiddleware("Ver todas las cuestionario", "Acceder a cuestionario vistas")
	canSubmitQuestionnaire := middleware.PermissionMiddleware("Acceder a cuestionario vistas")
	canViewQResponses := middleware.PermissionMiddleware("Ver todas las respuestas en cuestionario", "Acceder a cuestionario vistas")

	canViewGroups := middleware.PermissionMiddleware("Ver grupos propios", "Editar grupos")
	canEditGroups := middleware.PermissionMiddleware("Editar grupos")
	canFormRandom := middleware.PermissionMiddleware("Crear grupos aleatorios")
	canFormWithPlugins := middleware.PermissionMiddleware("Crear grupos con plugins")

	// Crear handlers
	groupHandler := handlers.NewGroupHandler(database.GetDatabase(), pluginManager)
	sectionCompleteDataHandler := handlers.NewSectionCompleteDataHandler(database.GetDatabase(), pluginManager)

	// Middleware de caché
	redisCache := middleware.RedisCacheMiddleware(database.GetRedisClient(), 15*time.Minute)

	api := app.Group("/api/v1")

	// ─────────────────────────────────────────
	// Sections (resumen de datos de sección)
	// ─────────────────────────────────────────
	sections := api.Group("/sections", authMW)

	sections.Get("/:section_id/complete-groups-data", canViewGroups, sectionCompleteDataHandler.GetCompleteSectionGroupsData)
	sections.Get("/:section_id/group-data", canViewGroups, sectionCompleteDataHandler.GetSectionGroupData)
	sections.Get("/:section_id/survey-data", canViewGroups, sectionCompleteDataHandler.GetSectionSurveyData)

	// ─────────────────────────────────────────
	// Groups
	// ─────────────────────────────────────────
	groups := api.Group("/groups", authMW)

	groups.Get("/section/:sectionId", canViewGroups, groupHandler.GetGroupsBySectionID)
	groups.Get("/section/:sectionId/export", canViewGroups, groupHandler.ExportGroupsToExcel)
	groups.Get("/section/:sectionId/my-group", middleware.PermissionMiddleware("Ver grupos propios"), groupHandler.GetStudentGroupMates)
	groups.Post("/", canEditGroups, groupHandler.CreateOrUpdateGroup)
	groups.Post("/bulk", canEditGroups, groupHandler.CreateOrUpdateGroupsBulk)
	groups.Delete("/section/:sectionId/student/:studentId", canEditGroups, groupHandler.RemoveStudentFromGroup)
	groups.Delete("/section/:sectionId", adminMW, groupHandler.DeleteGroupsBySection)
	groups.Post("/section/:sectionId/form-by-enneagram", canFormWithPlugins, groupHandler.FormGroupsByEnneagram)
	groups.Post("/section/:sectionId/form-random", canFormRandom, groupHandler.FormRandomGroups)
	groups.Patch("/section/:sectionId/scores", canEditGroups, groupHandler.PatchGroupScores)

	// ─────────────────────────────────────────
	// Plugins (solo admin)
	// ─────────────────────────────────────────
	pluginsGroup := api.Group("/plugins", authMW)

	pluginsGroup.Get("/", redisCache, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"data":    pluginManager.GetLoadedPlugins(),
		})
	})
	pluginsGroup.Put("/:name/enable", adminMW, func(c *fiber.Ctx) error {
		name := c.Params("name")
		if err := pluginManager.EnablePlugin(name); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}
		return c.JSON(fiber.Map{"success": true, "plugin": name, "enabled": true})
	})
	pluginsGroup.Put("/:name/disable", adminMW, func(c *fiber.Ctx) error {
		name := c.Params("name")
		if err := pluginManager.DisablePlugin(name); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}
		return c.JSON(fiber.Map{"success": true, "plugin": name, "enabled": false})
	})

	// Endpoint para obtener metadatos unificados de todos los plugins
	pluginsGroup.Get("/metadata", func(c *fiber.Ctx) error {
		plugins := pluginManager.GetLoadedPlugins()
		metadataList := make([]interface{}, 0)

		for _, pluginInfo := range plugins {
			result, err := pluginManager.ExecutePlugin(pluginInfo.Name, map[string]interface{}{
				"action": "get_metadata",
			})
			if err != nil {
				continue
			}
			metadataList = append(metadataList, result)
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data":    metadataList,
		})
	})

	// Endpoint para obtener metadatos extendidos con preguntas completas y opciones
	pluginsGroup.Get("/metadata/extended", func(c *fiber.Ctx) error {
		plugins := pluginManager.GetLoadedPlugins()
		metadataList := make([]interface{}, 0)

		for _, pluginInfo := range plugins {
			result, err := pluginManager.ExecutePlugin(pluginInfo.Name, map[string]interface{}{
				"action": "get_metadata_extended",
			})
			if err != nil {
				continue
			}
			metadataList = append(metadataList, result)
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data":    metadataList,
		})
	})

	// ─────────────────────────────────────────────────────────────────────────
	// Specific routes for plugins with non-standard patterns
	// ─────────────────────────────────────────────────────────────────────────
	// group_compatibility uses /surveys and /responses/section/:sectionId
	groupCompatProxy := api.Group("/plugin/group_compatibility", authMW)
	groupCompatProxy.Get("/responses/section/:sectionId", canViewQResponses, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "group_compatibility")
		return proxy.GetSectionResponsesGroupCompat(c)
	})
	groupCompatProxy.Get("/surveys/:surveyId/responses/:studentId", canViewQResponses, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "group_compatibility")
		return proxy.GetStudentSurveyResponse(c)
	})

	// group_compatibility legacy routes (required by api-gateway and frontend)
	groupCompatLegacyProxy := api.Group("/group-compatibility", authMW)
	groupCompatLegacyProxy.Get("/surveys/section/:section_id", canViewQuestionnaires, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "group_compatibility")
		return proxy.GetTests(c)
	})
	groupCompatLegacyProxy.Get("/surveys/:surveyId/responses/:studentId", canViewQResponses, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "group_compatibility")
		return proxy.GetStudentSurveyResponse(c)
	})
	groupCompatLegacyProxy.Get("/surveys/:surveyId/responses", canViewQResponses, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "group_compatibility")
		return proxy.GetSurveyResponses(c)
	})

	// peer_evaluation legacy routes (required by frontend)
	peerEvalProxy := api.Group("/peer-evaluation", authMW)
	peerEvalProxy.Get("/surveys", canViewQuestionnaires, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "peer_evaluation")
		return proxy.GetTests(c)
	})
	peerEvalProxy.Post("/surveys", canCreateQuestionnaire, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "peer_evaluation")
		return proxy.CreateTest(c)
	})
	peerEvalProxy.Get("/surveys/:id", canViewQuestionnaires, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "peer_evaluation")
		return proxy.GetTest(c)
	})
	peerEvalProxy.Put("/surveys/:id", canEditQuestionnaire, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "peer_evaluation")
		return proxy.UpdateTest(c)
	})
	peerEvalProxy.Delete("/surveys/:id", canEditQuestionnaire, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "peer_evaluation")
		return proxy.DeleteTest(c)
	})
	peerEvalProxy.Post("/surveys/:test_id/responses", canSubmitQuestionnaire, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "peer_evaluation")
		return proxy.SubmitResponse(c)
	})
	peerEvalProxy.Get("/surveys/:test_id/responses", canViewQResponses, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "peer_evaluation")
		return proxy.GetResponses(c)
	})
	peerEvalProxy.Get("/surveys/:test_id/results", canViewQResponses, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "peer_evaluation")
		return proxy.GetResults(c)
	})
	peerEvalProxy.Post("/survey-peers", canSubmitQuestionnaire, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "peer_evaluation")
		return proxy.GetSurveyPeers(c)
	})

	// Enneagram legacy routes (required by frontend)
	eneagramProxy := api.Group("/enneagram", authMW)
	eneagramProxy.Get("/sections/:section_id/student-eneatypes", canViewQResponses, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "enneagram")
		return proxy.GetStudentEneatypes(c)
	})
	eneagramProxy.Get("/questionnaires", canViewQuestionnaires, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "enneagram")
		return proxy.GetTests(c)
	})
	eneagramProxy.Get("/questionnaires/:id/responses", canViewQResponses, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "enneagram")
		return proxy.GetResponses(c)
	})

	// BFI-44 legacy routes (required by frontend)
	bfi44Proxy := api.Group("/bfi44", authMW)
	bfi44Proxy.Get("/questionnaires", canViewQuestionnaires, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "bfi44")
		return proxy.GetTests(c)
	})
	bfi44Proxy.Get("/questionnaires/:id/responses", canViewQResponses, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "bfi44")
		return proxy.GetResponses(c)
	})

	// MBTI legacy routes (required by frontend)
	mbtiProxy := api.Group("/mbti", authMW)
	mbtiProxy.Get("/questionnaires", canViewQuestionnaires, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "mbti")
		return proxy.GetTests(c)
	})
	mbtiProxy.Get("/questionnaires/:id/responses", canViewQResponses, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "mbti")
		return proxy.GetResponses(c)
	})

	// Demographic legacy routes (required by frontend)
	demographicProxy := api.Group("/demographic", authMW)
	demographicProxy.Get("/tests", canViewQuestionnaires, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "demographic")
		return proxy.GetTests(c)
	})
	demographicProxy.Get("/tests/:id/responses", canViewQResponses, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, "demographic")
		return proxy.GetResponses(c)
	})

	// ─────────────────────────────────────────────────────────────────────────
	// Generic Plugin Proxy (for ALL psychometric plugins)
	// This allows adding new plugins WITHOUT creating new handlers or routes.
	// Just add the plugin binary to plugins_dist/ and it works!
	// These routes come AFTER specific routes so specific routes take precedence.
	// ─────────────────────────────────────────────────────────────────────────
	pluginProxy := api.Group("/plugin/:pluginName", authMW)

	// Specific routes for peer_evaluation (uses surveys, not tests)
	pluginProxy.Get("/surveys", canViewQuestionnaires, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, c.Params("pluginName"))
		return proxy.GetSurveys(c)
	})

	pluginProxy.Post("/tests", canCreateQuestionnaire, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, c.Params("pluginName"))
		return proxy.CreateTest(c)
	})
	pluginProxy.Get("/tests", canViewQuestionnaires, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, c.Params("pluginName"))
		return proxy.GetTests(c)
	})
	pluginProxy.Get("/tests/:id", canViewQuestionnaires, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, c.Params("pluginName"))
		return proxy.GetTest(c)
	})
	pluginProxy.Put("/tests/:id", canEditQuestionnaire, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, c.Params("pluginName"))
		return proxy.UpdateTest(c)
	})
	pluginProxy.Delete("/tests/:id", canEditQuestionnaire, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, c.Params("pluginName"))
		return proxy.DeleteTest(c)
	})
	pluginProxy.Post("/tests/:test_id/responses", canSubmitQuestionnaire, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, c.Params("pluginName"))
		return proxy.SubmitResponse(c)
	})
	pluginProxy.Get("/tests/:test_id/responses", canViewQResponses, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, c.Params("pluginName"))
		return proxy.GetResponses(c)
	})
	pluginProxy.Get("/tests/:test_id/responses/:studentId", canViewQResponses, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, c.Params("pluginName"))
		return proxy.GetStudentTestResponse(c)
	})
	pluginProxy.Get("/responses", canViewQResponses, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, c.Params("pluginName"))
		return proxy.GetSectionResponses(c)
	})
	pluginProxy.Get("/tests/:test_id/results", canViewQResponses, func(c *fiber.Ctx) error {
		proxy := handlers.NewGenericPluginProxy(pluginManager, c.Params("pluginName"))
		return proxy.GetResults(c)
	})
}

// SetupTeamFormationRoutes configura las rutas para el motor de formación de equipos
func SetupTeamFormationRoutes(app *fiber.App, handler *handlers.TeamFormationHandler) {
	cfg := config.LoadConfig()
	authMW := middleware.AuthMiddleware(cfg.JWTSecret)

	teamFormation := app.Group("/api/v1/team-formation", authMW)

	teamFormation.Post("/form", middleware.PermissionMiddleware("Crear grupos aleatorios", "Crear grupos con plugins"), handler.FormTeams)
	teamFormation.Get("/providers", middleware.PermissionMiddleware("Crear grupos aleatorios", "Crear grupos con plugins"), handler.GetProviders)
	teamFormation.Put("/weights", middleware.RoleMiddleware("admin"), handler.UpdateWeights)
	teamFormation.Get("/affinity", middleware.PermissionMiddleware("Ver grupos propios", "Editar grupos"), handler.CalculateAffinity)
	teamFormation.Post("/form-flexible", middleware.PermissionMiddleware("Crear grupos con plugins"), handler.FlexibleFormTeams)
	teamFormation.Post("/form-flexible/async", middleware.PermissionMiddleware("Crear grupos con plugins"), handler.FlexibleFormTeamsAsync)
	teamFormation.Get("/options", middleware.PermissionMiddleware("Crear grupos aleatorios", "Crear grupos con plugins"), handler.GetAvailableOptions)
	teamFormation.Post("/evaluate-groups", middleware.PermissionMiddleware("Ver grupos propios", "Editar grupos"), handler.EvaluateGroups)
	teamFormation.Post("/evaluate-groups/async", middleware.PermissionMiddleware("Ver grupos propios", "Editar grupos"), handler.EvaluateGroupsAsync)

	// ── Job polling ────────────────────────────────────────────────────────────
	// GET /api/v1/jobs/:id — polls the result of any previously enqueued async job.
	// Registered under /api/v1 (not /team-formation) so it is reachable from any
	// client without knowing which subsystem produced the job.
	jobs := app.Group("/api/v1/jobs", authMW)
	jobs.Get("/:id", middleware.PermissionMiddleware(
		"Crear grupos con plugins",
		"Crear grupos aleatorios",
		"Ver grupos propios",
		"Editar grupos",
	), handler.GetJobResult)
}
