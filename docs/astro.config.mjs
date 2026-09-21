import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://discordkit.freitaseric.com',
  integrations: [starlight({
    title: 'DiscordKit',
    description: 'Build Discord applications in Go.',
    customCss: ['./src/styles/docs.css'],
    components: {
      Head: './src/components/AnalyticsHead.astro',
      SiteTitle: './src/components/BrandTitle.astro',
      PageTitle: './src/components/DocTitle.astro',
      Footer: './src/components/DocFooter.astro',
    },
    defaultLocale: 'root',
    locales: {
      root: { label: 'English', lang: 'en' },
      'pt-br': { label: 'Português (Brasil)', lang: 'pt-BR' },
    },
    social: [{ icon: 'github', href: 'https://github.com/freitaseric/discordkit', label: 'GitHub' }],
    editLink: { baseUrl: 'https://github.com/freitaseric/discordkit/edit/main/docs/' },
    sidebar: [
    {
        "label": "Start here",
        "translations": {
            "pt-BR": "Comece aqui"
        },
        "collapsed": false,
        "items": [
            {
                "slug": "getting-started/introduction"
            },
            {
                "slug": "getting-started/installation"
            },
            {
                "slug": "getting-started/first-bot"
            },
            {
                "slug": "guides/troubleshooting"
            }
        ]
    },
    {
        "label": "Conventions",
        "translations": {
            "pt-BR": "Convenções"
        },
        "collapsed": false,
        "items": [
            {
                "slug": "conventions"
            },
            {
                "slug": "conventions/structure"
            },
            {
                "slug": "conventions/environment"
            },
            {
                "slug": "conventions/intents"
            },
            {
                "slug": "conventions/bootstrap"
            }
        ]
    },
    {
        "label": "Commands",
        "translations": {
            "pt-BR": "Comandos"
        },
        "collapsed": false,
        "items": [
            {
                "slug": "commands"
            },
            {
                "slug": "getting-started/first-command"
            },
            {
                "slug": "commands/options"
            },
            {
                "slug": "commands/autocomplete"
            }
        ]
    },
    {
        "label": "Interactions",
        "translations": {
            "pt-BR": "Interações"
        },
        "collapsed": false,
        "items": [
            {
                "slug": "interactions"
            },
            {
                "slug": "interactions/buttons"
            },
            {
                "slug": "interactions/selects"
            },
            {
                "slug": "interactions/modals"
            },
            {
                "slug": "interactions/parameters"
            }
        ]
    },
    {
        "label": "Cookbook",
        "translations": {"pt-BR": "Cookbook"},
        "collapsed": false,
        "items": [
    {
        "slug": "cookbook"
    },
    {
        "slug": "cookbook/support-bot"
    },
    {
        "slug": "cookbook/setup"
    },
    {
        "slug": "cookbook/architecture"
    },
    {
        "slug": "cookbook/commands"
    },
    {
        "slug": "cookbook/components"
    },
    {
        "slug": "cookbook/persistence"
    },
    {
        "slug": "cookbook/tickets"
    },
    {
        "slug": "cookbook/queue"
    },
    {
        "slug": "cookbook/operations"
    },
    {
        "slug": "cookbook/laboratory"
    },
    {
        "slug": "cookbook/api-map"
    }
]
    },
    {
        "label": "Reference",
        "translations": {
            "pt-BR": "Referência"
        },
        "collapsed": false,
        "items": [
            {
                "slug": "concepts/discordkit-and-discordgo"
            },
            {
                "slug": "concepts/interaction-model"
            },
            {
                "slug": "concepts/router"
            },
            {
                "slug": "reference/api-overview"
            }
        ]
    }
],
  })],
});
