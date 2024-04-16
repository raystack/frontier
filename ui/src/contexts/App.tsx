import * as R from "ramda";
import { PropsWithChildren, createContext, useEffect, useState } from "react";
import { V1Beta1Organization, V1Beta1Plan } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";

type OrgMap = Record<string, V1Beta1Organization>;

interface AppContextValue {
  orgMap: OrgMap;
  isAdmin: boolean;
  isLoading: boolean;
  organizations: V1Beta1Organization[];
  plans: V1Beta1Plan[];
}

const AppContextDefaultValue = {
  orgMap: {},
  isAdmin: false,
  isLoading: false,
  organizations: [],
  plans: [],
};

export const AppContext = createContext<AppContextValue>(
  AppContextDefaultValue
);

export const AppConextProvider: React.FC<PropsWithChildren> = function ({
  children,
}) {
  const { client, user, isUserLoading } = useFrontier();
  const [isOrgListLoading, setIsOrgListLoading] = useState(false);
  const [isAdmin, setIsAdmin] = useState(false);
  const [enabledOrganizations, setEnabledOrganizations] = useState<
    V1Beta1Organization[]
  >([]);
  const [disabledOrganizations, setDisabledOrganizations] = useState<
    V1Beta1Organization[]
  >([]);
  const [plans, setPlans] = useState<V1Beta1Plan[]>([]);
  const [isPlansLoading, setIsPlansLoading] = useState(false);

  const isUserEmpty = R.either(R.isEmpty, R.isNil)(user);

  useEffect(() => {
    async function getOrganizations() {
      setIsOrgListLoading(true);
      try {
        const [orgResp, disabledOrgResp] = await Promise.all([
          client?.adminServiceListAllOrganizations(),
          client?.adminServiceListAllOrganizations({ state: "disabled" }),
        ]);
        if (orgResp?.data?.organizations) {
          setEnabledOrganizations(orgResp?.data?.organizations);
        }
        if (disabledOrgResp?.data?.organizations) {
          setDisabledOrganizations(disabledOrgResp?.data?.organizations);
        }
        setIsAdmin(true);
      } catch (error) {
        setIsAdmin(false);
      } finally {
        setIsOrgListLoading(false);
      }
    }

    if (!isUserEmpty) {
      getOrganizations();
    }
  }, [client, isUserEmpty]);

  useEffect(() => {
    async function getPlans() {
      setIsPlansLoading(true);
      try {
        const resp = await client?.frontierServiceListPlans();
        const planList = resp?.data?.plans || [];
        setPlans(planList);
      } catch (error) {
        console.error(error);
      } finally {
        setIsPlansLoading(false);
      }
    }

    if (isAdmin) {
      getPlans();
    }
  }, [client, isAdmin]);

  const isLoading = isOrgListLoading || isUserLoading || isPlansLoading;
  const organizations = [...enabledOrganizations, ...disabledOrganizations];

  const orgMap = organizations.reduce((acc, org) => {
    const orgId = org?.id || "";
    if (orgId) acc[orgId] = org;
    return acc;
  }, {} as OrgMap);

  return (
    <AppContext.Provider
      value={{ orgMap, isLoading, isAdmin, organizations, plans }}
    >
      {children}
    </AppContext.Provider>
  );
};
