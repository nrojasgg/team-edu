---
description: Estas reglas deben considerarse después de generar código.
# applyTo: 'Describe when these instructions should be loaded by the agent based on task context' # when provided, instructions will automatically be added to the request context when the pattern matches an attached file
---

<!-- Tip: Use /create-instructions in chat to generate content with agent assistance -->

- El proyecto está en front nuxt 4 y backend go fiber.
- Trabajamos con bun, no con npm.
- Después de generar código en el frontend, nunca ejecutes comandos de bun automáticamente.
- Después de generar código en el backend, nunca compiles el código ni ejecutes pruebas automáticamente.