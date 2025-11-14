import { User } from "@raystack/proton/frontier";

export const getUserName = (user?: User) =>
  user?.title || user?.name || "";

export const USER_STATES = {
  enabled: "Active",
  disabled: "Suspended",
} as const;

export type UserState = keyof typeof USER_STATES;
