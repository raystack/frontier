module.exports = {
  docsSidebar: [
    'introduction',
    {
      type: "category",
      label: "Getting Started",
      collapsed: true,
      link: {
        type: "doc",
        id: "installation",
      },
      items: [
        'installation',
        'configurations',
        'admin-portal',
      ],
    },
    {
      type: "category",
      label: "Concepts",
      link: {
        type: "doc",
        id: "concepts/architecture",
      },
      items: [
        "concepts/architecture",
        "concepts/glossary",
      ],
    },
    {
      type: "category",
      label: "Tour",
      collapsed: true,
      link: {
        type: "doc",
        id: "tour/setup-idp-oidc",
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
    // {
    //   type: "category",
    //   label: "Authentication",
    //   collapsed: true,
    //   link: {
    //     type: "generated-index",
    //     title: "Overview",
    //     slug: "authentication/overview",
    //   },
    //   items: [
    //     "authn/intro",
    //   ],
    // },
    {
      type: "category",
      label: "Authorization",
      collapsed: true,
      items: [
        "authz/permission",
        "authz/role",
        "authz/policy",
      ],
    },
    {
      type: "category",
      label: "Tenants",
      collapsed: true,
      link: {
        type: "doc",
        id: "tenants/org",
      },
      items: [
        "tenants/org",
        "tenants/project",
        {
          type: "category",
          label: "Managing Tenants",
          collapsed: false,
          link: {
            type: "doc",
            id: "tenants/managing-organization",
          },
          items: [
            "tenants/managing-organization",
            "tenants/managing-project",
            "tenants/managing-resource",
          ],
        }
      ],
    },
    {
      type: "category",
      label: "Principals",
      collapsed: true,
      link: {
        type: "doc",
        id: "users/principal",
      },
      items: [
        "users/principal",
        {
          type: "category",
          label: "Managing Principals",
          collapsed: false,
          link: {
            type: "doc",
            id: "users/managing-user",
          },
          items: [
            "users/managing-user",
            "users/managing-group",
          ],
        }
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
      label: "Reference",
      link: {
        type: "generated-index",
        title: "Overview",
        slug: "reference/overview",
      },
      items: [
        "reference/configurations",
        "reference/cli",
        "reference/metaschemas",
        "reference/api-definitions",
        "reference/shell-autocomplete",
      ],
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