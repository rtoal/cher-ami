package query

import (
	"encoding/json"
	"fmt"
	"github.com/dchest/uniuri"
	"github.com/jmcvetta/neoism"
	"time"
)

type Query struct {
	Db *neoism.Database
}

//
// Initialization
//

// Constructor, use this when creating a new Query struct.
func NewQuery(uri string) *Query {
	neo4jdb, err := neoism.Connect(uri)
	if err != nil {
		panicIfErr(err)
	}
	query := Query{neo4jdb}
	query.DatabaseInit()
	return &query
}

// Initializes the Neo4j Database
func (q Query) DatabaseInit() {
	if publicDomain := q.CreateUniquePublicDomain(); publicDomain == nil {
		fmt.Println("Unexpected database state, possible lack of PublicDomain")
	}
}

//
// Private Utilities
//

// Preforms a Cypher query, catching any unexpected behavior in a panic.
// It is ok to panic in this case as a panic at the db query level almost
// always indicates an incorrectly constructed query.
func (q Query) cypherOrPanic(query *neoism.CypherQuery) {
	panicIfErr(q.Db.Cypher(query))
}

// Asserts that err is non-nil then panics if so
func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Constants //
const (
	// Reserved Circles
	GOLD                = "Gold"
	BROADCAST           = "Broadcast"
	AUTH_TOKEN_DURATION = time.Hour
)

// Return types //

type Message struct {
	Id      string    `json:"m.id"`
	Author  string    `json:"t.handle"`
	Content string    `json:"m.content"`
	Created time.Time `json:"m.created"`
}

//
// Create
//

func (q Query) CreateUniquePublicDomain() *neoism.Node {
	// Initialize PublicDomain node
	// Nodes must have at least one property to allow unique creation
	if publicDomain, _, err := q.Db.GetOrCreateNode("PublicDomain", "iam", neoism.Props{
		"iam": "PublicDomain",
	}); err != nil {
		panic(err)
	} else {
		// Label (has to be) added separately
		panicIfErr(publicDomain.AddLabel("PublicDomain"))

		return publicDomain
	}
}

func (q Query) CreateUser(handle, email, passwordHash string) bool {
	newUser := []struct {
		Handle string    `json:"u.handle"`
		Email  string    `json:"u.email"`
		Joined time.Time `json:"u.joined"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            CREATE (u:User {
                handle:   {handle},
                name:     "",
                email:    {email},
                password: {password},
                joined:   {joined}
            })
            RETURN u.handle, u.email, u.joined
        `,
		Parameters: neoism.Props{
			"handle":   handle,
			"email":    email,
			"password": passwordHash,
			"joined":   time.Now().Local(),
		},
		Result: &newUser,
	})
	return len(newUser) > 0
}

func (q Query) CreateDefaultCirclesForUser(handle string) bool {
	created := []struct {
		Handle    string `json:"u.handle"`
		Gold      string `json:"g.name"`
		Broadcast string `json:"br.name"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH         (p:PublicDomain)
            WHERE         p.iam    = "PublicDomain"
            MATCH         (u:User)
            WHERE         u.handle = {handle}
            CREATE        (g:Circle  {name: {gold}})
            CREATE        (br:Circle {name: {broadcast}})
            CREATE 	      (u)-[:CHIEF_OF]->(g)
            CREATE        (u)-[:CHIEF_OF]->(br)
            CREATE UNIQUE (br)-[:PART_OF]->(p)
            RETURN        u.handle, g.name, br.name
        `,
		Parameters: neoism.Props{
			"handle":    handle,
			"gold":      GOLD,
			"broadcast": BROADCAST,
		},
		Result: &created,
	})
	return len(created) > 0
}

func (q Query) CreateCircle(handle, circleName string, isPublic bool,
) (circleid string, ok bool) {
	created := []struct {
		CircleName string `json:"c.name"`
		CircleId   string `json:"c.id"`
	}{}

	query := `
        MATCH   (u:User)
        WHERE   u.handle = {handle}
        CREATE  (u)-[:CHIEF_OF]->(c:Circle)
        SET     c.name   = {name}
        SET     c.id     = {id}
    `
	if isPublic {
		query = query + `
            WITH    u, c
            MATCH   (p:PublicDomain)
            WHERE   p.iam = "PublicDomain"
            CREATE  (c)-[:PART_OF]->(p)
        `
	}
	query = query + `
        RETURN c.name, c.id
    `

	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: query,
		Parameters: neoism.Props{
			"handle": handle,
			"name":   circleName,
			"id":     uniuri.NewLen(uniuri.UUIDLen),
		},
		Result: &created,
	})

	if ok = len(created) > 0; ok {
		return created[0].CircleId, ok
	} else {
		return "", ok
	}
}

func (q Query) CreateMessage(handle, content string) (messageid string, ok bool) {
	created := []struct {
		Content string `json:"m.content"`
		Id      string `json:"m.id"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (u:User)
            WHERE   u.handle = {handle}
            CREATE  (m:Message {
                content:   {content}
              , created:   {now}
              , lastsaved: {now}
              , id:        {id}
            })
            CREATE  (u)-[r:WROTE]->(m)
            RETURN  m.content, m.id
        `,
		Parameters: neoism.Props{
			"handle":  handle,
			"content": content,
			"now":     time.Now().Local(),
			"id":      uniuri.NewLen(uniuri.UUIDLen),
		},
		Result: &created,
	})

	if ok = len(created) > 0; ok {
		return created[0].Id, ok
	} else {
		return "", ok
	}
}

func (q Query) CreatePublishedRelation(messageid, circleid string) bool {
	created := []struct {
		R neoism.Relationship `json:"r"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (m:Message), (c:Circle)
            WHERE   m.id           = {messageid}
            AND     c.id           = {circleid}
            CREATE  (m)-[r:PUB_TO]->(c)
            SET     r.published_at = {now}
            RETURN  r
        `,
		Parameters: neoism.Props{
			"messageid": messageid,
			"circleid":  circleid,
			"now":       time.Now().Local(),
		},
		Result: &created,
	})
	return len(created) > 0
}

func (q Query) CreateMemberOfRelation(handle, circleid string) bool {
	joined := []struct {
		At time.Time `json:"r.at"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (u:User), (c:Circle)
            WHERE   u.handle = {handle}
            AND     c.id     = {id}
            CREATE  (u)-[r:MEMBER_OF]->(c)
            SET     r.at     = {now}
            RETURN  r.at
        `,
		Parameters: neoism.Props{
			"handle": handle,
			"id":     circleid,
			"now":    time.Now().Local(),
		},
		Result: &joined,
	})
	return len(joined) > 0
}

func (q Query) JoinBroadcastCircleOfUser(handle, target string) bool {
	created := []struct {
		At time.Time `json:"r.at"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH          (u:User)
            WHERE          u.handle = {handle}
            MATCH          (t:User)-[:CHIEF_OF]->(c:Circle)
            WHERE          t.handle = {target}
            AND            c.name   = {broadcast}
            CREATE UNIQUE  (u)-[r:MEMBER_OF]->(c)
            SET            r.at     = {now}
            RETURN         r.at
        `,
		Parameters: neoism.Props{
			"handle":    handle,
			"broadcast": BROADCAST,
			"target":    target,
			"now":       time.Now().Local(),
		},
		Result: &created,
	})
	return len(created) > 0
}

func (q Query) CreateBlockRelationFromTo(handle, target string) bool {
	res := []struct {
		Handle string              `json:"u.handle"`
		Target string              `json:"t.handle"`
		R      neoism.Relationship `json:"r"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH (u:User), (t:User)
            WHERE         u.handle = {handle}
            AND           t.handle = {target}
            CREATE UNIQUE (u)-[r:BLOCKED]->(t)
            RETURN        u.handle, t.handle, r
        `,
		Parameters: neoism.Props{
			"handle": handle,
			"target": target,
		},
		Result: &res,
	})
	return len(res) > 0
}

//
// Read
//

// Checks //

func (q Query) UserExistsByHandle(handle string) bool {
	found := []struct {
		Handle string `json:"u.handle"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (u:User)
            WHERE   u.handle = {handle}
            RETURN  u.handle
        `,
		Parameters: neoism.Props{
			"handle": handle,
		},
		Result: &found,
	})
	return len(found) > 0
}

func (q Query) CircleLinkedToPublicDomain(circleid string) bool {
	found := []struct {
		Id string `json:"c.id"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (c:Circle)-[:PART_OF]->(p:PublicDomain)
            WHERE   c.id = {id}
            RETURN  c.id
        `,
		Parameters: neoism.Props{
			"id": circleid,
		},
		Result: &found,
	})
	return len(found) > 0
}

func (q Query) UserPartOfCircle(handle, circleid string) bool {
	found := []struct {
		Id string `json:"c.id"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
			MATCH   (u:User)-[:MEMBER_OF|CHIEF_OF]->(c:Circle)
			WHERE   u.handle = {handle}
			AND     c.id     = {id}
			RETURN  c.id
		`,
		Parameters: neoism.Props{
			"handle": handle,
			"id":     circleid,
		},
		Result: &found,
	})
	return len(found) > 0
}

func (q Query) MessageIsPublished(handle, messageid, circleid string) bool {
	found := []struct {
		R *neoism.Relationship `json:"r"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (u:User)-[:WROTE]->(m:Message)-[r:PUB_TO]->(c:Circle)
            WHERE   u.handle = {handle}
            AND     m.id     = {messageid}
            AND     c.id     = {circleid}
            RETURN  r
        `,
		Parameters: neoism.Props{
			"handle":    handle,
			"messageid": messageid,
			"circleid":  circleid,
		},
		Result: &found,
	})
	return len(found) > 0
}

func (q Query) GetMessageById(messageid string) bool {
	found := []struct {
		Id int `json:"m.id"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (m:Message)
            WHERE   m.id = {id}
            RETURN  m.id
        `,
		Parameters: neoism.Props{
			"id": messageid,
		},
		Result: &found,
	})
	return len(found) > 0
}

func (q Query) HandleExists(handle string) bool {
	found := []struct {
		Handle string `json:"u.handle"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (u:User)
            WHERE   u.handle = {handle}
            RETURN  u.handle
        `,
		Parameters: neoism.Props{
			"handle": handle,
		},
		Result: &found,
	})
	return len(found) > 0
}

func (q Query) EmailExists(email string) bool {
	found := []struct {
		Email string `json:"u.email"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (u:User)
            WHERE   u.email = {email}
            RETURN  u.email
        `,
		Parameters: neoism.Props{
			"email": email,
		},
		Result: &found,
	})
	return len(found) > 0
}

func (q Query) AuthTokenBelongsToSomeUser(token string) bool {
	found := []struct {
		Handle string `json:"u.handle"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (u:User)<-[:SESSION_OF]-(a:AuthToken)
            WHERE   a.value   = {token}
            AND     a.expires > {now}
            RETURN  u.handle
        `,
		Parameters: neoism.Props{
			"token": token,
			"now":   time.Now().Local(),
		},
		Result: &found,
	})
	return len(found) == 1
}

func (q Query) BlockExistsFromTo(handle, target string) bool {
	found := []struct {
		Relation int `json:"r"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (u:User), (t:User)
            WHERE   u.handle = {handle}
            AND     t.handle = {target}
            MATCH   (u)-[r:BLOCKED]->(t)
            RETURN  r
        `,
		Parameters: neoism.Props{
			"handle": handle,
			"target": target,
		},
		Result: &found,
	})
	return len(found) > 0
}

// Get Data //

func (q Query) SearchForUsers(circle, namePrefix string, skip, limit int, sortBy string,
) (results string, count int) {
	res := []struct {
		Handle string `json:"u.handle"`
		Name   string `json:"u.name"`
		Id     int    `json:"id(u)"`
	}{}

	var query string
	props := neoism.Props{}
	regex := "(?i)" + namePrefix + ".*"

	if circle != "" {
		query = `
			MATCH  (u:User)-[]->(c:Circle)
			WHERE  c.name   =  {circle}
			AND    u.handle =~ {regex}
		`
		props = neoism.Props{
			"circle": circle,
			"regex":  regex,
			"skip":   skip,
			"limit":  limit,
			"sort":   sortBy,
		}
	} else {
		query = `
			MATCH  (u:User)
			WHERE  u.handle =~ {regex}
		`
		props = neoism.Props{
			"regex": regex,
			"skip":  skip,
			"limit": limit,
			"sort":  sortBy,
		}
	}
	query = query + `
        RETURN  u.handle, u.name, id(u)
        SKIP    {skip}
        LIMIT   {limit}
	`

	q.cypherOrPanic(&neoism.CypherQuery{
		Statement:  query,
		Parameters: props,
		Result:     &res,
	})

	if len(res) == 0 {
		return "", 0
	} else {
		bytes, err := json.Marshal(res)
		panicIfErr(err)
		return string(bytes), len(res)
	}
}

func (q Query) SearchCircles(user string, skip, limit int) (results string, count int) {
	res := []struct {
		//
		//
		// TODO JUST NAME AND ID FOR NOW.  THIS HAS TO BE FIXED TO BE LIKE API SPEC
		//
		//
		Name string `json:"c.name"`
		Id   int    `json:"id(c)"`
	}{}

	query := `
        MATCH   (u:User)-[]->(c:Circle)
        WHERE   u.handle = {user}
        RETURN  c.name, id(c)
        SKIP    {skip}
        LIMIT   {limit}
    `
	props := neoism.Props{
		"user":  user,
		"skip":  skip,
		"limit": limit,
	}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement:  query,
		Parameters: props,
		Result:     &res,
	})

	if len(res) == 0 {
		return "", 0
	} else {
		bytes, err := json.Marshal(res)
		panicIfErr(err)
		return string(bytes), len(res)
	}
}

func (q Query) GetPasswordHash(handle string) (passwordHash []byte, ok bool) {
	found := []struct {
		PasswordHash string `json:"u.password"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (u:User)
            WHERE   u.handle = {handle}
            RETURN  u.password
        `,
		Parameters: neoism.Props{
			"handle": handle,
		},
		Result: &found,
	})

	if ok := len(found) > 0; !ok {
		return []byte{}, ok
	} else {
		return []byte(found[0].PasswordHash), ok
	}
}

func (q Query) GetCircleIdByName(handle, circleName string) (circleid string) {
	found := []struct {
		Id string `json:"c.id"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
			MATCH   (u:User)-[:CHIEF_OF]->(c:Circle)
			WHERE   u.handle = {handle}
			AND     c.name   = {circle}
			RETURN  c.id
		`,
		Parameters: neoism.Props{
			"handle": handle,
			"circle": circleName,
		},
		Result: &found,
	})
	if len(found) > 0 {
		return found[0].Id
	} else {
		return ""
	}
}

func (q Query) GetAllMessagesByHandle(target string) []Message {
	messages := make([]Message, 0)
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH     (t:User)-[:WROTE]->(m:Message)
            WHERE     t.handle = {target}
            RETURN    m.id, t.handle, m.content, m.created
            ORDER BY  m.created
        `,
		Parameters: neoism.Props{
			"target": target,
		},
		Result: &messages,
	})
	return messages
}

func (q Query) GetVisibleMessageById(handle, messageid string) (message *Message, found bool) {
	messages := make([]Message, 0)
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
			MATCH   (t:User)-[:WROTE]->(m:Message)-[:PUB_TO]->(c:Circle)<-[:MEMBER_OF|CHIEF_OF]-(u:User)
			WHERE   u.handle = {handle}
            AND     m.id     = {messageid}
			RETURN  m.id, t.handle, m.content, m.created
		`,
		Parameters: neoism.Props{
			"handle":    handle,
			"messageid": messageid,
		},
		Result: &messages,
	})
	if ok := len(messages) > 0; ok {
		return &messages[0], ok
	} else {
		return nil, ok
	}
}

func (q Query) DeriveHandleFromAuthToken(token string) (handle string, ok bool) {
	found := []struct {
		Handle string `json:"u.handle"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
			MATCH   (u:User)<-[:SESSION_OF]-(a:AuthToken)
			WHERE   a.value  = {token}
			AND     {now}    < a.expires
			RETURN  u.handle
		`,
		Parameters: neoism.Props{
			"token": token,
			"now":   time.Now().Local(),
		},
		Result: &found,
	})
	if ok = len(found) > 0; ok {
		return found[0].Handle, ok
	} else {
		return "", ok
	}
}

//
// Update
//

func (q Query) SetGetNewAuthTokenForUser(handle string) string {
	created := []struct {
		Token string `json:"a.value"`
	}{}
	now := time.Now().Local()
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
                MATCH   (u:User)
                WHERE   u.handle     = {handle}
                WITH    u
                OPTIONAL MATCH (u)<-[old_r:SESSION_OF]-(old_a:AuthToken)
                DELETE  old_r, old_a
                WITH    u
                CREATE  (u)<-[r:SESSION_OF]-(a:AuthToken)
                SET     r.created_at = {now}
                SET     a.value      = {token}
                SET     a.expires    = {time}
                RETURN  a.value
            `,
		Parameters: neoism.Props{
			"handle": handle,
			"token":  "Token " + uniuri.NewLen(uniuri.UUIDLen),
			"time":   now.Add(AUTH_TOKEN_DURATION),
			"now":    now,
		},
		Result: &created,
	})
	if len(created) > 0 {
		return created[0].Token
	} else {
		return ""
	}
}

func (q Query) UpdatePassword(handle, newPasswordHash string) bool {
	updated := []struct {
		Password string `json:"u.password"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (u:User)
            WHERE   u.handle = {handle}
            SET     u.password = {new_pass}
            RETURN  u.password
        `,
		Parameters: neoism.Props{
			"handle":   handle,
			"new_pass": newPasswordHash,
		},
		Result: &updated,
	})
	return len(updated) > 0
}

func (q Query) SetGetUserName(handle, newName string) (string, bool) {
	updated := []struct {
		Name string `json:"u.name"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (u:User)
            WHERE   u.handle = {handle}
            SET     u.name = {name}
            RETURN  u.name
        `,
		Parameters: neoism.Props{
			"handle": handle,
			"name":   newName,
		},
		Result: &updated,
	})
	if ok := len(updated) > 0; ok {
		return updated[0].Name, ok
	} else {
		return "", ok
	}
}

func (q Query) UpdateMessageContent(messageid, newContent string) bool {
	updated := []struct {
		Content string
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (m:Message)
            WHERE   m.id        = {messageid}
            SET     m.content   = {content}
            SET     m.lastsaved = {now}
            RETURN  m.content
        `,
		Parameters: neoism.Props{
			"messageid": messageid,
			"content":   newContent,
			"now":       time.Now().Local(),
		},
		Result: &updated,
	})
	return len(updated) > 0
}

//
// Delete
//

func (q Query) DeleteAllNodesAndRelations() {
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH           (n)
            OPTIONAL MATCH  (n)-[r]-()
            DELETE          n, r
        `,
	})
}

func (q Query) DisconnectTargetFromAllHeldCircles(handle, target string) {
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (u:User)
            WHERE   u.handle = {handle}
            MATCH   (t:User)
            WHERE   t.handle = {target}
            OPTIONAL MATCH (u)-[:CHIEF_OF]->(c:Circle)
            OPTIONAL MATCH (t)-[r:MEMBER_OF]->(c)
            DELETE  r
        `,
		Parameters: neoism.Props{
			"handle": handle,
			"target": target,
		},
	})
}

func (q Query) DeleteUser(handle string) bool {
	deleted := []struct {
		Count int `json:"count(u)"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
                MATCH   (u:User)
                WHERE   u.handle = {handle}
                WITH    u
                OPTIONAL MATCH (a:AuthToken)-[r:SESSION_OF]->(u)
                DELETE  a, r
                WITH    u
                MATCH   (u)-[wr:WROTE]->(m:Message)-[pt:PUB_TO]->(:Circle)
                DELETE  pt, m, wr
                WITH    u
                MATCH   (u)-[mo:MEMBER_OF]->(:Circle)
                DELETE  mo
                WITH    u
                MATCH   (u)-[b:BLOCKED]->(:User)
                DELETE  b
                WITH    u
                MATCH   (u)-[co_my:CHIEF_OF]->(c:Circle)-[po_my:PART_OF]->(:PublicDomain)
                MATCH   (c)<-[mo_my:MEMBER_OF]-(:User)
                MATCH   (c)<-[pt_my:PUB_TO]-(:Message)
                DELETE  pt_my, mo_my, co_my, po_my, c, u
                RETURN  count(u)
            `,
		Parameters: neoism.Props{
			"handle": handle,
		},
		Result: &deleted,
	})
	return len(deleted) > 0 && deleted[0].Count > 0
}

func (q Query) DeletePublishedRelation(messageid, circleid string) bool {
	deleted := []struct {
		Count int `json:"count(r)"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (m:Message)-[r:PUB_TO]->(c:Circle)
            WHERE   m.id = {messageid}
            AND     c.id = {circleid}
            DELETE  r
            RETURN  count(r)
        `,
		Parameters: neoism.Props{
			"messageid": messageid,
			"circleid":  circleid,
		},
		Result: &deleted,
	})
	return len(deleted) > 0 && deleted[0].Count > 0
}

func (q Query) DestroyAuthToken(token string) bool {
	deleted := []struct {
		Handle string `json:"u.handle"`
	}{}
	q.cypherOrPanic(&neoism.CypherQuery{
		Statement: `
            MATCH   (u:User)<-[so:SESSION_OF]-(a:AuthToken)
            WHERE   a.value = {token}
            DELETE  so, a
            RETURN  u.handle
        `,
		Parameters: neoism.Props{
			"token": token,
		},
		Result: &deleted,
	})
	return len(deleted) > 0
}