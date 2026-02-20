import { PlansView } from "@raystack/frontier/admin";
import { useNavigate, useParams } from "react-router-dom";

export function PlansPage() {
  const { planId } = useParams();
  const navigate = useNavigate();

  return (
    <PlansView
      selectedPlanId={planId}
      onCloseDetail={() => navigate("/plans")}
    />
  );
}
