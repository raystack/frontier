import { V1Beta1Feature } from "@raystack/frontier";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { api } from "~/api";

export function useFeatures() {
  const [features, setFeatures] =
    useState<{ label: string | undefined; value: string | undefined }[]>();

  useEffect(() => {
    async function getFeatures() {
      try {
        const res = await api?.frontierServiceListFeatures();
        const features = res?.data?.features ?? [];
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
