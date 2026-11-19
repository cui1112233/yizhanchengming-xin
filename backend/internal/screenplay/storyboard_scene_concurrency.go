package screenplay

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"qiantie/backend/internal/prompt"
)

// This bounds scene requests in this process, including separate owners/projects.
// It is not a distributed queue or a measured provider/channel capacity.
var storyboardSceneSlots = make(chan struct{}, sceneConcurrencySetting("QIANTIE_SCREENPLAY_SCENE_GLOBAL_LIMIT", 4, 32))

func sceneConcurrencySetting(name string, fallback, max int) int {
	n, err := strconv.Atoi(os.Getenv(name))
	if err != nil || n < 1 {
		n = fallback
	}
	if n > max {
		n = max
	}
	return n
}

func (s *StoryboardService) sceneConcurrency() int {
	n := s.SceneConcurrency
	if n < 1 {
		n = sceneConcurrencySetting("QIANTIE_SCREENPLAY_SCENE_CONCURRENCY", 2, cap(storyboardSceneSlots))
	}
	if n > cap(storyboardSceneSlots) {
		n = cap(storyboardSceneSlots)
	}
	return n
}

func (s *StoryboardService) generateConcurrentScenes(ctx context.Context, req StoryboardRequest, episode episodeInput, units []storyboardSceneInput, selected prompt.CreativePromptSnapshot, roles, scenes, props, style, synopsis, prefix, extra, planning, model, aspectRatio string) ([]StoryboardShot, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	type sceneResult struct {
		shots []StoryboardShot
		err   error
	}
	results := make([]sceneResult, len(units))
	jobs := make(chan int, len(units))
	for index := range units {
		jobs <- index
	}
	close(jobs)
	var wg sync.WaitGroup
	workers := s.sceneConcurrency()
	if workers > len(units) {
		workers = len(units)
	}
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				if ctx.Err() != nil {
					return
				}
				select {
				case storyboardSceneSlots <- struct{}{}:
				case <-ctx.Done():
					return
				}
				local := req
				local.sceneOnlyIndex = &index
				shots, err := s.generateSceneShots(ctx, local, episode, units, selected, roles, scenes, props, style, synopsis, prefix, extra, planning, model, false, aspectRatio)
				<-storyboardSceneSlots
				results[index] = sceneResult{shots, err}
				// An uncertain request must be reconciled before scheduling more paid work.
				if err != nil {
					status := req.generationSession.callStatus(episode.No, fmt.Sprintf("storyboard_scene_%d", index+1))
					if status == "unknown" || status == "calling" {
						cancel()
						return
					}
				}
			}
		}()
	}
	wg.Wait()
	var all []StoryboardShot
	var firstErr error
	for _, result := range results {
		if result.err != nil && (firstErr == nil || (errors.Is(firstErr, context.Canceled) && !errors.Is(result.err, context.Canceled))) {
			firstErr = result.err
		}
		for _, shot := range result.shots {
			shot.Seq = len(all) + 1
			all = append(all, shot)
		}
	}
	if firstErr == nil && ctx.Err() != nil {
		firstErr = ctx.Err()
	}
	return all, firstErr
}
