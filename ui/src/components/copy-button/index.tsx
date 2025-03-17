import { IconButton } from "@raystack/apsara/v1";
import { useCopyToClipboard } from "usehooks-ts";
import { CopyIcon } from "@radix-ui/react-icons";
import CheckCircleFilledIcon from "~/assets/icons/check-circle-filled.svg?react";
import { useState } from "react";

export const CopyButton = ({
  text,
  resetDelay,
  ...props
}: {
  text: string;
  resetDelay?: number;
}) => {
  const [_, copy] = useCopyToClipboard();
  const [isCopied, setIsCopied] = useState(false);

  async function onCopy() {
    const res = await copy(text);
    if (res) {
      setIsCopied(true);
      if (resetDelay) {
        setTimeout(() => {
          setIsCopied(false);
        }, resetDelay);
      }
    }
  }

  return (
    <IconButton {...props} onClick={onCopy} data-test-id="copy-button">
      {isCopied ? (
        <CheckCircleFilledIcon
          color={"var(--rs-color-foreground-success-primary)"}
        />
      ) : (
        <CopyIcon />
      )}
    </IconButton>
  );
};
