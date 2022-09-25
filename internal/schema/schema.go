package schema

type NamespaceConfig struct {
	InheritedNamespaces []string
	Roles               map[string][]string
	Permissions         map[string][]string
}

type NamespaceConfigMapType map[string]NamespaceConfig

func RunMigrations() {

}
