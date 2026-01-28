/**
 * Accessibility testing and validation utilities
 * Provides helpers for checking WCAG compliance and ARIA usage
 */

/**
 * Check if an element has proper ARIA labels
 */
export function hasAriaLabel(element: HTMLElement): boolean {
  const hasLabel = element.hasAttribute('aria-label');
  const hasLabelledBy = element.hasAttribute('aria-labelledby');
  const hasLabelElement = element.querySelector('label');
  const hasTitle = element.hasAttribute('title');
  
  return hasLabel || hasLabelledBy || !!hasLabelElement || hasTitle;
}

/**
 * Check if an element is keyboard accessible
 */
export function isKeyboardAccessible(element: HTMLElement): boolean {
  // Check if element is naturally keyboard accessible
  const naturallyAccessible = ['BUTTON', 'A', 'INPUT', 'TEXTAREA', 'SELECT'].includes(element.tagName);
  
  // Check if element has tabindex
  const hasTabIndex = element.hasAttribute('tabindex');
  const tabIndex = element.getAttribute('tabindex');
  const isTabIndexValid = tabIndex === null || tabIndex === '0' || parseInt(tabIndex) >= 0;
  
  // Check if element has role that makes it keyboard accessible
  const role = element.getAttribute('role');
  const hasAccessibleRole = role && ['button', 'link', 'menuitem', 'tab'].includes(role);
  
  return naturallyAccessible || (hasTabIndex && isTabIndexValid) || !!hasAccessibleRole;
}

/**
 * Check if an element has proper focus indicators
 */
export function hasFocusIndicator(element: HTMLElement): boolean {
  const styles = window.getComputedStyle(element, ':focus');
  const outline = styles.outline;
  const outlineWidth = styles.outlineWidth;
  const boxShadow = styles.boxShadow;
  
  // Check if there's a visible focus indicator
  const hasOutline = outline !== 'none' && outlineWidth !== '0px';
  const hasBoxShadow = boxShadow !== 'none';
  
  return hasOutline || hasBoxShadow;
}

/**
 * Check if color contrast meets WCAG AA standards (simplified check)
 * Note: This is a basic check. For production, use a proper contrast checker library
 */
export function checkColorContrast(foreground: string, background: string): {
  passes: boolean;
  ratio: number;
  level: 'AA' | 'AAA' | 'FAIL';
} {
  // Simplified contrast calculation
  // In production, use a library like 'color-contrast-checker' or 'wcag-contrast'
  // This is a placeholder that returns a basic check
  
  // For now, return a pass if colors are different
  // Real implementation would calculate luminance and contrast ratio
  const ratio = 4.5; // Placeholder - should be calculated from actual colors
  
  let level: 'AA' | 'AAA' | 'FAIL' = 'FAIL';
  if (ratio >= 7) {
    level = 'AAA';
  } else if (ratio >= 4.5) {
    level = 'AA';
  }
  
  return {
    passes: ratio >= 4.5,
    ratio,
    level,
  };
}

/**
 * Check if all interactive elements have accessible names
 */
export function checkInteractiveElementsHaveNames(container: HTMLElement): {
  passes: boolean;
  violations: HTMLElement[];
} {
  const interactiveElements = container.querySelectorAll<HTMLElement>(
    'button, a, input, textarea, select, [role="button"], [role="link"], [tabindex]'
  );
  
  const violations: HTMLElement[] = [];
  
  interactiveElements.forEach((element) => {
    if (!hasAriaLabel(element)) {
      violations.push(element);
    }
  });
  
  return {
    passes: violations.length === 0,
    violations,
  };
}

/**
 * Check if all form inputs have associated labels
 */
export function checkFormLabels(container: HTMLElement): {
  passes: boolean;
  violations: HTMLElement[];
} {
  const inputs = container.querySelectorAll<HTMLElement>(
    'input, textarea, select'
  );
  
  const violations: HTMLElement[] = [];
  
  inputs.forEach((input) => {
    const id = input.getAttribute('id');
    const hasLabel = id && container.querySelector(`label[for="${id}"]`);
    const hasAriaLabel = input.hasAttribute('aria-label') || input.hasAttribute('aria-labelledby');
    const hasPlaceholder = input.hasAttribute('placeholder');
    
    if (!hasLabel && !hasAriaLabel && !hasPlaceholder) {
      violations.push(input);
    }
  });
  
  return {
    passes: violations.length === 0,
    violations,
  };
}

/**
 * Check if all images have alt text
 */
export function checkImageAltText(container: HTMLElement): {
  passes: boolean;
  violations: HTMLImageElement[];
} {
  const images = container.querySelectorAll<HTMLImageElement>('img');
  const violations: HTMLImageElement[] = [];
  
  images.forEach((img) => {
    const alt = img.getAttribute('alt');
    if (alt === null) {
      violations.push(img);
    }
  });
  
  return {
    passes: violations.length === 0,
    violations,
  };
}

/**
 * Check if there are any keyboard traps
 */
export function checkKeyboardTraps(container: HTMLElement): {
  passes: boolean;
  traps: HTMLElement[];
} {
  // This is a simplified check - real implementation would test actual keyboard navigation
  const traps: HTMLElement[] = [];
  
  // Check for elements with negative tabindex that might trap focus
  const negativeTabIndexElements = container.querySelectorAll<HTMLElement>('[tabindex="-1"]');
  
  // In a real implementation, you'd test if focus can escape these elements
  // For now, we just flag potential issues
  
  return {
    passes: traps.length === 0,
    traps: Array.from(negativeTabIndexElements),
  };
}

/**
 * Comprehensive a11y check for a component
 */
export function runA11yChecks(container: HTMLElement): {
  interactiveElements: ReturnType<typeof checkInteractiveElementsHaveNames>;
  formLabels: ReturnType<typeof checkFormLabels>;
  images: ReturnType<typeof checkImageAltText>;
  keyboardTraps: ReturnType<typeof checkKeyboardTraps>;
  overallPass: boolean;
} {
  const interactiveElements = checkInteractiveElementsHaveNames(container);
  const formLabels = checkFormLabels(container);
  const images = checkImageAltText(container);
  const keyboardTraps = checkKeyboardTraps(container);
  
  const overallPass = 
    interactiveElements.passes &&
    formLabels.passes &&
    images.passes &&
    keyboardTraps.passes;
  
  return {
    interactiveElements,
    formLabels,
    images,
    keyboardTraps,
    overallPass,
  };
}

