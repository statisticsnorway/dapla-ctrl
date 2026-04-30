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
				messageId = Save("messageId"),
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
				messageId = Save("messageId"),
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
