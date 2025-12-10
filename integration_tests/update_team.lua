local teamSlug = "someteamname"

local user = User.new("test", "test@test.com", "exttest")

Test.gql("Create team", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			createTeam(
				input: {
					slug: "%s"
					purpose: "some purpose"
					sectionCode: "724"
				}
			) {
				team {
					slug
				}
			}
		}
	]], teamSlug))

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

Test.gql("Update team", function(t)
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
				}
			}
		}
	]], teamSlug))

	t.check {
		data = {
			updateTeam = {
				team = {
					purpose = "new-purpose",
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
				}
			}
		}
	]], teamSlug))

	t.check {
		data = {
			updateTeam = {
				team = {
					purpose = "new-purpose",
				},
			},
		},
	}
end)

-- TODO(chredvar): Add tests for invalid input for create and update team
