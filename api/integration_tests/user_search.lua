local dev = User.new("dev usersen", "dev.usersen@example.com")
local bob = User.new("Bob Example", "bob@example.com")
local searcher = User.new("Searcher", "searcher@example.com")

Test.gql("Search existing user by name (type USER)", function(t)
	t.addHeader("x-user-email", dev:email())

	t.query([[
		query {
			search(first: 1, filter: { query: "dev", type: USER }) {
				pageInfo {
					totalCount
				}
				nodes {
					... on User {
						id
						name
						email
					}
				}
			}
		}
	]])

	t.check {
		data = {
			search = {
				pageInfo = {
					totalCount = 1,
				},
				nodes = {
					{
						id = Save("devID"),
						name = "dev usersen",
						email = "dev.usersen@example.com",
					},
				},
			},
		},
	}
end)

Test.gql("Search non-existing user returns empty result", function(t)
	t.addHeader("x-user-email", searcher:email())

	t.query([[
		query {
			search(first: 1, filter: { query: "nonexistentuser123", type: USER }) {
				pageInfo {
					totalCount
				}
				nodes {
					... on User {
						id
						name
						email
					}
				}
			}
		}
	]])

	t.check {
		data = {
			search = {
				pageInfo = {
					totalCount = 0,
				},
				nodes = {},
			},
		},
	}
end)
