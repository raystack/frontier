import PageHeader from "~/components/page-header";
import { Grid, TextField } from "@raystack/apsara";
import { Button, Flex, Separator, Switch, Text } from "@raystack/apsara/v1";
import { useCallback, useEffect, useState } from "react";
import { V1Beta1Preference, V1Beta1PreferenceTrait } from "@raystack/frontier";
import { useOutletContext, useParams } from "react-router-dom";
import Skeleton from "react-loading-skeleton";
import dayjs from "dayjs";
import * as R from "ramda";
import { toast } from "sonner";
import { api } from "~/api";

interface ContextType {
  preferences: V1Beta1Preference[];
  traits: V1Beta1PreferenceTrait[];
  isPreferencesLoading: boolean;
}

export function usePreferences() {
  return useOutletContext<ContextType>();
}

interface PreferenceValueProps {
  trait: V1Beta1PreferenceTrait;
  value: string;
  onChange: (v: string) => void;
}

function PreferenceValue({ value, trait, onChange }: PreferenceValueProps) {
  if (R.has("checkbox")(trait)) {
    const checked = value === "true";
    return (
      <Switch
        // @ts-ignore
        checked={checked}
        onCheckedChange={(v: boolean) => onChange(v.toString())}
        data-test-id="admin-ui-preference-select"
      />
    );
  } else if (R.has("text")(trait) || R.has("textarea")(trait)) {
    return (
      <TextField
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
  const [isActionLoading, setIsActionLoading] = useState(false);
  const { preferences, traits, isPreferencesLoading } = usePreferences();
  const preference = preferences?.find((p) => p.name === name);
  const trait = traits?.find((t) => t.name === name);

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
      value: trait?.sub_heading,
    },
    {
      key: "Resource type",
      value: trait?.resource_type,
    },
    {
      key: "Default value",
      value: trait?.default,
    },
    {
      key: "Last updated",
      value:
        preference?.updated_at &&
        dayjs(preference?.updated_at).format("MMM DD, YYYY hh:mm:A"),
    },
  ];

  const onSave = useCallback(async () => {
    setIsActionLoading(true);
    try {
      const resp = await api?.adminServiceCreatePreferences({
        preferences: [
          {
            name,
            value,
          },
        ],
      });
      if (resp?.status === 200) {
        toast.success("preference updated");
      }
    } catch (err) {
      console.error(err);
      toast.error("something went wrong");
    } finally {
      setIsActionLoading(false);
    }
  }, [name, value]);

  return (
    <Flex direction={"column"} style={{ width: "100%" }} gap="large">
      <PageHeader
        title={pageHeader.title}
        breadcrumb={pageHeader.breadcrumb}
        style={{ borderBottom: "1px solid var(--border-base)", gap: "16px" }}
      />
      <Flex direction="column" gap="large" style={{ padding: "0 24px" }}>
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
          <Flex direction={"column"} gap={"medium"}>
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
