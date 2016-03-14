package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/thrawn01/canis"
	"github.com/thrawn01/canis/middleware/logger"
	"github.com/thrawn01/canis/middleware/openstack"
	"github.com/thrawn01/canis/middleware/request"
	"github.com/thrawn01/canis/middleware/throttle"
)

func main() {
	router := canis.Router()

	db, err := rethink.Connect(rethink.ConnectOpts{
		Addresses: conf.RdbAddresses,
		AuthKey:   conf.RdbAuthKey,
		Database:  DatabaseName,
	})
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	router.GET("/pies/:id", func(ctx canis.Context, resp http.ResponseWriter, req *http.Request) {
		// Fetch our pie by ID from the database
		cursor, err := rethink.Table("pies").Get(ctx.ByName("id")).Run(db)
		if err != nil {
			log.Error(err)
			canis.Abort(resp, http.StatusText(500), 500)
		}

		var pie interface{}
		err = cursor.All(&pie)
		if err := cursor.One(&pie); err != nil {
			canis.Abort(resp, http.StatusText(401), 401)
		}
		canis.ToJson(resp, pie)
	})

	chain := canis.Chain(
		logger.Access(logger.ToFile("/var/log/pie/access.log")),
		logger.Error(logger.ToFile("/var/log/pie/error.log")),
		request.CatchPanic(),
		request.Timeout(30),
		throttle.Throttler(
			throttle.VaryByIpAddress(),
			throttle.VaryByPath(),
			throttle.PerMin(100),
			throttle.Burst(50),
			throttle.MemStore(1000),
		),
		cors.Cors(
			cors.Origin("api.rackspace.com"),
			cors.Allow("friend.rackspace.com"),
		),
	)

	// Middleware that accepts a context
	chain.Add(openstack.KeystoneAuth(
		openstack.AuthUrl("http://identity.rackspace.com")
	)

	// Middlware the does not accept a context
	chain.Use(
		noncontext.Middlware()
	)

	chain.Then(server)
}
