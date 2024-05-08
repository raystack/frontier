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
        'basics',
        'admin-portal',
        'admin-settings'
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
      ],
    },
    {
      type: "category",
      label: "Authentication",
      collapsed: true,
      link: {
        type: "doc",
        id: "authn/introduction",
      },
      items: [
        "authn/introduction",
        "authn/user",
        "authn/serviceuser",
        "authn/org-domain",
      ],
    },
    {
      type: "category",
      label: "Authorization",
      collapsed: true,
      link: {
        type: "doc",
        id: "authz/overview",
      },
      items: [
        "authz/overview",
        "authz/permission",
        "authz/role",
        "authz/policy",
        "authz/example",
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
        "users/group",
      ],
    },
    {
      type: "category",
      label: "Billing",
      collapsed: true,
      link: {
        type: "doc",
        id: "billing/introduction",
      },
      items: [
        "billing/introduction",
      ],
    },
    {
      type: "category",
      label: "APIs",
      link: {
        type: "doc",
        id: "apis/frontier-administration-api",
      },
      items: [
        require("./docs/apis/sidebar.js")
      ]
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
        "reference/smtp",
        "reference/webhook",
        "reference/api-auth",
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