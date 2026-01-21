import React, { Component, ReactNode } from "react";
import { EmptyState } from "@raystack/apsara";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error("Error caught by boundary:", error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <EmptyState
          icon={<ExclamationTriangleIcon />}
          heading="Something went wrong"
          subHeading={
            this.state.error?.message ||
            "An unexpected error occurred. Please refresh the page and try again."
          }
        />
      );
    }

    return this.props.children;
  }
}
