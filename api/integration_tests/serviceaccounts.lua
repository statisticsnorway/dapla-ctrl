local user = User.new("dev-user", "devuser@example.com", "someid1")
local admin = User.new("admin", "admin@example.com", "someid")
admin:admin(true)

local teamslug = "slug-team"
local team = Team.new(teamslug, "724")

Test.gql("Create service account as authenticated user", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		mutation {
			createServiceAccount(
				input: {
					name: "team-sa"
					description: "some description"
					teamSlug: "%s"
				}
			) {
				serviceAccount {
					id
					description
					roles {
						nodes {
							name
						}
					}
				}
			}
		}
	]], teamslug))

	t.check {
		data = {
			createServiceAccount = {
				serviceAccount = {
					id = Save("saID"),
					description = "some description",
					roles = {
						nodes = {},
					},
				},
			},
		},
	}
end)


Test.gql("Create service account token", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		mutation {
			createServiceAccountToken(
				input: {
					serviceAccountID: "%s"
					name: "my-token"
					description: "some description"
				}
			) {
				secret
				serviceAccount {
					id
				}
			}
		}
	]], State.saID))

	t.check {
		data = {
			createServiceAccountToken = {
				secret = Save("token"),
				serviceAccount = {
					id = State.saID,
				},
			},
		},
	}
end)

local sa1HeaderValue = string.format("Bearer %s", State.token)

Test.gql("Create new team as service account without permission", function(t)
	t.addHeader("authorization", sa1HeaderValue)

	t.query(string.format([[
		mutation {
			createTeam(
				input: {
					slug: "%s"
					displayName: "Test Team"
					sectionCode: "724"
				}
			) {
				team {
					id
					slug
				}
			}
		}
	]], teamslug))

	t.check {
		data = Null,
		errors = {
			{
				message = "You are authenticated, but your account is not authorized to perform this action.",
				path = {
					"createTeam",
				},
			},
		},
	}
end)


Test.gql("Create group", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		mutation {
			developers: createGroup(
				input: {teamSlug: "%s", category: "developers", suffix: ""}
			) {
				group {
					id
					name
					teamSlug
					category
					suffix
				}
			}
		}
	]], teamslug, teamslug))

	t.check {
		data = {
			developers = {
				group = {
					id = NotNull(),
					name = teamslug .. "-developers",
					teamSlug = teamslug,
					category = "developers",
					suffix = "",
				},
			},
		},
	}
end)

Test.gql("Add team member without permission", function(t)
	t.addHeader("authorization", sa1HeaderValue)

	t.query(string.format([[
		mutation {
			addGroupMember(
				input: {
					groupName: "%s-developers"
					userEmail: "authenticated@example.com"
				}
			) {
				member {
					user {
						email
					}
				}
			}
		}
	]], teamslug))

	t.check {
		data = Null,
		errors = {
			{
				message = "You are authenticated, but your account is not authorized to perform this action.",
				path = {
					"addGroupMember",
				},
			},
		},
	}
end)

Test.gql("Assign team owner role to service account as admin", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		mutation {
			assignRoleToServiceAccount(
				input: {
					serviceAccountID: "%s"
					roleName: "Team owner"
				}
			) {
				serviceAccount {
					id
					roles {
						nodes {
							name
						}
					}
				}
			}
		}
	]], State.saID))

	t.check {
		data = {
			assignRoleToServiceAccount = {
				serviceAccount = {
					id = State.saID,
					roles = {
						nodes = {
							{
								name = "Team owner",
							},
						},
					},
				},
			},
		},
	}
end)

Test.gql("Add team member with correct permission", function(t)
	t.addHeader("authorization", sa1HeaderValue)

	t.query(string.format([[
		mutation {
			addGroupMember(
				input: {
					groupName: "%s-developers"
					userEmail: "%s"
				}
			) {
				member {
					user {
						email
					}
				}
			}
		}
	]], teamslug, user:email()))


	t.check {
		data = {
			addGroupMember = {
				member = {
					user = {
						email = user:email(),
					},
				},
			},
		},
	}
end)



Test.gql("Send message as service account without role should be unauthorized", function(t)
	t.addHeader("authorization", sa1HeaderValue)

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
		data = Null,
		errors = {
			{
				message = Contains("Specifically, you need the \"messages:send\" authorization."),
				path = {
					"sendMessage",
				},
			},
		},
	}
end)

Test.gql("Assign role to service account should be unauthorized for user", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[

		mutation role {
		  assignRoleToServiceAccount(input: {
		    serviceAccountID: "%s"
		    roleName: "Message sender"
		  }){
		    serviceAccount {
		      name
		    }
		  }
		}
		]], State.saID))


	t.check {
		data = Null,
		errors = {
			{
				message = Contains("Specifically, you need the \"service_accounts:update\" authorization."),
				path = {
					"assignRoleToServiceAccount",
				},
			},
		},
	}
end)


Test.gql("Assign message sender role to service account should work with admin", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
	mutation role {
	  assignRoleToServiceAccount(input: {
	    serviceAccountID: "%s"
	    roleName: "Message sender"
	  }){
	    serviceAccount {
		  id
	      roles {
			nodes {
			  name
		    }
	      }
	    }
	  }
	}
	]], State.saID))


	t.check {
		data = {
			assignRoleToServiceAccount = {
				serviceAccount = {
					id = State.saID,
					roles = {
						nodes = {
							{
								name = "Message sender",
							},
							{
								name = "Team owner",
							},
						},
					},
				},
			},
		},
	}
end)



-- local result = Helper.SQLQueryRow([[
-- select service_account_id from service_account_roles
-- ]])
-- print(result)


Test.gql("Service account can send message when correct role", function(t)
	t.addHeader("authorization", sa1HeaderValue)

	Helper.emptyPubSubTopic("topic")

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
				messageId = NotNull(),
			},
		},
	}
end)
