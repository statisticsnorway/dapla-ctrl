import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import { viteEnvs } from 'vite-envs'


// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
    viteEnvs({
      declarationFile: ".env.local",
    })
  ],
})
