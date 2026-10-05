export function validateTimeline(shots, expectedDurationSec, maxShotDurationSec = 15) {
  const epsilon = 0.005
  if (!Array.isArray(shots) || shots.length === 0) throw new Error('timeline must contain at least one shot')
  if (Math.abs(Number(shots[0].start) - 0) > epsilon) throw new Error('timeline must start at 0')

  for (let index = 0; index < shots.length; index += 1) {
    const shot = shots[index]
    const start = Number(shot.start)
    const end = Number(shot.end)
    if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start) throw new Error(`invalid shot ${index}`)
    if (end - start > maxShotDurationSec + epsilon) throw new Error(`shot ${index} exceeds ${maxShotDurationSec}s limit`)
    if (index > 0) {
      const previousEnd = Number(shots[index - 1].end)
      if (Math.abs(start - previousEnd) > epsilon) throw new Error(`gap/overlap before shot ${index}`)
    }
  }

  const finalEnd = Number(shots[shots.length - 1].end)
  if (Math.abs(finalEnd - expectedDurationSec) > epsilon) {
    throw new Error(`timeline ends at ${finalEnd}, expected ${expectedDurationSec}`)
  }

  return true
}
