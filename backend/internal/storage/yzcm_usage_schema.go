package storage

import "qiantie/backend/internal/yzcmusage"

func YZCMUsageMigrations() []Migration {
	return []Migration{{Version: 8800402, SQL: yzcmusage.SchemaStatements(), CallbackChecksum: "yzcm-member-request-billing-ledger-v1"}}
}
