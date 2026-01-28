import React, { useState, useEffect } from 'react';
import { workspaceApi } from '../services/workspaceApi';
import { SimpleErrorResponse } from '../shared/interfaces';

interface Collaborator {
  id: string;
  user_id: string;
  page_id: string;
  role: 'viewer' | 'editor' | 'owner';
  created_at: string;
}

interface ShareDialogProps {
  pageId: string;
  onClose: () => void;
}

/**
 * ShareDialog component for managing page sharing and permissions.
 * Allows users to grant/revoke access to pages with different roles.
 */
export function ShareDialog({ pageId, onClose }: ShareDialogProps) {
  const [collaborators, setCollaborators] = useState<Collaborator[]>([]);
  const [userEmail, setUserEmail] = useState('');
  const [role, setRole] = useState<'viewer' | 'editor'>('viewer');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    loadCollaborators();
  }, [pageId]);

  const loadCollaborators = async () => {
    try {
      setLoading(true);
      const result = await workspaceApi.listCollaborators(pageId);
      if (result.success && result.data?.collaborators) {
        setCollaborators(result.data.collaborators);
      }
    } catch (err) {
      setError('Failed to load collaborators');
    } finally {
      setLoading(false);
    }
  };

  const handleGrantAccess = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!userEmail.trim()) {
      setError('Please enter a user email');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      // In a real app, we'd first resolve the email to a user ID
      // For now, we'll use the email as a placeholder
      const userId = userEmail; // This should come from user lookup

      const result = await workspaceApi.grantAccess(pageId, userId, role);
      if (result.success) {
        setSuccess(`Access granted to ${userEmail} as ${role}`);
        setUserEmail('');
        setRole('viewer');
        await loadCollaborators();
        setTimeout(() => setSuccess(null), 3000);
      } else {
        setError(result.error?.message || 'Failed to grant access');
      }
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  };

  const handleRevokeAccess = async (userId: string) => {
    try {
      setLoading(true);
      setError(null);

      const result = await workspaceApi.revokeAccess(pageId, userId);
      if (result.success) {
        setSuccess('Access revoked');
        await loadCollaborators();
        setTimeout(() => setSuccess(null), 3000);
      } else {
        setError(result.error?.message || 'Failed to revoke access');
      }
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        {/* Header */}
        <div className="flex justify-between items-center p-6 border-b">
          <h2 className="text-lg font-semibold">Share Page</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
            aria-label="Close"
          >
            ✕
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-4">
          {/* Error message */}
          {error && (
            <div className="bg-red-50 border border-red-200 rounded px-3 py-2 text-sm text-red-700">
              {error}
            </div>
          )}

          {/* Success message */}
          {success && (
            <div className="bg-green-50 border border-green-200 rounded px-3 py-2 text-sm text-green-700">
              {success}
            </div>
          )}

          {/* Grant Access Form */}
          <form onSubmit={handleGrantAccess} className="space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                User Email
              </label>
              <input
                type="email"
                value={userEmail}
                onChange={(e) => setUserEmail(e.target.value)}
                placeholder="user@example.com"
                disabled={loading}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Role
              </label>
              <select
                value={role}
                onChange={(e) => setRole(e.target.value as 'viewer' | 'editor')}
                disabled={loading}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="viewer">Viewer (Read-only)</option>
                <option value="editor">Editor (Can edit)</option>
              </select>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-medium py-2 rounded-md transition"
            >
              {loading ? 'Sharing...' : 'Grant Access'}
            </button>
          </form>

          {/* Collaborators List */}
          <div className="border-t pt-4">
            <h3 className="font-medium text-sm text-gray-900 mb-3">Collaborators</h3>
            {collaborators.length === 0 ? (
              <p className="text-sm text-gray-500">No collaborators yet</p>
            ) : (
              <div className="space-y-2">
                {collaborators.map((collab) => (
                  <div
                    key={collab.id}
                    className="flex justify-between items-center p-2 bg-gray-50 rounded"
                  >
                    <div className="flex-1">
                      <p className="text-sm font-medium text-gray-900">
                        {collab.user_id}
                      </p>
                      <p className="text-xs text-gray-500 capitalize">{collab.role}</p>
                    </div>
                    <button
                      onClick={() => handleRevokeAccess(collab.user_id)}
                      disabled={loading}
                      className="text-red-600 hover:text-red-800 disabled:text-gray-400 text-sm font-medium"
                    >
                      Remove
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
