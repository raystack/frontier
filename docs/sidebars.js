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
      collapsed: false,
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
      collapsed: false,
      items: [
        "guides/check-permission",
        "guides/manage-resources",
      ],
    },
    {
      type: "category",
      label: "Reference",
      items: [
        "reference/configurations",
      ],
    },
  ],
};