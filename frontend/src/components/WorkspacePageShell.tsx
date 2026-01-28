import React from "react";
import { WorkspaceSidebar } from "./WorkspaceSidebar";

type Props = {
  children: React.ReactNode;
};

export function WorkspacePageShell({ children }: Props) {
  return (
    <div className="flex h-full">
      <WorkspaceSidebar />
      <main 
        id="main-content"
        className="flex-1 overflow-y-auto focus:outline-none"
        role="main"
        aria-label="Main content"
      >
        <div className="max-w-4xl mx-auto py-6 px-4 lg:px-8">{children}</div>
      </main>
    </div>
  );
}

