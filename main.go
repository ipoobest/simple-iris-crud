package main

import (
	"fmt"
	"time"

	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	ID         bson.ObjectId `bson:"_id,omitempty"`
	Firstname  string        `json:"firstname"`
	Lastname   string        `json:"lastname"`
	Age        int           `json:"age"`
	Msisdn     string        `json:"msisdn"`
	InsertedAt time.Time     `json:"inserted_at" bson:"inserted_at"`
	LastUpdate time.Time     `json:"last_update" bson:"last_update"`
}

var c *mgo.Collection

func main() {
	app := iris.New()
	app.Logger().SetLevel("debug")

	app.Use(recover.New())
	app.Use(logger.New())

	session, err := mgo.Dial("127.0.0.1")
	if nil != err {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c = session.DB("usergo").C("profiles")

	// Index
	index := mgo.Index{
		Key:        []string{"msisdn"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	app.Get("/", HelloWorld)
	app.Get("/users", GetAllUsers)
	app.Get("/users/{msisdn : string}", GetUser)
	app.Post("/users", CreateUser)
	app.Put("/users/{msisdn: string}", UpdateUser)
	app.Delete("/users/{msisdn: string}", DeleteUser)
	app.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
}

func HelloWorld(ctx iris.Context) {
	ctx.JSON(context.Map{"message": "Welcome User Micro Service"})
}

func GetAllUsers(ctx iris.Context) {
	results := []User{}

	err := c.Find(nil).All(&results)
	if err != nil {
		panic(err)
	}
	ctx.JSON(context.Map{"response": results})
}

func GetUser(ctx iris.Context) {
	msisdn := ctx.Params().Get("msisdn")
	fmt.Println(msisdn)
	if msisdn == "" {
		ctx.JSON(context.Map{"response": "please pass a valid msisdn"})
	}
	result := User{}
	err := c.Find(bson.M{"msisdn": msisdn}).One(&result)
	if err != nil {
		ctx.JSON(context.Map{"response": err.Error()})
	}
	ctx.JSON(context.Map{"response": result})
}

func CreateUser(ctx iris.Context) {
	params := &User{}

	err := ctx.ReadJSON(params)
	if err != nil {
		ctx.JSON(context.Map{"response": err.Error()})
	} else {
		params.LastUpdate = time.Now()
		err := c.Insert(params)
		if err != nil {
			ctx.JSON(context.Map{"response": err.Error()})
		}
		fmt.Println("Successfully inserted into database")
		result := User{}
		err = c.Find(bson.M{"msisdn": params.Msisdn}).One(&result)
		if err != nil {
			ctx.JSON(context.Map{"response": err.Error()})
		}
		ctx.JSON(context.Map{"response": "User succesfully created", "message": result})
	}
}

func UpdateUser(ctx iris.Context) {
	msisdn := ctx.Params().Get("msisdn")
	fmt.Println(msisdn)
	if msisdn == "" {
		ctx.JSON(context.Map{"response": "please pass a valid msisdn"})
	}
	params := &User{}
	err := ctx.ReadJSON(params)
	if err != nil {
		ctx.JSON(context.Map{"response": err.Error()})
	} else {
		params.InsertedAt = time.Now()
		query := bson.M{"msisdn": msisdn}
		err = c.Update(query, params)
		if err != nil {
			ctx.JSON(context.Map{"response": err.Error()})
		} else {
			result := User{}
			err = c.Find(bson.M{"msisdn": params.Msisdn}).One(&result)
			if err != nil {
				ctx.JSON(context.Map{"response": err.Error()})
			}
			ctx.JSON(context.Map{"response": "user record successfully updated", "data": result})
		}
	}

}

func DeleteUser(ctx iris.Context) {
	msisdn := ctx.Params().Get("msisdn")
	fmt.Println(msisdn)
	if msisdn == "" {
		ctx.JSON(context.Map{"response": "please pass a valid msisdn"})
	}
	params := &User{}
	err := ctx.ReadJSON(params)
	if err != nil {
		ctx.JSON(context.Map{"response": err.Error()})
	} else {
		params.InsertedAt = time.Now()
		query := bson.M{"msisdn": msisdn}
		err = c.Remove(query)
		if err != nil {
			ctx.JSON(context.Map{"response": err.Error()})
		} else {
			ctx.JSON(context.Map{"response": "user record successfully deleted"})
		}
	}

}
