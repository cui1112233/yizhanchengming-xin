import os
import subprocess
import time
import unittest
import urllib.request

from playwright.sync_api import sync_playwright

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), '../../../..'))
URL = 'http://127.0.0.1:4186/src/user/components/director-desk/DirectorInspector.browser.html'


class InspectorBrowserTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.server = subprocess.Popen(['node', 'node_modules/vite/bin/vite.js', '--host', '127.0.0.1', '--port', '4186', '--strictPort'], cwd=ROOT, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
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

    def test_edit_and_invalid_restore_with_theme_contrast(self):
        page = self.browser.new_page(viewport={'width': 1440, 'height': 800})
        self.addCleanup(page.close)
        page.goto(URL)
        x = page.get_by_label('角色位置 X')
        y = page.get_by_label('角色位置 Y')
        x.fill('2.5')
        y.focus()
        self.assertEqual(x.input_value(), '2.5')
        self.assertEqual(y.input_value(), '2')
        x.fill('3.5')
        x.press('Enter')
        self.assertEqual(x.input_value(), '3.5')
        self.assertFalse(x.evaluate('(e) => document.activeElement === e'))
        x.fill('invalid')
        x.press('Tab')
        self.assertEqual(x.input_value(), '3.5')
        self.assertEqual(x.get_attribute('aria-invalid'), 'true')
        alert = page.get_by_role('alert')
        self.assertEqual(alert.inner_text(), '请输入有效数值')
        for theme in ('light', 'dark'):
            page.evaluate('(theme) => document.documentElement.dataset.theme = theme', theme)
            ratio = alert.evaluate('''e => {
              const parse = c => c.match(/\\d+/g).slice(0, 3).map(Number).map(v => { v /= 255; return v <= .04045 ? v / 12.92 : ((v + .055) / 1.055) ** 2.4; });
              const lum = c => { const a = parse(c); return .2126*a[0] + .7152*a[1] + .0722*a[2]; };
              const fg = lum(getComputedStyle(e).color), bg = lum(getComputedStyle(e.closest('.agent-director-shell__right')).backgroundColor);
              return (Math.max(fg, bg) + .05) / (Math.min(fg, bg) + .05);
            }''')
            self.assertGreaterEqual(ratio, 4.5, f'{theme} contrast {ratio}')
        self.assertTrue(page.get_by_role('heading', name='角色素材').is_visible())


if __name__ == '__main__':
    unittest.main()
