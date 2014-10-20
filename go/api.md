# The CherAmi API

This is an all-JSON API. All requests and responses should have a content-type header set to `application/json`.

All endpoints except signup (`POST /users`) and login (`POST /sessions`) require an `Authorization` header in which you pass in the token that you previously received from a successful login request. For example:

    Authorization: Token 8dsfg87ef23dkos9r9wjr32232re

If the authorization header is missing, or the token is invalid or expired, an HTTP 401 response is returned. After receiving a 401, a client should try to login (`POST /sessions`) again to obtain a new token.


# Group Users

## Signup [/users]
### Signup/create a new user [POST]
Create a user given only a handle, name, email, and password.  The service will create an initial status, reputation, and default circles, as well as record the creation timestamp.  All other profile information is set in a different operation.
+ Request

        {
            "handle": "pelé",
            "name": "Edson Arantes do Nascimento",
            "email": "number10@brasil.example.com",
            "password": "Brasil Uber Alles"
        }
+ Response 201

        {
            "url": "http://cher-ami.example.com/users/pelé",
            "handle": "pelé",
            "name": "Edson Arantes do Nascimento",
            "email": "number10@brasil.example.com",
            "status": "new",
            "reputation": 1,
            "joined": "2011-10-20T08:15Z",
            "circles": [
                {"name": "public", "url": "http://cher-ami.example.com/circles/207"},
                {"name": "gold", "url": "http://cher-ami.example.com/circles/208"}
            ]
        }
+ Response 400

        {
            "reason": ("malformed json"|"missing handle"|"missing name"|"missing email"|"missing password")
        }
+ Response 403

        {
            "reason": ("invalid handle"|"invalid name"|"invalid email"|"password too weak")
        }
+ Response 409

        {
            "reason": ("handle already used"|"email already used")
        }

## Login and Logout [/sessions]

### Login [POST]
If the given username-password combination is valid, generate and return a token.
+ Request

        {
            "handle": "a string",
            "password": "a string"
        }
+ Response 201

        {
           "token": "hu876xvyft3ufib230ffn0spdfmwefna"
        }
+ Response 400

        {
            "reason": "malformed json"
        }
+ Response 403

        {
            "reason": "invalid handle-password combination"
        }

### Logout [DELETE]
The token is passed in a header (not as a parameter in the URL) and, if it is valid, the server will invalidate it.
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
+ Response 204
+ Response 403

        {
            "reason": "cannot invalidate token because it is missing or already invalid or expired"
        }


## User Search [/users{?circle,nameprefix,skip,limit,sort}]

### Get users [GET]
Fetch a desired set of users. You may filter by circle or leading characters of a name. You _must_ specify a sort order. The results _will_ be paginated since there is a potential for returning millions of users. Only a subset of user data is returned; however, the url to get the complete data _is_ returned, in good HATEOAS-style.

+ Parameters
    + circle (optional, string, `393`) ... only return users from the circle with this id
    + nameprefix (optional, string, `sta`) ... only return users whose names begin with this value (good for autocomplete)
    + skip (optional, number, `0`) ... number of results to skip, for pagination, default 0, min 0
    + limit (optional, number, `20`) ... max number of results to return, for pagination, default 20, min 1, max 100
    + sort (required, string, `joined`)

        sort results by name ascending, reputation descending, or join datetime descending (newest users first)
        + Values
            + `name`
            + `reputation`
            + `joined`

+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
+ Response 200

        [
            {
                "url": "http://cher-ami.example.com/users/pelé",
                "handle": "pelé",
                "name": "Edson Arantes do Nascimento",
                "reputation": 303,
                "joined": "2011-10-20T08:15Z"
            },
            . . .
        ]
+ Response 400

        {
            "reason": ("malformed json"|"missing sort"|"no such sort"|"malformed skip"|"malformed limit")
        }
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 403

        {
            "reason": ("you do not own or belong to this circle"|"skip out of range"|"limit out of range")
        }

## User [/users/{handle}]

### Get user by handle [GET]
Get _complete_ user data, including all profile information as well as blocked users and circle membership.
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
+ Response 200

        {
            "url": "http://cher-ami.example.com/users/pelé",
            "avatar_url": "https://images.cher-ami.example.com/users/pelé",
            "handle": "pelé",
            "name": "Edson Arantes do Nascimento",
            "email": "number10@brasil.example.com",
            "status": "retired, but coaching",
            "reputation": 1435346,
            "joined": "2011-10-20T08:15Z",
            "circles": [
                {"name": "public", "url": "http://cher-ami.example.com/circles/207"},
                {"name": "gold", "url": "http://cher-ami.example.com/circles/208"}
                {"name": "coaches", "url": "http://cher-ami.example.com/circles/5922"}
            ]
        }
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 404

        {
            "reason": "no such user"
        }


### Edit user [PATCH]
Change only basic user information here such as display name, email, and status. Use a different endpoint for complex properties like the set of users that this user has blocked, or the circles in which this user participates. Also use different endpoints to adjust reputation and to upload a new avatar picture. Note that certain user data, such as the internal id, handle, and join date, cannot be changed at all.
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    + Body

            {
                "name": "New name (optional)",
                "email": "New email (optional)",
                "status": "New status (optional)",
            }
+ Response 204
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 403

        {
            "reason": "you can only edit yourself unless you are an admin"
        }

### Delete user [DELETE]
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
+ Response 204
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 403

        {
            "reason": "you can only delete yourself unless you are an admin"
        }

## Blocking [/users/{handle}/blocked]

### Block or unblock user [PATCH]
If user A blocks user B, then B is removed from all of A's circles, public and private.  As long as B is blocked by A, B will not be allowed to join any of A's circles.
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    + Body

            {
                "handle": "pelé",
                "action": ("block"|"unblock")
            }
+ Response 204
+ Response 400

        {
            "reason": ("malformed json"|"missing handle"|"missing action"|"unknown action")
        }
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 403

        {
            "reason": "you can not block yourself"
        }
+ Response 404

        {
            "reason": "no such user"
        }

## Viewing Blocked Users [/users/{handle}/blocked{?skip,limit}]
### Get blocked users [GET]
Fetch the list of blocked users for the given user, paginated. The blocked users will always be returned in alphabetical order by handle.

+ Parameters
    + skip (optional, number, `10`) ... number of results to skip, default is 0, min 0
    + limit (optional, number, `20`) ... max number of results to return, for pagination, default 20, min 1, max 100

+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
+ Response 200

        [
            {
                "url": "http://cher-ami.example.com/users/pelé",
                "handle": "pelé",
                "name": "Edson Arantes do Nascimento",
                "reputation": 303,
                "joined": "2011-10-20T08:15Z"
            },
            . . .
        ]
+ Response 400

        {
            "reason": ("malformed json"|"malformed skip"|"malformed limit")
        }
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 403

        {
            "reason": ("limit out of range"|"you not allowed to see this user's blocked list")
        }
+ Response 404

        {
            "reason": "no such user"
        }


## Reputation [/users/{handle}/reputation]

### Adjust reputation +/- [PATCH]
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    + Body

            {
                "delta": 5,
                "action": ("inc"|"dec")
            }
+ Response 204
+ Response 400

        {
            "reason": ("malformed json"|"missing delta"|"missing action"|"unknown action"|"delta not an integer")
        }
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 403

        {
            "reason": "you can not adjust others' reputations unless you are an admin"
        }
+ Response 404

        {
            "reason": "no such user"
        }

### Set reputation directly [PUT]
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    + Body

            {
                "reputatiom": 1005,
            }
+ Response 204
+ Response 400

        {
            "reason": ("malformed json"|"missing reputation"|"reputation not an integer")
        }
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 403

        {
            "reason": "you can not set others' reputations unless you are an admin"
        }
+ Response 404

        {
            "reason": "no such user"
        }

## Avatar [/users/{handle}/avatar]
### Upload avatar [PUT]
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
            Content-type: image/png
    + Body

            ... image content ...
+ Response 204
+ Response 400

        {
            "reason": "bad media"
        }
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 403

        {
            "reason": "you can not set others' avatars unless you are an admin"
        }
+ Response 404

        {
            "reason": "no such user"
        }


# Group Circles

## Circle Creation [/circles]
### Create circle [POST]
Create a circle given only a name and description, setting the owner to the currently logged-in user. Members will be added to the circle using a different endpoint.
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    + Body

            {
                "name": "coaches",
                "description": "All my coaching friends"
            }
+ Response 201

        {
            "name": "coaches",
            "url": "http://cher-ami.example.com/circles/2997",
            "description": "All my coaching friends",
            "owner": "pelé",
            "members": "http://cher-ami.example.com/circles/2997/members",
            "creation": "2011-10-20T14:22:09Z"
        }
+ Response 400

        {
            "reason": ("malformed json"|"missing name"|"missing description")
        }
+ Response 403

        {
            "reason": ("invalid name"|"invalid description")
        }
+ Response 409

        {
            "reason": "name already used"
        }


## Circle Search [/circles{?user,before,limit}]
### Search for circles [GET]
Fetch circles, optionally restricted to those with a given owner. The results will be paginated. Only basic circle data is returned; however, the url to get the complete data is also returned. Circles are returned in order of descending creation date.  We may add custom sorting capability in the future.

+ Parameters
    + user (optional, string, `alice`) ... only return circles owned by this user
    + before (optional, string, `2015-02-28`) ... only return circles created before this date (YYYY-MM-DD)
    + limit (optional, number, `20`) ... max number of results to return, for pagination, default 20, min 1, max 100

+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
+ Response 200

        [
            {
                "name": "coaches",
                "url": "http://cher-ami.example.com/circles/2997",
                "description": "All my coaching friends",
                "owner": "pelé",
                "members": "http://cher-ami.example.com/circles/2997/members",
                "creation": "2011-10-20T14:22:09Z"
            },
            . . .
        ]
+ Response 400

        {
            "reason": ("malformed json"|"malformed before"|"malformed limit")
        }
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 403

        {
            "reason": "limit out of range"
        }

## Circle [/circles/{id}]
### Get circle by id [GET]
Get complete circle data for the circle with the given id.
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
+ Response 200

        {
            "name": "coaches",
            "url": "http://cher-ami.example.com/circles/2997",
            "description": "All my coaching friends",
            "owner": "pelé",
            "members": "http://cher-ami.example.com/circles/2997/members",
            "creation": "2011-10-20T14:22:09Z"
        },
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 404

        {
            "reason": "no such circle"
        }


### Edit circle info [PATCH]
Edits only the name and description of the circle. Members are managed elsewhere. You cannot ever change the owner or creation time.
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    + Body

            {
                "name": "New name (optional)",
                "description": "New description (optional)",
            }
+ Response 204
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 403

        {
            "reason": "you can only edit circles you own unless you are an admin"
        }


## Get Circle Members [/circles/{id}/members{?skip,limit}]

### Get circle members [GET]
Fetch the list of members of this circle, paginated. The members will always be returned in alphabetical order by handle.

+ Parameters
    + skip (optional, number, `10`) ... number of results to skip, default is 0, min 0
    + limit (optional, number, `20`) ... max number of results to return, for pagination, default 20, min 1, max 100

+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
+ Response 200

        [
            {
                "url": "http://cher-ami.example.com/users/pelé",
                "handle": "pelé",
                "name": "Edson Arantes do Nascimento",
                "reputation": 303,
                "joined": "2011-10-20T08:15Z"
            },
            . . .
        ]
+ Response 400

        {
            "reason": ("malformed json"|"malformed skip"|"malformed limit")
        }
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 403

        {
            "reason": ("limit out of range"|"you not allowed to see this circle's members")
        }
+ Response 404

        {
            "reason": "no such circle"
        }


## Manage Circle Members [/circles/{id}/members]

### Add/remove circle members [PATCH]
If a circle is public, all user can let themselves in, unless blocked by the circle owner. If private, only the owner can add. To remove a user, the requestor must be that very user, the circle owner, or an admin.
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    + Body

            {
                "handle": "pelé",
                "action": ("add"|"remove")
            }
+ Response 204
+ Response 400

        {
            "reason": ("malformed json"|"missing handle"|"missing action"|"unknown action")
        }
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 403

        {
            "reason": ("only owner can add others to private circles"|"blocked by circle owner"|"not allowed to remove")
        }
+ Response 404

        {
            "reason": "no such circle"
        }


# Group Messages

## Message Creation [/messages]

### Create message [POST]
Creates a message given content only. Server sets the id, creation timestamp, and author.
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    + Body

            {
                "circle": 488,
                "content": "Awesome trip. Check out my pics at http://bit.ly/0000000000"
            }
+ Response 201

        {
            "content": "Awesome trip. Check out my pics at http://bit.ly/0000000000",
            "url": "http://cher-ami.example.com/messages/98",
            "author": "pelé",
            "creation": "2011-10-20T14:22:09Z"
        }
+ Response 400

        {
            "reason": ("malformed json"|"missing circle"|"missing content")
        }
+ Response 403

        {
            "reason": "no such circle or you are not allowed to post to it"
        }
+ Response 413

        {
            "reason": "content too large"
        }

## Message Search [/messages{?circle,before,limit}]

### Get messages [GET]
+ Response 200

## Message [/messages/{id}]

### Get message by id [GET]
Get the message with the given id.
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
+ Response 200

            {
                "url": "http://cher-ami.example.com/messages/4",
                "author": "alice",
                "content": "I had a nice ski trip",
                "date": "2012-10-18T14:22:09Z"
            }
+ Response 401

            {
                "reason": "missing, illegal, or expired token"
            }
+ Response 404

            {
                "reason": "no such message in any circle you can see"
            }


### Delete message [DELETE]
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
+ Response 204
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 404

        {
            "reason": "you are not the author of any such message"
        }

## Comment Creation [/messages/{id}/comments]

### Post comment [POST]
Post a comment to the given message. Comments are text-only. The server sets the timestamp and sets the author to the currently logged in user.
+ Response 201


## Comment Search [/messages/{id}/comments{?before,limit}]

### Get comments for message [GET]
Fetch the comments for the given message, paginated. The comments will always be returned in order of descending creation date. The message must be of a public circle or a private circle to which the current user belongs.

+ Parameters
    + before (optional, string, `2015-02-28`) ... only return comments created before this date (YYYY-MM-DD)
    + limit (optional, number, `20`) ... max number of results to return, for pagination, default 20, min 1, max 100

+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
+ Response 200

        [
            {
                "url": "http://cher-ami.example.com/messages/4/comments/7",
                "author": "alice",
                "content": "You stupid bastard",
                "date": "2012-10-20T14:22:09Z"
            },
            . . .
        ]
+ Response 400

        {
            "reason": ("malformed json"|"malformed before"|"malformed limit")
        }
+ Response 401

        {
            "reason": "missing, illegal, or expired token"
        }
+ Response 403

        {
            "reason": "limit out of range"
        }
+ Response 404

        {
            "reason": "no such message in any circle you can see"
        }

## Comment [/messages/{id}/comments/{id}]

### Get comment by id [GET]
Get the comment with the given id. Comment must be for a message of a public circle or a private circle to which the current user belongs.
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
+ Response 200

        {
            "url": "http://cher-ami.example.com/messages/4/comments/7",
            "author": "alice",
            "content": "You stupid bastard",
            "date": "2012-10-20T14:22:09Z"
        }
+ Response 401

        {
            "reason": "missing, illegal, or expired auth token"
        }
+ Response 404

        {
            "reason": "no such comment in any circles you can see"
        }


### Delete comment [DELETE]
Permanently deletes a comment. Only succeeds if current user is the comment author, or is an admin user.
+ Request
    + Headers

            Authorization: Token xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
+ Response 204
+ Response 401

        {
            "reason": "missing, illegal, or expired auth token"
        }
+ Response 404

        {
            "reason": "you are not the author of any such comment"
        }