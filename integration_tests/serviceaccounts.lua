local user = User.new("authenticated", "authenticated@example.com", "someid")
user:admin(true)

local teamslug = "slug-team"
local team = Team.new(teamslug, "724")

Test.gql("Create service account as authenticated user", function(t)
	t.addHeader("x-user-email", user:email())

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
	t.addHeader("x-user-email", user:email())

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
					slackChannel: "#some-channel"
				}
			) {
				team {
					id
					slug
					slackChannel
				}
			}
		}
	]], teamslug))

	t.check {
		errors = {
			{
				message = "Unauthorized",
			},
		},
	}
end)

Test.gql("Assign team creator role to service account as admin", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			assignRoleToServiceAccount(
				input: {
					serviceAccountID: "%s"
					roleName: "Team creator"
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
		data = Null,
		errors = { {
			message = "You are authenticated, but this functionality is not supported",
			path = {
				"assignRoleToServiceAccount",
			},
		} },
	}
end)

Test.gql("Add team member without permission", function(t)
	t.addHeader("authorization", sa1HeaderValue)

	Helper.emptyPubSubTopic("topic")

	t.query(string.format([[
		mutation {
			addTeamMember(
				input: {
					teamSlug: "%s"
					userEmail: "authenticated@example.com"
					role: MEMBER
				}
			) {
				member {
					role
				}
			}
		}
	]], teamSlug))

	t.check {
		errors = {
			{
				message = "Unauthorized",
			},
		},
	}
end)
