local user = User.new()

local managed = Team.new("managed", "724")

local environments = {
	{ env = "dev",  folder_id = "77777777" },
	{ env = "test", folder_id = "88888888" },
	{ env = "prod", folder_id = "99999999" },
}

for _, environment in ipairs(environments) do
	Helper.SQLExec([[
		INSERT INTO gcp_team_folders (team_slug, env, folder_id)
		VALUES ($1, $2, $3)
	]], managed:slug(), environment.env, environment.folder_id)
end

Test.gql("Managed team has a folder per environment", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
	query {
		team(slug: "%s") {
			teamFolders(first: 3) {
				pageInfo {
					totalCount
				}
				nodes {
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
				teamFolders = {
					pageInfo = {
						totalCount = 3,
					},
					nodes = {
						{
							env = "dev",
							name = managed:slug(),
						},
						{
							env = "prod",
							name = managed:slug(),
						},
						{
							env = "test",
							name = managed:slug(),
						},
					},
				},
			},
		},
	}
end)
