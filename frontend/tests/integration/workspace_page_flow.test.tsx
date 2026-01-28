import React from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { WorkspacePageView } from "../../src/pages/WorkspacePageView";

// Mock Next.js navigation
vi.mock("next/navigation", () => ({
  useSearchParams: vi.fn(() => ({
    get: vi.fn((key: string) => {
      if (key === "pageId") return null;
      return null;
    }),
  })),
}));

// Mock the API
vi.mock("@/services/workspaceApi", () => ({
  workspaceApi: {
    getPage: vi.fn(),
    updatePage: vi.fn(),
    deletePage: vi.fn(),
    listWorkspaces: vi.fn().mockResolvedValue([]),
  },
}));

// Mock child components that might have dependencies
vi.mock("@/components/BlockList", () => ({
  BlockList: ({ pageId }: { pageId?: string }) => (
    <div data-testid="block-list">
      {pageId ? `BlockList for page ${pageId}` : "Content blocks will be rendered here."}
    </div>
  ),
}));

vi.mock("@/components/ShareDialog", () => ({
  ShareDialog: ({ pageId, onClose }: { pageId: string; onClose: () => void }) => (
    <div data-testid="share-dialog">Share Dialog for {pageId}</div>
  ),
}));

vi.mock("@/components/WorkspaceSidebar", () => ({
  WorkspaceSidebar: () => (
    <aside data-testid="workspace-sidebar">Workspace Sidebar</aside>
  ),
}));

describe("WorkspacePageView", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders workspace shell with sidebar when no page is selected", () => {
    render(<WorkspacePageView />);

    expect(screen.getByTestId("workspace-sidebar")).toBeInTheDocument();
    expect(screen.getByText("No page selected. Select a page from the sidebar.")).toBeInTheDocument();
  });

  it("renders page content when pageId is provided", async () => {
    const mockGet = vi.fn((key: string) => {
      if (key === "pageId") return "test-page-id";
      return null;
    });
    
    const { useSearchParams } = await import("next/navigation");
    vi.mocked(useSearchParams).mockReturnValue({
      get: mockGet,
    } as any);

    const { workspaceApi } = await import("@/services/workspaceApi");
    vi.mocked(workspaceApi.getPage).mockResolvedValue({
      id: "test-page-id",
      title: "Test Page",
      workspaceId: "workspace-1",
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    } as any);

    render(<WorkspacePageView />);

    await screen.findByText("Test Page");
    expect(screen.getByTestId("block-list")).toBeInTheDocument();
  });
});

