import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import { viteEnvs } from 'vite-envs'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
    viteEnvs({
      computedEnv: async ({ resolvedConfig }) => {
        const path = await import('path')
        const fs = await import('fs/promises')

        const packageJson = JSON.parse(await fs.readFile(path.join(resolvedConfig.root, 'package.json'), 'utf-8'))

        /*
         * Here you can define any arbitrary value they will be available
         * in `import.meta.env` and it's type definitions.
         * You can also compute defaults for variable declared in `.env` files.
         */
        return {
          BUILD_TIME: Date.now(),
          VERSION: packageJson.version,
          DAPLA_CTRL_ADMIN_GROUPS: process.env.DAPLA_CTRL_ADMIN_GROUPS,
          DAPLA_CTRL_DOCUMENTATION_URL: process.env.DAPLA_CTRL_DOCUMENTATION_URL,
          DAPLA_CTRL_DAPLA_START_URL: process.env.DAPLA_CTRL_DAPLA_START_URL,
        }
      },
    }),
  ],
  build: {
    sourcemap: true,
  },
})
