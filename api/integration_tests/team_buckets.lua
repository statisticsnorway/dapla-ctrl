local user = User.new()

local managed = Team.new("managed", "724")
local selfManaged = Team.new("self-managed", "724", false)

Test.gql("Managed team has the expected 4 team buckets", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
	query {
		team(slug: "%s") {
			teamBuckets(first: 4, orderBy: {
			field: NAME
			direction: ASC
			}) {
				pageInfo {
					totalCount
				}
				nodes {
					kind
					env
					name
				}
			}
		}
	}
	]], managed:slug()))

	t.check {
		data = {
			team = {
				teamBuckets = {
					pageInfo = {
						totalCount = 4,
					},
					nodes = {
						{
							kind = "kilde",
							env = "prod",
							name = Save("bucketName"),
						},
						{
							kind = "kilde",
							env = "test",
							name = Ignore(),
						},
						{
							kind = "produkt",
							env = "prod",
							name = Ignore(),
						},
						{
							kind = "produkt",
							env = "test",
							name = Ignore(),
						},
					},
				},
			},
		},
	}
end)

Test.gql("Self-managed team has no buckets", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
	query {
		team(slug: "%s") {
			teamBuckets(first: 4, orderBy: {
			field: NAME
			direction: ASC
			}) {
				pageInfo {
					totalCount
				}
			}
		}
	}
	]], selfManaged:slug()))

	t.check {
		data = {
			team = {
				teamBuckets = {
					pageInfo = {
						totalCount = 0,
					},
				},
			},
		},
	}
end)

Test.gql("Can get bucket by its name", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
	query {
		teamBucket(name: "%s") {
			name
		}
	}
	]], State.bucketName))

	t.check {
		data = {
			teamBucket = {
				name = State.bucketName,
			},
		},
	}
end)


Test.gql("Can list all buckets", function(t)
	t.addHeader("x-user-email", user:email())

	t.query([[
	query {
			teamBuckets(first: 1) {
				pageInfo {
					totalCount
				}
			}
	}
	]])

	t.check {
		data = {
			teamBuckets = {
				pageInfo = {
					totalCount = Save("totalBuckets"),
				},
			},
		},
	}
end)

Test.gql("Can filter buckets by kind and env (and get 1/4 of the total back..)", function(t)
	t.addHeader("x-user-email", user:email())

	t.query([[
	query {
			teamBuckets(first: 1000, orderBy: {
			field: ENV
			direction: ASC
			}, filter: {
				kinds: ["produkt"]
				envs: ["prod"]
			}) {
				pageInfo {
					totalCount
				}
				nodes {
					kind
					env
				}
			}
	}
	]])

	t.check {
		data = {
			teamBuckets = {
				pageInfo = {
					totalCount = State.totalBuckets / 4,
				},
				nodes = {
					{
						kind = "produkt",
						env = "prod",
					},
				},
			},
		},
	}
end)
