// Package appctx — app context inject vào transport/biz (mẫu learn_go).
package appctx

import (
	"household-finance/api/component/pubsub"
	"household-finance/api/component/wshub"

	"gorm.io/gorm"
)

type AppContext interface {
	GetDB() *gorm.DB
	GetSecret() string
	GetPubSub() pubsub.PubSub
	GetWSHub() *wshub.Hub
}

type appCtx struct {
	db     *gorm.DB
	secret string
	ps     pubsub.PubSub
	hub    *wshub.Hub
}

func New(db *gorm.DB, secret string, ps pubsub.PubSub, hub *wshub.Hub) AppContext {
	return &appCtx{db: db, secret: secret, ps: ps, hub: hub}
}

func (a *appCtx) GetDB() *gorm.DB          { return a.db }
func (a *appCtx) GetSecret() string        { return a.secret }
func (a *appCtx) GetPubSub() pubsub.PubSub { return a.ps }
func (a *appCtx) GetWSHub() *wshub.Hub     { return a.hub }
