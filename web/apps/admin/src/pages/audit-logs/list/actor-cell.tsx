import { Avatar, Flex, getAvatarColor, Text } from "@raystack/apsara";
import { AuditRecordActor } from "@raystack/proton/frontier";
import { ACTOR_TYPES, getAuditLogActorName } from "../util";
import systemIcon from "~/assets/images/system.jpg";
import KeyIcon from "~/assets/icons/key.svg?react";

type ActorCellProps = {
  size?: "large" | "small";
  value: AuditRecordActor;
  maxLength?: number;
};

export default function ActorCell({
  size = "large",
  value,
  maxLength,
}: ActorCellProps) {
  const name = getAuditLogActorName(value, maxLength);
  const isSystem = value.type === ACTOR_TYPES.SYSTEM;
  const isServiceUser = value.type === ACTOR_TYPES.SERVICE_USER;

  return (
    <Flex gap={size === "large" ? 4 : 3} align="center">
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
      {size === "large" ? <Text size="regular">{name}</Text> : name}
    </Flex>
  );
}
