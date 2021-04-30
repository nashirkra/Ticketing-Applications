package valueObjects

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/nashirkra/Ticketing-Applications/entity"
	"gorm.io/gorm"
)

type Event struct {
	ID                int            `json:"id" form:"id"`
	CreatorId         int            `json:"creator_id" form:"creator_id"`
	TitleEvent        string         `json:"title_event" form:"title_event"`
	LinkWebinar       string         `json:"link_webinar" form:"link_webinar"`
	Description       string         `json:"description" form:"description"`
	TypeEvent         string         `json:"type_event" form:"type_event"`
	Banner            string         `json:"banner" form:"banner"`
	Price             float64        `json:"price" form:"price"`
	Quantity          int            `json:"quantity" form:"quantity"`
	Status            string         `json:"status" form:"status"`
	EventStartDate    *time.Time     `json:"event_start_date" form:"event_start_date" time_format:"2006-01-02"`
	EventEndDate      *time.Time     `json:"event_end_date" form:"event_end_date" time_format:"2006-01-02"`
	CampaignStartDate *time.Time     `json:"campaign_start_date" form:"campaign_start_date" time_format:"2006-01-02"`
	CampaignEndDate   *time.Time     `json:"campaign_end_date" form:"campaign_end_date" time_format:"2006-01-02"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" form:"deleted_at" time_format:"unixNano"`
	CreatedAt         time.Time      `json:"created_at" form:"created_at" time_format:"unixNano"`
	UpdatedAt         time.Time      `json:"updated_at" form:"updated_at" time_format:"unixNano"`
}

type EventVO interface {
	CreateEvent(ctx context.Context, c *redis.Client) (entity.Event, error)
	GetEvent(ctx context.Context, c *redis.Client, key string) (entity.Event, error)
	GetAllEvent(ctx context.Context, c *redis.Client) ([]entity.Event, error)
}

type eventVO struct {
	event     *entity.Event
	validator *validator.Validate
}

func NewEventVO(ev *entity.Event) EventVO {
	v := validator.New()
	return &eventVO{
		event:     ev,
		validator: v,
	}
}

func (ev *eventVO) CreateEvent(ctx context.Context, c *redis.Client) (entity.Event, error) {
	// validate URL first
	_, err := url.ParseRequestURI(ev.event.LinkWebinar)
	if err != nil {
		return *ev.event, fmt.Errorf("link is not valid")
	}
	// validate Unique Key for Title & link
	uk, erruk := CheckUnique(ctx, c, "uk_event", ev.event.TitleEvent)
	if erruk == nil {
		if len(uk) > 0 {
			return *ev.event, fmt.Errorf("title is already in use, please choose another title, or update it")
		}
	}
	uk, erruk = CheckUnique(ctx, c, "uk_event", ev.event.LinkWebinar)
	if erruk == nil {
		if len(uk) > 0 {
			return *ev.event, fmt.Errorf("link is already in use, please choose another link, or update it")
		}
	}

	// get index event_[id]
	keyFound, err := GetIndex(ctx, c, "idx_event")
	if err != nil {
		keyFound = 0
	}
	keyFound++

	// generate event key event_[id]
	key := "event_" + strconv.Itoa(keyFound)

	ev.event.ID = keyFound
	ev.event.CreatedAt = idn()
	ev.event.UpdatedAt = idn()
	ev.event.Status = "Draft"
	ev.event.TypeEvent = "Online"

	// TRY to generate hash args
	args := setArgs(ev.event)

	// fmt.Printf("key: %+v\nargs:%+v", key, args)
	// insert All field
	_, err = c.HSet(ctx, key, args).Result()
	if err != nil {
		return *ev.event, fmt.Errorf("hset: %v", err)
	}

	// AddUnique()

	uk_event := []string{key, ev.event.TitleEvent, ev.event.LinkWebinar}
	b, err := json.Marshal(uk_event)
	if err != nil {
		return *ev.event, err
	}

	_ = AddUnique(ctx, c, "uk_event", string(b))
	_ = SetIndex(ctx, c, "idx_event", keyFound)

	newEvent, err := ev.GetEvent(ctx, c, key)

	if err != nil {
		return *ev.event, err
	}
	ev.event = &newEvent
	return *ev.event, err
}

func (ev *eventVO) GetEvent(ctx context.Context, c *redis.Client, key string) (entity.Event, error) {
	var data entity.Event
	var t1 time.Time
	var t2 time.Time
	var t3 time.Time
	var t4 time.Time
	var t5 time.Time
	var t6 time.Time

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
	t3 = time.Unix(0, strToint64(res["EventStartDate"]))
	t4 = time.Unix(0, strToint64(res["EventEndDate"]))
	t5 = time.Unix(0, strToint64(res["CampaignStartDate"]))
	t6 = time.Unix(0, strToint64(res["CampaignEndDate"]))
	id, _ := strconv.Atoi(res["ID"])
	creatorId, _ := strconv.Atoi(res["CreatorId"])
	typeEv := res["TypeEvent"]
	titleEv := res["TitleEvent"]

	args := make(map[string]string, len(res)-1)

	for k, v := range res {
		if !(k == "Creator") || !(k == "Participant") || !(k == "Event") {
			args[k] = v
		}
	}

	b, err := json.Marshal(args)
	if err != nil {
		fmt.Printf("(GetUser[1]:%s)", err)
	}

	data.CreatedAt = t1
	data.UpdatedAt = t2
	data.EventStartDate = &t3
	data.EventEndDate = &t4
	data.CampaignStartDate = &t5
	data.CampaignEndDate = &t6
	data.ID = id
	data.CreatorId = creatorId
	data.TypeEvent = typeEv
	data.TitleEvent = titleEv

	// get previous data
	// Unmarshal or Decode the JSON to the interface.
	err = json.Unmarshal(b, &data)
	// fmt.Println(data)

	return data, err
}
func (ev *eventVO) GetAllEvent(ctx context.Context, c *redis.Client) ([]entity.Event, error) {
	var data []entity.Event
	// find all key first
	keys, _, err := c.Scan(ctx, 0, "event_*", 1000).Result()
	if err != nil {
		return data, err
	}
	var user entity.User
	var gUser = NewUserVO(&user)
	// now, get all event
	for _, v := range keys {
		ev, err := ev.GetEvent(ctx, c, v)
		if err != nil {
			fmt.Printf("1:(%+v:%s)\n", v, err)
		} else {
			err = gUser.FindUser(ctx, c, "user_"+strconv.Itoa(ev.CreatorId))
			if err != nil {
				fmt.Printf("2:(%+v:%s)\n", v, err)
			}
			ev.Creator = user
			data = append(data, ev)
		}
	}

	return data, nil
}

// generate value type as given by json
func getField(v interface{}, field string) string {
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	fieldValue := f.Interface()
	// zero := f.IsZero()
	// current := fieldValue

	// fmt.Printf("zero:%+v\ncurrent:%+v", zero, current)

	switch v := fieldValue.(type) {
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int:
		return strconv.FormatInt(int64(v), 10)
	case float64:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case entity.User:
		return string(redis.Nil)
	case time.Time:
		t, err := time.Parse("2006-01-02 15:04:05 +0700 +07", v.String())
		if err != nil {
			fmt.Printf("time.Parse: %+v\nerror:%+v", v, err)
		}
		return strconv.FormatInt(t.UnixNano(), 10)
	case *time.Time:
		t, err := time.Parse("2006-01-02 15:04:05 +0700 +07", v.String())
		if err != nil {
			fmt.Printf("time.Parse: %+v\nerror:%+v", v, err)
		}
		return strconv.FormatInt(t.UnixNano(), 10)
	default:
		return ""
	}
}

func setField(obj interface{}, field string, value string) error {

	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(field)
	fieldValue := structFieldValue.Interface()

	if !structFieldValue.IsValid() {
		return fmt.Errorf("no such field: %s in obj", field)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("cannot set %s field value", field)
	}

	switch fieldValue.(type) {
	case int64:
		gv, _ := strconv.Atoi(value)
		structFieldValue.SetInt(int64(gv))
	case int32:
		gv, _ := strconv.Atoi(value)
		structFieldValue.SetInt(int64(gv))
	case int:
		gv, _ := strconv.Atoi(value)
		structFieldValue.SetInt(int64(gv))
	case float64:
		gv, _ := strconv.Atoi(value)
		structFieldValue.SetFloat(float64(gv))
	case string:
		structFieldValue.SetString(value)
	case bool:
		gv, _ := strconv.ParseBool(value)
		structFieldValue.SetBool(gv)
	case time.Time:
		gv, _ := strconv.Atoi(value)
		t := time.Unix(0, int64(gv))
		structFieldValue.Set(reflect.ValueOf(t))
		// if err != nil {
		// 	fmt.Printf("time.Parse: %+v\nerror:%+v", v, err)
		// }
		// return strconv.FormatInt(t.UnixNano(), 10)
	case *time.Time:
		gv, _ := strconv.Atoi(value)
		t := time.Unix(0, int64(gv))
		structFieldValue.Set(reflect.ValueOf(t))
		// t, err := time.Parse("2006-01-02 15:04:05 +0700 +07", v.String())
		// if err != nil {
		// 	fmt.Printf("time.Parse: %+v\nerror:%+v", v, err)
		// }
		// return strconv.FormatInt(t.UnixNano(), 10)
	default:
		structFieldValue.Set(structFieldValue)
	}

	// fmt.Printf("setField:\n"+
	// 	"structFieldValue:%+v\n"+
	// 	"val:%+v\n\n",
	// 	structFieldValue, val)
	return nil
}

func fillStruct(s interface{}, m map[string]string) error {
	for k, v := range m {
		err := setField(s, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func setArgs(value interface{}) []interface{} {
	var args []interface{}

	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	message_fields := make([]string, val.NumField())

	for i := 0; i < len(message_fields); i++ {
		fieldType := val.Type().Field(i)
		if !isZero(val.Interface(), fieldType.Name) {
			if !((fieldType.Name == "Creator") || (fieldType.Name == "Participant") || (fieldType.Name == "Event")) {
				args = append(args, fieldType.Name, getField(val.Interface(), fieldType.Name))
			}
		}
	}
	return args
}

// generate value type as given by json
func isZero(v interface{}, field string) bool {
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	return f.IsZero()
}
