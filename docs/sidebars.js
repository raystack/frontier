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
      ],
    },
    {
      type: "category",
      label: "Concepts",
      items: [
        "concepts/architecture",
        "concepts/glossary",
      ],
    },
    {
      type: "category",
      label: "Tour",
      collapsed: true,
      items: [
        "tour/intro",
        "tour/setup-server",
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
      collapsed: true,
      items: [
        "guides/overview",
        "guides/managing-organization",
        "guides/managing-project",
        "guides/managing-group",
        "guides/check-permission",
        "guides/managing-resource",
        "guides/managing-relation",
        "guides/managing-user",
        "guides/managing-metaschemas",
      ],
    },
    {
      type: "category",
      label: "Reference",
      items: [
        "reference/configurations",
        "reference/api",
        "reference/cli"
      ],
    },
    {
      type: "category",
      label: "Contribute",
      items: [
        "contribution/contribute",
      ],
    }
  ],
};