import PageHeader from "~/components/page-header";
import {
  Grid,
  Button,
  Flex,
  Separator,
  Switch,
  Text,
  InputField,
} from "@raystack/apsara";
import { useCallback, useEffect, useState } from "react";
import { useOutletContext, useParams } from "react-router-dom";
import Skeleton from "react-loading-skeleton";
import dayjs from "dayjs";
import { toast } from "sonner";
import { useMutation, createConnectQueryKey, useTransport } from "@connectrpc/connect-query";
import { AdminServiceQueries, Preference, PreferenceTrait, PreferenceTrait_InputType } from "@raystack/proton/frontier";
import { useQueryClient } from "@tanstack/react-query";

interface ContextType {
  preferences: Preference[];
  traits: PreferenceTrait[];
  isPreferencesLoading: boolean;
}

export function usePreferences() {
  return useOutletContext<ContextType>();
}

interface PreferenceValueProps {
  trait: PreferenceTrait;
  value: string;
  onChange: (v: string) => void;
}

function PreferenceValue({ value, trait, onChange }: PreferenceValueProps) {
  if (trait.inputType === PreferenceTrait_InputType.CHECKBOX) {
    const checked = value === "true";
    return (
      <Switch
        checked={checked}
        onCheckedChange={(v: boolean) => onChange(v.toString())}
        data-test-id="admin-ui-preference-select"
      />
    );
  } else if (
    trait.inputType === PreferenceTrait_InputType.TEXT ||
    trait.inputType === PreferenceTrait_InputType.TEXTAREA
  ) {
    return (
      <InputField
        value={value}
        onChange={(e) => onChange(e.target.value)}
        data-test-id="admin-ui-preference-value-input"
      />
    );
  } else {
    return null;
  }
}

export default function PreferenceDetails() {
  const { name } = useParams();
  const [value, setValue] = useState("");
  const { preferences, traits, isPreferencesLoading } = usePreferences();
  const preference = preferences?.find((p) => p.name === name);
  const trait = traits?.find((t) => t.name === name);

  const queryClient = useQueryClient();
  const transport = useTransport();

  const { mutateAsync: createPreferences, isPending: isActionLoading } =
    useMutation(AdminServiceQueries.createPreferences, {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            schema: AdminServiceQueries.listPreferences,
            transport,
            input: {},
            cardinality: "finite",
          }),
        });
      },
    });

  const pageHeader = {
    title: "Preference",
    breadcrumb: [
      {
        href: `/preferences`,
        name: `Preferences`,
      },
      {
        href: `/preferences/${trait?.name}`,
        name: `${trait?.title}`,
      },
    ],
  };

  useEffect(() => {
    const v =
      preference?.value !== "" && preference?.value !== undefined
        ? preference?.value
        : trait?.default;
    setValue(v || "");
  }, [preference?.value, trait?.default]);

  const detailList = [
    {
      key: "Title",
      value: trait?.title,
    },
    {
      key: "Name",
      value: trait?.name,
    },
    {
      key: "Description",
      value: trait?.description,
    },
    {
      key: "Heading",
      value: trait?.heading,
    },
    {
      key: "Sub heading",
      value: trait?.subHeading,
    },
    {
      key: "Resource type",
      value: trait?.resourceType,
    },
    {
      key: "Default value",
      value: trait?.default,
    },
    {
      key: "Last updated",
      value:
        preference?.updatedAt &&
        dayjs(Number(preference.updatedAt.seconds) * 1000).format("MMM DD, YYYY hh:mm:A"),
    },
  ];

  const onSave = useCallback(async () => {
    try {
      await createPreferences({
        preferences: [
          {
            name,
            value,
          },
        ],
      });
      toast.success("preference updated");
    } catch (err) {
      console.error("ConnectRPC Error:", err);
      toast.error("something went wrong");
    }
  }, [name, value, createPreferences]);

  return (
    <Flex direction="column" style={{ width: "100%" }} gap={9}>
      <PageHeader
        title={pageHeader.title}
        breadcrumb={pageHeader.breadcrumb}
        style={{
          borderBottom: "1px solid var(--rs-color-border-base-primary)",
          gap: "16px",
        }}
      />
      <Flex direction="column" gap={9} style={{ padding: "0 24px" }}>
        {detailList.map((detailItem) =>
          isPreferencesLoading ? (
            <Grid columns={2} gap="small" key={detailItem.key}>
              <Skeleton />
              <Skeleton />
            </Grid>
          ) : (
            <Grid columns={2} gap="small" key={detailItem.key}>
              <Text size={1} weight={500}>
                {detailItem.key}
              </Text>
              <Text size={1}>{detailItem.value}</Text>
            </Grid>
          ),
        )}
        <Separator />
        {isPreferencesLoading ? (
          <Skeleton />
        ) : (
          <Text size={1} weight={500}>
            Value
          </Text>
        )}
        {trait ? (
          <Flex direction="column" gap={"medium"}>
            <PreferenceValue
              trait={trait}
              value={value}
              onChange={setValue}
              data-test-id="preference-value-save"
            />
            <Button
              onClick={onSave}
              disabled={isActionLoading}
              loading={isActionLoading}
              loaderText="Saving..."
              data-test-id="admin-ui-preference-value-save-btn"
            >
              Save
            </Button>
          </Flex>
        ) : null}
      </Flex>
    </Flex>
  );
}
