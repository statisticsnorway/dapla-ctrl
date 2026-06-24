local alice = User.new("Alice Example", "alice@example.com", "alice")
local searcher = User.new("Searcher", "searcher@example.com", "searcher")

Helper.SQLExec([[
		SELECT pg_notify(
			'api_notify',
			jsonb_build_object(
				'table', 'users',
				'op', 'INSERT',
				'data', jsonb_build_object('id', $1::uuid, 'name', $2::text::text, 'email', $3::text)
			)::text
		);
	]], alice:id(), alice:name(), alice:email())

Test.gql("Search existing user)", function(t)
	t.addHeader("x-user-email", searcher:email())

	os.execute("sleep 0.2")

	t.query([[
		query {
			search(first: 1, filter: { query: "Alice Example", type: USER }) {
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
