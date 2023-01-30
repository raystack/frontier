module.exports = {
  docsSidebar: [
    'introduction',
    'installation',
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
        "tour/what-is-in-shield",
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
        "guides/adding-metadata-key",
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
  ],
};