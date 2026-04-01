package notification

import (
	"context"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/engine/notification/provider"
)

type Engine struct {
	db *sqlc.Queries
	producer Producer
	providers map[string]provider.Provider 
}

type Producer interface{
	Publish(ctx context.Context, topic string,msg []byte)error
}

func NewEngine(db *sqlc.Queries,producer Producer)*Engine{
	e := &Engine{
        db:        db,
        producer:  producer,
        providers: make(map[string]provider.Provider), 
    }
	e.providers["email"]=provider.NewmockProvider("email")
	e.providers["sms"] = provider.NewmockProvider("sms")

	return e
}