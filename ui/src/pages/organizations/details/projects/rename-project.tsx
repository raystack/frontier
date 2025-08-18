import { Button, Dialog, Flex, InputField, toast } from "@raystack/apsara";
import { api } from "~/api";
import { SearchOrganizationProjectsResponseOrganizationProject } from "~/api/frontier";

import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm, Controller } from "react-hook-form";
import React from "react";

const projectRenameSchema = z.object({
  title: z.string(),
  name: z.string(),
  org_id: z.string(),
});

type ProjectRenameSchema = z.infer<typeof projectRenameSchema>;

interface RenameProjectDialogProps {
  onClose: () => void;
  project: SearchOrganizationProjectsResponseOrganizationProject;
  onRename: (
    project: SearchOrganizationProjectsResponseOrganizationProject,
  ) => void;
}

export function RenameProjectDialog({
  onClose,
  project,
  onRename,
}: RenameProjectDialogProps) {
  const preventClickPropagation = (event: React.MouseEvent<HTMLDivElement>) => {
    event.stopPropagation();
  };

  const {
    handleSubmit,
    control,
    formState: { isSubmitting },
  } = useForm<ProjectRenameSchema>({
    defaultValues: {
      title: project?.title,
      name: project?.name,
      org_id: project?.organization_id,
    },
    resolver: zodResolver(projectRenameSchema),
  });

  const submit = async (data: ProjectRenameSchema) => {
    try {
      const resp = await api?.frontierServiceUpdateProject(
        project?.id || "",
        data,
      );
      const newProject = resp.data.project;

      if (newProject) {
        toast.success("Project renamed successfully");
        onRename({ ...project, ...newProject });
      }
    } catch (error) {
      toast.error("Failed to rename project");
      console.error(error);
    }
  };

  return (
    <Dialog open onOpenChange={onClose}>
      <Dialog.Content
        width={400}
        onClick={preventClickPropagation}
        data-test-id="rename-project-dialog"
      >
        <form onSubmit={handleSubmit(submit)}>
          <Dialog.Header>
            <Dialog.Title>Rename Project</Dialog.Title>
            <Dialog.CloseButton data-test-id="rename-project-close-button" />
          </Dialog.Header>
          <Dialog.Body>
            <Flex direction="column" gap={5}>
              <Controller
                name="title"
                control={control}
                render={({ field }) => (
                  <InputField
                    {...field}
                    label="Project name"
                    placeholder="Project name"
                    data-test-id="rename-project-title-input"
                  />
                )}
              />
            </Flex>
          </Dialog.Body>
          <Dialog.Footer>
            <Dialog.Close asChild>
              <Button
                type="button"
                variant="outline"
                color="neutral"
                data-test-id="rename-project-cancel-button"
              >
                Cancel
              </Button>
            </Dialog.Close>
            <Button
              type="submit"
              data-test-id="rename-project-update-button"
              loading={isSubmitting}
              loaderText="Renaming..."
            >
              Rename
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
}
