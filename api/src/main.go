package main

import (
    "github.com/ant0ine/go-json-rest/rest"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "log"
    "net/http"
    "fmt"
    "time"
)

func main() {
    port := "8228"
    handler := rest.ResourceHandler{
        EnableRelaxedContentType: true,
    }

    session, err := mgo.Dial("mongodb://localhost")
    if err != nil {
        log.Fatal(err)
    }

    database := session.DB("cher-ami")
    api := Api{session, database}

    err = handler.SetRoutes(
        &rest.Route{"POST", "/signup", api.Signup},
        //&rest.Route{"GET", "/message", GetAllMessages},
        &rest.Route{"POST", "/messages", api.CreateMessage},
        &rest.Route{"GET", "/message/:id", api.GetMessage},
        //&rest.Route{"DELETE", "/message/:id", DeleteMessage},
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Listening on port %s\n", port)
    log.Fatal(http.ListenAndServe(":"+port, &handler))
}

//
// Application Types
//

type Api struct {
    session *mgo.Session
    db      *mgo.Database // main db
}

//
// Data types
// All data types are stored in mongodb,
// this gives them an '_id' identifier
//

type Message struct {
    Owner      string
    Created    time.Time
    Content    string
    ResponseTo string       // "" if not response
    RepostOf   string       // "" if not repost
    Circles    []string
}

type UserProposal struct {
    Handle          string
    Password        string
    ConfirmPassword string
}

type UserId string

type User struct {
    Handle string
    Password string
    Joined time.Time
    Follows []UserId
    BlockedUsers []UserId
}

//
// API
//

func (a Api) Signup(w rest.ResponseWriter, r *rest.Request) {
    proposal := UserProposal{}

    // expects a json POST with "Username", "Password", "ConfirmPassword"
    err:= r.DecodeJsonPayload(&proposal)
    if err != nil {
        rest.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // ensure unique handle
    count, err := a.db.C("users").Find(bson.M{ "username": proposal.Handle }).Count()
    if count > 0 {
        rest.Error(w, proposal.Handle+" is already taken", 400)
        return
    }
    if err != nil {
        rest.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // password checks
    if proposal.Password != proposal.ConfirmPassword {
        rest.Error(w, "Passwords do not match", 400)
        return
    }

    user := User{
        proposal.Handle,
        proposal.Password,  // plaintext for now
        time.Now().Local(),
        []UserId{},
        []UserId{},
    }
    err = a.db.C("users").Insert(user)
    if err != nil {
        log.Fatal("Can't insert user: %v\n", err)
    }
}

func (a Api) CreateMessage(w rest.ResponseWriter, r *rest.Request) {
    message := Message{}

    // expects a json POST with Message properties, case-sensitive
    // use custom strings as placeholders for testing. This will be
    // remidied as the user api grows.
    err := r.DecodeJsonPayload(&message)
    if err != nil {
        rest.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if message.Content == "" {
        rest.Error(w, "please enter some content for your message", 400)
        return
    }
    message.Created = time.Now().Local()
    err = a.db.C("messages").Insert(message)
    if err != nil {
        log.Fatal("Can't insert document: %v\n", err)
    }
}

func (a Api) GetMessage(w rest.ResponseWriter, r *rest.Request) {
    id := r.PathParam("id")
    // sample
    message := Message{
        id,
        time.Now().Local(),
        "This is a sample message, ayeee",
        "",
        "",
        []string{"c_777", "c_w0qweq45", "c_888282"},
    }
    w.WriteJson(message)
}