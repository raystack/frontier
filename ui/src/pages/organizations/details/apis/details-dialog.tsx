import { Dialog, Flex, Skeleton, Tabs, Text } from "@raystack/apsara/v1";
import styles from "./apis.module.css";
import { useCallback, useEffect, useState } from "react";
import { V1Beta1Project } from "@raystack/frontier";
import { api } from "~/api";
import { V1Beta1ServiceUserToken } from "~/api/frontier";
import dayjs from "dayjs";

interface ServiceUserDetailsDialogProps {
  serviceUserId: string;
  organizationId: string;
}

export const ServiceUserDetailsDialog = ({
  serviceUserId,
  organizationId,
}: ServiceUserDetailsDialogProps) => {
  const [projects, setProjects] = useState<V1Beta1Project[]>([]);
  const [tokens, setTokens] = useState<V1Beta1ServiceUserToken[]>([]);
  const [isProjectLoading, setIsProjectLoading] = useState(false);
  const [isTokenLoading, setIsTokenLoading] = useState(false);

  const fetchProjects = useCallback(async () => {
    setIsProjectLoading(true);
    try {
      const resp = await api?.frontierServiceListServiceUserProjects(
        organizationId,
        serviceUserId,
      );
      const list = resp?.data?.projects || [];
      setProjects(list || []);
    } catch (error) {
      console.error(error);
    } finally {
      setIsProjectLoading(false);
    }
  }, [organizationId, serviceUserId]);

  const fetchTokens = useCallback(async () => {
    setIsTokenLoading(true);
    try {
      const resp = await api?.frontierServiceListServiceUserTokens(
        organizationId,
        serviceUserId,
      );
      const list = resp?.data?.tokens || [];
      setTokens(list || []);
    } catch (error) {
      console.error(error);
    } finally {
      setIsTokenLoading(false);
    }
  }, [organizationId, serviceUserId]);

  useEffect(() => {
    fetchProjects();
    fetchTokens();
  }, [fetchProjects, fetchTokens]);

  const isLoading = isProjectLoading || isTokenLoading;

  return (
    <Dialog open>
      <Dialog.Content className={styles["details-dialog"]}>
        <Dialog.Header>
          <Dialog.Title>Service Account Details</Dialog.Title>
          <Dialog.CloseButton />
        </Dialog.Header>
        <Dialog.Body>
          <Tabs.Root defaultValue="keys" className={styles["tab-root"]}>
            <Tabs.List>
              <Tabs.Trigger value="keys">
                API keys {tokens.length > 0 ? `(${tokens.length})` : ""}
              </Tabs.Trigger>
              <Tabs.Trigger value="projects">
                Projects {projects.length > 0 ? `(${projects.length})` : ""}
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
