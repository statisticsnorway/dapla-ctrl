local alice = User.new("Alice Example", "alice@example.com", "alice")
local bob = User.new("Bob Example", "bob@example.com", "bob")
local searcher = User.new("Searcher", "searcher@example.com", "searcher")

Test.gql("Search existing user by name (type USER)", function(t)
	t.addHeader("x-user-email", searcher:email())

	t.query([[
		query {
			search(first: 1, filter: { query: "Alice", type: USER }) {
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
						id = Save("aliceID"),
						name = "Alice Example",
						email = "alice@example.com",
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
