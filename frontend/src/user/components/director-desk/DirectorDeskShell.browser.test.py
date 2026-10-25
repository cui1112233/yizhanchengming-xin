import os
import subprocess
import time
import unittest
import urllib.request

from playwright.sync_api import sync_playwright

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), '../../../..'))
URL = 'http://127.0.0.1:4179/src/user/components/director-desk/DirectorDeskShell.browser.html'


class ShellBrowserTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.server = subprocess.Popen(['node', 'node_modules/vite/bin/vite.js', '--host', '127.0.0.1', '--port', '4179', '--strictPort'], cwd=ROOT, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        for _ in range(100):
            if cls.server.poll() is not None:
                raise RuntimeError('fixture server stopped')
            try:
                urllib.request.urlopen(URL, timeout=1).close()
                break
            except OSError:
                time.sleep(.1)
        else:
            cls.server.terminate()
            raise RuntimeError('fixture server did not start')
        cls.playwright = sync_playwright().start()
        cls.browser = cls.playwright.chromium.launch(headless=True, executable_path=os.environ.get('DIRECTOR_CHROMIUM', r'C:\Program Files\Google\Chrome\Application\chrome.exe'))

    @classmethod
    def tearDownClass(cls):
        cls.browser.close()
        cls.playwright.stop()
        cls.server.terminate()
        cls.server.wait(timeout=10)

    def test_desktop_theme_and_rail(self):
        page = self.browser.new_page(viewport={'width': 1440, 'height': 800})
        self.addCleanup(page.close)
        page.goto(URL)
        page.evaluate("document.documentElement.dataset.theme = 'light'")
        selected = page.locator('.agent-director-scene-tree__select[aria-pressed=true]')
        selected.wait_for()
        self.assertEqual(selected.evaluate('(e) => getComputedStyle(e).color'), 'rgb(24, 91, 184)')
        self.assertNotEqual(selected.evaluate('(e) => getComputedStyle(e).backgroundColor'), 'rgb(29, 42, 51)')
        self.assertNotEqual(page.get_by_role('button', name='右侧操作').evaluate('(e) => getComputedStyle(e).color'), 'rgb(243, 247, 248)')
        self.assertGreater(page.locator('.agent-director-shell').bounding_box()['height'], 700)
        rail = page.locator('.agent-director-tool-rail').bounding_box()
        tree = page.locator('.agent-director-scene-tree').bounding_box()
        self.assertEqual(round(rail['width']), 44)
        self.assertAlmostEqual(rail['x'] + rail['width'], tree['x'], delta=1)
        self.assertEqual(round(tree['width']), 228)
        page.get_by_role('button', name='项目图片', exact=True).click()
        self.assertEqual(page.get_by_role('button', name='项目图片', exact=True).get_attribute('aria-pressed'), 'true')
        camera = page.locator('.agent-director-scene-tree__select').nth(1)
        camera.click()
        self.assertEqual(camera.get_attribute('aria-pressed'), 'true')
        self.assertEqual(selected.count(), 1)
        page.evaluate("document.documentElement.dataset.theme = 'dark'")
        self.assertNotEqual(selected.evaluate('(e) => getComputedStyle(e).color'), 'rgb(24, 91, 184)')
        page.keyboard.press('Tab')
        self.assertNotEqual(page.evaluate('getComputedStyle(document.activeElement).outlineStyle'), 'none')

    def test_narrow_panels_and_viewport_are_accessible(self):
        page = self.browser.new_page(viewport={'width': 390, 'height': 800})
        self.addCleanup(page.close)
        page.goto(URL)
        viewport = page.locator('.agent-director-shell__viewport').bounding_box()
        self.assertGreaterEqual(viewport['width'], 200)
        self.assertEqual(page.locator('.agent-director-shell__left-content').count(), 1)
        self.assertEqual(page.locator('.agent-director-shell__right-content').count(), 0)
        page.get_by_role('button', name='项目图片', exact=True).click()
        page.get_by_role('button', name='展开右侧面板').click()
        self.assertEqual(page.locator('.agent-director-shell__left-content').count(), 0)
        page.get_by_role('button', name='右侧操作').click()
        page.get_by_role('button', name='展开左侧面板').click()
        self.assertEqual(page.locator('.agent-director-shell__right-content').count(), 0)
        page.get_by_role('button', name='项目图片', exact=True).click()
        page.get_by_role('button', name='收起左侧面板').click()
        page.get_by_role('button', name='展开右侧面板').click()
        page.get_by_role('button', name='收起右侧面板').click()

    def test_resize_from_desktop_keeps_only_one_narrow_panel(self):
        page = self.browser.new_page(viewport={'width': 1440, 'height': 800})
        self.addCleanup(page.close)
        page.goto(URL)
        page.locator('.agent-director-shell__right-content').wait_for()
        self.assertEqual(page.locator('.agent-director-shell__left-content').count(), 1)
        page.set_viewport_size({'width': 390, 'height': 800})
        page.get_by_role('button', name='展开右侧面板').wait_for()
        self.assertEqual(page.locator('.agent-director-shell__left-content').count(), 1)
        self.assertEqual(page.locator('.agent-director-shell__right-content').count(), 0)
        page.set_viewport_size({'width': 1440, 'height': 800})
        page.locator('.agent-director-shell__right-content').wait_for()
        self.assertEqual(page.locator('.agent-director-shell__left-content').count(), 1)
        page.set_viewport_size({'width': 390, 'height': 800})
        page.get_by_role('button', name='展开右侧面板').wait_for()
        page.get_by_role('button', name='展开右侧面板').click()
        page.get_by_role('button', name='右侧操作').click()
        self.assertEqual(page.locator('.agent-director-shell__left-content').count(), 0)


if __name__ == '__main__':
    unittest.main()
