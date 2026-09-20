import { defineConfig } from 'astro/config'
import starlight from '@astrojs/starlight'

export default defineConfig({
	integrations: [
		starlight({
			title: 'DiscordKit',

			components: {
				Head: './src/components/AnalyticsHead.astro',
			},

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

			social: [
				{ icon: "github", href: 'https://github.com/freitaseric/discordkit', label: "Github" }
			],

			sidebar: [
				{
					label: 'Getting Started',
					translations: {
						'pt-BR': 'Primeiros passos',
					},
					items: [
						{ slug: 'getting-started/introduction' },
						{ slug: 'getting-started/installation' },
						{ slug: 'getting-started/first-bot', label: 'Cookbook: Build your first bot', translations: { 'pt-BR': 'Cookbook: Seu primeiro bot' } },
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
						{ slug: 'reference/api-overview' },
					],
				},
			],
		}),
	],
})
