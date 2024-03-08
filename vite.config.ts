import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import { viteEnvs } from 'vite-envs'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
    viteEnvs({
      /*
       * Uncomment the following line if `.env` is gitignored in your project.
       * This enables you to use another file for declaring your variables.
       */
      // declarationFile: '.env.declaration',
      
      /*
       * This is completely optional.  
       * It enables you to define environment 
       * variables that are computed at build time.
       */
      computedEnv: async ({ resolvedConfig, /*declaredEnv, localEnv*/ }) => {

        const path = await import('path');
        const fs = await import('fs/promises');

        const packageJson = JSON.parse(
          await fs.readFile(
            path.join(resolvedConfig.root, 'package.json'),
            'utf-8'
          )
        );
        
        const DAPLA_TEAM_API_URL=process.env.DAPLA_TEAM_API_URL
        const DAPLA_TEAM_API_CLUSTER_URL=process.env.DAPLA_TEAM_API_CLUSTER_URL
        const PORT=process.env.PORT || 8080
        /*
         * Here you can define any arbitrary value they will be available 
         * in `import.meta.env` and it's type definitions.  
         * You can also compute defaults for variable declared in `.env` files.
         */
        return {
          BUILD_TIME: Date.now(),
          VERSION: packageJson.version,
          DAPLA_TEAM_API_URL: DAPLA_TEAM_API_URL,
          DAPLA_TEAM_API_CLUSTER_URL: DAPLA_TEAM_API_CLUSTER_URL,
          PORT: PORT,
        };

      }
    })
  ],
  build: {
    sourcemap: true
  }
})
