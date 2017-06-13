// users
package main

import (
	//"code.google.com/p/go.crypto/scrypt"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/scrypt"
	"regexp"
	"strconv"
	"time"
)

type Facebook struct {
	//URL         string
	ID          string
	AccessToken string
}

type Twitter struct {
	Name string
	ID   string
}

type GooglePlus struct {
	ID string
}

type User struct {
	ID          uint
	Name        Name
	Salt        string
	Hash        string
	Permissions Permissions
	Description DisplayText
	Email       DisplayText
	Karma       uint
	AvatarValue string
	Facebook    Facebook
	Twitter     Twitter
	GooglePlus  GooglePlus
	JoinDate    string
	Enabled     string
	Preferences UserPreferences
}

func (user User) AvatarSmall() string {
	if user.AvatarValue != "Y" {
		return DEFAULT_AVATAR
	}
	return fmt.Sprintf("%savatars/%d-small.jpg", URL_WRITABLE_PATH, user.ID)
}
func (user User) AvatarLarge() string {
	if user.AvatarValue != "Y" {
		return DEFAULT_AVATAR
	}
	return fmt.Sprintf("%savatars/%d-large.jpg", URL_WRITABLE_PATH, user.ID)
}

////////////////////////////////////////////////////////////////////////////////

type Permissions struct {
	str string
	rx  *regexp.Regexp
}

// check to see if permissions match
func (perm Permissions) Match(s string) bool {
	if s == "" {
		return true
	}
	if perm.rx == nil {
		perm.rx = regexCompile(perm.RegEx())
	}

	return perm.rx.MatchString(s)
}

func (perm Permissions) RegEx() string {
	if perm.str == "" {
		return "[0]"
	}
	return "[" + perm.str + "]"
}

////////////////////////////////////////////////////////////////////////////////

type Name struct {
	Alias DisplayText
	First DisplayText
	Full  DisplayText
}

func (name Name) Short() DisplayText {
	if name.Alias.Value == "" {
		return DisplayText{name.First.Value}
	}
	return DisplayText{name.Alias.Value}
}

func (name Name) Long() DisplayText {
	if name.Alias.Value == "" {
		return DisplayText{name.Full.Value}
	}
	return DisplayText{name.Alias.Value}
}

////////////////////////////////////////////////////////////////////////////////

type UserPreferences struct {
	PublicEmail      string
	PublicTwitter    string
	PublicGooglePlus string
	PublicFacebook   string
}

func (pref UserPreferences) ExportPreferences(user *User) {
	if pref.PublicEmail != "Y" {
		user.Email.Value = ""
	}

	if pref.PublicFacebook != "Y" {
		user.Facebook.ID = ""
		user.Twitter.Name = ""
	}

	if pref.PublicTwitter != "Y" {
		user.Twitter.ID = ""
		user.Twitter.Name = ""
	}

	if pref.PublicGooglePlus != "Y" {
		user.GooglePlus.ID = ""
	}
}

////////////////////////////////////////////////////////////////////////////////

func passwordHash(session *Session, password string) string {
	// scrypt password
	r, err := scrypt.Key([]byte(password), []byte((session.User.Salt + session.User.JoinDate)), 16384, 8, 1, 32)
	isErr(session, err, false, "creating scrypt key", "passwordHash")
	skey := base64.URLEncoding.EncodeToString(r)

	// create a sha512 hash from that key
	hash := sha512.New()
	_, err = hash.Write([]byte(skey + session.User.Salt))
	isErr(session, err, false, "creating sha512 hash", "passwordHash")

	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

////////////////////////////////////////////////////////////////////////////////

func userAdd(session *Session, password string) (err error) {
	session.User.JoinDate = time.Now().Format(DATE_FMT_DATABASE)

	var (
		password_hash string
		pw_seed       int64
	//	avatar        string = "N"
	)

	session.User.Hash = validationHash(session, session.User.Name.Long().Value+session.r.RemoteAddr+session.User.Name.Short().Value+session.User.JoinDate)
	session.ID = session.CreateSessionID()

	fb_seed, _ := strconv.ParseInt(session.User.Facebook.ID, 10, 64)
	tw_seed, _ := strconv.ParseInt(session.User.Twitter.ID, 10, 64)
	gp_seed, _ := strconv.ParseInt(session.User.GooglePlus.ID, 10, 64)
	for i := 0; i < len([]byte(password)) && i < 8; i++ {
		pw_seed += int64([]byte(password)[i])
	}
	seed := time.Now().UnixNano() - fb_seed - tw_seed - gp_seed - pw_seed
	session.User.Salt = randomString(seed, 15)

	if password != "" {
		password_hash = passwordHash(session, password)
	}

	//if session.User.AvatarValue != "" {
	//	avatar = "Y"
	//}
	transaction, r, err := dbInsertRow(session, true, false, SQL_ADD_NEW_USER,
		session.User.Name.Alias.Value, session.User.Name.First.Value, session.User.Name.Full.Value, password_hash, session.User.Salt, session.User.Hash, session.User.JoinDate,
		session.User.Description.Value, session.User.Email.Value, session.ID,
		session.User.Twitter.ID, session.User.Facebook.ID, session.User.Facebook.AccessToken, session.User.AvatarValue)

	if err != nil {
		isErr(session, err, true, `"dbInsertRow(..., SQL_ADD_NEW_USER, ...)"`, "userAdd")
		return err
	}

	row_id, err := r.LastInsertId()
	if err != nil {
		isErr(session, err, true, "getting row ID", "userAdd")
		return err
	}

	_, err = dbInsertTransaction(session, transaction, true, true, SQL_ADD_USER_PREFERENCES, row_id)
	if err != nil {
		isErr(session, err, true, `"dbInsertTransaction(..., SQL_ADD_USER_PREFERENCES, ...)"`, "userAdd")
		return err
	}

	session.User.ID = uint(row_id)
	session.WriteSessionCookies(true)
	return nil
}
