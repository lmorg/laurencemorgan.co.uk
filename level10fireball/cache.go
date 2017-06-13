// cache.go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var cache_forums chan *CacheForums

const (
	CACHE_FORUMS_METHOD_GET        = 1
	CACHE_FORUMS_METHOD_SET        = 2
	CACHE_FORUMS_METHOD_NEW_THREAD = 3
)

type CacheForumsData struct {
	f      []Forum
	Parent map[uint][]*Forum
	Forums map[uint]*Forum
}

type CacheForums struct {
	Cache  *CacheForumsData
	Return chan *CacheForumsData
	Method int
}

func (c *CacheForums) Init() {
	c.Cache = new(CacheForumsData)
	c.Cache.Parent = make(map[uint][]*Forum)
	c.Cache.Forums = make(map[uint]*Forum)

	c.Return = make(chan *CacheForumsData)
}

func (c *CacheForums) Get() {
	c.Method = CACHE_FORUMS_METHOD_GET
	cache_forums <- c
	c.Cache = <-c.Return
}

func cacheForumsManager() {
	debugLog("Starting cacheForumManager....")
	session := new(Session)
	//session.Now = time.Now()
	var failed bool
	if failed, session.db = dbOpen(); failed == true {
		err := errors.New(`"Failed: !session.db = dbOpen()"`)
		isErr(nil, err, true, "unable to open database connection", "cacheForumsManager")
		return
	}

	defer dbClose(session.db)

	f := new(CacheForums)
	f.Init()
	cache_forums = make(chan *CacheForums)
	go func() {
		for {
			c := new(CacheForums)
			c.Init()

			//debugLog("[cache] 'cache_forums' waiting for cache request")
			c = <-cache_forums
			switch c.Method {
			default:
				debugLog("[cache] [cacheForumsManager] No method set")
				continue

			case CACHE_FORUMS_METHOD_GET:
				c.Return <- f.Cache

			case CACHE_FORUMS_METHOD_SET:
				f.Cache = c.Cache

			case CACHE_FORUMS_METHOD_NEW_THREAD:
				//fmt.Println(c.Cache.f)
				//fmt.Println(f.Cache.Forums[c.Cache.f[0].ForumID])
				//os.Exit(1)
				if len(c.Cache.f) == 0 {
					errLog("[cache] [cacheForumsManager] len(c.Cache.f) == 0")
					continue
				}
				f.Cache.Forums[c.Cache.f[0].ForumID].UpdatedDate = c.Cache.f[0].UpdatedDate
				f.Cache.Forums[c.Cache.f[0].ForumID].ThreadCount++
			}
		}
	}()

	for {
		var (
			err                  error
			forum                Forum
			forums               CacheForumsData
			f_obj                *CacheForums
			forum_thread_count   interface{}
			forum_thread_updated string
			start_time           time.Time
		)

		start_time = time.Now()

		forums.Parent = make(map[uint][]*Forum)
		forums.Forums = make(map[uint]*Forum)
		f_obj = new(CacheForums)

		// step through the forums
		subforums := dbQueryRows(session, true, &err, SQL_SHOW_ALL_FORUMS)

		//if dbOutOfSessions(session, err) {
		if err != nil {
			errLog("[cache] [cacheForumsManager] [#1] " + err.Error())
			time.Sleep(CORE_CACHE_FORUMS_ERR_TIMEOUT)
			continue
		}

		// forums[0] placeholder to prevent nil on top level forum
		forums.f = append(forums.f, Forum{})
		forums.Forums[0] = &forums.f[0]

		for dbSelectRows(session, false, &err, subforums,
			&forum.ForumID, &forum.ParentID, &forum.Title, &forum.Description, &forum.ReadPerm, &forum.NewThreadPerm, &forum.ThreadType, &forum.ThreadModel, &forum_thread_updated, &forum_thread_count) {

			if forum_thread_count != nil {
				forum.ThreadCount, _ = Atoui(string(forum_thread_count.([]uint8)))
				if forum_thread_updated[:4] == "0000" {
					forum.Latest = &forum.CreatedDate
				} else {
					forum.UpdatedDate = dateParse(forum_thread_updated, "cacheForumsManager")
					forum.UpdatedStr = forum.UpdatedDate.Format(DATE_FMT_THREAD)
					forum.Latest = &forum.UpdatedDate
				}
			} else {
				forum.ThreadCount = 0
				forum.UpdatedStr = ""
			}

			forums.f = append(forums.f, forum)
			l := len(forums.f) - 1
			forums.Parent[forum.ParentID] = append(forums.Parent[forum.ParentID], &forums.f[l])
			forums.Forums[forum.ForumID] = &forums.f[l]
		}

		if err != nil {
			errLog("[cache] [cacheForumsManager] [#2] " + err.Error())
			time.Sleep(CORE_CACHE_FORUMS_ERR_TIMEOUT)
			continue
		}

		f_obj.Cache = &forums
		f_obj.Method = CACHE_FORUMS_METHOD_SET
		cache_forums <- f_obj
		debugLog("[cache] forum cache updated in", time.Now().Sub(start_time).Nanoseconds()/1000, "µs (10^−6)")
		time.Sleep(CORE_CACHE_FORUMS_UPDATE_TIMEOUT)
	}

}

////////////////////////////////////////////////////////////////////////////////

type CacheUsers struct {
	ID      map[uint]User
	Initial map[string][]uint
	JSON    map[string]string
}

var cache_users CacheUsers

func cacheUsersManager() {
	time.Sleep(1 * time.Second)
	debugLog("Starting cacheUsersManager....")

	cache_users.ID = make(map[uint]User)
	cache_users.Initial = make(map[string][]uint)
	cache_users.JSON = make(map[string]string)

	var (
		err    error
		failed bool
	)

	generateCache := func() {
		var (
			errp    error
			u       User
			initial map[string][]uint              = make(map[string][]uint)
			j_array map[string][]map[string]string = make(map[string][]map[string]string)
			c       string
		)

		session := new(Session)
		session.Now = time.Now()

		if failed, session.db = dbOpen(); failed == true {
			err = errors.New(`"Failed: !session.db = dbOpen()"`)
			isErr(nil, err, true, fmt.Sprintf("unable to open database connection", CORE_USER_CACHE_TIMEOUT), "cacheUsersManager")
			return
		}

		defer dbClose(session.db)

		rows := dbQueryRows(session, true, &errp, SQL_ALL_USERS_CACHE)

		// pretty out of db sessions error message
		if errp != nil {
			isErr(nil, err, true, `"dbQueryRows(session,  true, &errp, SQL_ALL_USERS_CACHE)"`, "cacheUsersManager")
			return
		}

		for dbSelectRows(session, true, &errp, rows,
			&u.ID, &u.Name.Alias.Value, &u.Name.First.Value, &u.Name.Full.Value,
			&u.Description.Value, &u.AvatarValue, &u.Permissions.str) {

			if errp != nil {
				isErr(nil, err, true, `"dbSelectRows(session, true, &errp, rows, ...)"`, "cacheUsersManager")
				return
			}
			cache_users.ID[u.ID] = u

			c = strings.ToLower(u.Name.Long().Value[:1])
			initial[c] = append(initial[c], u.ID)

			j := make(map[string]string)
			j["ID"] = Itoa(u.ID)
			j["Name"] = u.Name.Full.HTMLEscaped()
			j["Alias"] = u.Name.Short().HTMLEscaped()
			j_array[c] = append(j_array[c], j)
		}

		dbClose(session.db)

		cache_users.Initial = initial

		for c = range cache_users.Initial {
			b, err := json.Marshal(j_array[c])
			if err != nil {
				isErr(session, err, true, `"json.Marshal(j_array[j])"`, "cacheUsersManager")
			} else {
				cache_users.JSON[c] = string(b)
			}
		}

		debugLog("[cache] user list cache updated in", time.Now().Sub(session.Now).Nanoseconds()/1000, "µs (10^−6)")

	}

	for {
		go generateCache()
		time.Sleep(CORE_USER_CACHE_TIMEOUT * time.Second)
	}
}
