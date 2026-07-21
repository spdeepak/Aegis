import { defineConfig } from 'orval'

export default defineConfig({
  aegis: {
    input: {
      target: '../server/openapi.yaml',
    },
    output: {
      target: './src/lib/api.gen.ts',
      client: 'fetch',
      override: {
        fetch: {
          includeHttpResponseReturnType: false,
        },
      },
    },
  },
})