import io
import json
import threading
import time
import unittest

from worker.ocr_worker import Worker, WorkerError, append_scroll_text


class ResidentWorkerProtocolTest(unittest.TestCase):
    def test_completed_job_returns_idle_and_second_job_reuses_process(self):
        output = io.StringIO()
        calls = []

        def extractor(request, progress, cancel_event):
            calls.append(request["jobId"])
            progress(1, 1, 100)
            return {"text": f"正文-{request['jobId']}", "characters": 4, "frames": 1, "duplicates": 0}

        worker = Worker(output, extractor=extractor)
        worker.handle_command({"type": "ensure_model", "modelDir": "C:/models/v1"})
        worker.handle_command({"type": "extract", "jobId": "one", "videoUrl": "https://material.hnqingyuwen.top/one.mp4", "durationSeconds": 1})
        worker.handle_command({"type": "extract", "jobId": "two", "videoUrl": "https://material.hnqingyuwen.top/two.mp4", "durationSeconds": 1})
        self.assertTrue(worker.wait_until_idle(timeout=2))

        events = [json.loads(line) for line in output.getvalue().splitlines()]
        self.assertEqual(calls, ["one", "two"])
        self.assertEqual([event["type"] for event in events].count("complete"), 2)
        self.assertEqual([event["type"] for event in events].count("idle"), 2)
        self.assertEqual([event["type"] for event in events].count("ready"), 1)

    def test_unsafe_video_url_fails_without_starting_ocr(self):
        output = io.StringIO()
        calls = []

        def extractor(request, progress, cancel_event):
            calls.append(request)
            return {"text": "不应执行"}

        worker = Worker(output, extractor=extractor)
        worker.handle_command({"type": "extract", "jobId": "bad", "videoUrl": "http://example.com/video.mp4", "durationSeconds": 1})
        events = [json.loads(line) for line in output.getvalue().splitlines()]
        self.assertEqual(calls, [])
        self.assertEqual(events[-2]["type"], "failed")
        self.assertEqual(events[-2]["code"], "OCR_VIDEO_NOT_ALLOWED")
        self.assertEqual(events[-1]["type"], "idle")

    def test_allows_only_the_two_known_qingyu_media_hosts(self):
        output = io.StringIO()
        calls = []

        def extractor(request, progress, cancel_event):
            calls.append(request["jobId"])
            return {"text": "正文", "characters": 2, "frames": 1, "duplicates": 0}

        worker = Worker(output, extractor=extractor)
        worker.handle_command({"type": "extract", "jobId": "legacy", "videoUrl": "https://material.hnqingyuwen.top/video.mp4", "durationSeconds": 1})
        worker.handle_command({"type": "extract", "jobId": "mlzr", "videoUrl": "https://mlzr-material.hnqingyuwen.top/video.mp4", "durationSeconds": 1})
        worker.handle_command({"type": "extract", "jobId": "ml", "videoUrl": "https://ml-material.hnqingyuwen.top/video.mp4", "durationSeconds": 1})
        self.assertTrue(worker.wait_until_idle(timeout=2))
        worker.handle_command({"type": "extract", "jobId": "lookalike", "videoUrl": "https://evilmaterial.hnqingyuwen.top/video.mp4", "durationSeconds": 1})

        events = [json.loads(line) for line in output.getvalue().splitlines()]
        self.assertEqual(calls, ["legacy", "mlzr", "ml"])
        self.assertEqual(events[-2]["code"], "OCR_VIDEO_NOT_ALLOWED")

    def test_cancel_emits_one_failure_then_idle(self):
        output = io.StringIO()
        started = threading.Event()

        def extractor(request, progress, cancel_event):
            started.set()
            while not cancel_event.is_set():
                time.sleep(0.005)
            raise WorkerError("OCR_CANCELLED")

        worker = Worker(output, extractor=extractor)
        worker.handle_command({"type": "extract", "jobId": "cancel-me", "videoUrl": "https://material.hnqingyuwen.top/video.mp4", "durationSeconds": 1})
        self.assertTrue(started.wait(1))
        worker.handle_command({"type": "cancel", "jobId": "cancel-me"})
        self.assertTrue(worker.wait_until_idle(timeout=2))
        events = [json.loads(line) for line in output.getvalue().splitlines()]
        failures = [event for event in events if event["type"] == "failed"]
        self.assertEqual(len(failures), 1)
        self.assertEqual(failures[0]["code"], "OCR_CANCELLED")
        self.assertEqual(events[-1]["type"], "idle")

    def test_scroll_merge_preserves_new_text_and_does_not_drop_periods(self):
        overlap = "姥姥怕我长大进去尝试改变我的想法发现无用无奈妥协老话说得好事不过三"
        previous = "我答应了。" + overlap + "。”"
        current = overlap + "。”\n如果有人惹你了至少给对方三次机会"
        merged, skipped = append_scroll_text(previous, previous, current)
        self.assertFalse(skipped)
        self.assertEqual(merged, previous + "\n如果有人惹你了至少给对方三次机会")


if __name__ == "__main__":
    unittest.main()
