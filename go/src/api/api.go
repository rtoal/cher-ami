package api

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/dchest/uniuri"
	"github.com/jmcvetta/neoism"
	//"github.com/gorilla/schema"
	"fmt"
	"net/http"
	"time"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func httpError(w rest.ResponseWriter, err error) {
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

//
// Application Types
//

type Api struct {
	Db *neoism.Database
}

//
// Data types
// All data types are stored in mongodb,
// this gives them an '_id' identifier
//

/*type Message struct {
    Id         bson.ObjectId
    Owner      bson.ObjectId
    Created    time.Time
    Content    string
    ResponseTo bson.ObjectId
    RepostOf   bson.ObjectId
    Circles    []bson.ObjectId
}

type Circle struct {
    Owner      bson.ObjectId
    Members    []bson.ObjectId
    Name       string
}*/

type UserProposal struct {
	Handle          string
	Email           string
	Password        string
	ConfirmPassword string
}

//
// API
//

/*
 * Expects a json POST with "Username", "Password", "ConfirmPassword"
 */
func (a Api) Signup(w rest.ResponseWriter, r *rest.Request) {
	proposal := UserProposal{}
	err := r.DecodeJsonPayload(&proposal)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Password checks
	if proposal.Password != proposal.ConfirmPassword {
		rest.Error(w, "Passwords do not match", 400)
		return
	}

	// Ensure unique handle
	foundUsers := []struct {
		Handle string `json:"user.handle"`
	}{}
	err = a.Db.Cypher(&neoism.CypherQuery{
		Statement: `
            MATCH (user:User {handle:{handle}})
            RETURN user.handle
        `,
		Parameters: neoism.Props{
			"handle": proposal.Handle,
		},
		Result: &foundUsers,
	})
	httpError(w, err)
	if len(foundUsers) > 0 {
		rest.Error(w, proposal.Handle+" is already taken", 400)
		return
	}

	newUser := []struct {
		Handle string    `json:"user.handle"`
		Joined time.Time `json:"user.joined"`
	}{}
	err = a.Db.Cypher(&neoism.CypherQuery{
		Statement: `
            CREATE (user:User { handle:{handle}, password:{password}, joined: {joined} })
            RETURN user.handle, user.joined
        `,
		Parameters: neoism.Props{
			"handle":   proposal.Handle,
			"password": proposal.Password,
			"joined":   time.Now().Local(),
		},
		Result: &newUser,
	})
	panicErr(err)

	if len(newUser) != 1 {
		panic(fmt.Sprintf("Incorrect results len in query1()\n\tgot %d, expected 1\n", len(newUser)))
	}

	var handle string = newUser[0].Handle
	var joined string = newUser[0].Joined.Format(time.RFC1123)

	w.WriteJson(map[string]string{
		"Response": "Signed up a new user!",
		"Handle":   handle,
		"Joined":   joined,
	})

}

func (a Api) Login(w rest.ResponseWriter, r *rest.Request) {
	credentials := struct {
		Handle   string
		Password string
	}{}
	err := r.DecodeJsonPayload(&credentials)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	found := []struct {
		Handle string `json:"user.handle"`
	}{}
	err = a.Db.Cypher(&neoism.CypherQuery{
		Statement: `
            MATCH (user:User {handle:{handle}, password:{password}})
            RETURN user.handle
        `,
		Parameters: neoism.Props{
			"handle":   credentials.Handle,
			"password": credentials.Password,
		},
		Result: &found,
	})
	panicErr(err)

	// Add session hash (16 chars) to node and return it to client in json res
	if len(found) > 0 {
		sessionHash := uniuri.New()

		idResponse := []struct {
			SessionId string `json:"user.sessionid"`
		}{}
		a.Db.Cypher(&neoism.CypherQuery{
			Statement: `
				MATCH (user:User {handle:{handle}, password:{password}})
				SET user.sessionid = {sessionid}
				return user.sessionid
			`,
			Parameters: neoism.Props{
				"handle":    credentials.Handle,
				"password":  credentials.Password,
				"sessionid": sessionHash,
			},
			Result: &idResponse,
		})
		if len(idResponse) != 1 {
			panic(fmt.Sprintf("Incorrect results len in query1()\n\tgot %d, expected 1\n", len(idResponse)))
		}

		w.WriteJson(map[string]string{
			"Response":  "Logged in " + credentials.Handle + ". Note your session id.",
			"SessionId": sessionHash,
		})
		return
	} else {
		rest.Error(w, "Invaid username or password, please try again.", 400)
	}
}

func (a Api) GetUser(w rest.ResponseWriter, r *rest.Request) {
	querymap := r.URL.Query()

	// Get by handle
	if handle, ok := querymap["handle"]; ok {
		stmt := `MATCH (user:User)
                 WHERE user.handle = {handle}
                 RETURN user`
		params := neoism.Props{
			"handle": handle[0],
		}
		res := []struct {
			User neoism.Node
		}{}

		err := a.Db.Cypher(&neoism.CypherQuery{
			Statement:  stmt,
			Parameters: params,
			Result:     &res,
		})
		panicErr(err)

		u := res[0].User.Data

		w.WriteJson(u)
		return
	}

	// All users
	stmt := `MATCH (user:User)
             RETURN user.handle, user.joined
             ORDER BY user.handle`
	res := []struct {
		Handle string    `json:"user.handle"`
		Joined time.Time `json:"user.joined"`
	}{}

	err := a.Db.Cypher(&neoism.CypherQuery{
		Statement:  stmt,
		Parameters: neoism.Props{},
		Result:     &res,
	})
	panicErr(err)

	if len(res) > 0 {
		w.WriteJson(res)
	} else {
		w.WriteJson(map[string]string{
			"Response": "No results found",
		})
	}
}

func (a Api) DeleteUser(w rest.ResponseWriter, r *rest.Request) {
	querymap := r.URL.Query()

	if handle, ok := querymap["handle"]; ok {
		if password, okok := querymap["password"]; okok {

			var handle = handle[0]
			var password = password[0]

			res := []struct {
				HandleToBeDeleted string `json:"user.handle"`
			}{}
			err := a.Db.Cypher(&neoism.CypherQuery{
				Statement: `
                    MATCH (user:User {handle:{handle}, password:{password}})
                    RETURN user.handle
                `,
				Parameters: neoism.Props{
					"handle":   handle,
					"password": password,
				},
				Result: &res,
			})
			panicErr(err)

			if len(res) > 0 {
				err := a.Db.Cypher(&neoism.CypherQuery{
					// Delete user node
					Statement: `
                        MATCH (u:User {handle: {handle}})
                        DELETE u
                    `,
					Parameters: neoism.Props{
						"handle": handle,
					},
					Result: nil,
				})
				panicErr(err)

				w.WriteJson(map[string]string{
					"Response": "Deleted " + handle,
				})
				return
			} else {
				w.WriteHeader(403)
				w.WriteJson(map[string]string{
					"Response": "Could not delete user with supplied credentials",
				})
				return
			}
		}
	}
	w.WriteHeader(403)
	w.WriteJson(map[string]string{
		"Error": "Bad request parameters for delete, expected handle:String, password:String",
	})
}

// func (a Api) CreateMessage(w rest.ResponseWriter, r *rest.Request) {
//     message := Message{
//         bson.NewObjectId(),
//         bson.NewObjectId(),     // owner ID
//         time.Now().Local(),
//         "",                     // content
//         NIL_ID,
//         NIL_ID,
//         []bson.ObjectId{},
//     }

//     payload := Message{}
//     err     := r.DecodeJsonPayload(&payload)
//     if err != nil {
//         rest.Error(w, err.Error(), http.StatusInternalServerError)
//         return
//     }
//     message.Content = payload.Content

//     if message.Content == "" {
//         rest.Error(w, "please enter some content for your message", 400)
//         return
//     }

//     err = a.Db.C("messages").Insert(message)
//     if err != nil {
//         log.Fatal("Can't insert document: %v\n", err)
//     }
// }

// func (a Api) GetMessage(w rest.ResponseWriter, r *rest.Request) {
//     bid     := bson.ObjectIdHex(r.PathParam("id"))
//     message := Message{}
//     err     := a.Db.C("messages").Find(bson.M{"id": bid}).One(&message)
//     if err != nil {
//         rest.Error(w, err.Error(), http.StatusInternalServerError)
//         return
//     }
//     w.WriteJson(message)
// }

// func (a Api) DeleteMessage(w rest.ResponseWriter, r *rest.Request) {
//     bid := bson.ObjectIdHex(r.PathParam("id"))
//     err := a.Db.C("messages").Remove(bson.M{"id": bid})
//     if err != nil {
//         rest.Error(w, err.Error(), http.StatusInternalServerError)
//         return
//     }
// }
