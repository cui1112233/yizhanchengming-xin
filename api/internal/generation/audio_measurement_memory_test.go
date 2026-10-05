package generation

import "context"

var memoryAudioMeasurements = map[*memoryStore][]AudioMeasurement{}

func (s *memoryStore) AudioMeasurementByAsset(_ context.Context, projectID, bookID int64, assetHash string) (AudioMeasurement, error) {
	for _, value := range memoryAudioMeasurements[s] {
		if value.BatchProjectID == projectID && value.BookID == bookID && value.AssetHash == assetHash {
			return value, nil
		}
	}
	return AudioMeasurement{}, ErrNotFound
}

func (s *memoryStore) LatestAudioMeasurement(_ context.Context, projectID, bookID int64) (AudioMeasurement, error) {
	values := memoryAudioMeasurements[s]
	for i := len(values) - 1; i >= 0; i-- {
		if values[i].BatchProjectID == projectID && values[i].BookID == bookID {
			return values[i], nil
		}
	}
	return AudioMeasurement{}, ErrNotFound
}

func (s *memoryStore) CreateAudioMeasurement(_ context.Context, value AudioMeasurement) (AudioMeasurement, error) {
	values := memoryAudioMeasurements[s]
	value.ID = int64(len(values) + 1)
	value.CreatedAt = s.now
	memoryAudioMeasurements[s] = append(values, value)
	return value, nil
}
