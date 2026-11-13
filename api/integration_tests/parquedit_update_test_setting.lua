local teamSlug = "parqueditsettings"

local user = User.new("Parquedit settings admin", "parquedit-settings@test.com", "parquedit-settings-admin")
user:admin(true)

Test.gql("Create team for independent Parquedit settings", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			createTeam(
				input: {
					slug: "%s"
					displayName: "Parquedit settings"
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
			createTeam = { team = { slug = teamSlug } },
		},
	}
end)

Test.gql("Enable test Parquedit without changing production", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			updateTeam(input: { slug: "%s", hasManualEditingTest: true }) {
				team {
					hasManualEditing
					hasManualEditingTest
				}
			}
		}
	]], teamSlug))

	t.check {
		data = {
			updateTeam = {
				team = {
					hasManualEditing = false,
					hasManualEditingTest = true,
				},
			},
		},
	}
end)

Test.gql("Enable production Parquedit without changing test", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			updateTeam(input: { slug: "%s", hasManualEditing: true }) {
				team {
					hasManualEditing
					hasManualEditingTest
				}
			}
		}
	]], teamSlug))

	t.check {
		data = {
			updateTeam = {
				team = {
					hasManualEditing = true,
					hasManualEditingTest = true,
				},
			},
		},
	}
end)

Test.gql("Disable test Parquedit without changing production", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			updateTeam(input: { slug: "%s", hasManualEditingTest: false }) {
				team {
					hasManualEditing
					hasManualEditingTest
				}
			}
		}
	]], teamSlug))

	t.check {
		data = {
			updateTeam = {
				team = {
					hasManualEditing = true,
					hasManualEditingTest = false,
				},
			},
		},
	}
end)
