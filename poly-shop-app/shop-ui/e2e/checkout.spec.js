// @ts-check
const { test, expect } = require('@playwright/test');

test('has title', async ({ page }) => {
    await page.goto('http://localhost:8070');
    await expect(page).toHaveTitle(/PolyShop/);
});

test('checkout flow', async ({ page }) => {
    await page.goto('http://localhost:8070');

    // Wait for products to load
    // Assuming products are loaded and displayed as .group or similar selector from ProductGrid
    const firstProductBtn = page.locator('button[aria-label="Add to cart"]').first();
    await firstProductBtn.waitFor({ state: 'visible', timeout: 10000 });

    await firstProductBtn.click();

    // Open cart
    await page.click('button[aria-label="Open cart"]');

    // Check cart contents
    await expect(page.locator('text=Total')).toBeVisible();

    // Checkout
    page.on('dialog', dialog => dialog.accept());
    await page.click('button:has-text("Checkout")');

    // Verify empty cart message
    await expect(page.locator('text=Your cart is empty')).toBeVisible();
});
