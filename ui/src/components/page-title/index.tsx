import { useContext, useEffect } from "react";
import { AppContext } from "~/contexts/App";
import { defaultConfig } from "~/utils/constants";

interface PageTitleProps {
  title?: string;
  appName?: string;
}

export default function PageTitle({ title, appName }: PageTitleProps) {
  const { config } = useContext(AppContext);
  const titleAppName = appName || config?.title || defaultConfig?.title;
  const fullTitle = title ? `${title} | ${titleAppName}` : titleAppName;

  useEffect(() => {
    document.title = fullTitle;

    return () => {
      document.title = titleAppName;
    };
  }, [fullTitle, titleAppName]);
  return null;
}
