import { V1Beta1User } from "~/api/frontier";

export const getUserName = (user?: V1Beta1User) =>
  user?.title || user?.name || "";

export const USER_STATES = {
  enabled: "Active",
  disabled: "Suspended",
} as const;

export type UserState = keyof typeof USER_STATES;
