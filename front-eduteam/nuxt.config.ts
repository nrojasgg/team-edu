// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  devServer: {
    port: 2999,
    host: process.env.HOST || '0.0.0.0',
    hmr: {
      server: false
    }
  },
  vite: {
    server: {
      allowedHosts: [
        'danyelle-unpreempted-pachydermatously.ngrok-free.dev',
        'localhost',
        '127.0.0.1'
      ]
    }
  },
  nitro: {
    // Bundlear vue y dependencias relacionadas en el output en vez de externalizarlas.
    // Necesario para que el servidor SSR funcione correctamente en Docker sin node_modules.
    externals: {
      inline: [
        'vue',
        'vue-router',
        '@vue/shared',
        '@vue/reactivity',
        '@vue/runtime-core',
        '@vue/runtime-dom',
        '@vue/server-renderer',
        'vue/server-renderer',
      ]
    },
    // Configuración para permitir imágenes externas
    routeRules: {
      '/api/proxy-image/**': { 
        headers: { 
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
          'Access-Control-Allow-Headers': 'Content-Type, Authorization'
        } 
      }
    }
  },
  runtimeConfig: {
    public: {
      // URLs permitidas para imágenes
      allowedImageDomains: [
        'lh3.googleusercontent.com',
        'lh4.googleusercontent.com', 
        'lh5.googleusercontent.com',
        'lh6.googleusercontent.com'
      ],
      // API URLs
      coreApiUrl: process.env.CORE_API_URL || 'http://127.0.0.1:3003'
    }
  }
})
