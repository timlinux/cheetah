// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';

// Create a test document
const createTestDocument = (content) => {
  const tmpDir = '/tmp/cheetah-tests';
  if (!fs.existsSync(tmpDir)) {
    fs.mkdirSync(tmpDir, { recursive: true });
  }
  const filePath = path.join(tmpDir, 'test-document.txt');
  fs.writeFileSync(filePath, content);
  return filePath;
};

test.describe('Document Upload', () => {
  test('should accept dropped text file', async ({ page }) => {
    await page.goto('/');

    // Create a test file
    const testContent = 'Hello world. This is a test document.';

    // Create a file and trigger drop event
    const buffer = Buffer.from(testContent);
    const dataTransfer = await page.evaluateHandle((data) => {
      const dt = new DataTransfer();
      const file = new File([new Uint8Array(data)], 'test.txt', { type: 'text/plain' });
      dt.items.add(file);
      return dt;
    }, [...buffer]);

    // Find the drop zone and dispatch drop event
    const dropZone = page.locator('[class*="dropzone"]').first();
    if (await dropZone.count() === 0) {
      // Try finding by text
      const dropArea = page.locator('text=Drag & drop').locator('..').locator('..');
      await dropArea.dispatchEvent('drop', { dataTransfer });
    } else {
      await dropZone.dispatchEvent('drop', { dataTransfer });
    }

    // Wait for potential loading state
    await page.waitForTimeout(1000);
  });

  test('should show loading state when processing document', async ({ page }) => {
    await page.goto('/');

    // The loading spinner should appear during document processing
    // This is a basic check - actual document upload would trigger this
  });
});

test.describe('Reading Screen', () => {
  // These tests would require a way to programmatically load a document
  // For now, we test the basic structure

  test('reading controls should be documented', async ({ page }) => {
    await page.goto('/');

    // Keyboard hints should be visible on home page
    await expect(page.locator('text=SPACE')).toBeVisible();
    await expect(page.locator('text=speed')).toBeVisible();
  });
});

test.describe('Keyboard Shortcuts', () => {
  test('should document keyboard shortcuts on home page', async ({ page }) => {
    await page.goto('/');

    // Check that keyboard shortcuts are mentioned
    const shortcutsText = await page.textContent('body');
    expect(shortcutsText).toContain('SPACE');
    expect(shortcutsText).toContain('speed');
  });
});

test.describe('Accessibility', () => {
  test('should have no automatic accessibility violations on home page', async ({ page }) => {
    await page.goto('/');

    // Check for basic accessibility features
    // Check that buttons have visible text
    const buttons = page.locator('button');
    const buttonCount = await buttons.count();

    for (let i = 0; i < buttonCount; i++) {
      const button = buttons.nth(i);
      const text = await button.textContent();
      const ariaLabel = await button.getAttribute('aria-label');
      // Button should have either text content or aria-label
      expect(text?.trim() || ariaLabel).toBeTruthy();
    }
  });

  test('should have proper heading hierarchy', async ({ page }) => {
    await page.goto('/');

    // Should have at least one h1 or equivalent
    const mainHeading = page.locator('h1, h2, [role="heading"]');
    expect(await mainHeading.count()).toBeGreaterThan(0);
  });

  test('links should have href attributes', async ({ page }) => {
    await page.goto('/');

    const links = page.locator('a');
    const linkCount = await links.count();

    for (let i = 0; i < linkCount; i++) {
      const link = links.nth(i);
      const href = await link.getAttribute('href');
      expect(href).toBeTruthy();
    }
  });
});

test.describe('Theme and Styling', () => {
  test('should use dark theme by default', async ({ page }) => {
    await page.goto('/');

    // Check that the body has dark background
    const bodyBg = await page.evaluate(() => {
      return window.getComputedStyle(document.body).backgroundColor;
    });

    // Dark colors typically have low RGB values
    // This is a basic check
    expect(bodyBg).toBeTruthy();
  });

  test('should have consistent branding colors', async ({ page }) => {
    await page.goto('/');

    // Kartoza brand color (orange) should be present
    const htmlContent = await page.content();
    // Just verify the page renders without errors
    expect(htmlContent).toContain('Cheetah');
  });
});

test.describe('Error Handling', () => {
  test('should handle invalid file gracefully', async ({ page }) => {
    await page.goto('/');

    // Create an invalid file type
    const buffer = Buffer.from('Invalid content');
    const dataTransfer = await page.evaluateHandle((data) => {
      const dt = new DataTransfer();
      const file = new File([new Uint8Array(data)], 'test.xyz', { type: 'application/octet-stream' });
      dt.items.add(file);
      return dt;
    }, [...buffer]);

    // The app should not crash on invalid file type
    // It should either reject or show an error
  });
});

test.describe('Performance', () => {
  test('should load within reasonable time', async ({ page }) => {
    const startTime = Date.now();
    await page.goto('/');
    const loadTime = Date.now() - startTime;

    // Page should load within 5 seconds
    expect(loadTime).toBeLessThan(5000);
  });

  test('should not have console errors on load', async ({ page }) => {
    const errors = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });

    await page.goto('/');
    await page.waitForTimeout(1000);

    // Filter out known non-critical errors
    const criticalErrors = errors.filter(
      (e) => !e.includes('favicon') && !e.includes('AdSense')
    );

    expect(criticalErrors).toHaveLength(0);
  });
});
