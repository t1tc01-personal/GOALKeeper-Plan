'use client';

import React, { useEffect, useState } from 'react';
import { WorkspacePageShell } from '../components/WorkspacePageShell';
import { PageEditor } from '../components/PageEditor';
import { ShareDialog } from '../components/ShareDialog';
import { workspaceApi, type Page } from '@/services/workspaceApi';
import { useSearchParams } from 'next/navigation';

export function WorkspacePageView() {
  const searchParams = useSearchParams();
  const pageId = searchParams?.get('pageId') || '';
  
  const [page, setPage] = useState<Page | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const [editedTitle, setEditedTitle] = useState('');
  const [showShareDialog, setShowShareDialog] = useState(false);

  useEffect(() => {
    if (pageId) {
      loadPage(pageId);
    }
  }, [pageId]);

  const loadPage = async (id: string) => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await workspaceApi.getPage(id);
      setPage(data);
      setEditedTitle(data.title);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load page');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSaveTitle = async () => {
    if (!page || !editedTitle.trim()) {
      setIsEditingTitle(false);
      return;
    }

    if (editedTitle === page.title) {
      setIsEditingTitle(false);
      return;
    }

    try {
      setIsLoading(true);
      setError(null);
      await workspaceApi.updatePage(page.id, {
        title: editedTitle,
      });
      setPage({ ...page, title: editedTitle });
      setIsEditingTitle(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update page title');
      setEditedTitle(page.title);
    } finally {
      setIsLoading(false);
    }
  };

  const handleDeletePage = async () => {
    if (!page || !confirm('Are you sure you want to delete this page?')) {
      return;
    }

    try {
      setIsLoading(true);
      setError(null);
      await workspaceApi.deletePage(page.id);
      // Redirect to workspace or show message
      window.location.href = '/';
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete page');
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading && !page) {
    return (
      <WorkspacePageShell>
        <div className="flex items-center justify-center h-full">
          <div className="text-muted-foreground">Loading page...</div>
        </div>
      </WorkspacePageShell>
    );
  }

  if (!page && !isLoading) {
    return (
      <WorkspacePageShell>
        <div className="flex items-center justify-center h-full">
          <div className="text-muted-foreground">
            {error || 'No page selected. Select a page from the sidebar.'}
          </div>
        </div>
      </WorkspacePageShell>
    );
  }

  return (
    <WorkspacePageShell>
      {error && (
        <div className="mb-4 p-3 bg-red-50 text-red-700 text-sm rounded">
          {error}
        </div>
      )}
      
      <header className="mb-6 flex items-center justify-between">
        {isEditingTitle ? (
          <div className="flex-1 flex gap-2">
            <input
              type="text"
              value={editedTitle}
              onChange={e => setEditedTitle(e.target.value)}
              onBlur={handleSaveTitle}
              onKeyDown={e => {
                if (e.key === 'Enter') handleSaveTitle();
                if (e.key === 'Escape') {
                  setIsEditingTitle(false);
                  setEditedTitle(page?.title || '');
                }
              }}
              className="flex-1 text-2xl font-semibold border rounded px-2 py-1"
              autoFocus
              disabled={isLoading}
            />
          </div>
        ) : (
          <h1
            onClick={() => setIsEditingTitle(true)}
            className="text-2xl font-semibold tracking-tight cursor-pointer hover:text-muted-foreground transition-colors"
          >
            {page?.title || 'Untitled page'}
          </h1>
        )}
        <div className="ml-4 flex gap-2">
          <button
            onClick={() => setShowShareDialog(true)}
            className="px-3 py-2 text-sm text-blue-600 hover:bg-blue-50 rounded transition-colors"
            disabled={isLoading}
          >
            Share
          </button>
          <button
            onClick={handleDeletePage}
            className="px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded transition-colors"
            disabled={isLoading}
          >
            Delete
          </button>
        </div>
      </header>

      <div className="text-xs text-muted-foreground mb-4">
        Created: {page?.created_at ? new Date(page.created_at).toLocaleDateString() : 'Unknown'}
      </div>

      <PageEditor pageId={page?.id} />

      {showShareDialog && page && (
        <ShareDialog pageId={page.id} onClose={() => setShowShareDialog(false)} />
      )}
    </WorkspacePageShell>
  );
}

export default WorkspacePageView;

