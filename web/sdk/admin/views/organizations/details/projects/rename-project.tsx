import { Button, Dialog, Field, Flex, Input, toastManager } from "@raystack/apsara";
import {
  FrontierServiceQueries,
  UpdateProjectRequestSchema,
  type SearchOrganizationProjectsResponse_OrganizationProject,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm, Controller } from "react-hook-form";
import React from "react";
import { useMutation } from "@connectrpc/connect-query";
import { useTerminology } from "../../../../hooks/useTerminology";

const projectRenameSchema = z.object({
  title: z.string(),
  name: z.string(),
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
  const t = useTerminology();
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
        toastManager.add({ title: `${t.project({ case: "capital" })} renamed successfully`, type: "success" });
        onRename({
          ...project,
          title: newProject.title,
          name: newProject.name,
        });
      }
    } catch (error) {
      toastManager.add({ title: `Failed to rename ${t.project({ case: "lower" })}`, type: "error" });
      console.error(error);
    }
  };

  return (
    <Dialog open onOpenChange={onClose}>
      <Dialog.Content
        onClick={preventClickPropagation}
        data-test-id="rename-project-dialog"
      >
        <form onSubmit={handleSubmit(submit)}>
          <Dialog.Header>
            <Dialog.Title>Rename {t.project({ case: "capital" })}</Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Flex direction="column" gap={5}>
              <Controller
                name="title"
                control={control}
                render={({ field }) => (
                  <Field label={`${t.project({ case: "capital" })} name`}>
                    <Input
                      {...field}
                      placeholder={`${t.project({ case: "capital" })} name`}
                      data-test-id="rename-project-title-input"
                    />
                  </Field>
                )}
              />
            </Flex>
          </Dialog.Body>
          <Dialog.Footer>
            <Dialog.Close
              render={
                <Button
                  type="button"
                  variant="outline"
                  color="neutral"
                  data-test-id="rename-project-cancel-button"
                >
                  Cancel
                </Button>
              }
            />
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
