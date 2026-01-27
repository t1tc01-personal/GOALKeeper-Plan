'use client';

import React, { useEffect, useState, useRef, useCallback } from 'react';
import { Card } from './ui/card';
import { workspaceApi, type Workspace, type Page } from '@/services/workspaceApi';

export function WorkspaceSidebar() {
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [pages, setPages] = useState<Page[]>([]);
  const [selectedWorkspace, setSelectedWorkspace] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showCreateWorkspace, setShowCreateWorkspace] = useState(false);
  const [showCreatePage, setShowCreatePage] = useState(false);
  const [newWorkspaceName, setNewWorkspaceName] = useState('');
  const [newPageTitle, setNewPageTitle] = useState('');
  const [focusedWorkspaceId, setFocusedWorkspaceId] = useState<string | null>(null);
  const workspaceListRef = useRef<HTMLDivElement>(null);
  const workspaceInputRef = useRef<HTMLInputElement>(null);
  const pageInputRef = useRef<HTMLInputElement>(null);
  const skipLinkRef = useRef<HTMLAnchorElement>(null);

  // Load workspaces on mount
  useEffect(() => {
    loadWorkspaces();
  }, []);

  // Load pages when workspace is selected
  useEffect(() => {
    if (selectedWorkspace) {
      loadPages(selectedWorkspace);
    } else {
      setPages([]);
    }
  }, [selectedWorkspace]);

  // Handle keyboard navigation in workspaces list
  const handleWorkspaceKeyDown = useCallback((e: React.KeyboardEvent, wsId: string) => {
    switch (e.key) {
      case 'ArrowDown': {
        e.preventDefault();
        const currentIndex = workspaces.findIndex(w => w.id === wsId);
        if (currentIndex < workspaces.length - 1) {
          setFocusedWorkspaceId(workspaces[currentIndex + 1].id);
        }
        break;
      }
      case 'ArrowUp': {
        e.preventDefault();
        const currentIndex = workspaces.findIndex(w => w.id === wsId);
        if (currentIndex > 0) {
          setFocusedWorkspaceId(workspaces[currentIndex - 1].id);
        }
        break;
      }
      case 'Home': {
        e.preventDefault();
        if (workspaces.length > 0) {
          setFocusedWorkspaceId(workspaces[0].id);
        }
        break;
      }
      case 'End': {
        e.preventDefault();
        if (workspaces.length > 0) {
          setFocusedWorkspaceId(workspaces[workspaces.length - 1].id);
        }
        break;
      }
      case 'Enter':
      case ' ': {
        e.preventDefault();
        setSelectedWorkspace(wsId);
        break;
      }
      case 'Escape': {
        e.preventDefault();
        if (showCreateWorkspace) {
          setShowCreateWorkspace(false);
          setNewWorkspaceName('');
        }
        break;
      }
    }
  }, [workspaces, showCreateWorkspace]);

  // Update focus on workspace
  useEffect(() => {
    if (focusedWorkspaceId && workspaceListRef.current) {
      const button = workspaceListRef.current.querySelector(
        `[data-workspace-id="${focusedWorkspaceId}"]`
      ) as HTMLButtonElement;
      if (button) {
        button.focus();
      }
    }
  }, [focusedWorkspaceId]);

  const loadWorkspaces = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await workspaceApi.listWorkspaces();
      setWorkspaces(data);
      if (data.length > 0 && !selectedWorkspace) {
        setSelectedWorkspace(data[0].id);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load workspaces');
    } finally {
      setIsLoading(false);
    }
  };

  const loadPages = async (workspaceId: string) => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await workspaceApi.getPageHierarchy(workspaceId);
      setPages(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load pages');
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateWorkspace = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newWorkspaceName.trim()) return;

    try {
      setIsLoading(true);
      setError(null);
      await workspaceApi.createWorkspace({
        name: newWorkspaceName,
      });
      setNewWorkspaceName('');
      setShowCreateWorkspace(false);
      await loadWorkspaces();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create workspace');
    } finally {
      setIsLoading(false);
    }
  };

  // Focus input when create workspace modal opens
  useEffect(() => {
    if (showCreateWorkspace && workspaceInputRef.current) {
      workspaceInputRef.current.focus();
    }
  }, [showCreateWorkspace]);

  const handleCreatePage = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newPageTitle.trim() || !selectedWorkspace) return;

    try {
      setIsLoading(true);
      setError(null);
      await workspaceApi.createPage({
        workspaceId: selectedWorkspace,
        title: newPageTitle,
      });
      setNewPageTitle('');
      setShowCreatePage(false);
      await loadPages(selectedWorkspace);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create page');
    } finally {
      setIsLoading(false);
    }
  };

  // Focus input when create page modal opens
  useEffect(() => {
    if (showCreatePage && pageInputRef.current) {
      pageInputRef.current.focus();
    }
  }, [showCreatePage]);

  const renderPages = (parentId: string | undefined = undefined, depth: number = 0) => {
    const childPages = pages.filter(p => 
      parentId ? p.parent_page_id === parentId : !p.parent_page_id
    );

    if (childPages.length === 0) {
      return null;
    }

    return (
      <ul className={`space-y-1 ${depth > 0 ? 'ml-4' : ''}`}>
        {childPages.map(page => (
          <li key={page.id}>
            <button
              className="w-full text-left px-2 py-1 text-sm rounded hover:bg-accent transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-0 text-foreground"
              onClick={() => {
                // Page selection would trigger navigation
                console.log('Selected page:', page.id);
              }}
              aria-label={`Page: ${page.title}${depth > 0 ? `, nested level ${depth}` : ''}`}
              title={`Open page: ${page.title}`}
            >
              {page.title}
            </button>
            {renderPages(page.id, depth + 1)}
          </li>
        ))}
      </ul>
    );
  };

  return (
    <aside 
      className="w-64 border-r bg-background/60 backdrop-blur-sm flex flex-col h-full"
      role="navigation"
      aria-label="Workspace navigation"
    >
      {/* Skip to main content link for keyboard navigation */}
      <a
        ref={skipLinkRef}
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:top-0 focus:left-0 focus:z-50 focus:bg-primary focus:text-white focus:p-2"
        aria-label="Skip to main content"
      >
        Skip to main content
      </a>

      <div className="p-4 border-b">
        <h2 
          className="text-sm font-semibold tracking-tight text-muted-foreground mb-3"
          id="workspaces-heading"
        >
          Workspaces
        </h2>
        
        {error && (
          <div 
            className="text-xs text-red-600 mb-2 p-2 bg-red-50 rounded border border-red-200"
            role="alert"
            aria-live="polite"
          >
            <span className="font-semibold">Error:</span> {error}
          </div>
        )}
        
        <div 
          className="space-y-2"
          role="listbox"
          aria-labelledby="workspaces-heading"
          aria-describedby="workspaces-help"
          ref={workspaceListRef}
        >
          {workspaces.map(ws => (
            <button
              key={ws.id}
              data-workspace-id={ws.id}
              onClick={() => setSelectedWorkspace(ws.id)}
              onKeyDown={(e) => handleWorkspaceKeyDown(e, ws.id)}
              className={`w-full text-left px-3 py-2 text-sm rounded transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-0 ${
                selectedWorkspace === ws.id
                  ? 'bg-primary text-primary-foreground shadow-sm'
                  : 'hover:bg-accent text-foreground'
              }`}
              role="option"
              aria-selected={selectedWorkspace === ws.id}
              aria-label={`Workspace: ${ws.name}${selectedWorkspace === ws.id ? ', currently selected' : ''}`}
              title={`Open workspace: ${ws.name}`}
            >
              {ws.name}
            </button>
          ))}
          {isLoading && workspaces.length === 0 && (
            <div 
              className="text-xs text-muted-foreground p-2"
              aria-live="polite"
              aria-busy="true"
            >
              Loading workspaces...
            </div>
          )}
        </div>
        
        <p id="workspaces-help" className="sr-only">
          Use arrow keys to navigate workspaces, enter to select
        </p>
        
        <button
          onClick={() => setShowCreateWorkspace(true)}
          className="w-full mt-2 px-3 py-2 text-xs text-center rounded border border-dashed border-muted-foreground/40 hover:bg-accent hover:border-muted-foreground/70 transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-0 text-foreground hover:text-foreground"
          aria-label="Create new workspace"
          title="Create a new workspace"
        >
          + New Workspace
        </button>
      </div>

      {showCreateWorkspace && (
        <div className="p-4 border-b bg-accent/50">
          <form onSubmit={handleCreateWorkspace} className="space-y-2" aria-label="Create workspace form">
            <label htmlFor="workspace-name" className="sr-only">
              Workspace name
            </label>
            <input
              id="workspace-name"
              ref={workspaceInputRef}
              type="text"
              placeholder="Workspace name"
              value={newWorkspaceName}
              onChange={e => setNewWorkspaceName(e.target.value)}
              className="w-full px-2 py-1 text-sm border rounded focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary bg-background text-foreground border-input"
              disabled={isLoading}
              aria-required="true"
              aria-describedby="workspace-name-help"
            />
            <p id="workspace-name-help" className="sr-only">
              Enter a descriptive name for your workspace
            </p>
            <div className="flex gap-2">
              <button
                type="submit"
                className="flex-1 px-2 py-1 text-xs bg-primary text-primary-foreground rounded hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-offset-0 focus:ring-primary disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                disabled={isLoading}
                aria-label="Create workspace"
              >
                Create
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowCreateWorkspace(false);
                  setNewWorkspaceName('');
                }}
                className="flex-1 px-2 py-1 text-xs border rounded hover:bg-accent focus:outline-none focus:ring-2 focus:ring-offset-0 focus:ring-primary disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                disabled={isLoading}
                aria-label="Cancel creating workspace"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="flex-1 overflow-auto">
        {selectedWorkspace && (
          <div className="p-4">
            <div className="flex items-center justify-between mb-3">
              <h3 
                className="text-xs font-semibold text-muted-foreground"
                id="pages-heading"
              >
                Pages
              </h3>
              <button
                onClick={() => setShowCreatePage(true)}
                className="text-xs px-2 py-1 rounded hover:bg-accent focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-0 transition-colors"
                aria-label="Create new page in workspace"
                title="Create a new page"
              >
                +
              </button>
            </div>

            {showCreatePage && (
              <form onSubmit={handleCreatePage} className="mb-3 p-2 bg-accent/50 rounded space-y-2" aria-label="Create page form">
                <label htmlFor="page-title" className="sr-only">
                  Page title
                </label>
                <input
                  id="page-title"
                  ref={pageInputRef}
                  type="text"
                  placeholder="Page title"
                  value={newPageTitle}
                  onChange={e => setNewPageTitle(e.target.value)}
                  className="w-full px-2 py-1 text-sm border rounded focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary bg-background text-foreground border-input"
                  disabled={isLoading}
                  aria-required="true"
                  aria-describedby="page-title-help"
                />
                <p id="page-title-help" className="sr-only">
                  Enter a title for your new page
                </p>
                <div className="flex gap-2">
                  <button
                    type="submit"
                    className="flex-1 px-2 py-1 text-xs bg-primary text-primary-foreground rounded hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-offset-0 focus:ring-primary disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                    disabled={isLoading}
                    aria-label="Create page"
                  >
                    Create
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setShowCreatePage(false);
                      setNewPageTitle('');
                    }}
                    className="flex-1 px-2 py-1 text-xs border rounded hover:bg-accent focus:outline-none focus:ring-2 focus:ring-offset-0 focus:ring-primary disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                    disabled={isLoading}
                    aria-label="Cancel creating page"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            )}

            {isLoading && pages.length === 0 && (
              <div 
                className="text-xs text-muted-foreground p-2"
                aria-live="polite"
                aria-busy="true"
              >
                Loading pages...
              </div>
            )}
            
            {pages.length === 0 && !isLoading && !showCreatePage && (
              <Card 
                className="p-3 text-xs text-muted-foreground border-dashed"
                role="status"
                aria-live="polite"
              >
                No pages yet. Create your first page!
              </Card>
            )}

            {pages.length > 0 && (
              <nav aria-labelledby="pages-heading">
                {renderPages()}
              </nav>
            )}
          </div>
        )}
      </div>
    </aside>
  );
}

