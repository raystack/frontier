import { useContext, useEffect } from "react";
import { AppContext } from "~/contexts/App";

interface PageTitleProps {
  title?: string;
  appName?: string;
}

export default function PageTitle({ title, appName }: PageTitleProps) {
  const { config } = useContext(AppContext);
  const titleAppName = appName || config?.title;
  const fullTitle = title ? `${title} | ${titleAppName}` : titleAppName;

  useEffect(() => {
    document.title = fullTitle;
  }, [fullTitle]);
  return null;
}
