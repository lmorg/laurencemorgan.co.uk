// sql
package main

const (
	SQL_LAYOUT  = `SELECT content FROM layout_page_design WHERE name = ?`
	SQL_MENUBAR = `
						SELECT
							lm.label,
							lm.description,
							lm.url
						FROM
							layout_menubar lm
						WHERE
							lm.enabled = "Y"
						ORDER BY
							lm.order`

	SQL_SHOW_ARTICLE = `
						SELECT
						    ca.title,
						    ca.content,
						    ca.created,
							ca.updated,
							ca.topic_id,
						    ct.title,
						    ct.description,
							ca.thread_id
						FROM
						    content_articles    ca,
						    content_topics      ct
						WHERE
						    ca.article_id   = ?
						    AND ct.topic_id = ca.topic_id
						    AND ca.enabled  = "Y"
                            AND ct.enabled  = "Y"
                            AND (ca.permissions = '' OR ca.permissions REGEXP ?)
							AND (ct.permissions = '' OR ct.permissions REGEXP ?)`

	SQL_SHOW_ALL_TOPICS = `
						SELECT
						    ct.topic_id,
						    ct.title,
						    ct.description
						FROM
						    content_topics  ct
						WHERE
						    ct.enabled = "Y"
                            AND (ct.permissions = '' OR ct.permissions REGEXP ?)
						ORDER BY
						    ct.order`

	SQL_SHOW_ALL_TOPICS_ARTICLES = `
						SELECT
						    ca.article_id,
						    ca.title,
						    ca.content
						FROM
						    content_articles  ca
						WHERE
						    ca.enabled      = "Y"
                            AND ca.hidden   = "N"
						    AND ca.topic_id = ?
                            AND (ca.permissions = '' OR ca.permissions REGEXP ?)
						ORDER BY
						    ca.created desc
						LIMIT
						    ?`

	SQL_LIST_ALL_ARTICLES = `
						SELECT
						    ca.article_id,
						    ca.topic_id,
						    ca.title,
						    ca.content,
							ct.title,
							ct.description,
							c.total
						FROM
						    content_articles  ca,
						    content_topics    ct,
							(
						        SELECT  count(ca2.article_id)   AS total
						        FROM    content_articles        ca2,
						                content_topics          ct2
						        WHERE   ca2.enabled  = "Y"
						            AND ct2.enabled  = "Y"
						            AND ca2.topic_id = ct2.topic_id
                                    AND (
                                               ca2.permissions =     ''
                                            OR ca2.permissions REGEXP ?
                                        )
                                    AND (
                                               ct2.permissions =     ''
                                            OR ct2.permissions REGEXP ?
                                        )
    						) as c
						WHERE
						    ca.enabled          = "Y"
						    AND ct.enabled      = "Y"
                            AND ca.hidden       = "N"
						    AND ct.topic_id     = ca.topic_id
                            AND (ca.permissions = '' OR ca.permissions REGEXP ?)
                            AND (ct.permissions = '' OR ct.permissions REGEXP ?)
						ORDER BY
						    ca.created desc
						LIMIT ?, ?`

	SQL_LIST_ALL_ARTICLES_BY_TOPIC = `
						SELECT
						    ca.article_id,
						    ca.topic_id,
						    ca.title,
						    ca.content,
						    ct.title,
						    ct.description,
						    c.total
						FROM
						    content_articles    ca,
						    content_topics      ct,
						    (
						        SELECT  count(ca2.article_id)   AS total
						        FROM    content_articles        ca2
						        WHERE   ca2.enabled  = "Y"
						            AND ca2.topic_id = ?
                                    AND (
                                               ca2.permissions =     ''
                                            OR ca2.permissions REGEXP ?
                                        )
						    ) as c
						WHERE
						    ca.enabled          = "Y"
                            AND ct.enabled      = "Y"
                            AND ca.hidden       = "N"
						    AND ca.topic_id     = ?
						    AND ct.topic_id     = ca.topic_id
                            AND (ca.permissions = '' OR ca.permissions REGEXP ?)
                            AND (ct.permissions = '' OR ct.permissions REGEXP ?)
						ORDER BY
						    ca.created desc
						LIMIT ?, ?`

	SQL_UPDATE_THREAD_VIEWED = `
                        INSERT INTO
                            forum_notifications
                                (
                                    thread_id,
                                    user_id,
                                    last_visit,
                                    read_comments
                                )
                            values
                                (
                                    ?,
                                    ?,
                                    ?,
                                    ?
                                )
                        ON DUPLICATE KEY UPDATE
                            last_visit    = values(last_visit),
                            read_comments = values(read_comments)`

	/*	SQL_UPDATE_COMMENT_VIEWED = `
		INSERT INTO
		    forum_notifications
		        (
		            thread_id,
		            user_id,
		            read_comments
		        )
		    values
		        (
		            ?,
		            ?,
		            ?
		        )
		ON DUPLICATE KEY UPDATE
		    read_comments = values(read_comments)`
	*/
	SQL_THREAD_HEADERS = `
						SELECT
						    -- thread details
						    ft.title,
						    ft.locked,
						    ft.thread_type,
                            ft.model,
                            fn.read_comments,

                            -- attach to forum (if applicable)
                            ff.forum_id,
                            ff.title,
                            ff.description,

						    -- attach to article (if applicable)
                            -- TODO: write this as a seperate script
						    ca.article_id,
						    ca.title,
						    ca.content,
						    ct.topic_id,
						    ct.title,
						    ct.description,

                            -- links
                            ft.link_url,
                            ft.link_content_type,
                            ft.link_description
						FROM
						    forum_threads               ft
                            LEFT OUTER JOIN forum_forums        ff
                                ON ff.forum_id      = ft.forum_id
                            LEFT OUTER JOIN content_articles    ca
						        ON ca.article_id    = ft.meta_reference
						    LEFT OUTER JOIN content_topics      ct
						        ON ct.topic_id      = ca.topic_id
                            LEFT OUTER JOIN forum_notifications fn
                                ON (
                                        fn.user_id    = ?
                                    AND fn.thread_id  = ft.thread_id
                                   )
						WHERE
						    ft.thread_id    = ?
                            AND ft.thread_type != "P"
						    -- only show if thread not deleted
						    AND ft.enabled      = "Y"
                            AND (
                                       ff.read_permissions = ''
                                    OR ff.read_permissions is null
                                    OR ff.read_permissions REGEXP ?
                                )`

	SQL_PM_HEADERS = `
						SELECT
						    -- thread details
						    ft.title,
						    ft.locked,
						    ft.thread_type,
                            ft.model,
                            fn.read_comments,

                            -- attach to forum (if applicable)
                            null,
                            null,
                            null,

						    -- attach to article (if applicable)
                            -- TODO: write this as a seperate script
						    null,
						    null,
						    null,
						    null,
						    null,
						    null,

                            -- links
                            ft.link_url,
                            ft.link_content_type,
                            ft.link_description
						FROM
						    forum_threads               ft,
                            forum_private_messages      pm
							LEFT OUTER JOIN forum_notifications fn
                            ON (
                                        fn.thread_id    = ?
                                    AND fn.user_id      = ?
                               )
						WHERE
						    ft.thread_id    = ?
                            AND ft.thread_type  = "P"
						    -- only show if thread not deleted
						    AND ft.enabled      = "Y"
                            AND pm.user_id      = ?
                            AND pm.thread_id    = ft.thread_id`

	SQL_SHOW_COMMENTS_THREADED = `
						SELECT
						    -- thread details
						    ft.title,
						    ft.locked,
						    ft.thread_type,
							-- ft.karma,

						    -- comment details
						    fc.comment_id,
						    fc.parent_id,
						    fc.user_id,
						    u.alias,
							u.first_name,
							u.full_name,
							u.avatar,
						    fc.content,
						    fc.cached,
						    fc.created,
						    fc.updated,
						    fc.karma,
                            (UNIX_TIMESTAMP(fc.created) - UNIX_TIMESTAMP(ft.created)) secs
                        FROM
						    forum_comments      fc
						    JOIN forum_threads  ft
						        ON ft.thread_id = fc.thread_id
						    JOIN users          u
						        ON u.user_id    = fc.user_id

						WHERE
						    fc.thread_id        =  ?
							AND (
                                       fc.parent_id  >= ?
                                    OR fc.comment_id  = ?
                                )

						    -- only show if thread not deleted
						    AND fc.enabled      = "Y"
							AND ft.enabled		= "Y"
							AND (ft.permissions = '' OR ft.permissions REGEXP ?)

						-- ORDER BY
						    -- fc.karma desc,`

	SQL_SHOW_COMMENTS_FLAT = `
						SELECT
						    -- thread details
						    ft.title,
						    ft.locked,
						    ft.thread_type,
							-- ft.karma,

						    -- comment details
						    fc.comment_id,
						    fc.parent_id,
						    fc.user_id,
						    u.alias,
							u.first_name,
							u.full_name,
							u.avatar,
						    fc.content,
						    fc.cached,
						    fc.created,
						    fc.updated,
						    fc.karma,
                      --      fk.karma_id,

                            c.comment_count
						FROM
						    (
						        SELECT  count(*)		as comment_count
						        FROM    forum_comments  fc2,
										forum_threads   ft2
						        WHERE   ft2.thread_id    = fc2.thread_id
						            AND ft2.thread_id	 = ?
						            AND fc2.enabled      = "Y"
						    ) c,
						    forum_comments      fc
						    JOIN forum_threads  ft
						        ON ft.thread_id = fc.thread_id
						    JOIN users          u
						        ON u.user_id    = fc.user_id
                     /*       LEFT OUTER JOIN forum_karma fk
                                ON (
                                        fk.comment_id   = fc.comment_id
                                    AND fk.given_by = ?
                                   ) */
						WHERE
						    ft.thread_id   		= ?

						    -- only show if thread not deleted
						    AND fc.enabled      = "Y"
							AND ft.enabled		= "Y"
							AND (ft.permissions = '' OR ft.permissions REGEXP ?)
						ORDER BY
						    fc.comment_id asc
						LIMIT ?, ?`

	SQL_SHOW_COMMENTS_FLAT_BY_PID = /* TODO: what's this used for? */ `
                        SELECT
                        	p.comment_num
                        FROM
                        	(
                        		SELECT
                        			@comment_num:=@comment_num+1 AS comment_num,
                        			c.comment_id,
                        			c.content
                        		FROM
                        			forum_comments c,
                                    (
                                        SELECT @comment_num:=0
                                    ) initialiser
                        		WHERE
                        			c.thread_id = ?
                        	) p
                        WHERE
                        	p.comment_id = ?`

	SQL_SHOW_COMMENT_HIGHLIGHTS = `
						SELECT
						    -- thread details
						    ft.locked,
                            ft.model,

						    -- comment details
						    fc.comment_id,
						    fc.parent_id,
						    fc.user_id,
						    u.alias,
							u.first_name,
							u.full_name,
							u.avatar,
						    fc.content,
						    fc.cached,
						    fc.created,
						    fc.updated,
						    fc.karma,

							c.comment_count
						FROM
							(
						        SELECT  count(*)		as comment_count
						        FROM    forum_comments  fc2,
										forum_threads   ft2
						        WHERE   ft2.thread_id    = fc2.thread_id
						            AND ft2.thread_id	 = ?
						            AND fc2.enabled      = "Y"
						    ) c,
						    forum_comments      fc
						    JOIN forum_threads  ft
						        ON ft.thread_id = fc.thread_id
						    JOIN users          u
						        ON u.user_id    = fc.user_id

						WHERE
						    ft.thread_id   		= ?

						    -- only show if comment and/or thread not 'deleted'
						    AND fc.enabled      = "Y"
							AND ft.enabled		= "Y"

						ORDER BY
						    fc.karma    DESC,
						    fc.created  DESC
						LIMIT ?`

	SQL_SHOW_SINGLE_COMMENT = `
						SELECT
						    -- comment details
						    fc.parent_id,
						    fc.user_id,
						    u.alias,
							u.first_name,
							u.full_name,
							u.avatar,
						    fc.content,
						    fc.cached,
						    fc.created,
						    fc.updated,
						    fc.karma,
							ft.thread_id,
							ft.locked,
							ft.title,
                            ft.model,
                            ft.thread_type
                        FROM
						    forum_comments      fc
                            JOIN users          u
						        ON u.user_id    = fc.user_id
						    JOIN forum_threads  ft
						        ON ft.thread_id = fc.thread_id
					--	    LEFT OUTER JOIN forum_private_messages pm
					--	    	ON  pm.user_id	= u.user_id
					--	    	AND pm.thread_id= ft.thread_id
               				LEFT OUTER JOIN forum_forums   ff
						        ON ff.forum_id  = ft.forum_id
                    --        LEFT OUTER JOIN
                    --            forum_notifications fn
                    --            ON (
                    --                    fn.user_id    = ?
                    --                AND fn.thread_id  = ft.thread_id
                    --               )
						WHERE
							fc.comment_id  = ?
							AND ft.thread_type != "P" -- TODO: write seperate SQL for PMs

						    -- only show if thread not deleted
						    AND fc.enabled      = "Y"
							AND ft.enabled		= "Y"
							AND (ft.permissions = '' OR ft.permissions REGEXP ?)
                            AND (ff.enabled		= "Y"     OR ff.enabled IS NULL)
                            AND (ff.enabled IS NULL OR ff.read_permissions = '' OR ff.read_permissions REGEXP ?)`

	SQL_SELECT_LATEST_COMMENT_BBCODE = `
    					SELECT		content
						FROM		forum_comment_history
						WHERE		comment_id = ?
						ORDER BY	date DESC
						LIMIT 		1`

	SQL_SHOW_ALL_FORUMS = `
						SELECT
						    ff.forum_id,
                            ff.parent_id,
						    ff.title,
						    ff.description,
                            ff.read_permissions,
                            ff.new_thread_permissions,
                            ff.thread_type,
                            ff.thread_model,
                           -- IF(fl.latest IS NULL, "", fl.latest) as latest,
                            IFNULL(fl.latest, "") latest,
							c.count
						FROM
						    forum_forums  ff
                            LEFT OUTER JOIN (
								SELECT  max(IF(ft.updated = "0000-00-00 00:00:00", ft.created, ft.updated)) as latest,
									ft.forum_id      as forum_id
								FROM     forum_threads   ft
								WHERE    ft.enabled = "Y"
								GROUP BY forum_id DESC
							) fl on fl.forum_id = ff.forum_id
                            LEFT OUTER JOIN (
                                SELECT   count(ft.thread_id) as count,
                                         ft.forum_id     as forum_id
                                FROM     forum_threads   ft
								WHERE    ft.enabled = "Y"
                                  /*   AND (
                                            ft.updated     != "0000-00-00 00:00:00"
                                             OR ft.thread_type != "A"
                                         ) */
                                GROUP BY ft.forum_id
                            ) c ON c.forum_id = ff.forum_id
						WHERE
						    ff.enabled          = "Y"
                  --        AND (ff.read_permissions = '' OR ff.read_permissions REGEXP ?)
						GROUP BY
							ff.forum_id
						ORDER BY
						    ff.order,
                            ff.parent_id,
                            ff.forum_id`

	SQL_SHOW_ALL_FORUM_THREADS = `
                        SELECT
                            ft.thread_id,
                            ft.title,
                            ft.created,
                            ft.updated,
                            IF(ft.updated = "0000-00-00 00:00:00", ft.created, ft.updated) as latest,
                            fn.last_visit,
                            fn.read_comments,
                            cu.alias,
                            cu.first_name,
                            cu.full_name,
                            uu.alias,
                            uu.first_name,
                            uu.full_name,
                            fn.subscribed,
                            (
                                SELECT count(*)
                                FROM   forum_comments      fc
                                WHERE  fc.enabled   = 'Y'
                                   AND fc.thread_id = ft.thread_id
                            )  count,
                            ft.model,
                            ft.thread_type,
                            ca.title
                        FROM
                            users                   cu,
                            users                   uu
                            JOIN
                                forum_threads       ft
                            LEFT OUTER JOIN
                                forum_notifications fn
                                ON (
                                        fn.user_id    = ?
                                    AND fn.thread_id  = ft.thread_id
                                   )
                            LEFT OUTER JOIN content_articles    ca
						        ON ca.article_id      = ft.meta_reference
                        WHERE
                            ft.enabled      = "Y"
                            AND ft.forum_id = ?
                            AND (ft.permissions = '' OR ft.permissions REGEXP ?)
                            AND ft.created_by = cu.user_id
                            AND ft.updated_by = uu.user_id
                        ORDER BY
                            latest DESC`

	SQL_SHOW_ALL_FORUM_PMS = `
                        SELECT
                            ft.thread_id,
                            ft.title,
                            ft.created,
                            ft.updated,
                            IF(ft.updated = "0000-00-00 00:00:00", ft.created, ft.updated) as latest,
                            fn.last_visit,
                            fn.read_comments,
                            cu.alias,
                            cu.first_name,
                            cu.full_name,
                            uu.alias,
                            uu.first_name,
                            uu.full_name,
                            fn.subscribed,
                            (
                                SELECT count(*)
                                FROM   forum_comments      fc
                                WHERE  fc.enabled   = 'Y'
                                   AND fc.thread_id = ft.thread_id
                            )  count,
                            ft.model,
                            ft.thread_type
                        FROM
                            users                   cu,
                            users                   uu,
                            forum_private_messages  pm
                            JOIN
                                forum_threads       ft
                            LEFT OUTER JOIN
                                forum_notifications fn
                                ON (
                                        fn.user_id    = ?
                                    AND fn.thread_id  = ft.thread_id
                                   )
                        WHERE
                            ft.enabled      = "Y"
                            AND ft.forum_id = ?
                            AND (ft.permissions = '' OR ft.permissions REGEXP ?)
                            AND ft.created_by = cu.user_id
                            AND ft.updated_by = uu.user_id
                            AND pm.thread_id  = ft.thread_id
                            AND pm.user_id	  = ?
                        ORDER BY
                            latest DESC`

	SQL_VALIDATE_REGISTRATION = `
						SELECT
							u.count,
							e.count
						FROM
							(
								SELECT  count(*)	as count
								FROM    users
								WHERE   lower(alias) = ?
							) u,
							(
								SELECT  count(*)	as count
								FROM    users
								WHERE   lower(email) = ?
							) e`

	SQL_VALIDATE_FACEBOOK = `
						SELECT
							count(*),
							user_id,
							user_hash,
							session_id,
							salt
						FROM
							users
						WHERE
							facebook_id = ?`
	/*
		SQL_VALIDATE_FACEBOOK = `
							SELECT
								count(*),
								u.user_id,
								u.user_hash,
								u.session_id,
								u.salt
							FROM
								( -- TOOD: Why did I write this voodoo crap? I don't need a nested select for that!
									SELECT	user_id,
											user_hash,
											session_id,
											salt
									FROM	users
									WHERE	facebook_id = ?
								) u
								LEFT JOIN users
									ON facebook_id = ?
							LIMIT 1`
	*/
	SQL_LOGIN_FACEBOOK = `
						UPDATE  users
						SET     session_id          = ?,
                                facebook_auth_token = ?,
                                avatar              = ?
						WHERE   facebook_id	        = ?`

	SQL_ADD_NEW_USER = `
						INSERT INTO
						    users (
								alias,
								first_name,
								full_name,
						        password,
                                salt,
								user_hash,
								join_date,
								description,
						        email,
								session_id,
						        twitter_id,
								facebook_id,
								facebook_auth_token,
								avatar,
								permissions
						    )
						VALUES (
                                ?,
						        ?,
						        ?,
						        ?,
								?,
								?,
						        ?,
						        ?,
								?,
						        ?,
								?,
						        ?,
						        ?,
								?,
								'1'
						    );`

	SQL_ADD_USER_PREFERENCES = `INSERT INTO user_preferences (user_id) VALUES (?)`

	SQL_AUTO_LOGIN = `
						SELECT
							count(session_id),
							user_id,
							alias,
							first_name,
							full_name,
							join_date,
                            permissions,
                            salt
						FROM
							users
						WHERE
							session_id = ?
							AND user_hash = ?` // TODO: when banning users, delete the session_id and user_hash too.

	SQL_VALIDATE_LOGIN = `
						SELECT
						    u.user_id,
                            u.join_date,
                            u.password,
                            u.session_id,
                            u.user_hash,
                            u.salt
						FROM
						    users   u
						WHERE
						    u.alias = ?
                            AND u.enabled = 'Y'`

	SQL_VALIDATE_COMMENT_POST = `
						SELECT
							c.count,
                            ft.thread_id,
                            ft.thread_type,
                            ft.meta_reference,
                            ft.locked,
                            ft.enabled,
                            ft.permissions
						FROM
                            (
                                SELECT
                                    count(*)    count
                                FROM
                                    forum_comments      fc
        						WHERE
        								fc.thread_id 	= ?
        							AND fc.parent_id	= ?
        							AND fc.user_id		= ?
        							AND fc.content		= ?
                            ) c
                            LEFT OUTER JOIN forum_threads    ft
                                on thread_id = ?`

	SQL_INSERT_COMMENT = `
						INSERT INTO
						    forum_comments (
						        thread_id,
						        content,
						        cached,
						        parent_id,
						        user_id
						    )
						VALUES (
						        ?,
						        ?,
						        ?,
						        ?,
						        ?
						    )`

	SQL_INSERT_COMMENT_UPDATE_THREAD = `
						UPDATE forum_threads
                        SET    updated    = ?,
                               updated_by = ?
                        WHERE  thread_id  = ?`

	SQL_INSERT_COMMENT_HISTORY = `
    					INSERT INTO
    						forum_comment_history(
								comment_id,
								content,
								reason,
								ip_address,
								user_agent
							)
						VALUES (
								?,
								?,
								?,
								?,
								?
							)`

	SQL_VALIDATE_COMMENT_KARMA = `
                        SELECT
                        	fc.comment_id,
                        	fk.enabled
                        FROM
                        	forum_comments	      fc
                        	JOIN forum_threads    ft
                                ON ft.thread_id	   = fc.thread_id
                        	LEFT OUTER JOIN forum_forums   ff
                                ON ff.forum_id	   = ft.forum_id
                        	LEFT OUTER JOIN forum_karma    fk
                        		ON (
                        			fk.comment_id  = ?
                        			AND fk.enabled = "N"
                        		   )
                        WHERE
                        	fc.comment_id 		= ?

                            -- not own comment?
                            AND fc.user_id      != ?

                        	-- check permissions
                        	AND fc.enabled		= "Y"
                        	AND ft.enabled		= "Y"
                        	AND ft.locked		= "N"
                        	AND (ft.permissions = ''  OR ft.permissions REGEXP ?)
                        	AND (ff.enabled		= "Y" OR ff.enabled IS NULL)
                        	AND (
                                       ff.read_permissions IS NULL
                                    OR ff.read_permissions = ''
                                    OR ff.read_permissions REGEXP ?
                                )`

	SQL_UPDATE_COMMENT_KARMA = `
                        INSERT INTO
                            forum_karma
                                (
                                    comment_id,
                                    given_by,
                                    modifier,
                                    message,
                                    date,
                                    enabled
                                )
                            values
                                (
                                    ?,
                                    ?,
                                    ?,
                                    ?,
                                    ?,
                                    ?
                                )
                        ON DUPLICATE KEY UPDATE
                            modifier = values(modifier),
                            message  = values(message),
                            date     = values(date),
                            enabled  = values(enabled)`

	SQL_SUM_COMMENT_KARMA = `
                        SELECT
                        	sum(fk.modifier)
                        FROM
                            forum_karma     fk
                        WHERE
                  			fk.comment_id  = ?
                   			AND fk.enabled = "Y"`

	SQL_UPDATE_COMMENT_SUMMED_KARMA = `
   						UPDATE forum_comments
                        SET    karma       = ?
                        WHERE  comment_id  = ?`

	SQL_INSERT_THREAD = `
						INSERT INTO
						    forum_threads (
						        forum_id,
                                title,
						        link_url,
						        link_content_type,
						        link_description,
						        thread_type,
						        meta_reference,
                                locked,
                                enabled,
                                permissions,
                                created_by,
                                updated_by,
                                model,
                                ip_address,
                                user_agent
						    )
						VALUES (
								?,
								?,
						        ?,
                                ?,
						        ?,
						        ?,
						        0,
                                "N",
                                "Y",
                                "", -- TODO: threads inherit forum permissions
                                ?,
                                ?,
                                ?,
                                ?,
                                ?
						    )`

	SQL_INSERT_THREAD_PM = `
						INSERT INTO
							forum_private_messages
								(thread_id, user_id)
							VALUES`

	SQL_SHOW_USER = `
                        SELECT
                            u.alias,
                            u.first_name,
                            u.full_name,
                            u.description,
                            u.email,
                            u.join_date,
                            u.enabled,
                            IFNULL(k.karma,0),
                            u.twitter_id,
                            u.google_plus_id,
                            u.facebook_id,
                            u.avatar,
                            u.permissions,
                            p.public_email,
                            p.public_twitter,
                            p.public_google_plus,
                            p.public_facebook
                        FROM
                            users    			u,
                            user_preferences	p,
                            (
                                SELECT sum(fc.karma) karma
                                FROM   forum_comments fc
                                WHERE  fc.user_id  = ?
                                   AND fc.enabled = "Y"
                            ) k
                        WHERE
                            u.user_id = ?
                            AND p.user_id = u.user_id`

	SQL_ALL_USERS_CACHE = `
						SELECT
							user_id,
							alias,
							first_name,
							full_name,
							description,
							avatar,
							permissions
						FROM
							users
						WHERE
							enabled = "Y"
						ORDER BY
							lcase(CONCAT(alias, full_name)),
							join_date`
)
