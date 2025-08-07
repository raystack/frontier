import { Flex } from "@raystack/apsara";
import { V1Beta1Preference, V1Beta1PreferenceTrait } from "@raystack/frontier";
import { useEffect, useState } from "react";

import { Outlet } from "react-router-dom";
import { api } from "~/api";

export default function PreferencesLayout() {
  const [preferences, setPreferences] = useState<V1Beta1Preference[]>([]);
  const [traits, setTraits] = useState<V1Beta1PreferenceTrait[]>([]);
  const [isPreferencesLoading, setIsPreferencesLoading] = useState(false);

  useEffect(() => {
    async function getPreferences() {
      try {
        setIsPreferencesLoading(true);
        const [traitResp, valuesMapResp] = await Promise.all([
          api?.frontierServiceDescribePreferences(),
          api?.adminServiceListPreferences(),
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
