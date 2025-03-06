import { Flex, Separator, Switch, Text } from "@raystack/apsara/v1";
import { V1Beta1Preference } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import PageHeader from "~/components/page-header";

export default function OrgSettingPage() {
  let { organisationId } = useParams();
  const [socialLogin, setSocialLogin] = useState<boolean>(false);
  const [mailLink, setMailLink] = useState<boolean>(false);

  const [preferences, setPreferences] = useState<V1Beta1Preference[]>([]);
  const { client } = useFrontier();

  const fetchOrganizationPreferences = useCallback(async () => {
    try {
      const res = await client?.frontierServiceListOrganizationPreferences(
        organisationId ?? ""
      );
      const preferences = res?.data?.preferences ?? [];
      setPreferences(preferences);
    } catch (error) {
      console.error(error);
    }
  }, [client, organisationId]);

  useEffect(() => {
    if (organisationId) fetchOrganizationPreferences();
  }, [organisationId, client, fetchOrganizationPreferences]);

  const preferencesMap = useMemo(() => {
    return preferences.reduce<Record<string, Record<string, string>>>(
      (map, el) => {
        // @ts-ignore
        map[el.name] = el;
        return map;
      },
      {}
    );
  }, [preferences]);

  useEffect(() => {
    if (preferencesMap["social_login"])
      setSocialLogin(preferencesMap["social_login"]?.value === "true");

    if (preferencesMap["mail_link"])
      setMailLink(preferencesMap["mail_link"]?.value === "true");
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [preferences]);

  const onValueChange = useCallback(
    async (key: string, checked: boolean) => {
      if (key === "mail_link") setMailLink(checked);
      if (key === "social_login") setSocialLogin(checked);
      await client?.frontierServiceCreateOrganizationPreferences(
        organisationId as string,
        {
          bodies: [
            {
              name: key,
              value: `${checked}`,
            },
          ],
        }
      );
    },
    [client, organisationId]
  );
  const pageHeader = {
    title: "Organizations",
    breadcrumb: [
      {
        href: `/organisations`,
        name: `Organizations list`,
      },
      {
        href: `/organisations/${organisationId}`,
        name: `Organization`,
      },
      {
        href: ``,
        name: `Organizations Service Users`,
      },
    ],
  };

  return (
    <Flex
      direction="column"
      gap="large"
      style={{
        width: "100%",
        height: "calc(100vh - 60px)",
        borderLeft: "1px solid var(--border-base)",
      }}
    >
      <PageHeader
        title={pageHeader.title}
        breadcrumb={pageHeader.breadcrumb}
        style={{ borderBottom: "1px solid var(--border-base)" }}
      ></PageHeader>
      <Flex direction="column" gap="large" style={{ padding: "0 24px" }}>
        <Flex direction="column" style={{ width: "100%" }}>
          <Flex direction="column" gap="large">
            <SecurityCheckbox
              label="Google"
              text="Allow logins through Google&#39;s single sign-on functionality"
              name="social_login"
              value={socialLogin}
              onValueChange={onValueChange}
            />
            <Separator></Separator>
            <SecurityCheckbox
              label="Email code"
              text="Allow password less logins through magic links or a code delivered
      over email."
              name="mail_link"
              value={mailLink}
              onValueChange={onValueChange}
            />
            <Separator></Separator>
          </Flex>
        </Flex>
      </Flex>
    </Flex>
  );
}
export type SecurityCheckboxTypes = {
  label: string;
  name: string;
  text: string;
  value: boolean;
  canUpdatePrefrence?: boolean;
  onValueChange: (key: string, checked: boolean) => void;
};

export const SecurityCheckbox = ({
  label,
  text,
  name,
  value,
  onValueChange,
  canUpdatePrefrence = true,
}: SecurityCheckboxTypes) => {
  return (
    <Flex direction="row" justify="between" align="center">
      <Flex direction="column" gap="small">
        <Text size={6}>{label}</Text>
        <Text size={4} variant="secondary">
          {text}
        </Text>
      </Flex>

      {canUpdatePrefrence ? (
        <Switch
          // @ts-ignore
          name={name}
          checked={value}
          onCheckedChange={(checked: boolean) => onValueChange(name, checked)}
        />
      ) : null}
    </Flex>
  );
};
