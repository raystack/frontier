import { useEffect } from "react";

interface PageTitleProps {
  title?: string;
  appName?: string;
}

export function PageTitle({ title, appName }: PageTitleProps) {
  const fullTitle = title && appName ? `${title} | ${appName}` : title ?? appName ?? "";

  useEffect(() => {
    if (fullTitle) document.title = fullTitle;
    return () => {
      document.title = appName ?? "";
    };
  }, [fullTitle, appName]);

  return null;
}
