import { Dialog, Flex, Skeleton, Tabs, Text } from "@raystack/apsara";
import styles from "./apis.module.css";
import { useCallback, useEffect, useState } from "react";
import { api } from "~/api";
import type {
  V1Beta1ServiceUserToken,
  SearchOrganizationServiceUsersResponseOrganizationServiceUser,
  Frontierv1Beta1Project,
} from "~/api/frontier";
import dayjs from "dayjs";

interface ServiceUserDetailsDialogProps {
  onClose: () => void;
  serviceUser: SearchOrganizationServiceUsersResponseOrganizationServiceUser | null;
}

export const ServiceUserDetailsDialog = ({
  serviceUser,
  onClose,
}: ServiceUserDetailsDialogProps) => {
  const { id = "", org_id = "", title = "" } = serviceUser || {};
  const [projects, setProjects] = useState<Frontierv1Beta1Project[]>([]);
  const [tokens, setTokens] = useState<V1Beta1ServiceUserToken[]>([]);
  const [isProjectLoading, setIsProjectLoading] = useState(false);
  const [isTokenLoading, setIsTokenLoading] = useState(false);

  const fetchProjects = useCallback(async () => {
    if (!org_id || !id) return;
    setIsProjectLoading(true);
    try {
      const resp = await api?.frontierServiceListServiceUserProjects(
        org_id,
        id,
      );
      const list = resp?.data?.projects || [];
      setProjects(list || []);
    } catch (error) {
      console.error(error);
    } finally {
      setIsProjectLoading(false);
    }
  }, [org_id, id]);

  const fetchTokens = useCallback(async () => {
    if (!org_id || !id) return;
    setIsTokenLoading(true);
    try {
      const resp = await api?.frontierServiceListServiceUserTokens(org_id, id);
      const list = resp?.data?.tokens || [];
      setTokens(list || []);
    } catch (error) {
      console.error(error);
    } finally {
      setIsTokenLoading(false);
    }
  }, [org_id, id]);

  useEffect(() => {
    fetchProjects();
    fetchTokens();
  }, [fetchProjects, fetchTokens]);

  function onOpenChange(val: boolean) {
    if (!val) {
      onClose();
    }
  }

  return (
    <Dialog open={id !== ""} onOpenChange={onOpenChange}>
      <Dialog.Content className={styles["details-dialog"]}>
        <Dialog.Header>
          <Dialog.Title>{title}</Dialog.Title>
          <Dialog.CloseButton />
        </Dialog.Header>
        <Dialog.Body className={styles["dialog-body"]}>
          <Tabs.Root defaultValue="keys" className={styles["tab-root"]}>
            <Tabs.List>
              <Tabs.Trigger value="keys">
                API keys{" "}
                {!isTokenLoading && tokens.length > 0
                  ? `(${tokens.length})`
                  : ""}
              </Tabs.Trigger>
              <Tabs.Trigger value="projects">
                Projects{" "}
                {!isProjectLoading && projects.length > 0
                  ? `(${projects.length})`
                  : ""}
              </Tabs.Trigger>
            </Tabs.List>
            <Tabs.Content value="keys" className={styles["tab-content"]}>
              {isTokenLoading ? (
                <Skeleton
                  containerClassName={styles["skeleton"]}
                  height={24}
                  count={10}
                />
              ) : (
                <Flex direction="column">
                  {tokens.map((token) => (
                    <Flex
                      key={token.id}
                      direction="column"
                      className={styles["list-item"]}
                      gap={2}
                    >
                      <Text weight="medium">{token.title}</Text>
                      <Text size="micro">
                        {dayjs(token.created_at).format("YYYY-MM-DD")}
                      </Text>
                    </Flex>
                  ))}
                </Flex>
              )}
            </Tabs.Content>
            <Tabs.Content value="projects" className={styles["tab-content"]}>
              {isProjectLoading ? (
                <Skeleton
                  containerClassName={styles["skeleton"]}
                  height={24}
                  count={10}
                />
              ) : (
                <Flex direction="column">
                  {projects.map((project) => (
                    <Flex
                      key={project.id}
                      direction="column"
                      className={styles["list-item"]}
                      gap={2}
                    >
                      <Text weight="medium">{project.title}</Text>
                    </Flex>
                  ))}
                </Flex>
              )}
            </Tabs.Content>
          </Tabs.Root>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
