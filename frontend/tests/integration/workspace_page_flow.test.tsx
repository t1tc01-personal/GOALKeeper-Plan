import React from "react";
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { WorkspacePageView } from "../../src/pages/WorkspacePageView";

describe("WorkspacePageView", () => {
  it("renders workspace shell with sidebar and page title", () => {
    render(<WorkspacePageView />);

    expect(screen.getByText("Workspace")).toBeInTheDocument();
    expect(screen.getByText("Untitled page")).toBeInTheDocument();
    expect(
      screen.getByText("Content blocks will be rendered here."),
    ).toBeInTheDocument();
  });
});

