local user = User.new()

Helper.SQLExec([[
    INSERT INTO sections (code, name) VALUES
        ('606',   'Section with three teams'),
        ('707', 'Section with another team'),
        ('808', 'Section without teams')
]])

Team.new("review-a", "606", true)
Team.new("review-b", "606", false)
Team.new("review-c", "606", true)
Team.new("other-team", "707", true)

Helper.SQLExec([[
    UPDATE teams
    SET last_successful_sync = '2026-01-02 03:04:05'
    WHERE slug = 'review-c'
]])

Test.gql("section teams: filters, sorts and paginates", function(t)
    t.addHeader("x-user-email", user:email())

    t.query [[
        query {
            section(code: "606") {
                teams(first: 2, orderBy: {field: SLUG, direction: DESC}) {
                    nodes { slug isManaged lastSuccessfulSync }
                    pageInfo { totalCount hasNextPage endCursor }
                }
            }
        }
    ]]

    t.check {
        data = {
            section = {
                teams = {
                    nodes = {
                        { slug = "review-c", isManaged = true, lastSuccessfulSync = "2026-01-02T03:04:05Z" },
                        { slug = "review-b", isManaged = false, lastSuccessfulSync = Null },
                    },
                    pageInfo = {
                        totalCount = 3,
                        hasNextPage = true,
                        endCursor = Save("sectionTeamsCursor"),
                    },
                },
            },
        },
    }

    -- Use the cursor returned by the API to fetch the remaining team.
    t.query(string.format([[
        query {
            section(code: "606") {
                teams(first: 2, after: "%s", orderBy: {field: SLUG, direction: DESC}) {
                    nodes { slug }
                    pageInfo { totalCount hasNextPage }
                }
            }
        }
    ]], State.sectionTeamsCursor))

    t.check {
        data = {
            section = {
                teams = {
                    nodes = { { slug = "review-a" } },
                    pageInfo = { totalCount = 3, hasNextPage = false },
                },
            },
        },
    }
end)

Test.gql("section teams: a section without teams returns an empty list", function(t)
    t.addHeader("x-user-email", user:email())

    t.query [[
        query {
            section(code: "808") {
                teams(first: 2) {
                    nodes { slug }
                    pageInfo { totalCount hasNextPage }
                }
            }
        }
    ]]

    t.check {
        data = {
            section = {
                teams = {
                    nodes = {},
                    pageInfo = { totalCount = 0, hasNextPage = false },
                },
            },
        },
    }
end)
