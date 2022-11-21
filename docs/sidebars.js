module.exports = {
  docsSidebar: [
    'introduction',
    'installation',
    {
      type: "category",
      label: "Tour",
      collapsed: false,
      items: [
        "tour/setup-server",
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
      label: "Concepts",
      items: [
        "concepts/architecture",
        "concepts/glossary",
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