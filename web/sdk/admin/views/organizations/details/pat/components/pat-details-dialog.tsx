import { useMemo } from "react";
import {
  Dialog,
  Flex,
  Skeleton,
  Tabs,
  Text,
  EmptyState,
} from "@raystack/apsara";
import { useQuery } from "@connectrpc/connect-query";
import {
  FrontierServiceQueries,
  type Project,
  type Role,
  type SearchOrganizationPATsResponse_OrganizationPAT,
} from "@raystack/proton/frontier";
import { useTerminology } from "../../../../../hooks/useTerminology";
import { SCOPES } from "../../../../../utils/constants";
import styles from "./pat-details-dialog.module.css";

interface PatDetailsDialogProps {
  pat: SearchOrganizationPATsResponse_OrganizationPAT | null;
  projectsMap: Record<string, Project>;
  onClose: () => void;
}

export function PatDetailsDialog({
  pat,
  projectsMap,
  onClose,
}: PatDetailsDialogProps) {
  const t = useTerminology();
  const open = pat !== null;

  const projectScope = useMemo(
    () => pat?.scopes?.find((s) => s.resourceType === SCOPES.PROJECT),
    [pat],
  );
  const orgScope = useMemo(
    () => pat?.scopes?.find((s) => s.resourceType === SCOPES.ORG),
    [pat],
  );

  const scopeProjectIds = projectScope?.resourceIds ?? [];
  const scopeProjects = useMemo(
    () =>
      scopeProjectIds
        .map((id) => projectsMap[id])
        .filter((p): p is Project => Boolean(p)),
    [scopeProjectIds, projectsMap],
  );

  const { data: rolesData, isLoading: isRolesLoading } = useQuery(
    FrontierServiceQueries.listRolesForPAT,
    { scopes: [SCOPES.ORG, SCOPES.PROJECT] },
    { enabled: open },
  );
  const rolesMap = useMemo(() => {
    const map: Record<string, Role> = {};
    (rolesData?.roles ?? []).forEach((role) => {
      if (role.id) map[role.id] = role;
    });
    return map;
  }, [rolesData]);

  const orgRole = orgScope?.roleId ? rolesMap[orgScope.roleId] : undefined;
  const projectRole = projectScope?.roleId
    ? rolesMap[projectScope.roleId]
    : undefined;

  const projectCount = scopeProjects.length;
  const roleEntries = [
    ...(orgScope
      ? [
        {
          id: `org-${orgScope.roleId}`,
          roleTitle: orgRole?.title || orgRole?.name || orgScope.roleId,
          scopeLabel: t.organization({ case: "capital" }),
        },
      ]
      : []),
    ...(projectScope
      ? [
        {
          id: `project-${projectScope.roleId}`,
          roleTitle:
            projectRole?.title || projectRole?.name || projectScope.roleId,
          scopeLabel: t.project({ plural: true, case: "capital" }),
        },
      ]
      : []),
  ];

  const onOpenChange = (val: boolean) => {
    if (!val) onClose();
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <Dialog.Content className={styles["details-dialog"]}>
        <Dialog.Header>
          <Dialog.Title>{pat?.title || ""}</Dialog.Title>
          <Dialog.CloseButton data-test-id="frontier-sdk-pat-details-dialog-close-btn" />
        </Dialog.Header>
        <Dialog.Body className={styles["dialog-body"]}>
          <Tabs defaultValue="projects" className={styles["tab-root"]}>
            <Tabs.List>
              <Tabs.Trigger value="projects">
                {t.project({ plural: true, case: "capital" })} ({projectCount})
              </Tabs.Trigger>
              <Tabs.Trigger value="roles">
                Roles ({roleEntries.length})
              </Tabs.Trigger>
            </Tabs.List>
            <Tabs.Content value="projects" className={styles["tab-content"]}>
              {scopeProjects.length === 0 ? null : (
                <Flex direction="column">
                  {scopeProjects.map((project) => (
                    <div key={project.id} className={styles["list-item"]}>
                      <Text
                        size="small"
                        weight="medium"
                        className={styles["list-item-text"]}
                      >
                        {project.title || project.name}
                      </Text>
                    </div>
                  ))}
                </Flex>
              )}
            </Tabs.Content>
            <Tabs.Content value="roles" className={styles["tab-content"]}>
              {isRolesLoading ? (
                <Skeleton
                  containerClassName={styles["skeleton"]}
                  height={20}
                  count={4}
                />
              ) : roleEntries.length === 0 ? (
                null
              ) : (
                <Flex direction="column">
                  {roleEntries.map((entry) => (
                    <Flex
                      key={entry.id}
                      direction="column"
                      gap={2}
                      className={styles["list-item"]}
                    >
                      <Text
                        size="small"
                        weight="medium"
                        className={styles["list-item-text"]}
                      >
                        {entry.roleTitle}
                      </Text>
                      <Text
                        size="micro"
                        variant="tertiary"
                        className={styles["list-item-text"]}
                      >
                        {entry.scopeLabel}
                      </Text>
                    </Flex>
                  ))}
                </Flex>
              )}
            </Tabs.Content>
          </Tabs>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog >
  );
}
