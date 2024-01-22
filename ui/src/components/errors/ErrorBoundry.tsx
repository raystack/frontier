import React, { Component, ErrorInfo, ReactNode } from "react";
import { ErrorState, Flex } from "@raystack/apsara";
interface Props {
  children?: ReactNode;
}

interface State {
  hasError: boolean;
}

export default class ErrorBoundary extends Component<Props, State> {
  constructor(props: any) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError() {
    // Update state so the next render will show the fallback UI.
    return { hasError: true };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error("Uncaught error:", error, errorInfo);
  }

  render() {
    return this.state.hasError ? (
      <Flex
        justify={"center"}
        align={"center"}
        style={{ minHeight: "inherit" }}
      >
        <ErrorState />
      </Flex>
    ) : (
      this.props.children
    );
  }
}
