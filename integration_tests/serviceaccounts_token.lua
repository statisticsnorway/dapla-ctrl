local user = User.new("user", "user@usersen.com", "123")
local admin = User.new("admin", "admin@adminsen.com", "4332")
admin:admin(true)

Team.new("someteamname", "724")

Test.gql("Create global service account as non-admin", function(t)
	t.addHeader("x-user-email", user:email())

	t.query [[
		mutation {
			createServiceAccount(
				input: {
					name: "my-sa"
					description: "some description"
					teamSlug: "someteamname"
				}
			) {
				serviceAccount {
					id
				}
			}
		}
	]]

	t.check {
		data = Null,
		errors = {
			{
				message = "You are authenticated, but your account is not authorized to perform this action.",
				path = {
					"createServiceAccount",
				},
			},
		},
	}
end)

Test.gql("Create team service account as admin", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query([[
		mutation {
			createServiceAccount(
				input: {
					name: "team-sa"
					description: "some description"
					teamSlug: "someteamname"
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
	]])

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

Test.gql("Update service account as non-admin", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			updateServiceAccount(
				input: {
					serviceAccountID: "%s"
					description: "new description"
				}
			) {
				serviceAccount {
					id
				}
			}
		}
	]], State.saID))

	t.check {
		data = Null,
		errors = {
			{
				message = "You are authenticated, but your account is not authorized to perform this action.",
				path = {
					"updateServiceAccount",
				},
			},
		},
	}
end)

Test.gql("Update service account as admin", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		mutation {
			updateServiceAccount(
				input: {
					serviceAccountID: "%s"
					description: "new description"
				}
			) {
				serviceAccount {
					id
					description
				}
			}
		}
	]], State.saID))

	t.check {
		data = {
			updateServiceAccount = {
				serviceAccount = {
					id = State.saID,
					description = "new description",
				},
			},
		},
	}
end)

Test.gql("Assign role to service account as non-admin", function(t)
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
				}
			}
		}
	]], State.saID))

	t.check {
		data = Null,
		errors = {
			{
				message = "You are authenticated, but this functionality is not supported",
				path = {
					"assignRoleToServiceAccount",
				},
			},
		},
	}
end)

Test.gql("Assign role to service account as admin", function(t)
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
		data = Null,
		errors = {
			{
				message = "You are authenticated, but this functionality is not supported",
				path = {
					"assignRoleToServiceAccount",
				},
			},
		},
	}
end)

Test.gql("Assign duplicate role to service account as admin", function(t)
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
		errors = {
			{
				message = "You are authenticated, but this functionality is not supported",
				path = {
					"assignRoleToServiceAccount",
				},
			},
		},
		data = Null,
	}
end)

Test.gql("Assign another role to service account as admin", function(t)
	t.addHeader("x-user-email", admin:email())

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
		errors = {
			{
				message = "You are authenticated, but this functionality is not supported",
				path = {
					"assignRoleToServiceAccount",
				},
			},
		},
	}
end)

Test.gql("Revoke role from service account as non-admin", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			revokeRoleFromServiceAccount(
				input: {
					serviceAccountID: "%s"
					roleName: "Team creator"
				}
			) {
				serviceAccount {
					id
				}
			}
		}
	]], State.saID))

	t.check {
		data = Null,
		errors = {
			{
				message = "You are authenticated, but this functionality is not supported",
				path = {
					"revokeRoleFromServiceAccount",
				},
			},
		},
	}
end)

Test.gql("Revoke role from service account as admin", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		mutation {
			revokeRoleFromServiceAccount(
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
		errors = {
			{
				message = "You are authenticated, but this functionality is not supported",
				path = { "revokeRoleFromServiceAccount" },
			},
		},
	}
end)

Test.gql("Revoke unassigned role from service account as admin", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		mutation {
			revokeRoleFromServiceAccount(
				input: {
					serviceAccountID: "%s"
					roleName: "Deploy key viewer"
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
		errors = {
			{
				message = "You are authenticated, but this functionality is not supported",
				path = {
					"revokeRoleFromServiceAccount",
				},
			},
		},
	}
end)

Test.gql("Delete service account as non-admin", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			deleteServiceAccount(
				input: {
					serviceAccountID: "%s"
				}
			) {
				serviceAccountDeleted
			}
		}
	]], State.saID))

	t.check {
		data = Null,
		errors = {
			{
				message = "You are authenticated, but your account is not authorized to perform this action.",
				path = {
					"deleteServiceAccount",
				},
			},
		},
	}
end)

Test.gql("Delete service account as admin", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		mutation {
			deleteServiceAccount(
				input: {
					serviceAccountID: "%s"
				}
			) {
				serviceAccountDeleted
			}
		}
	]], State.saID))

	t.check {
		data = {
			deleteServiceAccount = {
				serviceAccountDeleted = true,
			},
		},
	}
end)
