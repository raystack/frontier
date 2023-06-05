module.exports = {
  docsSidebar: [
    'introduction',
    {
      type: "category",
      label: "Getting Started",
      collapsed: true,
      items: [
        'installation',
        'configurations',
        'admin-portal',
      ],
    },
    {
      type: "category",
      label: "Concepts",
      items: [
        "concepts/architecture",
        "concepts/org",
        "concepts/user",
        "concepts/project",
        "concepts/policy",
        "concepts/role",
        "concepts/glossary",
      ],
    },
    {
      type: "category",
      label: "Tour",
      collapsed: true,
      link: {
        type: "doc",
        id: "tour/intro",
      },
      items: [
        "tour/setup-idp-oidc",
        "tour/creating-user",
        "tour/creating-organization",
        "tour/creating-project",
        "tour/creating-group",
        "tour/add-to-group",
        "tour/check-permissions",
        "tour/shield-as-proxy"
      ],
    },
    {
      type: "category",
      label: "Guides",
      link: {
        type: "generated-index",
        title: "Overview",
        slug: "guides/overview",
      },
      collapsed: true,
      items: [
        "guides/managing-organization",
        "guides/managing-project",
        "guides/managing-group",
        "guides/check-permission",
        "guides/managing-resource",
        "guides/managing-user",
        "guides/managing-metaschemas",
      ],
    },
    {
      type: "category",
      label: "Reference",
      link: {
        type: "generated-index",
        title: "Overview",
        slug: "reference/overview",
      },
      items: [
        "reference/configurations",
        "reference/cli",
        "reference/api-definitions",
        "reference/shell-autocomplete",
      ],
    },
    {
      type: "category",
      label: "APIs",
      link: {
        type: "doc",
        id: "apis/shield-administration-api",
      },
      items: require("./docs/apis/sidebar.js"),
    },
    {
      type: "category",
      label: "Contribute",
      link: {
        type: "doc",
        id: "contribution/contribute",
      },
      items: [
        "contribution/contribute",
      ]
    }
  ],
};