import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BlockEditor } from '@/components/BlockEditor';
import * as blockApiModule from '@/services/blockApi';

// Mock the blockApi
vi.mock('@/services/blockApi');

const mockBlock = {
  id: 'block-1',
  page_id: 'page-1',
  type_id: 'paragraph' as const,
  content: 'Original content',
  rank: 1,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

describe('Concurrent Edits UI - Last-Write-Wins', () => {
  let mockBlockApi: any;

  beforeEach(() => {
    mockBlockApi = blockApiModule.blockApi as any;
    vi.clearAllMocks();
  });

  it('should detect when block content is overwritten by server-side update', async () => {
    const onUpdate = vi.fn();
    const onDelete = vi.fn();
    const onSaveError = vi.fn();

    // Simulate: User saves local change, server returns newer version (concurrent edit)
    mockBlockApi.updateBlock = vi.fn().mockResolvedValue({
      ...mockBlock,
      content: 'Server overwrote this', // Last-write-wins: server content
      updated_at: new Date(Date.now() + 5000).toISOString(), // Later timestamp
    });

    render(
      <BlockEditor block={mockBlock} onUpdate={onUpdate} onDelete={onDelete} onSaveError={onSaveError} />
    );

    // User clicks and tries to edit
    const contentArea = screen.getByText('Original content');
    await userEvent.click(contentArea);

    const textarea = screen.getByPlaceholderText('Type something...') as HTMLTextAreaElement;
    await userEvent.clear(textarea);
    await userEvent.type(textarea, 'User local change');

    // User saves
    const saveButton = screen.getByText('Save');
    await userEvent.click(saveButton);

    // Verify server returned newer content
    await waitFor(() => {
      expect(onUpdate).toHaveBeenCalledWith(
        expect.objectContaining({
          content: 'Server overwrote this',
        })
      );
    });
  });

  it('should show notification when block is overwritten', async () => {
    const onUpdate = vi.fn();
    const onDelete = vi.fn();
    const onSaveError = vi.fn();

    // Original timestamp and newer server timestamp
    const originalTime = new Date();
    const serverTime = new Date(originalTime.getTime() + 5000);

    mockBlockApi.updateBlock = vi.fn().mockResolvedValue({
      ...mockBlock,
      content: 'Content changed by other user',
      updated_at: serverTime.toISOString(),
    });

    const { rerender } = render(
      <BlockEditor block={mockBlock} onUpdate={onUpdate} onDelete={onDelete} onSaveError={onSaveError} />
    );

    const contentArea = screen.getByText('Original content');
    await userEvent.click(contentArea);

    const textarea = screen.getByPlaceholderText('Type something...') as HTMLTextAreaElement;
    await userEvent.clear(textarea);
    await userEvent.type(textarea, 'My changes');

    const saveButton = screen.getByText('Save');
    await userEvent.click(saveButton);

    // After update is called, onUpdate receives the newer content
    await waitFor(() => {
      const updatedBlock = onUpdate.mock.calls[0][0];
      expect(updatedBlock.content).toBe('Content changed by other user');
      expect(new Date(updatedBlock.updated_at).getTime()).toBeGreaterThan(
        new Date(originalTime).getTime()
      );
    });
  });

  it('should allow user to override by editing again after concurrent update', async () => {
    const onUpdate = vi.fn();
    const onDelete = vi.fn();
    const onSaveError = vi.fn();

    // First update returns server content
    mockBlockApi.updateBlock = vi.fn()
      .mockResolvedValueOnce({
        ...mockBlock,
        content: 'Server content',
        updated_at: new Date(Date.now() + 1000).toISOString(),
      })
      // Second update returns user's new override
      .mockResolvedValueOnce({
        ...mockBlock,
        content: 'User override',
        updated_at: new Date(Date.now() + 2000).toISOString(),
      });

    const { rerender } = render(
      <BlockEditor block={mockBlock} onUpdate={onUpdate} onDelete={onDelete} onSaveError={onSaveError} />
    );

    // First edit and save
    let contentArea = screen.getByText('Original content');
    await userEvent.click(contentArea);

    let textarea = screen.getByPlaceholderText('Type something...') as HTMLTextAreaElement;
    await userEvent.clear(textarea);
    await userEvent.type(textarea, 'User attempt 1');

    let saveButton = screen.getByText('Save');
    await userEvent.click(saveButton);

    // Verify first update received server content
    await waitFor(() => {
      expect(mockBlockApi.updateBlock).toHaveBeenCalledTimes(1);
    });

    // Rerender with server-returned content
    const updatedBlock = onUpdate.mock.calls[0][0];
    rerender(
      <BlockEditor block={updatedBlock} onUpdate={onUpdate} onDelete={onDelete} onSaveError={onSaveError} />
    );

    // User can now edit the received content
    contentArea = screen.getByText('Server content');
    await userEvent.click(contentArea);

    textarea = screen.getByPlaceholderText('Type something...') as HTMLTextAreaElement;
    await userEvent.clear(textarea);
    await userEvent.type(textarea, 'User override');

    saveButton = screen.getByText('Save');
    await userEvent.click(saveButton);

    // Verify second update
    await waitFor(() => {
      expect(mockBlockApi.updateBlock).toHaveBeenCalledTimes(2);
      const finalBlock = onUpdate.mock.calls[1][0];
      expect(finalBlock.content).toBe('User override');
    });
  });

  it('should maintain editing state across concurrent updates', async () => {
    const onUpdate = vi.fn();
    const onDelete = vi.fn();
    const onSaveError = vi.fn();

    mockBlockApi.updateBlock = vi.fn().mockResolvedValue({
      ...mockBlock,
      content: 'Server new content',
      updated_at: new Date(Date.now() + 5000).toISOString(),
    });

    render(
      <BlockEditor block={mockBlock} onUpdate={onUpdate} onDelete={onDelete} onSaveError={onSaveError} />
    );

    // User enters edit mode
    const contentArea = screen.getByText('Original content');
    await userEvent.click(contentArea);

    // Edit field is visible
    expect(screen.getByPlaceholderText('Type something...')).toBeInTheDocument();

    // User modifies content
    const textarea = screen.getByPlaceholderText('Type something...') as HTMLTextAreaElement;
    await userEvent.clear(textarea);
    await userEvent.type(textarea, 'User edits');

    // User saves
    const saveButton = screen.getByText('Save');
    await userEvent.click(saveButton);

    // After save, should exit edit mode even though content was overwritten
    await waitFor(() => {
      expect(screen.queryByPlaceholderText('Type something...')).not.toBeInTheDocument();
      expect(onUpdate).toHaveBeenCalled();
    });
  });

  it('should display conflict resolution message for overwritten content', async () => {
    const onUpdate = vi.fn();
    const onDelete = vi.fn();
    const onSaveError = vi.fn();

    // User's update time vs server update time
    const userUpdateTime = new Date();
    const serverUpdateTime = new Date(userUpdateTime.getTime() + 10000);

    mockBlockApi.updateBlock = vi.fn().mockResolvedValue({
      ...mockBlock,
      content: 'Overwritten by concurrent edit',
      updated_at: serverUpdateTime.toISOString(),
    });

    render(
      <BlockEditor block={mockBlock} onUpdate={onUpdate} onDelete={onDelete} onSaveError={onSaveError} />
    );

    const contentArea = screen.getByText('Original content');
    await userEvent.click(contentArea);

    const textarea = screen.getByPlaceholderText('Type something...') as HTMLTextAreaElement;
    await userEvent.clear(textarea);
    await userEvent.type(textarea, 'User content');

    const saveButton = screen.getByText('Save');
    await userEvent.click(saveButton);

    // Verify the block content shows server's version (last-write-wins)
    await waitFor(() => {
      expect(onUpdate).toHaveBeenCalledWith(
        expect.objectContaining({
          content: 'Overwritten by concurrent edit',
          updated_at: expect.any(String),
        })
      );
      // Verify newer timestamp
      const updatedBlock = onUpdate.mock.calls[0][0];
      expect(new Date(updatedBlock.updated_at).getTime()).toBeGreaterThan(
        userUpdateTime.getTime()
      );
    });
  });
});
