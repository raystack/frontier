import React, { createContext, useEffect, useState } from "react";
import Shield from "../shield";
import {
  Group,
  ShieldClientOptions,
  ShieldProviderProps,
  User,
} from "../types";
import { Organization } from "../types/organization";
import { GroupContext } from "./GroupContext";
import { OrganizationContext } from "./OrganizationContext";

import type { Strategy } from "./StrategyContext";
import { StrategyContext } from "./StrategyContext";
import { UserContext } from "./UserContext";

type ShieldContextProviderProps = {};
const initialValues: ShieldContextProviderProps = {};
export const ShieldContext =
  createContext<ShieldContextProviderProps>(initialValues);
ShieldContext.displayName = "ShieldContext ";

export const ShieldContextProvider = (props: ShieldProviderProps) => {
  const { children, initialState, ...options } = props;
  const { shieldClient } = useShieldClient(options);
  const [strategies, setStrategies] = useState<Strategy[]>([]);
  const [user, setUser] = useState<User | null>(null);
  const [groups, setGroups] = useState<Group[]>([]);
  const [organizations, setOrganizations] = useState<Organization[]>([]);

  useEffect(() => {
    async function getShieldInformation() {
      const {
        data: { strategies },
      } = await shieldClient.getAuthAtrategies();

      const strategiesPromises = strategies.map(async (s) => {
        const {
          data: { endpoint },
        } = await shieldClient.getAuthAtrategyEndpoint(s.name);
        return {
          ...s,
          endpoint,
        };
      });
      const response = await Promise.all(strategiesPromises);
      setStrategies(response);
    }
    getShieldInformation();
  }, []);

  useEffect(() => {
    async function getShieldCurrentUser() {
      const {
        data: { user },
      } = await shieldClient.getCurrentUser();
      setUser(user);
    }
    getShieldCurrentUser();
  }, []);

  useEffect(() => {
    async function getShieldCurrentUserGroups(userId: string) {
      const {
        data: { groups },
      } = await shieldClient.getUserGroups(userId);
      setGroups(groups);
    }

    if (user) {
      getShieldCurrentUserGroups(user.id);
    }
  }, [user]);

  useEffect(() => {
    async function getShieldCurrentUserOrganizations(userId: string) {
      const {
        data: { organizations },
      } = await shieldClient.getUserOrganisations(userId);
      setOrganizations(organizations);
    }

    if (user) {
      getShieldCurrentUserOrganizations(user.id);
    }
  }, [user]);

  return (
    <ShieldContext.Provider value={{ client: shieldClient }}>
      <StrategyContext.Provider value={{ strategies }}>
        <OrganizationContext.Provider value={{ organizations }}>
          <GroupContext.Provider value={{ groups }}>
            <UserContext.Provider value={{ user }}>
              {children}
            </UserContext.Provider>
          </GroupContext.Provider>
        </OrganizationContext.Provider>
      </StrategyContext.Provider>
    </ShieldContext.Provider>
  );
};

export const useShieldClient = (options: ShieldClientOptions) => {
  const shieldClient = React.useMemo(
    () => Shield.getOrCreateInstance(options),
    []
  );

  return { shieldClient };
};
