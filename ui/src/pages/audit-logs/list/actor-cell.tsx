import { Avatar, Flex, getAvatarColor, Text } from "@raystack/apsara";
import { AuditRecordActor } from "@raystack/proton/frontier";
import { ACTOR_TYPES, getAuditLogActorName } from "../util";
import systemIcon from "~/assets/images/system.jpg";
import KeyIcon from "~/assets/icons/key.svg?react";

type ActorCellProps = {
  size?: "large" | "small";
  hideName?: boolean;
  value: AuditRecordActor;
  hideAvatar?: boolean;
};

export default function ActorCell({
  size = "large",
  hideName = false,
  value,
  hideAvatar = false,
}: ActorCellProps) {
  const name = getAuditLogActorName(value);
  const isSystem = value.type === ACTOR_TYPES.SYSTEM;
  const isServiceUser = value.type === ACTOR_TYPES.SERVICE_USER;

  return (
    <Flex gap={size === "large" ? 4 : 3} align="center">
      {!hideAvatar && (
        <Avatar
          size={size === "large" ? 3 : 1}
          fallback={
            isServiceUser ? (
              <KeyIcon width={12} height={12} />
            ) : (
              name?.[0]?.toUpperCase()
            )
          }
          color={isServiceUser ? "neutral" : getAvatarColor(value?.id ?? "")}
          radius="full"
          src={isSystem ? systemIcon : undefined}
        />
      )}
      {!hideName &&
        (size === "large" ? <Text size="regular">{name}</Text> : name)}
    </Flex>
  );
}
