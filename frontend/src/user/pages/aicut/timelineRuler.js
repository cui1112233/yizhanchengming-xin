import { timelinePixelsToTime, timelineTimeToPixels } from '../aicutTimeline.js';

export function timelineRulerTicks(width, zoom = 1) {
  const endMs = timelinePixelsToTime(width, zoom);
  const minimumStep = Math.max(timelinePixelsToTime(90, zoom), endMs / 1000);
  const intervals = [250, 500, 1000, 2000, 5000, 10000, 15000, 30000, 60000, 120000, 300000, 600000, 1800000, 3600000];
  const step = intervals.find(value => value >= minimumStep) || Math.ceil(minimumStep / 3600000) * 3600000;
  const ticks = [];
  for (let timeMs = 0; timeMs <= endMs; timeMs += step) {
    const seconds = Math.floor(timeMs / 1000);
    const label = `${String(Math.floor(seconds / 60)).padStart(2, '0')}:${String(seconds % 60).padStart(2, '0')}${timeMs % 1000 ? `.${String(timeMs % 1000).padStart(3, '0')}` : ''}`;
    ticks.push({ timeMs, left: timelineTimeToPixels(timeMs, zoom), label });
  }
  return ticks;
}
