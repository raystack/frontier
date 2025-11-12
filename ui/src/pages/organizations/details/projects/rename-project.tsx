import { Button, Dialog, Flex, InputField, toast } from "@raystack/apsara";
import type {
  SearchOrganizationProjectsResponse_OrganizationProject,
} from "@raystack/proton/frontier";
import { FrontierServiceQueries, UpdateProjectRequestSchema } from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm, Controller } from "react-hook-form";
import React from "react";
import { useMutation } from "@connectrpc/connect-query";

const projectRenameSchema = z.object({
  title: z.string(),
  name: z.string(),
  orgId: z.string(),
});

type ProjectRenameSchema = z.infer<typeof projectRenameSchema>;

interface RenameProjectDialogProps {
  onClose: () => void;
  project: SearchOrganizationProjectsResponse_OrganizationProject;
  onRename: (
    project: SearchOrganizationProjectsResponse_OrganizationProject,
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
      orgId: project?.organizationId,
    },
    resolver: zodResolver(projectRenameSchema),
  });

  const { mutateAsync: updateProject } = useMutation(
    FrontierServiceQueries.updateProject,
  );

  const submit = async (data: ProjectRenameSchema) => {
    try {
      const resp = await updateProject(
        create(UpdateProjectRequestSchema, {
          id: project?.id || "",
          body: {...data},
        }),
      );
      const newProject = resp.project;

      if (newProject) {
        toast.success("Project renamed successfully");
        onRename({
          ...project,
          title: newProject.title,
          name: newProject.name,
        });
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
