'use client'

import { useEffect, useState } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { workspaceApi, type Workspace, type Page } from '@/services/workspaceApi'
import { blockApi, type Block } from '@/services/blockApi'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { toast } from 'sonner'
import { PageEditor } from '@/components/PageEditor'
import { useAuthStore } from '@/features/auth/store/authStore'

export default function WorkspacePage() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const { isAuthenticated } = useAuthStore()
  const workspaceId = searchParams?.get('workspaceId') || ''
  const pageId = searchParams?.get('pageId') || ''

  const [workspace, setWorkspace] = useState<Workspace | null>(null)
  const [pages, setPages] = useState<Page[]>([])
  const [selectedPage, setSelectedPage] = useState<Page | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [showCreatePage, setShowCreatePage] = useState(false)
  const [newPageTitle, setNewPageTitle] = useState('')
  const [isCreating, setIsCreating] = useState(false)
  const [isEditingTitle, setIsEditingTitle] = useState(false)
  const [editedTitle, setEditedTitle] = useState('')

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login')
      return
    }
  }, [isAuthenticated, router])

  useEffect(() => {
    if (workspaceId && isAuthenticated) {
      loadWorkspace()
      loadPages()
    }
  }, [workspaceId, isAuthenticated])

  useEffect(() => {
    if (pageId && pages.length > 0) {
      const page = pages.find(p => p.id === pageId)
      if (page) {
        setSelectedPage(page)
        setEditedTitle(page.title)
      }
    }
  }, [pageId, pages])

  const loadWorkspace = async () => {
    try {
      const data = await workspaceApi.getWorkspace(workspaceId)
      setWorkspace(data)
    } catch (error) {
      console.error('Failed to load workspace:', error)
      toast.error('Failed to load workspace')
      router.push('/')
    }
  }

  const loadPages = async () => {
    try {
      setIsLoading(true)
      const data = await workspaceApi.listPages(workspaceId)
      setPages(data)
      
      // If pageId is in URL, select that page
      if (pageId) {
        const page = data.find(p => p.id === pageId)
        if (page) {
          setSelectedPage(page)
          setEditedTitle(page.title)
        }
      } else if (data.length > 0 && !selectedPage) {
        // Auto-select first page if no page selected
        setSelectedPage(data[0])
        setEditedTitle(data[0].title)
        router.push(`/workspace?workspaceId=${workspaceId}&pageId=${data[0].id}`)
      }
    } catch (error) {
      console.error('Failed to load pages:', error)
      toast.error('Failed to load pages')
    } finally {
      setIsLoading(false)
    }
  }

  const handleCreatePage = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newPageTitle.trim()) {
      toast.error('Page title is required')
      return
    }

    try {
      setIsCreating(true)
      const page = await workspaceApi.createPage({
        workspaceId,
        title: newPageTitle.trim(),
      })
      toast.success('Page created successfully')
      setNewPageTitle('')
      setShowCreatePage(false)
      await loadPages()
      // Navigate to the new page
      router.push(`/workspace?workspaceId=${workspaceId}&pageId=${page.id}`)
    } catch (error) {
      console.error('Failed to create page:', error)
      toast.error(error instanceof Error ? error.message : 'Failed to create page')
    } finally {
      setIsCreating(false)
    }
  }

  const handleSaveTitle = async () => {
    if (!selectedPage || !editedTitle.trim()) {
      setIsEditingTitle(false)
      return
    }

    if (editedTitle === selectedPage.title) {
      setIsEditingTitle(false)
      return
    }

    try {
      await workspaceApi.updatePage(selectedPage.id, {
        title: editedTitle.trim(),
      })
      toast.success('Page title updated')
      setIsEditingTitle(false)
      await loadPages()
    } catch (error) {
      console.error('Failed to update page title:', error)
      toast.error('Failed to update page title')
    }
  }

  const handleDeletePage = async (id: string, title: string) => {
    if (!confirm(`Are you sure you want to delete "${title}"? This action cannot be undone.`)) {
      return
    }

    try {
      await workspaceApi.deletePage(id)
      toast.success('Page deleted successfully')
      if (selectedPage?.id === id) {
        setSelectedPage(null)
        router.push(`/workspace?workspaceId=${workspaceId}`)
      }
      await loadPages()
    } catch (error) {
      console.error('Failed to delete page:', error)
      toast.error(error instanceof Error ? error.message : 'Failed to delete page')
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
    <div className="min-h-screen bg-neutral-50 flex">
      {/* Sidebar */}
      <div className="w-64 bg-white border-r border-neutral-200 flex flex-col">
        <div className="p-4 border-b border-neutral-200">
          <Button
            variant="outline"
            onClick={() => router.push('/')}
            className="w-full mb-2"
          >
            ← Back to Workspaces
          </Button>
          <h2 className="font-semibold text-lg">{workspace?.name || 'Workspace'}</h2>
        </div>

        <div className="flex-1 overflow-y-auto p-4">
          <Button
            onClick={() => setShowCreatePage(true)}
            className="w-full mb-4"
          >
            + New Page
          </Button>

          {showCreatePage && (
            <Card className="mb-4">
              <CardContent className="p-4">
                <form onSubmit={handleCreatePage} className="space-y-2">
                  <Input
                    placeholder="Page title"
                    value={newPageTitle}
                    onChange={(e) => setNewPageTitle(e.target.value)}
                    autoFocus
                    disabled={isCreating}
                  />
                  <div className="flex gap-2">
                    <Button type="submit" size="sm" disabled={isCreating || !newPageTitle.trim()}>
                      Create
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setShowCreatePage(false)
                        setNewPageTitle('')
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

          <div className="space-y-1">
            {pages.map((page) => (
              <div
                key={page.id}
                className={`p-2 rounded cursor-pointer hover:bg-neutral-100 flex justify-between items-center ${
                  selectedPage?.id === page.id ? 'bg-neutral-100' : ''
                }`}
                onClick={() => {
                  setSelectedPage(page)
                  setEditedTitle(page.title)
                  router.push(`/workspace?workspaceId=${workspaceId}&pageId=${page.id}`)
                }}
              >
                <span className="truncate flex-1">{page.title}</span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation()
                    handleDeletePage(page.id, page.title)
                  }}
                  className="text-red-600 hover:text-red-700 h-6 w-6 p-0"
                >
                  ×
                </Button>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-y-auto">
        {selectedPage ? (
          <div className="max-w-4xl mx-auto p-8">
            {/* Page Header */}
            <div className="mb-6">
              {isEditingTitle ? (
                <div className="flex gap-2 items-center">
                  <Input
                    value={editedTitle}
                    onChange={(e) => setEditedTitle(e.target.value)}
                    onBlur={handleSaveTitle}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        handleSaveTitle()
                      } else if (e.key === 'Escape') {
                        setEditedTitle(selectedPage.title)
                        setIsEditingTitle(false)
                      }
                    }}
                    autoFocus
                    className="text-3xl font-bold border-none shadow-none p-0 h-auto"
                  />
                </div>
              ) : (
                <h1
                  className="text-3xl font-bold cursor-pointer hover:bg-neutral-100 p-2 -m-2 rounded"
                  onClick={() => setIsEditingTitle(true)}
                >
                  {selectedPage.title}
                </h1>
              )}
            </div>

            {/* Blocks */}
            <PageEditor pageId={selectedPage.id} />
          </div>
        ) : (
          <div className="flex items-center justify-center h-full">
            <div className="text-center">
              <p className="text-neutral-600 mb-4">Select a page to start editing</p>
              {pages.length === 0 && (
                <Button onClick={() => setShowCreatePage(true)}>
                  Create your first page
                </Button>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
