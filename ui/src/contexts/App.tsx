import * as R from "ramda";
import React, {
  PropsWithChildren,
  createContext,
  useCallback,
  useEffect,
  useState,
} from "react";
import {
  V1Beta1ListPlatformUsersResponse,
  V1Beta1Organization,
  V1Beta1Plan,
  V1Beta1User,
} from "@raystack/frontier";
import { Config, defaultConfig } from "~/utils/constants";
import { api } from "~/api";

// TODO: Setting this to 1000 initially till APIs support filters and sorting.
const page_size = 1000;

type OrgMap = Record<string, V1Beta1Organization>;

interface AppContextValue {
  orgMap: OrgMap;
  isAdmin: boolean;
  isLoading: boolean;
  organizations: V1Beta1Organization[];
  plans: V1Beta1Plan[];
  platformUsers?: V1Beta1ListPlatformUsersResponse;
  fetchPlatformUsers: () => void;
  loadMoreOrganizations: () => void;
  config: Config;
  user?: V1Beta1User;
}

const AppContextDefaultValue = {
  orgMap: {},
  isAdmin: false,
  isLoading: false,
  organizations: [],
  plans: [],
  platformUsers: {
    users: [],
    serviceusers: [],
  },
  fetchPlatformUsers: () => {},
  loadMoreOrganizations: () => {},
  config: defaultConfig,
};

export const AppContext = createContext<AppContextValue>(
  AppContextDefaultValue
);

export const AppContextProvider: React.FC<PropsWithChildren> = function ({
  children,
}) {
  const [user, setUser] = useState<V1Beta1User | undefined>();
  const [isUserLoading, setIsUserLoading] = useState(true);

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

  const [isPlatformUsersLoading, setIsPlatformUsersLoading] = useState(false);
  const [platformUsers, setPlatformUsers] =
    useState<V1Beta1ListPlatformUsersResponse>();

  const [config, setConfig] = useState<Config>(defaultConfig);

  const [page, setPage] = useState(1);
  const [enabledOrgHasMoreData, setEnabledOrgHasMoreData] = useState(true);
  const [disabledOrgHasMoreData, setDisabledOrgHasMoreData] = useState(true);

  const isUserEmpty = R.either(R.isEmpty, R.isNil)(user);

  const fetchOrganizations = useCallback(async () => {
    if (!enabledOrgHasMoreData && !disabledOrgHasMoreData) return;

    setIsOrgListLoading(true);
    try {
      const [orgResp, disabledOrgResp] = await Promise.all([
        api?.adminServiceListAllOrganizations({ page_num: page, page_size }),
        api?.adminServiceListAllOrganizations({
          state: "disabled",
          page_num: page,
          page_size,
        }),
      ]);

      if (orgResp?.data?.organizations?.length) {
        setEnabledOrganizations((prev: V1Beta1Organization[]) => [
          ...prev,
          ...(orgResp.data.organizations || []),
        ]);
      } else {
        setEnabledOrgHasMoreData(false);
      }

      if (disabledOrgResp?.data?.organizations?.length) {
        setDisabledOrganizations((prev: V1Beta1Organization[]) => [
          ...prev,
          ...(disabledOrgResp.data.organizations || []),
        ]);
      } else {
        setDisabledOrgHasMoreData(false);
      }
      setIsAdmin(true);
    } catch (error) {
      console.error(error);
      setIsAdmin(false);
      setEnabledOrgHasMoreData(false);
      setDisabledOrgHasMoreData(false);
    } finally {
      setIsOrgListLoading(false);
    }
  }, [page, enabledOrgHasMoreData, disabledOrgHasMoreData]);

  const loadMoreOrganizations = () => {
    if (
      !isOrgListLoading &&
      (enabledOrgHasMoreData || disabledOrgHasMoreData)
    ) {
      setPage((prevPage: number) => prevPage + 1);
    }
  };

  useEffect(() => {
    async function fetchUser() {
      setIsUserLoading(true);
      try {
        const resp = await api.frontierServiceGetCurrentUser();
        setUser(resp?.data?.user);
      } catch (error) {
        console.error(error);
      } finally {
        setIsUserLoading(false);
      }
    }
    fetchUser();
  }, []);

  useEffect(() => {
    if (!isUserEmpty) {
      fetchOrganizations();
    }
  }, [isUserEmpty, page, fetchOrganizations]);

  const fetchPlatformUsers = useCallback(async () => {
    setIsPlatformUsersLoading(true);
    try {
      const resp = await api?.adminServiceListPlatformUsers();
      if (resp?.data) {
        setPlatformUsers(resp?.data);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setIsPlatformUsersLoading(false);
    }
  }, []);

  const fetchConfig = useCallback(async () => {
    setIsPlatformUsersLoading(true);
    try {
      const resp = await fetch("/configs");
      const data = (await resp?.json()) as Config;
      setConfig(data);
    } catch (err) {
      console.error(err);
    } finally {
      setIsPlatformUsersLoading(false);
    }
  }, []);

  useEffect(() => {
    async function getPlans() {
      setIsPlansLoading(true);
      try {
        const resp = await api?.frontierServiceListPlans();
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
      fetchPlatformUsers();
    }
    fetchConfig();
  }, [isAdmin, fetchPlatformUsers, fetchConfig]);

  const isLoading =
    isOrgListLoading ||
    isUserLoading ||
    isPlansLoading ||
    isPlatformUsersLoading;
  const organizations = [...enabledOrganizations, ...disabledOrganizations];

  const orgMap = organizations.reduce((acc, org) => {
    const orgId = org?.id || "";
    if (orgId) acc[orgId] = org;
    return acc;
  }, {} as OrgMap);

  return (
    <AppContext.Provider
      value={{
        orgMap,
        isLoading,
        isAdmin,
        organizations,
        plans,
        platformUsers,
        fetchPlatformUsers,
        loadMoreOrganizations,
        config,
        user,
      }}
    >
      {children}
    </AppContext.Provider>
  );
};
