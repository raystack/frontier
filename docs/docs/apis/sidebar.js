module.exports = [
  {
    "type": "doc",
    "id": "apis/frontier-administration-api"
  },
  {
    "type": "category",
    "label": "User",
    "link": {
      "type": "generated-index",
      "title": "User",
      "slug": "/category/apis/user"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/admin-service-list-all-users",
        "label": "List all users",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-search-user-organizations",
        "label": "Search users organizations",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-search-user-projects",
        "label": "Search projects of a user within an org",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-search-users",
        "label": "Search users",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-users",
        "label": "List public users",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-user",
        "label": "Create user",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-user",
        "label": "Get user",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-user",
        "label": "Delete user",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-update-user",
        "label": "Update user",
        "className": "api-method put"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-disable-user",
        "label": "Disable user",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-enable-user",
        "label": "Enable user",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-user-groups",
        "label": "List user groups",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-user-invitations",
        "label": "List user invitations",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-organizations-by-user",
        "label": "Get user organizations",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-projects-by-user",
        "label": "Get user projects",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-current-user",
        "label": "Get current user",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-update-current-user",
        "label": "Update current user",
        "className": "api-method put"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-current-user-groups",
        "label": "List my groups",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-current-user-invitations",
        "label": "List user invitations",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-organizations-by-current-user",
        "label": "Get my organizations",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-projects-by-current-user",
        "label": "Get my projects",
        "className": "api-method get"
      }
    ]
  },
  {
    "type": "category",
    "label": "Group",
    "link": {
      "type": "generated-index",
      "title": "Group",
      "slug": "/category/apis/group"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/admin-service-list-groups",
        "label": "List all groups",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-organization-groups",
        "label": "List organization groups",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-group",
        "label": "Create group",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-group",
        "label": "Get group",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-group",
        "label": "Delete group",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-update-group",
        "label": "Update group",
        "className": "api-method put"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-disable-group",
        "label": "Disable group",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-enable-group",
        "label": "Enable group",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-group-users",
        "label": "List group users",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-add-group-users",
        "label": "Add group user",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-remove-group-user",
        "label": "Remove group user",
        "className": "api-method delete"
      }
    ]
  },
  {
    "type": "category",
    "label": "Organization",
    "link": {
      "type": "generated-index",
      "title": "Organization",
      "slug": "/category/apis/organization"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/admin-service-list-all-organizations",
        "label": "List all organizations",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-admin-create-organization",
        "label": "Create organization as superadmin",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-search-organization-invoices",
        "label": "Search organization invoices",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-search-organization-projects",
        "label": "Search organization projects",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-search-organization-service-user-credentials",
        "label": "Search organization service user credentials",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-search-organization-tokens",
        "label": "Search organization tokens",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-search-organization-users",
        "label": "Search organization users",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-search-organizations",
        "label": "Search organizations",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-organizations",
        "label": "List organizations",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-organization",
        "label": "Create organization",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-organization",
        "label": "Get organization",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-organization",
        "label": "Delete organization",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-update-organization",
        "label": "Update organization",
        "className": "api-method put"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-organization-admins",
        "label": "List organization admins",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-disable-organization",
        "label": "Disable organization",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-enable-organization",
        "label": "Enable organization",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-organization-projects",
        "label": "Get organization projects",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-organization-service-users",
        "label": "List organization service users",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-organization-users",
        "label": "List organization users",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-add-organization-users",
        "label": "Add organization user",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-remove-organization-user",
        "label": "Remove organization user",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-organization-domains",
        "label": "List org domains",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-organization-domain",
        "label": "Create org domain",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-organization-domain",
        "label": "Get org domain",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-organization-domain",
        "label": "Delete org domain",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-verify-organization-domain",
        "label": "Verify org domain",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-organization-invitations",
        "label": "List pending invitations",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-organization-invitation",
        "label": "Invite user",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-organization-invitation",
        "label": "Get pending invitation",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-organization-invitation",
        "label": "Delete pending invitation",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-accept-organization-invitation",
        "label": "Accept pending invitation",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-join-organization",
        "label": "Join organization",
        "className": "api-method post"
      }
    ]
  },
  {
    "type": "category",
    "label": "Project",
    "link": {
      "type": "generated-index",
      "title": "Project",
      "slug": "/category/apis/project"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/admin-service-list-projects",
        "label": "List all projects",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-search-project-users",
        "label": "Search Project users",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-project",
        "label": "Create project",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-project",
        "label": "Get project",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-project",
        "label": "Delete Project",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-update-project",
        "label": "Update project",
        "className": "api-method put"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-project-admins",
        "label": "List project admins",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-disable-project",
        "label": "Disable project",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-enable-project",
        "label": "Enable project",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-project-groups",
        "label": "List project groups",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-project-service-users",
        "label": "List project serviceusers",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-project-users",
        "label": "List project users",
        "className": "api-method get"
      }
    ]
  },
  {
    "type": "category",
    "label": "Relation",
    "link": {
      "type": "generated-index",
      "title": "Relation",
      "slug": "/category/apis/relation"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/admin-service-list-relations",
        "label": "List all relations",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-relation",
        "label": "Create relation",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-relation",
        "label": "Get relation",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-relation",
        "label": "Delete relation",
        "className": "api-method delete"
      }
    ]
  },
  {
    "type": "category",
    "label": "Resource",
    "link": {
      "type": "generated-index",
      "title": "Resource",
      "slug": "/category/apis/resource"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/admin-service-list-resources",
        "label": "List all resources",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-project-resources",
        "label": "Get all resources",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-project-resource",
        "label": "Create resource",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-project-resource",
        "label": "Get resource",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-project-resource",
        "label": "Delete resource",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-update-project-resource",
        "label": "Update resource",
        "className": "api-method put"
      }
    ]
  },
  {
    "type": "category",
    "label": "Policy",
    "link": {
      "type": "generated-index",
      "title": "Policy",
      "slug": "/category/apis/policy"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-list-policies",
        "label": "List all policies",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-policy",
        "label": "Create policy",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-policy",
        "label": "Get policy",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-policy",
        "label": "Delete Policy",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-update-policy",
        "label": "Update policy",
        "className": "api-method put"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-policy-for-project",
        "label": "Create Policy for Project",
        "className": "api-method post"
      }
    ]
  },
  {
    "type": "category",
    "label": "Role",
    "link": {
      "type": "generated-index",
      "title": "Role",
      "slug": "/category/apis/role"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-list-organization-roles",
        "label": "List organization roles",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-organization-role",
        "label": "Create organization role",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-organization-role",
        "label": "Get organization role",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-organization-role",
        "label": "Delete organization role",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-update-organization-role",
        "label": "Update organization role",
        "className": "api-method put"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-roles",
        "label": "List platform roles",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-create-role",
        "label": "Create platform role",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-delete-role",
        "label": "Delete platform role",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-update-role",
        "label": "Update role",
        "className": "api-method put"
      }
    ]
  },
  {
    "type": "category",
    "label": "Permission",
    "link": {
      "type": "generated-index",
      "title": "Permission",
      "slug": "/category/apis/permission"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-list-permissions",
        "label": "Get all permissions",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-create-permission",
        "label": "Create platform permission",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-permission",
        "label": "Get permission",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-delete-permission",
        "label": "Delete platform permission",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-update-permission",
        "label": "Update platform permission",
        "className": "api-method put"
      }
    ]
  },
  {
    "type": "category",
    "label": "Authz",
    "link": {
      "type": "generated-index",
      "title": "Authz",
      "slug": "/category/apis/authz"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-get-jw-ks-2",
        "label": "Get well known JWKs",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-check-federated-resource-permission",
        "label": "Check",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-jw-ks",
        "label": "Get well known JWKs",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-batch-check-permission",
        "label": "Batch check",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-check-resource-permission",
        "label": "Check",
        "className": "api-method post"
      }
    ]
  },
  {
    "type": "category",
    "label": "Billing",
    "link": {
      "type": "generated-index",
      "title": "Billing",
      "slug": "/category/apis/billing"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/admin-service-list-all-billing-accounts",
        "label": "List all billing accounts",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-revert-billing-usage",
        "label": "Revert billing usage",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-get-billing-account-details",
        "label": "Get billing account details",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-update-billing-account-details",
        "label": "Update billing account details",
        "className": "api-method put"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-update-billing-account-limits",
        "label": "Update billing account limits",
        "className": "menu__list-item--deprecated api-method put"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-billing-accounts",
        "label": "List billing accounts",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-billing-account",
        "label": "Create billing account",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-billing-account",
        "label": "Get billing account",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-billing-account",
        "label": "Delete billing account",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-update-billing-account",
        "label": "Update billing account",
        "className": "api-method put"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-billing-balance",
        "label": "Get billing balance",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-disable-billing-account",
        "label": "Disable billing account",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-enable-billing-account",
        "label": "Enable billing account",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-has-trialed",
        "label": "Has trialed",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-register-billing-account",
        "label": "Register billing account to provider",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-revert-billing-usage-2",
        "label": "Revert billing usage",
        "className": "api-method post"
      }
    ]
  },
  {
    "type": "category",
    "label": "Invoice",
    "link": {
      "type": "generated-index",
      "title": "Invoice",
      "slug": "/category/apis/invoice"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/admin-service-list-all-invoices",
        "label": "List all invoices",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-generate-invoices",
        "label": "Trigger invoice generation",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-invoices",
        "label": "List invoices",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-upcoming-invoice",
        "label": "Get upcoming invoice",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-invoices-2",
        "label": "List invoices",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-upcoming-invoice-2",
        "label": "Get upcoming invoice",
        "className": "api-method get"
      }
    ]
  },
  {
    "type": "category",
    "label": "Checkout",
    "link": {
      "type": "generated-index",
      "title": "Checkout",
      "slug": "/category/apis/checkout"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/admin-service-delegated-checkout",
        "label": "Checkout a product or subscription",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-delegated-checkout-2",
        "label": "Checkout a product or subscription",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-checkouts",
        "label": "List checkouts",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-checkout",
        "label": "Checkout a product or subscription",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-checkout",
        "label": "Get checkout",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-checkouts-2",
        "label": "List checkouts",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-checkout-2",
        "label": "Checkout a product or subscription",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-checkout-2",
        "label": "Get checkout",
        "className": "api-method get"
      }
    ]
  },
  {
    "type": "category",
    "label": "OrganizationKyc",
    "link": {
      "type": "generated-index",
      "title": "OrganizationKyc",
      "slug": "/category/apis/organization-kyc"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/admin-service-set-organization-kyc",
        "label": "Set KYC information of an organization",
        "className": "api-method put"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-list-organizations-kyc",
        "label": "List KYC information of all organizations",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-organization-kyc",
        "label": "Get KYC info of an organization",
        "className": "api-method get"
      }
    ]
  },
  {
    "type": "category",
    "label": "Platform",
    "link": {
      "type": "generated-index",
      "title": "Platform",
      "slug": "/category/apis/platform"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/admin-service-list-platform-users",
        "label": "List platform users",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-add-platform-user",
        "label": "Add platform user",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-remove-platform-user",
        "label": "Remove platform user",
        "className": "api-method post"
      }
    ]
  },
  {
    "type": "category",
    "label": "Prospect",
    "link": {
      "type": "generated-index",
      "title": "Prospect",
      "slug": "/category/apis/prospect"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/admin-service-create-prospect",
        "label": "Create prospect",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-get-prospect",
        "label": "Get prospect",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-delete-prospect",
        "label": "Delete prospect",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-update-prospect",
        "label": "Update prospect",
        "className": "api-method put"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-list-prospects",
        "label": "List prospects",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-prospect-public",
        "label": "Create prospect",
        "className": "api-method post"
      }
    ]
  },
  {
    "type": "category",
    "label": "Webhook",
    "link": {
      "type": "generated-index",
      "title": "Webhook",
      "slug": "/category/apis/webhook"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/admin-service-list-webhooks",
        "label": "List webhooks",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-create-webhook",
        "label": "Create webhook",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-delete-webhook",
        "label": "Delete webhook",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-update-webhook",
        "label": "Update webhook",
        "className": "api-method put"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-billing-webhook-callback",
        "label": "Accept Billing webhook",
        "className": "api-method post"
      }
    ]
  },
  {
    "type": "category",
    "label": "Authn",
    "link": {
      "type": "generated-index",
      "title": "Authn",
      "slug": "/category/apis/authn"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-list-auth-strategies",
        "label": "List authentication strategies",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-auth-callback",
        "label": "Callback from a strategy",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-auth-callback-2",
        "label": "Callback from a strategy",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-auth-logout",
        "label": "Logout from a strategy",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-auth-logout-2",
        "label": "Logout from a strategy",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-authenticate",
        "label": "Authenticate with a strategy",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-authenticate-2",
        "label": "Authenticate with a strategy",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-auth-token",
        "label": "Generate access token by given credentials",
        "className": "api-method post"
      }
    ]
  },
  {
    "type": "category",
    "label": "Feature",
    "link": {
      "type": "generated-index",
      "title": "Feature",
      "slug": "/category/apis/feature"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-list-features",
        "label": "List features",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-feature",
        "label": "Create feature",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-feature",
        "label": "Get feature",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-update-feature",
        "label": "Update feature",
        "className": "api-method put"
      }
    ]
  },
  {
    "type": "category",
    "label": "Plan",
    "link": {
      "type": "generated-index",
      "title": "Plan",
      "slug": "/category/apis/plan"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-list-plans",
        "label": "List plans",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-plan",
        "label": "Create plan",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-plan",
        "label": "Get plan",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-update-plan",
        "label": "Update plan",
        "className": "api-method put"
      }
    ]
  },
  {
    "type": "category",
    "label": "Product",
    "link": {
      "type": "generated-index",
      "title": "Product",
      "slug": "/category/apis/product"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-list-products",
        "label": "List products",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-product",
        "label": "Create product",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-product",
        "label": "Get product",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-update-product",
        "label": "Update product",
        "className": "api-method put"
      }
    ]
  },
  {
    "type": "category",
    "label": "Preference",
    "link": {
      "type": "generated-index",
      "title": "Preference",
      "slug": "/category/apis/preference"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-list-group-preferences",
        "label": "List group preferences",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-group-preferences",
        "label": "Create group preferences",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-organization-preferences",
        "label": "List organization preferences",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-organization-preferences",
        "label": "Create organization preferences",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-list-preferences",
        "label": "List platform preferences",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/admin-service-create-preferences",
        "label": "Create platform preferences",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-describe-preferences",
        "label": "Describe preferences",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-project-preferences",
        "label": "List project preferences",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-project-preferences",
        "label": "Create project preferences",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-user-preferences",
        "label": "List user preferences",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-user-preferences",
        "label": "Create user preferences",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-current-user-preferences",
        "label": "List current user preferences",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-current-user-preferences",
        "label": "Create current user preferences",
        "className": "api-method post"
      }
    ]
  },
  {
    "type": "category",
    "label": "MetaSchema",
    "link": {
      "type": "generated-index",
      "title": "MetaSchema",
      "slug": "/category/apis/meta-schema"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-list-meta-schemas",
        "label": "List metaschemas",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-meta-schema",
        "label": "Create metaschema",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-meta-schema",
        "label": "Get metaschema",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-meta-schema",
        "label": "Delete metaschema",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-update-meta-schema",
        "label": "Update metaschema",
        "className": "api-method put"
      }
    ]
  },
  {
    "type": "category",
    "label": "Namespace",
    "link": {
      "type": "generated-index",
      "title": "Namespace",
      "slug": "/category/apis/namespace"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-list-namespaces",
        "label": "Get all namespaces",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-namespace",
        "label": "Get namespace",
        "className": "api-method get"
      }
    ]
  },
  {
    "type": "category",
    "label": "AuditLog",
    "link": {
      "type": "generated-index",
      "title": "AuditLog",
      "slug": "/category/apis/audit-log"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-list-organization-audit-logs",
        "label": "List audit logs",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-organization-audit-logs",
        "label": "Create audit log",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-organization-audit-log",
        "label": "Get audit log",
        "className": "api-method get"
      }
    ]
  },
  {
    "type": "category",
    "label": "Entitlement",
    "link": {
      "type": "generated-index",
      "title": "Entitlement",
      "slug": "/category/apis/entitlement"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-check-feature-entitlement",
        "label": "Check entitlement",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-check-feature-entitlement-2",
        "label": "Check entitlement",
        "className": "api-method post"
      }
    ]
  },
  {
    "type": "category",
    "label": "Transaction",
    "link": {
      "type": "generated-index",
      "title": "Transaction",
      "slug": "/category/apis/transaction"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-total-debited-transactions",
        "label": "Sum of amount of debited transactions including refunds",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-billing-transactions",
        "label": "List billing transactions",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-billing-transactions-2",
        "label": "List billing transactions",
        "className": "api-method get"
      }
    ]
  },
  {
    "type": "category",
    "label": "Subscription",
    "link": {
      "type": "generated-index",
      "title": "Subscription",
      "slug": "/category/apis/subscription"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-list-subscriptions",
        "label": "List subscriptions",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-subscription",
        "label": "Get subscription",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-update-subscription",
        "label": "Update subscription",
        "className": "api-method put"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-cancel-subscription",
        "label": "Cancel subscription",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-change-subscription",
        "label": "Change subscription plan",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-subscriptions-2",
        "label": "List subscriptions",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-subscription-2",
        "label": "Get subscription",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-update-subscription-2",
        "label": "Update subscription",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-cancel-subscription-2",
        "label": "Cancel subscription",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-change-subscription-2",
        "label": "Change subscription plan",
        "className": "api-method post"
      }
    ]
  },
  {
    "type": "category",
    "label": "Usage",
    "link": {
      "type": "generated-index",
      "title": "Usage",
      "slug": "/category/apis/usage"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-create-billing-usage",
        "label": "Create billing usage",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-billing-usage-2",
        "label": "Create billing usage",
        "className": "api-method post"
      }
    ]
  },
  {
    "type": "category",
    "label": "ServiceUser",
    "link": {
      "type": "generated-index",
      "title": "ServiceUser",
      "slug": "/category/apis/service-user"
    },
    "items": [
      {
        "type": "doc",
        "id": "apis/frontier-service-create-service-user",
        "label": "Create service user",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-service-user",
        "label": "Get service user",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-service-user",
        "label": "Delete service user",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-service-user-jw-ks",
        "label": "List service user keys",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-service-user-jwk",
        "label": "Create service user public/private key pair",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-get-service-user-jwk",
        "label": "Get service user key",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-service-user-jwk",
        "label": "Delete service user key",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-service-user-projects",
        "label": "List service sser projects",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-service-user-credentials",
        "label": "List service user credentials",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-service-user-credential",
        "label": "Create service user client credentials",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-service-user-credential",
        "label": "Delete service user credentials",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-service-user-tokens",
        "label": "List service user tokens",
        "className": "api-method get"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-create-service-user-token",
        "label": "Create service user token",
        "className": "api-method post"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-delete-service-user-token",
        "label": "Delete service user token",
        "className": "api-method delete"
      },
      {
        "type": "doc",
        "id": "apis/frontier-service-list-service-users",
        "label": "List org service users",
        "className": "api-method get"
      }
    ]
  }
]
