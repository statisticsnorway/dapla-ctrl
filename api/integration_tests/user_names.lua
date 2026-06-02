local user = User.new("Ettersen, Fyrst", "fet@test.no")
local contrarian = User.new("Mr. Commaless Contrarian", "mcc@commal.es")

Test.gql("First, last and full name behave as exptected", function(t)
	t.addHeader("x-user-email", user:email())

	t.query([[
		query {
			users {
				pageInfo {
					totalCount
				}
				nodes {
					name
					firstName
					lastName
					email
				}
			}
		}
		]])

	t.check {
		data = {
			users = {
				pageInfo = {
					totalCount = 2,
				},
				nodes = {
					{
						name = "Ettersen, Fyrst",
						firstName = "Fyrst",
						lastName = "Ettersen",
						email = "fet@test.no",
					},
					{
						name = "Mr. Commaless Contrarian",
						firstName = "Mr. Commaless Contrarian",
						lastName = "",
						email = "mcc@commal.es",
					},
				},
			},
		},
	}
end)
