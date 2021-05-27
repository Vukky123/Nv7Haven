package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/Nv7-Github/Nv7Haven/elemental"
	"github.com/Nv7-Github/Nv7Haven/nv7haven"
	"github.com/Nv7-Github/Nv7Haven/single"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"

	_ "embed"

	_ "github.com/go-sql-driver/mysql" // mysql

	"github.com/r3labs/sse/v2"
	"github.com/soheilhy/cmux"
)

const (
	dbUser = "u57_fypTHIW9t8"
	dbName = "s57_nv7haven"
)

//go:embed password.txt
var dbPassword string

func main() {
	logFile, err := os.OpenFile("logs.txt", os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	syscall.Dup2(int(logFile.Fd()), 2)

	// Error logging
	//defer recoverer()

	app := fiber.New(fiber.Config{
		BodyLimit: 1000000000,
	})
	app.Use(cors.New())
	app.Use(pprof.New())
	app.Use(recover.New(recover.Config{
		Next:              nil,
		EnableStackTrace:  true,
		StackTraceHandler: traceHandler,
	}))
	app.Get("/freememory", func(c *fiber.Ctx) error {
		debug.FreeOSMemory()
		return nil
	})

	/* Testing*/
	websockets(app)

	app.Static("/", "./index.html")

	db, err := sql.Open("mysql", dbUser+":"+dbPassword+"@tcp("+os.Getenv("MYSQL_HOST")+":3306)/"+dbName)
	if err != nil {
		panic(err)
	}

	//mysqlsetup.Mysqlsetup()

	sseServer := sse.New()

	e, err := elemental.InitElemental(app, db)
	if err != nil {
		panic(err)
	}

	err = nv7haven.InitNv7Haven(app, db)
	if err != nil {
		panic(err)
	}

	single.InitSingle(app, db)
	//b := discord.InitDiscord(db, e)
	//eod := eod.InitEoD(db)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("Gracefully shutting down...")
		app.Shutdown()
	}()

	// Set up cmux
	l, err := net.Listen("tcp", ":"+os.Getenv("PORT"))
	if err != nil {
		log.Fatal(err)
	}

	mux := cmux.New(l)
	sseL := mux.Match(cmux.HTTP1HeaderField("Accept", "text/event-stream"))
	appL := mux.Match(cmux.Any())

	go func() {
		if err := app.Listener(appL); err != nil {
			panic(err)
		}
	}()
	go func() {
		httpS := &http.Server{
			Handler: NewSseServer(sseServer),
		}
		defer httpS.Close()
		if err := httpS.Serve(sseL); err != nil {
			panic(err)
		}
	}()

	mux.Serve()

	e.Close()
	/*b.Close()
	eod.Close()*/
	db.Close()
}
