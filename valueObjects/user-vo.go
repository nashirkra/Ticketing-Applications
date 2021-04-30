package valueObjects

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/fatih/structs"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/nashirkra/Ticketing-Applications/entity"
	"github.com/nashirkra/Ticketing-Applications/helper"
)

type UserVO interface {
	GetUser(ctx context.Context, c *redis.Client, key string) (entity.User, error)
	GetAllUser(ctx context.Context, c *redis.Client) ([]entity.User, error)
	InsertUser(ctx context.Context, c *redis.Client) (int64, error)
	UpdateUser(ctx context.Context, c *redis.Client) (int64, error)
	FindUser(ctx context.Context, c *redis.Client, value string) error
}

type User struct {
	ID       int    `json:"id,string"`
	Username string `json:"username" validate:"username"`
	Fullname string `json:"fullname"`
	Email    string `json:"email" validate:"email"`
	// EmailVerifiedAt string `json:"email_verified_at"`
	Password  string `json:"password"`
	Role      string `json:"role"`
	DeletedAt int64  `json:"deleted_at,string"`
	CreatedAt int64  `json:"created_at,string"`
	UpdatedAt int64  `json:"updated_at,string"`
}

type userVO struct {
	user      *entity.User
	validator *validator.Validate
}

func NewUserVO(u *entity.User) UserVO {

	v := validator.New()

	v.RegisterValidation("username", usernameVal)

	v.RegisterValidation("email", emailVal)

	return &userVO{
		user:      u,
		validator: v,
	}
}

func usernameVal(fl validator.FieldLevel) bool {
	return len(fl.Field().String()) > 4
}

func emailVal(fl validator.FieldLevel) bool {
	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	if len(fl.Field().String()) < 3 && len(fl.Field().String()) > 254 {
		return false
	}
	if !emailRegex.MatchString(fl.Field().String()) {
		return false
	}
	// parts := strings.Split(fl.Field().String(), "@")
	// mx, err := net.LookupMX(parts[1])
	// if err != nil || len(mx) == 0 {
	// 	return false
	// }
	return true
}

// set time UTC
func idn() time.Time {
	return helper.TimeIn("Jakarta")
}

// Insert User to Redis with initial key: user_[id]
func (u *userVO) InsertUser(ctx context.Context, c *redis.Client) (int64, error) {
	// First, validate all form
	err := u.validator.Struct(u.user)
	if err != nil {
		return 0, err
	}
	if len(u.user.Fullname) < 3 {
		return 0, errs("Fullname Required")
	}
	if len(u.user.Password) < 6 {
		return 0, errs("Password Required minimum 6 Character")
	}
	if u.user.Role != "admin" {
		if u.user.Role != "creator" {
			if u.user.Role != "participant" {
				return 0, errs("error type Role User")
			}
		}
	}
	// validate Unique Key for Username & Email
	uk, erruk := CheckUnique(ctx, c, "uk_user", u.user.Username)
	if erruk == nil {
		if len(uk) > 0 {
			return 0, fmt.Errorf("username is already in use, please choose another username of at least 6 characters")
		}
	}
	uk, erruk = CheckUnique(ctx, c, "uk_user", u.user.Email)
	if erruk == nil {
		if len(uk) > 0 {
			return 0, fmt.Errorf("email is already in use, please choose another email")
		}
	}

	// get index user_[id]
	keyFound, err := GetIndex(ctx, c, "idx_user")
	if err != nil {
		keyFound = 0
	}
	keyFound++

	// generate key
	key := "user_" + strconv.Itoa(keyFound)

	// create new struct from destination interface to define autofill value
	// thanks to Fatih for usefull utilities "github.com/fatih/structs"
	s := structs.New(u.user)

	// fill ID value
	id := s.Field("ID")
	id.Set(keyFound)

	// fill form timestamp for redis
	createdAt := s.Field("CreatedAt")
	createdAt.Set(idn().UnixNano())
	updatedAt := s.Field("UpdatedAt")
	updatedAt.Set(idn().UnixNano())

	// insert All field
	res, err := c.HSet(ctx, key, s.Map()).Result()
	if err != nil {
		return 1, err
	}

	uk_user := []string{key, u.user.Username, u.user.Email}
	b, err := json.Marshal(uk_user)
	if err != nil {
		return 1, err
	}

	_ = AddUnique(ctx, c, "uk_user", string(b))
	_ = SetIndex(ctx, c, "idx_user", keyFound)
	// result, err := fmt.Printf("result: %+v\nindexAdd: %+v\nuniqueAdd: %+v\nuniqueAdd: %+v\n", res, indexAdd, uniqueAdd1, uniqueAdd2)

	newUser, err := u.GetUser(ctx, c, key)

	if err != nil {
		return 0, err
	}

	u.user = &newUser

	return res, err
}

func (u *userVO) UpdateUser(ctx context.Context, c *redis.Client) (int64, error) {
	if u.user.ID == 0 {
		return 0, errs("User ID not match")
	}
	// get key from user
	key := "user_" + strconv.Itoa(u.user.ID)

	// get previous data
	oldData, err := u.GetUser(ctx, c, key)

	if err != nil {
		return 0, err
	}

	// define new value to be update
	// this method is taken to solve problems
	// when set new timestamp and determine null data
	// which will overwrite old data on redis
	// thanks to Fatih for usefull utilities "github.com/fatih/structs"
	s := structs.New(u.user)
	sOld := structs.New(oldData)
	for _, f := range s.Fields() {
		// fmt.Printf("field: %+v (%+v)\n", f.Name(), f.Value())

		if f.IsExported() {
			// fmt.Printf("value   : %+v\n", f.Value())
			// fmt.Printf("is zero : %+v\n", f.IsZero())

			// otherwise value will not be updated, define it with old value
			if f.IsZero() {
				if f.Name() == "UpdatedAt" {
					// update timestamp
					// f.Set(idn().Format(time.RFC3339))
					f.Set(idn().UnixNano())
				} else {
					// set old data on zero form
					f.Set(sOld.Field(f.Name()).Value())
				}
				// fmt.Printf("(result of %+v:%+v)\n", f.Name(), fv)
			}
		}
	}
	// test check fields
	// fmt.Printf("%#v", s.Map())

	// Re-validate all values
	err = u.validator.Struct(u.user)
	if err != nil {
		return 0, err
	}
	if len(u.user.Fullname) < 3 {
		return 0, errs("Fullname Required")
	}
	if len(u.user.Password) < 6 {
		return 0, errs("Password Required minimum 6 Character")
	}
	if u.user.Role != "admin" {
		if u.user.Role != "creator" {
			if u.user.Role != "participant" {
				return 0, errs("error type Role User")
			}
		}
	}

	// if Username & Email will be updated and this's validated
	// we must remove old data from redis
	usernameUpdate := false
	emailUpdate := false

	// validate Unique Key for Username & Email if will be updated
	if s.Field("Username").Value().(string) != oldData.Username {
		uk, erruk := CheckUnique(ctx, c, "uk_user", s.Field("Username").Value().(string))
		if erruk == nil {
			if len(uk) > 0 {
				return 0, fmt.Errorf("username is already in use, please choose another username of at least 6 characters")
			}
		}
		usernameUpdate = true
		// fmt.Printf("new Username: %#v, old username: %#v", s.Field("Username").Value(), oldData["Username"])
	}
	if s.Field("Email").Value().(string) != oldData.Email {
		uk, erruk := CheckUnique(ctx, c, "uk_user", s.Field("Email").Value().(string))
		if erruk == nil {
			if len(uk) > 0 {
				return 0, fmt.Errorf("email is already in use, please choose another email")
			}
		}
		emailUpdate = true
		// fmt.Printf("new Email: %#v, old Email: %#v", s.Field("Email").Value(), oldData["Email"])
	}

	// insert All field
	res, err := c.HSet(ctx, key, s.Map()).Result()
	if err != nil {
		return 1, err
	}

	if usernameUpdate || emailUpdate {
		uk_user := []string{key, s.Field("Username").Value().(string), s.Field("Email").Value().(string)}
		b, err := json.Marshal(uk_user)
		if err != nil {
			return 1, err
		}

		old_uk_user := []string{key, oldData.Username, oldData.Email}
		old_b, err := json.Marshal(old_uk_user)
		if err != nil {
			return 1, err
		}

		// add new username and/or email
		res += AddUnique(ctx, c, "uk_user", string(b)).Val()
		// remove old username and/or email
		res += DelUnique(ctx, c, "uk_user", string(old_b)).Val()
	}

	updatedUser, err := u.GetUser(ctx, c, key)

	if err != nil {
		return 0, err
	}

	u.user = &updatedUser
	// res = int64(1)
	return res, err
}

// find user by username/email (from uniquekey of user table)
func (u *userVO) FindUser(ctx context.Context, c *redis.Client, value string) error {
	found, err := CheckUnique(ctx, c, "uk_user", value)
	if err != nil {
		return err
	}
	var res []string
	if len(found) > 0 {
		err = json.Unmarshal([]byte(found[0]), &res)
		if err != nil {
			return err
		}
		if len(res) > 0 {
			*u.user, err = u.GetUser(ctx, c, res[0])
			if err != nil {
				return fmt.Errorf("%+v", err)
			}
		}
	}
	//
	return nil
}

func (u *userVO) GetUser(ctx context.Context, c *redis.Client, key string) (entity.User, error) {
	var data entity.User
	var t1 time.Time
	var t2 time.Time

	res, err := c.HGetAll(ctx, key).Result()
	if err != nil {
		return data, err
	}
	t0 := time.Unix(0, strToint64(res["DeletedAt"]))
	if !t0.IsZero() && strToint64(res["DeletedAt"]) > 0 {
		res = nil
		return data, errs("deleted")
	}
	t1 = time.Unix(0, strToint64(res["CreatedAt"]))
	t2 = time.Unix(0, strToint64(res["UpdatedAt"]))

	b, err := json.Marshal(res)
	if err != nil {
		fmt.Printf("(GetUser[1]:%s)", err)
	}

	data.CreatedAt = t1.UnixNano()
	data.UpdatedAt = t2.UnixNano()
	data.Password = res["Password"]

	// get previous data
	// Unmarshal or Decode the JSON to the interface.
	err = json.Unmarshal(b, &data)
	// fmt.Println(data)

	return data, err
}

func (u *userVO) GetAllUser(ctx context.Context, c *redis.Client) ([]entity.User, error) {

	var data []entity.User

	keys, _, err := c.Scan(ctx, 0, "user_*", 0).Result()

	if err != nil {
		return data, err
	}
	for _, v := range keys {
		user, err := u.GetUser(ctx, c, v)
		if err != nil {
			fmt.Printf("(%+v:%s)\n", v, err)
		} else {
			data = append(data, user)
		}
	}

	return data, nil
}

func errs(message string) error {
	return fmt.Errorf("%s", message)
}

// Set index for key
//  set key idx_user for user table
//  set key idx_event for event table
//  set key idx_transaction for transaction table
func SetIndex(ctx context.Context, c *redis.Client, key string, value int) *redis.Cmd {
	return c.Do(ctx, "SET", key, value)

}

// Get index for key
//  set key idx_user for user table
//  set key idx_event for event table
//  set key idx_transaction for transaction table
func GetIndex(ctx context.Context, c *redis.Client, key string) (int, error) {

	idxUser := c.Do(ctx, "GET", key).Val()
	keyFound, err := strconv.Atoi(reflect.ValueOf(idxUser).String())
	if err != nil {
		return 0, nil
	}

	return keyFound, err

}

// Add Unique key
//  set key uk_user[FieldName] for user table
//  set key uk_event[FieldName] for event table
//  set key uk_transaction[FieldName] for transaction table
func AddUnique(ctx context.Context, c *redis.Client, key string, value string) *redis.IntCmd {
	return c.SAdd(ctx, key, value)
}

// Remove Unique key
//  set key uk_user[FieldName] for user table
//  set key uk_event[FieldName] for event table
//  set key uk_transaction[FieldName] for transaction table
func DelUnique(ctx context.Context, c *redis.Client, key string, value string) *redis.IntCmd {
	return c.SRem(ctx, key, value)
}

// "*\"fathimzr\"*"
func CheckUnique(ctx context.Context, c *redis.Client, key string, value string) ([]string, error) {

	keys, _, err := c.SScan(ctx, key, uint64(c.Options().DB), "*\""+value+"\"*", 0).Result()
	return keys, err
}

func strToint64(value string) int64 {
	r, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		// fmt.Printf("error %+v:%#v\n", r, err)
		return int64(0)
	}
	return r
}
