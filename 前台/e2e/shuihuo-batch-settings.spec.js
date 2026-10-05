import { expect, test } from './support/acceptance.js'
import { projectFixture, routeAuthenticated, routeBatchProjectFixtures } from './support/fixtures.js'

test.describe('@batch shuihuo / Batch Factory / settings acceptance', () => {
  test.beforeEach(async ({ page }) => {
    await routeAuthenticated(page)
    await routeBatchProjectFixtures(page)
  })

  test('@batch /shuihuo-production renders without black screen or unsafe empty values', async ({ page }) => {
    await page.goto('/shuihuo-production')

    await expect(page.getByRole('heading', { name: '小说获取工作台' })).toBeVisible()
    await expect(page.getByText('书籍结果')).toBeVisible()
    const body = await page.locator('body').innerText()
    expect(body).not.toContain('[object Object]')
    expect(body).not.toMatch(/\bundefined\b/)
    expect(body).not.toMatch(/\bnull\b/)
  })

  test('@batch BatchProject list shows project metadata and preserves project query context', async ({ page }) => {
    await page.goto('/batch-factory?project=123')

    await expect(page).toHaveURL(/\/batch-factory\?project=123$/)
    await expect(page.getByRole('heading', { name: '批量工厂' })).toBeVisible()
    await expect(page.getByRole('button', { name: projectFixture.name })).toBeVisible()
    await expect(page.getByText('知乎')).toBeVisible()
    await expect(page.getByText('常读')).toBeVisible()
    await expect(page.getByText('女频')).toBeVisible()
    await expect(page.getByText('男频')).toBeVisible()
    await expect(page.getByText('现实情感')).toBeVisible()
    await expect(page.getByText('都市')).toBeVisible()
    await expect(page.getByText('待执行')).toBeVisible()
  })

  test('@batch project detail shows Book ID/title/store/gender/style/status/error with safe fallbacks', async ({ page }) => {
    await page.goto('/batch-factory')
    await page.getByRole('button', { name: projectFixture.name }).click()

    await expect(page.getByRole('heading', { name: projectFixture.name })).toBeVisible()
    await expect(page.getByText('1001')).toBeVisible()
    await expect(page.getByText('林子深处有声音')).toBeVisible()
    await expect(page.getByText('知乎')).toBeVisible()
    await expect(page.getByText('女频')).toBeVisible()
    await expect(page.getByText('现实情感')).toBeVisible()
    await expect(page.getByText('已获取')).toBeVisible()
    await expect(page.getByText('fixture upstream failure')).toBeVisible()
    const body = await page.locator('body').innerText()
    expect(body).not.toContain('[object Object]')
    expect(body).not.toMatch(/\bundefined\b|\bnull\b/)
  })

  test('@batch production settings opens a Drawer, saves, and keeps current URL', async ({ page }) => {
    await page.goto('/batch-factory?project=123')
    const before = page.url()

    await page.getByRole('button', { name: '生产统一设置' }).click()
    await expect(page.locator('.ant-drawer-title').filter({ hasText: '生产统一设置' })).toBeVisible()
    await expect(page.getByText(/应用于当前 BatchProject/)).toBeVisible()
    await page.getByRole('button', { name: '保存生产统一设置' }).click()
    await expect(page.getByText('生产统一设置已保存')).toBeVisible()
    expect(page.url()).toBe(before)
  })

  test('@batch publishing settings opens a Drawer, saves, and keeps current URL', async ({ page }) => {
    await page.goto('/batch-factory?project=123')
    const before = page.url()

    await page.getByRole('button', { name: '发布统一设置' }).click()
    await expect(page.locator('.ant-drawer-title').filter({ hasText: '发布统一设置' })).toBeVisible()
    await page.getByRole('button', { name: '保存发布统一设置' }).click()
    await expect(page.getByText('发布统一设置已保存')).toBeVisible()
    expect(page.url()).toBe(before)
  })

  test('@batch version profile exists, sync actions are available, and legacy 解析输入 is not actionable', async ({ page }) => {
    await page.goto('/batch-factory?project=123')
    const before = page.url()

    await expect(page.getByRole('button', { name: '解析输入' })).toHaveCount(0)
    await page.getByRole('button', { name: '版本对应配置档' }).click()
    await expect(page.locator('.ant-drawer-title').filter({ hasText: '版本对应配置档' })).toBeVisible()
    await page.getByRole('tab', { name: '同步' }).click()
    await expect(page.getByRole('button', { name: '同步 121 网站配置' })).toBeVisible()
    await expect(page.getByRole('button', { name: '同步批量风格类型' })).toBeVisible()
    await page.getByRole('tab', { name: '配置档' }).click()
    await page.getByRole('button', { name: '保存配置档' }).click()
    await expect(page.getByText('版本对应配置档已保存')).toBeVisible()
    expect(page.url()).toBe(before)
    await expect(page.getByRole('button', { name: '解析输入' })).toHaveCount(0)
  })
})
