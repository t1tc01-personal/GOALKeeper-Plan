import React from 'react';

interface PermissionGuardProps {
  canEdit: boolean;
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

/**
 * PermissionGuard component that disables editing UI for users without edit permission.
 * Shows a read-only message when the user doesn't have permission to edit.
 */
export function PermissionGuard({ canEdit, children, fallback }: PermissionGuardProps) {
  if (!canEdit) {
    return (
      <div className="space-y-4">
        <div className="bg-yellow-50 border border-yellow-200 rounded px-4 py-3 text-sm text-yellow-800">
          <p className="font-medium">Read-only access</p>
          <p className="text-xs mt-1">You don't have permission to edit this page. Contact the owner to request edit access.</p>
        </div>
        {fallback}
      </div>
    );
  }

  return <>{children}</>;
}

/**
 * Hook to check if user has edit permission for a page
 */
export function usePagePermission(pageId: string, userRole?: string) {
  return {
    canEdit: userRole === 'editor' || userRole === 'owner',
    canView: !!userRole,
    role: userRole as 'viewer' | 'editor' | 'owner' | undefined,
  };
}
