export const cinematicWallpapers = Object.freeze([
  { id: 'cinema-0', label: '映窗·晚霞', src: '/assets/cinematic/scene-0.mp4', poster: '/assets/cinematic/poster-0.jpg' },
  { id: 'cinema-1', label: '映窗·山行', src: '/assets/cinematic/scene-1.mp4', poster: '/assets/cinematic/poster-1.jpg' },
  { id: 'cinema-2', label: '映窗·雪境', src: '/assets/cinematic/scene-2.mp4', poster: '/assets/cinematic/poster-2.jpg' },
  { id: 'cinema-3', label: '映窗·原野', src: '/assets/cinematic/scene-3.mp4', poster: '/assets/cinematic/poster-3.jpg' }
]);

export function windowVideoDefinition(id, index = 0) {
  if (id === 'cinema-auto') return cinematicWallpapers[((index % 4) + 4) % 4];
  return cinematicWallpapers.find(item => item.id === id) || null;
}
