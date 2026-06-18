// Package contract re-exports data-plane contract runners and integration smoke helpers.
package contract

import (
	cachecontract "github.com/EthanShen10086/voxera-kit/cache/contract"
	dbcontract "github.com/EthanShen10086/voxera-kit/database/contract"
	mqcontract "github.com/EthanShen10086/voxera-kit/mq/contract"
	secretcontract "github.com/EthanShen10086/voxera-kit/secret/contract"
	storagecontract "github.com/EthanShen10086/voxera-kit/storage/contract"
	taskcontract "github.com/EthanShen10086/voxera-kit/task/contract"
)

// Factory type aliases for contract test setup.
type (
	// CacheFactory creates a cache adapter for contract tests.
	CacheFactory = cachecontract.Factory
	// DatabaseFactory creates a database adapter for contract tests.
	DatabaseFactory = dbcontract.Factory
	// MQFactory creates mq publisher/subscriber pairs for contract tests.
	MQFactory = mqcontract.Factory
	// SecretFactory creates a secret manager for contract tests.
	SecretFactory = secretcontract.Factory
	// TaskFactory creates a task queue for contract tests.
	TaskFactory = taskcontract.Factory
)

// Contract runners re-exported from data-plane modules.
var (
	// RunCacheContract exercises cache.Cache implementations.
	RunCacheContract = cachecontract.RunCacheContract
	// RunDatabaseContract exercises database.Database implementations.
	RunDatabaseContract = dbcontract.RunDatabaseContract
	// RunMQContract exercises mq publish/subscribe implementations.
	RunMQContract = mqcontract.RunMQContract
	// RunObjectStoreContract exercises storage.ObjectStore implementations.
	RunObjectStoreContract = storagecontract.RunObjectStoreContract
	// RunMultipartContract exercises multipart upload flows.
	RunMultipartContract = storagecontract.RunMultipartContract
	// RunVersioningContract exercises versioned object store flows.
	RunVersioningContract = storagecontract.RunVersioningContract
	// RunSecretContract exercises secret.Manager implementations.
	RunSecretContract = secretcontract.RunSecretContract
	// RunTaskContract exercises task.TaskQueue implementations.
	RunTaskContract = taskcontract.RunTaskContract
)
