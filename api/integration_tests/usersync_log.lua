local admin = User.new("Sync Log Admin", "syncadmin@example.com", "sync-admin", "724")
admin:admin(true)


Helper.SQLExec([[
	INSERT INTO usersync_log_entries (created_at, action, user_id, user_name, user_email, old_user_name, old_user_email, changes)
	VALUES
	(now() - interval '2 seconds', 'create_user', $1::uuid, 'USL Created User', 'usl-created@example.com', NULL, NULL, '{}'),
	(now() - interval '1 second', 'update_user', $1::uuid, 'USL Full New', 'usl-full-new@example.com', 'USL Full Old', 'usl-full-old@example.com', $2::jsonb),
	(now(), 'update_user', $1::uuid, 'USL Partial New', 'usl-same@example.com', 'USL Partial Old', 'usl-same@example.com', $3::jsonb)
]], admin:id(),
	'{"name":{"old":"USL Full Old","new":"USL Full New"},"email":{"old":"usl-full-old@example.com","new":"usl-full-new@example.com"},"sectionCode":{"old":"A","new":"B"},"jobTitle":{"old":"Dev","new":"Lead"}}',
	'{"name":{"old":"USL Partial Old","new":"USL Partial New"}}')

Test.gql("userSyncLog returns inserted entries with expected changes", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query([[
		query {
			userSyncLog(first: 10) {
				pageInfo { totalCount }
				nodes {
					__typename
					message
					userName
					userEmail
					... on UserUpdatedUserSyncLogEntry {
						oldUserName
						oldUserEmail
						changes {
							name { old new }
							email { old new }
							sectionCode { old new }
							jobTitle { old new }
						}
					}
				}
			}
		}
	]])

	t.check({
		data = {
			userSyncLog = {
				pageInfo = { totalCount = 3 },
				nodes = {
					{
						__typename = "UserUpdatedUserSyncLogEntry",
						message = "Updated user",
						userName = "USL Partial New",
						userEmail = "usl-same@example.com",
						oldUserName = "USL Partial Old",
						oldUserEmail = "usl-same@example.com",
						changes = {
							name = { old = "USL Partial Old", new = "USL Partial New" },
							email = Null,
							sectionCode = Null,
							jobTitle = Null,
						},
					},
					{
						__typename = "UserUpdatedUserSyncLogEntry",
						message = "Updated user",
						userName = "USL Full New",
						userEmail = "usl-full-new@example.com",
						oldUserName = "USL Full Old",
						oldUserEmail = "usl-full-old@example.com",
						changes = {
							name = { old = "USL Full Old", new = "USL Full New" },
							email = { old = "usl-full-old@example.com", new = "usl-full-new@example.com" },
							sectionCode = { old = "A", new = "B" },
							jobTitle = { old = "Dev", new = "Lead" },
						},
					},
					{
						__typename = "UserCreatedUserSyncLogEntry",
						message = "Created user",
						userName = "USL Created User",
						userEmail = "usl-created@example.com",
					},
				},
			},
		},
	})
end)
