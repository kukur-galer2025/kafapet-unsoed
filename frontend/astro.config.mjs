import { defineConfig } from 'astro/config';
import tailwind from '@astrojs/tailwind';

// https://astro.build/config
export default defineConfig({
  integrations: [tailwind()],
  output: 'static',
  image: {
    domains: ['localhost', '127.0.0.1'],
  },
  server: {
    port: 4321
  }
});
