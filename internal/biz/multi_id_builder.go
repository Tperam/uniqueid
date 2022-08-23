package biz

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"sync"
)

type IDBuilderBizs struct {
	builderMap map[string]*IDBuilderBiz
	rwLock     sync.RWMutex
	logger     zerolog.Logger
	db         *gorm.DB
}

func NewIDBuilderBizs(logger zerolog.Logger, db *gorm.DB) *IDBuilderBizs {
	builderMap := make(map[string]*IDBuilderBiz, 1)
	return &IDBuilderBizs{builderMap: builderMap, logger: logger, db: db}
}

func (m *IDBuilderBizs) GetID(bizTag string) uint64 {
	m.rwLock.RLock()
	v, ok := m.builderMap[bizTag]
	m.rwLock.RUnlock()

	if ok {

		return v.GetID()
	}

	v = m.addNewBuilder(bizTag)
	return v.GetID()
}

func (m *IDBuilderBizs) addNewBuilder(bizTag string) *IDBuilderBiz {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	v, ok := m.builderMap[bizTag]
	if !ok {
		v = NewIDBuidlerBiz(m.logger, m.db, bizTag)
		m.builderMap[bizTag] = v
	}
	return v
}
