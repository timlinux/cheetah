// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

import { test, expect } from '@playwright/test';

test.describe('Home Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display the Cheetah title', async ({ page }) => {
    await expect(page.locator('text=Cheetah')).toBeVisible();
  });

  test('should display RSVP Speed Reading subtitle', async ({ page }) => {
    await expect(page.locator('text=RSVP Speed Reading')).toBeVisible();
  });

  test('should display the main heading', async ({ page }) => {
    await expect(page.locator('text=Read at 1000+ WPM')).toBeVisible();
  });

  test('should display the drop zone', async ({ page }) => {
    await expect(page.locator('text=Drag & drop a document')).toBeVisible();
  });

  test('should display supported formats', async ({ page }) => {
    await expect(page.locator('text=PDF')).toBeVisible();
    await expect(page.locator('text=DOCX')).toBeVisible();
    await expect(page.locator('text=EPUB')).toBeVisible();
    await expect(page.locator('text=TXT')).toBeVisible();
  });

  test('should display Kartoza branding in footer', async ({ page }) => {
    await expect(page.locator('text=Kartoza')).toBeVisible();
  });

  test('should display donate link in footer', async ({ page }) => {
    await expect(page.locator('text=Donate!')).toBeVisible();
  });

  test('should display GitHub link in footer', async ({ page }) => {
    await expect(page.locator('a:has-text("GitHub")')).toBeVisible();
  });

  test('should display feature cards', async ({ page }) => {
    await expect(page.locator('text=Lightning Fast')).toBeVisible();
    await expect(page.locator('text=Privacy First')).toBeVisible();
    await expect(page.locator('text=Auto-Save')).toBeVisible();
  });

  test('drop zone should be clickable', async ({ page }) => {
    const dropZone = page.locator('text=Drag & drop a document').locator('..');
    await expect(dropZone).toBeVisible();
  });
});

test.describe('Navigation', () => {
  test('Kartoza link should open kartoza.com', async ({ page }) => {
    await page.goto('/');
    const kartozaLink = page.locator('a:has-text("Kartoza")');
    await expect(kartozaLink).toHaveAttribute('href', 'https://kartoza.com');
  });

  test('GitHub link should open correct repository', async ({ page }) => {
    await page.goto('/');
    const githubLink = page.locator('a:has-text("GitHub")');
    await expect(githubLink).toHaveAttribute('href', /github\.com.*cheetah/);
  });
});

test.describe('Responsive Design', () => {
  test('should work on mobile viewport', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/');
    await expect(page.locator('text=Cheetah')).toBeVisible();
    await expect(page.locator('text=Drag & drop')).toBeVisible();
  });

  test('should work on tablet viewport', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.goto('/');
    await expect(page.locator('text=Cheetah')).toBeVisible();
  });

  test('should work on desktop viewport', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.goto('/');
    await expect(page.locator('text=Cheetah')).toBeVisible();
  });
});
