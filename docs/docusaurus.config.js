const lightCodeTheme = require('prism-react-renderer/themes/dracula');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

// With JSDoc @type annotations, IDEs can provide config autocompletion
/** @type {import('@docusaurus/types').DocusaurusConfig} */
(module.exports = {
  title: 'Frontier',
  tagline: 'Identity and authorization for your APIs',
  url: 'https://raystack.github.io/',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',
  organizationName: 'Raystack',
  projectName: 'frontier',

  presets: [
    [
      '@docusaurus/preset-classic',
      ({
        gtag: {
          trackingID: 'G-XXX',
        },
        docs: {
          routeBasePath: '/',
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/raystack/frontier/edit/master/docs/',
          sidebarCollapsed: true,
          breadcrumbs: false,
          docLayoutComponent: "@theme/DocPage",
          docItemComponent: "@theme/ApiItem"
        },
        blog: false,
        theme: {
          customCss: [
            require.resolve('./src/css/theme.css'),
            require.resolve('./src/css/custom.css'),
            require.resolve('./src/css/icons.css'),
          ],
        },
      })
    ],
  ],
  plugins: [
    [
      'docusaurus-plugin-openapi-docs',
      {
        id: "apiDocs",
        docsPluginId: "classic",
        config: {
          auth: {
            specPath: "../proto/apidocs.swagger.yaml",
            outputDir: "docs/apis",
            sidebarOptions: {
              groupPathsBy: "tag",
            },
            hideSendButton: false,
          }
        }
      },
    ],
  ],
  themes: ["docusaurus-theme-openapi-docs"],
  themeConfig:
    ({
      colorMode: {
        defaultMode: 'light',
        respectPrefersColorScheme: true,
      },
      docs: {
        sidebar: {
          autoCollapseCategories: true,
          hideable: true,
        },
      },
      navbar: {
        title: 'Frontier',
        logo: { src: 'img/frontier.svg', },
        hideOnScroll: true,
        items: [
          {
            type: 'doc',
            docId: 'introduction',
            position: 'right',
            label: 'Documentation',
          },
          { to: 'support', label: 'Support', position: 'right' },
          {
            href: 'https://bit.ly/2RzPbtn',
            position: 'right',
            className: 'header-slack-link',
          },
          {
            href: 'https://github.com/raystack/frontier',
            className: 'navbar-item-github',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'light',
        copyright: `Open DataOps Foundation © ${new Date().getFullYear()}`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
      announcementBar: {
        id: 'star-repo',
        content: '⭐️ If you like Frontier, give it a star on <a target="_blank" rel="noopener noreferrer" href="https://github.com/raystack/frontier">GitHub</a>! ⭐',
        backgroundColor: '#222',
        textColor: '#eee',
        isCloseable: true,
      },
    }),
});
