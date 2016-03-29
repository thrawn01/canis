package examples

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/thrawn01/canis"
	"github.com/thrawn01/canis/request/openstack"
	"github.com/thrawn01/canis/request/request"
	"github.com/thrawn01/canis/request/throttle"
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

	common := canis.Chain(
		request.ErrorLogger(log),
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

	// Extend the common chain with authentication, returning a new chain
	requireAuth := canis.Chain(
		openstack.AuthUrl("http://identity.rackspace.com"),
	)

	v1API := canis.Router()

	// Matches /v1/pies/:id
	v1API.GET("/:id", func(ctx canis.Context, resp http.ResponseWriter, req *http.Request) {
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

	v1 := canis.Router()
	// Matches GET /v1
	v1.GET("/", func(ctx canis.Context, resp http.ResponseWriter, req *http.Request) {
		// TODO: Serve up some docs
	})
	// POST to /pies require auth
	v1.POST("/pies", requireAuth.Then(v1API))
	// GET to /pies do not require auth
	v1.GET("/pies", v1API)

	// Handle any /v1 requests with common middleware
	router.Handle("", "/v1", common.Then(v1))

	v2API := canis.Router()

	// Matches /v2
	v1.GET("/", func(ctx canis.Context, resp http.ResponseWriter, req *http.Request) {
		// TODO: Serve up some docs
	})

	// Matches /v2/pies/:id
	v2API.GET("/pies/:id", requireAuth.Then(func(ctx canis.Context, resp http.ResponseWriter, req *http.Request) {
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
	}))
	// All the /v2 calls should have the common middleware
	v2 := common.Then(v2API)

	// Handle any /v2 requests
	router.Handle("", "/v2", v2)

	err = http.ListenAndServe("localhost:8080", router)
	if err != nil {
		log.Fatal(err)
	}

	// Asking for / and /docs does not require auth, but requests are logged and throttled
	router.GET("/", common.Then(serveDocs))
	router.GET("/docs", common.Then(serveDocs))

	// All other requests require auth
	router.Then(requireAuth.Then(API))
}
