import { Button } from "@raystack/apsara";

export const DOCS_URL = "https://raystack-frontier.vercel.app/docs";

interface DocumentationButtonProps {
  /** Deep link to a specific page. Defaults to the docs root. */
  href?: string;
}

/*
 * The secondary action every zero state carries in the design.
 * - Rendered as an anchor via Apsara's `render` prop, so it is a real link
 *   rather than a button with an onClick.
 * - Opens in a new tab: the docs are a separate site, and losing admin state
 *   to navigate away would be worse than a new tab.
 */
export const DocumentationButton = ({
  href = DOCS_URL,
}: DocumentationButtonProps) => {
  return (
    <Button
      variant="outline"
      color="neutral"
      size="small"
      data-test-id="admin-zero-state-documentation-btn"
      render={
        <a href={href} target="_blank" rel="noreferrer">
          Documentation
        </a>
      }
    />
  );
};
