import { V1Beta1Feature } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

export function useFeatures() {
  const { client } = useFrontier();
  const [features, setFeatures] = useState();

  useEffect(() => {
    async function getFeatures() {
      try {
        const {
          // @ts-ignore
          data: { features },
        } = await client?.frontierServiceListFeatures() || {};
        setFeatures(
          features.map((f: V1Beta1Feature) => ({
            label: f.name,
            value: f.name,
          }))
        );
      } catch (error: any) {
        toast.error("Something went wrong", {
          description: error.message,
        });
      }
    }
    getFeatures();
  }, []);

  return { features };
}
