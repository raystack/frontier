import { PlansView } from "@raystack/frontier/admin";
import { useNavigate, useParams } from "react-router-dom";
import PlansIcon from "~/assets/icons/plans.svg?react";

export function PlansPage() {
  const { planId } = useParams();
  const navigate = useNavigate();

  return (
    <PlansView
      selectedPlanId={planId}
      onCloseDetail={() => navigate("/plans")}
      onSelectPlan={(id: string) => navigate(`/plans/${id}`)}
      icon={<PlansIcon />}
    />
  );
}
