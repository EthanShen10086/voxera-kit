package contract

import (
	cachecontract "github.com/EthanShen10086/voxera-kit/cache/contract"
	dbcontract "github.com/EthanShen10086/voxera-kit/database/contract"
	mqcontract "github.com/EthanShen10086/voxera-kit/mq/contract"
	secretcontract "github.com/EthanShen10086/voxera-kit/secret/contract"
	storagecontract "github.com/EthanShen10086/voxera-kit/storage/contract"
	taskcontract "github.com/EthanShen10086/voxera-kit/task/contract"
)

// Re-export contract runners for a single testkit entry point.

type (
	CacheFactory    = cachecontract.Factory
	DatabaseFactory = dbcontract.Factory
	MQFactory       = mqcontract.Factory
	SecretFactory   = secretcontract.Factory
	TaskFactory     = taskcontract.Factory
)

var (
	RunCacheContract       = cachecontract.RunCacheContract
	RunDatabaseContract    = dbcontract.RunDatabaseContract
	RunMQContract          = mqcontract.RunMQContract
	RunObjectStoreContract = storagecontract.RunObjectStoreContract
	RunMultipartContract   = storagecontract.RunMultipartContract
	RunVersioningContract  = storagecontract.RunVersioningContract
	RunSecretContract      = secretcontract.RunSecretContract
	RunTaskContract        = taskcontract.RunTaskContract
)
