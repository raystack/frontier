import { Flex } from "@raystack/apsara/v1";
import { V1Beta1Preference, V1Beta1PreferenceTrait } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";

import { Outlet } from "react-router-dom";

export default function PreferencesLayout() {
  const { client } = useFrontier();
  const [preferences, setPreferences] = useState<V1Beta1Preference[]>([]);
  const [traits, setTraits] = useState<V1Beta1PreferenceTrait[]>([]);
  const [isPreferencesLoading, setIsPreferencesLoading] = useState(false);

  useEffect(() => {
    async function getPreferences() {
      try {
        setIsPreferencesLoading(true);
        const [traitResp, valuesMapResp] = await Promise.all([
          client?.frontierServiceDescribePreferences(),
          client?.adminServiceListPreferences(),
        ]);

        if (valuesMapResp?.data?.preferences) {
          setPreferences(valuesMapResp?.data?.preferences);
        }
        if (traitResp?.data?.traits) {
          setTraits(traitResp?.data?.traits);
        }
      } catch (err) {
        console.error(err);
      } finally {
        setIsPreferencesLoading(false);
      }
    }
    getPreferences();
  }, []);

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <Outlet
        context={{
          preferences,
          traits,
          isPreferencesLoading,
        }}
      />
    </Flex>
  );
}
