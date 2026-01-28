'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { workspaceApi, type Workspace, type Page } from '@/services/workspaceApi'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { toast } from 'sonner'
import Link from 'next/link'
import { useAuthStore } from '@/features/auth/store/authStore'
import { useAuth } from '@/shared/contexts/AuthProvider'

export default function Home() {
  const router = useRouter()
  const { isAuthenticated, user } = useAuthStore()
  const { handleLogout } = useAuth()
  const [workspaces, setWorkspaces] = useState<Workspace[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [newWorkspaceName, setNewWorkspaceName] = useState('')
  const [isCreating, setIsCreating] = useState(false)

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login')
      return
    }
    loadWorkspaces()
  }, [isAuthenticated, router])

  const loadWorkspaces = async () => {
    try {
      setIsLoading(true)
      const data = await workspaceApi.listWorkspaces()
      setWorkspaces(data)
    } catch (error) {
      console.error('Failed to load workspaces:', error)
      toast.error('Failed to load workspaces')
    } finally {
      setIsLoading(false)
    }
  }

  const handleCreateWorkspace = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newWorkspaceName.trim()) {
      toast.error('Workspace name is required')
      return
    }

    try {
      setIsCreating(true)
      const workspace = await workspaceApi.createWorkspace({
        name: newWorkspaceName.trim(),
      })
      toast.success('Workspace created successfully')
      setNewWorkspaceName('')
      setShowCreateDialog(false)
      await loadWorkspaces()
      // Navigate to the new workspace
      router.push(`/workspace?workspaceId=${workspace.id}`)
    } catch (error) {
      console.error('Failed to create workspace:', error)
      toast.error(error instanceof Error ? error.message : 'Failed to create workspace')
    } finally {
      setIsCreating(false)
    }
  }

  const handleDeleteWorkspace = async (id: string, name: string) => {
    if (!confirm(`Are you sure you want to delete "${name}"? This action cannot be undone.`)) {
      return
    }

    try {
      await workspaceApi.deleteWorkspace(id)
      toast.success('Workspace deleted successfully')
      await loadWorkspaces()
    } catch (error) {
      console.error('Failed to delete workspace:', error)
      toast.error(error instanceof Error ? error.message : 'Failed to delete workspace')
    }
  }

  // Show loading or redirect if not authenticated
  if (!isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-neutral-50">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-neutral-50">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-neutral-50 to-neutral-100 p-8">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-8 flex justify-between items-center">
          <div>
            <h1 className="text-4xl font-bold text-neutral-900">GOALKeeper Plan</h1>
            <p className="text-neutral-600 mt-2">
              Welcome, {user?.name || user?.email || 'User'}! Your Notion-like workspace
            </p>
          </div>
          <div className="flex gap-2">
            <Button onClick={() => setShowCreateDialog(true)}>
              + New Workspace
            </Button>
            <Button variant="outline" onClick={handleLogout}>
              Logout
            </Button>
          </div>
        </div>

        {/* Create Workspace Dialog */}
        {showCreateDialog && (
          <Card className="mb-6">
            <CardHeader>
              <CardTitle>Create New Workspace</CardTitle>
              <CardDescription>Give your workspace a name to get started</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleCreateWorkspace} className="space-y-4">
                <Input
                  placeholder="Workspace name"
                  value={newWorkspaceName}
                  onChange={(e) => setNewWorkspaceName(e.target.value)}
                  autoFocus
                  disabled={isCreating}
                />
                <div className="flex gap-2">
                  <Button type="submit" disabled={isCreating || !newWorkspaceName.trim()}>
                    {isCreating ? 'Creating...' : 'Create'}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setShowCreateDialog(false)
                      setNewWorkspaceName('')
                    }}
                    disabled={isCreating}
                  >
                    Cancel
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>
        )}

        {/* Workspaces Grid */}
        {workspaces.length === 0 ? (
          <Card>
            <CardContent className="py-12 text-center">
              <p className="text-neutral-600 mb-4">No workspaces yet</p>
              <Button onClick={() => setShowCreateDialog(true)}>
                Create your first workspace
              </Button>
            </CardContent>
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {workspaces.map((workspace) => (
              <Card key={workspace.id} className="hover:shadow-lg transition-shadow">
                <CardHeader>
                  <CardTitle className="text-xl">{workspace.name}</CardTitle>
                  {workspace.description && (
                    <CardDescription>{workspace.description}</CardDescription>
                  )}
                </CardHeader>
                <CardContent>
                  <div className="flex gap-2 mt-4">
                    <Link href={`/workspace?workspaceId=${workspace.id}`} className="flex-1">
                      <Button className="w-full">Open</Button>
                    </Link>
                    <Button
                      variant="outline"
                      onClick={() => handleDeleteWorkspace(workspace.id, workspace.name)}
                      className="text-red-600 hover:text-red-700"
                    >
                      Delete
                    </Button>
                  </div>
                  <p className="text-xs text-neutral-500 mt-4">
                    Created {new Date(workspace.created_at).toLocaleDateString()}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
