local user = User.new()
local admin = User.new()
local unauthorized = User.new()

Helper.SQLExec([[
	DELETE FROM user_roles WHERE user_id = $1
]], unauthorized:id())

Helper.SQLExec([[
	INSERT INTO user_roles (role_name, user_id) VALUES ('Message sender', $1)
]], user:id())

Test.gql("Send message as unauthorized user", function(t)
	t.addHeader("x-user-email", unauthorized:email())

	t.query(string.format([[
		mutation {
  			sendMessage(input: {
     		recipient: "%s"
       		subject: "hello, world"
         	message: "this is a test message"
        }) {
        	messageId
        }
    }
		]], unauthorized:email()))
	t.check {
		data = Null,
		errors = {
			{
				message = 'You are authenticated, but your account is not authorized to perform this action. Specifically, you need the \"messages:send\" authorization.',
				path = { "sendMessage" },
			},
		},
	}
end)

admin:admin(true)

Test.gql("Send message as admin", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		mutation {
  			sendMessage(input: {
     		recipient: "%s"
       		subject: "hello, world"
         	message: "this is a test message"
        }) {
        	messageId
        }
    }
		]], admin:email()))
	t.check {
		data = {
			sendMessage = {
				messageId = Save("adminMessageId"),
			},
		},
	}
end)

Test.gql("Send message as user with Message sender role", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
  			sendMessage(input: {
     		recipient: "%s"
       		subject: "hello, world"
         	message: "this is a test message"
        }) {
        	messageId
        }
    }
		]], user:email()))
	t.check {
		data = {
			sendMessage = {
				messageId = Save("userMessageId"),
			},
		},
	}
end)

Test.gql("Send message to non existent recipient", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
  			sendMessage(input: {
     		recipient: "xyz@example.com"
       		subject: "hello, world"
         	message: "this is a test message"
        }) {
        	messageId
        }
    }
		]]))
	t.check {
		data = Null,
		errors = {
			{
				extensions = {
					field = "Recipient",
				},
				message = "User does not exists.",
				path = { "sendMessage" },
			},
		},
	}
end)

Test.gql("List messages", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query [[
    query {
      messages {
        nodes {
          messageId
          recipient {
         	email
          }
          actor
          status
        }
      }
    }]]

	t.check {
		data = {
			messages = {
				nodes = {
					{
						actor = admin:email(),
						messageId = State.adminMessageId,
						recipient = {
							email = admin:email(),
						},
						status = "PENDING",
					},
					{
						actor = user:email(),
						messageId = State.userMessageId,
						recipient = {
							email = user:email(),
						},
						status = "PENDING",
					},
				},
			},
		},
	}
end)

Test.gql("List messages as user without permission", function(t)
	t.addHeader("x-user-email", unauthorized:email())

	t.query [[
       query {
         messages {
           nodes {
             messageId
             recipient {
            	email
             }
             actor
             status
           }
         }
       }]]

	t.check {
		data = Null,
		errors = {
			{
				message = "You are authenticated, but your account is not authorized to perform this action.",
				path = { "messages" },
			},
		},
	}
end)

Test.gql("List messages to non existing user", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query [[
		query {
		  messages(filter: {recipient: "xyz@example.com"}) {
		    nodes {
		      actor
		      recipient {
			    email
			  }
		      status
		      messageId
		    }
		  }
		}
	]]

	t.check {
		data = Null,
		errors = {
			{
				message = "Object was not found in the database. This usually means you specified a non-existing team identifier or e-mail address.",
				path = { "messages" },
			},
		},
	}
end)

Test.gql("List messages from non existing user", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query [[
		query {
		  messages(filter: {actor: "xyz@example.com"}) {
		    nodes {
		      actor
		      recipient {
				email
	  		  }
		      status
		      messageId
		    }
		  }
		}
	]]

	t.check {
		data = {
			messages = {
				nodes = {},
			},
		},
	}
end)

Test.gql("List messages with unused status", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query [[
		query {
		  messages(filter: {status: "NONSENSE"}) {
		    nodes {
		      actor
		      recipient {
				email
			  }
		      status
		      messageId
		    }
		  }
		}
	]]

	t.check {
		data = {
			messages = {
				nodes = {},
			},
		},
	}
end)

Test.gql("Get message by messageId", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		query {
		  message(messageId: "%s") {
		    actor
		    recipient {
			  email
			}
		    status
		    messageId
		  }
		}
	]], State.adminMessageId))

	t.check {
		data = {
			message = {
				actor = admin:email(),
				messageId = State.adminMessageId,
				recipient = {
					email = admin:email(),
				},
				status = "PENDING",
			},
		},
	}
end)

Test.gql("Get message by messageId without permission", function(t)
	t.addHeader("x-user-email", unauthorized:email())

	t.query(string.format([[
		query {
		  message(messageId: "%s") {
		    actor
		    recipient {
			  email
			}
		    status
		    messageId
		  }
		}
	]], State.adminMessageId))

	t.check {
		data = {
			message = Null,
		},
		errors = {
			{
				message = "You are authenticated, but your account is not authorized to perform this action.",
				path = { "message" },
			},
		},
	}
end)

Test.gql("Get message by non existing messageId", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query [[
		query {
		  message(messageId: "00000000-0000-0000-0000-000000000000") {
		    actor
		    recipient {
			  email
			}
		    status
		    messageId
		  }
		}
	]]

	t.check {
		data = {
			message = Null,
		},
		errors = {
			{
				message = "The specified message was not found.",
				path = { "message" },
			},
		},
	}
end)
