local teamSlug = "someteamname"

local user = User.new("test", "test@test.com", "exttest")
user:admin(true)

Test.gql("Create team", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			createTeam(
				input: {
					slug: "%s"
					displayName: "%s"
					purpose: "some purpose"
					sectionCode: "724"
				}
			) {
				team {
					slug
				}
			}
		}
	]], teamSlug, teamSlug))

	t.check {
		data = {
			createTeam = {
				team = {
					slug = teamSlug,
				},
			},
		},
	}
end)

Test.gql("Update purpose team", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			updateTeam(
				input: {
					slug: "%s"
					purpose: "new-purpose"
				}
			) {
				team {
					purpose
					displayName
				}
			}
		}
	]], teamSlug))

	t.check {
		data = {
			updateTeam = {
				team = {
					purpose = "new-purpose",
					displayName = "someteamname",
				},
			},
		},
	}
end)

Test.gql("Update display name team", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			updateTeam(
				input: {
					slug: "%s"
					displayName: "My Awesome Team"
				}
			) {
				team {
					purpose
					displayName
				}
			}
		}
	]], teamSlug))

	t.check {
		data = {
			updateTeam = {
				team = {
					purpose = "new-purpose",
					displayName = "My Awesome Team",
				},
			},
		},
	}
end)

Test.gql("Nothing to update", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			updateTeam(
				input: {
					slug: "%s"
				}
			) {
				team {
					purpose
					displayName
				}
			}
		}
	]], teamSlug))

	t.check {
		data = {
			updateTeam = {
				team = {
					purpose = "new-purpose",
					displayName = "My Awesome Team",
				},
			},
		},
	}
end)

-- TODO(chredvar): Add tests for invalid input for create and update team
