package storage

import "qiantie/backend/internal/imagecapacity"

func ImageCapacityMigrations() []Migration {
	return []Migration{{Version: 8800310, SQL: imagecapacity.SchemaStatements(), CallbackChecksum: "v88-shared-image-capacity-v1"}}
}
