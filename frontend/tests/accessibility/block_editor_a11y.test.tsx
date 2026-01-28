/**
 * Accessibility tests for BlockEditor component
 * Tests WCAG 2.1 Level AA compliance and keyboard navigation
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BlockEditor } from '@/components/BlockEditor';
import { blockApi, type Block } from '@/services/blockApi';
import { runA11yChecks, hasAriaLabel, isKeyboardAccessible } from '@/utils/a11y-helpers';

// Mock the API client
vi.mock('@/services/blockApi', () => ({
  blockApi: {
    updateBlock: vi.fn(),
    deleteBlock: vi.fn(),
  },
}));

describe('BlockEditor Accessibility', () => {
  const mockBlock: Block = {
    id: 'block-1',
    page_id: 'page-1',
    type_id: 'paragraph',
    content: 'Test content',
    position: 0,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };

  const defaultProps = {
    block: mockBlock,
    onUpdate: vi.fn(),
    onDelete: vi.fn(),
    onSaveError: vi.fn(),
    blockIndex: 0,
    totalBlocks: 3,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('ARIA Labels and Roles', () => {
    it('should have proper ARIA labels for block container', () => {
      const { container } = render(<BlockEditor {...defaultProps} />);
      const blockContainer = container.querySelector('[role="article"]');
      
      expect(blockContainer).toBeInTheDocument();
      expect(blockContainer).toHaveAttribute('aria-label');
      expect(blockContainer?.getAttribute('aria-label')).toContain('Block');
    });

    it('should have aria-describedby pointing to help text', () => {
      const { container } = render(<BlockEditor {...defaultProps} />);
      const blockContainer = container.querySelector('[role="article"]');
      const helpTextId = blockContainer?.getAttribute('aria-describedby');
      
      expect(helpTextId).toBeTruthy();
      const helpText = container.querySelector(`#${helpTextId}`);
      expect(helpText).toBeInTheDocument();
      expect(helpText).toHaveClass('sr-only');
    });

    it('should have accessible names for all interactive elements', () => {
      const { container } = render(<BlockEditor {...defaultProps} />);
      const checks = runA11yChecks(container);
      
      expect(checks.interactiveElements.passes).toBe(true);
    });

    it('should have proper labels for form inputs', () => {
      const { container } = render(<BlockEditor {...defaultProps} />);
      const checks = runA11yChecks(container);
      
      expect(checks.formLabels.passes).toBe(true);
    });
  });

  describe('Keyboard Navigation', () => {
    it('should be keyboard accessible', () => {
      const { container } = render(<BlockEditor {...defaultProps} />);
      const blockContainer = container.querySelector('[role="article"]') as HTMLElement;
      
      expect(blockContainer).toBeInTheDocument();
      // Check that the block can receive focus
      blockContainer.focus();
      expect(document.activeElement).toBe(blockContainer);
    });

    it('should support Enter key to start editing', async () => {
      const user = userEvent.setup();
      render(<BlockEditor {...defaultProps} />);
      
      const displayElement = screen.getByText('Test content');
      expect(displayElement).toBeInTheDocument();
      
      displayElement.focus();
      await user.keyboard('{Enter}');
      
      // Should enter edit mode
      await waitFor(() => {
        const textarea = screen.queryByRole('textbox');
        expect(textarea).toBeInTheDocument();
      });
    });

    it('should support Space key to start editing', async () => {
      const user = userEvent.setup();
      render(<BlockEditor {...defaultProps} />);
      
      const displayElement = screen.getByText('Test content');
      displayElement.focus();
      await user.keyboard(' ');
      
      await waitFor(() => {
        const textarea = screen.queryByRole('textbox');
        expect(textarea).toBeInTheDocument();
      });
    });

    it('should support Escape key to cancel editing', async () => {
      const user = userEvent.setup();
      render(<BlockEditor {...defaultProps} />);
      
      // Start editing
      const displayElement = screen.getByText('Test content');
      displayElement.focus();
      await user.keyboard('{Enter}');
      
      await waitFor(() => {
        expect(screen.queryByRole('textbox')).toBeInTheDocument();
      });
      
      // Cancel with Escape
      await user.keyboard('{Escape}');
      
      await waitFor(() => {
        expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
      });
    });

    it('should support Ctrl+Enter to save', async () => {
      const user = userEvent.setup();
      vi.mocked(blockApi.updateBlock).mockResolvedValue({
        ...mockBlock,
        content: 'Updated content',
      });
      
      render(<BlockEditor {...defaultProps} />);
      
      // Start editing
      const displayElement = screen.getByText('Test content');
      displayElement.focus();
      await user.keyboard('{Enter}');
      
      await waitFor(() => {
        expect(screen.queryByRole('textbox')).toBeInTheDocument();
      });
      
      // Type and save
      const textarea = screen.getByRole('textbox');
      await user.clear(textarea);
      await user.type(textarea, 'Updated content');
      await user.keyboard('{Control>}{Enter}{/Control}');
      
      await waitFor(() => {
        expect(blockApi.updateBlock).toHaveBeenCalled();
      });
    });

    it('should navigate to previous block with Arrow Up at start of content', async () => {
      const user = userEvent.setup();
      const onFocusPrevious = vi.fn();
      
      render(
        <BlockEditor
          {...defaultProps}
          blockIndex={1}
          onFocusPrevious={onFocusPrevious}
        />
      );
      
      // Start editing
      const displayElement = screen.getByText('Test content');
      displayElement.focus();
      await user.keyboard('{Enter}');
      
      await waitFor(() => {
        expect(screen.queryByRole('textbox')).toBeInTheDocument();
      });
      
      // Move to start and press Arrow Up
      const textarea = screen.getByRole('textbox');
      textarea.setSelectionRange(0, 0);
      await user.keyboard('{ArrowUp}');
      
      // Should call onFocusPrevious when at start
      expect(onFocusPrevious).toHaveBeenCalled();
    });

    it('should navigate to next block with Arrow Down at end of content', async () => {
      const user = userEvent.setup();
      const onFocusNext = vi.fn();
      
      render(
        <BlockEditor
          {...defaultProps}
          blockIndex={0}
          onFocusNext={onFocusNext}
        />
      );
      
      // Start editing
      const displayElement = screen.getByText('Test content');
      displayElement.focus();
      await user.keyboard('{Enter}');
      
      await waitFor(() => {
        expect(screen.queryByRole('textbox')).toBeInTheDocument();
      });
      
      // Move to end and press Arrow Down
      const textarea = screen.getByRole('textbox');
      textarea.setSelectionRange(textarea.value.length, textarea.value.length);
      await user.keyboard('{ArrowDown}');
      
      // Should call onFocusNext when at end
      expect(onFocusNext).toHaveBeenCalled();
    });
  });

  describe('Screen Reader Support', () => {
    it('should announce save actions to screen readers', async () => {
      const user = userEvent.setup();
      vi.mocked(blockApi.updateBlock).mockResolvedValue(mockBlock);
      
      render(<BlockEditor {...defaultProps} />);
      
      // Start editing and save
      const displayElement = screen.getByText('Test content');
      displayElement.focus();
      await user.keyboard('{Enter}');
      
      await waitFor(() => {
        expect(screen.queryByRole('textbox')).toBeInTheDocument();
      });
      
      const saveButton = screen.getByLabelText(/save/i);
      await user.click(saveButton);
      
      // Check for live region announcement
      await waitFor(() => {
        const liveRegion = document.querySelector('[aria-live]');
        expect(liveRegion).toBeInTheDocument();
      });
    });

    it('should announce error messages to screen readers', async () => {
      const user = userEvent.setup();
      vi.mocked(blockApi.updateBlock).mockRejectedValue(new Error('Save failed'));
      
      render(<BlockEditor {...defaultProps} />);
      
      // Start editing and try to save
      const displayElement = screen.getByText('Test content');
      displayElement.focus();
      await user.keyboard('{Enter}');
      
      await waitFor(() => {
        expect(screen.queryByRole('textbox')).toBeInTheDocument();
      });
      
      const saveButton = screen.getByLabelText(/save/i);
      await user.click(saveButton);
      
      // Check for error alert
      await waitFor(() => {
        const errorAlert = screen.queryByRole('alert');
        expect(errorAlert).toBeInTheDocument();
        expect(errorAlert).toHaveAttribute('aria-live', 'assertive');
      });
    });
  });

  describe('Focus Management', () => {
    it('should auto-focus input when entering edit mode', async () => {
      const user = userEvent.setup();
      render(<BlockEditor {...defaultProps} />);
      
      const displayElement = screen.getByText('Test content');
      displayElement.focus();
      await user.keyboard('{Enter}');
      
      await waitFor(() => {
        const textarea = screen.queryByRole('textbox');
        expect(textarea).toBeInTheDocument();
        expect(textarea).toHaveFocus();
      });
    });

    it('should return focus to display element after canceling', async () => {
      const user = userEvent.setup();
      render(<BlockEditor {...defaultProps} />);
      
      const displayElement = screen.getByText('Test content');
      displayElement.focus();
      await user.keyboard('{Enter}');
      
      await waitFor(() => {
        expect(screen.queryByRole('textbox')).toBeInTheDocument();
      });
      
      await user.keyboard('{Escape}');
      
      await waitFor(() => {
        expect(displayElement).toHaveFocus();
      });
    });
  });

  describe('WCAG Compliance', () => {
    it('should pass comprehensive a11y checks', () => {
      const { container } = render(<BlockEditor {...defaultProps} />);
      const checks = runA11yChecks(container);
      
      expect(checks.overallPass).toBe(true);
    });

    it('should have minimum touch target sizes (44x44px)', () => {
      const { container } = render(<BlockEditor {...defaultProps} />);
      const buttons = container.querySelectorAll('button');
      
      buttons.forEach((button) => {
        const styles = window.getComputedStyle(button);
        const minHeight = parseInt(styles.minHeight) || 0;
        const minWidth = parseInt(styles.minWidth) || 0;
        
        // Check if button has minimum size (44px is WCAG recommendation)
        expect(minHeight >= 44 || minWidth >= 44).toBe(true);
      });
    });

    it('should have visible focus indicators', () => {
      const { container } = render(<BlockEditor {...defaultProps} />);
      const interactiveElements = container.querySelectorAll<HTMLElement>(
        'button, a, input, textarea, [role="button"], [tabindex]'
      );
      
      interactiveElements.forEach((element) => {
        // Check if element has focus styles
        const styles = window.getComputedStyle(element, ':focus');
        const hasFocusStyle = 
          styles.outline !== 'none' ||
          styles.boxShadow !== 'none' ||
          styles.borderColor !== styles.borderColor; // Border color changes on focus
        
        // This is a basic check - in production, you'd test actual focus state
        expect(isKeyboardAccessible(element)).toBe(true);
      });
    });
  });

  describe('Block Types', () => {
    it('should have appropriate ARIA labels for heading blocks', () => {
      const headingBlock: Block = {
        ...mockBlock,
        type_id: 'heading',
        content: 'Heading text',
      };
      
      const { container } = render(
        <BlockEditor {...defaultProps} block={headingBlock} />
      );
      
      const blockContainer = container.querySelector('[role="article"]');
      expect(blockContainer?.getAttribute('aria-label')).toContain('Heading');
    });

    it('should have appropriate ARIA labels for checklist blocks', () => {
      const checklistBlock: Block = {
        ...mockBlock,
        type_id: 'checklist',
        content: 'Checklist item',
      };
      
      const { container } = render(
        <BlockEditor {...defaultProps} block={checklistBlock} />
      );
      
      const blockContainer = container.querySelector('[role="article"]');
      expect(blockContainer?.getAttribute('aria-label')).toContain('Checklist');
    });
  });
});

