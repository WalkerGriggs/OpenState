module.exports = {
  title: 'OpenState',
  tagline: 'OpenState is a language agnostic task runner built with modern technologies for modern workflows.',
  url: 'https://github.com/walkergriggs/openstate',
  baseUrl:  '/',
  favicon: '/img/favicon.png',
  organizationName: 'WalkerGriggs',
  projectName: 'OpenState',
  scripts: [
    {
      src: 'https://buttons.github.io/buttons.js',
      async: true,
      defer: true,
    }
  ],
  themeConfig: {
    colorMode: {
      defaultMode: 'light',
      disableSwitch: true,
    },
    navbar: {
      logo: {
        alt: 'OpenState Logo',
        src: 'img/eightshift-dev-kit-logo.svg',
      },
      items: [
        {
          to: '/',
          activeBasePath: '/',
          label: 'Overview',
          position: 'right',
        },
        {
          to: '/docs/welcome',
          activeBasePath: 'welcome',
          label: 'Docs',
          position: 'right',
        },
        {
          to: '/docs/api/http-api',
          activeBasePath: 'http-api',
          label: 'API',
          position: 'right',
        },
        {
          to: '/community',
          activeBasePath: 'community',
          label: 'Community',
          position: 'right',
        },
        {
          to: 'https://github.com/walkergriggs/openstate',
          activeBasePath: 'openstate',
          label: 'Github',
          position: 'right',
        },
      ],
    },
    // algolia: {
    //   apiKey: '',
    //   indexName: '',
    // },
  },
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
        },
        theme: {
          customCss: require.resolve('./src/scss/_application.scss'),
        },
      },
    ],
  ],
  plugins: [
    'docusaurus-plugin-sass',
    [
      '@docusaurus/plugin-sitemap',
      {
        changefreq: 'weekly',
        priority: 0.5,
        trailingSlash: false,
      },
    ],
  ],
  customFields: {
    keywords: [
        'workflow manager',
        'task runner',
        'language agnostic',
        'public cloud',
        'orchestration',
        'state machines',
    ],
    image: 'img-why-boilerplate@2x.png',
  }
};
