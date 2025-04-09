import { Dialog } from "@raystack/apsara/v1";
import { useCallback, useEffect, useState } from "react";
import Skeleton from "react-loading-skeleton";
import { api } from "~/api";
import { V1Beta1Project } from "~/api/frontier";
import styles from "./members.module.css";

export const ProjectMembersDialog = ({
  projectId,
  onClose,
}: {
  projectId: string;
  onClose: () => void;
}) => {
  const [project, setProject] = useState<V1Beta1Project>({});
  const [isProjectLoading, setIsProjectLoading] = useState<boolean>(false);

  const fetchProject = useCallback(async (id: string) => {
    setIsProjectLoading(true);
    try {
      const resp = await api?.frontierServiceGetProject(id);
      const project = resp.data?.project || {};
      setProject(project);
    } catch (error) {
      console.error(error);
    } finally {
      setIsProjectLoading(false);
    }
  }, []);

  useEffect(() => {
    if (projectId) {
      fetchProject(projectId);
    }
  }, [projectId, fetchProject]);

  return (
    <Dialog open onOpenChange={onClose}>
      <Dialog.Content className={styles["dialog-content"]}>
        <Dialog.Header>
          {isProjectLoading ? (
            <Skeleton containerClassName={styles["flex1"]} width={"200px"} />
          ) : (
            <Dialog.Title>{project.title}</Dialog.Title>
          )}
          <Dialog.CloseButton data-test-id="close-button" />
        </Dialog.Header>
      </Dialog.Content>
    </Dialog>
  );
};
