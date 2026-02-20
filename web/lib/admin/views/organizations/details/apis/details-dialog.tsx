import { Dialog, Flex, Skeleton, Tabs, Text, toast } from "@raystack/apsara";
import styles from "./apis.module.css";
import { useCallback, useEffect, useMemo } from "react";
import dayjs from "dayjs";
import { useQuery } from "@connectrpc/connect-query";
import {
  FrontierServiceQueries,
  ListServiceUserProjectsRequestSchema,
  ListServiceUserTokensRequestSchema,
  type SearchOrganizationServiceUsersResponse_OrganizationServiceUser
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { timestampToDayjs } from "../../../../utils/connect-timestamp";

interface ServiceUserDetailsDialogProps {
  onClose: () => void;
  serviceUser: SearchOrganizationServiceUsersResponse_OrganizationServiceUser | null;
}

export const ServiceUserDetailsDialog = ({
  serviceUser,
  onClose,
}: ServiceUserDetailsDialogProps) => {
  const { id = "", orgId = "", title = "" } = serviceUser || {};

  const projectsRequest = useMemo(
    () =>
      create(ListServiceUserProjectsRequestSchema, {
        orgId: orgId,
        id: id,
      }),
    [orgId, id],
  );

  const tokensRequest = useMemo(
    () =>
      create(ListServiceUserTokensRequestSchema, {
        orgId: orgId,
        id: id,
      }),
    [orgId, id],
  );

  const {
    data: projects = [],
    isLoading: isProjectLoading,
    error: projectError,
  } = useQuery(FrontierServiceQueries.listServiceUserProjects, projectsRequest, {
    enabled: !!orgId && !!id,
    select: (data) => data?.projects || [],
  });

  const {
    data: tokens = [],
    isLoading: isTokenLoading,
    error: tokenError,
  } = useQuery(FrontierServiceQueries.listServiceUserTokens, tokensRequest, {
    enabled: !!orgId && !!id,
    select: (data) => data?.tokens || [],
  });

  useEffect(() => {
    if (projectError) {
      toast.error("Something went wrong", {
        description: "Unable to fetch projects",
      });
      console.error("Unable to fetch projects:", projectError);
    }
    if (tokenError) {
      toast.error("Something went wrong", {
        description: "Unable to fetch tokens",
      });
      console.error("Unable to fetch tokens:", tokenError);
    }
  }, [projectError, tokenError]);

  const onOpenChange = useCallback(
    (val: boolean) => {
      if (!val) {
        onClose();
      }
    },
    [onClose],
  );

  return (
    <Dialog open={id !== ""} onOpenChange={onOpenChange}>
      <Dialog.Content className={styles["details-dialog"]}>
        <Dialog.Header>
          <Dialog.Title>{title}</Dialog.Title>
          <Dialog.CloseButton />
        </Dialog.Header>
        <Dialog.Body className={styles["dialog-body"]}>
          <Tabs defaultValue="keys" className={styles["tab-root"]}>
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
                        {timestampToDayjs(token.createdAt)?.format("YYYY-MM-DD")}
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
          </Tabs>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
