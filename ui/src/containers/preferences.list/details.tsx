import PageHeader from "~/components/page-header";
import {
  Button,
  Flex,
  Grid,
  Separator,
  Switch,
  Text,
  TextField,
} from "@raystack/apsara";
import { useCallback, useEffect, useState } from "react";
import { V1Beta1Preference, V1Beta1PreferenceTrait } from "@raystack/frontier";
import { useParams } from "react-router-dom";
import { useFrontier } from "@raystack/frontier/react";
import Skeleton from "react-loading-skeleton";
import dayjs from "dayjs";
import * as R from "ramda";
import { toast } from "sonner";

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
      />
    );
  } else if (R.or(R.has("text"), R.has("textarea"))(trait)) {
    return (
      <TextField value={value} onChange={(e) => onChange(e.target.value)} />
    );
  } else {
    return null;
  }
}

export default function PreferenceDetails() {
  const { client } = useFrontier();
  const { name } = useParams();

  const [isPreferencesLoading, setIsPreferencesLoading] = useState(false);

  const [trait, setTrait] = useState<V1Beta1PreferenceTrait | undefined>();
  const [preference, setPreference] = useState<V1Beta1Preference | undefined>();
  const [value, setValue] = useState("");
  const [isActionLoading, setIsActionLoading] = useState(false);

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
    async function getPreference(name: string) {
      try {
        setIsPreferencesLoading(true);
        const [traitResp, valuesMapResp] = await Promise.all([
          client?.frontierServiceDescribePreferences(),
          client?.adminServiceListPreferences(),
        ]);

        const newPreference = valuesMapResp?.data?.preferences?.find(
          (p) => p.name === name
        );
        const newTrait = traitResp?.data?.traits?.find((t) => t.name === name);
        setPreference(newPreference);
        setTrait(newTrait);
        const v =
          newPreference?.value !== "" && newPreference?.value !== undefined
            ? newPreference?.value
            : newTrait?.default;
        setValue(v || "");
      } catch (err) {
        console.error(err);
      } finally {
        setIsPreferencesLoading(false);
      }
    }
    if (name) {
      getPreference(name);
    }
  }, [name]);

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
      const resp = await client?.adminServiceCreatePreferences({
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
  }, [client, name, value]);

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
          )
        )}
        <Separator />
        {isPreferencesLoading ? (
          <Skeleton />
        ) : (
          <Text size={1} weight={500}>
            Value
          </Text>
        )}
        {preference && trait ? (
          <Flex direction={"column"} gap={"medium"}>
            <PreferenceValue trait={trait} value={value} onChange={setValue} />
            <Button
              variant={"primary"}
              onClick={onSave}
              disabled={isActionLoading}
            >
              {isActionLoading ? "Saving..." : "Save"}
            </Button>
          </Flex>
        ) : null}
      </Flex>
    </Flex>
  );
}
