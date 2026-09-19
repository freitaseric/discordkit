import { defineConfig } from 'astro/config'
import starlight from '@astrojs/starlight'

export default defineConfig({
  integrations: [
    starlight({
      title: 'DiscordKit',

      description:
        'An ergonomic framework for building Discord applications in Go.',

      defaultLocale: 'root',

      locales: {
        root: {
          label: 'English',
          lang: 'en',
        },
        'pt-br': {
          label: 'Português (Brasil)',
          lang: 'pt-BR',
        },
      },

      social: {
        github: 'https://github.com/freitaseric/discordkit',
      },

      sidebar: [
        {
          label: 'Getting Started',
          translations: {
            'pt-BR': 'Primeiros passos',
          },
          items: [
            { slug: 'getting-started/introduction' },
            { slug: 'getting-started/installation' },
            { slug: 'getting-started/first-bot' },
            { slug: 'getting-started/first-command' },
          ],
        },
        {
          label: 'Concepts',
          translations: {
            'pt-BR': 'Conceitos',
          },
          items: [
            { slug: 'concepts/discordkit-and-discordgo' },
            { slug: 'concepts/interaction-model' },
            { slug: 'concepts/router' },
          ],
        },
      ],
    }),
  ],
})
