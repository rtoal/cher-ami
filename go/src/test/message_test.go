package api_test

import (
	"./helper"
	"encoding/json"
	. "gopkg.in/check.v1"
	"time"
)

//
// Get Authored Messages Tests
//
func (s *TestSuite) TestGetAuthoredMessagesInvalidAuth(c *C) {
	req.PostSignup("handleA", "handleA@test.io", "password1", "password1")
	res, _ := req.GetAuthoredMessages("handleA", "")
	c.Check(res.StatusCode, Equals, 401)
}

func (s *TestSuite) TestGetAuthoredMessagesOK(c *C) {
	req.PostSignup("handleA", "handleA@test.io", "password1", "password1")

	response, _ := req.PostSessions("handleA", "password1")
	sessionid_A := helper.GetSessionFromResponse(response)

	req.PostMessages("Go is going gophers!", sessionid_A)
	req.PostMessages("Hypothesize about stuff", sessionid_A)
	req.PostMessages("The nearest exit may be behind you", sessionid_A)
	req.PostMessages("I make soap.", sessionid_A)

	res, _ := req.GetAuthoredMessages("handleA", sessionid_A)

	data := struct {
		Response string
		Objects  string
		Count    int
	}{}
	helper.Unmarshal(res, &data)
	type Message struct {
		Id      string
		Author  string
		Content string
		Created time.Time
	}

	objects := make([]Message, 0)
	json.Unmarshal([]byte(data.Objects), &objects)

	c.Check(data.Response, Equals, "Found messages for user handleA")
	c.Check(res.StatusCode, Equals, 200)
	c.Check(data.Count, Equals, 4)
	c.Check(objects[0].Author, Equals, "handleA")
	c.Check(objects[0].Content, Equals, "Go is going gophers!")
	c.Check(objects[1].Content, Equals, "Hypothesize about stuff")
	c.Check(objects[2].Content, Equals, "The nearest exit may be behind you")
	c.Check(objects[3].Content, Equals, "I make soap.")
}
